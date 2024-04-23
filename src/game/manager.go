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

// func (gm *GameManager) AddPlayer(p *Player) error {}
func (gm *GameManager) Manage() {

	for {

		select {
		case p := <-gm.AddPlayer:
			{
				if gm.WaitingPlayer == nil {
					gm.WaitingPlayer = p
				} else {
					g := NewGame()
					if randBool() {
						g.BlackPlayer = gm.WaitingPlayer
						g.WhitePlayer = p
					} else {
						g.BlackPlayer = p
						g.WhitePlayer = gm.WaitingPlayer
					}
					gm.WaitingPlayer = nil
					go g.start()
				}

			}
			// case p := <-gm.RemovePlayer:
		}

	}
}

func randBool() bool {
	return rand.Intn(2) == 1
}
