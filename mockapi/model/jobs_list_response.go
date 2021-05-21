package model

import (
	dbmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
)

// JobsListResponse represents Databricks jobs/list response object
type JobsListResponse struct {
	Jobs []dbmodel.Job `json:"jobs"`
}
