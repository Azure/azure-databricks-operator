package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/microsoft/azure-databricks-operator/mockapi/repository"
	dbmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
)

//CreateJob handles the job create endpoint
func CreateJob(j *repository.JobRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: CreateJob - starting\n", requestID)

		var job dbmodel.JobSettings
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("RequestID:%6d: CreateJob - Error reading the body: %v", requestID, err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &job); err != nil {
			log.Printf("RequestID:%6d: CreateJob - Error parsing the body: %v", requestID, err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		response := dbmodel.Job{
			JobID: j.CreateJob(job),
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("RequestID:%6d: CreateJob(%d) - Error writing the response: %v", requestID, response.JobID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: CreateJob(%d) - completed\n", requestID, response.JobID)
	}
}

// ListJobs handles the job list endpoint
func ListJobs(j *repository.JobRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: ListJobs - starting\n", requestID)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(j.GetJobs()); err != nil {
			log.Printf("RequestID:%6d: ListJobs - Error writing the response: %v", requestID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: ListJobs - completed\n", requestID)
	}
}

// GetJob handles the job get endpoint
func GetJob(j *repository.JobRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: GetJob - starting\n", requestID)

		vars := mux.Vars(r)
		jobID, err := strconv.ParseInt(vars["job_id"], 10, 64)
		if err != nil {
			log.Printf("RequestID:%6d: GetJob - Invalid job_id: %v", requestID, err)
			http.Error(w, "Invalid job_id", http.StatusBadRequest)
			return
		}
		job := j.GetJob(jobID)
		if job.JobID <= 0 {
			log.Printf("RequestID:%6d: GetJob(%d) - Not found", requestID, jobID)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(job); err != nil {
			log.Printf("RequestID:%6d: GetJob(%d) - Error writing the response: %v", requestID, jobID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: GetJob(%d) - completed\n", requestID, jobID)
	}
}

// DeleteJob handles the job delete endpoint
func DeleteJob(j *repository.JobRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: DeleteJob - starting\n", requestID)

		var request struct {
			JobID int64 `json:"job_id"`
		}
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("RequestID:%6d: DeleteJob - Error reading the body: %v", requestID, err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &request); err != nil {
			log.Printf("RequestID:%6d: DeleteJob - Error parsing the body: %v", requestID, err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		if err := j.DeleteJob(request.JobID); err != nil {
			log.Printf("RequestID:%6d: DeleteJob(%d) - Not found", requestID, request.JobID)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		// Set status here as we're not writing to the response
		w.WriteHeader(http.StatusOK)
		log.Printf("RequestID:%6d: DeleteJob(%d) - completed\n", requestID, request.JobID)
	}
}
