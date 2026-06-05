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

	// Init Dependencies
	ghClient, err := github.NewClient(token)
	if err != nil {
		log.Fatalf("Error creating GitHub client: %v", err)
	}
	issueHandler := handlers.NewIssueHandler(ghClient)
	prHandler := handlers.NewPRHandler(ghClient)

	// Endpoints---------------
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", handlers.HealthHandler)

	// Issues endpoints
	mux.HandleFunc("GET /repos/{owner}/{repo}/issues", issueHandler.ListIssues)
	mux.HandleFunc("POST /repos/{owner}/{repo}/issues", issueHandler.CreateIssue)
	mux.HandleFunc("DELETE /repos/{owner}/{repo}/issues/{id}", issueHandler.DeleteIssue)

	// Pull Requests endpoints
	mux.HandleFunc("GET /repos/{owner}/{repo}/prs/{id}/status", prHandler.CheckPRStatus)

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	port := ":8080"
	log.Printf("Server running on port %s", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
