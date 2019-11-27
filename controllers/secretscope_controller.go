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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	corev1 "k8s.io/api/core/v1"
)

// SecretScopeReconciler reconciles a SecretScope object
type SecretScopeReconciler struct {
	client.Client
	Log logr.Logger

	Recorder  record.EventRecorder
	APIClient dbazure.DBClient
}

// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=secretscopes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=secretscopes/status,verbs=get;update;patch

// Reconcile implements the reconciliation loop for the operator
func (r *SecretScopeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("secretscope", req.NamespacedName)

	// your logic here
	instance := &databricksv1alpha1.SecretScope{}
	err := r.Get(context.Background(), req.NamespacedName, instance)

	r.Log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer r.Log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.IsBeingDeleted() {
		err = r.handleFinalizer(instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	if !instance.HasFinalizer(databricksv1alpha1.SecretScopeFinalizerName) {
		err = r.addFinalizer(instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error when handling secret scope finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		return ctrl.Result{}, nil
	}

	if !instance.IsVerified() {
		if err = r.verifyWorkspace(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Failed", err.Error())
			return ctrl.Result{}, nil
		}
	}

	if !instance.IsSecretAvailable() {
		if err = r.checkSecrets(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Failed", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, fmt.Errorf("error when submitting secret scope to the API: %v", err)
		}
	}

	if !instance.IsSubmitted() {
		if err = r.submit(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Failed", err.Error())
			return ctrl.Result{}, fmt.Errorf("error when submitting secret scope to the API: %v", err)
		}
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Object has completed")
	return ctrl.Result{}, nil
}

// SetupWithManager adds the controller manager
func (r *SecretScopeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databricksv1alpha1.SecretScope{}).
		Complete(r)
}
