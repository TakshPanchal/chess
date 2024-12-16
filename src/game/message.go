package game

import (
	"time"
)

// message types
const (
	INIT      = "init"
	MOVE      = "move"
	GAME_OVER = "over"
	JOIN_GAME = "join_game"
	ERROR     = "error"
)

// return types
const (
	OK = "ok"
)

// Request Structs
type Request struct {
	Type string          `json:"type"`
	Data map[string]any `json:"data"`
}

type Message[T InitData | MoveData | JoinGameData] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type InitData struct {
	// Empty struct as we don't need any data for init request
}

type MoveData struct {
	To      string     `json:"to"`
	From    string     `json:"from"`
	Player  PlayerType `json:"color"`     // The color of the player making the move
	Outcome string     `json:"outcome"`   // Game outcome after this move
	Turn    string     `json:"turn"`      // Next player's turn after this move (white/black)
}

type JoinGameData struct {
	GameID      string `json:"gameId"`
	IsSpectator bool   `json:"isSpectator"`
	IsPlayer    bool   `json:"isPlayer"`
}

// Response Structs
type ResponseTypes interface {
	InitResponseData | *MoveData | JoinGameResponseData | ErrorResponseData
}

type Response[T ResponseTypes] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type InitResponseData struct {
	Time       time.Time `json:"time"`
	Color      string    `json:"color"`      // Player's assigned color (white/black) or current turn for spectators
	GameID     string    `json:"gameId"`     // For sharing
	ViewerURL  string    `json:"viewerUrl"`  // URL for spectators
	IsSpectator bool     `json:"isSpectator"` // Whether the client is a spectator
}

type JoinGameResponseData struct {
	Success     bool      `json:"success"`
	GameID      string    `json:"gameId"`
	PlayerType  string    `json:"playerType"` // white/black/spectator
	GameState   string    `json:"gameState"`  // FEN notation of current board
	StartTime   time.Time `json:"startTime"`
	IsSpectator bool      `json:"isSpectator"`
}

type ErrorResponseData struct {
	Message string `json:"message"`
}

func NewResponse[T ResponseTypes](t string, d T) *Response[T] {
	return &Response[T]{
		Type: t,
		Data: d,
	}
}

type Pair[T, U any] struct {
	First  T
	Second U
}
