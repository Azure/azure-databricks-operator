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

package notebookjob

import (
	"context"
	"fmt"
	microsoftv1beta1 "microsoft/azure-databricks-operator/pkg/apis/microsoft/v1beta1"
	"strings"
	"time"

	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileNotebookJob) deleteRunFromDatabricks(runID int64) error {
	if runID == 0 {
		return nil
	}
	err := r.apiClient.Jobs().RunsCancel(runID)
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Second)
	return r.apiClient.Jobs().RunsDelete(runID)
}

func (r *ReconcileNotebookJob) submitRunToDatabricks(instance *microsoftv1beta1.NotebookJob) error {

	// runName
	var runName = instance.ObjectMeta.Name

	// clusterSpec
	instance = instance.LoadDefaultConfig()
	clusterSpec := dbmodels.ClusterSpec{
		NewCluster: &dbmodels.NewCluster{
			SparkEnvVars: &dbmodels.SparkEnvPair{
				Key:   "PYSPARK_PYTHON",
				Value: "/databricks/python3/bin/python3",
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
			clusterSpec.Libraries[i].Pypi.Package = v.Properties["package"]
			clusterSpec.Libraries[i].Pypi.Repo = v.Properties["repo"]
		}
		if v.Type == "maven" {
			clusterSpec.Libraries[i].Maven.Coordinates = v.Properties["coordinates"]
			clusterSpec.Libraries[i].Maven.Repo = v.Properties["repo"]
			// TODO the spec doesn't support array
			// clusterSpec.Libraries[i].Maven.Exclusions = v.Properties["exclusions"]
		}
		if v.Type == "cran" {
			clusterSpec.Libraries[i].Cran.Package = v.Properties["package"]
			clusterSpec.Libraries[i].Cran.Repo = v.Properties["repo"]
		}
	}

	// jobTask
	jobTask := dbmodels.JobTask{
		NotebookTask: &dbmodels.NotebookTask{
			NotebookPath:   instance.Spec.NotebookTask.NotebookPath,
			BaseParameters: make([]dbmodels.ParamPair, len(instance.Spec.NotebookSpec)),
		},
	}
	counter := 0
	for k, v := range instance.Spec.NotebookSpec {
		jobTask.NotebookTask.BaseParameters[counter] = dbmodels.ParamPair{
			Key: k, Value: v,
		}
		counter++
	}

	// timeoutSeconds
	var timeoutSeconds = instance.Spec.TimeoutSeconds

	// scopeSecrents
	var scopeSecrets = make(map[string]string, len(instance.Spec.NotebookSpecSecrets))
	for _, notebookSpecSecret := range instance.Spec.NotebookSpecSecrets {
		secretName := notebookSpecSecret.SecretName
		secret := &v1.Secret{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, secret)
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
	err := r.createSecretScopeWithSecrets(secretScopeName, scopeSecrets)
	if err != nil {
		return err
	}

	// submit run
	runResponse, err := r.apiClient.Jobs().RunsSubmit(runName, clusterSpec, jobTask, int32(timeoutSeconds))
	if err != nil {
		return err
	}

	if runResponse.RunID == 0 {
		return fmt.Errorf("result from API didn't return any values")
	}

	// write information back to instance
	instance.Spec.NotebookTask.RunID = int(runResponse.RunID)
	err = r.Update(context.TODO(), instance)
	if err != nil {
		return fmt.Errorf("error when updating NotebookJob after submitting to API: %v", err)
	}

	r.recorder.Event(instance, "Normal", "Updated", "runID added")
	return nil
}

func (r *ReconcileNotebookJob) createSecretScopeWithSecrets(scope string, secrets map[string]string) error {
	err := r.apiClient.Secrets().CreateSecretScope(scope, "users")
	if err != nil && !strings.Contains(err.Error(), "RESOURCE_ALREADY_EXISTS") {
		return err
	}
	for k, v := range secrets {
		err = r.apiClient.Secrets().PutSecretString(v, scope, k)
		if err != nil {
			return err
		}
	}
	return nil
}
