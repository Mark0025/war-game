package game

// This file adds a STEP-WISE, SILENT, DATA-RETURNING way to play — the engine
// the web server drives. The original playRound/Play (game.go) prints as it
// goes and resolves a whole round at once, which is perfect for the CLI but
// wrong for a click-to-flip UI. Rather than rip that out, we ADD this layer.
// Same domain types (Card, Deck, Player); a different way to advance.
//
// Key differences from game.go:
//   - returns data (a StepResult) instead of calling fmt.Println
//   - resolves exactly ONE face-off per call, so the UI controls the pace
//   - a war is reported, then funded+resolved across subsequent steps so the
//     browser can animate it

// Outcome describes what a single Step produced. It's a small enumerated type
// (same iota idiom as Suit) so the front-end can branch on it cleanly.
type Outcome int

const (
	RoundWon Outcome = iota // a face-off was decided; someone took the pot
	WarStart                // ranks tied — a war begins; stakes laid this step
	GameOver                // a player ran out of cards (possibly mid-war)
)

func (o Outcome) String() string {
	return [...]string{"RoundWon", "WarStart", "GameOver"}[o]
}

// StepResult is the data the API hands back after each click. Notice the
// `json:"..."` STRUCT TAGS: they tell Go's encoding/json package what to name
// each field in the JSON sent to the browser. Without them you'd get the Go
// field names (Card1) instead of tidy lowercase keys (card1). Tags are how Go
// bridges a struct to an external format.
type StepResult struct {
	Outcome   Outcome `json:"-"`       // omitted from JSON; we send the string below
	OutcomeID string  `json:"outcome"` // "RoundWon" | "WarStart" | "GameOver"
	Card1     *Card   `json:"card1"`   // the card P1 just flipped (nil if none)
	Card2     *Card   `json:"card2"`   // the card P2 just flipped
	PotSize   int     `json:"potSize"` // cards currently on the table
	Message   string  `json:"message"` // human-readable summary for the UI
	Count1    int     `json:"count1"`  // P1's hand size after this step
	Count2    int     `json:"count2"`  // P2's hand size after this step
	Winner    string  `json:"winner"`  // name of game winner, "" until GameOver
}

// WebGame wraps a *Game with the extra round-in-progress state that step play
// needs. Composition over modification: rather than bloat Game, we embed it.
// The pot can persist across several Steps (a war spans multiple clicks), so it
// lives here as unexported bookkeeping the UI never sees directly.
//
// `*Game` with no field name is an EMBEDDED field. Go "promotes" the embedded
// type's methods, so a *WebGame can still call .Play(), .New isn't a method
// etc. It's Go's lightweight alternative to inheritance: has-a that reads
// like is-a.
type WebGame struct {
	*Game
	pot      []Card // cards on the table for the round in progress
	warArmed bool   // true when the previous step started a war we must fund
}

// NewWebGame builds a fresh, shuffled, dealt game ready for step play.
func NewWebGame(seed int64) *WebGame {
	return &WebGame{Game: New("You", "Computer", seed)}
}

// Step advances the game by exactly one face-off and returns what happened.
// The browser calls this once per click.
func (w *WebGame) Step() StepResult {
	// If a war was started last step, fund it first: each player lays up to 3
	// face-down cards. Running out here ends the game.
	if w.warArmed {
		w.warArmed = false
		if win, _, decided := w.layWarStakes(&w.pot); decided {
			return w.gameOver(win)
		}
	}

	// Either player empty at the start of a face-off => game over.
	if !w.P1.HasCards() {
		return w.gameOver(w.P2)
	}
	if !w.P2.HasCards() {
		return w.gameOver(w.P1)
	}

	c1 := w.P1.Draw()
	c2 := w.P2.Draw()
	w.pot = append(w.pot, c1, c2)

	switch {
	case c1.Rank > c2.Rank:
		w.collectPot(w.P1, w.pot)
		return w.resolved(c1, c2, takesMsg(w.P1.Name))
	case c2.Rank > c1.Rank:
		w.collectPot(w.P2, w.pot)
		return w.resolved(c1, c2, takesMsg(w.P2.Name))
	default:
		// Tie: report a war. The NEXT step funds it (warArmed) and the one
		// after draws the deciding face-up pair. This staging is what lets the
		// browser animate "WAR!" before the resolution.
		w.warArmed = true
		return StepResult{
			Outcome:   WarStart,
			OutcomeID: WarStart.String(),
			Card1:     &c1,
			Card2:     &c2,
			PotSize:   len(w.pot),
			Message:   "⚔️ WAR! Both tied on " + rankName(c1.Rank) + " — click to continue.",
			Count1:    len(w.P1.Hand),
			Count2:    len(w.P2.Hand),
		}
	}
}

// takesMsg returns a grammatically correct "X takes the round" line. "You" is
// second person ("You take"); a named third party ("Computer takes") gets the
// -s. A tiny detail, but wrong grammar reads as a bug to a player.
func takesMsg(name string) string {
	if name == "You" {
		return "You take the round!"
	}
	return name + " takes the round."
}

// resolved builds a StepResult for a decided face-off. After collectPot the
// pot is emptied for the next round.
func (w *WebGame) resolved(c1, c2 Card, msg string) StepResult {
	w.pot = nil
	return StepResult{
		Outcome:   RoundWon,
		OutcomeID: RoundWon.String(),
		Card1:     &c1,
		Card2:     &c2,
		PotSize:   0,
		Message:   msg,
		Count1:    len(w.P1.Hand),
		Count2:    len(w.P2.Hand),
	}
}

func (w *WebGame) gameOver(winner *Player) StepResult {
	// Same second-person grammar fix as takesMsg: "You win" vs "Computer wins".
	verb := " wins the game!"
	if winner.Name == "You" {
		verb = " win the game!"
	}
	return StepResult{
		Outcome:   GameOver,
		OutcomeID: GameOver.String(),
		PotSize:   len(w.pot),
		Message:   "🏆 " + winner.Name + verb,
		Count1:    len(w.P1.Hand),
		Count2:    len(w.P2.Hand),
		Winner:    winner.Name,
	}
}

// rankName turns a rank int into its display word, reusing the same mapping
// Card.String relies on. Small helper so war messages read naturally.
func rankName(r int) string {
	switch r {
	case 11:
		return "Jack"
	case 12:
		return "Queen"
	case 13:
		return "King"
	case 14:
		return "Ace"
	default:
		// fmt would work but we avoid importing it here; build the digits.
		return itoa(r)
	}
}

// itoa converts a small positive int (2..10) to a string without importing
// strconv — a tiny illustration that you can build primitives yourself.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
