package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/NunoVz/Scalabit-Challenge/internal/github"
)

type CreateIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (r *CreateIssueRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	return nil
}

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

	writeJSON(w, http.StatusOK, issues)
}

func (h *IssueHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var req CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	issue, err := h.client.CreateIssue(r.Context(), h.owner, h.repo, req.Title, req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating issue: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, issue)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
