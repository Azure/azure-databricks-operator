import uuid
import time
import pytest
from locust import events
from db_locust.constant import RUN_RESOURCE, K8_GROUP, VERSION
from db_locust.db_run_client import DbRunClient
from kubernetes import client
from kubernetes.client.rest import ApiException
from locust.exception import InterruptTaskSet


def test_create_run(mocker):
    expectedRunName = "run-wibble"

    mocker.patch.dict(RUN_RESOURCE, {"metadata": {"name": "xyz"}}, clear=True)

    mocker.patch("uuid.uuid4")
    uuid.uuid4.return_value = "wibble"

    stub = mocker.stub(name="api")
    stub.create_namespaced_custom_object = mocker.stub(name="create_run_stub")

    db_run_client = DbRunClient(stub)
    result = db_run_client.create_run()

    stub.create_namespaced_custom_object.assert_called_once_with(
        group=K8_GROUP,
        version=VERSION,
        namespace="default",
        plural="runs",
        body={"metadata": {"name": expectedRunName}},
    )
    assert result == expectedRunName


def test_create_run_throws_when_create_object_fails(mocker):
    mock = mocker.Mock()
    mock.create_namespaced_custom_object = mocker.Mock()
    mock.create_namespaced_custom_object.side_effect = ApiException("failed to create")

    db_run_client = DbRunClient(mock)

    with pytest.raises(InterruptTaskSet):
        db_run_client.create_run()


def test_delete_run(mocker):
    run_name = "test_run_wibble"

    stub = mocker.stub(name="api")
    stub.delete_namespaced_custom_object = mocker.stub(name="delete_run_stub")

    db_run_client = DbRunClient(stub)
    db_run_client.delete_run(run_name)

    stub.delete_namespaced_custom_object.assert_called_once_with(
        group=K8_GROUP,
        version=VERSION,
        namespace="default",
        plural="runs",
        name=run_name,
        body=client.V1DeleteOptions(),
    )


def test_delete_run_throws_when_delete_object_fails(mocker):
    run_name = "test_run_wibble"

    mock = mocker.Mock()
    mock.delete_namespaced_custom_object = mocker.Mock()
    mock.delete_namespaced_custom_object.side_effect = ApiException("doesn't exist")

    db_run_client = DbRunClient(mock)

    with pytest.raises(InterruptTaskSet):
        db_run_client.delete_run(run_name)


def test_get_run(mocker):
    run_name = "test_run_wibble"

    mock = mocker.Mock()
    mock.get_namespaced_custom_object = mocker.Mock()
    mock.get_namespaced_custom_object.return_value = {"name": run_name}

    db_run_client = DbRunClient(mock)
    result = db_run_client.get_run(run_name)

    mock.get_namespaced_custom_object.assert_called_once_with(
        group=K8_GROUP,
        version=VERSION,
        namespace="default",
        plural="runs",
        name=run_name,
    )

    assert result == {"name": run_name}


def test_get_run_throws_when_get_object_fails(mocker):
    run_name = "test_run_wibble"

    mock = mocker.Mock()
    mock.get_namespaced_custom_object = mocker.Mock()
    mock.get_namespaced_custom_object.side_effect = ApiException("doesn't exist")

    db_run_client = DbRunClient(mock)

    with pytest.raises(InterruptTaskSet):
        db_run_client.get_run(run_name)


def test_poll_run_await_completion_polls_until_complete(mocker):
    run_name = "test_run_poll_stops_when_run_complete"

    mocker.patch("time.sleep")

    mock = mocker.Mock()
    mock.get_namespaced_custom_object = mocker.Mock()
    mock.get_namespaced_custom_object.side_effect = [
        {},
        {"wibble": "wobble"},
        {"status": {"metadata": {"state": {"life_cycle_state": "wibble"}}}},
        {"status": {"metadata": {"state": {"life_cycle_state": "TERMINATED"}}}},
    ]

    db_run_client = DbRunClient(mock)
    db_run_client.poll_run_await_completion(run_name, 5, 0.1)

    mock.get_namespaced_custom_object.assert_called_with(
        group=K8_GROUP,
        version=VERSION,
        namespace="default",
        plural="runs",
        name=run_name,
    )
    assert time.sleep.call_count == 3
    assert mock.get_namespaced_custom_object.call_count == 4


def test_poll_run_await_completion_polls_throws_exception_on_invalid(mocker):
    run_name = "test_run_poll_stops_when_run_complete"

    completed_life_states = ["INTERNAL_ERROR", "SKIPPED"]
    for life_state in completed_life_states:
        mock = mocker.Mock()
        mock.get_namespaced_custom_object = mocker.Mock()
        mock.get_namespaced_custom_object.side_effect = [
            {},
            {"wibble": "wobble"},
            {"status": {"metadata": {"state": {"life_cycle_state": "wibble"}}}},
            {"status": {"metadata": {"state": {"life_cycle_state": life_state}}}},
        ]

        stub = mocker.stub(name="locust_event_fired")
        events.request_failure = stub
        events.request_failure.fire = mocker.stub(name="fire_stub")

        db_run_client = DbRunClient(mock)
        db_run_client.poll_run_await_completion(run_name, 5, 0.1)

        mock.get_namespaced_custom_object.assert_called_with(
            group=K8_GROUP,
            version=VERSION,
            namespace="default",
            plural="runs",
            name=run_name,
        )

        assert mock.get_namespaced_custom_object.call_count == 4

        events.request_failure.fire.assert_called_once_with(
            request_type="db_client",
            name="poll_run_await_completion",
            response_length=0,
            response_time=mocker.ANY,
            exception=mocker.ANY,
        )


def test_poll_run_await_completion_fails_when_run_fails_to_complete_in_time(mocker):
    run_name = "test_run_wibble"

    stub = mocker.stub(name="locust_event_fired")
    mocker.patch("locust.events")
    events.request_failure = stub
    events.request_failure.fire = mocker.stub(name="fire_stub")

    mock = mocker.Mock()
    mock.get_namespaced_custom_object = mocker.Mock()
    mock.get_namespaced_custom_object.return_value = {
        "status": {"metadata": {"state": {"life_cycle_state": "wibble"}}}
    }

    db_run_client = DbRunClient(mock)
    db_run_client.poll_run_await_completion(run_name, 2, 0.1)

    events.request_failure.fire.assert_called_once_with(
        request_type="db_client",
        name="poll_run_await_completion",
        response_length=0,
        response_time=mocker.ANY,
        exception=mocker.ANY,
    )
