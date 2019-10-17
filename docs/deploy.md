# Deploy the operator

## Prerequests

- You have `kubectl` configured pointing to the target Kubernetes cluster.
- You have access to a DataBricks cluster and able to generate PAT token. To generate a token, check
  [generate a DataBricks token](https://docs.databricks.com/api/latest/authentication.html#generate-a-token).

## Step-by-step guide

This will deploy the operator in namespace `azure-databricks-operator-system`. If you want to customise
the namespace, you can either search-replace the namespace, or use `kustomise` by following the next
section.

1. Download [the latest release manifests](https://github.com/microsoft/azure-databricks-operator/releases):

```sh
wget https://github.com/microsoft/azure-databricks-operator/releases/latest/download/release.zip
unzip release.zip
```

2. Create the `azure-databricks-operator-system` namespace:

```sh
kubectl create namespace azure-databricks-operator-system
```

3. Create Kubernetes secrets with values for `DATABRICKS_HOST` and `DATABRICKS_TOKEN`:

```shell
kubectl --namespace azure-databricks-operator-system \
    create secret generic dbrickssettings \
    --from-literal=DatabricksHost="https://xxxx.azuredatabricks.net" \
    --from-literal=DatabricksToken="xxxxx"
```

4. Apply the manifests for the Operator and CRDs in `release/config`:

```sh
kubectl apply -f release/config
```

## Use kustomize to customise your deployment

1. Clone the source code:

```sh
git clone git@github.com:microsoft/azure-databricks-operator.git
```

2. Edit file `config/default/kustomization.yaml` file to change your preferences

3. Use `kustomize` to generate the final manifests and deploy:

```sh
kustomize build config/default | kubectl apply -f -
```

4. Deploy the CRDs:

```sh
kubectl apply -f config/crd/bases
```

## Test your deployment

1. Deploy a sample job, this will create a job in the default namespace:

```sh
curl https://raw.githubusercontent.com/microsoft/azure-databricks-operator/master/config/samples/databricks_v1alpha1_djob.yaml | kubectl apply -f -
```

2. Check the Job in Kubernetes:

```sh
kubectl get djob
```

3. Check the job is created successfully in DataBricks.

## Troubleshooting

If you encounter any issue, you can check the log of the operator by pulling it from Kubernetes:

```sh
# get the pod name of your operator
kubectl --namespace azure-databricks-operator-system get pods

# pull the logs
kubectl --namespace azure-databricks-operator-system logs -f [name_of_the_operator_pod]
```
