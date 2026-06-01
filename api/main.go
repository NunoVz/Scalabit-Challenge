package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v88/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprint(w, "OK"); err != nil {
		log.Printf("Error writing health response: %v", err)
	}
}
func main() {
	//.env
	_ = godotenv.Load()
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")


	//GIT AUTH
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	
	client, err := github.NewClient(github.WithHTTPClient(tc))
	if err != nil {
		log.Fatalf("Error creating GitHub client: %v", err)
	}





	// Endpoints---------------
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler)

	// List issues 
	mux.HandleFunc("GET /issues", func(w http.ResponseWriter, r *http.Request) {
		opts := &github.IssueListByRepoOptions{
			State: "all", 
		}

		issues, _, err := client.Issues.ListByRepo(r.Context(), owner, repo, opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error presenting issues: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		
		
		if err := json.NewEncoder(w).Encode(issues); err != nil {
			log.Printf("Error encoding issues JSON: %v", err)
		}
	})
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