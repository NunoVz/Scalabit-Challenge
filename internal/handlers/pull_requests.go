package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/NunoVz/Scalabit-Challenge/internal/github"
)

type PRHandler struct {
	client github.Client
}

func NewPRHandler(client github.Client) *PRHandler {
	return &PRHandler{
		client: client,
	}
}

func (h *PRHandler) CheckPRStatus(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	idStr := r.PathValue("id")
	prNumber, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid PR ID. Must be a number.", http.StatusBadRequest)
		return
	}

	status, err := h.client.GetPRStatus(r.Context(), owner, repo, prNumber)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			http.Error(w, "Repository, Owner, or PR not found", http.StatusNotFound)
			return
		}
		slog.Error("Error checking PR status", "error", err, "owner", owner, "repo", repo, "pr_number", prNumber)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pr_number": prNumber,
		"status":    status,
	}
	writeJSON(w, http.StatusOK, response)
}
