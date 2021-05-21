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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbazure "github.com/polar-rams/databricks-sdk-golang/azure"
)

// SecretScopeReconciler reconciles a SecretScope object
type SecretScopeReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
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
			r.Recorder.Event(instance, corev1.EventTypeWarning, "deleting finalizer", fmt.Sprintf("Failed to delete finalizer: %s", err))
			return reconcile.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	if !instance.HasFinalizer(databricksv1alpha1.SecretScopeFinalizerName) {
		err = r.addFinalizer(instance)
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Adding finalizer", fmt.Sprintf("Failed to add finalizer: %s", err))
			return reconcile.Result{}, fmt.Errorf("error when handling secret scope finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		return ctrl.Result{}, nil
	}

	if !instance.IsSecretAvailable() {
		if err = r.checkSecrets(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Failed", err.Error())
			return ctrl.Result{RequeueAfter: 30 * time.Second}, fmt.Errorf("error when submitting secret scope to the API: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Passed", "Secrets are available")
		return ctrl.Result{}, nil
	}

	if !instance.IsSubmitted() {
		var requeue bool
		requeue, err = r.submit(instance)
		if err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Failed", fmt.Sprintf("Failed to submit object: %s", err))
			if requeue {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, fmt.Errorf("error when submitting secret scope to the API: %v", err)
			}
			return ctrl.Result{}, nil
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Submitted", "Object is submitted")
		return ctrl.Result{}, nil
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
