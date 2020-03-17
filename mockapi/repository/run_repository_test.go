package repository

import (
	"testing"
	"time"

	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	mockableClock "github.com/stephanos/clock"
	"github.com/stretchr/testify/assert"
	dbmodel "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

var testLifeCyclePeriodLength int64 = 500
var mockClock = mockableClock.NewMock()
var testTime = time.Date(2019, 01, 01, 01, 00, 00, 00, time.UTC)

var lifeCycleTests = []struct {
	name                   string
	waitTime               int64
	expectedLifeCycleState string
	expectedResultState    string
}{
	{"LifeCycleState while run is PENDING", 0, dbmodel.RunLifeCycleStatePending, ""},
	{"LifeCycleState while run is RUNNING", testLifeCyclePeriodLength, dbmodel.RunLifeCycleStateRunning, ""},
	{"LifeCycleState while run is TERMINATING", testLifeCyclePeriodLength * 2, dbmodel.RunLifeCycleStateTerminating, ""},
	{"LifeCycleState while run is TERMINATED", testLifeCyclePeriodLength * 3, dbmodel.RunLifeCycleStateTerminated, dbmodel.RunResultStateSuccess},
}

func TestGetRun_LifeCycle(t *testing.T) {
	// Arrange
	repo, runID := setupLifeCycleTests()

	for _, tt := range lifeCycleTests {
		t.Run(tt.name, func(t *testing.T) {
			mockClock.Set(testTime.Add(time.Duration(tt.waitTime) * time.Millisecond))

			// Act
			run := repo.GetRun(runID)

			//Assert
			var lifeCycleState string
			if run.State.LifeCycleState != nil {
				lifeCycleState = string(*run.State.LifeCycleState)
			}
			assert.Equal(t, tt.expectedLifeCycleState, lifeCycleState)

			var resultState string
			if run.State.ResultState != nil {
				resultState = string(*run.State.ResultState)
			}
			assert.Equal(t, tt.expectedResultState, resultState)
		})
	}
}

func TestGetRunOutput_LifeCycle(t *testing.T) {
	// Arrange
	repo, runID := setupLifeCycleTests()

	for _, tt := range lifeCycleTests {
		t.Run(tt.name, func(t *testing.T) {
			mockClock.Set(testTime.Add(time.Duration(tt.waitTime) * time.Millisecond))

			// Act
			response := repo.GetRunOutput(runID)

			//Assert
			var lifeCycleState string
			if response.Metadata.State.LifeCycleState != nil {
				lifeCycleState = string(*response.Metadata.State.LifeCycleState)
			}
			assert.Equal(t, tt.expectedLifeCycleState, lifeCycleState)

			var resultState string
			if response.Metadata.State.ResultState != nil {
				resultState = string(*response.Metadata.State.ResultState)
			}
			assert.Equal(t, tt.expectedResultState, resultState)

		})
	}
}

func TestListRun_LifeCycle(t *testing.T) {
	// Arrange
	repo, _ := setupLifeCycleTests()

	for _, tt := range lifeCycleTests {
		t.Run(tt.name, func(t *testing.T) {
			mockClock.Set(testTime.Add(time.Duration(tt.waitTime) * time.Millisecond))

			// Act
			response := repo.GetRuns()

			//Assert
			var lifeCycleState string
			if response.Runs[0].State.LifeCycleState != nil {
				lifeCycleState = string(*response.Runs[0].State.LifeCycleState)
			}
			assert.Equal(t, tt.expectedLifeCycleState, lifeCycleState)

			var resultState string
			if response.Runs[0].State.ResultState != nil {
				resultState = string(*response.Runs[0].State.ResultState)
			}
			assert.Equal(t, tt.expectedResultState, resultState)

		})
	}
}

var cancelRunTests = []struct {
	name                string
	waitTime            int64
	expectedResultState string
}{
	{"Cancel run while PENDING", 0, dbmodel.RunResultStateCanceled},
	{"Cancel run while RUNNING", testLifeCyclePeriodLength * 2, dbmodel.RunResultStateCanceled},
	{"Cancel run while TERMINATING", testLifeCyclePeriodLength * 3, dbmodel.RunResultStateCanceled},
	{"Cancel run while TERMINATED", testLifeCyclePeriodLength * 6, dbmodel.RunResultStateSuccess},
}

func TestCancelRun(t *testing.T) {
	// Arrange
	repo := NewRunRepository(testLifeCyclePeriodLength)
	mockClock.Set(testTime)
	clock = mockClock

	for i, tt := range cancelRunTests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			id := repo.CreateRun(model.JobsRunsSubmitRequest{}, int64(i))
			mockClock.Set(testTime.Add(time.Duration(tt.waitTime) * time.Millisecond))
			_ = repo.CancelRun(id)

			//Assert
			run := repo.GetRun(id)

			var resultState string
			if run.State.ResultState != nil {
				resultState = string(*run.State.ResultState)
			}
			assert.Equal(t, tt.expectedResultState, resultState)
		})
	}
}

func setupLifeCycleTests() (*RunRepository, int64) {
	mockClock.Set(testTime)
	clock = mockClock

	repo := NewRunRepository(testLifeCyclePeriodLength)
	request := model.JobsRunsSubmitRequest{}
	id := repo.CreateRun(request, 1)
	return repo, id
}
