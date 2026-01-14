package errors

import (
	"errors"
	"testing"
)

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "settings.poll_interval",
		Message: "must be at least 1",
	}

	expected := "config error: settings.poll_interval: must be at least 1"
	if err.Error() != expected {
		t.Errorf("ConfigError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestGitHubError(t *testing.T) {
	innerErr := errors.New("connection refused")
	err := &GitHubError{
		Operation: "list issues",
		Repo:      "owner/repo",
		Err:       innerErr,
	}

	expected := "github error: list issues on owner/repo: connection refused"
	if err.Error() != expected {
		t.Errorf("GitHubError.Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap
	if err.Unwrap() != innerErr {
		t.Error("GitHubError.Unwrap() should return the wrapped error")
	}

	// Test errors.Is
	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should find the wrapped error")
	}
}

func TestSessionError(t *testing.T) {
	innerErr := errors.New("process killed")
	err := &SessionError{
		SessionID: "owner/repo#42",
		Operation: "spawn",
		Err:       innerErr,
	}

	expected := "session error: spawn for owner/repo#42: process killed"
	if err.Error() != expected {
		t.Errorf("SessionError.Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap
	if err.Unwrap() != innerErr {
		t.Error("SessionError.Unwrap() should return the wrapped error")
	}
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are properly defined
	sentinelErrors := []struct {
		err  error
		name string
	}{
		{ErrConfigNotFound, "ErrConfigNotFound"},
		{ErrConfigInvalid, "ErrConfigInvalid"},
		{ErrAlreadyRunning, "ErrAlreadyRunning"},
		{ErrNotRunning, "ErrNotRunning"},
		{ErrGHNotInstalled, "ErrGHNotInstalled"},
		{ErrGHNotAuthenticated, "ErrGHNotAuthenticated"},
		{ErrClaudeNotInstalled, "ErrClaudeNotInstalled"},
		{ErrRepoNotFound, "ErrRepoNotFound"},
		{ErrRepoNotAccessible, "ErrRepoNotAccessible"},
		{ErrPathNotFound, "ErrPathNotFound"},
		{ErrNotGitRepo, "ErrNotGitRepo"},
		{ErrWorktreeExists, "ErrWorktreeExists"},
		{ErrWorktreeNotFound, "ErrWorktreeNotFound"},
		{ErrSessionExists, "ErrSessionExists"},
		{ErrMaxSessionsReached, "ErrMaxSessionsReached"},
	}

	for _, tt := range sentinelErrors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s.Error() is empty", tt.name)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that we can wrap and unwrap errors properly
	baseErr := errors.New("base error")

	ghErr := &GitHubError{
		Operation: "test",
		Repo:      "test/repo",
		Err:       baseErr,
	}

	// Should be able to find base error with errors.Is
	if !errors.Is(ghErr, baseErr) {
		t.Error("errors.Is should find wrapped error")
	}

	// Should be able to unwrap to get base error
	var unwrapped error = ghErr
	for {
		if u, ok := unwrapped.(interface{ Unwrap() error }); ok {
			unwrapped = u.Unwrap()
		} else {
			break
		}
	}
	if unwrapped != baseErr {
		t.Error("Should unwrap to base error")
	}
}
