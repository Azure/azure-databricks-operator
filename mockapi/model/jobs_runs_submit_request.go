package model

import (
	dbjobsmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
	dblibmodel "github.com/polar-rams/databricks-sdk-golang/azure/libraries/models"
)

//JobsRunsSubmitRequest represents DataBricks run submit request
type JobsRunsSubmitRequest struct {
	RunName      string                   `json:"run_name"`
	NewCluster   dbjobsmodel.NewCluster   `json:"new_cluster"`
	Libraries    []dblibmodel.Library     `json:"libraries"`
	SparkJarTask dbjobsmodel.SparkJarTask `json:"spark_jar_task"`
}
