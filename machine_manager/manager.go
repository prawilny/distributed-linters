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

    //"k8s.io/apimachinery/pkg/api/errors"
    //metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    //"k8s.io/client-go/rest"
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
	for k := range mis {
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
			for mach := range machMap {
				workers = append(workers, &pb.Worker{
					Address:  mach.Address,
					Port:     mach.Port,
                    Attrs:    &pb.LinterAttributes {  
                        Language: lang,
                        Version:  ver,
                    },
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

func (w *WorkersInfo) removeMachinesForLinter(lang language, ver version) {
	map_, exists1 := w.Machines[lang]
	if !exists1 {
		return
	}
    delete(map_, ver)
    if len(map_) == 0 {
        delete(w.Machines, lang)
    }
}

var (
    listen_addr = flag.String("address", "0.0.0.0:2137", "The Admin CLI Listen address (with port)")
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
    client kubernetes.Clientset
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

/*
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
*/

func (s *machineManagerServer) createLinterMachine(lang language, ver version) (MachineInfo, error) {
    // TODO: actually create this linter
    return MachineInfo{Address: "abcd(change this)", Port: int32(2137)}, nil
}

func (s *machineManagerServer) removeLinterMachine(info MachineInfo) {
    // TODO: actually remove the machine
}

func (s *machineManagerServer) removeLinterMachines(lang language, ver version) {
    // TODO: actually create this linter
    map1, exists1 := s.state.Linters.Machines[lang]
    if !exists1 {
        return
    }
    map2, exists2 := map1[ver]
    if !exists2 {
        return
    }
    for info := range map2 {
        s.removeLinterMachine(info)
    }
}

func (s *machineManagerServer) AppendLinter(ctx context.Context, req *pb.LinterAttributes) (*pb.LinterResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
    machine, err := s.createLinterMachine(req.Language, req.Version)
    if err != nil {
        return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
    }
	s.state.Linters.addMachine(req.Language, req.Version, machine)
	s.storeState()
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *machineManagerServer) RemoveLinter(ctx context.Context, req *pb.LinterAttributes) (*pb.LinterResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

    s.removeLinterMachines(req.Language, req.Version)
    s.state.Linters.removeMachinesForLinter(req.Language, req.Version)
	s.storeState()
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *machineManagerServer) SetProportions(ctx context.Context, req *pb.LoadBalancingProportions) (*pb.LinterResponse, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	for machine := range s.state.Load_balancers.Machines {
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
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func main() {
	flag.Parse()
	machine_manager := makeMachineManager()

	lis, err := net.Listen("tcp", *listen_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var server_opts []grpc.ServerOption
	grpcServer := grpc.NewServer(server_opts...)
	pb.RegisterMachineManagerServer(grpcServer, &machine_manager)
	log.Fatal(grpcServer.Serve(lis))

    //machine_spawner := makeMachineSpawner()
    //pb.RegisterMachineSpawnerServer(grpcServer, &machine_spawner)
    //log.Fatal(grpcServer.Serve(lis))
}

/*
type adminRespondingServer struct {
    pb.UnimplementedMachineSpawnerServer
    client kubernetes.Clientset
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
*/
