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
	g := &Game{
		id:          generateShortID(),
		startTime:   time.Now(),
		BlackPlayer: nil,
		WhitePlayer: nil,
		Spectators:  make([]*Player, 0),
		ChessGame:   chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Moves:       make(chan Pair[*MoveData, *Player]),
		End:         make(chan bool),
	}
	go g.gameLoop()
	return g
}

func (g *Game) AddSpectator(p *Player) {
	p.game = g
	p.Type = "spectator"
	p.IsSpectator = true
	g.Spectators = append(g.Spectators, p)

	// Send initial game state to spectator
	viewerURL := ""
	if g.id != "" {
		viewerURL = "/play/" + g.id + "?spectator=true"
	}

	// Determine current turn for spectator view
	var currentTurn string
	if g.ChessGame != nil {
		if g.ChessGame.Position().Turn() == chess.White {
			currentTurn = "white"
		} else {
			currentTurn = "black"
		}
	}

	sendInitResponse(p, NewResponse(INIT, InitResponseData{
		Time:       g.startTime,
		Color:      currentTurn, // Send current turn instead of "spectator"
		GameID:     g.id,
		ViewerURL:  viewerURL,
		IsSpectator: true, // Add explicit spectator flag
	}))

	log.Printf("Added spectator to game %s", g.id)
}

func sendInitResponse(p *Player, resp *Response[InitResponseData]) {
	respStruct, err := json.Marshal(resp)
	if err != nil {
		helpers.HandleError(err)
		return
	}
	p.Send <- respStruct
	log.Printf("Sent init response to player: %+v", resp.Data)
}

func (g *Game) start() {
	log.Printf("Starting game %s", g.id)

	if g.WhitePlayer == nil || g.BlackPlayer == nil {
		log.Printf("Cannot start game %s: missing players", g.id)
		return
	}

	g.WhitePlayer.game = g
	g.BlackPlayer.game = g
	g.WhitePlayer.Type = WhitePlayer
	g.BlackPlayer.Type = BlackPlayer

	// Send updated init responses to both players to confirm game start
	viewerURL := "/play/" + g.id + "?spectator=true"

	// White player init
	whiteInit := NewResponse(INIT, InitResponseData{
		Time:       g.startTime,
		Color:      WhitePlayer,
		GameID:     g.id,
		ViewerURL:  viewerURL,
		IsSpectator: false,
	})
	if respBytes, err := json.Marshal(whiteInit); err == nil {
		g.WhitePlayer.Send <- respBytes
	}

	// Black player init
	blackInit := NewResponse(INIT, InitResponseData{
		Time:       g.startTime,
		Color:      BlackPlayer,
		GameID:     g.id,
		ViewerURL:  viewerURL,
		IsSpectator: false,
	})
	if respBytes, err := json.Marshal(blackInit); err == nil {
		g.BlackPlayer.Send <- respBytes
	}

	log.Printf("Game %s started. White: %d, Black: %d", g.id, g.WhitePlayer.id, g.BlackPlayer.id)
}

func (g *Game) broadcastToSpectators(msg []byte) {
	for _, spectator := range g.Spectators {
		spectator.Send <- msg
	}
}

func (g *Game) gameLoop() {
	log.Printf("Game loop started for game %s", g.id)
	for {
		select {
		case p := <-g.Moves:
			g.makeMove(p)
		case <-g.End:
			{
				log.Printf("Game %s ended", g.id)
				if g.WhitePlayer != nil {
					g.WhitePlayer.Conn.Close()
				}
				if g.BlackPlayer != nil {
					g.BlackPlayer.Conn.Close()
				}
				for _, spectator := range g.Spectators {
					spectator.Conn.Close()
				}
				return
			}
		}
	}
}

func (g *Game) makeMove(p Pair[*MoveData, *Player]) {
	move, player := p.First, p.Second
	log.Printf("Game %s: Move attempt from %s to %s by %s", g.id, move.From, move.To, player.Type)

	// Validate player's turn
	currentTurn := g.ChessGame.Position().Turn()
	isWhiteTurn := currentTurn == chess.White
	if (isWhiteTurn && player.Type != WhitePlayer) || (!isWhiteTurn && player.Type != BlackPlayer) {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Not your turn"})
		respBytes, _ := json.Marshal(errResp)
		player.Send <- respBytes
		log.Printf("Game %s: Wrong turn for %s player", g.id, player.Type)
		return
	}

	// Validate and make the move
	err := g.ChessGame.MoveStr(move.From + move.To)
	if err != nil {
		errResp := NewResponse(ERROR, ErrorResponseData{Message: "Invalid move"})
		respBytes, _ := json.Marshal(errResp)
		player.Send <- respBytes
		log.Printf("Game %s: Invalid move from %s player: %v", g.id, player.Type, err)
		return
	}

	// Set move outcome
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

	// Create move response
	moveResp := NewResponse(MOVE, move)
	msg, err := json.Marshal(moveResp)
	if err != nil {
		helpers.HandleError(err)
		return
	}

	// Send move to both players and spectators
	g.WhitePlayer.Send <- msg
	g.BlackPlayer.Send <- msg
	g.broadcastToSpectators(msg)

	if move.Outcome != "*" {
		log.Printf("Game %s over. Outcome: %v", g.id, move.Outcome)
		g.End <- true
	}

	log.Printf("Game %s: Move processed from %s to %s, outcome: %s",
		g.id, move.From, move.To, move.Outcome)
}

func (g *Game) GetID() string {
	return g.id
}
