package game

import (
	"math/rand"
	"testing"
)

func TestNewDeckHas52UniqueCards(t *testing.T) {
	d := NewDeck()

	if len(d.Cards) != 52 {
		t.Fatalf("deck has %d cards, want 52", len(d.Cards))
	}

	// A set of seen cards proves uniqueness: if any two were equal, the set
	// would have fewer than 52 entries. Structs are comparable in Go when all
	// their fields are, so a Card works directly as a map key.
	seen := make(map[Card]bool)
	for _, c := range d.Cards {
		if seen[c] {
			t.Errorf("duplicate card: %v", c)
		}
		seen[c] = true
		// also sanity-check the rank range while we're walking the deck
		if c.Rank < 2 || c.Rank > 14 {
			t.Errorf("card %v has out-of-range rank %d", c, c.Rank)
		}
	}
	if len(seen) != 52 {
		t.Errorf("found %d unique cards, want 52", len(seen))
	}
}

// TestSeededShuffleIsReproducible is the test that justifies the whole
// dependency-injection design from Phase 2: same seed in -> same order out.
// Everything downstream (deterministic game tests) relies on this property.
func TestSeededShuffleIsReproducible(t *testing.T) {
	a := NewDeck()
	a.Shuffle(rand.New(rand.NewSource(42)))

	b := NewDeck()
	b.Shuffle(rand.New(rand.NewSource(42)))

	for i := range a.Cards {
		if a.Cards[i] != b.Cards[i] {
			t.Fatalf("same seed produced different order at index %d: %v vs %v",
				i, a.Cards[i], b.Cards[i])
		}
	}
}

func TestDealSplitsEvenly(t *testing.T) {
	d := NewDeck()
	h1, h2 := d.Deal()
	if len(h1) != 26 || len(h2) != 26 {
		t.Errorf("deal gave %d + %d, want 26 + 26", len(h1), len(h2))
	}
	if len(h1)+len(h2) != len(d.Cards) {
		t.Errorf("halves (%d) don't sum to deck size (%d)", len(h1)+len(h2), len(d.Cards))
	}
}
