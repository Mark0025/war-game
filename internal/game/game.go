package game

import (
	"fmt"
	"math/rand"
)

// ─── Game ────────────────────────────────────────────────────────────────
//
// The first three types (Card, Deck, Player) are NOUNS — data that sits there.
// Game is the VERB: it holds the two players and runs rounds until someone is
// out of cards. Nothing about the language is new here; this phase is all
// about COMPOSITION — small correct pieces assembling into behavior.
type Game struct {
	P1, P2 *Player    // pointers: the loop mutates both hands constantly
	rng    *rand.Rand // kept around so we can shuffle the pot (see playRound)
}

// New builds a ready-to-play game: fresh deck, seeded shuffle, dealt 26/26.
//
// The seed is a parameter (not read from the clock in here) so a caller can
// reproduce an exact match — clock-seed for real play, fixed-seed for tests.
// That's the seeded-RNG design from Phase 2 paying off.
//
// `rng` is lowercase, so it's UNEXPORTED — private to package game. Capitalized
// names (P1, New, Card) are exported/public; lowercase ones are package-private.
// That capitalization IS Go's access control — there's no public/private
// keyword.
func New(name1, name2 string, seed int64) *Game {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(seed))
	deck.Shuffle(rng)
	h1, h2 := deck.Deal()
	return &Game{
		P1:  &Player{Name: name1, Hand: h1},
		P2:  &Player{Name: name2, Hand: h2},
		rng: rng,
	}
}

// playRound plays ONE full round, including any chained wars, and hands the
// pot to the winner.
//
// The three return values express every possible outcome of a round:
//   - winner, loser, decided=true  -> the GAME ended this round (someone ran
//     out of cards, possibly mid-war). The caller should announce the winner.
//   - nil, nil, decided=false      -> the round resolved normally; play on.
//
// Returning a small bundle of named results like this is very idiomatic Go —
// no exceptions, no special "game over" signal object, just plain values the
// caller switches on.
func (g *Game) playRound() (winner *Player, loser *Player, decided bool) {
	pot := []Card{} // the cards on the table this round; winner takes all

	for {
		// Both players need at least one card to face off. If either is empty
		// AT THE START of a face-off, the game is over.
		if !g.P1.HasCards() {
			return g.P2, g.P1, true
		}
		if !g.P2.HasCards() {
			return g.P1, g.P2, true
		}

		c1 := g.P1.Draw()
		c2 := g.P2.Draw()
		pot = append(pot, c1, c2)

		fmt.Printf("  %s plays %-16s | %s plays %s\n", g.P1.Name, c1, g.P2.Name, c2)

		switch {
		case c1.Rank > c2.Rank:
			g.collectPot(g.P1, pot)
			return nil, nil, false
		case c2.Rank > c1.Rank:
			g.collectPot(g.P2, pot)
			return nil, nil, false
		default:
			// Equal rank -> WAR. The tie stays on the table; both players lay
			// down stakes, then the loop comes back around and draws the next
			// face-up pair to compare.
			fmt.Println("  ⚔️  WAR!")
			if w, l, d := g.layWarStakes(&pot); d {
				// A player ran out mid-war: the other wins the whole game.
				return w, l, true
			}
			// not decided: loop again, next iteration draws the face-up cards
		}
	}
}

// layWarStakes moves up to 3 face-down cards from each player into the pot.
//
// THE EDGE CASE that separates a toy from a correct program: a player may not
// have 3 cards to stake (or may run out partway). If a player can't keep
// funding the war, they LOSE THE GAME — we must handle "ran out of cards
// mid-war," not only "ran out at the start of a round."
//
// We take *[]Card (a pointer to the slice) so appends here are visible to the
// caller's pot. (Passing the slice by value would append to a copy — the same
// class of bug as the Player value-receiver trap from Phase 3.)
func (g *Game) layWarStakes(pot *[]Card) (winner, loser *Player, decided bool) {
	for i := 0; i < 3; i++ {
		if !g.P1.HasCards() {
			return g.P2, g.P1, true
		}
		if !g.P2.HasCards() {
			return g.P1, g.P2, true
		}
		*pot = append(*pot, g.P1.Draw(), g.P2.Draw())
	}
	return nil, nil, false
}

// collectPot shuffles the pot, then gives it to the winner's bottom.
//
// Why shuffle before collecting? Without it, two evenly matched players can
// fall into a DETERMINISTIC infinite loop — the same cards cycle in the same
// order forever and the game never ends. Shuffling the won pile breaks that
// symmetry, so games terminate naturally. (This is the spec's "add
// rng.Shuffle(pot) before Collect" note, included on purpose because we want
// games that actually finish.)
func (g *Game) collectPot(p *Player, pot []Card) {
	g.rng.Shuffle(len(pot), func(i, j int) {
		pot[i], pot[j] = pot[j], pot[i]
	})
	p.Collect(pot...)
}

// Play runs the whole game to a conclusion (or to a round cap, since War can
// theoretically run a very long time even with pot-shuffling).
//
// The maxRounds cap is a TERMINATION GUARANTEE: a correct simulation must be
// guaranteed to stop. If we hit the cap, we declare the winner by who holds
// more cards — a sensible tiebreak when nobody has been knocked out yet.
func (g *Game) Play(maxRounds int) {
	for round := 1; round <= maxRounds; round++ {
		winner, _, decided := g.playRound()
		if decided {
			fmt.Printf("\n🏆 %s wins the game on round %d!\n", winner.Name, round)
			return
		}
		fmt.Printf("After round %d: %s=%d, %s=%d\n\n",
			round, g.P1.Name, len(g.P1.Hand), g.P2.Name, len(g.P2.Hand))
	}

	// Reached the cap without a knockout: most cards wins.
	fmt.Printf("\nReached %d rounds. Deciding on card count.\n", maxRounds)
	switch {
	case len(g.P1.Hand) > len(g.P2.Hand):
		fmt.Printf("🏆 %s wins on cards (%d vs %d)!\n", g.P1.Name, len(g.P1.Hand), len(g.P2.Hand))
	case len(g.P2.Hand) > len(g.P1.Hand):
		fmt.Printf("🏆 %s wins on cards (%d vs %d)!\n", g.P2.Name, len(g.P2.Hand), len(g.P1.Hand))
	default:
		fmt.Println("🤝 Draw!")
	}
}
