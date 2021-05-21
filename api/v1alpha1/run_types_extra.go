package v1alpha1

import (
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
	dblibsmodels "github.com/polar-rams/databricks-sdk-golang/azure/libraries/models"
)

// ClusterSpec is similar to dbmodels.ClusterSpec, the reason it
// exists is because dbmodels.ClusterSpec doesn't support ExistingClusterName
// ExistingClusterName allows discovering databricks clusters by it's kubernetese object name
type ClusterSpec struct {
	ExistingClusterID   string                 `json:"existing_cluster_id,omitempty" url:"existing_cluster_id,omitempty"`
	ExistingClusterName string                 `json:"existing_cluster_name,omitempty" url:"existing_cluster_name,omitempty"`
	NewCluster          *dbmodels.NewCluster   `json:"new_cluster,omitempty" url:"new_cluster,omitempty"`
	Libraries           []dblibsmodels.Library `json:"libraries,omitempty" url:"libraries,omitempty"`
}

// ToK8sClusterSpec converts a databricks ClusterSpec object to k8s ClusterSpec object.
// It is needed to add ExistingClusterName and follow k8s camleCase naming convention
func ToK8sClusterSpec(dbjs *dbmodels.ClusterSpec) ClusterSpec {
	var k8sjs ClusterSpec
	k8sjs.ExistingClusterID = dbjs.ExistingClusterID
	k8sjs.NewCluster = &dbjs.NewCluster
	k8sjs.Libraries = *dbjs.Libraries
	return k8sjs
}

// ToDatabricksClusterSpec converts a k8s ClusterSpec object to a DataBricks ClusterSpec object.
// It is needed to add ExistingClusterName and follow k8s camleCase naming convention
func ToDatabricksClusterSpec(k8sjs *ClusterSpec) dbmodels.ClusterSpec {

	var dbjs dbmodels.ClusterSpec
	dbjs.ExistingClusterID = k8sjs.ExistingClusterID
	dbjs.NewCluster = *k8sjs.NewCluster
	dbjs.Libraries = &k8sjs.Libraries
	return dbjs
}
