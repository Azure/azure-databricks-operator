# Azure-Databricks-Operator Metrics

To help diagnose issues the operator exposes a set of [Prometheus metrics](https://prometheus.io/). Also included with this repo is a ServiceMonitor definition `yaml` that can be deployed to enable an existing (or new) Prometheus deployment to scrape these metrics.

## Operator metrics

- Enabling the Operator to output prometheus metrics is done via the customization of `config/default/kustomization.yaml`:
- If you don't want Prometheus-Operator configuration generated, it can be disabled by commenting out the line indicated in `config/default/kustomization.yaml`
> *NOTE:* If you don't have the Prometheus-Operator installed, the ServiceMonitor CRD will not be available to you. Please see the section below for further information about installation.
- All custom operator metrics exposed on the metrics endpoint are prefixed `databricks_`

In addition to the standard metrics that kubebuilder provides, the following custom metrics have been added.

The `databricks_request_duration_seconds` histogram provides metrics on the duration of calls via the databricks SDK and has the following labels:

|Name|Description|
|-|-|
|`object_type`|The type of CRD that the call relates to, e.g. `dcluster`|
|`action`| The action being performed, e.g. `get`, `create`|
|`outcome`| `success` or `failure`|

## Accessing Prometheus
- [Prometheus-Operator](https://github.com/coreos/prometheus-operator) can be installed in your cluster easily via Helm
> This repo provides an easy `make install-prometheus` to perform the Helm installtion
- Determine the name of Prometheus service running in your cluster (If you used our `make` command then this will default to `prom-azure-databricks-oper-prometheus`)
- Port forward localhost:9090 to your service: `kubectl port-forward service/prom-azure-databricks-oper-prometheus 9090:9090`
>If using VSCode and Dev Container, you may need to expose the internal port out to your host machine (Command Pallete > Remote Containers Forward Port From Container) 
- Using a browser navigate to `http://localhost:9090` to view the Prometheus dashboard
- For more information regarding the usage of Prometheus please view the [docs here](https://prometheus.io/)

## Grafana Dashboard
This repo also includes a Grafana dashboard named `Databricks Operator` that can be installed:
- If Prometheus-Operator is being used ensure then by default a sidecar is available to automatically install dashboards via `configmap`:
  - Update `config/prometheus/grafana-dashboard-configmap.yaml` to have a namespace matching your Grafana service
  - Apply `configmap` into the same namespace as your Grafana service running the sidecar `kubectl apply -f ./config/prometheus/grafana-dashboard-configmap.yaml`
- If you are not using Grafana/Prometheus-Operator, then the json can be extracted and imported manually
- The dashboard provides you general metrics regarding the health of your operator (see below for information about interpretting the chart data)

## Dashboard Charts

| Panel Name | Description | Usage |
|---|---|---|
| **Reconciliations Per Controller** | Increase/decrease in the total count of reconcile loops that are being performed | Graph is useful to determine the number of Reconcile loops that result in Error vs Success.  <br /><br />A spike in errors can indicate something wrong inside the operator logic such as missing config Secret containing Databricks uri etc.|
| **Controller Reconcile Time** | Median, 95% and mean time taken to perform a reconciliation loop  | Graph is useful to see how long the reconciliations take to complete as this is the complete lifecycle time and includes execution time in addition to upstream Databricks calls|
| **Workqueue Adds** | Increase/decrease of new work for the Operator to perform. | Graph is useful as it will show incoming rate of Operator work requests to create CRD's. <br /><br />Operator also re-queues items to re-process (polling runs for completion status for example) and so therefore graph will show rate increase even when not strictly "new work to be performed"<br /><br />*Note:* The Operator logic will re-queue certain tasks wIncrease/decrease of the Operator work queue depth | The work queue shows the number of reconcile loops currently awaiting and opportunity to run. <br /><br />Useful for seeing if the Operator is struggling to cope with incoming demands for work
| **Average Databricks Request Duration** | Average and 95% request duration when the Operator calls Databricks via its REST api | Useful for seeing how long Databricks is taking to respond to requests from Operator and can help diagnose network issues from the K8s cluster/potential timeout issues. |
| **Databricks REST endpoint calls - Success** | Increase/decrease of successful calls to databricks REST endpoints | Useful for identifying the throughput rate of the Operator calls to Databricks |
| **Databricks REST endpoint calls - Failure** | Increase/decrease of failed calls to databricks REST endpoints | Useful for identifying the error rate of external Databricks calls, a sudden spike could indicated a databricks outage or a potentially breaking change to the Databricks REST services causing all requests for a specific endpoint that is having issues |
| **Workqueue - Work Duration** | Median and 95% of how long in seconds processing an item from workqueue takes | Useful for measuring if one type of CRD request takes longer than others to complete<br /><br />*Note:* This metric is different to that of Controller Reconcile Time because it includes overhead execution time, not just the time spent executing with the Controller.
| **Workqueue - Queue Duration** | Median and 95% of how long in seconds an item stays in workqueue before being requested. | Useful for measuring if the work queue is backing up. Can indicate that something is starving the Operator of CPU