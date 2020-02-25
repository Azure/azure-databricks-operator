from locust import TaskSequence, seq_task, between
from locust_files.db_locust import DbLocust


class DatabricksRunSubmitUserBehaviour(TaskSequence):
    """
    This class defines a set of tasks that are performed in a
    predefined order (a sequence).
    """

    @seq_task(1)
    def create_databricks_run(self):
        self.run_name = self.client.runs.create_run()

    @seq_task(2)
    def await_databricks_run_complete(self):
        self.client.runs.poll_run_await_completion(self.run_name, 40, 10)

    @seq_task(3)
    def delete_databricks_run(self):
        self.client.runs.delete_run(self.run_name)


class DatabricksRunSubmitUser(DbLocust):
    """
    This class represents a user of the Databricks Operator
    with a particular behaviour
    """

    task_set = DatabricksRunSubmitUserBehaviour
    wait_time = between(1, 5)
