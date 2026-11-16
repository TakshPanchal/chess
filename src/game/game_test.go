package game

import (
	"testing"
	"time"

	"github.com/notnil/chess"
)

func TestCancelRemovesWaitingPlayerBeforeNextSearch(t *testing.T) {
	manager := NewGameManager()
	go manager.Manage()

	first := NewPlayer(nil, manager)
	first.handleInit(nil)
	waitFor(t, "first player to enter matchmaking", func() bool {
		return manager.WaitingPlayer == first
	})

	canceled := make(chan struct{})
	go func() {
		first.handleRequest("cancel", nil)
		close(canceled)
	}()
	waitFor(t, "cancellation request to complete", func() bool {
		select {
		case <-canceled:
			return true
		default:
			return false
		}
	})
	waitFor(t, "canceled player to leave matchmaking", func() bool {
		return manager.WaitingPlayer == nil
	})

	next := NewPlayer(nil, manager)
	next.handleInit(nil)
	waitFor(t, "new search to wait for an opponent", func() bool {
		return manager.WaitingPlayer == next
	})
	if next.game != nil {
		t.Fatal("new search was matched after the prior search was canceled")
	}
}

func TestMovePromotesPawnToQueenWhenPromotionIsOmitted(t *testing.T) {
	fen, err := chess.FEN("7k/P7/8/8/8/8/8/7K w - - 0 1")
	if err != nil {
		t.Fatalf("creating promotion position: %v", err)
	}

	game := NewGame()
	game.ChessGame = chess.NewGame(fen, chess.UseNotation(chess.UCINotation{}))
	white := NewPlayer(nil, nil)
	black := NewPlayer(nil, nil)
	white.Type = WhitePlayer
	black.Type = BlackPlayer
	white.send = make(chan []byte, 1)
	black.send = make(chan []byte, 1)
	game.WhitePlayer = white
	game.BlackPlayer = black

	game.makeMove(Pair[*MoveData, *Player]{
		First:  &MoveData{From: "a7", To: "a8", Player: WhitePlayer},
		Second: white,
	})

	if got := game.ChessGame.Position().Board().Piece(chess.A8); got != chess.WhiteQueen {
		t.Fatalf("promoted piece = %v, want white queen", got)
	}
}

func TestTerminalMoveDoesNotBlockGameLoop(t *testing.T) {
	game := NewGame()
	white := NewPlayer(nil, nil)
	black := NewPlayer(nil, nil)
	white.Type = WhitePlayer
	black.Type = BlackPlayer
	game.WhitePlayer = white
	game.BlackPlayer = black

	for _, move := range []string{"f2f3", "e7e5", "g2g4"} {
		if err := game.ChessGame.MoveStr(move); err != nil {
			t.Fatalf("setting up checkmate: %v", err)
		}
	}

	completed := make(chan struct{})
	go func() {
		game.makeMove(Pair[*MoveData, *Player]{
			First:  &MoveData{From: "d8", To: "h4", Player: BlackPlayer},
			Second: black,
		})
		close(completed)
	}()

	waitFor(t, "terminal move to finish", func() bool {
		select {
		case <-completed:
			return true
		default:
			return false
		}
	})

	select {
	case <-game.End:
	case <-time.After(time.Second):
		t.Fatal("terminal move did not end the game")
	}
}

func waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %s", description)
		case <-ticker.C:
		}
	}
}
