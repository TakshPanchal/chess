package game

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/takshpanchal/chess/src/helpers"
)

type PlayerType = string

const (
	BlackPlayer PlayerType = "black"
	WhitePlayer PlayerType = "white"
	cancelRequest PlayerType = "cancel"
)

type Player struct {
	Type        PlayerType
	id          uint32
	Conn        *websocket.Conn
	game        *Game
	gameManager *GameManager
	send        chan []byte
	done        chan struct{}
	flush       chan struct{}

	mu        sync.RWMutex
	closed    bool
	stopOnce  sync.Once
	flushOnce sync.Once
}

func NewPlayer(conn *websocket.Conn, gm *GameManager) *Player {
	return &Player{
		id:          uuid.New().ID(),
		Conn:        conn,
		game:        nil,
		gameManager: gm,
		send:        make(chan []byte, 16),
		done:        make(chan struct{}),
		flush:       make(chan struct{}),
		Type:        "",
	}
}

func (p *Player) isClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

func (p *Player) sendMessage(message []byte) bool {
	p.mu.RLock()
	closed := p.closed
	done := p.done
	send := p.send
	p.mu.RUnlock()
	if closed {
		return false
	}

	select {
	case send <- message:
		return true
	case <-done:
		return false
	}
}

func (p *Player) trySendMessage(message []byte) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return
	}
	select {
	case p.send <- message:
	default:
	}
}

func (p *Player) close() {
	p.stopOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		close(p.done)
		if p.Conn != nil {
			p.Conn.Close()
		}
	})
}

func (p *Player) closeAfterWrites() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()
	p.flushOnce.Do(func() {
		close(p.flush)
	})
}

// ReadPump reads from the WS connection and pump to the game
func (p *Player) ReadPump() {
	defer func() {
		p.close()
		if p.game != nil {
			p.game.end()
		} else if p.gameManager != nil {
			p.gameManager.RemovePlayer <- p
		}
	}()
	for {
		_, reader, err := p.Conn.NextReader()
		if err != nil {
			helpers.HandleError(err)
			return
		}
		msg, err := io.ReadAll(reader)
		if err != nil {
			helpers.HandleError(err)
			return
		}
		req := &Request{}
		if err := json.Unmarshal(msg, req); err != nil {
			helpers.HandleError(err)
			return
		}
		p.handleRequest(req.Type, msg)
	}
}

func (p *Player) handleRequest(reqType string, msg []byte) {
	switch reqType {
	case INIT:
		p.handleInit(msg)
	case MOVE:
		p.handleMove(msg)
	case GAME_OVER:
		p.handleGameOver(msg)
	case cancelRequest:
		p.handleCancel()
	default:
		p.sendMessage([]byte("Invalid Request"))
	}
}

func (p *Player) handleInit(_ []byte) {
	if p.game == nil && p.gameManager != nil && !p.isClosed() {
		p.gameManager.AddPlayer <- p
	}
}

func (p *Player) handleCancel() {
	if p.game == nil && p.gameManager != nil {
		p.gameManager.RemovePlayer <- p
	}
}

func (p *Player) handleMove(msg []byte) {
	resp, _ := json.Marshal(map[string]string{"type": "game_not_started", "data": "Game Not Started"})
	if p.game == nil {
		p.sendMessage(resp)
		return
	}
	moveMsg := &Message[MoveData]{}
	if err := json.Unmarshal(msg, &moveMsg); err != nil {
		p.sendMessage([]byte("Invalid Message Format"))
		return
	}

	select {
	case p.game.Moves <- Pair[*MoveData, *Player]{First: &moveMsg.Data, Second: p}:
	case <-p.game.End:
	}
}

func (p *Player) handleGameOver(_ []byte) {
	if p.game != nil {
		p.game.end()
	}
}

// WritePump write bytes to the connection and send channel
func (p *Player) WritePump() {
	defer p.close()
	for {
		select {
		case msg := <-p.send:
			if !p.writeMessage(msg) {
				return
			}
		case <-p.flush:
			for {
				select {
				case msg := <-p.send:
					if !p.writeMessage(msg) {
						return
					}
				default:
					return
				}
			}
		case <-p.done:
			return
		}
	}
}

func (p *Player) writeMessage(msg []byte) bool {
	if err := p.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		helpers.HandleError(err)
		return false
	}
	w, err := p.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		helpers.HandleError(err)
		return false
	}
	if _, err = w.Write(msg); err != nil {
		helpers.HandleError(err)
		return false
	}
	if err = w.Close(); err != nil {
		helpers.HandleError(err)
		return false
	}
	return true
}
