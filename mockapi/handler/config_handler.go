package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/microsoft/azure-databricks-operator/mockapi/middleware"
)

// Config is used to represent the mock api config
type Config struct {
	LatencySlowRequestMax       *int `json:"DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MAX"`
	LatencySlowRequestMin       *int `json:"DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_SLOW_REQUEST_MIN"`
	LatencyFastRequestMax       *int `json:"DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MAX"`
	LatencyFastRequestMin       *int `json:"DATABRICKS_MOCK_API_LATENCY_MILLISECONDS_FAST_REQUEST_MIN"`
	RateLimit                   *int `json:"DATABRICKS_MOCK_API_RATE_LIMIT"`
	Error500Probability         *int `json:"DATABRICKS_MOCK_API_ERROR_500_PROBABILITY"`
	ErrorSinkHoleProbability    *int `json:"DATABRICKS_MOCK_API_ERROR_SINKHOLE_PROBABILITY"`
	ErrorXMLResponseProbability *int `json:"DATABRICKS_MOCK_API_ERROR_XML_RESPONSE_PROBABILITY"`
}

// GetConfig gets the current config
func GetConfig() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		config := Config{
			LatencySlowRequestMax:       getIntSetting(middleware.LatencySlowRequestMaxEnvName),
			LatencySlowRequestMin:       getIntSetting(middleware.LatencySlowRequestMinEnvName),
			LatencyFastRequestMax:       getIntSetting(middleware.LatencyFastRequestMaxEnvName),
			LatencyFastRequestMin:       getIntSetting(middleware.LatencyFastRequestMinEnvName),
			RateLimit:                   getIntSetting(middleware.RateLimitEnvName),
			Error500Probability:         getIntSetting(middleware.Error500ResponseEnvName),
			ErrorSinkHoleProbability:    getIntSetting(middleware.ErrorSinkHoleResponseEnvName),
			ErrorXMLResponseProbability: getIntSetting(middleware.ErrorXMLResponseEnvName),
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

// SetConfig updates the config
func SetConfig() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var config Config
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("Config: Error reading the body: %v", err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &config); err != nil {
			log.Printf("Config: Error parsing the body: %v", err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		// TODO - would be nice to validate the config values here and return an error response
		// but that logic is currently buried in the middleware itself
		if err := updateIntSetting(w, middleware.LatencySlowRequestMaxEnvName, config.LatencySlowRequestMax); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.LatencySlowRequestMinEnvName, config.LatencySlowRequestMin); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.LatencyFastRequestMaxEnvName, config.LatencyFastRequestMax); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.LatencyFastRequestMinEnvName, config.LatencyFastRequestMin); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.RateLimitEnvName, config.RateLimit); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.Error500ResponseEnvName, config.Error500Probability); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.ErrorSinkHoleResponseEnvName, config.ErrorSinkHoleProbability); err != nil {
			return
		}
		if err := updateIntSetting(w, middleware.ErrorXMLResponseEnvName, config.ErrorXMLResponseProbability); err != nil {
			return
		}
	}
}

// PatchConfig allows updating a subset of the config
func PatchConfig() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		defer r.Body.Close() // nolint: errcheck
		if err != nil {
			log.Printf("Config: Error reading the body: %v", err)
			http.Error(w, "Error reading the body", http.StatusBadRequest)
			return
		}

		var configPatch map[string]*int
		if err := json.Unmarshal(body, &configPatch); err != nil {
			log.Printf("Config: Error parsing the body: %v", err)
			http.Error(w, "Error parsing body", http.StatusBadRequest)
			return
		}

		for name, value := range configPatch {
			switch name {
			case middleware.LatencyFastRequestMaxEnvName,
				middleware.LatencyFastRequestMinEnvName,
				middleware.LatencySlowRequestMaxEnvName,
				middleware.LatencySlowRequestMinEnvName,
				middleware.Error500ResponseEnvName,
				middleware.ErrorSinkHoleResponseEnvName,
				middleware.ErrorXMLResponseEnvName,
				middleware.RateLimitEnvName:
				if err := setIntSetting(name, value); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			default:
				log.Printf("Error: unexepected value '%s'\n", name)
				http.Error(w, fmt.Sprintf("Error: unexepected value '%s'\n", name), http.StatusBadRequest)
				return
			}
		}
	}
}

func getIntSetting(name string) *int {
	value := os.Getenv(name)
	if value == "" {
		return nil
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &intValue
}
func setIntSetting(name string, value *int) error {
	if value == nil {
		err := os.Unsetenv(name)
		if err != nil {
			return err
		}
		return nil
	}
	err := os.Setenv(name, fmt.Sprintf("%d", *value))
	if err != nil {
		return err
	}
	return nil
}
func updateIntSetting(w http.ResponseWriter, name string, value *int) error {
	if err := setIntSetting(name, value); err != nil {
		log.Printf("Config: Error updating %s: %v", name, err)
		http.Error(w, fmt.Sprintf("Error updating %s", name), http.StatusBadRequest)
		return err
	}
	return nil
}
