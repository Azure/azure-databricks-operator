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

package v1

import (
	"fmt"

	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

// DclusterInfo is similar to dbmodels.ClusterInfo, the reason it
// exists is because dbmodels.ClusterInfo has a float field which
// is not supported by Kubernetes API
type DclusterInfo struct {
	NumWorkers             int32                       `json:"num_workers,omitempty" url:"num_workers,omitempty"`
	AutoScale              *dbmodels.AutoScale         `json:"autoscale,omitempty" url:"autoscale,omitempty"`
	ClusterID              string                      `json:"cluster_id,omitempty" url:"cluster_id,omitempty"`
	CreatorUserName        string                      `json:"creator_user_name,omitempty" url:"creator_user_name,omitempty"`
	Driver                 *dbmodels.SparkNode         `json:"driver,omitempty" url:"driver,omitempty"`
	Executors              []dbmodels.SparkNode        `json:"executors,omitempty" url:"executors,omitempty"`
	SparkContextID         int64                       `json:"spark_context_id,omitempty" url:"spark_context_id,omitempty"`
	JdbcPort               int32                       `json:"jdbc_port,omitempty" url:"jdbc_port,omitempty"`
	ClusterName            string                      `json:"cluster_name,omitempty" url:"cluster_name,omitempty"`
	SparkVersion           string                      `json:"spark_version,omitempty" url:"spark_version,omitempty"`
	SparkConf              *dbmodels.SparkConfPair     `json:"spark_conf,omitempty" url:"spark_conf,omitempty"`
	NodeTypeID             string                      `json:"node_type_id,omitempty" url:"node_type_id,omitempty"`
	DriverNodeTypeID       string                      `json:"driver_node_type_id,omitempty" url:"driver_node_type_id,omitempty"`
	ClusterLogConf         *dbmodels.ClusterLogConf    `json:"cluster_log_conf,omitempty" url:"cluster_log_conf,omitempty"`
	InitScripts            []dbmodels.InitScriptInfo   `json:"init_scripts,omitempty" url:"init_scripts,omitempty"`
	SparkEnvVars           map[string]string           `json:"spark_env_vars,omitempty" url:"spark_env_vars,omitempty"`
	AutoterminationMinutes int32                       `json:"autotermination_minutes,omitempty" url:"autotermination_minutes,omitempty"`
	State                  *dbmodels.ClusterState      `json:"state,omitempty" url:"state,omitempty"`
	StateMessage           string                      `json:"state_message,omitempty" url:"state_message,omitempty"`
	StartTime              int64                       `json:"start_time,omitempty" url:"start_time,omitempty"`
	TerminateTime          int64                       `json:"terminate_time,omitempty" url:"terminate_time,omitempty"`
	LastStateLossTime      int64                       `json:"last_state_loss_time,omitempty" url:"last_state_loss_time,omitempty"`
	LastActivityTime       int64                       `json:"last_activity_time,omitempty" url:"last_activity_time,omitempty"`
	ClusterMemoryMb        int64                       `json:"cluster_memory_mb,omitempty" url:"cluster_memory_mb,omitempty"`
	ClusterCores           string                      `json:"cluster_cores,omitempty" url:"cluster_cores,omitempty"`
	DefaultTags            []dbmodels.ClusterTag       `json:"default_tags,omitempty" url:"default_tags,omitempty"`
	ClusterLogStatus       *dbmodels.LogSyncStatus     `json:"cluster_log_status,omitempty" url:"cluster_log_status,omitempty"`
	TerminationReason      *dbmodels.TerminationReason `json:"termination_reason,omitempty" url:"termination_reason,omitempty"`
}

func (i *DclusterInfo) FromDataBricksClusterInfo(ci dbmodels.ClusterInfo) *DclusterInfo {
	i.NumWorkers = ci.NumWorkers
	i.AutoScale = ci.AutoScale
	i.ClusterID = ci.ClusterID
	i.CreatorUserName = ci.CreatorUserName
	i.Driver = ci.Driver
	i.Executors = ci.Executors
	i.SparkContextID = ci.SparkContextID
	i.JdbcPort = ci.JdbcPort
	i.ClusterName = ci.ClusterName
	i.SparkVersion = ci.SparkVersion
	i.SparkConf = ci.SparkConf
	i.NodeTypeID = ci.NodeTypeID
	i.DriverNodeTypeID = ci.DriverNodeTypeID
	i.ClusterLogConf = ci.ClusterLogConf
	i.InitScripts = ci.InitScripts
	i.SparkEnvVars = ci.SparkEnvVars
	i.AutoterminationMinutes = ci.AutoterminationMinutes
	i.State = ci.State
	i.StateMessage = ci.StateMessage
	i.StartTime = ci.StartTime
	i.TerminateTime = ci.TerminateTime
	i.LastStateLossTime = ci.LastStateLossTime
	i.LastActivityTime = ci.LastActivityTime
	i.ClusterMemoryMb = ci.ClusterMemoryMb
	i.ClusterCores = fmt.Sprintf("%v", ci.ClusterCores)
	i.DefaultTags = ci.DefaultTags
	i.ClusterLogStatus = ci.ClusterLogStatus
	i.TerminationReason = ci.TerminationReason
	return i
}
