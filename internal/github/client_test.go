package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gh "github.com/google/go-github/v88/github"
)

func setupMockClient(mux *http.ServeMux) (*httptest.Server, Client) {
	server := httptest.NewServer(mux)

	baseURL := server.URL + "/"
	ghClient, _ := gh.NewClient(gh.WithEnterpriseURLs(baseURL, baseURL))

	client := &gitHubClient{client: ghClient}
	return server, client
}

func TestGetPRStatus(t *testing.T) {
	tests := []struct {
		name           string
		mockPR         func(w http.ResponseWriter, r *http.Request)
		mockChecks     func(w http.ResponseWriter, r *http.Request)
		expectedStatus string
		expectError    bool
	}{
		{
			name: "Success",
			mockPR: func(w http.ResponseWriter, r *http.Request) {
				pr := gh.PullRequest{Head: &gh.PullRequestBranch{SHA: gh.Ptr("mock-sha")}}
				_ = json.NewEncoder(w).Encode(pr)
			},
			mockChecks: func(w http.ResponseWriter, r *http.Request) {
				checks := gh.ListCheckRunsResults{
					Total:     gh.Ptr(1),
					CheckRuns: []*gh.CheckRun{{Status: gh.Ptr("completed"), Conclusion: gh.Ptr("success")}},
				}
				_ = json.NewEncoder(w).Encode(checks)
			},
			expectedStatus: "success",
			expectError:    false,
		},
		{
			name: "Error PR",
			mockPR: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectError: true,
		},
		{
			name: "Error Checks",
			mockPR: func(w http.ResponseWriter, r *http.Request) {
				pr := gh.PullRequest{Head: &gh.PullRequestBranch{SHA: gh.Ptr("mock-sha")}}
				_ = json.NewEncoder(w).Encode(pr)
				_ = json.NewEncoder(w).Encode(pr)
			},
			mockChecks: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectError: true,
		},
		{
			name: "No Checks Found",
			mockPR: func(w http.ResponseWriter, r *http.Request) {
				pr := gh.PullRequest{Head: &gh.PullRequestBranch{SHA: gh.Ptr("mock-sha")}}
				json.NewEncoder(w).Encode(pr)
			},
			mockChecks: func(w http.ResponseWriter, r *http.Request) {
				checks := gh.ListCheckRunsResults{Total: gh.Ptr(0)}
				_ = json.NewEncoder(w).Encode(checks)
			},
			expectedStatus: "no_checks_found",
		},
		{
			name: "Pending",
			mockPR: func(w http.ResponseWriter, r *http.Request) {
				pr := gh.PullRequest{Head: &gh.PullRequestBranch{SHA: gh.Ptr("mock-sha")}}
				_ = json.NewEncoder(w).Encode(pr)
			},
			mockChecks: func(w http.ResponseWriter, r *http.Request) {
				checks := gh.ListCheckRunsResults{
					Total:     gh.Ptr(1),
					CheckRuns: []*gh.CheckRun{{Status: gh.Ptr("in_progress")}},
				}
				_ = json.NewEncoder(w).Encode(checks)
			},
			expectedStatus: "pending",
		},
		{
			name: "Failure",
			mockPR: func(w http.ResponseWriter, r *http.Request) {
				pr := gh.PullRequest{Head: &gh.PullRequestBranch{SHA: gh.Ptr("mock-sha")}}
				_ = json.NewEncoder(w).Encode(pr)
			},
			mockChecks: func(w http.ResponseWriter, r *http.Request) {
				checks := gh.ListCheckRunsResults{
					Total:     gh.Ptr(1),
					CheckRuns: []*gh.CheckRun{{Status: gh.Ptr("completed"), Conclusion: gh.Ptr("failure")}},
				}
				_ = json.NewEncoder(w).Encode(checks)
			},
			expectedStatus: "failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()

			if tt.mockPR != nil {
				mux.HandleFunc("/api/v3/repos/test-owner/test-repo/pulls/1", tt.mockPR)
			}
			if tt.mockChecks != nil {
				mux.HandleFunc("/api/v3/repos/test-owner/test-repo/commits/mock-sha/check-runs", tt.mockChecks)
			}

			server, client := setupMockClient(mux)
			defer server.Close()

			status, err := client.GetPRStatus(context.Background(), "test-owner", "test-repo", 1)

			if (err != nil) != tt.expectError {
				t.Fatalf("Expected error: %v, got: %v", tt.expectError, err)
			}
			if !tt.expectError && status != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, status)
			}
		})
	}
}

func TestListIssues(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v3/repos/test-owner/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		issues := []*gh.Issue{
			{
				Number: gh.Ptr(1),
				Title:  gh.Ptr("Mock Issue"),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(issues)
	})

	server, client := setupMockClient(mux)
	defer server.Close()

	issues, err := client.ListIssues(context.Background(), "test-owner", "test-repo", 1, 30)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, but got %d", len(issues))
	}
	if issues[0].GetTitle() != "Mock Issue" {
		t.Errorf("Title does not match, got: %s", issues[0].GetTitle())
	}
}

func TestNewClient(t *testing.T) {
	_, err := NewClient("mock-token")
	if err != nil {
		t.Errorf("Did not expect an error when creating the client: %v", err)
	}
}

func TestNewClient_EmptyToken(t *testing.T) {
	_, err := NewClient("")
	if err != nil {
		t.Errorf("Did not expect an error when creating the client: %v", err)
	}
}

func TestCreateIssue(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v3/repos/test-owner/test-repo/issues", func(w http.ResponseWriter, r *http.Request) {
		issue := gh.Issue{
			Number: gh.Ptr(1),
			Title:  gh.Ptr("Mock Issue"),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(issue)
	})

	server, client := setupMockClient(mux)
	defer server.Close()

	issue, err := client.CreateIssue(context.Background(), "test-owner", "test-repo", "Mock Issue", "Mock Body")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if issue.GetTitle() != "Mock Issue" {
		t.Errorf("Expected 'Mock Issue', but got '%s'", issue.GetTitle())
	}
}

func TestCloseIssue(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v3/repos/test-owner/test-repo/issues/1", func(w http.ResponseWriter, r *http.Request) {
		issue := gh.Issue{
			Number: gh.Ptr(1),
			State:  gh.Ptr("closed"),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(issue)
	})

	server, client := setupMockClient(mux)
	defer server.Close()

	issue, err := client.CloseIssue(context.Background(), "test-owner", "test-repo", 1)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if issue.GetState() != "closed" {
		t.Errorf("Expected 'closed', but got '%s'", issue.GetState())
	}
}
