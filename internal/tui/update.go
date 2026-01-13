package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nathanbarrett/dev-swarm-go/internal/orchestrator"
	"github.com/nathanbarrett/dev-swarm-go/internal/session"
)

// StateUpdateMsg wraps an orchestrator state update
type StateUpdateMsg orchestrator.StateUpdate

// TickMsg triggers a refresh
type TickMsg struct{}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForUpdates(),
		m.refreshData(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case StateUpdateMsg:
		return m.handleStateUpdate(msg)

	case TickMsg:
		return m, m.refreshData()
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle help toggle first
	if m.showHelp {
		if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Escape) || key.Matches(msg, m.keys.Quit) {
			m.showHelp = false
			return m, nil
		}
		return m, nil
	}

	// Handle focused session mode
	if m.focusedSession != "" {
		if key.Matches(msg, m.keys.Escape) {
			m.focusedSession = ""
			return m, nil
		}
		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}
		// Scroll output
		if key.Matches(msg, m.keys.Up) {
			if m.outputScroll > 0 {
				m.outputScroll--
			}
			return m, nil
		}
		if key.Matches(msg, m.keys.Down) {
			m.outputScroll++
			return m, nil
		}
		return m, nil
	}

	// Normal mode
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.selectedIdx < m.TotalItems()-1 {
			m.selectedIdx++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		// Focus on selected session
		sess := m.GetSelectedSession()
		if sess != nil {
			m.focusedSession = sess.ID
			m.outputScroll = 0
		}
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		m.orchestrator.ForceRefresh()
		return m, m.refreshData()

	case key.Matches(msg, m.keys.Pause):
		if m.isPaused {
			m.orchestrator.Resume()
			m.isPaused = false
		} else {
			m.orchestrator.Pause()
			m.isPaused = true
		}
		return m, nil

	case key.Matches(msg, m.keys.ToggleLog):
		m.showLogs = !m.showLogs
		return m, nil

	case key.Matches(msg, m.keys.Help):
		m.showHelp = true
		return m, nil

	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleStateUpdate handles orchestrator state updates
func (m Model) handleStateUpdate(msg StateUpdateMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case orchestrator.UpdatePollComplete:
		// Refresh display data
		return m, m.refreshData()

	case orchestrator.UpdateSessionStarted:
		if info, ok := msg.Data.(session.SessionInfo); ok {
			m.sessions[info.ID] = info
		}
		return m, m.listenForUpdates()

	case orchestrator.UpdateSessionEnded:
		if event, ok := msg.Data.(session.StatusEvent); ok {
			delete(m.sessions, event.SessionID)
			if m.focusedSession == event.SessionID {
				m.focusedSession = ""
			}
		}
		return m, m.listenForUpdates()

	case orchestrator.UpdateSessionOutput:
		// Output is captured in the session, just trigger a repaint
		return m, m.listenForUpdates()

	default:
		return m, m.listenForUpdates()
	}
}

// listenForUpdates returns a command that listens for orchestrator updates
func (m Model) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		update, ok := <-m.updateChan
		if !ok {
			return nil
		}
		return StateUpdateMsg(update)
	}
}

// refreshData refreshes the display data from the orchestrator
func (m Model) refreshData() tea.Cmd {
	return func() tea.Msg {
		m.codebases = m.orchestrator.GetCodebaseInfo()
		stats := m.orchestrator.Stats()
		m.lastPoll = stats.LastPoll
		m.nextPoll = stats.NextPoll
		m.isPaused = stats.IsPaused

		// Update session info
		for _, info := range m.orchestrator.GetSessionManager().GetSessionInfo() {
			m.sessions[info.ID] = info
		}

		return TickMsg{}
	}
}
