package github

import (
	"context"
	"net/http"

	gh "github.com/google/go-github/v88/github"
	"golang.org/x/oauth2"
)

// Client establish Git Contact
type Client interface {
	ListIssues(ctx context.Context, owner, repo string) ([]*gh.Issue, error)
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
