package integration_test

import (
	"encoding/xml"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/microsoft/azure-databricks-operator/mockapi/middleware"
	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	"github.com/microsoft/azure-databricks-operator/mockapi/router"
	"github.com/stretchr/testify/assert"
)

func unsetLatencyEnvVars() {
	_ = os.Unsetenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN")
	_ = os.Unsetenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX")
	_ = os.Unsetenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MIN")
	_ = os.Unsetenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MAX")
}

func TestAPI_Index(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Get(server.URL + "/")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
}

func TestAPI_JobList_WithFastLatency(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN", "1000")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX", "2000")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MIN", "0")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MAX", "0")

	// Act
	start := time.Now()
	response, err := server.Client().Get(server.URL + "/api/2.0/jobs/list")
	elapsed := time.Since(start)

	// Assert
	elapsedMilliseconds := int64(elapsed / time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.GreaterOrEqual(t, elapsedMilliseconds, int64(1000))
	assert.LessOrEqual(t, elapsedMilliseconds, int64(3000))

	// Cleanup
	unsetLatencyEnvVars()
}

func TestAPI_JobList_WithSlowLatency(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN", "0")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX", "0")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MIN", "1000")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MAX", "2000")

	// Act
	start := time.Now()
	response, err := createJob(server)
	elapsed := time.Since(start)

	// Assert
	elapsedMilliseconds := int64(elapsed / time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.GreaterOrEqual(t, elapsedMilliseconds, int64(1000))
	assert.LessOrEqual(t, elapsedMilliseconds, int64(3000))

	// Cleanup
	unsetLatencyEnvVars()
}
func TestAPI_JobList_WithRateLimit(t *testing.T) {
	// Arrange
	_ = os.Setenv("DATABRICKS_MOCK_API_RATE_LIMIT", "10")
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	start := time.Now()
	end := start.Add(time.Second)
	successCount := 0
	for time.Now().Before(end) {
		response, err := server.Client().Get(server.URL + "/api/2.0/jobs/list")
		if err != nil {
			t.Errorf("*** Get failed: %v\n", err)
			break
		}
		_, _ = ioutil.ReadAll(response.Body)
		_ = response.Body.Close()
		if response.StatusCode == 200 {
			successCount++
		}
	}

	// Assert
	t.Logf("Success Count %d\n", successCount)
	assert.Equal(t, 10, successCount)

	// Cleanup
	_ = os.Unsetenv("DATABRICKS_MOCK_API_RATE_LIMIT")
}

func TestAPI_JobList_Error500_20PercentRate(t *testing.T) {

	_ = os.Setenv("DATABRICKS_MOCK_API_ERROR_500_PROBABILITY", "20")
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	error500Count := 0
	for i := 0; i < 10000; i++ {
		response, err := server.Client().Get(server.URL + "/api/2.0/jobs/list")
		if err != nil {
			t.Errorf("*** Get failed: %v\n", err)
			break
		}
		_, _ = ioutil.ReadAll(response.Body)
		_ = response.Body.Close()
		if response.StatusCode == 500 {
			error500Count++
		}
	}

	// 10000 requests with a 20% probability configured for Error500 responses
	// Treat 1900-2100 inclusive as success
	assert.GreaterOrEqual(t, error500Count, 1900, "For 10000 executions with 20%% Error500 rate, expected the error rate to be 1900-2100. Got %d", error500Count)
	assert.LessOrEqualf(t, error500Count, 2100, "For 10000 executions with 20% Error500 rate, expected the error rate to be 1900-2100. Got %d", error500Count)

	_ = os.Unsetenv("DATABRICKS_MOCK_API_ERROR_500_PROBABILITY")
}

func TestAPI_JobList_Error500_100PercentRate(t *testing.T) {

	_ = os.Setenv("DATABRICKS_MOCK_API_ERROR_500_PROBABILITY", "100")

	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	error500Count := 0
	for i := 0; i < 1000; i++ {
		response, err := server.Client().Get(server.URL + "/api/2.0/jobs/list")
		if err != nil {
			t.Errorf("*** Get failed: %v\n", err)
			break
		}
		_, _ = ioutil.ReadAll(response.Body)
		_ = response.Body.Close()
		if response.StatusCode == 500 {
			error500Count++
		}
	}

	assert.Equal(t, error500Count, 1000)

	_ = os.Unsetenv("DATABRICKS_MOCK_API_ERROR_500_PROBABILITY")
}

func TestAPI_JobList_WithLatencyAndError500_EnsureLatencyIsApplied(t *testing.T) {

	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN", "1000")
	_ = os.Setenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX", "1000")
	_ = os.Setenv("DATABRICKS_MOCK_API_ERROR_500_PROBABILITY", "100")

	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	start := time.Now()
	response, err := server.Client().Get(server.URL + "/api/2.0/jobs/list")
	if err != nil {
		t.Errorf("*** Get failed: %v\n", err)
	} else {
		_, _ = ioutil.ReadAll(response.Body)
		_ = response.Body.Close()

		elapsed := time.Since(start)

		// Assert
		elapsedMilliseconds := int64(elapsed / time.Millisecond)
		assert.GreaterOrEqual(t, elapsedMilliseconds, int64(1000))
	}

	_ = os.Unsetenv("DATABRICKS_MOCK_API_ERROR_500_PROBABILITY")
	_ = os.Unsetenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN")
	_ = os.Unsetenv("DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX")
}

func TestAPI_JobList_WithSinkHole_EnsureClientRequestTimesOut(t *testing.T) {
	_ = os.Setenv("DATABRICKS_MOCK_API_ERROR_SINKHOLE_PROBABILITY", "100")
	middleware.SetErrorSinkHoleDuration(15 * time.Second)

	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	client := server.Client()

	client.Timeout = 10 * time.Second
	_, err := client.Get(server.URL + "/api/2.0/jobs/list")

	assert.NotNil(t, err)

	_ = os.Unsetenv("DATABRICKS_MOCK_API_ERROR_SINKHOLE_PROBABILITY")
}

func TestAPI_JobList_WithXMLResponse(t *testing.T) {
	_ = os.Setenv("DATABRICKS_MOCK_API_ERROR_XML_RESPONSE_PROBABILITY", "100")
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	client := server.Client()
	response, err := client.Get(server.URL + "/api/2.0/jobs/list")
	assert.Nil(t, err)

	assert.Equal(t, 200, response.StatusCode)
	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)
	var xmlResponse model.TestXMLResponse
	err = xml.Unmarshal(body, &xmlResponse)

	assert.Nil(t, err)
	assert.Equal(t, "TestJob", xmlResponse.Name)

	_ = os.Unsetenv("DATABRICKS_MOCK_API_ERROR_XML_RESPONSE_PROBABILITY")

}

func TestAPI_PageNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Get(server.URL + "/unknown")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 404, response.StatusCode)
}

func TestAPI_MethodNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Get(server.URL + jobAPI + "create")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 405, response.StatusCode)
}
