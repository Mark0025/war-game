// Package game holds everything about playing War: the cards, the deck,
// the players, and the round loop. Keeping it in its own package (under
// internal/, so nothing outside this module can import it) lets main.go stay
// tiny — it just wires a game up and presses "go".
package game

import (
	"encoding/json"
	"fmt"
)

// ─── Suit ────────────────────────────────────────────────────────────────
//
// A Suit is one of the four card suits. Go has no "enum" keyword, so the
// idiom is: define a named integer type, then list its allowed values as
// constants using `iota`.
//
// `type Suit int` means "a Suit is an int, but a DISTINCT kind of int." The
// compiler now stops you from accidentally mixing a Suit with a plain number
// or a Rank — that type-safety is the whole reason we don't just use raw ints.
type Suit int

// iota is Go's auto-incrementing counter inside a const block. It starts at 0
// on the first line and adds 1 each line. So:
//
//	Spades   = 0
//	Clubs    = 1
//	Hearts   = 2
//	Diamonds = 3
//
// We never care about the numbers themselves — only that the four suits are
// distinct, ordered values. (In TypeScript you'd reach for a union type or an
// enum; this is Go's equivalent, and it's just integers under the hood.)
const (
	Spades Suit = iota
	Clubs
	Hearts
	Diamonds
)

// String makes Suit satisfy the fmt.Stringer interface.
//
// An interface in Go is a contract: "anything with a String() string method
// counts as a Stringer." We never *declare* "Suit implements Stringer" — Go
// figures it out automatically because the method signature matches. This is
// called structural (or "duck") typing: if it has the method, it qualifies.
//
// The payoff: fmt.Println(Hearts) prints "Hearts" instead of "2", because
// fmt looks for a String() method and uses it when it finds one.
//
// The [...]string{...} is an array literal indexed by the suit's integer
// value: Spades(0) -> "Spades", Clubs(1) -> "Clubs", and so on. We attach
// this to a VALUE receiver (s Suit, not *Suit) because a Suit is tiny and we
// never modify it — copying it is free and safe.
func (s Suit) String() string {
	return [...]string{"Spades", "Clubs", "Hearts", "Diamonds"}[s]
}

// ─── Card ────────────────────────────────────────────────────────────────
//
// A Card is the smallest, dumbest thing in the game: it knows two facts about
// itself and does nothing. This is a struct — Go's "bundle of named fields."
// There is NO class here, no constructor ceremony, no inheritance. Just data.
//
// Rank is a plain int (2..14) on purpose: comparison becomes trivial — a
// bigger number beats a smaller one, full stop. Jack/Queen/King/Ace are just
// 11/12/13/14 wearing fancy names, and we only translate to those names at
// the moment we print. Storing the comparable thing and prettifying late
// deletes a whole category of comparison bugs.
//
// Critical to War: Suit NEVER decides a round. Two equal ranks tie regardless
// of suit (that's what triggers a "war"). Suit exists only so "Ace of Spades"
// feels like a real card when printed.
type Card struct {
	Rank int  // 2..14  (11=Jack, 12=Queen, 13=King, 14=Ace)
	Suit Suit // cosmetic only — never compared
}

// String makes Card a Stringer too, so fmt prints "Ace of Spades" rather than
// the raw struct. Like Suit, it's a value receiver: a Card is small and
// immutable — once it's the Ace of Spades, it stays that forever — so we copy
// it freely instead of pointing at it. (Remember this contrast: in player.go
// the Player will need a POINTER receiver because its hand changes. Card
// never changes, so value is correct here.)
func (c Card) String() string {
	// Only the face cards get special names; 2..10 just print their number.
	// A small map is the clearest way to express "these four are special."
	names := map[int]string{11: "Jack", 12: "Queen", 13: "King", 14: "Ace"}
	if name, ok := names[c.Rank]; ok {
		// The comma-ok idiom: looking up a missing map key returns the
		// zero value AND ok=false. So `name, ok := names[c.Rank]` lets us
		// ask "was this rank in the map?" without a separate check.
		return fmt.Sprintf("%s of %s", name, c.Suit)
	}
	return fmt.Sprintf("%d of %s", c.Rank, c.Suit)
}

// MarshalJSON controls how a Card turns into JSON for the web API.
//
// This is json.Marshaler — the sibling of Stringer. Just as Stringer lets a
// type say "here's how I print myself," Marshaler lets it say "here's how I
// turn into JSON." Without it, the default encoding leaks the Go field names
// and a raw integer suit: {"Rank":8,"Suit":2}. The browser wants tidy,
// self-describing data instead: {"rank":8,"suit":"Hearts"}.
//
// We hand back the suit's NAME (via its Stringer) so the front-end never has
// to know that Hearts is internally the integer 2 — the same "prettify at the
// boundary" habit as Card.String. Implemented with a value receiver because,
// like String, it only reads the card.
func (c Card) MarshalJSON() ([]byte, error) {
	// An anonymous struct with json tags is the cleanest way to define a
	// one-off output shape. json.Marshal turns it into the bytes we return.
	return json.Marshal(struct {
		Rank int    `json:"rank"`
		Suit string `json:"suit"`
	}{
		Rank: c.Rank,
		Suit: c.Suit.String(),
	})
}
