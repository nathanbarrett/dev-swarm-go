package orchestrator

import (
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
	"github.com/nathanbarrett/dev-swarm-go/internal/session"
)

// StateUpdate represents a state change in the orchestrator
type StateUpdate struct {
	Type      UpdateType
	Codebase  string
	IssueNum  int
	Data      interface{}
	Timestamp time.Time
}

// UpdateType represents the type of state update
type UpdateType int

const (
	UpdateIssueFound UpdateType = iota
	UpdateIssueRemoved
	UpdateSessionStarted
	UpdateSessionOutput
	UpdateSessionEnded
	UpdateLabelChanged
	UpdatePollComplete
	UpdateError
)

func (t UpdateType) String() string {
	switch t {
	case UpdateIssueFound:
		return "issue_found"
	case UpdateIssueRemoved:
		return "issue_removed"
	case UpdateSessionStarted:
		return "session_started"
	case UpdateSessionOutput:
		return "session_output"
	case UpdateSessionEnded:
		return "session_ended"
	case UpdateLabelChanged:
		return "label_changed"
	case UpdatePollComplete:
		return "poll_complete"
	case UpdateError:
		return "error"
	default:
		return "unknown"
	}
}

// CodebaseState tracks the state of a single codebase
type CodebaseState struct {
	Config    *config.Codebase
	Issues    map[int]*IssueState
	LastPoll  time.Time
	IsHealthy bool
	Error     error
}

// IssueState tracks the state of a single issue
type IssueState struct {
	Issue       *github.Issue
	Label       string
	HasSession  bool
	SessionID   string
	LastChecked time.Time
}

// Stats contains orchestrator statistics
type Stats struct {
	ActiveSessions  int
	QueuedSessions  int
	WaitingSessions int
	TotalIssues     int
	LastPoll        time.Time
	NextPoll        time.Time
	IsPaused        bool
	Uptime          time.Duration
}

// IssueInfo contains display information about an issue
type IssueInfo struct {
	Number       int
	Title        string
	Label        string
	Status       session.Status
	HasSession   bool
	SessionID    string
	Duration     time.Duration
	CodebaseName string
	Repo         string
}

// CodebaseInfo contains display information about a codebase
type CodebaseInfo struct {
	Name     string
	Repo     string
	Issues   []IssueInfo
	IsIdle   bool
	IsHealthy bool
	Error    string
}
