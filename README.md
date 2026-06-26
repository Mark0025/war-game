# War — a Go + Git learning project

A command-line implementation of the card game **War**, built deliberately
slowly to learn two things at once:

1. **Go fundamentals** — structs, methods, value vs. pointer receivers,
   slices, `iota`, interfaces, table-driven tests.
2. **A real Git workflow** — one concept per branch, every change through a
   pull request, branch protection on `main`, and CI that must pass before
   anything merges.

The game is simple on purpose. Nothing hides behind it, so every line is
about a concept, not about the game being clever.

## How to run it

```bash
go run .          # play one game with a random shuffle
go test ./...     # run the test suite (added in a later phase)
```

## The rules of War (plain English)

- Deck of 52 is split evenly: 26 cards each.
- Each round, both players reveal their **top** card. **Higher rank wins**
  both cards, which go to the **bottom** of the winner's hand.
- Suit never matters — it's only for display. Equal rank is a **tie → WAR**.
- **War:** each player lays 3 cards face-down, then 1 face-up. The higher
  face-up card takes the whole pile. Another tie → war again, stacking.
- If a player can't field enough cards for a war, they **lose the game**.
- Because War can loop near-forever, the simulation has a max-round cap and
  falls back to "most cards wins."

## Learning roadmap

Each phase is one branch → one pull request → one review → one merge.

| Phase | Go concept                              | Git concept                          |
|-------|-----------------------------------------|--------------------------------------|
| 0     | `go mod init`, project layout           | repo init, `.gitignore`, first push  |
| 1     | `Card` struct, `iota`, `Stringer`       | feature branch → PR → review → merge |
| 2     | `Deck`, slices, seeded shuffle          | branch protection (no direct push)   |
| 3     | `Player`, pointer vs. value receivers   | atomic commits, good messages        |
| 4     | game loop + the war edge case           | keeping history clean                |
| 5     | table-driven tests, seeded RNG          | required CI status check             |
| 6     | `Stats`, refactor, polish               | tags / releases (`v0.1.0`)           |

## Project layout

```
war/
├── go.mod
├── main.go                 # entry point: wire up a game and play it
└── internal/
    └── game/               # the game lives here; `internal/` = private to this module
        ├── card.go         # Card, Suit, Rank — the dumb data
        ├── deck.go         # build 52, shuffle, deal
        ├── player.go       # a hand as a queue (draw top, win to bottom)
        └── game.go         # the round loop + war resolution
```

> `internal/` is a Go convention: packages under it can only be imported by
> code in this same module. It's how Go enforces "this is private API."
