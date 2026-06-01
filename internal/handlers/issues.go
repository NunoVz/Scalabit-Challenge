package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/NunoVz/Scalabit-Challenge/internal/github"
)

type IssueHandler struct {
	client github.Client
	owner  string
	repo   string
}

func NewIssueHandler(client github.Client, owner, repo string) *IssueHandler {
	return &IssueHandler{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (h *IssueHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	issues, err := h.client.ListIssues(r.Context(), h.owner, h.repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing issues: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(issues); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
