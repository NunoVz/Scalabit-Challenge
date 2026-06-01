package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/NunoVz/Scalabit-Challenge/internal/github"
	"github.com/NunoVz/Scalabit-Challenge/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	//.env
	_ = godotenv.Load()
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	// Init Dependencies
	ghClient, err := github.NewClient(token)
	if err != nil {
		log.Fatalf("Error creating GitHub client: %v", err)
	}
	issueHandler := handlers.NewIssueHandler(ghClient, owner, repo)

	// Endpoints---------------
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", handlers.HealthHandler)

	// Issues endpoints
	mux.HandleFunc("GET /issues", issueHandler.ListIssues)

	port := ":8080"
	log.Printf("Server running on port %s", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, 
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}