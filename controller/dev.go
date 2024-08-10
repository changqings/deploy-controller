package local_ctrl

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
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
	// for example;

	// for safe only run on annotation deploy,with local-controller="true"
	if dp.Annotations["local-controller"] != "true" {
		return reconcile.Result{}, nil
	}

	// inject all container env RUN_ON__prod, use double downline as split char
	nameValue, ok := dp.Annotations["env.inject.local"]
	if !ok {
		return reconcile.Result{}, nil
	}
	strs := strings.Split(nameValue, "__")

	if len(strs) != 2 {
		log.Error(nil, "env.inject.local format error", "name", nameValue)
		return reconcile.Result{}, nil
	}

	name, value := strs[0], strs[1]

	for i := range dp.Spec.Template.Spec.Containers {
		dp.Spec.Template.Spec.Containers[i].Env = append(dp.Spec.Template.Spec.Containers[i].Env, corev1.EnvVar{Name: name, Value: value})
	}
	dp.ResourceVersion = "0"

	err = r.Client.Update(ctx, dp)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("update deploy err=%v", err)
	}

	log.Info("update deploy success", "name", dp.Name, "namespace", dp.Namespace)
	return reconcile.Result{}, nil
}
