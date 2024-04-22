package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthCheck)

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panicf("Error: %v", err)
	}
}
