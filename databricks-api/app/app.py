from flask import Flask
from flask_restplus import Resource, Api, fields
import logging
from databricks_api import DatabricksAPI
import json
import os
from dbricks_client import dbricks_client

app = Flask(__name__)

api = Api(app,
          title='Databricks Controller',
          version='1.0',
          description='Manage Databricks')

runs_ns = api.namespace('api/jobs/runs',
                        description='Operations related to job management')


create_job_result = api.model('Create Job Result', {
    'run_id': fields.Integer
})

create_job_status = api.model('Create Job Status', {
    'run_name': fields.String(required=True,
                              example="aztest1-uppercase",
                              description='The name identifier of a job run'),
    'result': fields.List(fields.Nested(create_job_result))
})

new_cluster = api.model('new_cluster', {
    'spark_version': fields.String,
    'spark_conf': fields.Raw,
    'node_type_id': fields.String,
    'spark_env_vars': fields.Raw,
    'enable_elastic_disk': fields.Boolean,
    'num_workers': fields.Integer
})


cluster_spec = api.model('cluster_spec', {
    'new_cluster': fields.List(fields.Nested(new_cluster)),
    'libraries': fields.List(fields.Raw),
})

notebook_task = api.model('notebook_task', {
    'notebook_path': fields.String
})

run_definition = api.model('Run definition',
                           {
                               'run_name':
                               fields.String(required=True,
                                             example="aztest1-uppercase",
                                             description='The name identifier of a job run'),
                               'notebook_spec_secrets':
                               fields.Raw(required=False,
                                          example={
                                              "eventhub_source_cs": "Endpoint=sb://xxxx.servicebus.windows.net/;SharedAccessKeyName=xxxx;SharedAccessKey=xxxx=;EntityPath=sourceeh",
                                              "eventhub_destination_cs": "Endpoint=sb://xxxx.servicebus.windows.net/;SharedAccessKeyName=xxxx;SharedAccessKey=xxxx=;EntityPath=desteh",
                                              "adl2_destination_oauth2_clientid": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
                                              "adl2_destination_oauth2_clientsecret": "xxxx=",
                                              "adl2_destination_oauth2_tenantid": "xxxx=",
                                              "adl2_destination_cs": "abfss://<file-system-name>@<storage-account-name>.dfs.core.windows.net/folder1",
                                          }),
                               'notebook_spec':
                               fields.Raw(required=False,
                                          example={
                                              'TryCount': 3,
                                              'LogError': True
                                          }),
                               'notebook_additional_libraries':
                               fields.Raw(required=False,
                                          example=['t_uppercase',
                                                   't_lowercase',
                                                   't_append_a_to_attribute',
                                                   't_append_b_to_attribute']),
                               'new_cluster': fields.Nested(new_cluster),
                               'timeout_seconds': fields.Integer(required=False,
                                                                 example=3600,
                                                                 description='timeout in seconds'),
                               'notebook_task': fields.Nested(notebook_task)
                           }
                           )

state = api.model('state', {
    'life_cycle_state': fields.String,
    'result_state': fields.String,
    'state_message': fields.String,
})

cluster_instance = api.model('cluster_instance', {
    'cluster_id': fields.String})

run = api.model('Run', {
    'job_id': fields.Integer,
    'run_id': fields.Integer,
    'number_in_job': fields.Integer,
    'task': fields.Raw,
    'cluster_spec': fields.List(fields.Nested(cluster_spec)),
    'state': fields.List(fields.Nested(state)),
    'cluster_instance': fields.List(fields.Nested(cluster_instance)),
    'start_time': fields.Raw,
    'setup_duration': fields.Integer,
    'execution_duration': fields.Integer,
    'cleanup_duration': fields.Integer,
    'creator_user_name': fields.String,
    'run_name': fields.String,
    'run_page_url': fields.Url,
    'run_type': fields.String,
})

list_runs_result = api.model('List Runs Result', {
    'runs': fields.List(fields.Nested(run)),
    'has_more': fields.Boolean
})

list_runs_status = api.model('List Runs Status', {
    'result': fields.List(fields.Nested(list_runs_result))
})


@api.route('/status')
class Status(Resource):
    def get(self):
        return {'status': 'ok'}


@runs_ns.route('/')
class RunsList(Resource):
    '''Shows a list of all runs, and lets you POST to submit new run'''
    @runs_ns.doc('list_runs')
    @runs_ns.marshal_with(list_runs_status)
    def get(self):
        '''List all runs'''
        result = dbricks_client.list_runs(db=DATABRICKS_API)
        return {'result': result}

    @runs_ns.doc('submit run')
    @runs_ns.expect(run_definition)
    @runs_ns.marshal_with(create_job_status, code=201)
    def post(self):
        '''Create a new task'''

        data = api.payload
        if (data is None):
            return {
                "run_name": None,
                "result": None
            }

        run_name = data['run_name']
        notebook_spec_secrets = {}
        if 'notebook_spec_secrets' in data:
            notebook_spec_secrets = data['notebook_spec_secrets']

        notebook_spec = {}
        if 'notebook_spec' in data:
            notebook_spec = data['notebook_spec']

        if 'notebook_additional_libraries' in data:

            notebook_additional_libraries = data['notebook_additional_libraries']

        checkpoint_location = "dbfs:/checkpoints/" + run_name + "/output"

        # Read job config
        with open(JOB_CONFIG_PATH) as job_config_file:
            job_config = json.load(job_config_file)

        # Set job config
        job_config["run_name"] = run_name

        # Add additional library to libraries
        libraries = job_config["libraries"]

        for lib in notebook_additional_libraries:
            libraries.append({
                lib['type']: lib['properties']
            })

            # All transformer packages require .transformer module w/ transform func
            # transformer_names.append(transformer + ".transformer")

        if "timeout_seconds" in data:
            timeout_seconds = data["timeout_seconds"]
            if (timeout_seconds is not None and type(timeout_seconds) is int):
                job_config["timeout_seconds"] = timeout_seconds

        if "new_cluster" in data:
            new_cluster = data["new_cluster"]
            for key in new_cluster:
                value = new_cluster[key]
                job_config["new_cluster"][key] = value

        if "notebook_task" in data:
            notebook_task = data["notebook_task"]
            for key in notebook_task:
                value = notebook_task[key]
                job_config["notebook_task"][key] = value

        result = dbricks_client.create_job(db=DATABRICKS_API,
                                           run_name=run_name,
                                           notebook_spec_secrets=notebook_spec_secrets,
                                           notebook_spec=notebook_spec,
                                           checkpoint_location=checkpoint_location,
                                           job_config=job_config)
        return {'run_name': run_name, 'result': result}


@runs_ns.response(404, 'run not found')
@runs_ns.param('run_id', 'The run identifier')
@runs_ns.route('/<int:run_id>')
class Run(Resource):
    '''Show a single run job item and lets you delete it'''
    @runs_ns.doc('get run')
    # @jobs_ns.marshal_with(run_status)
    def get(self, run_id):
        print(run_id)
        result = dbricks_client.get_run(db=DATABRICKS_API, run_id=run_id)
        return {'result': result}

    @runs_ns.doc('delete run')
    # @runs_ns.response(204, 'run deleted')
    def delete(self, run_id):
        '''Delete a run given its identifier'''
        print(run_id)
        cancel_run_result, delete_run_result = dbricks_client.cancel_and_delete_run(
            db=DATABRICKS_API, run_id=run_id)
        return {'run_id': run_id, 'result': delete_run_result}


if __name__ == '__main__':
    log_fmt = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    logging.basicConfig(level=logging.INFO, format=log_fmt)

    JOB_CONFIG_PATH = os.path.join(os.path.dirname(
        __file__), "config", "job.config.json")

    # Provide a host and token for connecting to DataBricks
    DATABRICKS_HOST = os.getenv("DATABRICKS_HOST")
    DATABRICKS_TOKEN = os.getenv("DATABRICKS_TOKEN")

    PYPI_INDEX_URL = os.getenv("PYPI_INDEX_URL")

    DATABRICKS_API = DatabricksAPI(
        host=DATABRICKS_HOST,
        token=DATABRICKS_TOKEN
    )

    app.run(host='0.0.0.0')
