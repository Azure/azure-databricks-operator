# -*- coding: utf-8 -*-

"""Main module."""
from typing import Dict
from databricks_api import DatabricksAPI


def list_runs(db: DatabricksAPI,
              job_id=None,
              active_only=None,
              completed_only=None,
              offset=None,
              limit=None):
    """
    List runs from most recently started to least.
    :param DatabricksAPI db:
        DatabricksAPI with host and token
    :param int job_id:
        The job for which to list runs. If omitted,
        the Jobs service will list runs from all jobs.
    :param int job_id:
        The job for which to list runs.
        If omitted, the Jobs service will list runs from all jobs.
    :param bool active_only:
        If active_only, if true, only active runs will be included in the results;
        otherwise, lists both active and completed runs.
    :param bool completed_only:
        If completed_only, if true, only completed runs will be included in the results;
        otherwise, lists both active and completed runs.
        This field cannot be true when active_only is true.
    :param int offset:
        The offset of the first run to return, relative to the most recent run.
    :param int limit:
        The number of runs to return.
        This value should be greater than 0 and less than 1000
    """
    result = db.jobs.list_runs(
        job_id=job_id,
        active_only=active_only,
        completed_only=completed_only,
        offset=offset,
        limit=limit
    )
    return result


def get_run(db: DatabricksAPI, run_id: int):
    """
    Retrieve the metadata of a run.
    :param DatabricksAPI db:
        DatabricksAPI with host and token
    :param int run_id:
        databricks runs id
    """
    try:
        result = db.jobs.get_run(run_id=run_id)
        return result
    except Exception as ex:
        print(ex)
        return None


def cancel_run(db: DatabricksAPI, run_id: int):
    """
    Cancel a run in databricks

    :param DatabricksAPI db:
        DatabricksAPI with host and token
    :param int run_id:
        databricks runs id
    """
    try:
        result = db.jobs.cancel_run(run_id=run_id)
        return result
    except Exception as ex:
        print(ex)
        return None


def cancel_and_delete_run(db: DatabricksAPI, run_id: int):
    """
    Cancel and then delete a run in databricks

    :param DatabricksAPI db:
        DatabricksAPI with host and token
    :param int run_id:
        databricks runs id
    """
    try:
        cancel_run_result = db.jobs.cancel_run(run_id=run_id)
        delete_run_result = db.jobs.delete_run(run_id=run_id)
        return cancel_run_result, delete_run_result
    except Exception as ex:
        print(ex)
        return None


def put_secret(db: DatabricksAPI,
               secret_scope: str,
               key: str,
               string_value: str):
    if (string_value is not None):
        db.secret.put_secret(secret_scope, key=key, string_value=string_value)


def create_job(db: DatabricksAPI,
               run_name: str,
               notebook_spec_secrets: Dict,
               notebook_spec: Dict,
               checkpoint_location: str,
               job_config: Dict) -> Dict[str, int]:
    """
    Create a databricks streaming job

    :param DatabricksAPI db:
        DatabricksAPI with host and token
    :param str run_name:
        Name of job in databricks, each run_name should be unique
    :param dict notebook_spec_secrets
        Dictionary of Key,Values to pass to Databricks notebook by using databrick secrets (encrypted)
    :param dict notebook_spec
        Dictionary of Key,Values to pass to Databricks notebook by widgets
    :param str checkpoint_location:
        location of checkpoints for the job, it should be unique
    :param dict job_config
        job configuration
    """

    try:
        # Create secrets
        secret_scope = run_name + "_scope"
        create_or_update_secret_scope(db, secret_scope)

        for key in notebook_spec_secrets:
            value = notebook_spec_secrets[key]
            put_secret(db, secret_scope, key=key, string_value=value)

        # 1. You can set a cluster environment variable
        # job_config["new_cluster"]["spark_env_vars"]["SECRET_SCOPE"]=secret_scope
        # 2. OR pass it in via databricks widgets / notebooks base_parameters
        job_config["notebook_task"]["base_parameters"] = [
            {"key": "run_name", "value": run_name},
            {"key": "checkpoint_location", "value": checkpoint_location}
        ]

        for key in notebook_spec:
            value = notebook_spec[key]
            job_config["notebook_task"]["base_parameters"].append({"key": key, "value": value})

        # Create job
        # Todo: log potential errors
        result = db.jobs.submit_run(**job_config)
        return result
    except Exception as ex:
        print(ex)
        return None


def create_or_update_secret_scope(db: DatabricksAPI, secret_scope: str):

    try:
        scope_results = db.secret.list_scopes()

        if "scopes" not in scope_results:
            db.secret.create_scope(secret_scope)
            return

        scopes = scope_results["scopes"]
        if not any(d['name'] == secret_scope for d in scopes):
            db.secret.create_scope(secret_scope)
    except Exception as ex:
        print(ex)
        return None
