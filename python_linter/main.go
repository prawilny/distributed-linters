package main

import (
	"context"
	//"encoding/json"
	"flag"
	"fmt"
	//"io"
	//"io/ioutil"
	"log"
	//"math"
	"net"
	"regexp"
	"unicode/utf8"
	//"sync"
	//"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/credentials"
	//"google.golang.org/grpc/examples/data"
	//"github.com/golang/protobuf/proto"
    pb "irio/linter_proto"
)

func merge(a []int, b []int) []int {
    final := make([]int, len(a) + len(b))
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
	listen_addr = flag.String("address", "localhost", "The server address")
	listen_port = flag.Int("port", -1, "The server port")
)

type linterServer struct {
	pb.UnimplementedLinterServer
    regexLeft *regexp.Regexp
    regexRight *regexp.Regexp
}
func makeLinterServer() *linterServer {
    return &linterServer{
        regexLeft: regexp.MustCompile("[[:alnum:]]="),
        regexRight: regexp.MustCompile("=[^=[:space:]]"),
    }
}

func (s *linterServer) Lint(ctx context.Context, req *pb.LintRequest) (*pb.LintResponse, error) {
    text := req.Content
    if !utf8.Valid(req.Content) {
        return nil, status.Error(codes.InvalidArgument, "Request text is not valid utf-8")
    }

    ans := make([]*pb.LintResponse_Hint, 0, 0)
    a := s.regexLeft.FindAllIndex(text, -1)
    b := s.regexRight.FindAllIndex(text, -1)

    final := make([]int, 0, len(a) + len(b))
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
        ans = append(ans, &pb.LintResponse_Hint{HintText: "= is not surrounded by whitespace", StartByte: int32(v), EndByte: int32(v+1)})
    }
    return &pb.LintResponse{Hints: ans}, nil
}

func main() {
	flag.Parse()
    if *listen_port < 0 {
        log.Fatalf("Bad -port: %d", *listen_port)
    }
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *listen_addr, *listen_port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterLinterServer(grpcServer, makeLinterServer())
	grpcServer.Serve(lis)
}

