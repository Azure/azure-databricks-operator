package model

import (
	dbmodel "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

//JobsRunsSubmitRequest represents DataBricks run submit request
type JobsRunsSubmitRequest struct {
	RunName      string               `json:"run_name"`
	NewCluster   dbmodel.NewCluster   `json:"new_cluster"`
	Libraries    []dbmodel.Library    `json:"libraries"`
	SparkJarTask dbmodel.SparkJarTask `json:"spark_jar_task"`
}
