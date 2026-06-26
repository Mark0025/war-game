package game

import "testing"

// Go's testing is built into the language: files named *_test.go, functions
// named TestXxx(t *testing.T), run with `go test`. No framework to install.
//
// This file shows the TABLE-DRIVEN pattern — Go's signature test style. You
// define a slice of cases (the "table"), then loop over them. Adding a new
// case is one line, and t.Run gives each case its own named sub-test so a
// failure tells you exactly which row broke.

func TestCardString(t *testing.T) {
	// Each case pairs an input Card with the string we expect from it.
	cases := []struct {
		name string
		card Card
		want string
	}{
		{"number card", Card{Rank: 7, Suit: Hearts}, "7 of Hearts"},
		{"ten", Card{Rank: 10, Suit: Spades}, "10 of Spades"},
		{"jack", Card{Rank: 11, Suit: Clubs}, "Jack of Clubs"},
		{"queen", Card{Rank: 12, Suit: Diamonds}, "Queen of Diamonds"},
		{"king", Card{Rank: 13, Suit: Hearts}, "King of Hearts"},
		{"ace", Card{Rank: 14, Suit: Spades}, "Ace of Spades"},
	}

	for _, tc := range cases {
		// t.Run makes a sub-test named after tc.name. Run `go test -v` to see
		// them listed individually.
		t.Run(tc.name, func(t *testing.T) {
			got := tc.card.String()
			if got != tc.want {
				// t.Errorf marks the test failed but keeps going; t.Fatalf
				// would stop this sub-test immediately. The %q verb quotes the
				// string so trailing spaces etc. are visible.
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSuitString(t *testing.T) {
	cases := []struct {
		suit Suit
		want string
	}{
		{Spades, "Spades"},
		{Clubs, "Clubs"},
		{Hearts, "Hearts"},
		{Diamonds, "Diamonds"},
	}
	for _, tc := range cases {
		if got := tc.suit.String(); got != tc.want {
			t.Errorf("Suit(%d).String() = %q, want %q", tc.suit, got, tc.want)
		}
	}
}

// TestCardMarshalJSON guards the web API contract: the browser depends on the
// exact shape {"rank":N,"suit":"Name"}. If someone "tidies" the struct later
// and breaks this, the front-end silently stops rendering cards — this test
// turns that into a loud build failure instead.
func TestCardMarshalJSON(t *testing.T) {
	b, err := Card{Rank: 14, Suit: Hearts}.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON errored: %v", err)
	}
	got := string(b)
	want := `{"rank":14,"suit":"Hearts"}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
