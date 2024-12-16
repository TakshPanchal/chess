package game

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
	"github.com/takshpanchal/chess/src/helpers"
)

type PlayerType = string

const (
	BlackPlayer     PlayerType = "black"
	WhitePlayer     PlayerType = "white"
	SpectatorPlayer PlayerType = "spectator"

	// Constants for WebSocket timeouts
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Player struct {
	Type        PlayerType
	id          uint32
	Conn        *websocket.Conn
	game        *Game
	gameManager *GameManager
	Send        chan []byte
	IsSpectator bool
	mu          sync.RWMutex // Protects connection state
}

func NewPlayer(conn *websocket.Conn, gm *GameManager) *Player {
	return &Player{
		id:          uuid.New().ID(),
		Conn:        conn,
		game:        nil,
		gameManager: gm,
		Send:        make(chan []byte),
		Type:        "",
		IsSpectator: false,
	}
}

// IsActive checks if the player's connection is still active
func (p *Player) IsActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Conn != nil && p.Conn.WriteMessage(websocket.PingMessage, nil) == nil
}

func (p *Player) ReadPump() {
	defer func() {
		p.mu.Lock()
		if p.game != nil {
			p.gameManager.RemovePlayer <- p
			if !p.IsSpectator {
				p.game.End <- true
			}
		}
		if p.Conn != nil {
			p.Conn.Close()
			p.Conn = nil
		}
		p.mu.Unlock()
	}()

	p.Conn.SetReadDeadline(time.Now().Add(pongWait))
	p.Conn.SetPongHandler(func(string) error {
		p.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, reader, err := p.Conn.NextReader()
		if err != nil {
			helpers.HandleError(err)
			return
		}
		if msg, err := io.ReadAll(reader); err != nil {
			helpers.HandleError(err)
			return
		} else {
			req := &Request{}
			err = json.Unmarshal(msg, req)
			if err != nil {
				helpers.HandleError(err)
				return
			}
			log.Printf("Player %d received message type: %s", p.id, req.Type)
			p.HandleRequest(req.Type, msg)
		}
	}
}

func (p *Player) HandleRequest(reqType string, msg []byte) {
	log.Printf("Player %d handling request type: %s", p.id, reqType)
	switch reqType {
	case INIT:
		p.handleInit(msg)
	case MOVE:
		p.handleMove(msg)
	case GAME_OVER:
		p.handleGameOver(msg)
	case JOIN_GAME:
		p.handleJoinGame(msg)
	default:
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid request type"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.Send <- respBytes
		}
	}
}

func (p *Player) handleInit(_ []byte) {
	log.Printf("Player %d initializing game", p.id)
	if p.game == nil {
		// Send player to game manager to be matched
		log.Printf("Player %d sending to game manager", p.id)
		p.gameManager.AddPlayer <- p
	} else {
		log.Printf("Player %d already in game", p.id)
		// If already in a game, send current game state
		if p.game != nil {
			// Determine current turn for spectator view
			var currentTurn string
			if p.IsSpectator && p.game.ChessGame != nil {
				if p.game.ChessGame.Position().Turn() == chess.White {
					currentTurn = "white"
				} else {
					currentTurn = "black"
				}
			} else {
				currentTurn = p.Type
			}

			initResp := NewResponse(INIT, InitResponseData{
				Time:       p.game.startTime,
				Color:      currentTurn,
				GameID:     p.game.id,
				ViewerURL:  "/play/" + p.game.id + "?spectator=true",
				IsSpectator: p.IsSpectator,
			})
			if respBytes, err := json.Marshal(initResp); err == nil {
				p.Send <- respBytes
			}
		} else {
			errResp := NewResponse(ERROR, ErrorResponseData{Message: "Already in a game"})
			if respBytes, err := json.Marshal(errResp); err == nil {
				p.Send <- respBytes
			}
		}
	}
}

func (p *Player) handleJoinGame(msg []byte) {
	joinMsg := &Message[JoinGameData]{}
	err := json.Unmarshal(msg, joinMsg)
	if err != nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid join game message format"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.Send <- respBytes
		}
		return
	}

	log.Printf("Player %d joining game %s with isPlayer=%v", p.id, joinMsg.Data.GameID, joinMsg.Data.IsPlayer)
	p.gameManager.JoinGame <- &JoinGameRequest{
		GameID:      joinMsg.Data.GameID,
		Player:      p,
		IsSpectator: joinMsg.Data.IsSpectator,
		IsPlayer:    joinMsg.Data.IsPlayer,
	}
}

func (p *Player) handleMove(msg []byte) {
	if p.game == nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Game not started"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.Send <- respBytes
		}
		return
	}

	if p.IsSpectator {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Spectators cannot make moves"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.Send <- respBytes
		}
		return
	}

	moveMsg := &Message[MoveData]{}
	err := json.Unmarshal(msg, moveMsg)
	if err != nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid move format"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.Send <- respBytes
		}
		return
	}

	log.Printf("Player %d making move: %v", p.id, moveMsg.Data)
	p.game.Moves <- Pair[*MoveData, *Player]{First: &moveMsg.Data, Second: p}
}

func (p *Player) handleGameOver(_ []byte) {
	if p.game != nil {
		log.Printf("Game over for player %d", p.id)
		gameOver := NewResponse(GAME_OVER, ErrorResponseData{Message: "Game has ended"})
		gameOverBytes, _ := json.Marshal(gameOver)

		var wg sync.WaitGroup

		if p.game.WhitePlayer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.game.WhitePlayer.Send <- gameOverBytes
			}()
		}

		if p.game.BlackPlayer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.game.BlackPlayer.Send <- gameOverBytes
			}()
		}

		for _, spectator := range p.game.Spectators {
			wg.Add(1)
			go func(s *Player) {
				defer wg.Done()
				s.Send <- gameOverBytes
			}(spectator)
		}

		wg.Wait()

		p.gameManager.RemovePlayer <- p
		p.game.End <- true
	} else {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "No active game to end"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.Send <- respBytes
		}
	}
}

func (p *Player) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.mu.Lock()
		if p.Conn != nil {
			p.Conn.Close()
			p.Conn = nil
		}
		p.mu.Unlock()
	}()

	for {
		select {
		case msg, ok := <-p.Send:
			if !ok {
				p.mu.Lock()
				if p.Conn != nil {
					p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				p.mu.Unlock()
				return
			}

			log.Printf("Player %d sending message: %s", p.id, string(msg))
			p.mu.RLock()
			if p.Conn == nil {
				p.mu.RUnlock()
				return
			}
			w, err := p.Conn.NextWriter(websocket.TextMessage)
			p.mu.RUnlock()
			if err != nil {
				helpers.HandleError(err)
				return
			}
			w.Write(msg)

			if err := w.Close(); err != nil {
				helpers.HandleError(err)
				return
			}
		case <-ticker.C:
			p.mu.RLock()
			if p.Conn != nil && p.Conn.WriteMessage(websocket.PingMessage, nil) != nil {
				p.mu.RUnlock()
				return
			}
			p.mu.RUnlock()
		}
	}
}
