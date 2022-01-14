package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pb "irio/linter_proto"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	etcd3 "go.etcd.io/etcd/client/v3"
)

type language = string
type version = string
type address = string
type port = int32

type machineInfo struct {
	address address
	port    port
}
type machines = map[machineInfo]struct{}

type LBWorkersInfo machines
type WorkersInfo map[language]map[version]machines

func (lb LBWorkersInfo) addMachine(machine machineInfo) {
	lb[machine] = exists
}

func (lb LBWorkersInfo) removeMachine(machine machineInfo) {
	delete(lb, machine)
}

func (w WorkersInfo) addMachine(lang language, ver version, machine machineInfo) {
	if _, exists := w[lang]; !exists {
		w[lang] = make(map[version]map[machineInfo]struct{})
	}
	if _, exists := w[lang][ver]; !exists {
		w[lang][ver] = make(map[machineInfo]struct{})
	}
	w[lang][ver][machine] = exists
}

func (w WorkersInfo) removeMachine(lang language, ver version, machine machineInfo) {
	map1, exists1 := w[lang]
	if !exists1 {
		return
	}
	map2, exists2 := w[lang][ver]
	if !exists2 {
		return
	}
	delete(map2, machine)
	if len(map2) == 0 {
		delete(map1, lang)
	}
	if len(map1) == 0 {
		delete(w, lang)
	}
}

var (
	listen_addr = flag.String("address", "localhost", "The server address")
	grpc_port   = flag.Int("grpc-port", -1, "The gRPC port")
	grpc_opts   = []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}

	dialTimeout               = 2 * time.Second
	requestTimeout            = 10 * time.Second
	etcd_addrs                = []string{"etcd:2379"}
	machine_manager_state_key = "MACHINE_MANAGER_STATE"

	exists = struct{}{}
)

type machineManagerState struct {
	load_balancers LBWorkersInfo
	linters        WorkersInfo
}

type machineManagerServer struct {
	pb.UnimplementedMachineManagerServer
	state machineManagerState
	etcd  etcd3.KV
	mut   sync.Mutex
}

func makeMachineManager() machineManagerServer {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()
	cli, err := etcd3.New(etcd3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   etcd_addrs,
	})
	if err != nil {
		panic(fmt.Sprintf("error creating etcd connection: %s", err.Error()))
	}
	defer cli.Close()

	kv := etcd3.NewKV(cli)
	resp, err := kv.Get(ctx, machine_manager_state_key)
	if err != nil {
		panic(fmt.Sprintf("error getting lastest machine manager state from etcd: %s", err.Error()))
	}

	state := machineManagerState{
		load_balancers: make(LBWorkersInfo),
		linters:        make(WorkersInfo),
	}
	if len(resp.Kvs) != 0 {
		err = json.Unmarshal(resp.Kvs[0].Value, &state)
		if err != nil {
			panic(fmt.Sprintf("error unmarshaling lastest machine manager state: %s", err.Error()))
		}
	}

	return machineManagerServer{
		state: state,
		etcd:  kv,
	}
}

func (s *machineManagerServer) storeState() {
	serialized, err := json.Marshal(s.state)
	if err != nil {
		panic(fmt.Sprintf("error marshalling machine manager state: %s", err.Error()))
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()
	_, err = s.etcd.Put(ctx, machine_manager_state_key, string(serialized))
	if err != nil {
		panic(fmt.Sprintf("error saving machine manager state: %s", err.Error()))
	}
}

func (s *machineManagerServer) AppendLoadBalancer(ctx context.Context, req *pb.LBWorker) (*pb.AppendMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.load_balancers.addMachine(machineInfo{address: req.Address, port: req.Port})
	s.storeState()
	return &pb.AppendMachineResponse{Code: pb.AppendMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) RemoveLoadBalancer(ctx context.Context, req *pb.LBWorker) (*pb.RemoveMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.load_balancers.removeMachine(machineInfo{address: req.Address, port: req.Port})
	s.storeState()
	return &pb.RemoveMachineResponse{Code: pb.RemoveMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) AppendLinter(ctx context.Context, req *pb.Worker) (*pb.AppendMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.linters.addMachine(req.Language, req.Version, machineInfo{address: req.Address, port: req.Port})
	s.storeState()
	return &pb.AppendMachineResponse{Code: pb.AppendMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) RemoveLinter(ctx context.Context, req *pb.Worker) (*pb.RemoveMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.linters.removeMachine(req.Language, req.Version, machineInfo{address: req.Address, port: req.Port})
	s.storeState()
	return &pb.RemoveMachineResponse{Code: pb.RemoveMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) SetProportions(ctx context.Context, req *pb.LoadBalancingProportions) (*pb.SetProportionsResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	for machine, _ := range s.state.load_balancers {
		peerAddr := fmt.Sprintf("%s:%d", machine.address, machine.port)
		conn, err := grpc.Dial(peerAddr, grpc_opts...)
		log.Printf("Connected to: %v", peerAddr)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
			// TODO: better error handling?
		}
		defer conn.Close()
		client := pb.NewLoadBalancerClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client.SetConfig(ctx, &pb.SetConfigRequest{Workers: []*pb.Worker{}, Weights: []*pb.Weight{}})
	}
	return &pb.SetProportionsResponse{Code: pb.SetProportionsResponse_SUCCESS}, nil
}

func main() {
	flag.Parse()
	if *grpc_port < 0 {
		log.Fatalf("Bad -grpc-port: %d", *grpc_port)
	}
	machine_manager := makeMachineManager()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *listen_addr, *grpc_port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var server_opts []grpc.ServerOption
	grpcServer := grpc.NewServer(server_opts...)
	pb.RegisterMachineManagerServer(grpcServer, &machine_manager)
	log.Fatal(grpcServer.Serve(lis))
}
