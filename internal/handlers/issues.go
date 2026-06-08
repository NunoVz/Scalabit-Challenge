package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
}

func NewIssueHandler(client github.Client) *IssueHandler {
	return &IssueHandler{
		client: client,
	}
}

func (h *IssueHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage <= 0 {
		perPage = 30
	}

	issues, err := h.client.ListIssues(r.Context(), owner, repo, page, perPage)
	if err != nil {
		slog.Error("Error listing issues", "error", err, "owner", owner, "repo", repo)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, issues)
}

func (h *IssueHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	var req CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	issue, err := h.client.CreateIssue(r.Context(), owner, repo, req.Title, req.Body)
	if err != nil {
		slog.Error("Error creating issue", "error", err, "owner", owner, "repo", repo)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, issue)
}

func (h *IssueHandler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	idStr := r.PathValue("id")
	issueNumber, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid issue ID. Must be a number.", http.StatusBadRequest)
		return
	}

	issue, err := h.client.CloseIssue(r.Context(), owner, repo, issueNumber)
	if err != nil {
		slog.Error("Error closing issue", "error", err, "owner", owner, "repo", repo, "issue_id", issueNumber)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": fmt.Sprintf("Issue %d successfully closed", issueNumber),
		"title":   issue.GetTitle(),
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
}
