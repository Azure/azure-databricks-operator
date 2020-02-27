#!/bin/bash

# Find all already running kubectl port-forwards and kill them
ps aux | grep [k]ubectl | awk '{print $2}' | xargs kill > /dev/null 2>&1

echo "-------> Open port-forwards"
kubectl port-forward service/prom-azure-databricks-operator-grafana -n default 8080:80 &
kubectl port-forward service/prom-azure-databricks-oper-prometheus -n default 9091:9090 &
kubectl port-forward service/locust-loadtest 8089:8089 9090:9090 -n locust &

echo "Browse to locust webui   -> http://localhost:8089/"
echo "Browse to locust metrics -> http://localhost:9090/"
echo "Browse to Prometheus     -> http://localhost:9091/targets"
echo "Browse to Grafana        -> http://localhost:8080/"
