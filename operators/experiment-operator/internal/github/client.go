package github

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"

	gh "github.com/google/go-github/v68/github"
)

// Client wraps the GitHub Contents API for committing experiment results.
type Client struct {
	client *gh.Client
	owner  string
	repo   string
	branch string
	path   string // e.g. "site/data"
}

// NewClient creates a GitHub client for committing experiment results.
// owner and repo are parsed from the "owner/repo" format.
func NewClient(token, owner, repo, branch, path string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		client: gh.NewClient(tc),
		owner:  owner,
		repo:   repo,
		branch: branch,
		path:   path,
	}
}

// RepoPath returns "owner/repo" for logging.
func (c *Client) RepoPath() string {
	return c.owner + "/" + c.repo
}

// CreateBranch creates a new branch from the current HEAD of the configured base branch.
func (c *Client) CreateBranch(ctx context.Context, branchName string) error {
	// Get the SHA of the base branch
	baseRef, _, err := c.client.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/"+c.branch)
	if err != nil {
		return fmt.Errorf("get base branch %s ref: %w", c.branch, err)
	}

	// Create the new branch pointing at the same SHA
	newRef := &gh.Reference{
		Ref:    gh.Ptr("refs/heads/" + branchName),
		Object: baseRef.Object,
	}
	_, _, err = c.client.Git.CreateRef(ctx, c.owner, c.repo, newRef)
	if err != nil {
		return fmt.Errorf("create branch %s: %w", branchName, err)
	}

	return nil
}

// CreatePR creates a pull request from branchName into the base branch.
// Returns the PR number and HTML URL.
func (c *Client) CreatePR(ctx context.Context, title, body, branchName string) (int, string, error) {
	pr, _, err := c.client.PullRequests.Create(ctx, c.owner, c.repo, &gh.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &branchName,
		Base:  &c.branch,
	})
	if err != nil {
		return 0, "", fmt.Errorf("create PR from %s to %s: %w", branchName, c.branch, err)
	}

	return pr.GetNumber(), pr.GetHTMLURL(), nil
}

// PublishExperimentResult creates a branch, commits results to it, and opens a PR.
// Returns the branch name, PR number, and PR URL.
func (c *Client) PublishExperimentResult(ctx context.Context, expName string, summary any) (string, int, string, error) {
	branchName := "experiment/" + expName

	// Create the experiment branch from the base branch
	if err := c.CreateBranch(ctx, branchName); err != nil {
		return "", 0, "", fmt.Errorf("create experiment branch: %w", err)
	}

	// Commit results to the experiment branch
	if err := c.commitToBranch(ctx, branchName, expName, summary); err != nil {
		return "", 0, "", fmt.Errorf("commit results to branch: %w", err)
	}

	// Open a PR
	title := fmt.Sprintf("data: Add %s experiment results", expName)
	body := fmt.Sprintf("Experiment `%s` completed. Results committed to `%s` for review.\n\nPreview locally:\n```\ngit fetch && git checkout %s\ncd site && npm run dev\n```", expName, c.path+"/"+expName+".json", branchName)
	prNum, prURL, err := c.CreatePR(ctx, title, body, branchName)
	if err != nil {
		return branchName, 0, "", fmt.Errorf("create PR: %w", err)
	}

	return branchName, prNum, prURL, nil
}

// commitToBranch commits an experiment summary JSON to the given branch.
func (c *Client) commitToBranch(ctx context.Context, branch, expName string, summary any) error {
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal summary JSON: %w", err)
	}

	filePath := c.path + "/" + expName + ".json"
	commitMsg := fmt.Sprintf("data: Add %s experiment results", expName)

	// Check if file already exists on the target branch (need SHA for updates)
	existing, _, resp, err := c.client.Repositories.GetContents(ctx, c.owner, c.repo, filePath, &gh.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil && (resp == nil || resp.StatusCode != 404) {
		return fmt.Errorf("check existing file %s: %w", filePath, err)
	}

	opts := &gh.RepositoryContentFileOptions{
		Message: &commitMsg,
		Content: body,
		Branch:  &branch,
	}

	// If file exists, include SHA for update
	if existing != nil {
		sha := existing.GetSHA()
		opts.SHA = &sha
		commitMsg = fmt.Sprintf("data: Update %s experiment results", expName)
		opts.Message = &commitMsg
	}

	_, _, err = c.client.Repositories.CreateFile(ctx, c.owner, c.repo, filePath, opts)
	if err != nil {
		return fmt.Errorf("commit %s: %w", filePath, err)
	}

	return nil
}

// CommitResult commits an experiment summary JSON to the configured repo path.
// It creates or updates site/data/{expName}.json with indented JSON.
func (c *Client) CommitResult(ctx context.Context, expName string, summary any) error {
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal summary JSON: %w", err)
	}

	filePath := c.path + "/" + expName + ".json"
	commitMsg := fmt.Sprintf("data: Add %s experiment results", expName)

	// Check if file already exists (need SHA for updates)
	existing, _, resp, err := c.client.Repositories.GetContents(ctx, c.owner, c.repo, filePath, &gh.RepositoryContentGetOptions{
		Ref: c.branch,
	})
	if err != nil && (resp == nil || resp.StatusCode != 404) {
		return fmt.Errorf("check existing file %s: %w", filePath, err)
	}

	opts := &gh.RepositoryContentFileOptions{
		Message: &commitMsg,
		Content: body,
		Branch:  &c.branch,
	}

	// If file exists, include SHA for update
	if existing != nil {
		sha := existing.GetSHA()
		opts.SHA = &sha
		commitMsg = fmt.Sprintf("data: Update %s experiment results", expName)
		opts.Message = &commitMsg
	}

	_, _, err = c.client.Repositories.CreateFile(ctx, c.owner, c.repo, filePath, opts)
	if err != nil {
		return fmt.Errorf("commit %s: %w", filePath, err)
	}

	return nil
}
