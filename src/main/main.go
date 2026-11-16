package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/takshpanchal/chess/src/game"
	"github.com/takshpanchal/chess/src/ws"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func main() {

	// logger setup
	log.SetFlags(log.Llongfile | log.Ltime)

	gm := game.NewGameManager()
	go gm.Manage()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthCheck)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.HandleWS(w, r, gm)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Starting the server on port %s", port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panicf("Error: %v", err)
	}
}
