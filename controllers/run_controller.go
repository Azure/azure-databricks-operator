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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
)

// RunReconciler reconciles a Run object
type RunReconciler struct {
	client.Client
	Log logr.Logger

	Recorder  record.EventRecorder
	APIClient dbazure.DBClient
}

// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=runs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=runs/status,verbs=get;update;patch

func (r *RunReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("run", req.NamespacedName)

	instance := &databricksv1alpha1.Run{}
	if err := r.Get(context.Background(), req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.IsBeingDeleted() {
		if err := r.handleFinalizer(instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	if !instance.HasFinalizer(databricksv1alpha1.RunFinalizerName) {
		if err := r.addFinalizer(instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Added", "Object finalizer is added")
		return ctrl.Result{}, nil
	}

	if !instance.IsSubmitted() {
		if err := r.submit(instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when submitting run: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Submitted", "Object is submitted")
	}

	if instance.IsSubmitted() {
		if err := r.refresh(instance); err != nil {
			return ctrl.Result{}, fmt.Errorf("error when refreshing run: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Refreshed", "Object is refreshed")
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *RunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databricksv1alpha1.Run{}).
		Complete(r)
}
