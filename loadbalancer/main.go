package main

import (
    "context"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    //"path"
    "net/http"
    "net"
    "sync"
    "time"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/connectivity"
    pb "irio/linter_proto"
    //"github.com/davecgh/go-spew/spew"
)

var (
    listen_addr = flag.String("address", "localhost", "The server address")
    data_port = flag.Int("data-port", -1, "The HTTP data port")
    admin_port = flag.Int("admin-port", -1, "The GRPC admin port")
)

type workerConnectionState = int32

type workerState struct {
    connection *grpc.ClientConn
    client pb.LinterClient
    queueLength uint32
}

type version = string
type language = string
type addressPort = string

type languageConfig struct {
    weights map[version]float32
    workers map[version]map[addressPort]*workerState
    loadBalancingState map[version]float32
}
func makeLanguageConfig() languageConfig {
    return languageConfig{
        workers: make(map[version]map[addressPort]*workerState),
        weights: make(map[version]float32),
        loadBalancingState: make(map[version]float32),
    }
}

type loadBalancerServer struct {
    pb.UnimplementedLoadBalancerServer
    mut sync.Mutex
    workerContext context.Context
    workerCancel context.CancelFunc
    config map[language]languageConfig
}

func (s *loadBalancerServer) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
    conf := make(map[string]languageConfig) 
    for _, w := range req.Workers {
        addr := fmt.Sprintf("%s:%d", w.Address, w.Port)
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
                sum += 1/w
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
                        client: client,
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

func (lbServer *loadBalancerServer) handleLintRequest(res http.ResponseWriter, req *http.Request) (bool) {
    //p := path.Base(req.URL.Path)
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
        http.Error(res, fmt.Sprintf("Lang %s currently unavailable", lang), http.StatusServiceUnavailable)
        lbServer.mut.Unlock()
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
        fmt.Printf("%v\n", w.connection.GetState())
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
        Content: body,
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
    if *data_port < 0 {
        log.Fatalf("Bad -data-port: %d", *data_port)
    }
    if *admin_port < 0 {
        log.Fatalf("Bad -admin-port: %d", *admin_port)
    }

    lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *listen_addr, *admin_port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    workerContext, workerCancel := context.WithCancel(context.Background())
    defer workerCancel()
    lbServer := loadBalancerServer{
        workerContext: workerContext,
        workerCancel: workerCancel,
    }
    pb.RegisterLoadBalancerServer(grpcServer, &lbServer)
    go grpcServer.Serve(lis)

    mux := http.NewServeMux()
    mux.HandleFunc("/lint", func(res http.ResponseWriter, req *http.Request) {
        for {
            if lbServer.handleLintRequest(res, req) {
                break
            }
        }
    })
    mux.HandleFunc("/health", func(res http.ResponseWriter, req *http.Request) {
        fmt.Fprintf(res, "Healthy\n")
    })
    s := http.Server{
        Addr:    fmt.Sprintf("%s:%d", *listen_addr, *data_port),
        Handler: mux,
    }
    log.Fatal(s.ListenAndServe())
}
