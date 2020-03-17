package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/microsoft/azure-databricks-operator/mockapi/router"
	"github.com/stretchr/testify/assert"
	azure "github.com/xinsnake/databricks-sdk-golang/azure"
	dbmodel "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

const runSubmitFileLocation = "test_data/run/run_submit.json"
const runAPI = "/api/2.0/jobs/runs/"

func submitRun(server *httptest.Server) (*http.Response, error) {
	jsonFile, _ := os.Open(runSubmitFileLocation)
	return server.Client().Post(
		server.URL+runAPI+"submit",
		"application/json",
		jsonFile)
}

func TestAPI_RunsSubmit(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := submitRun(server)

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var runResponse dbmodel.Run
	err = json.Unmarshal(body, &runResponse)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), runResponse.RunID)
}

func TestAPI_RunsSubmit_JobIsCreated(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	_, _ = submitRun(server)

	// Assert
	response, err := server.Client().Get(server.URL + jobAPI + "get?job_id=1")
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var job dbmodel.Job
	err = json.Unmarshal(body, &job)
	assert.Nil(t, err)

	assert.Equal(t, int64(1), job.JobID)
	assert.Equal(t, "5.3.x-scala2.11", job.Settings.NewCluster.SparkVersion)
	assert.Equal(t, "com.databricks.ComputeModels", job.Settings.SparkJarTask.MainClassName)
	assert.Equal(t, "dbfs:/my-jar.jar", job.Settings.Libraries[0].Jar)

}

func TestAPI_RunsSubmit_Concurrently(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	const numberToAdd = 50
	var wg sync.WaitGroup
	wg.Add(numberToAdd)

	// Act
	for index := 0; index < numberToAdd; index++ {
		go func() {
			_, err := submitRun(server)
			if err != nil {
				t.Error(err)
			}

			wg.Done()
		}()
	}
	wg.Wait()

	// Assert
	response, err := server.Client().Get(server.URL + runAPI + "list")
	assert.Equal(t, 200, response.StatusCode)

	body, _ := ioutil.ReadAll(response.Body)
	var runsListResponse azure.JobsRunsListResponse
	_ = json.Unmarshal(body, &runsListResponse)
	assert.Nil(t, err)
	assert.Equal(t, numberToAdd, len(runsListResponse.Runs))
}

func TestAPI_RunsGet(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()
	_, _ = submitRun(server)

	// Act
	response, _ := server.Client().Get(server.URL + runAPI + "get?run_id=1")

	// Assert
	assert.Equal(t, 200, response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var runResponse dbmodel.Run
	err = json.Unmarshal(body, &runResponse)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), runResponse.RunID)
}

func TestAPI_RunsGetWithNoRuns(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, _ := server.Client().Get(server.URL + runAPI + "get?run_id=1")

	// Assert
	assert.Equal(t, 404, response.StatusCode)
}

func TestAPI_RunsGetOutput(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()
	_, _ = submitRun(server)

	// Act
	response, _ := server.Client().Get(server.URL + runAPI + "get-output?run_id=1")

	// Assert
	assert.Equal(t, 200, response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var runResponse azure.JobsRunsGetOutputResponse
	err = json.Unmarshal(body, &runResponse)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), runResponse.Metadata.RunID)
}

func TestAPI_RunsGetOutputWithNoRuns(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, _ := server.Client().Get(server.URL + runAPI + "get-output?run_id=1")

	// Assert
	assert.Equal(t, 404, response.StatusCode)
}

func TestAPI_RunsListWithTwoRuns(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	_, _ = submitRun(server)
	_, _ = submitRun(server)

	// Act
	response, err := server.Client().Get(server.URL + runAPI + "list")

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var runsListResponse azure.JobsRunsListResponse
	err = json.Unmarshal(body, &runsListResponse)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(runsListResponse.Runs))
	assert.Equal(t, false, runsListResponse.HasMore)
}

func TestAPI_RunsListWithEmptyList(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Get(server.URL + runAPI + "list")

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var runsListResponse azure.JobsRunsListResponse
	err = json.Unmarshal(body, &runsListResponse)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(runsListResponse.Runs))
}

func TestAPI_RunsCancel(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()
	response, err := submitRun(server)

	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	// Act
	response, err = server.Client().Post(
		server.URL+runAPI+"cancel",
		"application/json",
		bytes.NewBufferString("{\"run_id\":1}"))

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
}

func TestAPI_RunsCancel_WithInvalidJobID(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Post(
		server.URL+runAPI+"cancel",
		"application/json",
		bytes.NewBufferString("{\"run_id\":1}"))

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 404, response.StatusCode)
}

func TestAPI_RunsDelete(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()
	response, err := submitRun(server)

	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	// Act
	response, err = server.Client().Post(
		server.URL+runAPI+"delete",
		"application/json",
		bytes.NewBufferString("{\"run_id\":1}"))

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
}

func TestAPI_RunsDelete_WithInvalidJobID(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Post(
		server.URL+runAPI+"delete",
		"application/json",
		bytes.NewBufferString("{\"run_id\":1}"))

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 404, response.StatusCode)
}
