package local_manager

import (
	local_cotnroller "deploy-controller/controller"
	"slices"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	kubeExcludeNamespaces = []string{"kube-system", "kube-public"}
)

func init() {
	log.SetLogger(zap.New())
}

func RunManager() error {
	mgrLog := log.Log.WithName("kube-controller")

	// Setup a Manager
	mgrLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		mgrLog.Error(err, "err to set up local manager")
		return err
	}

	c, err := controller.New("deploy-controller", mgr, controller.Options{
		Reconciler:   &local_cotnroller.ReconcilePods{Client: mgr.GetClient()},
		RecoverPanic: func() *bool { t := true; return &t }(),
	})
	if err != nil {
		mgrLog.Error(err, "err to set up local controller")
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &appsv1.Deployment{}, &handler.TypedEnqueueRequestForObject[*appsv1.Deployment]{},
		predicate.TypedFuncs[*appsv1.Deployment]{
			CreateFunc: func(tce event.TypedCreateEvent[*appsv1.Deployment]) bool {

				// 过滤掉指定命令空间
				if slices.Contains(kubeExcludeNamespaces, tce.Object.Namespace) {
					return false
				}

				// 筛选出指定标签的deploy
				if v, ok := tce.Object.Labels["kube-controller"]; ok { // only watch labeled deploy
					return v == "true"
				}
				return false
			}, // only watch create event
			UpdateFunc:  func(tue event.TypedUpdateEvent[*appsv1.Deployment]) bool { return false },
			DeleteFunc:  func(tde event.TypedDeleteEvent[*appsv1.Deployment]) bool { return false },
			GenericFunc: func(tge event.TypedGenericEvent[*appsv1.Deployment]) bool { return false },
		},
		predicate.TypedGenerationChangedPredicate[*appsv1.Deployment]{},
	))
	if err != nil {
		mgrLog.Error(err, "unable to watch deploy")
		return err
	}

	mgrLog.Info("starting local manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		mgrLog.Error(err, "unable to run local manager")
		return err
	}
	return nil

}
