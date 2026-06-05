package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting cleanly")
}
