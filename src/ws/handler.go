package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/takshpanchal/chess/src/game"
	"github.com/takshpanchal/chess/src/helpers"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool { return true }}

// HandleWS handles the websocket connection and creates a new player
func HandleWS(w http.ResponseWriter, r *http.Request, gm *game.GameManager) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		helpers.HandleError(err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	p := game.NewPlayer(conn, gm)
	// start read and write pumps
	go p.ReadPump()
	go p.WritePump()
}
