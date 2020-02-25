# Mock Databricks API

The API found under `/mockapi` is a Databricks mock API for following success scenarios:

- [Jobs/](https://docs.databricks.com/dev-tools/api/latest/jobs.html):
  - Create
  - Get
  - List
  - Delete
  - Runs/
    - Submit
    - Get
    - GetOutput
    - List

In addition, each submitted run will cycle through it's life states (PENDING -> RUNNING -> TERMINATING -> TERMINATED).
The length of each state is set in run_repository.go : timePerRunLifeState.

## Table of contents <!-- omit in toc -->

- [Mock Databricks API](#mock-databricks-api)
  - [Features](#features)
    - [Configurable API latency](#configurable-api-latency)
    - [Configurable Rate Limiting](#configurable-rate-limiting)
    - [Configurable Errors](#configurable-errors)
    - [Dynamic Configuration](#dynamic-configuration)
  - [Running locally](#running-locally)
  - [Running in Kind](#running-in-kind)
  - [Running in a separate cluster](#running-in-a-separate-cluster)
    - [Prerequisites](#prerequisites)
    - [Deploy to the cluster](#deploy-to-the-cluster)

## Features

### Configurable API latency

To simulate Databricks API more accurately, we've added an option to configure a range of latency for each request.

The latency range can be configured by adding a min and max value for desired latency in milliseconds for a fast and slow requests using the environment variables:

```text
DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MIN
DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MAX
DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN
DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX
```

When set, for each request will sleep for a time chosen at random between the min and max values.

If either of the variables is not set, the API will default to running with no latency.

### Configurable Rate Limiting

To allow rate-limiting requests to match Databricks API behaviour, a rate limit can be specified by setting `DATABRICKS_MOCK_API_RATE_LIMIT` environment variable to the number of requests per second that should be allowed against the API.

### Configurable Errors

To configure a percentage of responses that return a status code 500 response in the mock-api you can set `DATABRICKS_MOCK_API_ERROR_500_PROBABILITY`.

E.g. setting `DATABRICKS_MOCK_API_ERROR_500_PROBABILITY` to `20` will return a status code 500 response for roughly 20% of responses.

To configure a percentage of calls that should sink-hole, i.e. return no response and keep the connection open for 10 minutes, you can set `DATABRICKS_MOCK_API_ERROR_SINKHOLE_PROBABILITY`. Probabilities are as for `DATABRICKS_MOCK_API_ERROR_500_PROBABILITY`.

To configure a percentage of calls that should respond xml response with status code 200 response in the mock-api you can set`DATABRICKS_MOCK_API_ERROR_XML_RESPONSE_PROBABILITY`.Probabilities are as for `DATABRICKS_MOCK_API_ERROR_500_PROBABILITY`.

> NB: The combined probabilities must be <=100

### Dynamic Configuration

The API includes a `/config` endpoint that can be used to `GET`, `PUT` or `PATCH` configuration values. This allows configuration values to be retrieved from the API as well as changed without restarting the API (which would lose in-memory run data).

`GET` returns the full set of configurable values.

`PUT` expects a full set of configurable values to be specified and applies all the values.

`PATCH` allows one or more configurable values to be specified and only applies the values from the body.

## Running locally

Open this repo in VS Code and select "Remote-Containers: Reopen in Container".

Once the devcontainer has built and started, use `make run-mock-api` to run the API.

`mockapi_samples/*_sample.http` files contain example calls that can be made against the running API endpoints.

## Running in Kind

To run the mock api in Kind run `make kind-deploy-mock-api`. This will ensure a Kind cluster is created, deploy promethous with helm, build and load a docker image for the mock api into the Kind cluster and then create a Deployment and Service.

To test, run `kubectl port-forward svc/databricks-mock-api 8085:8080 -n databricks-mock-api` and make a request to <http://localhost:8085> to verify that the API is running

## Running in a separate cluster

### Prerequisites

This assumes that you have a container registry and a Kubernetes cluster that is able to pull images from it.

> NB: For now, run the below outside the devcontainer as permissions don't allow the config to be modified currently

If using Azure Container registry then run `az acr login -n your-azure-container-registry-name` to ensure that you are authenticated to push images to the registry

Ensure your KUBECONFIG is set to point to the cluster you want to deploy to.

### Deploy to the cluster

Deploy to AKS with

```bash
IMG=your-container-registry.azurecr.io/databricks-mock-api:vsomething make aks-deploy-mock-api
```

To test, run `kubectl port-forward svc/databricks-mock-api 8085:8080` and make a request to <http://localhost:8085> to verify that the API is running

> NB: Error: 'unauthorized: authentication required' make sure that the image name matches the ACR login server casing or try pushing the docker file outside of the container
