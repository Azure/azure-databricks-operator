# Azure Databricks operator

## Introduction

Azure Databricks operator contains two projects. The golang application is a Kubernetes controller that watches CRDs that defines a Databricks job (input, output, functions, transformers, etc) and The Python Flask App sends commands to the Databricks.

![alt text](docs/images/azure-databricks-operator.jpg "high level architecture")

The project was built using

1. [Kubebuilder](https://book.kubebuilder.io/)
2. [Swagger Codegen](https://github.com/swagger-api/swagger-codegen)
3. [Flask-RESTPlus](http://flask-restplus.readthedocs.io)
4. [Flask](http://flask.pocoo.org/)

![alt text](docs/images/development-flow.jpg "development flow")

### Prerequisites And Assumptions

1. You have [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/),[Kind](https://github.com/kubernetes-sigs/kind) or docker for desktop installed on your local computer with RBAC enabled.
2. You have a Kubernetes cluster running.
3. You have the kubectl command line (kubectl CLI) installed.
4. You have Helm and Tiller installed.

* Configure a Kubernetes cluster in your machine
    > You need to make sure a kubeconfig file is configured.
    > if you opt AKS, you can use: `az aks get-credentials --resource-group $RG_NAME --name $Cluster_NAME`

Basic commands to check your cluster

```shell
    kubectl config get-contexts
    kubectl cluster-info
    kubectl version
    kubectl get pods -n kube-system
```

#### Kubernetes on WSL
    
on windows command line run `kubectl config view` to find the values of [windows-user-name],[minikubeip],[port]

```shell
mkdir ~/.kube \
&& cp /mnt/c/Users/[windows-user-name]/.kube/config ~/.kube

kubectl config set-cluster minikube --server=https://<minikubeip>:<port> --certificate-authority=/mnt/c/Users/<windows-user-name>/.minikube/ca.crt
kubectl config set-credentials minikube --client-certificate=/mnt/c/Users/<windows-user-name>/.minikube/client.crt --client-key=/mnt/c/Users/<windows-user-name>/.minikube/client.key
kubectl config set-context minikube --cluster=minikube --user=minikub

```

More info:

https://devkimchi.com/2018/06/05/running-kubernetes-on-wsl/

https://www.jamessturtevant.com/posts/Running-Kubernetes-Minikube-on-Windows-10-with-WSL/

### How to use operator

*Docs are work in progress*

Create .env file and set values of `DATABRICKS_HOST` and `DATABRICKS_TOKEN`

```
DATABRICKS_HOST=https://australiaeast.azuredatabricks.net
DATABRICKS_TOKEN=xxxx
```

1. To install CRDs into a cluster : `kubectl apply -f databricks-operator/config/crds` or `make install -C databricks-operator`
2. To deploy controller in the configured Kubernetes cluster in ~/.kube/config `kustomize build databricks-operator/config | kubectl apply -f -`
3. Change NotebookJob name from `sample1run1` to your desired name and  Update the values in `microsoft_v1beta2_notebookjob.yaml`

```
kubectl apply -f databricks-operator/config/samples/microsoft_v1beta2_notebookjob.yaml
kubectl get notebookjob
kubectl describe notebookjob kubectl sample1run1
```

Basic commands to check the new Notebookjob
```
kubectl get crd
kubectl -n databricks-operator-system get svc
kubectl -n databricks-operator-system get pod
kubectl -n databricks-operator-system describe  pod databricks-operator-controller-manager-0
kubectl -n databricks-operator-system logs  databricks-operator-controller-manager-0 -c dbricks -f
```

There is a also a Make file that you can use to install, test and deploy. 

 ```
make -C databricks-operator
make docker-build IMG=azadehkhojandi/databricks-operator -C databricks-operator
make docker-push IMG=azadehkhojandi/databricks-operator -C databricks-operator
make deploy -C databricks-operator

```


## Main Contributors

1. Jordan Knight [Github](https://github.com/jakkaj), [Linkedin](https://www.linkedin.com/in/jakkaj/)
2. Paul Bouwer [Github](https://github.com/paulbouwer), [Linkedin](https://www.linkedin.com/in/pbouwer/)
3. Lace Lofranco [Github](https://github.com/devlace), [Linkedin](https://www.linkedin.com/in/lacelofranco/)
4. Allan Targino [Github](https://github.com/allantargino), [Linkedin](https://www.linkedin.com/in/allan-targino//)
5. Rian Finnegan [Github](https://github.com/xtellurian), [Linkedin](https://www.linkedin.com/in/rian-finnegan-97651b55/)
6. Jason Goodselli [Github](https://github.com/JasonTheDeveloper), [Linkedin](https://www.linkedin.com/in/jason-goodsell-2505a3b2/)
7. Craig Rodger [Github](https://github.com/crrodger), [Linkedin](https://www.linkedin.com/in/craigrodger/)
8. Azadeh Khojandi [Github](https://github.com/Azadehkhojandi), [Linkedin](https://www.linkedin.com/in/azadeh-khojandi-ba441b3/)

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
