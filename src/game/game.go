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
	End         chan bool
}

func NewGame() *Game {
	return &Game{
		id:          uuid.New().ID(),
		startTime:   time.Now(),
		BlackPlayer: nil,
		WhitePlayer: nil,
		ChessGame:   chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Moves:       make(chan Pair[*MoveData, *Player]),
		End:         make(chan bool),
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
	go g.gameLoop()
}

func (g *Game) gameLoop() {

	// start the game
	for {
		select {
		case p := <-g.Moves:
			g.makeMove(p)
		case <-g.End:
			{
				log.Println("Server disconnected")
				g.WhitePlayer.Conn.Close()
				g.BlackPlayer.Conn.Close()
			}
		}
	}
}

func (g *Game) makeMove(p Pair[*MoveData, *Player]) {
	// TODO: Validate the move

	move, player := p.First, p.Second
	log.Printf("Player move %v, %v", move.Player, player.Type)
	errResp, _ := json.Marshal(map[string]string{"type": "ERROR", "data": "invalid move"})
	if player.Type != move.Player {
		player.send <- errResp
		return
	}
	err := g.ChessGame.MoveStr(move.From + move.To)

	if err != nil {
		player.send <- errResp
		helpers.HandleError(err)
	}

	outcome := g.ChessGame.Outcome()
	switch outcome {
	case chess.NoOutcome:
		move.Outcome = "*"
	case chess.BlackWon:
		move.Outcome = "black"
	case chess.WhiteWon:
		move.Outcome = "white"
	default:
		move.Outcome = "draw"
	}
	msg, err := json.Marshal(NewResponse(MOVE, move))
	if err != nil {
		helpers.HandleError(err)
		return
	}
	if move.Outcome == "*" {
		if player.Type == BlackPlayer {
			g.WhitePlayer.send <- msg
		} else {
			g.BlackPlayer.send <- msg
		}
	} else {
		g.WhitePlayer.send <- msg
		g.BlackPlayer.send <- msg
		log.Printf("Game over. Outcome: %v", move.Outcome)
		g.End <- true
	}
	log.Printf("Outcome is %v", move)
	//TODO: check for game over
}
