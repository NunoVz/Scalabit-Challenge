package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v88/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

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

	mux := http.NewServeMux()

	// Endpoint de health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})


	//List issues endpoint
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
		json.NewEncoder(w).Encode(issues)
	})

	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Error: %v", err)
	}
}