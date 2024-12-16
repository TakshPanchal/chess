package game

import (
	"log"
	"math/rand"
	"sync"
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
}

func NewGameManager() *GameManager {
	return &GameManager{
		Games:         make(map[string]*Game),
		WaitingPlayer: nil,
		AddPlayer:     make(chan *Player),
		RemovePlayer:  make(chan *Player),
		JoinGame:      make(chan *JoinGameRequest),
	}
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
	game := gm.GetGame(req.GameID)
	if game == nil {
		log.Printf("Failed to join game: Game not found with ID %s", req.GameID)
		req.Player.send <- []byte(`{"type":"ERROR","data":{"message":"Game not found"}}`)
		return
	}

	if req.IsSpectator {
		game.AddSpectator(req.Player)
		log.Printf("New spectator joined game %s", req.GameID)
		return
	}

	if game.WhitePlayer == nil || game.BlackPlayer == nil {
		gm.AddPlayer <- req.Player
		log.Printf("Player joined game %s as player", req.GameID)
	} else {
		game.AddSpectator(req.Player)
		log.Printf("Player joined game %s as spectator (game full)", req.GameID)
	}
}

func (gm *GameManager) Manage() {
	for {
		select {
		case p := <-gm.AddPlayer:
			{
				if gm.WaitingPlayer == nil {
					gm.WaitingPlayer = p
					log.Println("Player added to waiting queue")
				} else {
					g := NewGame()
					if randBool() {
						g.BlackPlayer = gm.WaitingPlayer
						g.WhitePlayer = p
					} else {
						g.BlackPlayer = p
						g.WhitePlayer = gm.WaitingPlayer
					}
					gm.WaitingPlayer = nil
					gm.addGameToMap(g)
					g.start()
				}
			}
		case req := <-gm.JoinGame:
			{
				gm.handleJoinGame(req)
			}
		case p := <-gm.RemovePlayer:
			{
				if p.game != nil {
					gameID := p.game.GetID()
					p.game.End <- true
					gm.removeGame(gameID)
				}
				if gm.WaitingPlayer == p {
					gm.WaitingPlayer = nil
					log.Println("Waiting player removed")
				}
			}
		}
	}
}

func randBool() bool {
	return rand.Intn(2) == 1
}
