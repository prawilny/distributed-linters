package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pb "irio/linter_proto"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	//appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	ADD_WORKER PodEventType = iota
	ADD_LB     PodEventType = iota
	REMOVE_POD PodEventType = iota
)

const (
	LOAD_BALANCER PodRole = iota
	LINTER        PodRole = iota
)

type PodRole int
type PodEventType int
type Language string
type Version string
type Address string
type Port int32

type MachineManagerServer struct {
	pb.UnimplementedMachineManagerServer
	State      MachineManagerState
	kubernetes kubernetes.Clientset
	mut        sync.Mutex
	eventChan  chan *PodEvent
}

type MachineManagerState struct {
	Persistent    MachineManagerPersistentState
	LoadBalancers map[types.NamespacedName]LBInfo
	Linters       map[types.NamespacedName]WorkerInfo
}

type MachineManagerPersistentState struct {
	Weights []*pb.Weight
}

type WorkerInfo struct {
	address Address
	lang    Language
	ver     Version
}

type LBInfo struct {
	conn        grpc.ClientConn
	cancelCtx   context.CancelFunc
	commandChan chan<- bool
}

type PodEvent struct {
	Type     PodEventType
	Name     *types.NamespacedName
	Addr     *Address
	Version  *Version
	Language *Language
}

type podInfo struct {
	address string
}

type PodReconciler struct {
	ctrlClient.Client
	mut    sync.Mutex
	pods   map[string]podInfo
	events chan<- *PodEvent
}

const (
	markerLabel                         = "xD"
	requestTimeout                      = 3 * time.Second
	machine_manager_config_map          = "machine-manager-config-map"
	machine_manager_config_map_data_key = "json"
	load_balancer_port                  = 33333
	linter_port                         = 55555
	num_retries                         = 3
)

var (
	listen_addr = flag.String("address", ":2137", "The Admin CLI Listen address (with port)")
	grpc_opts   = []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
)

func (a *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	err := a.Get(ctx, req.NamespacedName, pod)
	deleted := false
	created := false
	info := podInfo{}
	ok := true
	if apierrors.IsNotFound(err) {
		a.mut.Lock()
		if info, ok = a.pods[req.NamespacedName.String()]; ok {
			delete(a.pods, req.NamespacedName.String())
			deleted = true

			a.events <- &PodEvent{
				Type: REMOVE_POD,
				Name: &req.NamespacedName,
			}
		}
		a.mut.Unlock()
	} else if err != nil {
		fmt.Printf("Error processing event for %s: %s\n", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else {
		if pod.Status.Phase == corev1.PodRunning {
			a.mut.Lock()
			if _, ok = a.pods[req.NamespacedName.String()]; !ok {
				address := pod.Status.PodIP
				a.pods[req.NamespacedName.String()] = podInfo{
					address: address,
				}

				role := LOAD_BALANCER
				if pod.Labels[markerLabel] == "linter" {
					role = LINTER
				}
				event := PodEvent{
					Type: ADD_LB,
					Name: &req.NamespacedName,
					Addr: (*Address)(&address),
				}
				if role == LOAD_BALANCER {
					event.Type = ADD_LB
				} else {
					event.Type = ADD_WORKER
					ver := Version(pod.Labels["version"])
					lang := Language(pod.Labels["language"])
					event.Version = &ver
					event.Language = &lang
				}
				a.events <- &event

				info = a.pods[req.NamespacedName.String()]
				created = true
			}
			a.mut.Unlock()
		}
	}

	if created {
		log.Printf("Created: %s %v\n", req.NamespacedName, info)
	} else if deleted {
		log.Printf("Deleted: %s %v\n", req.NamespacedName, info)
	} else {
		log.Printf("Changed: %s\n", req.NamespacedName)
	}

	return ctrl.Result{}, nil
}

func (mms *MachineManagerState) getWorkers() []*pb.Worker {
	workers := []*pb.Worker{}
	for _, worker := range mms.Linters {
		workers = append(workers, &pb.Worker{
			Address: string(worker.address),
			Port:    linter_port,
			Attrs: &pb.LinterAttributes{
				Language: string(worker.lang),
				Version:  string(worker.ver),
			},
		})
	}
	return workers
}

func (mms *MachineManagerState) hasMachinesFor(lang Language, ver Version) bool {
	for _, worker := range mms.Linters {
		if worker.lang == lang && worker.ver == ver {
			return true
		}
	}
	return false
}

func makeMachineManager() MachineManagerServer {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	kubernetes, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	eventChan := make(chan *PodEvent)

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		panic("couldn't create manager: " + err.Error())
	}

	pred, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: markerLabel, Operator: metav1.LabelSelectorOpExists}}})
	if err != nil {
		panic("could not create predicate: " + err.Error())
	}

	err = ctrl.
		NewControllerManagedBy(manager).
		For(&corev1.Pod{}, builder.WithPredicates(pred)).
		Complete(&PodReconciler{
			Client: manager.GetClient(),
			pods:   make(map[string]podInfo),
			events: eventChan,
		})
	if err != nil {
		panic("could not create controller: " + err.Error())
	}

	err = manager.Start(ctrl.SetupSignalHandler())
	if err != nil {
		panic("could not start manager: " + err.Error())
	}

	state := MachineManagerState{
		Persistent: MachineManagerPersistentState{
			Weights: []*pb.Weight{},
		},
		LoadBalancers: make(map[types.NamespacedName]LBInfo),
		Linters:       make(map[types.NamespacedName]WorkerInfo),
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	get, err := kubernetes.CoreV1().ConfigMaps(apiv1.NamespaceDefault).Get(ctx, machine_manager_config_map, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	jsonData := get.Data[machine_manager_config_map_data_key]
	if len(jsonData) != 0 {
		err = json.Unmarshal([]byte(jsonData), &state)
		if err != nil {
			panic(err.Error())
		}
	}

	// TODO: set sensible capacities of channels
	return MachineManagerServer{
		State:      state,
		kubernetes: *kubernetes,
		eventChan:  eventChan,
	}
}

func int32Ptr(i int32) *int32 { return &i }

func deploymentNameFromAttrs(lang Language, ver Version) string {
	return fmt.Sprintf("linter-%s-%s", string(lang), strings.ReplaceAll(string(ver), ".", "-"))
}

func deploymentFromLabels(lang Language, ver Version, imageUrl string) appsv1.Deployment {
	linterName := deploymentNameFromAttrs(lang, ver)
	labels := map[string]string{
		"version":   string(ver),
		"language":  string(lang),
		"app":       linterName,
		markerLabel: "linter",
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

func (s *MachineManagerPersistentState) linterWeight(lang Language, ver Version) float32 {
	for _, w := range s.Weights {
		if w.Attrs.Language == string(lang) && w.Attrs.Version == string(ver) {
			return w.Weight
		}
	}
	return 0
}

func (s *MachineManagerServer) pushLoadbalancerProportions() {
	for _, lbInfo := range s.State.LoadBalancers {
		select {
		case lbInfo.commandChan <- true:
		default:
		}
	}
}

func (s *MachineManagerServer) storeState() {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	serializedState, err := json.Marshal(&s.State.Persistent.Weights)
	if err != nil {
		panic("error serializing state: " + err.Error())
	}
	updatedMap := apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-manager-config-map",
		},
		Data: map[string]string{
			machine_manager_config_map_data_key: string(serializedState),
		},
	}
	_, err = s.kubernetes.CoreV1().ConfigMaps(apiv1.NamespaceDefault).Update(ctx, &updatedMap, metav1.UpdateOptions{})
	if err != nil {
		panic("error saving state: " + err.Error())
	}
}

func (s *MachineManagerServer) AppendLinter(ctx context.Context, req *pb.AppendLinterRequest) (*pb.LinterResponse, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	deploymentsClient := s.kubernetes.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := deploymentFromLabels(Language(req.Attrs.Language), Version(req.Attrs.Version), req.ImageUrl)
	_, err := deploymentsClient.Create(ctx, &deployment, metav1.CreateOptions{})
	if err != nil {
		log.Println("error appending linter: " + err.Error())
		return &pb.LinterResponse{Code: 1}, nil
	}
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *MachineManagerServer) RemoveLinter(ctx context.Context, req *pb.LinterAttributes) (*pb.LinterResponse, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	lang := Language(req.Language)
	ver := Version(req.Version)
	deploymentsClient := s.kubernetes.AppsV1().Deployments(apiv1.NamespaceDefault)
	deploymentName := deploymentNameFromAttrs(lang, ver)
	s.mut.Lock()
	linterWeight := s.State.Persistent.linterWeight(lang, ver)
	s.mut.Unlock()
	if linterWeight > 0 {
		log.Printf("attempt to remove linter %s of nonzero weight %f\n", deploymentName, linterWeight)
		return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
	}
	err := deploymentsClient.Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		log.Println("error removing linter: " + err.Error())
		return &pb.LinterResponse{Code: 1}, nil
	}
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *MachineManagerServer) SetProportions(ctx context.Context, req *pb.LoadBalancingProportions) (*pb.LinterResponse, error) {
	for _, weight := range req.Weights {
		lang := Language(weight.Attrs.Language)
		ver := Version(weight.Attrs.Version)
		s.mut.Lock()
		hasMachines := s.State.hasMachinesFor(lang, ver)
		s.mut.Unlock()
		if !hasMachines {
			log.Printf("attempt to set nonzero weight of %f for nonexistent linter %s \n", weight.Weight, deploymentNameFromAttrs(lang, ver))
			return &pb.LinterResponse{Code: 1}, nil
		}
	}
	s.mut.Lock()
	s.State.Persistent.Weights = req.Weights
	s.storeState()
	s.mut.Unlock()
	s.pushLoadbalancerProportions()
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *MachineManagerServer) handlePodEvents() {
	for event := range s.eventChan {
		switch event.Type {
		case ADD_WORKER:
			s.mut.Lock()
			s.State.Linters[*event.Name] = WorkerInfo{
				address: *event.Addr,
				lang:    *event.Language,
				ver:     *event.Version,
			}
			s.storeState()
			s.mut.Unlock()
			s.pushLoadbalancerProportions()
		case ADD_LB:
			cmdChan := make(chan bool, 1)
			lbAddr := fmt.Sprintf("%s:%d", string(*event.Addr), load_balancer_port)

			conn, err := grpc.Dial(lbAddr, grpc_opts...)
			if err != nil {
				log.Println("failed to dial %s: %s\n", lbAddr, err)
			}
			log.Printf("connected to: %v", lbAddr)
			client := pb.NewLoadBalancerClient(conn)

			ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
			go func() {
				for _ = range cmdChan {
					s.mut.Lock()
					config := pb.SetConfigRequest{
						Workers: s.State.getWorkers(),
						Weights: s.State.Persistent.Weights,
					}
					s.mut.Unlock()
					for i := 0; i < num_retries; i++ {
						goCtx, cancelGoCtx := context.WithTimeout(ctx, requestTimeout)

						_, err = client.SetConfig(goCtx, &config)
						if err != nil {
							log.Printf("failed to set config on a load balancer %s: %s", event.Name, err)
						}
						cancelGoCtx()
					}
					if err != nil {
						log.Println("retries failed")
					}
				}
			}()
			s.mut.Lock()
			s.State.LoadBalancers[*event.Name] = LBInfo{
				conn:        *conn,
				cancelCtx:   cancelCtx,
				commandChan: cmdChan,
			}
			s.mut.Unlock()
		case REMOVE_POD:
			if lbInfo, isLb := s.State.LoadBalancers[*event.Name]; isLb {
				delete(s.State.LoadBalancers, *event.Name)
				lbInfo.cancelCtx()
				lbInfo.conn.Close()
				close(lbInfo.commandChan)
			} else {
				s.mut.Lock()
				_, isWorker := s.State.Linters[*event.Name]
				if isWorker {
					delete(s.State.Linters, *event.Name)
					s.storeState()
				}
				s.mut.Unlock()
				if isWorker {
					s.pushLoadbalancerProportions()
				}
			}
		default:
			log.Printf("wrong PodEvent type: %d", event.Type)
		}
	}
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

	go log.Fatal(grpcServer.Serve(lis))
	machine_manager.handlePodEvents()
}
