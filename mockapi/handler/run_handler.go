package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	"github.com/microsoft/azure-databricks-operator/mockapi/repository"
	dbjobsmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
)

//SubmitRun handles the runs submit endpoint
func SubmitRun(runRepo *repository.RunRepository, jobRepo *repository.JobRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: SubmitRun - starting\n", requestID)

		var run model.JobsRunsSubmitRequest
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("RequestID:%6d: SubmitRun - Error reading the body: %v", requestID, err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &run); err != nil {
			log.Printf("RequestID:%6d: SubmitRun - Error parsing the body: %v", requestID, err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		job := dbjobsmodel.JobSettings{
			NewCluster:   &run.NewCluster,
			Libraries:    &run.Libraries,
			SparkJarTask: &run.SparkJarTask,
		}

		response := dbjobsmodel.Run{
			RunID: runRepo.CreateRun(run, jobRepo.CreateJob(job)),
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("RequestID:%6d: SubmitRun(%d) - Error writing the response: %v", requestID, response.RunID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: SubmitRun(%d) - completed\n", requestID, response.RunID)
	}
}

// GetRun handles the runs get endpoint
func GetRun(j *repository.RunRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: GetRun - starting\n", requestID)

		vars := mux.Vars(r)
		runID, err := strconv.ParseInt(vars["run_id"], 10, 64)
		if err != nil {
			log.Printf("RequestID:%6d: GetRun - Invalid run_id: %v", requestID, err)
			http.Error(w, "Invalid run_id", http.StatusBadRequest)
			return
		}
		run := j.GetRun(runID)
		if run.RunID <= 0 {
			log.Printf("RequestID:%6d: GetRun(%d) - Not found", requestID, runID)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(run); err != nil {
			log.Printf("RequestID:%6d: GetRun(%d) - Error writing the response: %v", requestID, runID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: GetRun(%d) - completed\n", requestID, runID)
	}
}

// GetRunOutput handles the runs get output endpoint
func GetRunOutput(j *repository.RunRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: GetRunOutput - starting\n", requestID)

		vars := mux.Vars(r)
		runID, err := strconv.ParseInt(vars["run_id"], 10, 64)
		if err != nil {
			log.Printf("RequestID:%6d: GetRunOutput - Invalid run_id: %v", requestID, err)
			http.Error(w, "Invalid run_id", http.StatusBadRequest)
			return
		}
		run := j.GetRunOutput(runID)
		if run.Metadata.RunID <= 0 {
			log.Printf("RequestID:%6d: GetRunOutput(%d) - Not found", requestID, runID)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(run); err != nil {
			log.Printf("RequestID:%6d: GetRunOutput(%d) - Error writing the response: %v", requestID, runID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: GetRunOutput(%d) - completed\n", requestID, runID)
	}
}

// ListRuns handles the runs list endpoint
func ListRuns(j *repository.RunRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: ListRuns - starting\n", requestID)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(j.GetRuns()); err != nil {
			log.Printf("RequestID:%6d: ListRuns - Error writing the response: %v", requestID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("RequestID:%6d: ListRuns - completed\n", requestID)
	}
}

// CancelRun handles the runs cancel endpoint
func CancelRun(j *repository.RunRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: CancelRun - starting\n", requestID)

		var request struct {
			RunID int64 `json:"run_id"`
		}
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("RequestID:%6d: CancelRun - Error reading the body: %v", requestID, err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &request); err != nil {
			log.Printf("RequestID:%6d: CancelRun - Error parsing the body: %v", requestID, err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		if err := j.CancelRun(request.RunID); err != nil {
			log.Printf("RequestID:%6d: CancelRun(%d) - Not found", requestID, request.RunID)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		// Set status here as we're not writing to the response
		w.WriteHeader(http.StatusOK)
		log.Printf("RequestID:%6d: CancelRun(%d) - completed\n", requestID, request.RunID)
	}
}

// DeleteRun handles the runs delete endpoint
func DeleteRun(j *repository.RunRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getNewRequestID()
		log.Printf("RequestID:%6d: DeleteRun - starting\n", requestID)

		var request struct {
			RunID int64 `json:"run_id"`
		}
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("RequestID:%6d: DeleteRun - Error reading the body: %v", requestID, err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &request); err != nil {
			log.Printf("RequestID:%6d: DeleteRun - Error parsing the body: %v", requestID, err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		if err := j.DeleteRun(request.RunID); err != nil {
			log.Printf("RequestID:%6d: DeleteRun(%d) - Not found", requestID, request.RunID)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		// Set status here as we're not writing to the response
		w.WriteHeader(http.StatusOK)
		log.Printf("RequestID:%6d: DeleteRun(%d) - completed\n", requestID, request.RunID)
	}
}
