/*
The MIT License (MIT)

Copyright (c) 2019  Microsoft

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	"github.com/microsoft/azure-databricks-operator/controllers"
	db "github.com/polar-rams/databricks-sdk-golang"
	dbazure "github.com/polar-rams/databricks-sdk-golang/azure"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = databricksv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	apiClient := func() dbazure.DBClient {
		host, token := os.Getenv("DATABRICKS_HOST"), os.Getenv("DATABRICKS_TOKEN")
		if len(host) < 10 && len(token) < 10 {
			err = fmt.Errorf("no valid databricks host / key configured")
			setupLog.Error(err, "unable to initialize databricks api client")
			os.Exit(1)
		}

		opt := db.NewDBClientOption("", "", host, token, nil, false, 0)
		return *(dbazure.NewDBClient(opt))
	}()

	err = (&controllers.SecretScopeReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("SecretScope"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("secretscope-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretScope")
		os.Exit(1)
	}
	err = (&controllers.DjobReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Djob"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("djob-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Djob")
		os.Exit(1)
	}
	err = (&controllers.RunReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Run"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("run-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Run")
		os.Exit(1)
	}
	err = (&controllers.DclusterReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Dcluster"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("dcluster-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Dcluster")
		os.Exit(1)
	}
	err = (&controllers.DbfsBlockReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("DbfsBlock"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("dbfsblock-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DbfsBlock")
		os.Exit(1)
	}
	err = (&controllers.WorkspaceItemReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("WorkspaceItem"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("workspaceitem-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WorkspaceItem")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
