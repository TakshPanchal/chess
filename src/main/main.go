package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/takshpanchal/chess/src/game"
	"github.com/takshpanchal/chess/src/ws"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// logger setup
	log.SetFlags(log.Llongfile | log.Ltime)

	gm := game.NewGameManager()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthCheck)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.HandleWS(w, r, gm)
	})

	// Apply CORS middleware
	handler := corsMiddleware(mux)

	srv := &http.Server{
		Handler: handler,
		Addr:    ":8080",  // Use port 8080
	}

	log.Printf("Starting the server on port 8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panicf("Error: %v", err)
	}
}
