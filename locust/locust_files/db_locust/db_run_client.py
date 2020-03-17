from .db_decorator import locust_client_function
import logging
import uuid
from .constant import RUN_RESOURCE, K8_GROUP, VERSION
import copy
from kubernetes import client
from kubernetes.client.rest import ApiException
from locust import InterruptTaskSet
import time


class DbRunClient:
    """
    DbRunClient exposes the Databricks actions that a user will perform relating to Runs
    """

    def __init__(self, k8s_api):
        self.api = k8s_api

    @locust_client_function
    def create_run(self):
        logging.info("Run Creation: STARTED")

        run_name = "run-{}".format(str(uuid.uuid4()))
        run = copy.deepcopy(RUN_RESOURCE)
        run["metadata"]["name"] = run_name

        try:
            self.api.create_namespaced_custom_object(
                group=K8_GROUP,
                version=VERSION,
                namespace="default",
                plural="runs",
                body=run,
            )
        except ApiException as e:
            logging.error("API Exception: %s\n" % e)
            raise InterruptTaskSet(reschedule=False)

        logging.info("Run Creation: COMPLETE - {}".format(run_name))
        return run_name

    @locust_client_function
    def get_run(self, run_name):
        logging.info("Run Get: STARTED - %s", run_name)

        try:
            resource = self.api.get_namespaced_custom_object(
                group=K8_GROUP,
                version=VERSION,
                name=run_name,
                namespace="default",
                plural="runs",
            )
        except ApiException as e:
            logging.error("API Exception: %s\n" % e)
            raise InterruptTaskSet(reschedule=False)

        return resource

    @locust_client_function
    def delete_run(self, run_name):
        logging.info("Run Deletion: STARTED - %s", run_name)

        try:
            self.api.delete_namespaced_custom_object(
                group=K8_GROUP,
                version=VERSION,
                name=run_name,
                namespace="default",
                plural="runs",
                body=client.V1DeleteOptions(),
            )
        except ApiException as e:
            logging.error("API Exception: %s\n" % e)
            raise InterruptTaskSet(reschedule=False)

        logging.info("Run Deletion: COMPLETE - %s", run_name)

    @locust_client_function
    def poll_run_await_completion(self, run_name, max_attempts, polling_secs):
        logging.info("Waiting for run to complete: STARTED - %s", run_name)

        completed = False

        for x in range(max_attempts):
            run = self.get_run(run_name)

            if run is None:
                raise Exception("Unable to get job")

            if "status" in run:
                life_cycle_state = run["status"]["metadata"]["state"][
                    "life_cycle_state"
                ]

                if life_cycle_state == "TERMINATED":
                    logging.info("Run COMPLETED - %s" % run_name)
                    completed = True
                    break
                elif (
                    life_cycle_state == "SKIPPED"
                    or life_cycle_state == "INTERNAL_ERROR"
                ):
                    raise Exception(
                        "Run COMPLETED with error - life_cycle_state: %s"
                        % life_cycle_state
                    )

            logging.info("Run NOT yet complete: WAITING - %s" % run_name)
            time.sleep(polling_secs)

        if not completed:
            raise Exception("Run did not complete")

        logging.info("Waiting for run to complete: COMPLETE - %s" % run_name)
