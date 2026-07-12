// @ctx: GitLab API client stub — simulates API calls for testing
// @hlv:sec [AUTH_BOUNDARY] — auth token sent with every request

package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
)

type CreateIssueResult struct {
	IssueID  string
	IssueIID string
	IssueURL string
}

type UpdateIssueResult struct {
	IssueID  string
	IssueIID string
	IssueURL string
}

type GitLabClient interface {
	CreateIssue(projectID string, issue GitLabIssue) (*CreateIssueResult, error)
	UpdateIssue(projectID, issueID string, issue GitLabIssue) (*UpdateIssueResult, error)
	CloseIssue(projectID, issueID string) error
	ReopenIssue(projectID, issueID string) error
}

type StubGitLabClient struct {
	mu              sync.RWMutex
	ticketIssueMap  map[string]string
	issueIDCounter  int
	simulateFailure bool
	failureMode     string
}

func NewStubGitLabClient() *StubGitLabClient {
	return &StubGitLabClient{
		ticketIssueMap:  make(map[string]string),
		issueIDCounter:  10000,
		simulateFailure: false,
	}
}

func (c *StubGitLabClient) SetSimulateFailure(mode string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.simulateFailure = true
	c.failureMode = mode
	slog.Warn("gitlab.stub.failure_mode_set", "mode", mode)
}

func (c *StubGitLabClient) ClearSimulateFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.simulateFailure = false
	c.failureMode = ""
	slog.Info("gitlab.stub.failure_mode_cleared")
}

func (c *StubGitLabClient) CreateIssue(projectID string, issue GitLabIssue) (*CreateIssueResult, error) {
	slog.Info("gitlab.create_issue.enter",
		"project_id", projectID,
		"title", issue.Title,
	)

	if err := c.checkFailure("create"); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.issueIDCounter++
	cid := c.issueIDCounter
	issueID := fmt.Sprintf("%d", cid)
	issueIID := fmt.Sprintf("%d", cid-10000)
	url := fmt.Sprintf("https://gitlab.example.com/%s/-/issues/%s", projectID, issueIID)
	c.mu.Unlock()

	result := &CreateIssueResult{
		IssueID:  issueID,
		IssueIID: issueIID,
		IssueURL: url,
	}

	slog.Info("gitlab.create_issue.exit",
		"issue_id", result.IssueID,
		"issue_iid", result.IssueIID,
	)
	return result, nil
}

func (c *StubGitLabClient) UpdateIssue(projectID, issueID string, issue GitLabIssue) (*UpdateIssueResult, error) {
	slog.Info("gitlab.update_issue.enter",
		"project_id", projectID,
		"issue_id", issueID,
	)

	if err := c.checkFailure("update"); err != nil {
		return nil, err
	}

	issueIID := fmt.Sprintf("%d", atoi(issueID)-10000)
	url := fmt.Sprintf("https://gitlab.example.com/%s/-/issues/%s", projectID, issueIID)

	result := &UpdateIssueResult{
		IssueID:  issueID,
		IssueIID: issueIID,
		IssueURL: url,
	}

	slog.Info("gitlab.update_issue.exit",
		"issue_id", result.IssueID,
	)
	return result, nil
}

func (c *StubGitLabClient) CloseIssue(projectID, issueID string) error {
	slog.Info("gitlab.close_issue.enter",
		"project_id", projectID,
		"issue_id", issueID,
	)

	if err := c.checkFailure("close"); err != nil {
		return err
	}

	slog.Info("gitlab.close_issue.exit",
		"issue_id", issueID,
	)
	return nil
}

func (c *StubGitLabClient) ReopenIssue(projectID, issueID string) error {
	slog.Info("gitlab.reopen_issue.enter",
		"project_id", projectID,
		"issue_id", issueID,
	)

	if err := c.checkFailure("reopen"); err != nil {
		return err
	}

	slog.Info("gitlab.reopen_issue.exit",
		"issue_id", issueID,
	)
	return nil
}

// @hlv:sec [INPUT_VALIDATION] — check failure before calling simulated API
func (c *StubGitLabClient) checkFailure(action string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.simulateFailure {
		return nil
	}

	switch c.failureMode {
	case "unavailable":
		slog.Error("gitlab.stub.unavailable", "action", action)
		// @hlv SYNC-GITLAB-UNAVAILABLE
		return fmt.Errorf("simulated: GitLab API unavailable")
	case "rate_limited":
		slog.Error("gitlab.stub.rate_limited", "action", action)
		// @hlv SYNC-RATE-LIMITED
		return fmt.Errorf("simulated: GitLab API rate limited")
	case "conflict":
		slog.Error("gitlab.stub.conflict", "action", action)
		// @hlv SYNC-CONFLICT
		return fmt.Errorf("simulated: GitLab issue conflict")
	case "random":
		if rand.Intn(100) < 30 {
			slog.Error("gitlab.stub.random_failure", "action", action)
			return fmt.Errorf("simulated: random GitLab API failure")
		}
	}

	return nil
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
