/*
Copyright 2019 microsoft.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import(
	"time"
	"github.com/prometheus/client_golang/prometheus"
	models "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

func trackJobExecutionTime(histogram prometheus.Histogram, f func() (models.Job, error)) (models.Job, error) {
	startTime := time.Now()

	defer trackMetric(startTime, histogram)

	return f()
}

func trackExecutionTime(histogram prometheus.Histogram, f func() error) error {
	startTime := time.Now()

	defer trackMetric(startTime, histogram)

	return f()
}

func trackMetric(startTime time.Time, histogram prometheus.Histogram) {
	endTime := float64(time.Now().Sub(startTime).Nanoseconds() / int64(time.Millisecond))
	histogram.Observe(endTime)
}
