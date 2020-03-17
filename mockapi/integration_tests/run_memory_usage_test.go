package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/microsoft/azure-databricks-operator/mockapi/router"
	"github.com/stretchr/testify/assert"
)

const printMemEveryXItems = 50000

/*
*
* WARNING - THESE HAVE TO BE RUN INDEPENDENTLY SO MEMORY USAGE OF ONE BENCHMARK DOESN'T EFFECT THE OTHER. Use `make bench` to run correctly.
*
 */

func BenchmarkRunsMemoryUsage_Cleanup(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	memLimitMb := 2000

	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	index := 0
	for {
		index++

		submitAndCheckRun(server, b)

		// delete run - takes advantage of fact that we call synchronously and ids are incremented per call so match the `index` value
		response, err := server.Client().Post(
			server.URL+runAPI+"delete",
			"application/json",
			bytes.NewBufferString(fmt.Sprintf("{\"run_id\":%v}", index)))

		if err != nil {
			b.Error(err.Error())
			assert.Equal(b, nil, err)
			b.FailNow()
		}

		// devnull the response
		io.Copy(ioutil.Discard, response.Body) //nolint
		response.Body.Close()                  //nolint

		assert.Equal(b, 200, response.StatusCode)

		// Only check the memory every x items
		if shouldCheckMemory(index) {
			shouldExit := checkAndPrintMemoryUsage(index, b, memLimitMb)
			if shouldExit {
				break
			}
		}
	}

}

func BenchmarkRunsMemoryUsage_NeverDelete(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	memLimitMb := 2000

	// Arrange
	server := httptest.NewServer(router.NewRouter())
	defer server.Close()

	// Act
	index := 0
	for {
		index++

		submitAndCheckRun(server, b)

		// Only check the memory every x items
		if shouldCheckMemory(index) {
			shouldExit := checkAndPrintMemoryUsage(index, b, memLimitMb)
			if shouldExit {
				break
			}
		}
	}

}

func shouldCheckMemory(index int) bool {
	return float64(index)/printMemEveryXItems == math.Trunc(float64(index)/printMemEveryXItems)
}

func checkAndPrintMemoryUsage(index int, b *testing.B, memLimitMb int) bool {
	mbUsed := PrintMemUsage()
	fmt.Printf("-- InProgress - Used %vmb of memory for %v items \n\n", mbUsed, index)

	// If we've used more than 2GB stop the test
	if int(mbUsed) > memLimitMb {
		fmt.Printf("\n\n Test completed - Used >%vmb of memory. Used %vmb of memory for %vk items \n %vmb per item\n\n", memLimitMb, mbUsed, index/1000, float64(mbUsed)/float64(index))
		return true
	}
	return false
}

func submitAndCheckRun(server *httptest.Server, b *testing.B) {
	response, err := submitRun(server)
	if err != nil {
		b.Error(err.Error())
		assert.Equal(b, nil, err)
		b.FailNow()
	}

	// devnull the response
	io.Copy(ioutil.Discard, response.Body) //nolint
	response.Body.Close()                  //nolint

	assert.Equal(b, 200, response.StatusCode)
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garbage collection cycles completed.
func PrintMemUsage() (mbUsed uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)

	return bToMb(m.Sys)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
