package github

import (
	"context"

	gh "github.com/google/go-github/v88/github"
	"golang.org/x/oauth2"
)

// Client establish Git Contact
type Client interface {
	ListIssues(ctx context.Context, owner, repo string) ([]*gh.Issue, error)
	CreateIssue(ctx context.Context, owner, repo, title, body string) (*gh.Issue, error)
	CloseIssue(ctx context.Context, owner, repo string, issueNumber int) (*gh.Issue, error)
	GetPRStatus(ctx context.Context, owner, repo string, prNumber int) (string, error)
}

type gitHubClient struct {
	client *gh.Client
}

func NewClient(token string) (Client, error) {
	ctx := context.Background()
	var client *gh.Client
	var err error

	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client, err = gh.NewClient(gh.WithHTTPClient(tc))
	} else {
		client, err = gh.NewClient()
	}

	if err != nil {
		return nil, err
	}

	return &gitHubClient{client: client}, nil
}

func (g *gitHubClient) ListIssues(ctx context.Context, owner, repo string) ([]*gh.Issue, error) {
	opts := &gh.IssueListByRepoOptions{State: "all"}
	issues, _, err := g.client.Issues.ListByRepo(ctx, owner, repo, opts)
	return issues, err
}

func (g *gitHubClient) CreateIssue(ctx context.Context, owner, repo, title, body string) (*gh.Issue, error) {
	req := &gh.IssueRequest{
		Title: &title,
		Body:  &body,
	}
	issue, _, err := g.client.Issues.Create(ctx, owner, repo, req)
	return issue, err
}

func (g *gitHubClient) CloseIssue(ctx context.Context, owner, repo string, issueNumber int) (*gh.Issue, error) {
	state := "closed"
	req := &gh.IssueRequest{State: &state}
	issue, _, err := g.client.Issues.Edit(ctx, owner, repo, issueNumber, req)
	return issue, err
}

func (g *gitHubClient) GetPRStatus(ctx context.Context, owner, repo string, prNumber int) (string, error) {
	// 1. Get PR Details
	pr, _, err := g.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return "", err
	}

	sha := pr.GetHead().GetSHA()

	// 2. Get the list of checks/pipelines for that commit
	result, _, err := g.client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
	if err != nil {
		return "", err
	}

	if result.GetTotal() == 0 {
		return "no_checks_found", nil
	}

	// 3. Evaluate the pipeline result
	status := "success"
	for _, run := range result.CheckRuns {
		if run.GetStatus() != "completed" {
			status = "pending"
		} else if run.GetConclusion() == "failure" || run.GetConclusion() == "cancelled" || run.GetConclusion() == "timed_out" {
			return "failure", nil 
		}
	}

	return status, nil
}
