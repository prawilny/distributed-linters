package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	pb "irio/linter_proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listen_addr = flag.String("address", "", "The server address")
	data_addr   = flag.String("data-addr", ":20000", "The HTTP data and healthcheck port")
	admin_addr  = flag.String("admin-addr", ":10000", "The GRPC admin port")
	lintCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadbalancer_lints",
		Help: "The total number of received lint requests",
	})
	setConfigCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadbalancer_set_configs",
		Help: "The total number of received setConfig requests",
	})
	healthcheckCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadbalancer_healthchecks",
		Help: "The total number of received healthcheck requests",
	})
	metricCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loadbalancer_metrics",
		Help: "The total number of received metrics requests",
	})
)

type workerConnectionState = int32

type workerState struct {
	connection  *grpc.ClientConn
	client      pb.LinterClient
	queueLength uint32
}

type version = string
type language = string
type addressPort = string

type languageConfig struct {
	weights            map[version]float32
	workers            map[version]map[addressPort]*workerState
	loadBalancingState map[version]float32
}

func makeLanguageConfig() languageConfig {
	return languageConfig{
		workers:            make(map[version]map[addressPort]*workerState),
		weights:            make(map[version]float32),
		loadBalancingState: make(map[version]float32),
	}
}

type loadBalancerServer struct {
	pb.UnimplementedLoadBalancerServer
	mut           sync.Mutex
	workerContext context.Context
	workerCancel  context.CancelFunc
	config        map[language]languageConfig
}

func (s *loadBalancerServer) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
	setConfigCounter.Inc()
	conf := make(map[string]languageConfig)
	for _, w := range req.Workers {
		addr := w.Address
		if _, ok := conf[w.Attrs.Language]; !ok {
			conf[w.Attrs.Language] = makeLanguageConfig()
		}
		if _, ok := conf[w.Attrs.Language].workers[w.Attrs.Version]; !ok {
			conf[w.Attrs.Language].workers[w.Attrs.Version] = make(map[addressPort]*workerState)
		}
		conf[w.Attrs.Language].workers[w.Attrs.Version][addr] = &workerState{}
	}
	for _, w := range req.Weights {
		if _, ok := conf[w.Attrs.Language]; !ok {
			conf[w.Attrs.Language] = makeLanguageConfig()
		}
		conf[w.Attrs.Language].weights[w.Attrs.Version] = w.Weight
	}
	for lang := range conf {
		hasWeights := false
		sum := float32(0.0)
		for _, w := range conf[lang].weights {
			if w != 0.0 {
				hasWeights = true
				sum += 1 / w
			}
		}
		if !hasWeights {
			return nil, status.Errorf(codes.InvalidArgument, "Language %s has no nonzero weights", lang)
		}
		for v, w := range conf[lang].weights {
			if w != 0 {
				conf[lang].weights[v] = (1 / w) / sum
			} else {
				conf[lang].weights[v] = 0
			}
		}
	}
	s.mut.Lock()
	defer s.mut.Unlock()
	for lang := range s.config {
		for version := range s.config[lang].workers {
			for addressPort := range s.config[lang].workers[version] {
				_, ok := conf[lang]
				if ok {
					_, ok = conf[lang].workers[version]
				}
				if ok {
					_, ok = conf[lang].workers[version][addressPort]
				}
				if !ok {
					s.config[lang].workers[version][addressPort].connection.Close()
					delete(s.config[lang].workers[version], addressPort)
				}
			}
		}
	}
	for lang := range conf {
		for version := range conf[lang].workers {
			for addressPort := range conf[lang].workers[version] {
				_, ok := s.config[lang]
				if ok {
					_, ok = s.config[lang].workers[version]
				}
				if ok {
					_, ok = s.config[lang].workers[version][addressPort]
				}
				if !ok {
					log.Printf("Dialing: %v", addressPort)
					connection, err := grpc.Dial(addressPort, grpc.WithInsecure())
					if err != nil {
						log.Fatalf("Failed to dial %v: %s", addressPort, err)
					}
					client := pb.NewLinterClient(connection)
					conf[lang].workers[version][addressPort] = &workerState{
						connection: connection,
						client:     client,
					}
				} else {
					conf[lang].workers[version][addressPort] = s.config[lang].workers[version][addressPort]
				}
			}
		}
	}
	s.config = conf
	return &pb.SetConfigResponse{Code: pb.SetConfigResponse_SUCCESS}, nil
}

func (lbServer *loadBalancerServer) handleLintRequest(res http.ResponseWriter, req *http.Request) bool {
	lang := req.URL.Query()["lang"][0]

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	lbServer.mut.Lock()
	_, ok := lbServer.config[lang]
	if !ok {
		lbServer.mut.Unlock()
		http.Error(res, fmt.Sprintf("Lang %s currently unavailable", lang), http.StatusServiceUnavailable)
		return true
	}

	version, maxVal := "", float32(-1000.0)
	for ver := range lbServer.config[lang].workers {
		curVal := lbServer.config[lang].loadBalancingState[ver]
		if lbServer.config[lang].weights[ver] != 0.0 && curVal > maxVal {
			version = ver
			maxVal = curVal
		}
	}
	if maxVal <= float32(0.0) {
		for ver := range lbServer.config[lang].workers {
			lbServer.config[lang].loadBalancingState[ver] += 1.0
		}
	}
	lbServer.config[lang].loadBalancingState[version] -= lbServer.config[lang].weights[version]
	var ws *workerState
	var addr addressPort
	for a, w := range lbServer.config[lang].workers[version] {
		switch w.connection.GetState() {
		case connectivity.Idle, connectivity.Ready:
			if ws == nil || w.queueLength < ws.queueLength {
				ws = w
				addr = a
			}
		default:
			continue
		}
	}
	if ws != nil {
		ws.queueLength += 1
		defer func() {
			lbServer.mut.Lock()
			ws.queueLength -= 1
			lbServer.mut.Unlock()
		}()
	}

	lbServer.mut.Unlock()

	if ws == nil {
		http.Error(res, fmt.Sprintf("Lang %s currently under maintenance. Please try again.", lang), http.StatusServiceUnavailable)
		return true
	}

	reply, err := ws.client.Lint(ctx, &pb.LintRequest{
		Language: lang,
		Content:  body,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			st := status.Convert(err)
			http.Error(res, fmt.Sprintf("Invalid argument: %s", st.Message()), http.StatusBadRequest)
			return true
		case codes.Unavailable:
			return false
		default:
			st := status.Convert(err)
			fmt.Printf("Unexpected return code from worker %s (lang %s, version %s). Message: %s", addr, lang, version, st.Message())
			http.Error(res, fmt.Sprintf("Internal server error"), http.StatusInternalServerError)
			return true
		}
	}

	fmt.Fprintf(res, "%v\n", reply)
	return true
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *admin_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	workerContext, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	lbServer := loadBalancerServer{
		workerContext: workerContext,
		workerCancel:  workerCancel,
	}
	pb.RegisterLoadBalancerServer(grpcServer, &lbServer)
	go grpcServer.Serve(lis)

	mux := http.NewServeMux()
	mux.HandleFunc("/lint", func(res http.ResponseWriter, req *http.Request) {
		lintCounter.Inc()
		for {
			if lbServer.handleLintRequest(res, req) {
				break
			}
		}
	})
	mux.HandleFunc("/health", func(res http.ResponseWriter, req *http.Request) {
		healthcheckCounter.Inc()
		fmt.Fprintf(res, "Healthy\n")
	})
	mux.HandleFunc("/metrics", func(res http.ResponseWriter, req *http.Request) {
		metricCounter.Inc()
		promhttp.Handler().ServeHTTP(res, req)
	})
	s := http.Server{
		Addr:    *data_addr,
		Handler: mux,
	}
	log.Fatal(s.ListenAndServe())
}
