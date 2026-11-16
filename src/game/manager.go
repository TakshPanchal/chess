package game

import (
	"math/rand"
)

type GameManager struct {
	Games         []*Game
	WaitingPlayer *Player

	// channels
	AddPlayer    chan *Player
	RemovePlayer chan *Player
}

func NewGameManager() *GameManager {
	return &GameManager{
		Games:         make([]*Game, 0),
		WaitingPlayer: nil,
		AddPlayer:     make(chan *Player),
		RemovePlayer:  make(chan *Player),
	}
}

func (gm *GameManager) Manage() {
	for {
		select {
		case p := <-gm.AddPlayer:
			if p == nil || p.isClosed() || p == gm.WaitingPlayer {
				continue
			}

			if gm.WaitingPlayer == nil || gm.WaitingPlayer.isClosed() {
				gm.WaitingPlayer = p
				continue
			}

			g := NewGame()
			if randBool() {
				g.BlackPlayer = gm.WaitingPlayer
				g.WhitePlayer = p
			} else {
				g.BlackPlayer = p
				g.WhitePlayer = gm.WaitingPlayer
			}
			gm.WaitingPlayer = nil
			g.start()
		case p := <-gm.RemovePlayer:
			if gm.WaitingPlayer == p {
				gm.WaitingPlayer = nil
			}
		}
	}
}

func randBool() bool {
	return rand.Intn(2) == 1
}
