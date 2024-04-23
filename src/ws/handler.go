package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/takshpanchal/chess/src/game"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool { return true }}

func HandleWS(w http.ResponseWriter, r *http.Request, gm *game.GameManager) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error: %+v", err)
		return
	}

	p := game.NewPlayer(conn)
	gm.AddPlayer <- p
	go p.ReadPump()
	go p.WritePump()
}
