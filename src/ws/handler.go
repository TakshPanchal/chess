package ws

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/takshpanchal/chess/src/game"
	"github.com/takshpanchal/chess/src/helpers"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// HandleWS handles the websocket connection and creates a new player
func HandleWS(w http.ResponseWriter, r *http.Request, gm *game.GameManager) {
	// Get query parameters for game joining
	gameID := r.URL.Query().Get("gameId")
	isSpectator := r.URL.Query().Get("spectator") == "true"

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		helpers.HandleError(err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	p := game.NewPlayer(conn, gm)

	// If gameId is provided, attempt to join the specific game
	if gameID != "" {
		// Send join game message through the player's message handling system
		joinMsg := game.Message[game.JoinGameData]{
			Type: game.JOIN_GAME,
			Data: game.JoinGameData{
				GameID:      gameID,
				IsSpectator: isSpectator,
			},
		}

		// Marshal and send through the player's connection
		if msgBytes, err := json.Marshal(joinMsg); err == nil {
			// Use a goroutine to avoid blocking
			go func() {
				p.HandleRequest(game.JOIN_GAME, msgBytes)
			}()
		}
	}

	// Start read and write pumps
	go p.ReadPump()
	go p.WritePump()
}
