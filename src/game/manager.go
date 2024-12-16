package game

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"
)

type GameManager struct {
	Games         map[string]*Game
	WaitingPlayer *Player
	mu            sync.RWMutex

	// channels
	AddPlayer    chan *Player
	RemovePlayer chan *Player
	JoinGame     chan *JoinGameRequest
}

type JoinGameRequest struct {
	GameID      string
	Player      *Player
	IsSpectator bool
	IsPlayer    bool
}

func NewGameManager() *GameManager {
	rand.Seed(time.Now().UnixNano())
	gm := &GameManager{
		Games:         make(map[string]*Game),
		WaitingPlayer: nil,
		AddPlayer:     make(chan *Player),
		RemovePlayer:  make(chan *Player),
		JoinGame:      make(chan *JoinGameRequest),
	}
	go gm.Manage()
	return gm
}

func (gm *GameManager) GetGame(id string) *Game {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.Games[id]
}

func (gm *GameManager) addGameToMap(g *Game) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.Games[g.GetID()] = g
	log.Printf("New game created with ID: %s", g.GetID())
}

func (gm *GameManager) removeGame(id string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.Games, id)
	log.Printf("Game removed with ID: %s", id)
}

func (gm *GameManager) handleJoinGame(req *JoinGameRequest) {
	log.Printf("Handling join game request: gameId=%s, isSpectator=%v, isPlayer=%v", req.GameID, req.IsSpectator, req.IsPlayer)

	game := gm.GetGame(req.GameID)
	if game == nil {
		log.Printf("Failed to join game: Game not found with ID %s", req.GameID)
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Game not found"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			req.Player.Send <- respBytes
		}
		return
	}

	// If explicitly marked as spectator and not a player, add as spectator
	if req.IsSpectator && !req.IsPlayer {
		game.AddSpectator(req.Player)
		log.Printf("New spectator joined game %s", req.GameID)
		return
	}

	// Check if game is full and handle based on player/spectator status
	if game.WhitePlayer != nil && game.BlackPlayer != nil {
		if req.IsPlayer {
			// If trying to join as player but game is full, send error
			log.Printf("Player attempted to join full game %s", req.GameID)
			errResp := NewResponse(ERROR, ErrorResponseData{Message: "Game is full"})
			if respBytes, err := json.Marshal(errResp); err == nil {
				req.Player.Send <- respBytes
			}
		} else {
			// If not explicitly requesting to play, join as spectator
			log.Printf("Game %s is full, joining as spectator", req.GameID)
			game.AddSpectator(req.Player)
		}
		return
	}

	// Join as player in the first available slot
	var assignedColor string
	if game.WhitePlayer == nil {
		game.WhitePlayer = req.Player
		assignedColor = WhitePlayer
		req.Player.Type = WhitePlayer
		log.Printf("Player joined game %s as White", req.GameID)
	} else if game.BlackPlayer == nil {
		game.BlackPlayer = req.Player
		assignedColor = BlackPlayer
		req.Player.Type = BlackPlayer
		log.Printf("Player joined game %s as Black", req.GameID)
	}
	req.Player.game = game

	// Send init response with assigned color
	viewerURL := "/play/" + game.id + "?spectator=true"
	initResp := NewResponse(INIT, InitResponseData{
		Time:       game.startTime,
		Color:      assignedColor,
		GameID:     game.id,
		ViewerURL:  viewerURL,
		IsSpectator: false,
	})
	if respBytes, err := json.Marshal(initResp); err == nil {
		req.Player.Send <- respBytes
	}

	// If both players are now present, start the game
	if game.WhitePlayer != nil && game.BlackPlayer != nil {
		log.Printf("Both players present, starting game %s", req.GameID)
		game.start()
	} else {
		log.Printf("Waiting for opponent in game %s", req.GameID)
	}
}

func (gm *GameManager) handleNewPlayer(p *Player) {
	log.Printf("Handling new player")
	if gm.WaitingPlayer == nil {
		// Create a new game for this player
		g := NewGame()
		g.WhitePlayer = p
		p.Type = WhitePlayer
		p.game = g
		gm.WaitingPlayer = p
		gm.addGameToMap(g)

		// Send initial response to waiting player
		viewerURL := "/play/" + g.id + "?spectator=true"
		initResp := NewResponse(INIT, InitResponseData{
			Time:       g.startTime,
			Color:      WhitePlayer,
			GameID:     g.id,
			ViewerURL:  viewerURL,
			IsSpectator: false,
		})
		if respBytes, err := json.Marshal(initResp); err == nil {
			p.Send <- respBytes
		}
		log.Printf("Player added to waiting queue and assigned game %s", g.id)
	} else {
		log.Printf("Found waiting player, completing game")
		g := gm.WaitingPlayer.game
		g.BlackPlayer = p
		p.Type = BlackPlayer
		p.game = g

		// Send initial response to black player
		viewerURL := "/play/" + g.id + "?spectator=true"
		initResp := NewResponse(INIT, InitResponseData{
			Time:       g.startTime,
			Color:      BlackPlayer,
			GameID:     g.id,
			ViewerURL:  viewerURL,
			IsSpectator: false,
		})
		if respBytes, err := json.Marshal(initResp); err == nil {
			p.Send <- respBytes
		}

		gm.WaitingPlayer = nil
		g.start()
		log.Printf("Game %s started with both players", g.id)
	}
}

func (gm *GameManager) Manage() {
	log.Println("Game manager started")
	for {
		select {
		case p := <-gm.AddPlayer:
			gm.handleNewPlayer(p)
		case req := <-gm.JoinGame:
			gm.handleJoinGame(req)
		case p := <-gm.RemovePlayer:
			if p.game != nil {
				gameID := p.game.GetID()
				log.Printf("Removing player from game %s", gameID)
				p.game.End <- true
				gm.removeGame(gameID)
			}
			if gm.WaitingPlayer == p {
				gm.WaitingPlayer = nil
				log.Printf("Waiting player removed")
			}
		}
	}
}

func randBool() bool {
	return rand.Intn(2) == 1
}
