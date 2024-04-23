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

func main() {

	gm := game.NewGameManager()
	go gm.Manage()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthCheck)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.HandleWS(w, r, gm)
	})

	srv := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Println("Starting the server on port 8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Panicf("Error: %v", err)
	}
}
