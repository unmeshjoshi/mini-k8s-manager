package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clusterv1alpha1 "github.com/unmeshjoshi/mini-k8s-manager/pkg/api/v1alpha1"
	"github.com/unmeshjoshi/mini-k8s-manager/pkg/controllers"
	"github.com/unmeshjoshi/mini-k8s-manager/pkg/providers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	// Register our types with the scheme
	//_ = scheme.New(scheme)
	_ = clusterv1alpha1.AddToScheme(scheme)
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		//MetricsBindAddress:     metricsAddr,
		//Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "mini-k8s-manager.mini-k8s.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create Docker provider
	provider, err := providers.NewDockerProvider(&clusterv1alpha1.DockerProviderConfig{
		Spec: clusterv1alpha1.DockerProviderConfigSpec{
			Network: clusterv1alpha1.NetworkConfig{
				CIDR:          "10.10.0.0/16",
				SubnetMask:    24,
				ExposedPorts:  []int32{6443},
				EnableIPv6:    false,
				DNSNameserver: "8.8.8.8",
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to create Docker provider")
		os.Exit(1)
	}

	// Set up the cluster controller
	if err = (&controllers.ClusterReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Provider: provider,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
