# How to use the operator

*Documentation is a work in progress*

## Quick start

1. Download [latest release.zip](https://github.com/microsoft/azure-databricks-operator/releases)

```sh
wget https://github.com/microsoft/azure-databricks-operator/releases/latest/download/release.zip
unzip release.zip
```

2. Create the `azure-databricks-operator-system` namespace

```sh
kubectl create namespace azure-databricks-operator-system
```

3. [Generate a databricks token](https://docs.databricks.com/api/latest/authentication.html#generate-a-token), and create Kubernetes secrets with values for `DATABRICKS_HOST` and `DATABRICKS_TOKEN`

```shell
    kubectl  --namespace azure-databricks-operator-system create secret generic dbrickssettings --from-literal=DatabricksHost="https://xxxx.azuredatabricks.net" --from-literal=DatabricksToken="xxxxx"
```

4. Apply the manifests for the CRD and Operator in `release/config`:

```sh
kubectl apply -f release/config
```

5. Create a test secret, you can pass the value of Kubernetes secrets into your notebook as Databricks secrets

```sh
kubectl create secret generic test-secret --from-literal=my_secret_key="my_secret_value"
```

6. In Databricks, [create a new Python Notebook](https://docs.databricks.com/user-guide/notebooks/notebook-manage.html#create-a-notebook) called `test-notebook` in the root of your [Workspace](https://docs.databricks.com/user-guide/workspace.html#folders). Put the following in the first cell of the notebook:

```py
secret_scope = dbutils.widgets.get("secret_scope")

secret_value = dbutils.secrets.get(scope=secret_scope, key="dbricks_secret_key") # this will come from a kubernetes secret
print(secret_value) # will be redacted

value = dbutils.widgets.get("flag")
print(value) # 'true'
```

7. Define your Notebook job and apply it

```yaml
apiVersion: databricks.microsoft.com/v1
kind: NotebookJob
metadata:
  annotations:
    databricks.microsoft.com/author: azkhojan@microsoft.com
  name: sample1run1
spec:
  notebookTask:
    notebookPath: "/test-notebook"
  timeoutSeconds: 500
  notebookSpec:
    "flag": "true"
  notebookSpecSecrets:
    - secretName: "test-secret"
      mapping :
        - "secretKey": "my_secret_key"
          "outputKey": "dbricks_secret_key"
  notebookAdditionalLibraries:
    - type: "maven"
      properties:
        coordinates: "com.microsoft.azure:azure-eventhubs-spark_2.11:2.3.9"
  clusterSpec:
    sparkVersion: "5.2.x-scala2.11"
    nodeTypeId: "Standard_DS12_v2"
    numWorkers: 1
```

8. Check the NotebookJob and Operator pod

```sh
# list all notebook jobs
kubectl get notebookjob
# describe a notebook job
kubectl describe notebookjob sample1run1
# get pods
kubectl -n azure-databricks-operator-system get pods
# describe the manager pod
azure-databricks-operator-controller-manager-xxxxx
# get logs from the manager container
kubectl -n azure-databricks-operator-system logs databricks-operator-controller-manager-xxxxx -c manager
```

9. Check the job ran with expected output in the Databricks UI.