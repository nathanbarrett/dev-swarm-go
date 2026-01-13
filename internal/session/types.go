package session

import (
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

// Status represents the session status
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusCompleted
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// OutputLine represents a single line of output
type OutputLine struct {
	Timestamp time.Time
	Text      string
	Stream    string // "stdout" or "stderr"
}

// OutputEvent is sent when new output is available
type OutputEvent struct {
	SessionID string
	Line      OutputLine
}

// StatusEvent is sent when session status changes
type StatusEvent struct {
	SessionID string
	Status    Status
	ExitCode  *int
	Error     error
}

// SessionInfo contains information about a session for display
type SessionInfo struct {
	ID           string
	IssueNumber  int
	IssueTitle   string
	Repo         string
	CodebaseName string
	Status       Status
	Label        string
	StartedAt    time.Time
	CompletedAt  *time.Time
	Duration     time.Duration
	ExitCode     *int
	Error        error
}

// SpawnRequest contains all information needed to spawn a session
type SpawnRequest struct {
	Issue        *github.Issue
	Codebase     *config.Codebase
	CurrentLabel string
	AIAction     string
}
