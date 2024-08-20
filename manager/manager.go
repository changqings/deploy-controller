package local_manager

import (
	local_cotnroller "deploy-controller/controller"
	"flag"
	"os"
	"slices"

	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsServer "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	kubeExcludeNamespaces = []string{"kube-system", "kube-public"}
	logTimeLayout         = "2006-01-02-15:04:05.000-MST"
)

func init() {
	log.SetLogger(zap.New())
}

func RunManager() error {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var leaderNs string

	flag.StringVar(&leaderNs, "leader-namespace", "default", "leader namespace")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.TimeEncoderOfLayout(logTimeLayout),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	mgrLog := log.Log.WithName("kube-controller")

	// Setup a Manager
	mgrLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Metrics: metricsServer.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderNs,
		LeaderElectionID:        "2408201547.some-controll.cn",
	})
	if err != nil {
		mgrLog.Error(err, "err to set up local manager")
		return err
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		mgrLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		mgrLog.Error(err, "unable to set up ready check")
		os.Exit(1)
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
				if v, ok := tce.Object.Labels["kube-controller"]; ok {
					return v == "true"
				}
				return false
			}, // watch create event
			UpdateFunc: func(tue event.TypedUpdateEvent[*appsv1.Deployment]) bool {
				// 过滤掉指定命令空间
				if slices.Contains(kubeExcludeNamespaces, tue.ObjectOld.Namespace) {
					return false
				}

				// 筛选出指定标签的deploy
				if v, ok := tue.ObjectNew.Labels["kube-controller"]; ok {
					return v == "true"
				}

				return false
			},
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
