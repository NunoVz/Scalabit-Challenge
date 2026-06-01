package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Endpoint de health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Error: %v", err)
	}
}