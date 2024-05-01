package game

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/notnil/chess"
	"github.com/takshpanchal/chess/src/helpers"
)

type Game struct {
	id          uint32
	startTime   time.Time
	BlackPlayer *Player
	WhitePlayer *Player
	ChessGame   *chess.Game
	Moves       chan Pair[*MoveData, *Player]
}

func NewGame() *Game {
	return &Game{
		id:          uuid.New().ID(),
		startTime:   time.Now(),
		BlackPlayer: nil,
		WhitePlayer: nil,
		ChessGame:   chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Moves:       make(chan Pair[*MoveData, *Player]),
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
	g.WhitePlayer.game = g
	g.BlackPlayer.game = g
	// send init message to both players
	g.WhitePlayer.Type = WhitePlayer
	g.BlackPlayer.Type = BlackPlayer
	sendInitResponse(g.WhitePlayer, NewResponse(INIT, InitResponseData{
		Time:  g.startTime,
		Color: g.WhitePlayer.Type,
	}))
	sendInitResponse(g.BlackPlayer, NewResponse(INIT, InitResponseData{
		Time:  g.startTime,
		Color: g.BlackPlayer.Type,
	}))

	log.Printf("%d Game is started.", g.id)

	// start the game
	for {
		select {
		case p := <-g.Moves:
			g.makeMove(p)
		}
	}
}

func (g *Game) makeMove(p Pair[*MoveData, *Player]) {
	// TODO: Validate the move

	move, player := p.First, p.Second
	err := g.ChessGame.MoveStr(move.From + move.To)

	if err != nil {
		player.send <- []byte("Invalid Move")
		helpers.HandleError(err)
	}

	msg, err := json.Marshal(NewResponse(MOVE, move))
	if err != nil {
		helpers.HandleError(err)
		return
	}

	if player.Type == BlackPlayer {
		g.WhitePlayer.send <- msg
	} else {
		g.BlackPlayer.send <- msg
	}

	//TODO: check for game over
}
