// Package web serves War as a browser game. The Go game engine stays the
// single source of truth; this package just exposes it over HTTP and ships the
// HTML/CSS/JS that draws the felt.
package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"war/internal/game"
)

// ─── Embedding the front-end ───────────────────────────────────────────────
//
// The //go:embed directive bakes the static/ folder INTO the compiled binary
// at build time. The result: `go build` produces ONE self-contained executable
// with the HTML/CSS/JS inside it — no "ship the assets alongside the binary"
// problem. (This is the lesson from a recent gotcha: embed runtime assets, do
// not read them off disk at runtime.) The blank line between the directive and
// the var is NOT allowed — the comment+directive must sit directly above it.
//
//go:embed static
var staticFS embed.FS

// ─── Server state ──────────────────────────────────────────────────────────
//
// One game at a time (single-player vs the computer). The mutex guards it
// because an HTTP server handles requests CONCURRENTLY — two clicks could
// arrive at once, and Step() mutates shared state. A sync.Mutex is Go's basic
// "only one goroutine in here at a time" lock. Forgetting it would be a data
// race; `go test -race` (Phase 6) would catch it.
type Server struct {
	mu   sync.Mutex
	game *game.WebGame
}

// NewServer wires up the HTTP routes and returns something you can hand to
// http.ListenAndServe.
func NewServer() http.Handler {
	s := &Server{game: game.NewWebGame(time.Now().UnixNano())}

	mux := http.NewServeMux()

	// Serve the embedded static files. We strip the "static" prefix from the
	// embed FS so "/static/style.css" maps to the file "static/style.css".
	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	// The index page.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, _ := staticFS.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})

	// API: start a fresh game.
	mux.HandleFunc("/api/new", s.handleNew)
	// API: advance one step.
	mux.HandleFunc("/api/step", s.handleStep)

	return mux
}

// handleNew resets to a brand-new shuffled game and returns the opening state.
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock() // defer runs when the function returns — a tidy way to
	//                      guarantee the lock is released no matter which path
	//                      we exit through.

	s.game = game.NewWebGame(time.Now().UnixNano())
	// Hand back the starting counts (26/26) via a zero-value-ish StepResult.
	writeJSON(w, game.StepResult{
		OutcomeID: "New",
		Message:   "New game — 26 cards each.",
		Count1:    len(s.game.P1.Hand),
		Count2:    len(s.game.P2.Hand),
	})
}

// handleStep advances the current game by one face-off.
func (s *Server) handleStep(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := s.game.Step()
	writeJSON(w, result)
}

// writeJSON marshals any value to JSON and writes it with the right header.
// `v any` accepts anything; json.NewEncoder streams it to the response writer.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
