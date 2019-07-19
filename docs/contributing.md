# Requirements
If you're interested in contributing to this project, you'll need:
* Go installed - see this [Getting Started](https://golang.org/doc/install) guide for Go.
* `Kubebuilder` -  see this [Quick Start](https://book.kubebuilder.io/quick-start.html) guide for installation instructions.
* Kubernetes command-line tool `kubectl` 
* Access to a Kubernetes cluster. Some options are:
	* Locally hosted cluster, such as 
		* [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
		* [Kind](https://github.com/kubernetes-sigs/kind)
		* Docker for desktop installed localy with RBAC enabled.
	* Azure Kubernetes Service ([AKS](https://azure.microsoft.com/en-au/services/kubernetes-service/))
		* The [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/?view=azure-cli-latest) will be helpful here
		* Retrieve the config for the AKS cluster with `az aks get-credentials --resource-group $RG_NAME --name $Cluster_NAME`
* Setup access to your cluster using a `kubeconfig` file.  See [here](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) for more details
* Access to an Azure DataBricks instance.

Here are a few `kubectl` commands you can run to verify your `kubectl` installation and cluster setup
```
$ kubectl version
$ kubectl config get-contexts
$ kubectl config current-context
$ kubectl cluster-info
$ kubectl get pods -n kube-system
```

# Environment Variables
You will need to set some environment variables so that the operator knows what Azure DataBricks instance to reconcile with. 
```
$ export DATABRICKS_HOST=https://xxxx.azuredatabricks.net
$ export DATABRICKS_TOKEN=xxxxx
```

# Building and Running the operator

## Basics
The scaffolding for the project is generated using `Kubebuilder`. It is a good idea to become familiar with this [project](https://github.com/kubernetes-sigs/kubebuilder). The [quick start](https://book.kubebuilder.io/quick-start.html) guide is also quite useful.

See `Makefile` at the root directory of the project. By default, executing `make` will build the project and produce an executable at `./bin/manager`

For example, to quick start 
>this assumes dependencies have been downloaded and existing CRDs have been installed. See next section
```
$ git clone https://github.com/microsoft/azure-databricks-operator.git
$ cd azure-databricks-operator
$ make
$ ./bin/manager
```

Other tasks are defined in the `Makefile`. It would be good to familiarise yourself with them.

## Dependencies
The project uses external Go modules that are required to build/run. In addition, to run successfully, any CRDs defined in the project should be regenerated and installed. 

The following steps should illustrate what is required before the project can be run:
1. `go mod tidy` - download the dependencies (this can take a while and there is no progress bar - need to be patient for this one)
2. `make manifests` - regenerates the CRD manifests
3. `make install` -  installs the CRDs into the cluster
4. `make generate` - generate the code

At this point you will be able to build the binary with `go build -o bin/manager main.go`. Alternatively, this step and the required `make generate` step before hand is covered with the default `make`. 

## Running Tests
Running e2e tests require a configure Kubernetes cluster and Azure DataBricks connection (through specified environment variables)
```
make test
```

# Extending the Library
*This is a work in progress*

As previously mentioned, familiarity with `kubebuilder` is required for developing this operator. Kubebuilder generates the scaffolding for new Kubernetes APIs. 
```
$ kubebuilder create api --group databricks --version v1 --kind SecretScope
                                                                                                                Create Resource [y/n]
y
Create Controller [y/n]
y
Writing scaffold for you to edit...
api/v1/secretscope_types.go
controllers/secretscope_controller.go
Running make...
/Users/d886442/go/bin/controller-gen object:headerFile=./hack/boilerplate.go.txt paths=./api/...
go fmt ./...
go vet ./...
go build -o bin/manager main.go

$ 
```
You'll notice in the output above that the following files have been created:
* `api/v1/secretscope_types.go`
* `controllers/secretscope_controller.go`

These files are ready for you to fill in with the logic appropriate to the new resource you're creating. Once you've developed your API, ensure to regenerate and install your CRDs. See [Dependencies](#dependencies)
