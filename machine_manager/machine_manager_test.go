package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"reflect"
	"testing"
	"time"

	pb "irio/linter_proto"

	"google.golang.org/grpc"
)

var (
	machine_manager_port = 11111
	load_balancer_port   = 22222
	client_opts          = []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
	state                = testLoadBalancerState{
		workers: WorkersInfo{
			Machines: make(map[string]map[string]MachineInfoSet),
		},
		lastWeights: nil,
	}
)

type testLoadBalancerState struct {
	workers     WorkersInfo
	lastWeights []*pb.Weight
}

type testLoadBalancer struct {
	pb.UnimplementedLoadBalancerServer
	state *testLoadBalancerState
	t     *testing.T
}

func (s *testLoadBalancer) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
	weightsEqual := reflect.DeepEqual(s.state.lastWeights, req.Weights)
	weightsEqual = weightsEqual || s.state.lastWeights == nil && len(req.Weights) == 0
	weightsEqual = weightsEqual || req.Weights == nil && len(s.state.lastWeights) == 0
	if weightsEqual {
		fmt.Println("good weights")
	} else {
		s.t.Errorf("weights (%T, %T) not equal: %v != %v", s.state.lastWeights, req.Weights, s.state.lastWeights, req.Weights)
	}

	reqMachines := WorkersInfo{Machines: make(map[language]map[version]MachineInfoSet)}
	for _, m := range req.Workers {
		reqMachines.addMachine(m.Language, m.Version, MachineInfo{Address: m.Address, Port: m.Port})
	}

	workersEqual := reflect.DeepEqual(s.state.workers.Machines, reqMachines.Machines)
	if workersEqual {
		fmt.Println("good workers")
	} else {
		s.t.Errorf("workers not equal: %#v != %#v", state.workers.Machines, reqMachines.Machines)
	}

	return &pb.SetConfigResponse{Code: pb.SetConfigResponse_SUCCESS}, nil
}

func TestMachineManager(t *testing.T) {
	machineManagerServerAddr := fmt.Sprintf("localhost:%d", machine_manager_port)
	log.Printf("Dialing: %v", machineManagerServerAddr)
	conn, err := grpc.Dial(machineManagerServerAddr, client_opts...)
	log.Printf("Connected to: %v", machineManagerServerAddr)
	if err != nil {
		t.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewMachineManagerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", load_balancer_port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var server_opts []grpc.ServerOption
	grpcServer := grpc.NewServer(server_opts...)

	testLB := testLoadBalancer{t: t, state: &state}
	pb.RegisterLoadBalancerServer(grpcServer, &testLB)
	go grpcServer.Serve(lis)

	////////////////////////////////////////////////////////////////////////////////
	// setup above
	////////////////////////////////////////////////////////////////////////////////

	_, err = client.AppendLoadBalancer(ctx, &pb.LBWorker{Address: "localhost", Port: int32(load_balancer_port)})
	if err != nil {
		t.Errorf("AppendLoadBalancer() failed: %s", err.Error())
	}

	lang := "holyC"
	ver := "5.03"
	addr := "temple.os"
	port := int32(640)

	state.workers.addMachine(lang, ver, MachineInfo{Address: addr, Port: port})
	_, err = client.AppendLinter(ctx, &pb.Worker{Language: lang, Version: ver, Address: addr, Port: port})
	if err != nil {
		t.Errorf("AppendLinter() failed: %s", err.Error())
	}

	state.lastWeights = []*pb.Weight{}
	_, err = client.SetProportions(ctx, &pb.LoadBalancingProportions{Weights: state.lastWeights})
	if err != nil {
		t.Errorf("setting proportions failed: %s", err.Error())
	}
}
