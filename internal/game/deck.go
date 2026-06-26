package game

import "math/rand"

// ─── Deck ──────────────────────────────────────────────────────────────────
//
// A Deck is a container with exactly three jobs and NO opinions about the game:
// build itself, shuffle itself, and deal itself in half. It does not know how
// to play War — it doesn't compare cards, track score, or know about turns.
// That "each piece knows only its own job" separation is the habit this whole
// project is teaching.
//
// The Deck holds a SLICE of Cards. A slice ([]Card) is Go's resizable list —
// the everyday collection you'll reach for 90% of the time. (Go also has fixed
// ARRAYS like [52]Card, but those have a baked-in length; slices grow and
// shrink, which is what we want.)
type Deck struct {
	Cards []Card
}

// NewDeck builds a full, ordered 52-card deck.
//
// This is a "constructor" — but remember, Go has no special constructor
// syntax. It's just a plain function named NewDeck() by convention that
// returns a ready-to-use value. The `New<Type>` naming is how Go programmers
// signal "this builds a Type for you."
//
// We return *Deck (a POINTER to a Deck) rather than a Deck value. Why a
// pointer here when Card used a value? Because callers will mutate the deck —
// Shuffle() rearranges it in place. Handing back a pointer means everyone
// shares and modifies the SAME deck, not private copies. (Same reasoning is
// coming for Player in Phase 3.)
func NewDeck() *Deck {
	d := &Deck{} // &Deck{} = "make a Deck and give me its address." Cards is
	//              nil right now, which is fine: appending to a nil slice works.

	// Two nested loops generate all 52 cards: 4 suits × 13 ranks = 52.
	// We never type out 52 lines — we describe the pattern and let the loops
	// produce it. Outer loop walks the suits, inner loop walks ranks 2..14.
	for s := Spades; s <= Diamonds; s++ {
		for r := 2; r <= 14; r++ {
			// append is THE slice operation to know. It returns a (possibly
			// new) slice with the element added on the end, and you must
			// assign the result back: `d.Cards = append(d.Cards, ...)`.
			// Forgetting to reassign is a classic Go beginner bug — the
			// append happens but you throw the result away.
			d.Cards = append(d.Cards, Card{Rank: r, Suit: s})
		}
	}
	return d
}

// Shuffle randomizes the order of the cards in place.
//
// Notice we take an *rand.Rand (a random-number generator) as a PARAMETER
// instead of using Go's global randomness. This is "dependency injection,"
// and it's a deliberate design choice for a learning/testing project:
//
//   - For real play, you pass a generator seeded from the clock → unpredictable.
//   - For tests, you pass a generator seeded with a FIXED number → the same
//     shuffle every single time, so a test can assert an exact outcome.
//
// Reproducible randomness sounds like a contradiction, but it's the thing that
// makes "deal this precise game and check the result" possible. We build it in
// from the start rather than bolting it on later.
//
// rng.Shuffle does the Fisher–Yates shuffle for us: it calls our little swap
// function to exchange pairs of elements until the order is randomized. We
// don't implement the algorithm — we just tell it HOW to swap two cards.
func (d *Deck) Shuffle(rng *rand.Rand) {
	rng.Shuffle(len(d.Cards), func(i, j int) {
		// The multiple-assignment swap: Go lets you exchange two values on one
		// line with no temp variable. Reads as "i becomes j, j becomes i."
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Deal splits the deck evenly into two halves: 26 cards each.
//
// Here's the most important slice concept in the whole project: SLICING.
// `d.Cards[:mid]` means "from the start up to (not including) mid" and
// `d.Cards[mid:]` means "from mid to the end." These are cheap — they don't
// copy the cards, they create two new slice "windows" that both look at the
// SAME underlying array of 52 cards.
//
// ⚠️ The sharp edge (worth understanding now, even though it doesn't bite us
// here): because both halves share one backing array, writing through one
// slice could in theory affect the other. We get away with it because each
// player immediately treats their half as their own queue and we never write
// back into these exact windows. But this shared-backing-array behavior is the
// #1 surprise slices spring on newcomers — file it away.
func (d *Deck) Deal() ([]Card, []Card) {
	mid := len(d.Cards) / 2
	return d.Cards[:mid], d.Cards[mid:]
}
