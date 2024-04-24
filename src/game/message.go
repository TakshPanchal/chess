package game

import "time"

// message types
const (
	INIT      = "init"
	MOVE      = "move"
	GAME_OVER = "over"
)

// TODO: Do typed stuff
// type MSG_INIT = INIT

// return types
const (
	OK = "ok"
)

// Request Structs
// Message struct
type Request struct {
	Type string `json:"type"`
}
type Message[T InitData | MoveData] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type InitData struct {
}

type MoveData struct {
	To   string `json:"to"`
	From string `json:"from"`
}

// Response Structs
type ResponseTypes interface {
	InitResponseData
}
type Response[T ResponseTypes] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type InitResponseData struct {
	Time  time.Time `json:"time"`
	Color string    `json:"color"`
}

func NewResponse[T ResponseTypes](t string, d T) *Response[T] {
	return &Response[T]{
		Type: t,
		Data: d,
	}
}
