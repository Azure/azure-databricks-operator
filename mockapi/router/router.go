package router

import (
	"net/http"

	"github.com/microsoft/azure-databricks-operator/mockapi/handler"
	"github.com/microsoft/azure-databricks-operator/mockapi/middleware"
	"github.com/microsoft/azure-databricks-operator/mockapi/repository"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const apiPrefix = "/api/2.0/"

var phm = initPrometheusHTTPMetric("databricks_mock_api", prometheus.LinearBuckets(0, 5, 20))

// NewRouter creates a new Router with mappings configured
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Routes registered directly here don't have middleware applied
	router.Name("Index").Methods("GET").Path("/").Handler(http.HandlerFunc(handler.Index))
	router.Name("ConfigGet").Methods("GET").Path("/config").Handler(http.HandlerFunc(handler.GetConfig()))
	router.Name("ConfigSet").Methods("PUT").Path("/config").Handler(http.HandlerFunc(handler.SetConfig()))
	router.Name("ConfigPatch").Methods("PATCH").Path("/config").Handler(http.HandlerFunc(handler.PatchConfig()))
	router.Name("Metrics").Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	// Build up API Routes and apply middleware
	jobRepo := repository.NewJobRepository()
	routes := append(getJobRoutes(jobRepo), getClusterRoutes()...)
	routes = append(routes, getRunRoutes(jobRepo)...)

	for _, route := range routes {
		promMiddleware := phm.createHandlerWrapper(route.TypeLabel, route.ActionLabel)
		handler := middleware.Add(route.HandlerFunc, promMiddleware, middleware.RateLimit, middleware.AddLatency, middleware.ErrorResponse)
		router.
			Name(route.Name).
			Methods(route.Method).
			Path(route.Pattern).
			Handler(handler).
			Queries(route.Queries...)
	}

	router.NotFoundHandler = http.HandlerFunc(handler.NotFoundPage)

	router.MethodNotAllowedHandler = http.HandlerFunc(handler.MethodNotAllowed)

	return router
}

func getRunRoutes(jobRepo *repository.JobRepository) Routes {
	runRepo := repository.NewRunRepository(1500)

	return Routes{
		Route{
			"RunsSubmit",
			"runs",
			"submit",
			"POST",
			apiPrefix + "jobs/runs/submit",
			http.HandlerFunc(handler.SubmitRun(runRepo, jobRepo)),
			nil,
		},
		Route{
			"RunsList",
			"runs",
			"list",
			"GET",
			apiPrefix + "jobs/runs/list",
			http.HandlerFunc(handler.ListRuns(runRepo)),
			nil,
		},
		Route{
			"RunsGet",
			"runs",
			"get",
			"GET",
			apiPrefix + "jobs/runs/get",
			http.HandlerFunc(handler.GetRun(runRepo)),
			[]string{"run_id", "{run_id:[0-9]+}"},
		},
		Route{
			"RunsGetOutput",
			"runs",
			"get_output",
			"GET",
			apiPrefix + "jobs/runs/get-output",
			http.HandlerFunc(handler.GetRunOutput(runRepo)),
			[]string{"run_id", "{run_id:[0-9]+}"},
		},
		Route{
			"RunsDelete",
			"runs",
			"delete",
			"POST",
			apiPrefix + "jobs/runs/delete",
			http.HandlerFunc(handler.DeleteRun(runRepo)),
			nil,
		},
		Route{
			"RunsCancel",
			"runs",
			"cancel",
			"POST",
			apiPrefix + "jobs/runs/cancel",
			http.HandlerFunc(handler.CancelRun(runRepo)),
			nil,
		},
	}
}

func getJobRoutes(jobRepo *repository.JobRepository) Routes {
	return Routes{
		Route{
			"JobsCreate",
			"jobs",
			"create",
			"POST",
			apiPrefix + "jobs/create",
			http.HandlerFunc(handler.CreateJob(jobRepo)),
			nil,
		},
		Route{
			"JobsList",
			"jobs",
			"list",
			"GET",
			apiPrefix + "jobs/list",
			http.HandlerFunc(handler.ListJobs(jobRepo)),
			nil,
		},
		Route{
			"JobsGet",
			"jobs",
			"get",
			"GET",
			apiPrefix + "jobs/get",
			http.HandlerFunc(handler.GetJob(jobRepo)),
			[]string{"job_id", "{job_id:[0-9]+}"},
		},
		Route{
			"JobsDelete",
			"jobs",
			"delete",
			"POST",
			apiPrefix + "jobs/delete",
			http.HandlerFunc(handler.DeleteJob(jobRepo)),
			nil,
		},
	}
}

func getClusterRoutes() Routes {
	clusterRepo := repository.NewClusterRepository()
	return Routes{
		Route{
			"ClustersCreate",
			"clusters",
			"create",
			"POST",
			apiPrefix + "clusters/create",
			http.HandlerFunc(handler.CreateCluster(clusterRepo)),
			nil,
		},
		Route{
			"ClustersList",
			"clusters",
			"list",
			"GET",
			apiPrefix + "clusters/list",
			http.HandlerFunc(handler.ListClusters(clusterRepo)),
			nil,
		},
		Route{
			"ClustersGet",
			"clusters",
			"get",
			"GET",
			apiPrefix + "clusters/get",
			http.HandlerFunc(handler.GetCluster(clusterRepo)),
			[]string{"job_id", "{job_id}"},
		},
		Route{
			"ClustersEdit",
			"clusters",
			"edit",
			"POST",
			apiPrefix + "clusters/edit",
			http.HandlerFunc(handler.EditCluster(clusterRepo)),
			nil,
		},
		Route{
			"ClustersDelete",
			"clusters",
			"delete",
			"POST",
			apiPrefix + "clusters/delete",
			http.HandlerFunc(handler.DeleteCluster(clusterRepo)),
			nil,
		},
	}
}

// Route represents an API route/endpoint
type Route struct {
	Name        string
	TypeLabel   string
	ActionLabel string
	Method      string
	Pattern     string
	HandlerFunc http.Handler
	Queries     []string
}

// Routes is an array of Routes
type Routes []Route
