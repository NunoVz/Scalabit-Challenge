package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/NunoVz/Scalabit-Challenge/internal/github"
)

type PRHandler struct {
	client github.Client
	owner  string
	repo   string
}

func NewPRHandler(client github.Client, owner, repo string) *PRHandler {
	return &PRHandler{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (h *PRHandler) CheckPRStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	prNumber, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid PR ID. Must be a number.", http.StatusBadRequest)
		return
	}

	status, err := h.client.GetPRStatus(r.Context(), h.owner, h.repo, prNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking PR status: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pr_number": prNumber,
		"status":    status,
	}
	writeJSON(w, http.StatusOK, response)
}