package game

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/notnil/chess"
	"github.com/takshpanchal/chess/src/helpers"
)

const (
	idLength = 6
	idChars  = "abcdefghijklmnopqrstuvwxyz0123456789"
)

type Game struct {
	id          string
	startTime   time.Time
	BlackPlayer *Player
	WhitePlayer *Player
	Spectators  []*Player
	ChessGame   *chess.Game
	Moves       chan Pair[*MoveData, *Player]
	End         chan bool
}

func generateShortID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = idChars[rand.Intn(len(idChars))]
	}
	return string(b)
}

func NewGame() *Game {
	return &Game{
		id:          generateShortID(),
		startTime:   time.Now(),
		BlackPlayer: nil,
		WhitePlayer: nil,
		Spectators:  make([]*Player, 0),
		ChessGame:   chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Moves:       make(chan Pair[*MoveData, *Player]),
		End:         make(chan bool),
	}
}

func (g *Game) AddSpectator(p *Player) {
	p.game = g
	p.Type = "spectator"
	g.Spectators = append(g.Spectators, p)
	
	sendInitResponse(p, NewResponse(INIT, InitResponseData{
		Time:      g.startTime,
		Color:     "spectator",
		GameID:    g.id,
		ViewerURL: "",
	}))
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
	g.WhitePlayer.Type = WhitePlayer
	g.BlackPlayer.Type = BlackPlayer
	
	sendInitResponse(g.WhitePlayer, NewResponse(INIT, InitResponseData{
		Time:      g.startTime,
		Color:     g.WhitePlayer.Type,
		GameID:    g.id,
		ViewerURL: "",
	}))
	sendInitResponse(g.BlackPlayer, NewResponse(INIT, InitResponseData{
		Time:      g.startTime,
		Color:     g.BlackPlayer.Type,
		GameID:    g.id,
		ViewerURL: "",
	}))

	log.Printf("Game %s is started.", g.id)
	go g.gameLoop()
}

func (g *Game) broadcastToSpectators(msg []byte) {
	for _, spectator := range g.Spectators {
		spectator.send <- msg
	}
}

func (g *Game) gameLoop() {
	for {
		select {
		case p := <-g.Moves:
			g.makeMove(p)
		case <-g.End:
			{
				log.Println("Server disconnected")
				g.WhitePlayer.Conn.Close()
				g.BlackPlayer.Conn.Close()
				for _, spectator := range g.Spectators {
					spectator.Conn.Close()
				}
			}
		}
	}
}

func (g *Game) isCorrectTurn(player *Player) bool {
	currentTurn := g.ChessGame.Position().Turn()
	return (currentTurn == chess.White && player.Type == WhitePlayer) ||
		(currentTurn == chess.Black && player.Type == BlackPlayer)
}

func (g *Game) makeMove(p Pair[*MoveData, *Player]) {
	move, player := p.First, p.Second
	log.Printf("Received move from %v player: %v to %v", player.Type, move.From, move.To)

	// Validate it's the correct player's turn
	if !g.isCorrectTurn(player) {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Not your turn"})
		respBytes, _ := json.Marshal(errResp)
		player.send <- respBytes
		log.Printf("Wrong turn for %v player", player.Type)
		return
	}

	// Validate and make the move
	err := g.ChessGame.MoveStr(move.From + move.To)
	if err != nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid move"})
		respBytes, _ := json.Marshal(errResp)
		player.send <- respBytes
		log.Printf("Invalid move from %v player: %v", player.Type, err)
		return
	}

	// Set move outcome and next turn
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

	// Set the next turn
	nextTurn := "white"
	if g.ChessGame.Position().Turn() == chess.Black {
		nextTurn = "black"
	}
	move.Turn = nextTurn

	// Create move response
	moveResp := NewResponse(MOVE, move)
	msg, err := json.Marshal(moveResp)
	if err != nil {
		helpers.HandleError(err)
		return
	}

	// Send move to both players and spectators
	g.WhitePlayer.send <- msg
	g.BlackPlayer.send <- msg
	g.broadcastToSpectators(msg)

	if move.Outcome != "*" {
		log.Printf("Game over. Outcome: %v", move.Outcome)
		g.End <- true
	}
	
	log.Printf("Move processed: from %s to %s, outcome: %s, next turn: %s", 
		move.From, move.To, move.Outcome, nextTurn)
}

func (g *Game) GetID() string {
	return g.id
}
