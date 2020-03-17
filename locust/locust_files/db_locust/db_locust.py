from locust import Locust
from .db_client import DbClient


class DbLocust(Locust):
    """
    Custom Locust for Databricks Operator
    """

    def __init__(self, *args, **kwargs):
        super(DbLocust, self).__init__(*args, **kwargs)
        self.client = DbClient()
