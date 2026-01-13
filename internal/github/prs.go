package github

import (
	"fmt"
)

// GetPRForBranch finds a PR for a specific branch
func (c *Client) GetPRForBranch(repo, branch string) (*PullRequest, error) {
	var prs []PullRequest
	err := c.RunJSON(&prs,
		"pr", "list",
		"--repo", repo,
		"--head", branch,
		"--state", "open",
		"--json", "number,title,body,state,url,headRefName,baseRefName,merged,createdAt",
	)
	if err != nil {
		return nil, err
	}

	if len(prs) == 0 {
		return nil, nil
	}
	return &prs[0], nil
}

// GetPR returns a specific PR by number
func (c *Client) GetPR(repo string, number int) (*PullRequest, error) {
	var pr PullRequest
	err := c.RunJSON(&pr,
		"pr", "view", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "number,title,body,state,url,headRefName,baseRefName,merged,createdAt",
	)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// CreatePR creates a new pull request
func (c *Client) CreatePR(repo, title, body, head, base string) (*PullRequest, error) {
	_, err := c.Run(
		"pr", "create",
		"--repo", repo,
		"--title", title,
		"--body", body,
		"--head", head,
		"--base", base,
	)
	if err != nil {
		return nil, err
	}

	// Fetch the created PR
	return c.GetPRForBranch(repo, head)
}

// MergePR merges a pull request
func (c *Client) MergePR(repo string, number int, deleteRemoteBranch bool) error {
	args := []string{
		"pr", "merge", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--merge",
	}

	if deleteRemoteBranch {
		args = append(args, "--delete-branch")
	}

	_, err := c.Run(args...)
	return err
}

// GetPRReviews returns reviews on a PR
func (c *Client) GetPRReviews(repo string, number int) ([]PRReview, error) {
	var reviews []PRReview
	err := c.RunJSON(&reviews,
		"pr", "view", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "reviews",
	)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

// GetPRComments returns review comments on a PR
func (c *Client) GetPRComments(repo string, number int) ([]PRComment, error) {
	var result struct {
		Comments []PRComment `json:"comments"`
	}
	err := c.RunJSON(&result,
		"pr", "view", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "comments",
	)
	if err != nil {
		return nil, err
	}
	return result.Comments, nil
}

// GetMergedPRs returns recently merged PRs
func (c *Client) GetMergedPRs(repo string) ([]PullRequest, error) {
	var prs []PullRequest
	err := c.RunJSON(&prs,
		"pr", "list",
		"--repo", repo,
		"--state", "merged",
		"--json", "number,title,headRefName,merged",
		"--limit", "50",
	)
	return prs, err
}

// ListPRs returns all open PRs
func (c *Client) ListPRs(repo string) ([]PullRequest, error) {
	var prs []PullRequest
	err := c.RunJSON(&prs,
		"pr", "list",
		"--repo", repo,
		"--state", "open",
		"--json", "number,title,body,state,url,headRefName,baseRefName,createdAt",
	)
	return prs, err
}

// AddPRComment adds a comment to a PR
func (c *Client) AddPRComment(repo string, number int, body string) error {
	_, err := c.Run(
		"pr", "comment", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--body", body,
	)
	return err
}
