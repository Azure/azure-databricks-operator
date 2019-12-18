# Resources

## Kubernetes on WSL

On Windows command line run `kubectl config view` to find the values of [windows-user-name],[minikube-ip],[port]:

```sh
mkdir ~/.kube && cp /mnt/c/Users/[windows-user-name]/.kube/config ~/.kube
```

If you are using minikube you need to set bellow settings 
```sh
# allow kubectl to trust the certificate authority of minikube
kubectl config set-cluster minikube \
    --server=https://[minikube-ip]:[port] \
    --certificate-authority=/mnt/c/Users/[windows-user-name]/.minikube/ca.crt

# configure the client certificate to use when talking to minikube
kubectl config set-credentials minikube \
    --client-certificate=/mnt/c/Users/[windows-user-name]/.minikube/client.crt \
    --client-key=/mnt/c/Users/[windows-user-name]/.minikube/client.key

# create the context minikube with cluster and user info created above
kubectl config set-context minikube --cluster=minikube --user=minikub
```

More info:

- https://devkimchi.com/2018/06/05/running-kubernetes-on-wsl/
- https://www.jamessturtevant.com/posts/Running-Kubernetes-Minikube-on-Windows-10-with-WSL/

## Build pipelines

- [Create a pipeline and add a status badge to Github](https://docs.microsoft.com/en-us/azure/devops/pipelines/create-first-pipeline?view=azure-devops&tabs=tfs-2018-2)
- [Customize status badge with shields.io](https://shields.io/)

## Operator metrics

- Operator telemetry metrics are exposed via standard [Prometheus](https://prometheus.io/) format endpoints. 
- [Prometheus-Operator](https://github.com/coreos/prometheus-operator) is included as part of the operator deployment via Helm chart.
    - Prometheus configuration is generated via the `config/default/kustomization.yaml`
    - Installation of Prometheus-Operator can be manually triggered via command `make install-prometheus`
    - If you don't want Prometheus-Operator configuration generated, it can be disabled by commenting out the line indicated in `config/default/kustomization.yaml`
    - *NOTE:* If you don't have the Prometheus-Operator installed, the ServiceMonitor CRD will not be available to you
- Custom metrics exposed by the Operator can be found by searching for `databricks_` inside the Prometheus web ui
- Metrics follow the naming guidlines recommended by Prometheus 

### How to access the Prometheus instance
- Have the operator installed and running locally. See [deploy.md](https://github.com/microsoft/azure-databricks-operator/blob/master/docs/deploy.md)
- Determine the name of Prometheus service running in your cluster (by default this will be prom-azure-databricks-oper-prometheus)
- Port forward localhost:9090 to your service: `kubectl port-forward service/prom-azure-databricks-oper-prometheus 9090:9090`
    - If using VSCode and Dev Container, you may need to expose the internal port out to your host machine (Command Pallete > Remote Containers Forward Port From Container) 
- Using a browser navigate to `http://localhost:9090` to view the Prometheus dashboard
- For more information regarding the usage of Prometheus please view the [docs here](https://prometheus.io/)

### How To scrape the metrics from a single intance of the Operator running on a Pod: 
- Have the operator installed and running locally. See [deploy.md](https://github.com/microsoft/azure-databricks-operator/blob/master/docs/deploy.md)
- Determine the name of the pod running your operator: `kubectl get pods -n azure-databricks-operator-system`
- Port forward localhost:8080 to your pod: `kubectl port-forward -n azure-databricks-operator-system pod/azure-databricks-operator-controller-manager-<id> 8080:8080`
- Open another terminal and curl request the metric endpoint: `curl localhost:8080/metrics`

### Counter metrics

In addition to the standard metrics that kubebuilder provides, the following custom metrics have been added.

The `databricks_request_duration_seconds` histogram provides metrics on the duration of calls via the databricks SDK and has the following labels:

|Name|Description|
|-|-|
|`object_type`|The type of object that the call relatest to, e.g. `dcluster`|
|`action`| The action being performed, e.g. `get`, `create`|
|`outcome`| `success` or `failure`|
