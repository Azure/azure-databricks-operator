K8_GROUP = "databricks.microsoft.com"

VERSION = "v1alpha1"

RUN_RESOURCE = {
    "apiVersion": "databricks.microsoft.com/v1alpha1",
    "kind": "Run",
    "metadata": {"name": "run-sample"},
    "spec": {
        "name": "LocustRun",
        "new_cluster": {
            "spark_version": "5.3.x-scala2.11",
            "node_type_id": "Standard_D3_v2",
            "num_workers": 1,
        },
        "spark_submit_task": {
            "parameters": [
                "--class",
                "org.apache.spark.examples.SparkPi",
                # For tests against the real databricks API, change this to a valid file location
                "dbfs:/FileStore/tables/SparkPi_assembly_0_1-04ede.jar",
                "1",
            ]
        },
    },
}
