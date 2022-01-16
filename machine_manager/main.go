package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pb "irio/linter_proto"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
		log.Println("error unmarshaling JSON: " + err.Error())
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
					Address: mach.Address,
					Port:    mach.Port,
					Attrs: &pb.LinterAttributes{
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

	dialTimeout                = 2 * time.Second
	requestTimeout             = 10 * time.Second
	etcd_addrs                 = []string{"localhost:2379"}
	machine_manager_config_map = "machine-manager-config-map"
	linter_port                = 33333
)

type machineManagerState struct {
	Load_balancers LBWorkersInfo
	Linters        WorkersInfo
}

type machineManagerServer struct {
	pb.UnimplementedMachineManagerServer
	state  machineManagerState
	client kubernetes.Clientset
	mut    sync.Mutex
}

func makeMachineManager() machineManagerServer {
	l := log.New(os.Stderr, "", 0)
	l.Println(os.Environ())
	l.Println(os.Getenv("KUBERNETES_SERVICE_HOST"))
	l.Println(os.Getenv("KUBERNETES_SERVICE_PORT"))

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	state := machineManagerState{
		Load_balancers: LBWorkersInfo{Machines: make(MachineInfoSet)},
		Linters:        WorkersInfo{Machines: make(map[string]map[string]MachineInfoSet)},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	get, err := client.CoreV1().ConfigMaps("default").Get(ctx, machine_manager_config_map, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	jsonData := get.Data["json"]
	if len(jsonData) != 0 {
		err = json.Unmarshal([]byte(jsonData), &state)
		if err != nil {
			panic(err.Error())
		}
	}

	///////////////////////////////////////////////////////////////////////////
	log.Printf("get: %v\n", get)
	log.Println("jsonData: " + jsonData)
	serializedState, err := json.Marshal(&state)
	if err != nil {
		panic(err.Error())
	}
	log.Println("serializedState: " + string(serializedState))
	updatedMap := apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-manager-config-map",
		},
		Data: map[string]string{
			"json": string(serializedState),
		},
	}
	_, err = client.CoreV1().ConfigMaps("default").Update(ctx, &updatedMap, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	get, err = client.CoreV1().ConfigMaps("default").Get(ctx, machine_manager_config_map, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	log.Printf("get: %v\n", get)
	///////////////////////////////////////////////////////////////////////////

	return machineManagerServer{
		state:  state,
		client: *client,
		mut:    sync.Mutex{},
	}
}

func int32Ptr(i int32) *int32 { return &i }

func deploymentNameFromAttrs(lang language, ver version) string {
	return fmt.Sprintf("linter-%s-%s", lang, strings.ReplaceAll(ver, ".", "-"))
}

func deploymentFromLabels(lang language, ver version, imageUrl string) appsv1.Deployment {
	linterName := deploymentNameFromAttrs(lang, ver)
	labels := map[string]string{
		"version":  ver,
		"language": lang,
		"app":      linterName,
	}
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   linterName,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            linterName,
							Image:           imageUrl,
							ImagePullPolicy: "Never",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: int32(linter_port),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *machineManagerServer) AppendLinter(ctx context.Context, req *pb.AppendLinterRequest) (*pb.LinterResponse, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	deploymentsClient := s.client.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := deploymentFromLabels(req.Attrs.Language, req.Attrs.Version, req.ImageUrl)
	_, err := deploymentsClient.Create(ctx, &deployment, metav1.CreateOptions{})
	if err != nil {
		log.Println("error appending linter: " + err.Error())
		return &pb.LinterResponse{Code: 1}, nil
	}
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *machineManagerServer) RemoveLinter(ctx context.Context, req *pb.LinterAttributes) (*pb.LinterResponse, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	deploymentsClient := s.client.AppsV1().Deployments(apiv1.NamespaceDefault)
	deploymentName := deploymentNameFromAttrs(req.Language, req.Version)
	err := deploymentsClient.Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("error removing linter: " + err.Error())
		return &pb.LinterResponse{Code: 1}, nil
	}
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
}
