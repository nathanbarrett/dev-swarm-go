package github

import (
	"fmt"
)

// ListIssuesWithLabel returns all open issues with a specific label
func (c *Client) ListIssuesWithLabel(repo, label string) ([]Issue, error) {
	var issues []Issue
	err := c.RunJSON(&issues,
		"issue", "list",
		"--repo", repo,
		"--label", label,
		"--state", "open",
		"--json", "number,title,body,state,url,labels,createdAt,updatedAt",
		"--limit", "100",
	)
	return issues, err
}

// ListIssuesWithLabels returns all open issues with any of the specified labels
func (c *Client) ListIssuesWithLabels(repo string, labels []string) ([]Issue, error) {
	var allIssues []Issue
	seen := make(map[int]bool)

	for _, label := range labels {
		issues, err := c.ListIssuesWithLabel(repo, label)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			if !seen[issue.Number] {
				seen[issue.Number] = true
				allIssues = append(allIssues, issue)
			}
		}
	}

	return allIssues, nil
}

// GetIssue returns full issue details including comments
func (c *Client) GetIssue(repo string, number int) (*Issue, error) {
	var issue Issue
	err := c.RunJSON(&issue,
		"issue", "view", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "number,title,body,state,url,labels,comments,createdAt,updatedAt",
	)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetIssueComments returns comments for an issue
func (c *Client) GetIssueComments(repo string, number int) ([]Comment, error) {
	var result struct {
		Comments []Comment `json:"comments"`
	}
	err := c.RunJSON(&result,
		"issue", "view", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "comments",
	)
	if err != nil {
		return nil, err
	}
	return result.Comments, nil
}

// UpdateIssueLabels changes labels on an issue
func (c *Client) UpdateIssueLabels(repo string, number int, removeLabels, addLabels []string) error {
	args := []string{"issue", "edit", fmt.Sprintf("%d", number), "--repo", repo}

	for _, label := range removeLabels {
		args = append(args, "--remove-label", label)
	}
	for _, label := range addLabels {
		args = append(args, "--add-label", label)
	}

	_, err := c.Run(args...)
	return err
}

// AddIssueComment adds a comment to an issue
func (c *Client) AddIssueComment(repo string, number int, body string) error {
	_, err := c.Run(
		"issue", "comment", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--body", body,
	)
	return err
}

// CloseIssue closes an issue
func (c *Client) CloseIssue(repo string, number int) error {
	_, err := c.Run(
		"issue", "close", fmt.Sprintf("%d", number),
		"--repo", repo,
	)
	return err
}
