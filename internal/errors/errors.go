package errors

import (
	"errors"
	"fmt"
)

var (
	ErrConfigNotFound     = errors.New("config file not found")
	ErrConfigInvalid      = errors.New("config file is invalid")
	ErrAlreadyRunning     = errors.New("dev-swarm is already running")
	ErrNotRunning         = errors.New("dev-swarm is not running")
	ErrGHNotInstalled     = errors.New("gh CLI is not installed")
	ErrGHNotAuthenticated = errors.New("gh CLI is not authenticated")
	ErrClaudeNotInstalled = errors.New("claude CLI is not installed")
	ErrRepoNotFound       = errors.New("repository not found")
	ErrRepoNotAccessible  = errors.New("repository is not accessible")
	ErrPathNotFound       = errors.New("local path not found")
	ErrNotGitRepo         = errors.New("path is not a git repository")
	ErrWorktreeExists     = errors.New("worktree already exists")
	ErrWorktreeNotFound   = errors.New("worktree not found")
	ErrSessionExists      = errors.New("session already exists for this issue")
	ErrMaxSessionsReached = errors.New("maximum concurrent sessions reached")
)

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s: %s", e.Field, e.Message)
}

// GitHubError represents a GitHub API error
type GitHubError struct {
	Operation string
	Repo      string
	Err       error
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("github error: %s on %s: %v", e.Operation, e.Repo, e.Err)
}

func (e *GitHubError) Unwrap() error {
	return e.Err
}

// SessionError represents a session management error
type SessionError struct {
	SessionID string
	Operation string
	Err       error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session error: %s for %s: %v", e.Operation, e.SessionID, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}
