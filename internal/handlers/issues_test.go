package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gh "github.com/google/go-github/v88/github"
)

type mockGitHubClient struct {
	mockCreateIssue func(ctx context.Context, owner, repo, title, body string) (*gh.Issue, error)
}

func (m *mockGitHubClient) ListIssues(ctx context.Context, owner, repo string) ([]*gh.Issue, error) {
	return nil, nil
}

func (m *mockGitHubClient) CreateIssue(ctx context.Context, owner, repo, title, body string) (*gh.Issue, error) {
	if m.mockCreateIssue != nil {
		return m.mockCreateIssue(ctx, owner, repo, title, body)
	}
	return nil, nil
}

func (m *mockGitHubClient) CloseIssue(ctx context.Context, owner, repo string, issueNumber int) (*gh.Issue, error) {
	return nil, nil
}

func (m *mockGitHubClient) GetPRStatus(ctx context.Context, owner, repo string, prNumber int) (string, error) {
	return "", nil
}

func TestIssueHandler_CreateIssue(t *testing.T) {
	owner := "test-owner"
	repo := "test-repo"

	tests := []struct {
		name           string
		reqBody        string /
		mockCreate     func(ctx context.Context, owner, repo, title, body string) (*gh.Issue, error)
		expectedStatus int
		expectedSubstr string
	}{
		{
			name:    "Success",
			reqBody: `{"title": "Test Issue", "body": "This is a test issue"}`,
			mockCreate: func(ctx context.Context, o, r, title, body string) (*gh.Issue, error) {
				id := int64(123)
				return &gh.Issue{ID: &id, Title: &title, Body: &body}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedSubstr: `"title":"Test Issue"`,
		},
		{
			name:           "Invalid JSON",
			reqBody:        "{ invalid json",
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedSubstr: "Invalid request body",
		},
		{
			name:           "Validation Error - Empty Title",
			reqBody:        `{"title": "", "body": "No title here"}`,
			mockCreate:     nil, 
			expectedStatus: http.StatusBadRequest,
			expectedSubstr: "title is required",
		},
		{
			name:    "GitHub API Error",
			reqBody: `{"title": "Error Issue", "body": "This should fail at the mock"}`,
			mockCreate: func(ctx context.Context, o, r, title, body string) (*gh.Issue, error) {
				return nil, errors.New("simulated github api error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedSubstr: "Error creating issue: simulated github api error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGitHubClient{mockCreateIssue: tt.mockCreate}
			handler := NewIssueHandler(mockClient, owner, repo)

			req, err := http.NewRequest("POST", "/issues", strings.NewReader(tt.reqBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.CreateIssue(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedSubstr != "" && !strings.Contains(rr.Body.String(), tt.expectedSubstr) {
				t.Errorf("handler returned unexpected body: got %v want substring %v", rr.Body.String(), tt.expectedSubstr)
			}
		})
	}
}
