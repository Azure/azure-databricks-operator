from locust import events
from functools import wraps
import time
from locust.exception import InterruptTaskSet


def locust_client_function(func):
    """
    This decorator captures timings and fires the appropriate locust events
    """

    @wraps(func)
    def wrapper(*args, **kwargs):

        start_time = time.time()
        name = func.__name__

        try:
            response = func(*args, **kwargs)
        except InterruptTaskSet as e:
            _record_error_event(name, start_time, e)
            raise  # ensure we let Locust known exceptions bubble up
        except Exception as e:
            _record_error_event(name, start_time, e)
        else:
            total_time = _calc_time_taken(start_time)
            events.request_success.fire(
                request_type="db_client",
                name=name,
                response_time=total_time,
                response_length=0,
            )
            return response

    return wrapper


def _record_error_event(name, start_time, e):
    """
    Record the exception as a failure into the Locust event chain
    """
    total_time = _calc_time_taken(start_time)
    events.request_failure.fire(
        request_type="db_client",
        name=name,
        response_time=total_time,
        response_length=0,
        exception=e,
    )


def _calc_time_taken(start_time):
    """
    Calculate the amount of seconds elapsed between current time and start_time
    """
    return int((time.time() - start_time) * 1000)
