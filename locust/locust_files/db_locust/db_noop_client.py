from .db_decorator import locust_client_function
import logging
import time


class DbNoopClient:
    """
    DbNoopClient exposes the no-op actions purely for diagnostics
    and testing purposes
    """

    @locust_client_function
    def noop_with_delay(self, delay_in_seconds):
        logging.info(f"NoOp With Delay: STARTED")

        time.sleep(delay_in_seconds)

        logging.info("NoOp With Delay: COMPLETE")

    @locust_client_function
    def noop(self):
        logging.info(f"NoOp With Delay: STARTED")
        time.sleep(0.5)
        logging.info("NoOp With Delay: COMPLETE")

    @locust_client_function
    def noop_with_fail(self):
        logging.info(f"NoOp With Delay: STARTED")
        raise Exception("Stuff happened")
