package local_ctrl

import (
	"context"
	"fmt"
	"strings"

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
	// take this for example, warning: if pod labels can not be match with rs,
	// it will create pod containuerly

	// for example; value=RUN_ON__prod, it will inject all pod env.Name=RUN_ON, env.Vaule=prod
	value, ok := pod.Annotations["env-inject.pod-controller.local"]
	if !ok {
		return reconcile.Result{}, nil
	}

	strs := strings.Split(value, "__")
	if len(strs) != 2 {
		return reconcile.Result{}, nil
	}

	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, corev1.EnvVar{
			Name:  strs[0],
			Value: strs[1],
		})
	}

	err = r.Client.Update(ctx, pod)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("update pods err=%v", err)
	}

	return reconcile.Result{}, nil
}
