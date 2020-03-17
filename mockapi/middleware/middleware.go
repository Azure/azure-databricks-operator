package middleware

import (
	"encoding/xml"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/microsoft/azure-databricks-operator/mockapi/model"
	"golang.org/x/time/rate"
)

// LatencySlowRequestMaxEnvName is the name of the env var for the max latency setting of a slow request Eg: Post request
const LatencySlowRequestMaxEnvName = "DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MAX"

// LatencySlowRequestMinEnvName is the name of the env var for the min latency setting of a slow request Eg: Post request
const LatencySlowRequestMinEnvName = "DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MIN"

// LatencyFastRequestMaxEnvName is the name of the env var for the max latency setting of a fast request Eg: Get request
const LatencyFastRequestMaxEnvName = "DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX"

// LatencyFastRequestMinEnvName is the name of the env var for the min latency setting of a fast request Eg: Get request
const LatencyFastRequestMinEnvName = "DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN"

// RateLimitEnvName is the name of the env var for the rate limit setting
const RateLimitEnvName = "DATABRICKS_MOCK_API_RATE_LIMIT"
const rateLimitDefault = rate.Inf
const rateBurstAmount = 10 // Burst of 10 seems to work from testing with a client

// Error500ResponseEnvName is the name of the env var for specifying the probability of returning a 500 response.
// E.g. 20 => 20% chance of returning a 500 response
const Error500ResponseEnvName = "DATABRICKS_MOCK_API_ERROR_500_PROBABILITY"

// ErrorSinkHoleResponseEnvName is name of the env var for specifying the probability of a sink-hole response.
// A sink-hole response is a request that writes nothing as part of the response for an extended period (5 minutes)
// and holds open the connection during that time.
// E.g. 20 => 20% chance of returning a sink-hole  response
const ErrorSinkHoleResponseEnvName = "DATABRICKS_MOCK_API_ERROR_SINKHOLE_PROBABILITY"
const errorProbabilityDefault = 0
const errorSinkHoleDurationDefault = 10 * time.Minute

//ErrorXMLResponseEnvName is the name of the env var for specifying status code of 200 but with XML body
const ErrorXMLResponseEnvName = "DATABRICKS_MOCK_API_ERROR_XML_RESPONSE_PROBABILITY"

func init() {
	// From the docs, Seed should not be called concurrently with any other Rand method
	// so call Seed here to randomize on initialization rather than per-request
	rand.Seed(time.Now().UnixNano())
}

// Add applies a set of middleware to a http.Handler.
// Middleware is applied so that it executes in the order specified.
// E.g. middleware.Add(h, mw1, mw2) is equivalent to mw2(mw1(h))
func Add(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	// Reverse index to execute middleware in provided order
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}

	return h
}

// AddLatency applies random latency to the request
func AddLatency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			// Fast Request Sleep
			fastRequestMin, fastRequestMax, err := getLatencyValues(LatencyFastRequestMinEnvName, LatencyFastRequestMaxEnvName)
			if err == nil {
				log.Printf("Adding latency for GET requests between %v-%v ms", fastRequestMin, fastRequestMax)
				fastRequestSleepDuration := fastRequestMin + rand.Intn(fastRequestMax-fastRequestMin+1)
				time.Sleep(time.Duration(fastRequestSleepDuration) * time.Millisecond)
			}
		default:
			// Slow Request Sleep
			slowRequestMin, slowRequestMax, err := getLatencyValues(LatencySlowRequestMinEnvName, LatencySlowRequestMaxEnvName)
			if err == nil {
				log.Printf("Adding latency for default POST/PUT/DELETE requests between %v-%v ms", slowRequestMin, slowRequestMax)
				slowRequestSleepDuration := slowRequestMin + rand.Intn(slowRequestMax-slowRequestMin+1)
				time.Sleep(time.Duration(slowRequestSleepDuration) * time.Millisecond)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func getLatencyValues(minName string, maxName string) (minValue int, maxValue int, errorValue error) {
	requestMin, minErr := strconv.Atoi(os.Getenv(minName))
	requestMax, maxErr := strconv.Atoi(os.Getenv(maxName))

	// test minErr, maxErr and log + return error if either not nil
	if minErr != nil || maxErr != nil {
		log.Printf("Env variables are not set: %s, %s", minName, maxName)
		// errLatencyNotSet Env variable set
		errLatencyNotSet := errors.New("Env variables are not set ")
		return 0, 0, errLatencyNotSet
	}
	return requestMin, requestMax, nil

}

var limiter = rate.NewLimiter(rateLimitDefault, rateBurstAmount)

// RateLimit applies rate-limiting to the requests
func RateLimit(handler http.Handler) http.Handler {

	setRateLimit := func() {
		configValue, rateErr := strconv.Atoi(os.Getenv(RateLimitEnvName))
		var configRate rate.Limit
		if rateErr == nil {
			configRate = rate.Limit(configValue)
		} else {
			configRate = rateLimitDefault
		}
		if configRate != limiter.Limit() {
			log.Printf("%s has changed - updating limit to %.0f\n", RateLimitEnvName, configRate)
			limiter.SetLimit(rate.Limit(configRate))
		}
	}
	setRateLimit()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setRateLimit()

		if limiter.Allow() {
			handler.ServeHTTP(w, r)
		} else {
			log.Printf("429 Response: %s\n", r.RequestURI)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		}
	})
}

var error500Probabilty = -1
var errorSinkHoleProbability = -1
var errorXMLResponseProbability = -1
var errorSinkHoleDuration = errorSinkHoleDurationDefault

// GetErrorSinkHoleDuration returns the duration used for sink-hole requests
func GetErrorSinkHoleDuration() time.Duration {
	return errorSinkHoleDuration
}

// SetErrorSinkHoleDuration sets the duration used for sink-hole requests
func SetErrorSinkHoleDuration(duration time.Duration) {
	errorSinkHoleDuration = duration
}

// ErrorResponse injects error behavior based on config
func ErrorResponse(handler http.Handler) http.Handler {
	getConfigValue := func(name string) int {
		configValue, err := strconv.Atoi(os.Getenv(name))
		if err != nil {
			configValue = errorProbabilityDefault
		}
		if configValue < 0 || configValue > 100 {
			log.Printf("Invalid value for %s (%d). Using default of %d\n", name, configValue, errorProbabilityDefault)
			configValue = errorProbabilityDefault
		}
		return configValue
	}
	setErrorLimit := func() {
		error500ConfigValue := getConfigValue(Error500ResponseEnvName)
		errorSinkHoleConfigValue := getConfigValue(ErrorSinkHoleResponseEnvName)
		errorXMLResponseConfigValue := getConfigValue(ErrorXMLResponseEnvName)
		if error500ConfigValue+errorSinkHoleConfigValue+errorXMLResponseConfigValue > 100 {
			log.Printf("Invalid value for error rates - must add up to 100 or less. Using defaults of %d\n", errorProbabilityDefault)
			error500ConfigValue = errorProbabilityDefault
			errorSinkHoleConfigValue = errorProbabilityDefault
			errorXMLResponseConfigValue = errorProbabilityDefault
		}
		if error500ConfigValue != error500Probabilty {
			log.Printf("%s has changed - setting to %d\n", Error500ResponseEnvName, error500ConfigValue)
			error500Probabilty = error500ConfigValue
		}
		if errorSinkHoleConfigValue != errorSinkHoleProbability {
			log.Printf("%s has changed - setting to %d\n", ErrorSinkHoleResponseEnvName, errorSinkHoleConfigValue)
			errorSinkHoleProbability = errorSinkHoleConfigValue
		}
		if errorXMLResponseConfigValue != errorXMLResponseProbability {
			log.Printf("%s has changed - setting to %d\n", ErrorXMLResponseEnvName, errorXMLResponseConfigValue)
			errorXMLResponseProbability = errorXMLResponseConfigValue
		}
	}
	setErrorLimit()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setErrorLimit()
		randVal := rand.Intn(100)
		if randVal < error500Probabilty {
			sendStatus500Response(w, r)
			return
		}
		if randVal < (error500Probabilty + errorSinkHoleProbability) {
			handleSinkholeRequest(handler, w, r)
			return
		}
		if randVal < (error500Probabilty + errorSinkHoleProbability + errorXMLResponseProbability) {
			handleXMLResponseRequest(handler, w, r)
			return
		}
		// over the error threshold value so pass to inner handler
		handler.ServeHTTP(w, r)
	})
}

func sendStatus500Response(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
}
func handleSinkholeRequest(handler http.Handler, w http.ResponseWriter, r *http.Request) {
	time.Sleep(errorSinkHoleDuration)
	handler.ServeHTTP(w, r)
}

func handleXMLResponseRequest(handler http.Handler, w http.ResponseWriter, r *http.Request) {

	testdata := model.TestXMLResponse{Name: "TestJob"}
	x, err := xml.MarshalIndent(testdata, "", " ")

	if err != nil {
		log.Printf("Error marshaling xml data")
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	_, errWrite := w.Write(x)

	if errWrite != nil {
		log.Printf("Error writing XML response")
		return
	}
	handler.ServeHTTP(w, r)
}
