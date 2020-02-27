# Load testing with locust

The load testing project for the [azure-databricks-operator](https://github.com/microsoft/azure-databricks-operator/) can be found under `/locust`. Tests are built and run using the python [Locust load testing framework](https://docs.locust.io/en/stable/index.html).

## Table of contents <!-- omit in toc -->

- [Load testing with locust](#load-testing-with-locust)
  - [Build and run](#build-and-run)
    - [Run unit tests](#run-unit-tests)
    - [Run load tests in kind](#run-load-tests-in-kind)
      - [Set error conditions](#set-error-conditions)
  - [Contribute](#contribute)
    - [Extending the supported Databricks functionality](#extending-the-supported-databricks-functionality)
    - [Add load test scenarios](#add-load-test-scenarios)
    - [Adding unit tests](#adding-unit-tests)
    - [Updating a dashboard](#updating-a-dashboard)
  - [Prometheus Endpoint](#prometheus-endpoint)
  - [Known issues](#known-issues)

## Build and run

Everything needed to build and test the project is set up in the dev container.

To run the project without the dev container you need:

- Python 3
- Pip
- Python environment set up with:

    ```bash
    python -m venv venv
    source venv/bin/activate # You can also tell VSCode to use the interpretter in this location
    pip install -r requirements.dev.txt
    pip install -r requirements.txt
    ```

### Run unit tests

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

### Run load tests in kind

> Before proceeding make sure your container or environment is up and running

1. Deploy locust to local KIND instance. Set `LOCUST_FILE` to the  the locust scenario you'd like to run from  `locust/behaviours`.

    ```bash
    make run-load-testing LOCUST_FILE="behaviours/scenario1_run_submit_delete.py" LOCUST_ARGS=
    ```

2. Once all services are up, port-forward them for access

    ```bash
    make port-forward  
    ```

3. Navigate to http://localhost:8089 to start the load test from the locust web UI

4. Navigate to http://localhost:8080 with the below credentials to view the Grafana dashboards

    ```text
    Username: admin
    Password: prom-operator
    ```

> Note: If one of these port-forwards stops working use `ps aux | grep kubectl` and look for the process id of the one thats broken then use `kill 21283` (your id in there) to stop it. Then rerun the port-forward command

Good to know:

- Change the load test scenario file after deploying with:

    ```bash
    locust -f behaviours/<my_locust_file>.py
    ```

#### Set error conditions

For some of the load test scenarios we want to trigger error behaviour in the mockAPI during a test run.

1. Port-forward the mockAPI service

    ```bash
    # port-forward to localhost:8085
    kubectl port-forward -n databricks-mock-api svc/databricks-mock-api 8085:8080
    ```

2. Issue a `PATCH` request to update the error rate, e.g. to set 20% probability for status code 500 responses

    ```bash
    curl --request PATCH \
        --url http://localhost:8085/config \
        --header 'content-type: application/json'   \
        --data '{"DATABRICKS_MOCK_API_ERROR_500_PROBABILITY":20}'
    ```

> For more information see [mockAPI features](mockapi.md#Features)

## Contribute

### Extending the supported Databricks functionality

- `/locust_files/db_locust` contains all files related to how Locust can interact with Databricks using K8s via the [azure-databricks-operator](https://github.com/microsoft/azure-databricks-operator/)
  - `db_locust`: The brain of the behaviour driven tests. Inherits from the default `Locust`, read more [here](https://docs.locust.io/en/stable/testing-other-systems.html)
  - `db_client.py`: Core client used by the `db_locust`. It is a wrapper of "sub" clients that interface to specific databricks operator Kinds
    - `db_run_client.py`: all actions relating to `run` api interfaces
    - More clients to be added - ***this is where the majority of contributions will be made***
  - `db_decorator.py`: A simple decorator for Databricks operations that gives basic metric logging and error handling

### Add load test scenarios

- Load test scenario files are added to the `/behaviours` directory
- These files take the recommended format described by the Locust documentation representing the behvaiour of a single (or set of) users

### Adding unit tests

The project is setup to automatically discover any tests under the `locust/test` folder. Provided the following criteria are met:

- your test `.py` file follows the naming convention `<something>_test.py`
- within your test file your methods follow the naming convention `def test_<what you want to test>()`

### Updating a dashboard

Best way I've found is to import the JSON for the board into the Grafana instance, edit it using the UI then export it back to JSON and update the file in the repo.

## Prometheus Endpoint

This suite of Locust tests exposes stats to Prometheus via a web endpoint.

The endpoint is exposed at `/export/prometheus`. When running the tests with the web endpoints enabled, you can visit <http://localhost:8089/export/prometheus> to see the exported stats.

## Known issues

- When the port you're forwarding your Locust server to is not exposed from the container, you cannot hit it from your localhost machine. Use the [VSCode temporary port-forwarding](https://code.visualstudio.com/docs/remote/containers#_temporarily-forwarding-a-port) to resolve this.