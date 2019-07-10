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

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func (r *RunReconciler) submitRunToDatabricks(instance *databricksv1.Run) error {
	r.Log.Info("Submitting job " + instance.GetName())

	var runOutput dbazure.JobsRunsGetOutputResponse
	var run dbmodels.Run
	var err error

	var k8sJob databricksv1.Djob

	instance.Spec.RunName = instance.GetName()
	if instance.Spec.JobName != "" {
		// run existing job
		runParameters := dbmodels.RunParameters{
			JarParams:         instance.Spec.JarParams,
			NotebookParams:    instance.Spec.NotebookParams,
			PythonParams:      instance.Spec.PythonParams,
			SparkSubmitParams: instance.Spec.SparkSubmitParams,
		}

		k8sJobNamespacedName := types.NamespacedName{Namespace: instance.GetNamespace(), Name: instance.Spec.JobName}
		err = r.Client.Get(context.Background(), k8sJobNamespacedName, &k8sJob)
		if err != nil {
			return err
		}

		run, err = r.APIClient.Jobs().RunNow(k8sJob.Status.JobID, runParameters)
		if err != nil {
			return err
		}

		runOutput, err = r.APIClient.Jobs().RunsGetOutput(run.RunID)
		if err != nil {
			return err
		}

		instance.ObjectMeta.SetOwnerReferences([]metav1.OwnerReference{
			metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "Djob",
				Name:       k8sJob.GetName(),
				UID:        k8sJob.GetUID(),
			},
		})
	} else {
		// run directly
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

		runResp, err := r.APIClient.Jobs().RunsSubmit(instance.Spec.RunName, clusterSpec, jobTask, instance.Spec.TimeoutSeconds)
		if err != nil {
			return err
		}

		runOutput, err = r.APIClient.Jobs().RunsGetOutput(runResp.RunID)
		if err != nil {
			return err
		}
	}

	instance.Status = &runOutput
	err = r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("error when updating run after submitting to API: %v", err)
	}

	r.Recorder.Event(instance, "Normal", "Updated", "run submitted")
	return nil
}

func (r *RunReconciler) refreshDatabricksRun(instance *databricksv1.Run) error {
	r.Log.Info("Refreshing run " + instance.GetName())

	runID := instance.Status.Metadata.RunID
	runOutput, err := r.APIClient.Jobs().RunsGetOutput(runID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(instance.Status, &runOutput) {
		instance.Status = &runOutput
		err = r.Update(context.Background(), instance)
		if err != nil {
			return fmt.Errorf("error when updating run: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Updated", "run status updated")
	}
	return nil
}

func (r *RunReconciler) deleteRunFromDatabricks(instance *databricksv1.Run) error {
	r.Log.Info("Deleting run " + instance.GetName())

	if instance.Status == nil {
		return nil
	}
	runID := instance.Status.Metadata.RunID
	_, err := r.APIClient.Jobs().RunsGet(runID)

	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return nil
	}

	if err == nil {
		err = r.APIClient.Jobs().RunsCancel(runID)
		time.Sleep(15 * time.Second)
	}

	if err == nil {
		err = r.APIClient.Jobs().RunsDelete(runID)
	}
	return err
}
