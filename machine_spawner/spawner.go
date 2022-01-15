package main

import (
	"context"
	"fmt"
	pb "irio/linter_proto"
    "net"
	"time"
    "log"

	"google.golang.org/grpc"

    //"k8s.io/apimachinery/pkg/api/errors"
    //metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

type adminRespondingServer struct {
    pb.UnimplementedMachineSpawnerServer
    client kubernetes.Clientset
}

func (s *adminRespondingServer) AddLinter(ctx context.Context, in *pb.AddLinterRequest) (*pb.AddLinterResponse, error) {
    log.Printf("received %s", in.String())
    // 
    return &pb.AddLinterResponse{Resp: pb.AddLinterResponse_Ok}, nil
}

func makeMachineSpawner() adminRespondingServer {
    _, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

    config, err := rest.InClusterConfig()
    if err != nil {
        panic(err.Error())
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    return adminRespondingServer {client: *clientset}
}

func main() {
    machine_spawner := makeMachineSpawner()

    lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", 1337))
    if nil != err {
        panic("failed to listen")
    }

    var opts []grpc.ServerOption
    grpcServer := grpc.NewServer(opts...)
    pb.RegisterMachineSpawnerServer(grpcServer, &machine_spawner)
    log.Fatal(grpcServer.Serve(lis))
}
