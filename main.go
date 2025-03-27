package main

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler"

	"helm.sh/helm/v3/pkg/chart/loader"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: set value for the app label")
		os.Exit(1)
	}
	appLabelValue := os.Args[1]

	gvk := schema.GroupVersionKind{
		Group:   "demo.example.com",
		Version: "v1alpha1",
		Kind:    "Nginx",
	}

	selector := "app=" + appLabelValue

	labelSelector, err := metav1.ParseToLabelSelector(selector)
	if err != nil {
		panic(fmt.Sprintf("unable to parse label selector: %s", err))
	}

	eChart, err := loader.LoadDir("helm-charts/nginx")
	if err != nil {
		fmt.Printf("Error reading helm-charts directory: %v\n", err)
	}

	rec, err := reconciler.New(
		reconciler.WithChart(*eChart),
		reconciler.WithGroupVersionKind(gvk),
		reconciler.WithSelector(*labelSelector),
	)
	if err != nil {
		fmt.Printf("Error creating Reconciler: %v\n", err)
		return
	}

	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // disable the metrics server
		},
		LeaderElection:   false,
		LeaderElectionID: "86f835c3.example.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
	}

	if err := rec.SetupWithManager(mgr); err != nil {
		panic(fmt.Sprintf("unable to create reconciler: %s", err))
	}

	setupLog.Info("starting manager with label selector", "selector", selector)
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		fmt.Printf("problem running manager: %v\n", err)
		return
	}
}
