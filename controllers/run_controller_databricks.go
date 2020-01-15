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
	"reflect"
	"strings"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	"github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *RunReconciler) submit(instance *databricksv1alpha1.Run) (bool, error) {
	r.Log.Info(fmt.Sprintf("Submitting run %s", instance.GetName()))

	var run *dbmodels.Run
	var err error
	var requeue bool

	instance.Spec.RunName = instance.GetName()

	// If the run is not linked to a job, submit using RunsSubmit,
	// otherwise submit it as RunNow under the job, and make the
	// job the owner of the run
	if instance.Spec.JobName != "" {
		run, requeue, err = r.runUsingRunNow(instance)
		if requeue {
			return true, err
		}
	} else {
		run, err = r.runUsingRunsSubmit(instance)
	}

	if err != nil {
		return false, err
	}

	// update the run state now, in case the RunsGetOutput call below fails
	var pendingState dbmodels.RunLifeCycleState = dbmodels.RunLifeCycleStatePending
	run.State = &dbmodels.RunState{
		LifeCycleState: &pendingState,
	}
	instance.Status = &azure.JobsRunsGetOutputResponse{
		Metadata: *run,
	}

	err = r.Update(context.Background(), instance)
	if err != nil {
		return false, err
	}

	runOutput, err := r.getRunOutput(run.RunID)

	if err != nil {
		return false, err
	}

	instance.Status = &runOutput
	return false, r.Update(context.Background(), instance)
}

func (r *RunReconciler) refresh(instance *databricksv1alpha1.Run) error {
	r.Log.Info(fmt.Sprintf("Refreshing run %s", instance.GetName()))

	runID := instance.Status.Metadata.RunID

	runOutput, err := r.getRunOutput(runID)
	if err != nil {
		return err
	}

	err = r.Get(context.Background(), types.NamespacedName{
		Name:      instance.GetName(),
		Namespace: instance.GetNamespace(),
	}, instance)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(instance.Status, &runOutput) {
		return nil
	}

	instance.Status = &runOutput
	return r.Update(context.Background(), instance)
}

// delete attempts to cancel and delete a run. Returns bool indicating if complete (safe to retry if not and no error) and an error
func (r *RunReconciler) delete(instance *databricksv1alpha1.Run) (bool, error) {
	r.Log.Info(fmt.Sprintf("Deleting run %s", instance.GetName()))

	if instance.Status == nil {
		return true, nil
	}

	runID := instance.Status.Metadata.RunID

	// Check if the run exists before trying to delete it
	run, err := r.getRun(runID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return true, nil
		}
		return false, err
	}

	if run.State.ResultState == nil {
		// We will not check for error when cancelling a job,
		// if it fails just let it be
		execution := NewExecution("runs", "cancel")
		err := r.APIClient.Jobs().RunsCancel(runID)
		execution.Finish(err)
		return false, nil // no error, but indicate not completed to trigger a requeue to delete once cancelled
	}
	// job has reached a terminated state
	execution := NewExecution("runs", "delete")
	err = r.APIClient.Jobs().RunsDelete(runID)
	execution.Finish(err)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *RunReconciler) runUsingRunNow(instance *databricksv1alpha1.Run) (*dbmodels.Run, bool, error) {

	runParameters := dbmodels.RunParameters{
		JarParams:         instance.Spec.JarParams,
		NotebookParams:    instance.Spec.NotebookParams,
		PythonParams:      instance.Spec.PythonParams,
		SparkSubmitParams: instance.Spec.SparkSubmitParams,
	}

	// Here we set the owner attribute
	k8sJobNamespacedName := types.NamespacedName{Namespace: instance.GetNamespace(), Name: instance.Spec.JobName}
	var k8sJob databricksv1alpha1.Djob
	if err := r.Client.Get(context.Background(), k8sJobNamespacedName, &k8sJob); err != nil {
		return nil, false, err
	}

	instance.ObjectMeta.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "v1alpha1", // TODO should this be a referenced value?
			Kind:       "Djob",     // TODO should this be a referenced value?
			Name:       k8sJob.GetName(),
			UID:        k8sJob.GetUID(),
		},
	})

	if !k8sJob.IsSubmitted() {
		return nil, true, fmt.Errorf("Run references Djob that is not yet submitted")
	}

	execution := NewExecution("runs", "run_now")
	run, err := r.APIClient.Jobs().RunNow(k8sJob.Status.JobStatus.JobID, runParameters)
	execution.Finish(err)
	return &run, false, err
}

func (r *RunReconciler) runUsingRunsSubmit(instance *databricksv1alpha1.Run) (*dbmodels.Run, error) {

	//Check if dbricks run is set to run on exisiting dbricks cluster
	//Get exisiting dbricks cluster by cluster name and set ExistingClusterID or
	//Get exisiting dbricks cluster by cluster id
	var ownerInstance databricksv1alpha1.Dcluster
	if len(instance.Spec.ClusterSpec.ExistingClusterName) > 0 {
		dClusterNamespacedName := types.NamespacedName{Name: instance.Spec.ClusterSpec.ExistingClusterName, Namespace: instance.Namespace}
		err := r.Get(context.Background(), dClusterNamespacedName, &ownerInstance)
		if err != nil {
			return nil, err
		}
		if (ownerInstance.Status != nil) && (ownerInstance.Status.ClusterInfo != nil) && len(ownerInstance.Status.ClusterInfo.ClusterID) > 0 {
			instance.Spec.ClusterSpec.ExistingClusterID = ownerInstance.Status.ClusterInfo.ClusterID
		} else {
			return nil, fmt.Errorf("failed to get ClusterID of %v", instance.Spec.ExistingClusterName)
		}
	} else if len(instance.Spec.ClusterSpec.ExistingClusterID) > 0 {
		var dclusters databricksv1alpha1.DclusterList
		err := r.List(context.Background(), &dclusters, client.InNamespace(instance.Namespace), client.MatchingFields{dclusterIndexKey: instance.Spec.ClusterSpec.ExistingClusterID})
		if err != nil {
			return nil, err
		}
		if len(dclusters.Items) == 1 {
			ownerInstance = dclusters.Items[0]
		} else {
			return nil, fmt.Errorf("failed to get ClusterID of %v", instance.Spec.ExistingClusterID)
		}
	}
	//Set Exisiting cluster as Owner of Run
	if &ownerInstance != nil && len(ownerInstance.APIVersion) > 0 && len(ownerInstance.Kind) > 0 && len(ownerInstance.GetName()) > 0 {
		references := []metav1.OwnerReference{
			{
				APIVersion: ownerInstance.APIVersion,
				Kind:       ownerInstance.Kind,
				Name:       ownerInstance.GetName(),
				UID:        ownerInstance.GetUID(),
			},
		}
		instance.ObjectMeta.SetOwnerReferences(references)
	}

	clusterSpec := dbmodels.ClusterSpec{
		NewCluster:        instance.Spec.NewCluster,
		ExistingClusterID: instance.Spec.ExistingClusterID,
		Libraries:         instance.Spec.Libraries,
	}
	jobTask := dbmodels.JobTask{
		NotebookTask:    instance.Spec.NotebookTask,
		SparkJarTask:    instance.Spec.SparkJarTask,
		SparkPythonTask: instance.Spec.SparkPythonTask,
		SparkSubmitTask: instance.Spec.SparkSubmitTask,
	}

	execution := NewExecution("runs", "run_submit")
	run, err := r.APIClient.Jobs().RunsSubmit(instance.Spec.RunName, clusterSpec, jobTask, instance.Spec.TimeoutSeconds)
	execution.Finish(err)
	return &run, err
}

func (r *RunReconciler) getRun(runID int64) (dbmodels.Run, error) {
	execution := NewExecution("runs", "get")
	runOutput, err := r.APIClient.Jobs().RunsGet(runID)
	execution.Finish(err)

	return runOutput, err
}

func (r *RunReconciler) getRunOutput(runID int64) (azure.JobsRunsGetOutputResponse, error) {
	execution := NewExecution("runs", "run_get_output")
	runOutput, err := r.APIClient.Jobs().RunsGetOutput(runID)
	execution.Finish(err)

	return runOutput, err
}
