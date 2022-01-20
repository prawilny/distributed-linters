package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"unicode/utf8"

	pb "irio/linter_proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func merge(a []int, b []int) []int {
	final := make([]int, len(a)+len(b))
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			final[i+j] = a[i]
		} else {
			final[i+j] = b[j]
		}
	}
	for ; i < len(a); i++ {
		final[i+j] = a[i]
	}
	for ; j < len(b); j++ {
		final[i+j] = b[j]
	}
	return final
}

var (
	health_addr = flag.String("health-addr", ":60000", "The HTTP healthcheck port")
	data_addr   = flag.String("data-addr", ":20000", "The GRPC data port")
	lintCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "java_linter_lints",
		Help: "The total number of received lint requests",
	})
	healthcheckCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "java_linter_healthchecks",
		Help: "The total number of received healthcheck requests",
	})
	metricCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "java_linters_metrics",
		Help: "The total number of received metrics requests",
	})
)

type linterServer struct {
	pb.UnimplementedLinterServer
	regexLeft  *regexp.Regexp
	regexRight *regexp.Regexp
}

func makeLinterServer() *linterServer {
	return &linterServer{
		regexLeft:  regexp.MustCompile("[[:alnum:]]="),
		regexRight: regexp.MustCompile("=[^=[:space:]]"),
	}
}

func (s *linterServer) Lint(ctx context.Context, req *pb.LintRequest) (*pb.LintResponse, error) {
	lintCounter.Inc()
	text := req.Content
	if !utf8.Valid(req.Content) {
		return nil, status.Error(codes.InvalidArgument, "Request text is not valid utf-8")
	}

	ans := make([]*pb.LintResponse_Hint, 0, 0)
	a := s.regexLeft.FindAllIndex(text, -1)
	b := s.regexRight.FindAllIndex(text, -1)

	final := make([]int, 0, len(a)+len(b))
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		if a[i][1]-1 == b[j][0] {
			final = append(final, a[i][1]-1)
			i += 1
			j += 1
		} else if a[i][1]-1 < b[j][0] {
			final = append(final, a[i][1]-1)
			i += 1
		} else {
			final = append(final, b[j][0])
			j += 1
		}
	}
	for ; i < len(a); i++ {
		final = append(final, a[i][1]-1)
	}
	for ; j < len(b); j++ {
		final = append(final, b[j][0])
	}

	for _, v := range final {
		ans = append(ans, &pb.LintResponse_Hint{HintText: "= is not surrounded by whitespace", StartByte: int32(v), EndByte: int32(v + 1)})
	}
	return &pb.LintResponse{Hints: ans}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *data_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterLinterServer(grpcServer, makeLinterServer())
	go grpcServer.Serve(lis)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(res http.ResponseWriter, req *http.Request) {
		healthcheckCounter.Inc()
		ioutil.ReadAll(req.Body)
		fmt.Fprintf(res, "Healthy\n")
	})
	mux.HandleFunc("/metrics", func(res http.ResponseWriter, req *http.Request) {
		metricCounter.Inc()
		promhttp.Handler().ServeHTTP(res, req)
	})
	s := http.Server{
		Addr:    *health_addr,
		Handler: mux,
	}
	log.Fatal(s.ListenAndServe())
}
