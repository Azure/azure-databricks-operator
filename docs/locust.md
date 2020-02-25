# Load testing with locust

The load testing project for the [azure-databricks-operator](https://github.com/microsoft/azure-databricks-operator/) can be found under `/locust`. Tests are built and run using the python [Locust load testing framework](https://docs.locust.io/en/stable/index.html).

## Table of contents <!-- omit in toc -->

- [Load testing with locust](#load-testing-with-locust)
  - [Deploying dependencies](#deploying-dependencies)
  - [Build and Test](#build-and-test)
    - [Deploy to kind](#deploy-to-kind)
    - [Run tests](#run-tests)
    - [Adding tests](#adding-tests)
  - [Contribute](#contribute)
    - [Extending the supported Databricks functionality](#extending-the-supported-databricks-functionality)
  - [Prometheus Endpoint](#prometheus-endpoint)
  - [Running test under docker](#running-test-under-docker)
    - [Test locally against cluster](#test-locally-against-cluster)
    - [Deploy into the cluster and run](#deploy-into-the-cluster-and-run)
      - [How do I update a dashboard](#how-do-i-update-a-dashboard)
      - [How do I set error conditions](#how-do-i-set-error-conditions)
  - [Known issues](#known-issues)

## Deploying dependencies

For documentation on deploying the `azure-databricks-operator` and `databricks-mock-api` for testing see [deploy/README.md](deploy/README.md)

## Build and Test

Everything needed to build and test the project is set up in the dev container.

To run the project without the dev container you need:

- Python 3
- Pip
- Set up your python environment

    ```bash
    python -m venv venv
    source venv/bin/activate # You can also tell VSCode to use the interpretter in this location
    pip install -r requirements.dev.txt
    pip install -r requirements.txt
    ```

### Deploy to kind

> Before proceeding make sure your container or environment is up and running

1. Deploy locust to local KIND instance. Set `LOCUST_FILE` to the  the locust scenario you'd like to run from  `locust/behaviours`.

    ```bash
    make kind-deploy-locust LOCUST_FILE="behaviours/scenario1_run_submit_delete.py"
    ```

2. Start the test server

    ```bash
    locust -f behaviours/<my_locust_file>.py
    ```

### Run tests

Tests are written using `pytest`. More information [is available here](https://docs.pytest.org/en/latest/).

> Before proceeding make sure your container or environment is up and running

1. Run the tests from the root of the project

    ```bash
    make test-locust
    ```

2. All being well you should see the following output:

    ```bash
    ============================================================================================= test session starts ==============================================================================================
    platform linux -- Python 3.7.4, pytest-5.3.2, py-1.8.0, pluggy-0.13.1
    rootdir: /workspace
    plugins: mock-1.13.0
    collected 8 items


    test/unit/db_run_client_test.py ........
    ```

### Adding tests

The project is setup to automatically discover any tests under the `locust/test` folder. Provided the following criteria are met:

- your test `.py` file follows the naming convention `<something>_test.py`
- within your test file your methods follow the naming convention `def test_<what you want to test>()`

## Contribute

- Test files are added to the `/behaviours` directory
- These files take the recommended format described by the Locust documentation representing the behvaiour of a single (or set of) users

### Extending the supported Databricks functionality

- `/locust_files/db_locust` contains all files related to how Locust can interact with Databricks using K8s via the [azure-databricks-operator](https://github.com/microsoft/azure-databricks-operator/)
  - `db_locust`: The brain of the behaviour driven tests. Inherits from the default `Locust`, read more [here](https://docs.locust.io/en/stable/testing-other-systems.html)
  - `db_client.py`: Core client used by the `db_locust`. It is a wrapper of "sub" clients that interface to specific databricks operator Kinds
    - `db_run_client.py`: all actions relating to `run` api interfaces
    - More clients to be added - ***this is where the majority of contributions will be made***
  - `db_decorator.py`: A simple decorator for Databricks operations that gives basic metric logging and error handling

## Prometheus Endpoint

This suite of Locust tests exposes stats to Prometheus via a web endpoint.

The endpoint is exposed at `/export/prometheus`. When running the tests with the web endpoints enabled, you can visit <http://localhost:8089/export/prometheus> to see the exported stats.

## Running test under docker

This guide assumes you have used `./deploy/README.md` to deploy an AKS Engine cluster and have the `KUBECONFG` set correctly and also used `./deploy/prometheus-grafana` to setup the `prometheus` operator.

### Test locally against cluster

To build and test the locust image locally againt the cluster you can run:

```bash
make docker-run-local
```

This will build a docker image **which contains the kubeconfig** file.

> Why put the file in the docker image? As we're using a devcontainer the normal approach of mounting a file doesn't work as the path on the host to the file (which is what the dameon uses) isn't the same as the path in the devcontainer so the file is never mounted.

### Deploy into the cluster and run

1. To deploy into the cluster run and port forward:

    ```bash
    CONTAINER_REGISTRY=$ACR_LOGIN_SERVER make deploy-loadtest
    k port-forward service/locust-loadtest 8089:8089 9090:9090 -n locust
    ```

2. Visit `http://localhost:8089` to start the loadtest from the locust web UI.

3. View stats on the test

      ```bash
      kubectl port-forward service/prom-azure-databricks-operator-grafana 8080:80
      ```

      ```text
      Username: admin
      Password: prom-operator
      http://localhost:8080
      ```

    Then navigate to the locust dashboard to view the results.

If you want to setup port forwards for all the things then do the following:

```bash
k port-forward service/prom-azure-databricks-oper-prometheus 9091:9090 &
k port-forward service/locust-loadtest 8089:8089 9090:9090 &
k port-forward service/prom-azure-databricks-operator-grafana 8080:80 &

Browse to locust webui   -> http://localhost:8089/
Browse to locust metrics -> http://localhost:9090/
Browse to Prometheus     -> http://localhost:9091/targets
Browse to Grafana        -> http://localhost:8080/
```

> Note: If one of these port forwards stops working use `ps aux | grep kubectl` and look for the process id of the one thats broken then use `kill 21283` (your id in there) to stop it. Then rerun the port forward command.

#### How do I update a dashboard

Best way I've found is to import the JSON for the board into the grafana instance, edit it using the UI then export it back to JSON and update the file in the repo.

#### How do I set error conditions

For some of the load test scenarios we want to trigger error behaviour in the mock-api during a test run.

First step for this is to port-forward the mock-api service

```bash
# port-forward to localhost:8085
kubectl port-forward -n databricks-mock-api svc/databricks-mock-api 8085:8080
```

Next we can issue a `PATCH` request to update the error rate, e.g. to set 20% probability for status code 500 responses

```bash
curl --request PATCH \
     --url http://localhost:8085/config \
     --header 'content-type: application/json'   \
     --data '{"DATABRICKS_MOCK_API_ERROR_500_PROBABILITY":20}'
```
## Known issues

- When the port you're forwarding your Locust server to is not exposed from the container, you cannot hit it from your localhost machine. Use the [VSCode temporary port forwarding](https://code.visualstudio.com/docs/remote/containers#_temporarily-forwarding-a-port) to resolve this.