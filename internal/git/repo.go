package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsGitRepo checks if a path is a git repository
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetRepoRoot returns the root directory of a git repository
func GetRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDefaultBranch attempts to determine the default branch (main or master)
func GetDefaultBranch(path string) (string, error) {
	// Try to get the default branch from origin
	cmd := exec.Command("git", "-C", path, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		// Remove "origin/" prefix
		if strings.HasPrefix(branch, "origin/") {
			return strings.TrimPrefix(branch, "origin/"), nil
		}
		return branch, nil
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "-C", path, "rev-parse", "--verify", branch)
		if cmd.Run() == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

// Fetch fetches from remote
func Fetch(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "origin")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %s", stderr.String())
	}
	return nil
}

// FetchBranch fetches a specific branch from remote
func FetchBranch(path, branch string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "origin", branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %s", stderr.String())
	}
	return nil
}

// BranchExists checks if a branch exists locally
func BranchExists(path, branch string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--verify", branch)
	return cmd.Run() == nil
}

// RemoteBranchExists checks if a branch exists on remote
func RemoteBranchExists(path, branch string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
	return cmd.Run() == nil
}

// GetRemoteURL returns the remote URL for origin
func GetRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Pull pulls changes from remote
func Pull(path string) error {
	cmd := exec.Command("git", "-C", path, "pull", "--ff-only")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %s", stderr.String())
	}
	return nil
}

// Checkout checks out a branch
func Checkout(path, branch string) error {
	cmd := exec.Command("git", "-C", path, "checkout", branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %s", stderr.String())
	}
	return nil
}

// CreateBranch creates a new branch from a base
func CreateBranch(path, name, base string) error {
	cmd := exec.Command("git", "-C", path, "checkout", "-b", name, base)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git branch creation failed: %s", stderr.String())
	}
	return nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(path, branch string) error {
	cmd := exec.Command("git", "-C", path, "branch", "-D", branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git branch delete failed: %s", stderr.String())
	}
	return nil
}

// Push pushes changes to remote
func Push(path, branch string) error {
	cmd := exec.Command("git", "-C", path, "push", "-u", "origin", branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %s", stderr.String())
	}
	return nil
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
