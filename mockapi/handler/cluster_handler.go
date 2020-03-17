package handler

import (
	"net/http"

	"github.com/microsoft/azure-databricks-operator/mockapi/repository"
)

//CreateCluster handles the cluster create endpoint
func CreateCluster(j *repository.ClusterRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

// ListClusters handles the cluster list endpoint
func ListClusters(j *repository.ClusterRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

// GetCluster handles the cluster get endpoint
func GetCluster(j *repository.ClusterRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

// EditCluster handles the cluster edit endpoint
func EditCluster(j *repository.ClusterRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

// DeleteCluster handles the cluster delete endpoint
func DeleteCluster(j *repository.ClusterRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}
