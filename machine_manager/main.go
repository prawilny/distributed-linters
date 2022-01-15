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

	etcd3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type language = string
type version = string
type address = string
type port = int32

type MachineInfo struct {
	Address address
	Port    port
}
type MachineInfoSet map[MachineInfo]bool

func (mis MachineInfoSet) MarshalJSON() (data []byte, err error) {
	machines := make([]MachineInfo, 0, len(mis))
	for k, _ := range mis {
		machines = append(machines, k)
	}
	return json.Marshal(machines)
}

func (mis *MachineInfoSet) UnmarshalJSON(data []byte) error {
	*mis = make(MachineInfoSet)

	var machines []MachineInfo
	err := json.Unmarshal(data, &machines)
	if err != nil {
		return err
	}

	for _, machine := range machines {
		(*mis)[machine] = true
	}
	return nil
}

type LBWorkersInfo struct {
	Machines MachineInfoSet
}
type WorkersInfo struct {
	Machines map[language]map[version]MachineInfoSet
}

func (lb *LBWorkersInfo) addMachine(machine MachineInfo) {
	lb.Machines[machine] = true
}

func (lb *LBWorkersInfo) removeMachine(machine MachineInfo) {
	delete(lb.Machines, machine)
}

func (w *WorkersInfo) getMachines() []*pb.Worker {
	workers := []*pb.Worker{}
	for lang, verMap := range w.Machines {
		for ver, machMap := range verMap {
			for mach, _ := range machMap {
				workers = append(workers, &pb.Worker{
					Address:  mach.Address,
					Port:     mach.Port,
					Language: lang,
					Version:  ver,
				})
			}
		}
	}
	return workers
}

func (w *WorkersInfo) addMachine(lang language, ver version, machine MachineInfo) {
	if _, exists := w.Machines[lang]; !exists {
		w.Machines[lang] = make(map[version]MachineInfoSet)
	}
	if _, exists := w.Machines[lang][ver]; !exists {
		w.Machines[lang][ver] = make(MachineInfoSet)
	}
	w.Machines[lang][ver][machine] = true
}

func (w *WorkersInfo) removeMachine(lang language, ver version, machine MachineInfo) {
	map1, exists1 := w.Machines[lang]
	if !exists1 {
		return
	}
	map2, exists2 := w.Machines[lang][ver]
	if !exists2 {
		return
	}
	delete(map2, machine)
	if len(map2) == 0 {
		delete(map1, lang)
	}
	if len(map1) == 0 {
		delete(w.Machines, lang)
	}
}

var (
	listen_addr = flag.String("address", "localhost", "The server address")
	grpc_port   = flag.Int("grpc-port", -1, "The gRPC port")
	grpc_opts   = []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}

	dialTimeout               = 2 * time.Second
	requestTimeout            = 10 * time.Second
	etcd_addrs                = []string{"localhost:2379"}
	machine_manager_state_key = "MACHINE_MANAGER_STATE"
)

type machineManagerState struct {
	Load_balancers LBWorkersInfo
	Linters        WorkersInfo
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

	kv := etcd3.NewKV(cli)
	state := machineManagerState{
		Load_balancers: LBWorkersInfo{Machines: make(MachineInfoSet)},
		Linters:        WorkersInfo{Machines: make(map[string]map[string]MachineInfoSet)},
	}
	resp, err := kv.Get(ctx, machine_manager_state_key)
	if err != nil {
		panic(fmt.Sprintf("error getting lastest machine manager state from etcd: %s", err.Error()))
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
	serializedBytes, err := json.Marshal(s.state)
	if err != nil {
		panic(fmt.Sprintf("error marshalling machine manager state: %s", err.Error()))
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()
	_, err = s.etcd.Put(ctx, machine_manager_state_key, string(serializedBytes))
	if err != nil {
		panic(fmt.Sprintf("error saving machine manager state: %s", err.Error()))
	}
}

func (s *machineManagerServer) AppendLoadBalancer(ctx context.Context, req *pb.LBWorker) (*pb.AppendMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.Load_balancers.addMachine(MachineInfo{Address: req.Address, Port: req.Port})
	s.storeState()
	return &pb.AppendMachineResponse{Code: pb.AppendMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) RemoveLoadBalancer(ctx context.Context, req *pb.LBWorker) (*pb.RemoveMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.Load_balancers.removeMachine(MachineInfo{Address: req.Address, Port: req.Port})
	s.storeState()
	return &pb.RemoveMachineResponse{Code: pb.RemoveMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) AppendLinter(ctx context.Context, req *pb.Worker) (*pb.AppendMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.Linters.addMachine(req.Language, req.Version, MachineInfo{Address: req.Address, Port: req.Port})
	s.storeState()
	return &pb.AppendMachineResponse{Code: pb.AppendMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) RemoveLinter(ctx context.Context, req *pb.Worker) (*pb.RemoveMachineResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.state.Linters.removeMachine(req.Language, req.Version, MachineInfo{Address: req.Address, Port: req.Port})
	s.storeState()
	return &pb.RemoveMachineResponse{Code: pb.RemoveMachineResponse_SUCCESS}, nil
}

func (s *machineManagerServer) SetProportions(ctx context.Context, req *pb.LoadBalancingProportions) (*pb.SetProportionsResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	for machine, _ := range s.state.Load_balancers.Machines {
		peerAddr := fmt.Sprintf("%s:%d", machine.Address, machine.Port)
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

		client.SetConfig(ctx, &pb.SetConfigRequest{Workers: s.state.Linters.getMachines(), Weights: req.Weights})
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
