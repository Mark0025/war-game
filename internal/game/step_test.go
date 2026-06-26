package game

import "testing"

// These are the most important tests in the project. War's correctness lives
// almost entirely in the tie -> war -> resolution path, and especially in the
// edge case where a player can't fund a war. We construct EXACT hands (no
// randomness) so each scenario is forced and deterministic.

// newRiggedGame builds a WebGame, then replaces the dealt hands with the exact
// cards we want to test. Because the test is in `package game`, it can set the
// exported P1/P2 hands directly — a handy reason to keep tests in the same
// package as the code they exercise (called "white-box" testing).
func newRiggedGame(h1, h2 []Card) *WebGame {
	w := NewWebGame(0) // seed irrelevant; we overwrite the hands
	w.P1.Name = "P1"
	w.P2.Name = "P2"
	w.P1.Hand = h1
	w.P2.Hand = h2
	w.pot = nil
	w.warArmed = false
	return w
}

// Sanity: a plain higher-card win takes both cards and the loser shrinks.
func TestStepSimpleWin(t *testing.T) {
	w := newRiggedGame(
		[]Card{{Rank: 10, Suit: Spades}},
		[]Card{{Rank: 4, Suit: Hearts}},
	)
	r := w.Step()

	if r.OutcomeID != "RoundWon" {
		t.Fatalf("outcome = %q, want RoundWon", r.OutcomeID)
	}
	// P1 played the higher card, so P1 now holds BOTH cards (2), P2 holds 0.
	if r.Count1 != 2 || r.Count2 != 0 {
		t.Errorf("counts after win = %d/%d, want 2/0", r.Count1, r.Count2)
	}
}

// A tie reports WarStart and does NOT decide the round yet — the next Step
// funds the war.
func TestStepTieStartsWar(t *testing.T) {
	w := newRiggedGame(
		// both flip a 7 first -> tie -> war
		[]Card{{Rank: 7, Suit: Spades}, {Rank: 2, Suit: Spades}, {Rank: 3, Suit: Spades}, {Rank: 4, Suit: Spades}, {Rank: 13, Suit: Spades}},
		[]Card{{Rank: 7, Suit: Hearts}, {Rank: 2, Suit: Hearts}, {Rank: 3, Suit: Hearts}, {Rank: 4, Suit: Hearts}, {Rank: 5, Suit: Hearts}},
	)
	r := w.Step()
	if r.OutcomeID != "WarStart" {
		t.Fatalf("outcome = %q, want WarStart", r.OutcomeID)
	}
	if !w.warArmed {
		t.Error("warArmed should be true after a tie")
	}
	if r.PotSize != 2 {
		t.Errorf("pot size after tie = %d, want 2", r.PotSize)
	}
}

// Full war that RESOLVES: tie on 7, each stakes 3 face-down, then the deciding
// face-up card decides. P1's deciding card (King) beats P2's (5), so P1 sweeps
// the whole 10-card pot.
func TestStepWarResolves(t *testing.T) {
	w := newRiggedGame(
		// 7, then 3 face-down (2,3,4), then deciding King
		[]Card{{Rank: 7, Suit: Spades}, {Rank: 2, Suit: Spades}, {Rank: 3, Suit: Spades}, {Rank: 4, Suit: Spades}, {Rank: 13, Suit: Spades}},
		// 7, then 3 face-down (2,3,4), then deciding 5
		[]Card{{Rank: 7, Suit: Hearts}, {Rank: 2, Suit: Hearts}, {Rank: 3, Suit: Hearts}, {Rank: 4, Suit: Hearts}, {Rank: 5, Suit: Hearts}},
	)

	r1 := w.Step() // tie -> WarStart
	if r1.OutcomeID != "WarStart" {
		t.Fatalf("step1 outcome = %q, want WarStart", r1.OutcomeID)
	}
	r2 := w.Step() // funds 3 face-down each, draws deciding pair, resolves
	if r2.OutcomeID != "RoundWon" {
		t.Fatalf("step2 outcome = %q, want RoundWon", r2.OutcomeID)
	}
	// 10 cards were on the table (2 + 6 stakes + 2 deciding). P1 wins them all
	// and started with 5, P2 with 5; after the sweep P1 has all 10, P2 has 0.
	if r2.Count1 != 10 || r2.Count2 != 0 {
		t.Errorf("counts after war = %d/%d, want 10/0", r2.Count1, r2.Count2)
	}
}

// THE edge case the spec singled out: a player ties into a war but cannot
// field enough cards to fund it. They must LOSE THE GAME — not crash, not
// silently continue. P2 ties on 7 but then has nothing left to stake.
func TestStepWarWithInsufficientCardsLosesGame(t *testing.T) {
	w := newRiggedGame(
		// P1: the tie card + plenty to fund the war
		[]Card{{Rank: 7, Suit: Spades}, {Rank: 2, Suit: Spades}, {Rank: 3, Suit: Spades}, {Rank: 4, Suit: Spades}, {Rank: 13, Suit: Spades}},
		// P2: ONLY the tie card — cannot fund the war that follows
		[]Card{{Rank: 7, Suit: Hearts}},
	)

	r1 := w.Step() // tie -> WarStart
	if r1.OutcomeID != "WarStart" {
		t.Fatalf("step1 outcome = %q, want WarStart", r1.OutcomeID)
	}
	r2 := w.Step() // P2 can't fund -> game over, P1 wins
	if r2.OutcomeID != "GameOver" {
		t.Fatalf("step2 outcome = %q, want GameOver", r2.OutcomeID)
	}
	if r2.Winner != "P1" {
		t.Errorf("winner = %q, want P1 (P2 ran out mid-war)", r2.Winner)
	}
}

// An empty player at the very start of a face-off ends the game immediately.
func TestStepEmptyPlayerEndsGame(t *testing.T) {
	w := newRiggedGame(
		[]Card{{Rank: 5, Suit: Spades}},
		nil, // P2 starts with no cards
	)
	r := w.Step()
	if r.OutcomeID != "GameOver" {
		t.Fatalf("outcome = %q, want GameOver", r.OutcomeID)
	}
	if r.Winner != "P1" {
		t.Errorf("winner = %q, want P1", r.Winner)
	}
}
