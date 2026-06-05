package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPRHandler_CheckPRStatus(t *testing.T) {
	owner := "test-owner"
	repo := "test-repo"

	tests := []struct {
		name           string
		prID           string
		mockGetStatus  func(ctx context.Context, owner, repo string, prNumber int) (string, error)
		expectedStatus int
		expectedSubstr string
	}{
		{
			name: "Success",
			prID: "42",
			mockGetStatus: func(ctx context.Context, owner, repo string, prNumber int) (string, error) {
				return "success", nil
			},
			expectedStatus: http.StatusOK,
			expectedSubstr: `"status":"success"`,
		},
		{
			name: "Failure",
			prID: "42",
			mockGetStatus: func(ctx context.Context, owner, repo string, prNumber int) (string, error) {
				return "failure", nil
			},
			expectedStatus: http.StatusOK,
			expectedSubstr: `"status":"failure"`,
		},
		{
			name: "Pending",
			prID: "42",
			mockGetStatus: func(ctx context.Context, owner, repo string, prNumber int) (string, error) {
				return "pending", nil
			},
			expectedStatus: http.StatusOK,
			expectedSubstr: `"status":"pending"`,
		},
		{
			name:           "Invalid ID Format",
			prID:           "abc",
			mockGetStatus:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedSubstr: "Invalid PR ID. Must be a number.",
		},
		{
			name: "GitHub API Error",
			prID: "42",
			mockGetStatus: func(ctx context.Context, owner, repo string, prNumber int) (string, error) {
				return "", errors.New("simulated PR status error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedSubstr: "Error checking PR status: simulated PR status error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGitHubClient{mockGetPRStatus: tt.mockGetStatus}
			handler := NewPRHandler(mockClient)

			url := "/repos/" + owner + "/" + repo + "/prs/" + tt.prID + "/status"
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.SetPathValue("owner", owner)
			req.SetPathValue("repo", repo)
			req.SetPathValue("id", tt.prID)

			rr := httptest.NewRecorder()

			handler.CheckPRStatus(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
			if tt.expectedSubstr != "" && !strings.Contains(rr.Body.String(), tt.expectedSubstr) {
				t.Errorf("handler returned unexpected body: got %v want substring %v", rr.Body.String(), tt.expectedSubstr)
			}
		})
	}
}
