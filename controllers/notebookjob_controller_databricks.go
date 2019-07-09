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

	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *NotebookJobReconciler) deleteRunFromDatabricks(instance *databricksv1.NotebookJob) error {
	// cancel and delete the job
	if instance.Status.Run != nil {
		runID := int64(instance.Status.Run.RunID)
		_, err := r.APIClient.Jobs().RunsGet(runID)
		if err == nil {
			err = r.APIClient.Jobs().RunsCancel(runID)
			if err != nil {
				return err
			}
			time.Sleep(15 * time.Second)
			err = r.APIClient.Jobs().RunsDelete(runID)
			if err != nil {
				return err
			}
		} else {
			if !strings.Contains(err.Error(), "does not exist") {
				return err
			}
		}
	}

	// delete the scope
	secretScopeName := fmt.Sprintf("%s_scope", instance.ObjectMeta.Name)
	err := r.APIClient.Secrets().DeleteSecretScope(secretScopeName)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		return err
	}
	return nil
}

func (r *NotebookJobReconciler) submitRunToDatabricks(instance *databricksv1.NotebookJob) error {

	// runName
	var runName = instance.ObjectMeta.Name

	// clusterSpec
	instance = instance.LoadDefaultConfig()
	clusterSpec := dbmodels.ClusterSpec{
		NewCluster: &dbmodels.NewCluster{
			SparkEnvVars: map[string]string{
				"PYSPARK_PYTHON": "/databricks/python3/bin/python3",
			},
		},
	}
	clusterSpec.NewCluster.SparkVersion = instance.Spec.ClusterSpec.SparkVersion
	clusterSpec.NewCluster.NodeTypeID = instance.Spec.ClusterSpec.NodeTypeId
	clusterSpec.NewCluster.NumWorkers = int32(instance.Spec.ClusterSpec.NumWorkers)
	clusterSpec.Libraries = make([]dbmodels.Library, len(instance.Spec.NotebookAdditionalLibraries))
	for i, v := range instance.Spec.NotebookAdditionalLibraries {
		if v.Type == "jar" {
			clusterSpec.Libraries[i].Jar = v.Properties["path"]
		}
		if v.Type == "egg" {
			clusterSpec.Libraries[i].Egg = v.Properties["path"]
		}
		if v.Type == "whl" {
			clusterSpec.Libraries[i].Whl = v.Properties["path"]
		}
		if v.Type == "pypi" {
			clusterSpec.Libraries[i].Pypi = &dbmodels.PythonPyPiLibrary{
				Package: v.Properties["package"],
				Repo:    v.Properties["repo"],
			}
		}
		if v.Type == "maven" {
			clusterSpec.Libraries[i].Maven = &dbmodels.MavenLibrary{
				Coordinates: v.Properties["coordinates"],
				Repo:        v.Properties["repo"],
				// TODO: current exlusions is not a map
				// Exclusions:  v.Properties["exclusions"],
			}
		}
		if v.Type == "cran" {
			clusterSpec.Libraries[i].Cran = &dbmodels.RCranLibrary{
				Package: v.Properties["package"],
				Repo:    v.Properties["repo"],
			}
		}
	}

	// jobTask
	jobTask := dbmodels.JobTask{
		NotebookTask: &dbmodels.NotebookTask{
			NotebookPath:   instance.Spec.NotebookTask.NotebookPath,
			BaseParameters: instance.Spec.NotebookSpec,
		},
	}

	// timeoutSeconds
	var timeoutSeconds = instance.Spec.TimeoutSeconds

	// scopeSecrents
	var scopeSecrets = make(map[string]string, len(instance.Spec.NotebookSpecSecrets))
	for _, notebookSpecSecret := range instance.Spec.NotebookSpecSecrets {
		secretName := notebookSpecSecret.SecretName
		secret := &v1.Secret{}
		err := r.Get(context.Background(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, secret)
		if err != nil {
			return err
		}
		for _, mapping := range notebookSpecSecret.Mapping {
			secretvalue := secret.Data[mapping.SecretKey]
			tempkey := mapping.OutputKey
			scopeSecrets[tempkey] = fmt.Sprintf("%s", secretvalue)
		}
	}

	secretScopeName := runName + "_scope"
	jobTask.NotebookTask.BaseParameters["secret_scope"] = secretScopeName
	r.Log.Info(fmt.Sprintf("Creating secret scope %s with %d secrets", secretScopeName, len(scopeSecrets)))
	err := r.createSecretScopeWithSecrets(secretScopeName, scopeSecrets)
	if err != nil {
		return err
	}

	// submit run
	r.Log.Info("Submitting run " + runName)
	run, err := r.APIClient.Jobs().RunsSubmit(runName, clusterSpec, jobTask, int32(timeoutSeconds))
	if err != nil {
		return err
	}

	if run.RunID == 0 {
		return fmt.Errorf("result from API didn't return any values")
	}

	// write information back to instance
	instance.Status.Run = &run
	err = r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("error when updating NotebookJob after submitting to API: %v", err)
	}

	r.Recorder.Event(instance, "Normal", "Updated", "runID added")
	return nil
}

func (r *NotebookJobReconciler) refreshDatabricksRun(instance *databricksv1.NotebookJob) error {
	r.Log.Info(fmt.Sprintf("Refreshing Databricks run_id %v", instance.Status.Run.RunID))
	runID := instance.Status.Run.RunID
	run, err := r.APIClient.Jobs().RunsGet(int64(runID))
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(instance.Status.Run, &run) {
		instance.Status.Run = &run
		err = r.Update(context.Background(), instance)
		if err != nil {
			return fmt.Errorf("error when updating NotebookJob: %v", err)
		}
		r.Recorder.Event(instance, "Normal", "Updated", "run status updated")
	}
	return nil
}

func (r *NotebookJobReconciler) createSecretScopeWithSecrets(scope string, secrets map[string]string) error {
	err := r.APIClient.Secrets().CreateSecretScope(scope, "users")
	if err != nil && !strings.Contains(err.Error(), "RESOURCE_ALREADY_EXISTS") {
		return err
	}
	for k, v := range secrets {
		err = r.APIClient.Secrets().PutSecretString(v, scope, k)
		if err != nil {
			return err
		}
	}
	return nil
}
