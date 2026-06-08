package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NunoVz/Scalabit-Challenge/internal/github"
	"github.com/NunoVz/Scalabit-Challenge/internal/handlers"
	"github.com/joho/godotenv"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

func auditLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("Request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
			"client_ip", r.RemoteAddr,
		)
	})
}

func main() {
	//.env
	_ = godotenv.Load()
	token := os.Getenv("GITHUB_TOKEN")

	// Init Dependencies
	ghClient, err := github.NewClient(token)
	if err != nil {
		slog.Error("Error creating GitHub client", "error", err)
		os.Exit(1)
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
	slog.Info("Server starting", "port", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           auditLogMiddleware(secureHeadersMiddleware(mux)),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("Server exiting cleanly")
}
