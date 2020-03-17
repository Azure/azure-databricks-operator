package integration_test

import (
	"net/http/httptest"
	"testing"

	"github.com/microsoft/azure-databricks-operator/mockapi/router"
	"github.com/stretchr/testify/assert"
)

func TestAPI_Metrics(t *testing.T) {
	//Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	//Act
	response, err := server.Client().Get(server.URL + "/metrics")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
}
