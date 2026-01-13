package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/nathanbarrett/dev-swarm-go/internal/git"
)

// Manager manages all active sessions
type Manager struct {
	sessions          map[string]*Session
	maxActive         int
	outputBufferLines int
	worktreesDir      string
	mu                sync.RWMutex

	// Channels for communication
	outputChan chan OutputEvent
	statusChan chan StatusEvent
}

// NewManager creates a new session manager
func NewManager(maxActive, outputBufferLines int, worktreesDir string) *Manager {
	return &Manager{
		sessions:          make(map[string]*Session),
		maxActive:         maxActive,
		outputBufferLines: outputBufferLines,
		worktreesDir:      worktreesDir,
		outputChan:        make(chan OutputEvent, 1000),
		statusChan:        make(chan StatusEvent, 100),
	}
}

// CanSpawn returns true if we can spawn a new session
func (m *Manager) CanSpawn() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeCount := 0
	for _, s := range m.sessions {
		if s.Status == StatusRunning {
			activeCount++
		}
	}
	return activeCount < m.maxActive
}

// HasSession returns true if a session exists for the given ID
func (m *Manager) HasSession(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.sessions[sessionID]
	return exists
}

// ActiveCount returns the number of running sessions
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, s := range m.sessions {
		if s.Status == StatusRunning {
			count++
		}
	}
	return count
}

// SpawnSession creates and starts a new Claude session
func (m *Manager) SpawnSession(req SpawnRequest, aiInstructions string) (*Session, error) {
	sessionID := fmt.Sprintf("%s#%d", req.Codebase.Repo, req.Issue.Number)
	branchName := git.GetBranchName(req.Issue.Number)

	// Create worktree path
	worktreePath := git.GetWorktreePath(m.worktreesDir, req.Codebase.Name, req.Issue.Number)

	// Create worktree if it doesn't exist
	if !git.WorktreeExists(worktreePath) {
		err := git.CreateWorktree(req.Codebase.LocalPath, worktreePath, branchName, req.Codebase.DefaultBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	// Build context for Claude
	context := BuildContext(req.Issue, req.Codebase, req.CurrentLabel, req.AIAction, aiInstructions)

	// Write context to prompt file
	promptFile := filepath.Join(worktreePath, ".dev-swarm-prompt.md")
	if err := os.WriteFile(promptFile, []byte(context), 0644); err != nil {
		return nil, fmt.Errorf("failed to write prompt file: %w", err)
	}

	// Create Claude command
	cmd := exec.Command("claude",
		"--print",
		"--dangerously-skip-permissions",
		"--prompt-file", promptFile,
	)
	cmd.Dir = worktreePath
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DEV_SWARM_ISSUE=%d", req.Issue.Number),
		fmt.Sprintf("DEV_SWARM_REPO=%s", req.Codebase.Repo),
	)

	// Create session
	session := NewSession(
		sessionID,
		req.Issue,
		req.Codebase,
		worktreePath,
		branchName,
		req.CurrentLabel,
		cmd,
		m.outputBufferLines,
	)

	// Track session
	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	// Start session
	if err := session.Start(m.outputChan, m.statusChan); err != nil {
		m.mu.Lock()
		delete(m.sessions, sessionID)
		m.mu.Unlock()
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return session, nil
}

// GetSession returns a session by ID
func (m *Manager) GetSession(sessionID string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// GetAllSessions returns all sessions
func (m *Manager) GetAllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

// GetSessionForIssue returns a session for a specific issue
func (m *Manager) GetSessionForIssue(repo string, issueNumber int) *Session {
	sessionID := fmt.Sprintf("%s#%d", repo, issueNumber)
	return m.GetSession(sessionID)
}

// RemoveSession removes a session from tracking
func (m *Manager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// StopSession stops a specific session
func (m *Manager) StopSession(sessionID string) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		session.Stop()
	}
}

// StopAll stops all running sessions
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sessions {
		s.Stop()
	}
}

// CleanupCompleted removes completed sessions from tracking
func (m *Manager) CleanupCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, s := range m.sessions {
		if s.IsComplete() {
			delete(m.sessions, id)
		}
	}
}

// OutputChan returns the output event channel
func (m *Manager) OutputChan() <-chan OutputEvent {
	return m.outputChan
}

// StatusChan returns the status event channel
func (m *Manager) StatusChan() <-chan StatusEvent {
	return m.statusChan
}

// GetSessionInfo returns info for all sessions
func (m *Manager) GetSessionInfo() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SessionInfo, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s.Info())
	}
	return result
}
