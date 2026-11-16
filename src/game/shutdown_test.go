package game

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGameEndNotifiesPlayersOnce(t *testing.T) {
	g := NewGame()
	g.WhitePlayer = NewPlayer(nil, nil)
	g.BlackPlayer = NewPlayer(nil, nil)

	g.end()
	g.end()

	for _, player := range []*Player{g.WhitePlayer, g.BlackPlayer} {
		select {
		case msg := <-player.send:
			var response Request
			if err := json.Unmarshal(msg, &response); err != nil || response.Type != GAME_OVER {
				t.Fatalf("unexpected game-end message %q: %v", msg, err)
			}
		default:
			t.Fatal("player did not receive game-end notification")
		}
		select {
		case msg := <-player.send:
			t.Fatalf("duplicate game-end notification: %s", msg)
		default:
		}
	}
	select {
	case <-g.End:
	default:
		t.Fatal("game was not ended")
	}
}

func TestGameEndDoesNotWaitForFullPlayerQueue(t *testing.T) {
	g := NewGame()
	g.WhitePlayer = NewPlayer(nil, nil)
	g.BlackPlayer = NewPlayer(nil, nil)
	for range cap(g.WhitePlayer.send) {
		g.WhitePlayer.send <- []byte("queued")
	}

	finished := make(chan struct{})
	go func() {
		g.end()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		<-g.WhitePlayer.send
		t.Fatal("ending game blocked on full player queue")
	}
	select {
	case <-g.End:
	default:
		t.Fatal("game was not ended")
	}
	select {
	case msg := <-g.BlackPlayer.send:
		var response Request
		if err := json.Unmarshal(msg, &response); err != nil || response.Type != GAME_OVER {
			t.Fatalf("unexpected game-end message %q: %v", msg, err)
		}
	default:
		t.Fatal("other player did not receive game-end notification")
	}
}
