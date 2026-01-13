package github

import (
	"fmt"
)

// GetPRChecks returns CI check status for a PR
func (c *Client) GetPRChecks(repo string, number int) ([]PRCheck, error) {
	var checks []PRCheck
	err := c.RunJSON(&checks,
		"pr", "checks", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "name,state,conclusion",
	)
	if err != nil {
		// If there are no checks, gh returns an error
		return nil, nil
	}
	return checks, err
}

// PRChecksPassing returns true if all checks have passed
func (c *Client) PRChecksPassing(repo string, number int) (bool, error) {
	checks, err := c.GetPRChecks(repo, number)
	if err != nil {
		return false, err
	}

	if len(checks) == 0 {
		return true, nil // No checks configured
	}

	for _, check := range checks {
		if check.Status != "completed" {
			return false, nil // Still running
		}
		if check.Conclusion != "success" && check.Conclusion != "skipped" && check.Conclusion != "neutral" {
			return false, nil // Failed
		}
	}

	return true, nil
}

// PRChecksRunning returns true if any checks are still running
func (c *Client) PRChecksRunning(repo string, number int) (bool, error) {
	checks, err := c.GetPRChecks(repo, number)
	if err != nil {
		return false, err
	}

	for _, check := range checks {
		if check.Status != "completed" {
			return true, nil
		}
	}

	return false, nil
}

// PRChecksFailed returns true if any checks have failed
func (c *Client) PRChecksFailed(repo string, number int) (bool, error) {
	checks, err := c.GetPRChecks(repo, number)
	if err != nil {
		return false, err
	}

	for _, check := range checks {
		if check.Status == "completed" && check.Conclusion == "failure" {
			return true, nil
		}
	}

	return false, nil
}

// GetLatestWorkflowRun returns the latest workflow run for a branch
func (c *Client) GetLatestWorkflowRun(repo, branch string) (*WorkflowRun, error) {
	var runs []WorkflowRun
	err := c.RunJSON(&runs,
		"run", "list",
		"--repo", repo,
		"--branch", branch,
		"--json", "databaseId,name,status,conclusion,url,createdAt",
		"--limit", "1",
	)
	if err != nil {
		return nil, err
	}

	if len(runs) == 0 {
		return nil, nil
	}
	return &runs[0], nil
}

// GetWorkflowRunLogs returns the logs for a failed workflow run
func (c *Client) GetWorkflowRunLogs(repo string, runID int) (string, error) {
	output, err := c.Run(
		"run", "view", fmt.Sprintf("%d", runID),
		"--repo", repo,
		"--log-failed",
	)
	return output, err
}
