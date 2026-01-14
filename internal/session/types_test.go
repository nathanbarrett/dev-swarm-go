package session

import (
	"testing"
	"time"
)

func TestStatusConstants(t *testing.T) {
	// Verify status constants have expected values
	if StatusPending != 0 {
		t.Errorf("StatusPending = %d, want 0", StatusPending)
	}
	if StatusRunning != 1 {
		t.Errorf("StatusRunning = %d, want 1", StatusRunning)
	}
	if StatusCompleted != 2 {
		t.Errorf("StatusCompleted = %d, want 2", StatusCompleted)
	}
	if StatusFailed != 3 {
		t.Errorf("StatusFailed = %d, want 3", StatusFailed)
	}
}

func TestOutputEventFields(t *testing.T) {
	event := OutputEvent{
		SessionID: "owner/repo#42",
		Line: OutputLine{
			Timestamp: time.Now(),
			Text:      "test output",
			Stream:    "stdout",
		},
	}

	if event.SessionID != "owner/repo#42" {
		t.Errorf("SessionID = %q, want %q", event.SessionID, "owner/repo#42")
	}
	if event.Line.Text != "test output" {
		t.Errorf("Line.Text = %q, want %q", event.Line.Text, "test output")
	}
}

func TestStatusEventFields(t *testing.T) {
	exitCode := 0
	event := StatusEvent{
		SessionID: "owner/repo#42",
		Status:    StatusCompleted,
		ExitCode:  &exitCode,
		Error:     nil,
	}

	if event.SessionID != "owner/repo#42" {
		t.Errorf("SessionID = %q, want %q", event.SessionID, "owner/repo#42")
	}
	if event.Status != StatusCompleted {
		t.Errorf("Status = %d, want %d", event.Status, StatusCompleted)
	}
	if *event.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", *event.ExitCode)
	}
}

func TestSessionInfoFields(t *testing.T) {
	now := time.Now()
	duration := 5 * time.Minute
	exitCode := 0

	info := SessionInfo{
		ID:           "owner/repo#42",
		IssueNumber:  42,
		IssueTitle:   "Test Issue",
		Repo:         "owner/repo",
		CodebaseName: "test-repo",
		Status:       StatusCompleted,
		Label:        "user:ready-to-plan",
		StartedAt:    now,
		CompletedAt:  &now,
		Duration:     duration,
		ExitCode:     &exitCode,
		Error:        nil,
	}

	if info.ID != "owner/repo#42" {
		t.Errorf("ID = %q, want %q", info.ID, "owner/repo#42")
	}
	if info.IssueNumber != 42 {
		t.Errorf("IssueNumber = %d, want 42", info.IssueNumber)
	}
	if info.Status != StatusCompleted {
		t.Errorf("Status = %d, want %d", info.Status, StatusCompleted)
	}
	if info.Duration != duration {
		t.Errorf("Duration = %v, want %v", info.Duration, duration)
	}
}

func TestSpawnRequestFields(t *testing.T) {
	req := SpawnRequest{
		Issue:        nil, // Would normally be a *github.Issue
		Codebase:     nil, // Would normally be a *config.Codebase
		CurrentLabel: "user:ready-to-plan",
		AIAction:     "Create implementation plan",
	}

	if req.CurrentLabel != "user:ready-to-plan" {
		t.Errorf("CurrentLabel = %q, want %q", req.CurrentLabel, "user:ready-to-plan")
	}
	if req.AIAction != "Create implementation plan" {
		t.Errorf("AIAction = %q, want %q", req.AIAction, "Create implementation plan")
	}
}
