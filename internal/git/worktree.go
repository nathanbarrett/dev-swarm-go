package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CreateWorktree creates a new git worktree
func CreateWorktree(repoPath, worktreePath, branchName, baseBranch string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	// Fetch latest from remote
	if err := FetchBranch(repoPath, baseBranch); err != nil {
		// Non-fatal, continue anyway
	}

	var cmd *exec.Cmd
	var stderr bytes.Buffer

	// Check if branch already exists locally
	if BranchExists(repoPath, branchName) {
		// Branch exists, create worktree for existing branch
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", worktreePath, branchName)
	} else if RemoteBranchExists(repoPath, branchName) {
		// Remote branch exists, create worktree tracking it
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", worktreePath, branchName)
	} else {
		// Create new branch from base
		cmd = exec.Command("git", "-C", repoPath, "worktree", "add", "-b", branchName, worktreePath, fmt.Sprintf("origin/%s", baseBranch))
	}

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %s: %w", stderr.String(), err)
	}

	return nil
}

// RemoveWorktree removes a git worktree and optionally the branch
func RemoveWorktree(repoPath, worktreePath string, deleteBranch bool) error {
	// Get branch name before removing worktree
	var branchName string
	if deleteBranch {
		branch, err := GetCurrentBranch(worktreePath)
		if err == nil && branch != "HEAD" {
			branchName = branch
		}
	}

	// Remove the worktree
	cmd := exec.Command("git", "-C", repoPath, "worktree", "remove", worktreePath, "--force")
	if err := cmd.Run(); err != nil {
		// Try removing the directory directly if worktree remove fails
		os.RemoveAll(worktreePath)
	}

	// Prune worktree references
	pruneCmd := exec.Command("git", "-C", repoPath, "worktree", "prune")
	pruneCmd.Run() // Ignore errors

	// Delete the branch if requested
	if deleteBranch && branchName != "" && branchName != "HEAD" {
		DeleteBranch(repoPath, branchName) // Ignore errors - branch might not exist locally
	}

	return nil
}

// WorktreeExists checks if a worktree exists
func WorktreeExists(worktreePath string) bool {
	info, err := os.Stat(worktreePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ListWorktrees returns all worktrees for a repo
func ListWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var current Worktree
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch ")
			// Remove refs/heads/ prefix
			if strings.HasPrefix(current.Branch, "refs/heads/") {
				current.Branch = strings.TrimPrefix(current.Branch, "refs/heads/")
			}
		}
	}

	// Don't forget the last worktree
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// CleanupOrphanedWorktrees removes worktrees that are no longer needed
func CleanupOrphanedWorktrees(repoPath, worktreesDir string) error {
	worktrees, err := ListWorktrees(repoPath)
	if err != nil {
		return err
	}

	for _, wt := range worktrees {
		// Check if worktree is in our managed directory
		if strings.HasPrefix(wt.Path, worktreesDir) {
			// Check if the worktree directory still exists and is valid
			if !WorktreeExists(wt.Path) {
				// Prune this worktree reference
				RemoveWorktree(repoPath, wt.Path, false)
			}
		}
	}

	return nil
}

// GetWorktreePath returns the worktree path for an issue
func GetWorktreePath(worktreesDir, codebaseName string, issueNumber int) string {
	return filepath.Join(worktreesDir, codebaseName, fmt.Sprintf("issue-%d", issueNumber))
}

// GetBranchName returns the branch name for an issue
func GetBranchName(issueNumber int) string {
	return fmt.Sprintf("claude/issue-%d", issueNumber)
}
