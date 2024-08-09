package local_ctrl

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ReconcilePods struct {
	Client client.Client
}

var _ reconcile.Reconciler = &ReconcilePods{}

func (r *ReconcilePods) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	pod := &corev1.Pod{}
	err := r.Client.Get(ctx, request.NamespacedName, pod)
	if errors.IsNotFound(err) {
		log.Error(nil, "not found pod in cluster", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("err to get pod:  err=%v", err)
	}

	// TODO;
	// some pod logical
	// take this for example
	_, ok1 := pod.Labels["app"]
	_, ok2 := pod.Labels["version"]
	if !(ok1 && !ok2) {
		return reconcile.Result{}, nil
	}

	pod.Labels["version"] = "stable"
	err = r.Client.Update(ctx, pod)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("update pods err=%v", err)
	}

	return reconcile.Result{}, nil
}
