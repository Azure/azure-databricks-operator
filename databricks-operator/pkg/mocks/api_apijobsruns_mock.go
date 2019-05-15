package mock

import (
	"net/http"

	swagger "azure-databricks-operator/databricks-operator/pkg/swagger"

	"golang.org/x/net/context"
)

type MockedApiJobRuns struct {
}

func (*MockedApiJobRuns) DeleteRun(ctx context.Context, runId int32) (*http.Response, error) {
	return &http.Response{Status: "200 OK", StatusCode: 200}, nil
}
func (*MockedApiJobRuns) GetRun(ctx context.Context, runId int32) (*http.Response, error) {
	return nil, nil
}
func (*MockedApiJobRuns) ListRuns(ctx context.Context, localVarOptionals *swagger.ListRunsOpts) (swagger.ListRunsStatus, *http.Response, error) {
	return swagger.ListRunsStatus{}, nil, nil
}
func (*MockedApiJobRuns) SubmitRun(ctx context.Context, payload swagger.RunDefinition, localVarOptionals *swagger.SubmitRunOpts) (swagger.CreateJobStatus, *http.Response, error) {
	status := swagger.CreateJobStatus{}
	status.RunName = payload.RunName
	status.Result = []swagger.CreateJobResult{swagger.CreateJobResult{RunId: 42}}

	return status, nil, nil
}
