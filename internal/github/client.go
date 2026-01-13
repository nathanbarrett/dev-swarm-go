package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Client wraps the gh CLI for GitHub operations
type Client struct{}

// NewClient creates a new GitHub client
func NewClient() *Client {
	return &Client{}
}

// Run executes a gh command and returns stdout
func (c *Client) Run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh command failed: %s (exit code %d)", stderr.String(), exitErr.ExitCode())
		}
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunJSON executes a gh command and parses JSON output
func (c *Client) RunJSON(result interface{}, args ...string) error {
	output, err := c.Run(args...)
	if err != nil {
		return err
	}
	if output == "" {
		return nil
	}
	return json.Unmarshal([]byte(output), result)
}

// IsInstalled checks if gh CLI is installed
func (c *Client) IsInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// IsAuthenticated checks if gh CLI is authenticated
func (c *Client) IsAuthenticated() bool {
	_, err := c.Run("auth", "status")
	return err == nil
}

// GetAuthenticatedUser returns the currently authenticated username
func (c *Client) GetAuthenticatedUser() (string, error) {
	output, err := c.Run("api", "user", "-q", ".login")
	if err != nil {
		return "", err
	}
	return output, nil
}

// RepoExists checks if a repository exists and is accessible
func (c *Client) RepoExists(repo string) bool {
	_, err := c.Run("repo", "view", repo, "--json", "name")
	return err == nil
}
