package game

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/takshpanchal/chess/src/helpers"
)

type Game struct {
	id          uint32
	startTime   time.Time
	BlackPlayer *Player
	WhitePlayer *Player
	Moves       []string
	moves   chan chess.Move
	// control   chan []byte
}

func NewGame() *Game {
	return &Game{
		id:          uuid.New().ID(),
		startTime:   time.Now(),
		BlackPlayer: nil,
		WhitePlayer: nil,
		Moves:       make([]string, 0),
		// channels
		broadcast: make(chan []byte),
	}
}

func sendInitResponse(p *Player, resp *Response[InitResponseData]) {
	respStruct, err := json.Marshal(resp)
	if err != nil {
		helpers.HandleError(err)
	}

	p.send <- respStruct
}

func (g *Game) start() {
	// send init message to both players
	sendInitResponse(g.WhitePlayer, NewResponse(INIT, InitResponseData{
		Time:  g.startTime,
		Color: "white",
	}))
	sendInitResponse(g.BlackPlayer, NewResponse(INIT, InitResponseData{
		Time:  g.startTime,
		Color: "black",
	}))

	log.Printf("%d Game is started.", g.id)

	// start the game

	for {
	}
}

func (g *Game) makeMove() {
	// validation
	// update the
}
