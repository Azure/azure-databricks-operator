package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	dbmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
)

// JobRepository is a store for Job instances
type JobRepository struct {
	currentID int64
	jobs      map[int64]dbmodel.Job
	writeLock sync.Mutex
}

// NewJobRepository creates a new JobRepository
func NewJobRepository() *JobRepository {
	return &JobRepository{
		jobs: map[int64]dbmodel.Job{},
	}
}

// GetJob returns the Job with the specified ID or an empty Job
func (r *JobRepository) GetJob(id int64) dbmodel.Job {
	if job, ok := r.jobs[id]; ok {
		return job
	}

	return dbmodel.Job{}
}

// GetJobs returns all Jobs
func (r *JobRepository) GetJobs() model.JobsListResponse {
	arr := []dbmodel.Job{}
	for _, job := range r.jobs {
		arr = append(arr, job)
	}

	result := model.JobsListResponse{
		Jobs: arr,
	}

	return result
}

// CreateJob adds an ID to the specified job and adds it to the collection
func (r *JobRepository) CreateJob(jobCreateRequest dbmodel.JobSettings) int64 {
	newID := atomic.AddInt64(&r.currentID, 1)

	job := dbmodel.Job{
		JobID:       newID,
		Settings:    &jobCreateRequest,
		CreatedTime: makeTimestamp(),
	}

	job.JobID = newID

	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	r.jobs[newID] = job
	return job.JobID
}

// DeleteJob deletes the job with the specified ID
func (r *JobRepository) DeleteJob(id int64) error {
	if _, ok := r.jobs[id]; ok {
		delete(r.jobs, id)
		return nil
	}
	return fmt.Errorf("Could not find Job with id of %d to delete", id)
}
