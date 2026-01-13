package tui

import (
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/orchestrator"
	"github.com/nathanbarrett/dev-swarm-go/internal/session"
	"github.com/nathanbarrett/dev-swarm-go/pkg/version"
)

// Model represents the TUI state
type Model struct {
	// Orchestrator reference
	orchestrator *orchestrator.Orchestrator

	// Display data
	codebases []orchestrator.CodebaseInfo
	sessions  map[string]session.SessionInfo

	// UI state
	selectedIdx  int
	scrollOffset int
	outputScroll int
	focusedSession string // Session ID of focused session for full-screen output

	// Status
	lastPoll  time.Time
	nextPoll  time.Time
	isPaused  bool
	showHelp  bool
	showLogs  bool

	// Dimensions
	width  int
	height int

	// Key bindings
	keys KeyMap

	// Update channel
	updateChan <-chan orchestrator.StateUpdate

	// Quit flag
	quitting bool
}

// NewModel creates a new TUI model
func NewModel(orch *orchestrator.Orchestrator) Model {
	return Model{
		orchestrator: orch,
		sessions:     make(map[string]session.SessionInfo),
		keys:         DefaultKeyMap(),
		updateChan:   orch.StateChan(),
	}
}

// SelectionType represents what kind of item is selected
type SelectionType int

const (
	SelectionCodebase SelectionType = iota
	SelectionIssue
)

// SelectedItem returns information about the currently selected item
func (m *Model) SelectedItem() (codebaseIdx int, issueIdx int, selType SelectionType) {
	if len(m.codebases) == 0 {
		return -1, -1, SelectionCodebase
	}

	idx := m.selectedIdx
	for i, cb := range m.codebases {
		// Codebase header
		if idx == 0 {
			return i, -1, SelectionCodebase
		}
		idx--

		// Issues under this codebase
		for j := range cb.Issues {
			if idx == 0 {
				return i, j, SelectionIssue
			}
			idx--
		}
	}

	return len(m.codebases) - 1, -1, SelectionCodebase
}

// TotalItems returns the total number of selectable items
func (m *Model) TotalItems() int {
	count := 0
	for _, cb := range m.codebases {
		count++ // Codebase header
		count += len(cb.Issues)
	}
	return count
}

// GetSelectedSession returns the session for the selected issue, if any
func (m *Model) GetSelectedSession() *session.SessionInfo {
	cbIdx, issueIdx, selType := m.SelectedItem()
	if selType != SelectionIssue || cbIdx < 0 || issueIdx < 0 {
		return nil
	}

	if cbIdx >= len(m.codebases) {
		return nil
	}
	cb := m.codebases[cbIdx]

	if issueIdx >= len(cb.Issues) {
		return nil
	}
	issue := cb.Issues[issueIdx]

	if !issue.HasSession {
		return nil
	}

	info, exists := m.sessions[issue.SessionID]
	if !exists {
		return nil
	}

	return &info
}

// GetFocusedSession returns the session that is currently focused for full output view
func (m *Model) GetFocusedSession() *session.Session {
	if m.focusedSession == "" {
		return nil
	}
	return m.orchestrator.GetSessionManager().GetSession(m.focusedSession)
}

// Stats returns the current stats
func (m *Model) Stats() orchestrator.Stats {
	return m.orchestrator.Stats()
}

// Version returns the version string
func (m *Model) Version() string {
	return version.Info()
}
