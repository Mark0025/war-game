package game

// ─── Player ──────────────────────────────────────────────────────────────
//
// A Player is a name plus a hand of cards held in an order that MATTERS.
// The hand is the single most important data structure in War, so read slowly.
//
// We model the hand as a QUEUE built out of a slice. A queue has two ends and
// we use both:
//
//	index 0 (the front) = the TOP of the hand  -> you DRAW from here
//	the end  (appended) = the BOTTOM of the hand -> you WIN cards to here
//
// Top and bottom being different ends is the actual engine of War. Cards you
// win cycle to the back and won't come up again until you've played through
// everything ahead of them. Without that two-ends rule the game's rhythm
// breaks and it never ends. Go has no built-in queue type — we build one from
// a slice, and that's the whole trick.
type Player struct {
	Name string
	Hand []Card // index 0 = top (draw); append = bottom (win)
}

// HasCards reports whether the player still has any cards. When this returns
// false, the player has lost — that's the entire win condition of the game.
//
// This is a VALUE receiver (p Player) and that's fine: it only READS len(Hand),
// it changes nothing. Reading through a copy gives the same answer as reading
// through the original, so there's no reason to take a pointer here.
func (p Player) HasCards() bool {
	return len(p.Hand) > 0
}

// Draw removes and returns the TOP card (front of the queue).
//
// ★★★ THIS is the value-vs-pointer lesson the whole project builds toward. ★★★
//
// The receiver is a POINTER: (p *Player), not (p Player). Here's why it MUST be.
//
// Draw mutates the hand — it shortens it by one card via p.Hand = p.Hand[1:].
// A method with a VALUE receiver operates on a *copy* of the Player. If Draw
// used (p Player), it would reslice the COPY's hand, return the card, and then
// throw the copy away — the real player's hand would be untouched. The card
// would "refuse to leave the hand." You'd draw the 2 of Spades a thousand
// times and it would never go away.
//
// With a POINTER receiver (p *Player), p points at the REAL player, so
// p.Hand = p.Hand[1:] shrinks the actual hand. The mutation sticks.
//
// The rule you can carry forever:
//   - method only READS the struct        -> value receiver is fine  (HasCards)
//   - method MUTATES the struct's contents -> pointer receiver required (Draw)
//
// (We don't guard against an empty hand here on purpose — the caller checks
// HasCards() first. Drawing from an empty hand SHOULD crash loudly in dev
// rather than silently returning a fake card; that surfaces the bug. Phase 4's
// round loop always checks HasCards before drawing.)
func (p *Player) Draw() Card {
	c := p.Hand[0]      // grab the top card
	p.Hand = p.Hand[1:] // reslice past it: the hand is now one shorter
	return c
}

// Collect adds won cards to the BOTTOM of the hand (the back of the queue).
//
// Also a pointer receiver, for the same reason as Draw: it grows the real
// hand. The `cards ...Card` is a VARIADIC parameter — it accepts any number
// of cards: Collect(c1), Collect(c1, c2), or Collect(pot...) to spread a whole
// slice in. append handles them all in one shot.
//
// Winning to the BOTTOM (not the top) is what keeps the game progressing:
// freshly won cards go to the back of the line and won't be replayed until the
// player has cycled through everything ahead of them.
func (p *Player) Collect(cards ...Card) {
	p.Hand = append(p.Hand, cards...)
}
