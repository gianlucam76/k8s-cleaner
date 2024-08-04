/*
Copyright 2023. projectsveltos.io. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"github.com/go-logr/zapr"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller"
	//+kubebuilder:scaffold:imports
)

var (
	setupLog              = ctrl.Log.WithName("setup")
	diagnosticsAddress    string
	insecureDiagnostics   bool
	workers               int
	restConfigQPS         float32
	restConfigBurst       int
	webhookPort           int
	concurrentReconciles  int
	syncPeriod            time.Duration
	jitterWindowInSeconds int
	healthAddr            string
)

// Add RBAC for the authorized diagnostics endpoint.
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
//+kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

func main() {
	scheme, err := initScheme()
	if err != nil {
		os.Exit(1)
	}

	klog.InitFlags(nil)

	initFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	loggerCfg := createLogger()
	zapLogger, err := loggerCfg.Build()
	if err != nil {
		panic(1)
	}

	ctrl.SetLogger(zapr.NewLogger(zapLogger))

	ctx := ctrl.SetupSignalHandler()

	ctrlOptions := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                getDiagnosticsOptions(),
		HealthProbeBindAddress: healthAddr,
		WebhookServer: webhook.NewServer(
			webhook.Options{
				Port: webhookPort,
			}),
		Cache: cache.Options{
			SyncPeriod: &syncPeriod,
		},
	}

	restConfig := ctrl.GetConfigOrDie()
	restConfig.QPS = restConfigQPS
	restConfig.Burst = restConfigBurst
	mgr, err := ctrl.NewManager(restConfig, ctrlOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.CleanerReconciler{
		Client:                mgr.GetClient(),
		Scheme:                mgr.GetScheme(),
		ConcurrentReconciles:  concurrentReconciles,
		JitterWindowInSeconds: jitterWindowInSeconds,
	}).SetupWithManager(ctx, mgr, workers, ctrl.Log.WithName("worker")); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cleaner")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&diagnosticsAddress, "diagnostics-address", ":8443",
		"The address the diagnostics endpoint binds to. Per default metrics are served via https and with"+
			"authentication/authorization. To serve via http and without authentication/authorization set --insecure-diagnostics."+
			"If --insecure-diagnostics is not set the diagnostics endpoint also serves pprof endpoints and an endpoint to change the log level.")

	fs.BoolVar(&insecureDiagnostics, "insecure-diagnostics", false,
		"Enable insecure diagnostics serving. For more details see the description of --diagnostics-address.")

	const defaultWorkers = 5
	fs.IntVar(&workers, "worker-number", defaultWorkers,
		"Number of worker. Workers are used to process cleaner instances in background")

	const defaultJitterWindow = 15
	fs.IntVar(&jitterWindowInSeconds, "jitter-window", defaultJitterWindow,
		"The predefined time interval around a scheduled execution time.")

	const defaultReconcilers = 10
	fs.IntVar(&concurrentReconciles, "concurrent-reconciles", defaultReconcilers,
		"concurrent reconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 10")

	const defautlRestConfigQPS = 40
	fs.Float32Var(&restConfigQPS, "kube-api-qps", defautlRestConfigQPS,
		fmt.Sprintf("Maximum queries per second from the controller client to the Kubernetes API server. Defaults to %d",
			defautlRestConfigQPS))

	const defaultRestConfigBurst = 60
	fs.IntVar(&restConfigBurst, "kube-api-burst", defaultRestConfigBurst,
		fmt.Sprintf("Maximum number of queries that should be allowed in one burst from the controller client to the Kubernetes API server. Default %d",
			defaultRestConfigBurst))

	const defaultWebhookPort = 9443
	fs.IntVar(&webhookPort, "webhook-port", defaultWebhookPort,
		"Webhook Server port")

	const defaultSyncPeriod = 10
	fs.DurationVar(&syncPeriod, "sync-period", defaultSyncPeriod*time.Minute,
		fmt.Sprintf("The minimum interval at which watched resources are reconciled (e.g. 15m). Default: %d minutes",
			defaultSyncPeriod))

	fs.StringVar(&healthAddr, "health-addr", ":9440",
		"The address the health endpoint binds to.")
}

//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;delete

func initScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := appsv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}

	return s, nil
}

func createLogger() *zap.Config {
	loggerCfg := zap.NewProductionConfig()
	const level int = -2
	loggerCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(level))
	loggerCfg.EncoderConfig = zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		TimeKey:      "time",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	return &loggerCfg
}

// getDiagnosticsOptions returns metrics options which can be used to configure a Manager.
func getDiagnosticsOptions() metricsserver.Options {
	// If "--insecure-diagnostics" is set, serve metrics via http
	// and without authentication/authorization.
	if insecureDiagnostics {
		return metricsserver.Options{
			BindAddress:   diagnosticsAddress,
			SecureServing: false,
		}
	}

	// If "--insecure-diagnostics" is not set, serve metrics via https
	// and with authentication/authorization. As the endpoint is protected,
	// we also serve pprof endpoints and an endpoint to change the log level.
	return metricsserver.Options{
		BindAddress:    diagnosticsAddress,
		SecureServing:  true,
		FilterProvider: filters.WithAuthenticationAndAuthorization,
	}
}
