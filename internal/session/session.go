package session

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

// Session represents an active Claude Code session
type Session struct {
	ID           string
	Issue        *github.Issue
	Codebase     *config.Codebase
	WorktreePath string
	BranchName   string
	Label        string

	// Process
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	// Output
	output *OutputBuffer

	// Status
	Status      Status
	StartedAt   time.Time
	CompletedAt *time.Time
	ExitCode    *int
	Error       error

	// Control
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewSession creates a new session
func NewSession(
	id string,
	issue *github.Issue,
	codebase *config.Codebase,
	worktreePath string,
	branchName string,
	label string,
	cmd *exec.Cmd,
	bufferSize int,
) *Session {
	return &Session{
		ID:           id,
		Issue:        issue,
		Codebase:     codebase,
		WorktreePath: worktreePath,
		BranchName:   branchName,
		Label:        label,
		cmd:          cmd,
		output:       NewOutputBuffer(bufferSize),
		Status:       StatusPending,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the session and begins capturing output
func (s *Session) Start(outputChan chan<- OutputEvent, statusChan chan<- StatusEvent) error {
	// Set up pipes
	var err error
	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		return err
	}

	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}

	// Update status
	s.mu.Lock()
	s.Status = StatusRunning
	s.StartedAt = time.Now()
	s.mu.Unlock()

	// Start the process
	if err := s.cmd.Start(); err != nil {
		s.mu.Lock()
		s.Status = StatusFailed
		s.Error = err
		s.mu.Unlock()
		return err
	}

	// Start output capture goroutines
	go s.captureOutput(s.stdout, "stdout", outputChan)
	go s.captureOutput(s.stderr, "stderr", outputChan)

	// Wait for completion
	go s.waitForCompletion(statusChan)

	return nil
}

// captureOutput reads from a stream and sends lines to the output channel
func (s *Session) captureOutput(reader io.Reader, stream string, outputChan chan<- OutputEvent) {
	scanner := bufio.NewScanner(reader)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-s.stopChan:
			return
		default:
			line := OutputLine{
				Timestamp: time.Now(),
				Text:      scanner.Text(),
				Stream:    stream,
			}

			s.output.Append(line)

			select {
			case outputChan <- OutputEvent{SessionID: s.ID, Line: line}:
			default:
				// Channel full, drop message
			}
		}
	}
}

// waitForCompletion waits for the process to exit and updates status
func (s *Session) waitForCompletion(statusChan chan<- StatusEvent) {
	err := s.cmd.Wait()

	s.mu.Lock()
	now := time.Now()
	s.CompletedAt = &now

	if err != nil {
		s.Status = StatusFailed
		s.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			s.ExitCode = &code
		}
	} else {
		s.Status = StatusCompleted
		code := 0
		s.ExitCode = &code
	}
	s.mu.Unlock()

	select {
	case statusChan <- StatusEvent{
		SessionID: s.ID,
		Status:    s.Status,
		ExitCode:  s.ExitCode,
		Error:     s.Error,
	}:
	default:
		// Channel full
	}
}

// Stop terminates the session
func (s *Session) Stop() {
	close(s.stopChan)
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
}

// GetOutput returns all output lines
func (s *Session) GetOutput() []OutputLine {
	return s.output.GetAll()
}

// GetRecentOutput returns the last n output lines
func (s *Session) GetRecentOutput(n int) []OutputLine {
	return s.output.GetRecent(n)
}

// Duration returns how long the session has been running
func (s *Session) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.CompletedAt != nil {
		return s.CompletedAt.Sub(s.StartedAt)
	}
	return time.Since(s.StartedAt)
}

// Info returns session information for display
func (s *Session) Info() SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SessionInfo{
		ID:           s.ID,
		IssueNumber:  s.Issue.Number,
		IssueTitle:   s.Issue.Title,
		Repo:         s.Codebase.Repo,
		CodebaseName: s.Codebase.Name,
		Status:       s.Status,
		Label:        s.Label,
		StartedAt:    s.StartedAt,
		CompletedAt:  s.CompletedAt,
		Duration:     s.Duration(),
		ExitCode:     s.ExitCode,
		Error:        s.Error,
	}
}

// IsRunning returns true if the session is still running
func (s *Session) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status == StatusRunning
}

// IsComplete returns true if the session has finished (success or failure)
func (s *Session) IsComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status == StatusCompleted || s.Status == StatusFailed
}
