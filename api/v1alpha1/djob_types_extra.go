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

package v1alpha1

import (
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

// JobSettings is similar to dbmodels.JobSettings, the reason it
// exists is because dbmodels.JobSettings doesn't support ExistingClusterName
// ExistingClusterName allows discovering databricks clusters by it's kubernetese object name
type JobSettings struct {
	ExistingClusterID      string                          `json:"existing_cluster_id,omitempty" url:"existing_cluster_id,omitempty"`
	ExistingClusterName    string                          `json:"existing_cluster_name,omitempty" url:"existing_cluster_name,omitempty"`
	NewCluster             *dbmodels.NewCluster            `json:"new_cluster,omitempty" url:"new_cluster,omitempty"`
	NotebookTask           *dbmodels.NotebookTask          `json:"notebook_task,omitempty" url:"notebook_task,omitempty"`
	SparkJarTask           *dbmodels.SparkJarTask          `json:"spark_jar_task,omitempty" url:"spark_jar_task,omitempty"`
	SparkPythonTask        *dbmodels.SparkPythonTask       `json:"spark_python_task,omitempty" url:"spark_python_task,omitempty"`
	SparkSubmitTask        *dbmodels.SparkSubmitTask       `json:"spark_submit_task,omitempty" url:"spark_submit_task,omitempty"`
	Name                   string                          `json:"name,omitempty" url:"name,omitempty"`
	Libraries              []dbmodels.Library              `json:"libraries,omitempty" url:"libraries,omitempty"`
	EmailNotifications     *dbmodels.JobEmailNotifications `json:"email_notifications,omitempty" url:"email_notifications,omitempty"`
	TimeoutSeconds         int32                           `json:"timeout_seconds,omitempty" url:"timeout_seconds,omitempty"`
	MaxRetries             int32                           `json:"max_retries,omitempty" url:"max_retries,omitempty"`
	MinRetryIntervalMillis int32                           `json:"min_retry_interval_millis,omitempty" url:"min_retry_interval_millis,omitempty"`
	RetryOnTimeout         bool                            `json:"retry_on_timeout,omitempty" url:"retry_on_timeout,omitempty"`
	Schedule               *dbmodels.CronSchedule          `json:"schedule,omitempty" url:"schedule,omitempty"`
	MaxConcurrentRuns      int32                           `json:"max_concurrent_runs,omitempty" url:"max_concurrent_runs,omitempty"`
}

// ToK8sJobSettings converts a databricks JobSettings object to k8s JobSettings object.
// It is needed to add ExistingClusterName and follow k8s camleCase naming convention
func ToK8sJobSettings(dbjs *dbmodels.JobSettings) JobSettings {
	var k8sjs JobSettings
	k8sjs.ExistingClusterID = dbjs.ExistingClusterID
	k8sjs.NewCluster = dbjs.NewCluster
	k8sjs.NotebookTask = dbjs.NotebookTask
	k8sjs.SparkJarTask = dbjs.SparkJarTask
	k8sjs.SparkPythonTask = dbjs.SparkPythonTask
	k8sjs.SparkSubmitTask = dbjs.SparkSubmitTask
	k8sjs.Name = dbjs.Name
	k8sjs.Libraries = dbjs.Libraries
	k8sjs.EmailNotifications = dbjs.EmailNotifications
	k8sjs.TimeoutSeconds = dbjs.TimeoutSeconds
	k8sjs.MaxRetries = dbjs.MaxRetries
	k8sjs.MinRetryIntervalMillis = dbjs.MinRetryIntervalMillis
	k8sjs.RetryOnTimeout = dbjs.RetryOnTimeout
	k8sjs.Schedule = dbjs.Schedule
	k8sjs.MaxConcurrentRuns = dbjs.MaxConcurrentRuns
	return k8sjs
}

// ToDatabricksJobSettings converts a k8s JobSettings object to a DataBricks JobSettings object.
// It is needed to add ExistingClusterName and follow k8s camleCase naming convention
func ToDatabricksJobSettings(k8sjs *JobSettings) dbmodels.JobSettings {

	var dbjs dbmodels.JobSettings
	dbjs.ExistingClusterID = k8sjs.ExistingClusterID
	dbjs.NewCluster = k8sjs.NewCluster
	dbjs.NotebookTask = k8sjs.NotebookTask
	dbjs.SparkJarTask = k8sjs.SparkJarTask
	dbjs.SparkPythonTask = k8sjs.SparkPythonTask
	dbjs.SparkSubmitTask = k8sjs.SparkSubmitTask
	dbjs.Name = k8sjs.Name
	dbjs.Libraries = k8sjs.Libraries
	dbjs.EmailNotifications = k8sjs.EmailNotifications
	dbjs.TimeoutSeconds = k8sjs.TimeoutSeconds
	dbjs.MaxRetries = k8sjs.MaxRetries
	dbjs.MinRetryIntervalMillis = k8sjs.MinRetryIntervalMillis
	dbjs.RetryOnTimeout = k8sjs.RetryOnTimeout
	dbjs.Schedule = k8sjs.Schedule
	dbjs.MaxConcurrentRuns = k8sjs.MaxConcurrentRuns
	return dbjs
}
