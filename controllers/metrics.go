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
