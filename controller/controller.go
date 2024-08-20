package local_controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
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

	dp := &appsv1.Deployment{}
	err := r.Client.Get(ctx, request.NamespacedName, dp)
	if errors.IsNotFound(err) {
		log.Error(nil, "not found deploy in cluster", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("err to get deploy:  err=%v", err)
	}

	// TODO: your logic;

	// For example;
	if dp.Annotations == nil {
		dp.Annotations = make(map[string]string)
	}

	if _, ok := dp.Annotations["kube-controller"]; !ok {
		dp.Annotations["kube-controller"] = "test-by-controller"
	}

	// force update
	dp.ResourceVersion = "0"

	err = r.Client.Update(ctx, dp)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("update deploy err=%v", err)
	}

	log.Info("update deploy success", "name", dp.Name, "namespace", dp.Namespace)
	return reconcile.Result{}, nil
}
