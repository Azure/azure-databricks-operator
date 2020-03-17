from locust import TaskSequence, seq_task, between
from locust_files.db_locust import DbLocust
from random import randint


class NoopUserBehaviour(TaskSequence):
    """
    This class defines a set of tasks that are performed in a
    predefined order (a sequence).
    """

    @seq_task(1)
    def wait_short(self):
        request_time = randint(1, 5)
        self.client.noop.noop_with_delay(request_time)

    @seq_task(2)
    def wait_medium(self):
        self.client.noop.noop()

    @seq_task(3)
    def fail_fast(self):
        self.client.noop.noop_with_fail()


class NoopUser(DbLocust):
    """
    This class represents a user of the Databricks Operator
    with a particular behaviour
    """

    task_set = NoopUserBehaviour
    wait_time = between(1, 5)
