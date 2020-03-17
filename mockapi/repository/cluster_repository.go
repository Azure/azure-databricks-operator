package repository

import (
	"fmt"
	"github.com/google/uuid"
	dbmodel "github.com/xinsnake/databricks-sdk-golang/azure/models"
	"sync"
)

// ClusterRepository is a store for Cluster instances
type ClusterRepository struct {
	clusters  map[string]dbmodel.ClusterInfo
	writeLock sync.Mutex
}

// NewClusterRepository creates a new ClusterRepository
func NewClusterRepository() *ClusterRepository {
	return &ClusterRepository{
		clusters: map[string]dbmodel.ClusterInfo{},
	}
}

// GetCluster returns the Cluster with the specified ID or an empty Cluster
func (r *ClusterRepository) GetCluster(id string) dbmodel.ClusterInfo {
	if cluster, ok := r.clusters[id]; ok {
		return cluster
	}

	return dbmodel.ClusterInfo{}
}

// GetClusters returns all Clusters
func (r *ClusterRepository) GetClusters() []dbmodel.ClusterInfo {
	arr := []dbmodel.ClusterInfo{}
	for _, cluster := range r.clusters {
		arr = append(arr, cluster)
	}
	return arr
}

// CreateCluster adds an ID to the specified cluster and adds it to the collection
func (r *ClusterRepository) CreateCluster(cluster dbmodel.ClusterInfo) string {
	newID := uuid.New().String()
	cluster.ClusterID = newID

	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	r.clusters[newID] = cluster
	return cluster.ClusterID
}

// DeleteCluster deletes the cluster with the specified ID
func (r *ClusterRepository) DeleteCluster(id string) error {
	if _, ok := r.clusters[id]; ok {
		delete(r.clusters, id)
		return nil
	}
	return fmt.Errorf("Could not find Job with id of %s to delete", id)
}
