package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/microsoft/azure-databricks-operator/mockapi/router"
)

func main() {

	router := router.NewRouter()

	port := ":8085"
	fmt.Printf("API running on http://localhost%s\n", port)

	log.Fatal(http.ListenAndServe(port, router))
}
