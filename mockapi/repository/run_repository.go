package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	dbhttpmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/httpmodels"
	dbmodel "github.com/polar-rams/databricks-sdk-golang/azure/jobs/models"
)

// RunRepository is a store for Run instances
type RunRepository struct {
	runID               int64
	runs                map[int64]dbmodel.Run
	writeLock           sync.Mutex
	timePerRunLifeState int64
}

// NewRunRepository creates a new RunRepository
func NewRunRepository(timePerRunLifeState int64) *RunRepository {
	return &RunRepository{
		runs:                map[int64]dbmodel.Run{},
		timePerRunLifeState: timePerRunLifeState,
	}
}

// CreateRun adds an ID to the specified run and adds it to the collection
func (r *RunRepository) CreateRun(runReq model.JobsRunsSubmitRequest, jobID int64) int64 {

	newID := atomic.AddInt64(&r.runID, 1)
	lifeCycleState := dbmodel.RunLifeCycleState(dbmodel.RunLifeCycleStatePending)
	trigger := dbmodel.TriggerType(dbmodel.TriggerTypePeriodic)
	run := dbmodel.Run{
		RunID:       newID,
		JobID:       jobID,
		NumberInJob: int64(runReq.NewCluster.NumWorkers),
		State: dbmodel.RunState{
			LifeCycleState: lifeCycleState,
			StateMessage:   "Starting action",
		},
		ClusterInstance: dbmodel.ClusterInstance{
			ClusterID:      "1201-my-cluster",
			SparkContextID: "1102398-spark-context-id",
		},
		Task:                 dbmodel.JobTask{NotebookTask: dbmodel.NotebookTask{NotebookPath: "/Users/user@example.com/my-notebook"}},
		ClusterSpec:          dbmodel.ClusterSpec{ExistingClusterID: "1201-my-cluster"},
		OverridingParameters: dbmodel.RunParameters{JarParams: &[]string{"param1", "param2"}},
		StartTime:            makeTimestamp(),
		SetupDuration:        r.timePerRunLifeState,
		ExecutionDuration:    r.timePerRunLifeState,
		CleanupDuration:      r.timePerRunLifeState,
		Trigger:              trigger,
	}

	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	r.runs[newID] = run
	return newID
}

// GetRun returns the Run with the specified ID or an empty Run
func (r *RunRepository) GetRun(id int64) dbmodel.Run {
	if run, ok := r.runs[id]; ok {
		setRunState(&run)
		return run
	}

	return dbmodel.Run{}
}

// GetRunOutput returns the Run output along with the run as metadata or an empty run output
func (r *RunRepository) GetRunOutput(id int64) dbhttpmodel.RunsGetOutputResp {
	return dbhttpmodel.RunsGetOutputResp{
		Metadata: r.GetRun(id),
	}
}

// GetRuns returns all Runs
func (r *RunRepository) GetRuns() dbhttpmodel.RunsListResp {
	arr := []dbmodel.Run{}
	for _, run := range r.runs {
		setRunState(&run)
		arr = append(arr, run)
	}

	response := dbhttpmodel.RunsListResp{
		Runs: &arr,
	}

	return response
}

// DeleteRun deletes the run with the specified ID
func (r *RunRepository) DeleteRun(id int64) error {
	if _, ok := r.runs[id]; ok {
		delete(r.runs, id)
		return nil
	}
	return fmt.Errorf("Could not find Run with id of %d to delete", id)
}

//CancelRun cancels the run with the specified ID
func (r *RunRepository) CancelRun(id int64) error {
	if run, ok := r.runs[id]; ok {
		setRunState(&run)
		//If the run has already been completed, cancel becomes NOP
		if run.State.ResultState != "" {
			return nil
		}
		resultState := dbmodel.RunResultState(dbmodel.RunResultStateCanceled)
		run.State.ResultState = resultState
		run.State.LifeCycleState = dbmodel.RunLifeCycleStateTerminated
		r.runs[id] = run
		return nil
	}
	return fmt.Errorf("Could not find Run with id of %d to cancel", id)
}

func setRunState(run *dbmodel.Run) {
	currentTime := makeTimestamp()
	setupFinishedTime := run.StartTime + run.SetupDuration
	runFinishedTime := setupFinishedTime + run.ExecutionDuration
	cleanupFinishedTime := runFinishedTime + run.CleanupDuration

	if currentTime < setupFinishedTime || run.State.LifeCycleState == dbmodel.RunLifeCycleStateTerminated {
		return
	}

	if currentTime < runFinishedTime {
		run.State.LifeCycleState = dbmodel.RunLifeCycleStateRunning
		return
	}

	if currentTime < cleanupFinishedTime {
		run.State.LifeCycleState = dbmodel.RunLifeCycleStateTerminating
		return
	}

	run.State.LifeCycleState = dbmodel.RunLifeCycleStateTerminated
	resultState := dbmodel.RunResultState(dbmodel.RunResultStateSuccess)
	run.State.ResultState = resultState
}
