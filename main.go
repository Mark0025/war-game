// Command war plays the card game War on the command line.
//
// This is the program's entry point: execution begins at func main() in
// package main. It stays deliberately tiny — all the real logic lives in the
// internal/game package. main's only job is to wire up a game and press "go."
package main

import (
	"time"

	"war/internal/game"
)

func main() {
	// Seed from the clock for an unpredictable shuffle each run. (Pass a fixed
	// number instead to replay the exact same match every time — that's what
	// the tests in a later phase will do.)
	seed := time.Now().UnixNano()

	g := game.New("Player 1", "Player 2", seed)

	// The cap guarantees the program ends even if a game drags on. With
	// pot-shuffling enabled, real games finish well under this.
	g.Play(100000)
}
