package game

import "testing"

// These tests pin down the queue behavior — draw from the top, win to the
// bottom — and, crucially, that Draw actually MUTATES the hand (the Phase 3
// pointer-receiver lesson, now enforced by a test instead of a demo).

func TestDrawTakesFromTopAndShrinks(t *testing.T) {
	p := &Player{Name: "P", Hand: []Card{
		{Rank: 2, Suit: Spades},
		{Rank: 9, Suit: Hearts},
		{Rank: 14, Suit: Clubs},
	}}

	got := p.Draw()
	if got.Rank != 2 {
		t.Errorf("Draw returned rank %d, want 2 (the top card)", got.Rank)
	}
	if len(p.Hand) != 2 {
		// If Draw used a value receiver, the hand would still be length 3 here.
		// This is the pointer-vs-value bug, caught as a test failure.
		t.Errorf("hand length %d after Draw, want 2 (did Draw mutate the hand?)", len(p.Hand))
	}
	if p.Hand[0].Rank != 9 {
		t.Errorf("new top card rank %d, want 9", p.Hand[0].Rank)
	}
}

func TestCollectAddsToBottom(t *testing.T) {
	p := &Player{Name: "P", Hand: []Card{{Rank: 5, Suit: Spades}}}
	p.Collect(Card{Rank: 8, Suit: Hearts}, Card{Rank: 3, Suit: Clubs})

	if len(p.Hand) != 3 {
		t.Fatalf("hand length %d after Collect, want 3", len(p.Hand))
	}
	// Won cards go to the BOTTOM (end of the slice), in order.
	if p.Hand[1].Rank != 8 || p.Hand[2].Rank != 3 {
		t.Errorf("collected cards landed wrong: %v", p.Hand)
	}
}

func TestHasCards(t *testing.T) {
	empty := Player{Name: "E", Hand: nil}
	if empty.HasCards() {
		t.Error("empty hand reported HasCards() == true")
	}
	full := Player{Name: "F", Hand: []Card{{Rank: 2, Suit: Spades}}}
	if !full.HasCards() {
		t.Error("non-empty hand reported HasCards() == false")
	}
}
