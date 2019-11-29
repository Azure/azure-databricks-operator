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
	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DclusterReconciler reconciles a Dcluster object
type DclusterReconciler struct {
	client.Client
	Log logr.Logger

	Recorder  record.EventRecorder
	APIClient dbazure.DBClient
}

// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=dclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databricks.microsoft.com,resources=dclusters/status,verbs=get;update;patch

// Reconcile implements the reconciliation loop for the operator
func (r *DclusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("dcluster", req.NamespacedName)

	instance := &databricksv1alpha1.Dcluster{}

	r.Log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer r.Log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	if err := r.Get(context.Background(), req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.IsBeingDeleted() {
		r.Log.Info(fmt.Sprintf("HandleFinalizer for %v", req.NamespacedName))
		if err := r.handleFinalizer(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "deleting finalizer", fmt.Sprintf("Failed to delete finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	if !instance.HasFinalizer(databricksv1alpha1.DclusterFinalizerName) {
		r.Log.Info(fmt.Sprintf("AddFinalizer for %v", req.NamespacedName))
		if err := r.addFinalizer(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Adding finalizer", fmt.Sprintf("Failed to add finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		return ctrl.Result{}, nil
	}

	if !instance.IsSubmitted() {
		r.Log.Info(fmt.Sprintf("Submit for %v", req.NamespacedName))
		if err := r.submit(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Submitting object", fmt.Sprintf("Failed to submit object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when submitting cluster: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Submitted", "Object is submitted")
		return ctrl.Result{}, nil
	}

	if instance.IsSubmitted() {
		r.Log.Info(fmt.Sprintf("Refresh for %v", req.NamespacedName))
		if err := r.refresh(instance); err != nil {
			r.Recorder.Event(instance, corev1.EventTypeWarning, "Refreshing object", fmt.Sprintf("Failed to refresh object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when refreshing cluster: %v", err)
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Refreshed", "Object is refreshed")
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

const dclusterIndexKey = ".status.cluster_info.cluster_id"

// SetupWithManager adds the controller manager
func (r *DclusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&databricksv1alpha1.Dcluster{}, dclusterIndexKey, func(rawObj runtime.Object) []string {
		dcluster := rawObj.(*databricksv1alpha1.Dcluster)
		if dcluster == nil || dcluster.Status == nil || dcluster.Status.ClusterInfo == nil {
			return nil
		}
		return []string{dcluster.Status.ClusterInfo.ClusterID}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&databricksv1alpha1.Dcluster{}).
		Complete(r)
}
