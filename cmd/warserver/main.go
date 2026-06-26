// Command warserver runs War as a browser game on http://localhost:8080.
//
// This is a SECOND binary in the same module. Go finds a separate executable
// for each `package main` under cmd/. So:
//
//	go run .                 -> the auto-play CLI (root main.go)
//	go run ./cmd/warserver   -> this web server
//
// Keeping each entry point tiny and letting the packages do the work is the
// idiomatic Go layout.
package main

import (
	"log"
	"net/http"

	"war/web"
)

func main() {
	addr := ":8080"
	handler := web.NewServer()

	log.Printf("War web game running → http://localhost%s", addr)
	// ListenAndServe blocks until the server stops (or errors). log.Fatal
	// prints the error and exits non-zero if it ever returns.
	log.Fatal(http.ListenAndServe(addr, handler))
}
