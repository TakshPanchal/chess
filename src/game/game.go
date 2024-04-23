package game

import (
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Game struct {
	id          uint32
	BlackPlayer *Player
	WhitePlayer *Player
	Moves       []string
	broadcast   chan []byte
	// control   chan []byte
}

func NewGame() *Game {
	return &Game{
		id:          uuid.New().ID(),
		BlackPlayer: nil,
		WhitePlayer: nil,
		Moves:       make([]string, 0),
		// channels
		broadcast: make(chan []byte),
	}
}

func (g *Game) start() {
	g.WhitePlayer.send <- []byte("welcome, You're white")
	g.BlackPlayer.send <- []byte("welcome, You're black")
	log.Printf("%d Game is started.", g.id)

}

type Player struct {
	id   uint32
	Conn *websocket.Conn
	game *Game
	send chan []byte
}

func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		id:   uuid.New().ID(),
		Conn: conn,
		game: nil,
		send: make(chan []byte),
	}
}

// ReadPump reads from the connection and pump to the game
func (p *Player) ReadPump() {
	defer func() {
		// TODO: Close the game
		// p.game.
		p.Conn.Close()
	}()
	// p.Conn.SetReadDeadline()
	// p.Conn.SetPongHandler()
	for {
		_, reader, err := p.Conn.NextReader()
		if err != nil {
			log.Printf("Error: %+v", err)
			return
		}
		if p.game != nil {
			msg, err := io.ReadAll(reader)
			if err != nil {
				log.Printf("Error: %+v", err)
				return
			}
			p.game.broadcast <- msg
		}
	}
}

// WritePump writes to the connection and from the game
func (p *Player) WritePump() {
	defer func() {
		p.Conn.Close()
	}()
	for {
		msg, ok := <-p.send
		// if !ok {
		// 	// TODO: Do work on send channel close
		// }
		log.Println(ok)
		log.Println(string(msg))
		w, err := p.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Printf("Error: %+v", err)
			return
		}
		w.Write(msg)
		if err = w.Close(); err != nil {
			log.Printf("Error: %+v", err)
			return
		}
	}
}
