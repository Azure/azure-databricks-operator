/*
Copyright 2019 microsoft.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notebookjob

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"

	microsoftv1beta1 "microsoft/azure-databricks-operator/databricks-operator/pkg/apis/microsoft/v1beta1"

	db "github.com/xinsnake/databricks-sdk-golang"
	dbazure "github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const finalizerName = "notebookjob.finalizers.microsoft.k8s.io"

var log = logf.Log.WithName("notebookjob-controller")

// Add creates a new NotebookJob Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	var apiConfig db.DBClientOption
	//TODO Make it configurable
	apiConfig.Host = os.Getenv("DATABRICKS_HOST")
	apiConfig.Token = os.Getenv("DATABRICKS_TOKEN")
	var apiClient dbazure.DBClient
	apiClient.Init(apiConfig)

	return &ReconcileNotebookJob{
		Client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		recorder:  mgr.GetRecorder("notebookjob-controller"),
		apiClient: apiClient,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("notebookjob-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to NotebookJob
	err = c.Watch(&source.Kind{Type: &microsoftv1beta1.NotebookJob{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNotebookJob{}

// ReconcileNotebookJob reconciles a NotebookJob object
type ReconcileNotebookJob struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder

	apiClient dbazure.DBClient
}

// Reconcile reads that state of the cluster for a NotebookJob object and makes changes based on the state read
// and what is in the NotebookJob.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=microsoft.k8s.io,resources=notebookjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=microsoft.k8s.io,resources=notebookjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=microsoft.k8s.io,resources=events,verbs=create;patch
func (r *ReconcileNotebookJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the NotebookJob instance
	instance := &microsoftv1beta1.NotebookJob{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.IsBeingDeleted() {
		err := r.handleFinalizer(instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		return reconcile.Result{}, nil
	}

	if !instance.HasFinalizer(finalizerName) {
		err = r.addFinalizer(instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error when removing finalizer: %v", err)
		}
		return reconcile.Result{}, nil
	}

	if !instance.IsRunning() {
		err = r.submitRunToAPI(instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("error when submitting job to API: %v", err)
		}
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileNotebookJob) convertInstanceToRunDefinition(instance *microsoftv1beta1.NotebookJob) (
	string, dbmodels.ClusterSpec, dbmodels.JobTask, int, map[string]string, error) {

	var runName = instance.ObjectMeta.Name

	var clusterSpec dbmodels.ClusterSpec

	clusterSpec.NewCluster.SparkVersion = instance.Spec.ClusterSpec.SparkVersion
	clusterSpec.NewCluster.NodeTypeID = instance.Spec.ClusterSpec.NodeTypeId
	clusterSpec.NewCluster.NumWorkers = int32(instance.Spec.ClusterSpec.NumWorkers)
	clusterSpec.Libraries = make([]dbmodels.Library, len(instance.Spec.NotebookAdditionalLibraries))
	for i, v := range instance.Spec.NotebookAdditionalLibraries {
		if v.Type == "jar" {
			clusterSpec.Libraries[i].Jar = v.Properties["path"]
		}
		if v.Type == "egg" {
			clusterSpec.Libraries[i].Egg = v.Properties["path"]
		}
		if v.Type == "whl" {
			clusterSpec.Libraries[i].Whl = v.Properties["path"]
		}
		if v.Type == "pypi" {
			clusterSpec.Libraries[i].Pypi.Package = v.Properties["package"]
			clusterSpec.Libraries[i].Pypi.Repo = v.Properties["repo"]
		}
		if v.Type == "maven" {
			clusterSpec.Libraries[i].Maven.Coordinates = v.Properties["coordinates"]
			clusterSpec.Libraries[i].Maven.Repo = v.Properties["repo"]
			// TODO the spec doesn't support array
			// clusterSpec.Libraries[i].Maven.Exclusions = v.Properties["exclusions"]
		}
		if v.Type == "cran" {
			clusterSpec.Libraries[i].Cran.Package = v.Properties["package"]
			clusterSpec.Libraries[i].Cran.Repo = v.Properties["repo"]
		}
	}

	var jobTask dbmodels.JobTask
	jobTask.NotebookTask.NotebookPath = instance.Spec.NotebookTask.NotebookPath
	jobTask.NotebookTask.BaseParameters = make([]dbmodels.ParamPair, len(instance.Spec.NotebookSpec))
	counter := 0
	for k, v := range instance.Spec.NotebookSpec {
		jobTask.NotebookTask.BaseParameters[counter] = dbmodels.ParamPair{
			Key: k, Value: v,
		}
		counter++
	}

	var timeoutSeconds = instance.Spec.TimeoutSeconds

	var scopeSecrets = make(map[string]string, len(instance.Spec.NotebookSpecSecrets))
	for _, notebookSpecSecret := range instance.Spec.NotebookSpecSecrets {
		secretName := notebookSpecSecret.SecretName
		secret := &v1.Secret{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, secret)
		if err != nil {
			return runName, clusterSpec, jobTask, timeoutSeconds, scopeSecrets, err
		}
		for _, mapping := range notebookSpecSecret.Mapping {
			secretvalue := secret.Data[mapping.SecretKey]
			tempkey := mapping.OutputKey
			scopeSecrets[tempkey] = fmt.Sprintf("%s", secretvalue)
		}
	}

	return runName, clusterSpec, jobTask, timeoutSeconds, scopeSecrets, nil
}

func (r *ReconcileNotebookJob) getEventHubConnectionString(instance *microsoftv1beta1.NotebookJob, secretName string) (string, error) {
	secret := &v1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, secret)

	if err != nil {
		return "", err
	}

	eventHubName := secret.Data["eventHubName"]
	connectionString := secret.Data["connectionString"]
	fullConnectionString := fmt.Sprintf("%s;EntityPath=%s", connectionString, eventHubName)

	return fullConnectionString, nil
}

func (r *ReconcileNotebookJob) addFinalizer(instance *microsoftv1beta1.NotebookJob) error {
	instance.AddFinalizer(finalizerName)
	err := r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("failed to update finalizer: %v", err)
	}
	r.recorder.Event(instance, "Normal", "Updated", fmt.Sprintf("finalizer %s added", finalizerName))
	return nil
}

func (r *ReconcileNotebookJob) handleFinalizer(instance *microsoftv1beta1.NotebookJob) error {
	if instance.HasFinalizer(finalizerName) {
		// our finalizer is present, so lets handle our external dependency
		if err := r.deleteExternalDependency(instance); err != nil {
			return err
		}

		instance.RemoveFinalizer(finalizerName)
		if err := r.Update(context.Background(), instance); err != nil {
			return err
		}
	}
	// Our finalizer has finished, so the reconciler can do nothing.
	return nil
}

func (r *ReconcileNotebookJob) deleteExternalDependency(instance *microsoftv1beta1.NotebookJob) error {
	log.Info("deleting the external dependencies")
	runID := instance.Spec.NotebookTask.RunID
	return r.apiClient.Jobs().RunsDelete(int64(runID))
}

func (r *ReconcileNotebookJob) submitRunToAPI(instance *microsoftv1beta1.NotebookJob) error {
	// get definition
	runName, clusterSpec, jobTask, timeoutSeconds, scopeSecrets, err := r.convertInstanceToRunDefinition(instance)
	if err != nil {
		return err
	}

	// create scope and put secrets
	secretScopeName := runName + "_scope"
	err = r.apiClient.Secrets().CreateSecretScope(secretScopeName, "users")
	if err != nil {
		return err
	}
	for k, v := range scopeSecrets {
		r.apiClient.Secrets().PutSecretString(v, secretScopeName, k)
	}

	// submit run
	runResponse, err := r.apiClient.Jobs().RunsSubmit(runName, clusterSpec, jobTask, int32(timeoutSeconds))
	if err != nil {
		return err
	}

	if runResponse.RunID == 0 {
		return fmt.Errorf("result from API didn't return any values")
	}

	instance.Spec.NotebookTask.RunID = int(runResponse.RunID)
	err = r.Update(context.TODO(), instance)
	if err != nil {
		return fmt.Errorf("error when updating NotebookJob after submitting to API: %v", err)
	}

	r.recorder.Event(instance, "Normal", "Updated", "runID added")

	return nil
}
