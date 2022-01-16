package main

import (
	"context"
	"fmt"
	"sync"
	"os"

	//appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/client"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func main() {
    fmt.Printf("Controller: starting\n")
	var log = ctrl.Log.WithName("builder-examples")

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

    pred, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{metav1.LabelSelectorRequirement{Key: "xd", Operator: metav1.LabelSelectorOpExists}}})
    if err != nil {
		log.Error(err, "could not create predicate")
		os.Exit(1)
    }

	err = ctrl.
		NewControllerManagedBy(manager).
		For(&corev1.Pod{}, builder.WithPredicates(pred)).
		Complete(&PodReconciler{
            Client: manager.GetClient(),
            pods: make(map[string]podInfo),
        })
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

type podInfo struct {
    role string
    address string
}

type PodReconciler struct {
	client.Client
    mut sync.Mutex
    pods map[string]podInfo
}

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
        }
        a.mut.Unlock()
    } else if err != nil {
        fmt.Printf("Error processing event for %s: %s\n", req.NamespacedName, err)
		return ctrl.Result{}, err
	} else {
        if pod.Status.Phase == corev1.PodRunning {
            a.mut.Lock()
            if _, ok = a.pods[req.NamespacedName.String()]; !ok {
                a.pods[req.NamespacedName.String()] = podInfo{
                    address: pod.Status.PodIP,
                    role: pod.Labels["xd"],
                }
                info = a.pods[req.NamespacedName.String()]
                created = true
            }
            a.mut.Unlock()
        }
    }

    if created {
        fmt.Printf("Created: %s %v\n", req.NamespacedName, info)
    } else if deleted {
        fmt.Printf("Deleted: %s %v\n", req.NamespacedName, info)
    } else {
        fmt.Printf("Changed: %s\n", req.NamespacedName)
    }

	return ctrl.Result{}, nil
}
