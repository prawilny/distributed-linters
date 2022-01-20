package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	pb "irio/linter_proto"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	ADD_WORKER               PodEventType = iota
	ADD_LB                   PodEventType = iota
	REMOVE_WORKER            PodEventType = iota
	REMOVE_LB                PodEventType = iota
	PERSISTENT_STATE_UDPATED PodEventType = iota
	DEBOUNCE_TIMER_EXPIRED   PodEventType = iota
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
	kubernetes          *kubernetes.Clientset
	mut                 sync.Mutex
	eventChan           chan *PodEvent
	candidateUpdateChan chan bool
	CandidateState      MachineManagerPersistentState
	PersistentState     MachineManagerPersistentState
	LoadBalancers       map[types.NamespacedName]LBInfo
	Linters             map[types.NamespacedName]WorkerInfo
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
	conn        *grpc.ClientConn
	cancelCtx   context.CancelFunc
	commandChan chan<- bool
}

type PodEvent struct {
	Type     PodEventType
	Name     types.NamespacedName
	Addr     Address
	Version  Version
	Language Language
}

type PodReconciler struct {
	ctrlClient.Client
	mut    sync.Mutex
	pods   map[string]*corev1.Pod
	events chan<- *PodEvent
}

const (
	markerLabel                         = "xD"
	requestTimeout                      = 3 * time.Second
	machine_manager_config_map          = "machine-manager-config-map"
	machine_manager_config_map_data_key = "json"
	num_retries                         = 3
)

var (
	admin_addr  = flag.String("admin-addr", ":10000", "The Admin CLI Listen address (with port)")
	health_addr = flag.String("health-addr", ":60000", "The HTTP healthcheck listen address (with port)")

	warningCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machine_manager_warnings",
		Help: "The total number of warnings issued by the machine manager",
	})
	configPushCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machine_manager_config_pushes",
		Help: "The number (excluding retries) of attempts to propagate configuration to load balancers",
	})
	configPushRetriesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machine_manager_config_push_retries",
		Help: "The number of retries to propagate configuration to load balancers",
	})
	healthcheckCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machine_manager_healthchecks",
		Help: "The total number of received healthcheck requests",
	})
	metricCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machine_manager_metrics",
		Help: "The total number of received metrics requests",
	})
)

func (a *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	pod := &corev1.Pod{}
	err := a.Get(ctx, req.NamespacedName, pod)
	deleted := false
	created := false
	ok := true
	if apierrors.IsNotFound(err) {
		a.mut.Lock()
		if pod, ok = a.pods[req.NamespacedName.String()]; ok {
			delete(a.pods, req.NamespacedName.String())
			deleted = true
		}
		a.mut.Unlock()
	} else if err != nil {
		warningCounter.Inc()
		log.Printf("Error processing event for %s: %s\n", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else {
		if pod.Status.Phase == corev1.PodRunning {
			a.mut.Lock()
			if _, ok = a.pods[req.NamespacedName.String()]; !ok {
				a.pods[req.NamespacedName.String()] = pod
				created = true
			}
			a.mut.Unlock()
		}
	}

	if created {
		log.Printf("Created: pod %s\n", req.NamespacedName)
		switch pod.Labels[markerLabel] {
		case "linter":
			data_port := -1
			for _, p := range pod.Spec.Containers[0].Ports {
				if p.Name == "data" {
					data_port = int(p.ContainerPort)
					break
				}
			}
			if data_port > 0 {
				a.events <- &PodEvent{
					Type:     ADD_WORKER,
					Version:  Version(pod.Labels["version"]),
					Language: Language(pod.Labels["language"]),
					Addr:     (Address)(fmt.Sprintf("%s:%d", pod.Status.PodIP, data_port)),
					Name:     req.NamespacedName,
				}
			} else {
				warningCounter.Inc()
				log.Printf("Warning: created linter pod %s has no 'data' port. Ignoring it.\n")
			}
		case "loadbalancer":
			admin_port := -1
			for _, p := range pod.Spec.Containers[0].Ports {
				if p.Name == "admin" {
					admin_port = int(p.ContainerPort)
					break
				}
			}
			if admin_port > 0 {
				a.events <- &PodEvent{
					Type: ADD_LB,
					Name: req.NamespacedName,
					Addr: (Address)(fmt.Sprintf("%s:%d", pod.Status.PodIP, admin_port)),
				}
			} else {
				warningCounter.Inc()
				log.Printf("Warning: created load balancer pod %s has no 'admin' port. Ignoring it.\n")
			}
		default:
			warningCounter.Inc()
			log.Printf("Warning: created pod %s has unknown type : %s\n", req.NamespacedName, pod.Labels[markerLabel])
		}
	} else if deleted {
		log.Printf("Deleted: pod %s\n", req.NamespacedName)
		if pod.Labels[markerLabel] == "linter" {
			a.events <- &PodEvent{
				Type: REMOVE_WORKER,
				Name: req.NamespacedName,
			}
		} else if pod.Labels[markerLabel] == "loadbalancer" {
			a.events <- &PodEvent{
				Type: REMOVE_LB,
				Name: req.NamespacedName,
			}
		} else {
			warningCounter.Inc()
			log.Printf("Warning: deleted pod %s has unknown type : %s\n", req.NamespacedName, pod.Labels[markerLabel])
		}
	} else {
		log.Printf("Changed: pod %s\n", req.NamespacedName)
	}

	return ctrl.Result{}, nil
}

func (mms *MachineManagerServer) getWorkerListCopy() []*pb.Worker {
	workers := []*pb.Worker{}
	for _, worker := range mms.Linters {
		workers = append(workers, &pb.Worker{
			Address: string(worker.address),
			Attrs: &pb.LinterAttributes{
				Language: string(worker.lang),
				Version:  string(worker.ver),
			},
		})
	}
	return workers
}

func runReconciler(eventChan chan *PodEvent) {
	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Namespace: "default",
	})
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
			pods:   make(map[string]*corev1.Pod),
			events: eventChan,
		})
	if err != nil {
		panic("could not create controller: " + err.Error())
	}

	err = manager.Start(ctrl.SetupSignalHandler())
	if err != nil {
		panic("could not start manager: " + err.Error())
	}

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

	state := MachineManagerPersistentState{
		Weights: []*pb.Weight{},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	get, err := kubernetes.CoreV1().ConfigMaps(apiv1.NamespaceDefault).Get(ctx, machine_manager_config_map, metav1.GetOptions{})
	if err == nil {
		jsonData := get.Data[machine_manager_config_map_data_key]
		err = json.Unmarshal([]byte(jsonData), &state)
		if err != nil {
			panic(err.Error())
		}
	} else if apierrors.IsNotFound(err) {
		serializedState, err := json.Marshal(state)
		if err != nil {
			panic("error serializing state: " + err.Error())
		}
		newMap := apiv1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-manager-config-map",
			},
			Data: map[string]string{
				machine_manager_config_map_data_key: string(serializedState),
			},
		}
		_, err = kubernetes.CoreV1().ConfigMaps(apiv1.NamespaceDefault).Create(ctx, &newMap, metav1.CreateOptions{})
		if err != nil {
			panic(err.Error())
		}
	} else {
		panic(err.Error())
	}

	return MachineManagerServer{
		kubernetes:          kubernetes,
		eventChan:           make(chan *PodEvent, 1),
		candidateUpdateChan: make(chan bool, 1),
		PersistentState:     state,
		LoadBalancers:       make(map[types.NamespacedName]LBInfo),
		Linters:             make(map[types.NamespacedName]WorkerInfo),
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
							Name:  linterName,
							Image: imageUrl,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "health",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 60000,
								},
								{
									Name:          "data",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 20000,
								},
							},
							Args: []string{"-data-addr=:20000", "-health-addr=:60000"},
							LivenessProbe: &apiv1.Probe{
								ProbeHandler: apiv1.ProbeHandler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(60000),
									},
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

func (s *MachineManagerServer) ListVersions(ctx context.Context, req *pb.Language) (*pb.LoadBalancingProportions, error) {
	// We don't have to copy this list, because it's never modified – only replaced.
	return &pb.LoadBalancingProportions{Weights: s.PersistentState.Weights}, nil
}

func (s *MachineManagerServer) AppendLinter(ctx context.Context, req *pb.AppendLinterRequest) (*pb.LinterResponse, error) {
	log.Printf("APPEND LINTER\n")
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	deploymentsClient := s.kubernetes.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := deploymentFromLabels(Language(req.Attrs.Language), Version(req.Attrs.Version), req.ImageUrl)
	_, err := deploymentsClient.Create(ctx, &deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error creating linter deployment %s: %s\n", deployment.ObjectMeta.Name, err)
	}
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *MachineManagerServer) RemoveLinter(ctx context.Context, req *pb.LinterAttributes) (*pb.LinterResponse, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	lang := Language(req.Language)
	ver := Version(req.Version)
	deploymentName := deploymentNameFromAttrs(lang, ver)
	s.mut.Lock()
	linterWeight := s.PersistentState.linterWeight(lang, ver)
	s.mut.Unlock()
	if linterWeight > 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "Error: attempted to remove linter deployment %s of nonzero weight %f\n", deploymentName, linterWeight)
	}
	deploymentsClient := s.kubernetes.AppsV1().Deployments(apiv1.NamespaceDefault)
	err := deploymentsClient.Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.FailedPrecondition, "Error removing linter deployment %s: no such deployment\n", deploymentName)
		} else {
			return nil, status.Errorf(codes.Internal, "Error removing linter deployment %s: %s\n", deploymentName, err)
		}
	}
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *MachineManagerServer) SetProportions(ctx context.Context, req *pb.LoadBalancingProportions) (*pb.LinterResponse, error) {
	s.mut.Lock()
	s.CandidateState.Weights = req.Weights
	s.mut.Unlock()
	select {
	case s.candidateUpdateChan <- true:
	default:
	}
	return &pb.LinterResponse{Code: pb.LinterResponse_SUCCESS}, nil
}

func (s *MachineManagerServer) stateUpdater() {
	for _ = range s.candidateUpdateChan {
		log.Printf("Updating version load proportions\n")
		s.mut.Lock()
		// We don't have to copy this list, because it's never modified – only replaced.
		state := s.CandidateState
		s.mut.Unlock()

		ctx, cancelCtx := context.WithCancel(context.Background())
		defer cancelCtx()

		serializedState, err := json.Marshal(state)
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

		s.mut.Lock()
		s.PersistentState = state
		s.mut.Unlock()

		s.eventChan <- &PodEvent{
			Type: PERSISTENT_STATE_UDPATED,
		}
	}
}

func (s *MachineManagerServer) handlePodEvents() {
	var debounceTimer *time.Timer
	scheduleUpdate := func() {
		log.Printf("Update scheduled\n")
		if debounceTimer == nil {
			debounceTimer = time.AfterFunc(time.Second*1, func() { s.eventChan <- &PodEvent{Type: DEBOUNCE_TIMER_EXPIRED} })
		}
	}
	for event := range s.eventChan {
		event := event
		switch event.Type {
		case DEBOUNCE_TIMER_EXPIRED:
			debounceTimer = nil
			log.Printf("Pushing update to load balancers\n")
			for _, lbInfo := range s.LoadBalancers {
				select {
				case lbInfo.commandChan <- true:
				default:
				}
			}
		case ADD_WORKER:
			log.Printf("Adding linter %s\n", event.Name)
			s.mut.Lock()
			s.Linters[event.Name] = WorkerInfo{
				address: event.Addr,
				lang:    event.Language,
				ver:     event.Version,
			}
			s.mut.Unlock()
			scheduleUpdate()
			log.Printf("Added linter %s\n", event.Name)
		case ADD_LB:
			log.Printf("Adding load balancer %s\n", event.Name)
			cmdChan := make(chan bool, 1)
			lbAddr := string(event.Addr)

			conn, err := grpc.Dial(lbAddr, grpc.WithInsecure())
			if err != nil {
				panic(fmt.Sprintf("Failed to dial %s: %s\n", lbAddr, err))
			}
			log.Printf("Connected to load balancer %s\n", event.Name)
			client := pb.NewLoadBalancerClient(conn)

			lbCtx, cancelLbCtx := context.WithCancel(context.Background())
			go func() {
				for _ = range cmdChan {
					configPushCounter.Inc()
					for {
						s.mut.Lock()
						config := pb.SetConfigRequest{
							Workers: s.getWorkerListCopy(),
							// We don't have to copy this list, because it's never modified – only replaced.
							Weights: s.PersistentState.Weights,
						}
						s.mut.Unlock()
						log.Printf("Pushing new config to load balancer %s\n", event.Name)
						requestCtx, cancelRequestCtx := context.WithTimeout(lbCtx, requestTimeout)
						_, err = client.SetConfig(requestCtx, &config)
						cancelRequestCtx()
						if err != nil {
							switch status.Code(err) {
							case codes.Canceled:
								log.Printf("Disconnected from load balancer %s\n", event.Name, err)
								return
							default:
								configPushRetriesCounter.Inc()
								log.Printf("Failed to push config to load balancer %s: %s. Retrying in 1 second.\n", event.Name, err)
								select {
								case <-lbCtx.Done():
									log.Printf("Disconnected from load balancer %s\n", event.Name, err)
									return
								case <-time.After(1 * time.Second):
								}
							}
						} else {
							log.Printf("Pushed new config to load balancer %s\n", event.Name)
							break
						}
					}
				}
			}()
			s.LoadBalancers[event.Name] = LBInfo{
				conn:        conn,
				cancelCtx:   cancelLbCtx,
				commandChan: cmdChan,
			}
			scheduleUpdate()
			log.Printf("Added load balancer %s\n", event.Name)
		case REMOVE_LB:
			log.Printf("Removing load balancer %s\n", event.Name)
			lbInfo := s.LoadBalancers[event.Name]
			delete(s.LoadBalancers, event.Name)
			lbInfo.cancelCtx()
			lbInfo.conn.Close()
			close(lbInfo.commandChan)
			log.Printf("Removed load balancer %s\n", event.Name)
		case REMOVE_WORKER:
			log.Printf("Removing linter %s\n", event.Name)
			s.mut.Lock()
			delete(s.Linters, event.Name)
			s.mut.Unlock()
			scheduleUpdate()
			log.Printf("Removed linter %s\n", event.Name)
		case PERSISTENT_STATE_UDPATED:
			scheduleUpdate()
			log.Printf("Updated version load proportions\n")
		default:
			warningCounter.Inc()
			log.Printf("Unknown PodEvent type: %d\n", event.Type)
		}
	}
}

func main() {
	flag.Parse()
	log.Printf("Starting machine manager\n")

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
	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	lis, err := net.Listen("tcp", *admin_addr)
	if err != nil {
		log.Fatalf("Failed to listen on address %s: %s\n", *admin_addr, err)
	}
	log.Printf("Listening on %s\n", *admin_addr)

	machine_manager := makeMachineManager()
	go machine_manager.handlePodEvents()
	go machine_manager.stateUpdater()
	go runReconciler(machine_manager.eventChan)

	grpcServer := grpc.NewServer()
	pb.RegisterMachineManagerServer(grpcServer, &machine_manager)
	log.Fatal(grpcServer.Serve(lis))
}
