/*
Copyright 2019 microsoft.

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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
	"github.com/microsoft/azure-databricks-operator/controllers"
	db "github.com/xinsnake/databricks-sdk-golang"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	"k8s.io/apimachinery/pkg/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
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
	kscheme.AddToScheme(scheme)
	databricksv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	apiClient := func() dbazure.DBClient {
		host, token := os.Getenv("DATABRICKS_HOST"), os.Getenv("DATABRICKS_TOKEN")
		if len(host) < 10 && len(token) < 10 {
			err = fmt.Errorf("no valid databricks host / key configured")
			setupLog.Error(err, "unable to start notebookjob controller")
			os.Exit(1)
		}
		var apiClient dbazure.DBClient
		return apiClient.Init(db.DBClientOption{
			Host:  host,
			Token: token,
		})
	}()

	err = (&controllers.NotebookJobReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("NotebookJob"),
		Recorder:  mgr.GetEventRecorderFor("notebookjob-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NotebookJob")
		os.Exit(1)
	}
	err = (&controllers.SecretScopeReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SecretScope"),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretScope")
		os.Exit(1)
	}
	err = (&controllers.DjobReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("Djob"),
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
		Recorder:  mgr.GetEventRecorderFor("run-controller"),
		APIClient: apiClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Run")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
