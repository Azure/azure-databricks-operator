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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
)

// RunReconciler reconciles a Run object
type RunReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	APIClient dbazure.DBClient
}

// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=runs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=runs/status,verbs=get;update;patch

// Reconcile implements the reconciliation loop for the operator
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
		completed, err := r.handleFinalizer(instance)
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "deleting finalizer", fmt.Sprintf("Failed to delete finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		if completed {
			r.Recorder.Event(instance, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
			return ctrl.Result{}, nil
		}
		// no error but not completed removing the finalizer - requeue
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Deleting", "Pending deletion")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !instance.HasFinalizer(databricksv1alpha1.RunFinalizerName) {
		if err := r.addFinalizer(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Adding finalizer", fmt.Sprintf("Failed to add finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		return ctrl.Result{}, nil
	}

	if !instance.IsSubmitted() {
		if requeue, err := r.submit(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Submitting object", fmt.Sprintf("Failed to submit object: %s", err))
			if requeue {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, fmt.Errorf("error when submitting run: %v", err)
			}
			return ctrl.Result{}, fmt.Errorf("error when submitting run: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Submitted", "Object is submitted")
	}

	if instance.IsSubmitted() {
		if err := r.refresh(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Refreshing object", fmt.Sprintf("Failed to refresh object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when refreshing run: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Refreshed", "Object is refreshed")
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager adds the controller manager
func (r *RunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databricksv1alpha1.Run{}).
		Complete(r)
}
