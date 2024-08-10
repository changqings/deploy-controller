package local_mgr

import (
	local_ctrl "deploy-controller/controller"

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
		Reconciler:   &local_ctrl.ReconcilePods{Client: mgr.GetClient()},
		RecoverPanic: func() *bool { t := true; return &t }(),
	})
	if err != nil {
		mgrLog.Error(err, "err to set up local controller")
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &appsv1.Deployment{}, &handler.TypedEnqueueRequestForObject[*appsv1.Deployment]{},
		predicate.TypedFuncs[*appsv1.Deployment]{
			CreateFunc:  func(tce event.TypedCreateEvent[*appsv1.Deployment]) bool { return true }, // only watch create event
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
