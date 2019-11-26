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
	"time"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	"github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func (r *RunReconciler) submit(instance *databricksv1alpha1.Run) error {
	r.Log.Info(fmt.Sprintf("Submitting run %s", instance.GetName()))

	var run *dbmodels.Run
	var err error

	instance.Spec.RunName = instance.GetName()

	// If the run is not linked to a job, submit using RunsSubmit,
	// otherwise submit it as RunNow under the job, and make the
	// job the owner of the run
	if instance.Spec.JobName != "" {
		run, err = r.runUsingRunNow(instance)
	} else {
		run, err = r.runUsingRunsSubmit(instance)
	}

	if err != nil {
		return err
	}

	runOutput, err := r.getRunOutput(run.RunID)
	if err != nil {
		return err
	}

	instance.Status = &runOutput
	return r.Update(context.Background(), instance)
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

func (r *RunReconciler) delete(instance *databricksv1alpha1.Run) error {
	r.Log.Info(fmt.Sprintf("Deleting run %s", instance.GetName()))

	if instance.Status == nil {
		return nil
	}

	runID := instance.Status.Metadata.RunID

	// Check if the run exists before trying to delete it
	if _, err := r.getRun(runID); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return err
	}

	// We will not check for error when cancelling a job,
	// if it fails just let it be
	r.APIClient.Jobs().RunsCancel(runID) //nolint:errcheck

	// It takes time for DataBricks to cancel a run
	time.Sleep(15 * time.Second)

	return trackExecutionTime(runDeleteDuration, func() error {
		return r.APIClient.Jobs().RunsDelete(runID)
	})
}

func (r *RunReconciler) runUsingRunNow(instance *databricksv1alpha1.Run) (*dbmodels.Run, error) {
	defer trackMillisecondsTaken(time.Now(), runNowDuration)
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
		return nil, err
	}

	instance.ObjectMeta.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "v1alpha1", // TODO should this be a referenced value?
			Kind:       "Djob",     // TODO should this be a referenced value?
			Name:       k8sJob.GetName(),
			UID:        k8sJob.GetUID(),
		},
	})

	run, err := r.APIClient.Jobs().RunNow(k8sJob.Status.JobStatus.JobID, runParameters)
	trackSuccessFailure(err, runNowSuccess, runNowFailure)
	return &run, err
}

func (r *RunReconciler) runUsingRunsSubmit(instance *databricksv1alpha1.Run) (*dbmodels.Run, error) {
	defer trackMillisecondsTaken(time.Now(), runSubmitDuration)

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

	run, err := r.APIClient.Jobs().RunsSubmit(instance.Spec.RunName, clusterSpec, jobTask, instance.Spec.TimeoutSeconds)
	trackSuccessFailure(err, runSubmitSuccess, runSubmitFailure)
	return &run, err
}

func (r *RunReconciler) getRun(runID int64) (dbmodels.Run, error) {
	defer trackMillisecondsTaken(time.Now(), runGetDuration)

	runOutput, err := r.APIClient.Jobs().RunsGet(runID)

	trackSuccessFailure(err, runGetSuccess, runGetFailure)

	return runOutput, err
}

func (r *RunReconciler) getRunOutput(runID int64) (azure.JobsRunsGetOutputResponse, error) {
	defer trackMillisecondsTaken(time.Now(), runGetOutputDuration)

	runOutput, err := r.APIClient.Jobs().RunsGetOutput(runID)

	trackSuccessFailure(err, runGetOutputSuccess, runGetOutputFailure)

	return runOutput, err
}
