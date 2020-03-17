package model

import (
	dbmodel "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

// JobsListResponse represents Databricks jobs/list response object
type JobsListResponse struct {
	Jobs []dbmodel.Job `json:"jobs"`
}
