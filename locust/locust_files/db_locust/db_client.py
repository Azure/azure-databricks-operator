from .db_run_client import DbRunClient
from .db_noop_client import DbNoopClient
from kubernetes import client, config


class DbClient:
    """
    DbClient exposes the Databricks actions that a user will perform
    """

    def __init__(self):
        # Try and load the kubeconfig or incluster config
        try:
            config.load_kube_config()
        except:
            config.load_incluster_config()

        self.api = client.CustomObjectsApi()
        self.runs = DbRunClient(self.api)
        self.noop = DbNoopClient()
