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

	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	"github.com/microsoft/azure-databricks-operator/mockapi/router"
	"github.com/stretchr/testify/assert"
	dbmodel "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

const jobFileLocation = "test_data/job/job_create.json"
const jobAPI = "/api/2.0/jobs/"

func createJob(server *httptest.Server) (*http.Response, error) {
	jsonFile, _ := os.Open(jobFileLocation)
	return server.Client().Post(
		server.URL+jobAPI+"create",
		"application/json",
		jsonFile)
}

func TestAPI_JobsCreate(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := createJob(server)

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var jobResponse dbmodel.Job
	err = json.Unmarshal(body, &jobResponse)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), jobResponse.JobID)
}

func TestAPI_JobsCreate_ConcurrentRequest(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	const numberToAdd = 50

	var wg sync.WaitGroup
	wg.Add(numberToAdd)

	// Act
	for index := 0; index < numberToAdd; index++ {
		go func() {
			_, err := createJob(server)
			if err != nil {
				t.Error(err)
			}

			wg.Done()
		}()
	}

	wg.Wait()

	// Assert
	response, err := server.Client().Get(server.URL + jobAPI + "list")

	body, _ := ioutil.ReadAll(response.Body)
	var listResponse model.JobsListResponse
	_ = json.Unmarshal(body, &listResponse)

	assert.Nil(t, err)

	assert.Equal(t, numberToAdd, len(listResponse.Jobs))
	assert.Equal(t, 200, response.StatusCode)
}

func TestAPI_JobsGet(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	_, _ = createJob(server)

	// Act
	response, err := server.Client().Get(server.URL + jobAPI + "get?job_id=1")

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var jobCreateResponse dbmodel.Job
	err = json.Unmarshal(body, &jobCreateResponse)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), jobCreateResponse.JobID)
	assert.Equal(t, "Nightly model training", jobCreateResponse.Settings.Name)

}

func TestAPI_JobsListTwoJobs(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	_, _ = createJob(server)
	_, _ = createJob(server)

	// Act
	response, err := server.Client().Get(server.URL + jobAPI + "list")

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var listResponse model.JobsListResponse
	err = json.Unmarshal(body, &listResponse)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(listResponse.Jobs))
}

func TestAPI_JobsList_EmptyList(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Get(server.URL + jobAPI + "list")

	// Assert
	assert.Equal(t, 200, response.StatusCode)
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	var listResponse model.JobsListResponse
	err = json.Unmarshal(body, &listResponse)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(listResponse.Jobs))
}

func TestAPI_JobsDelete(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()
	response, err := createJob(server)

	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	// Act
	response, err = server.Client().Post(
		server.URL+jobAPI+"delete",
		"application/json",
		bytes.NewBufferString("{\"job_id\":1}"))

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
}

func TestAPI_DeleteJob_WithInvalidJobID(t *testing.T) {
	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	response, err := server.Client().Post(
		server.URL+jobAPI+"delete",
		"application/json",
		bytes.NewBufferString("{\"job_id\":1}"))

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 404, response.StatusCode)
}
