package main

import (
	"fmt"
	"os"
	"time"

	prometheus "github.com/ryotarai/prometheus-query/client"
)

func main() {
	client, _ := prometheus.NewClient("http://localhost:9091")
	timeout := time.After(7 * time.Minute)
	tick := time.Tick(10 * time.Second)

	// Keep trying until we're timed out, got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			onFailure("Load test timed out")
		// Got a tick, we should check on doSomething()
		case <-tick:
			userCount := getQueryResult(client, "locust_user_count")
			if userCount >= 25 {
				fmt.Printf("Load test finished with user count: %v\n", userCount)
				return
			}

			fmt.Printf("User count: %v\n", userCount)
			failRatio := getQueryResult(client, "sum(locust_fail_ratio)")
			fmt.Printf("Locust failure ratio: %v\n", failRatio)
			if failRatio > 0 {
				onFailure("locust_fail_ratio is higher than 0")
			}

			numOfFailures := getQueryResult(client, "locust_requests_num_failures{path=\"poll_run_await_completion\"}")
			fmt.Printf("Number of failed locust requests: %v\n", numOfFailures)
			if numOfFailures > 0 {
				onFailure("locust_requests_num_failures is higher than 0")
			}

			databricksFailures := getQueryResult(client, "sum(databricks_request_duration_seconds_count{object_type=\"runs\", outcome=\"failure\"})")
			fmt.Printf("Number of failed databricks requests: %v\n", databricksFailures)
			if databricksFailures > 0 {
				onFailure("databricks_request_duration_seconds_count failure count is higher than 0")
			}

		}
	}
}

func getQueryResult(client *prometheus.Client, query string) float64 {
	now := time.Now()
	step, _ := time.ParseDuration("15s")

	resp, err := client.QueryRange(query, now, now, step)

	if err != nil {
		onError(err)
	}

	if len(resp.Data.Result) == 0 {
		return 0
	}

	metric, _ := resp.Data.Result[0].Values[0].Value()

	return metric
}

func onFailure(err string) {
	fmt.Println(err)
	os.Exit(99)
}

func onError(err error) {
	fmt.Println(err)
	os.Exit(1)
}
