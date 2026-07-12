//go:build ignore

package main

import (
	"log"
	"net/http"
)

func main() {
	port := ":8080"
	log.Printf("Serving WASM app on http://localhost%s...\n", port)
	err := http.ListenAndServe(port, http.FileServer(http.Dir("wasm")))
	if err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
