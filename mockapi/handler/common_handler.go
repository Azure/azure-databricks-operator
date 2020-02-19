package handler

import (
	"log"
	"net/http"
)

//Index returns index page
func Index(w http.ResponseWriter, r *http.Request) {
	requestID := getNewRequestID()
	log.Printf("RequestID:%6d: Index - starting\n", requestID)
	_, _ = w.Write([]byte("Databricks mock API is up"))
	log.Printf("RequestID:%6d: Index - completed\n", requestID)
}

//NotFoundPage returns a not found page
func NotFoundPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("*** Not Found: %s", r.URL)
	http.Error(w, "404 page not found", http.StatusNotFound)
}

//MethodNotAllowed returns a method not allowed page
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	log.Printf("*** Method Not Allowed: %s - %s", r.Method, r.URL)
	http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
}
