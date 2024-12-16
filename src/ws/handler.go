package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/takshpanchal/chess/src/game"
	"github.com/takshpanchal/chess/src/helpers"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// activeConnections tracks active connections to prevent duplicates
var (
	activeConnections = make(map[string]*game.Player)
	connectionsMutex  sync.RWMutex
)

// HandleWS handles the websocket connection and creates a new player
func HandleWS(w http.ResponseWriter, r *http.Request, gm *game.GameManager) {
	// Get query parameters for game joining
	gameID := r.URL.Query().Get("gameId")
	isSpectator := r.URL.Query().Get("spectator") == "true"
	isPlayer := r.URL.Query().Get("play") == "true"

	// If play=true is set, override spectator mode
	if isPlayer {
		isSpectator = false
	}

	// Generate a unique connection ID based on request parameters
	connectionID := r.RemoteAddr
	if gameID != "" {
		connectionID = gameID + "-" + r.RemoteAddr
	}

	// Check for existing connection
	connectionsMutex.RLock()
	existingPlayer, exists := activeConnections[connectionID]
	connectionsMutex.RUnlock()

	if exists && existingPlayer.IsActive() {
		log.Printf("Duplicate connection attempt from %s for game %s", r.RemoteAddr, gameID)
		http.Error(w, "Already connected", http.StatusConflict)
		return
	}

	log.Printf("New WebSocket connection request: gameId=%s, isSpectator=%v, isPlayer=%v", gameID, isSpectator, isPlayer)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		helpers.HandleError(err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	log.Printf("WebSocket connection established successfully")

	p := game.NewPlayer(conn, gm)

	// Store the new connection
	connectionsMutex.Lock()
	activeConnections[connectionID] = p
	connectionsMutex.Unlock()

	// Clean up connection when done
	defer func() {
		connectionsMutex.Lock()
		delete(activeConnections, connectionID)
		connectionsMutex.Unlock()
	}()

	// Start read and write pumps before sending any messages
	go p.ReadPump()
	go p.WritePump()

	// If gameId is provided, attempt to join the specific game
	if gameID != "" {
		log.Printf("Joining existing game: %s", gameID)
		// Send join game message through the player's message handling system
		joinMsg := game.Message[game.JoinGameData]{
			Type: game.JOIN_GAME,
			Data: game.JoinGameData{
				GameID:      gameID,
				IsSpectator: isSpectator,
				IsPlayer:    isPlayer,
			},
		}

		// Marshal and send through the player's connection
		if msgBytes, err := json.Marshal(joinMsg); err == nil {
			log.Printf("Sending join game message for gameId: %s with isPlayer=%v", gameID, isPlayer)
			p.HandleRequest(game.JOIN_GAME, msgBytes)
		} else {
			log.Printf("Error marshaling join game message: %v", err)
		}
	} else if isSpectator {
		log.Printf("Invalid connection request: spectator without gameId")
		errMsg := game.NewResponse(game.ERROR, game.ErrorResponseData{Message: "Spectator must provide a game ID"})
		if msgBytes, err := json.Marshal(errMsg); err == nil {
			p.Send <- msgBytes
		}
		return
	}

	log.Printf("WebSocket handler completed setup")
}
