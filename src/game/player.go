package game

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	send        chan []byte
	IsSpectator bool
}

func NewPlayer(conn *websocket.Conn, gm *GameManager) *Player {
	return &Player{
		id:          uuid.New().ID(),
		Conn:        conn,
		game:        nil,
		gameManager: gm,
		send:        make(chan []byte),
		Type:        "",
		IsSpectator: false,
	}
}

func (p *Player) ReadPump() {
	defer func() {
		if p.game != nil {
			p.gameManager.RemovePlayer <- p
			if !p.IsSpectator {
				p.game.End <- true
			}
		}
		p.Conn.Close()
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
			p.HandleRequest(req.Type, msg)
		}
	}
}

func (p *Player) HandleRequest(reqType string, msg []byte) {
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
			p.send <- respBytes
		}
	}
}

func (p *Player) handleJoinGame(msg []byte) {
	joinMsg := &Message[JoinGameData]{}
	err := json.Unmarshal(msg, joinMsg)
	if err != nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid join game message format"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.send <- respBytes
		}
		return
	}

	p.gameManager.JoinGame <- &JoinGameRequest{
		GameID:      joinMsg.Data.GameID,
		Player:      p,
		IsSpectator: joinMsg.Data.IsSpectator,
	}
}

func (p *Player) handleInit(_ []byte) {
	if p.game == nil {
		p.gameManager.AddPlayer <- p
		// Send waiting status to client
		initResp := NewResponse(INIT, InitResponseData{
			Time:  time.Now(),
			Color: "",
		})
		if respBytes, err := json.Marshal(initResp); err == nil {
			p.send <- respBytes
		}
	} else {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Already in a game"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.send <- respBytes
		}
	}
}

func (p *Player) handleMove(msg []byte) {
	if p.game == nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Game not started"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.send <- respBytes
		}
		return
	}

	if p.IsSpectator {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Spectators cannot make moves"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.send <- respBytes
		}
		return
	}

	moveMsg := &Message[MoveData]{}
	err := json.Unmarshal(msg, moveMsg)
	if err != nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid move format"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.send <- respBytes
		}
		return
	}

	p.game.Moves <- Pair[*MoveData, *Player]{First: &moveMsg.Data, Second: p}
}

func (p *Player) handleGameOver(_ []byte) {
	if p.game != nil {
		log.Println("Inside game over function")
		gameOver := NewResponse(GAME_OVER, ErrorResponseData{Message: "Game has ended"})
		gameOverBytes, _ := json.Marshal(gameOver)

		var wg sync.WaitGroup

		if p.game.WhitePlayer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.game.WhitePlayer.send <- gameOverBytes
			}()
		}

		if p.game.BlackPlayer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.game.BlackPlayer.send <- gameOverBytes
			}()
		}

		for _, spectator := range p.game.Spectators {
			wg.Add(1)
			go func(s *Player) {
				defer wg.Done()
				s.send <- gameOverBytes
			}(spectator)
		}

		wg.Wait()

		p.gameManager.RemovePlayer <- p
		p.game.End <- true
	} else {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "No active game to end"})
		if respBytes, err := json.Marshal(errResp); err == nil {
			p.send <- respBytes
		}
	}
}

func (p *Player) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-p.send:
			if !ok {
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.Conn.NextWriter(websocket.TextMessage)
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
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
