package game

import (
	"encoding/json"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/takshpanchal/chess/src/helpers"
)

type PlayerType = string

const (
	BlackPlayer PlayerType = "black"
	WhitePlayer PlayerType = "white"
)

type Player struct {
	Type        PlayerType
	id          uint32
	Conn        *websocket.Conn
	game        *Game
	gameManager *GameManager
	send        chan []byte
}

func NewPlayer(conn *websocket.Conn, gm *GameManager) *Player {
	return &Player{
		id:          uuid.New().ID(),
		Conn:        conn,
		game:        nil,
		gameManager: gm,
		send:        make(chan []byte),
		Type:        "",
	}
}

// ReadPump reads from the WS connection and pump to the game
func (p *Player) ReadPump() {
	defer func() {
		if p.game != nil {
			// TODO: Close the game
		}
		p.Conn.Close()
	}()
	// p.Conn.SetReadDeadline()
	// p.Conn.SetPongHandler()
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
			p.handleRequest(req.Type, msg)
		}
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
	default:
		p.send <- []byte("Invalid Request")
	}
}

func (p *Player) handleInit(_ []byte) {
	// Execute the game init
	if p.game == nil {
		p.gameManager.AddPlayer <- p
		// TODO: send reply if neccessary to the client
	} else {
		// FIXME: What if game is already running
	}
	// p.send <- []byte("Init")
}

func (p *Player) handleMove(msg []byte) {
	if p.game == nil {
		p.send <- []byte("Game Not Started")
		return
	}
	moveMsg := &Message[MoveData]{}
	err := json.Unmarshal(msg, &moveMsg)
	if err != nil {
		//TODO: return structured error message
		p.send <- []byte("Invalid Message Format")
	}
	log.Println(moveMsg)

	p.game.Moves <- Pair[*MoveData, *Player]{First: &moveMsg.Data, Second: p}
}

func (p *Player) handleGameOver(_ []byte) {
	if p.game != nil {
		// TODO: Remove the player from the game
		// p.gameManager.RemovePlayer <- p
		p.game = nil
	} else {
		// TODO: What if game is not running
	}
}

// WritePump write bytes to the connection and send channel
func (p *Player) WritePump() {
	defer func() {
		p.Conn.Close()
	}()
	for {
		msg, ok := <-p.send
		if !ok {
			// TODO: Do work on send channel close
			log.Println("conn: closing | send channel closed")
			return
		}

		log.Printf("conn: working: %v", ok)
		log.Println(string(msg))
		w, err := p.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			helpers.HandleError(err)
			return
		}
		w.Write(msg)
		if err = w.Close(); err != nil {
			helpers.HandleError(err)
			return
		}
	}
}
