package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/git"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
	"github.com/nathanbarrett/dev-swarm-go/internal/session"
)

// Orchestrator manages the dev-swarm workflow
type Orchestrator struct {
	config         *config.Config
	ghClient       *github.Client
	sessionManager *session.Manager

	// State
	mu         sync.RWMutex
	codebases  map[string]*CodebaseState
	startedAt  time.Time
	lastPoll   time.Time
	isPaused   bool
	isRunning  bool

	// Control
	ctx    context.Context
	cancel context.CancelFunc

	// Communication
	stateChan chan StateUpdate

	// Logging
	logger *log.Logger
}

// New creates a new Orchestrator
func New(cfg *config.Config, logger *log.Logger) (*Orchestrator, error) {
	ctx, cancel := context.WithCancel(context.Background())

	worktreesDir := config.WorktreesDir()

	return &Orchestrator{
		config:         cfg,
		ghClient:       github.NewClient(),
		sessionManager: session.NewManager(cfg.Settings.MaxConcurrentSessions, cfg.Settings.OutputBufferLines, worktreesDir),
		codebases:      make(map[string]*CodebaseState),
		ctx:            ctx,
		cancel:         cancel,
		stateChan:      make(chan StateUpdate, 100),
		logger:         logger,
	}, nil
}

// Start begins the orchestration loop
func (o *Orchestrator) Start() error {
	o.mu.Lock()
	if o.isRunning {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator is already running")
	}
	o.isRunning = true
	o.startedAt = time.Now()
	o.mu.Unlock()

	// Initialize codebase states
	for _, cb := range o.config.GetEnabledCodebases() {
		o.codebases[cb.Name] = &CodebaseState{
			Config:    &cb,
			Issues:    make(map[int]*IssueState),
			IsHealthy: true,
		}
	}

	// Sync labels
	if err := o.syncLabels(); err != nil {
		o.log("Warning: failed to sync labels: %v", err)
	}

	// Ensure worktrees directory exists
	if err := config.EnsureWorktreesDir(); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Start main loop
	go o.mainLoop()

	// Start session event handler
	go o.handleSessionEvents()

	o.log("Orchestrator started. Monitoring %d codebases.", len(o.config.GetEnabledCodebases()))
	return nil
}

// Stop gracefully stops the orchestrator
func (o *Orchestrator) Stop() {
	o.mu.Lock()
	if !o.isRunning {
		o.mu.Unlock()
		return
	}
	o.isRunning = false
	o.mu.Unlock()

	o.log("Stopping orchestrator...")
	o.cancel()
	o.sessionManager.StopAll()
	close(o.stateChan)
	o.log("Orchestrator stopped.")
}

// Pause pauses polling
func (o *Orchestrator) Pause() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.isPaused = true
	o.log("Polling paused.")
}

// Resume resumes polling
func (o *Orchestrator) Resume() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.isPaused = false
	o.log("Polling resumed.")
}

// IsPaused returns true if polling is paused
func (o *Orchestrator) IsPaused() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.isPaused
}

// ForceRefresh triggers an immediate poll
func (o *Orchestrator) ForceRefresh() {
	o.poll()
}

// StateChan returns the state update channel
func (o *Orchestrator) StateChan() <-chan StateUpdate {
	return o.stateChan
}

// Stats returns current orchestrator statistics
func (o *Orchestrator) Stats() Stats {
	o.mu.RLock()
	defer o.mu.RUnlock()

	activeSessions := o.sessionManager.ActiveCount()
	pollInterval := o.getPollInterval()

	return Stats{
		ActiveSessions: activeSessions,
		TotalIssues:    o.countTotalIssues(),
		LastPoll:       o.lastPoll,
		NextPoll:       o.lastPoll.Add(pollInterval),
		IsPaused:       o.isPaused,
		Uptime:         time.Since(o.startedAt),
	}
}

// GetCodebaseInfo returns display information for all codebases
func (o *Orchestrator) GetCodebaseInfo() []CodebaseInfo {
	o.mu.RLock()
	defer o.mu.RUnlock()

	result := make([]CodebaseInfo, 0, len(o.codebases))
	for _, cb := range o.codebases {
		info := CodebaseInfo{
			Name:      cb.Config.Name,
			Repo:      cb.Config.Repo,
			Issues:    make([]IssueInfo, 0, len(cb.Issues)),
			IsIdle:    len(cb.Issues) == 0,
			IsHealthy: cb.IsHealthy,
		}
		if cb.Error != nil {
			info.Error = cb.Error.Error()
		}

		for _, issue := range cb.Issues {
			issueInfo := IssueInfo{
				Number:       issue.Issue.Number,
				Title:        issue.Issue.Title,
				Label:        issue.Label,
				HasSession:   issue.HasSession,
				SessionID:    issue.SessionID,
				CodebaseName: cb.Config.Name,
				Repo:         cb.Config.Repo,
			}

			if issue.HasSession {
				sess := o.sessionManager.GetSession(issue.SessionID)
				if sess != nil {
					issueInfo.Status = sess.Status
					issueInfo.Duration = sess.Duration()
				}
			}

			info.Issues = append(info.Issues, issueInfo)
		}

		result = append(result, info)
	}

	return result
}

// GetSessionManager returns the session manager
func (o *Orchestrator) GetSessionManager() *session.Manager {
	return o.sessionManager
}

// syncLabels ensures all required labels exist in all repos
func (o *Orchestrator) syncLabels() error {
	labels := o.config.Labels.GetAllLabels()
	labelInfos := make([]github.LabelInfo, 0, len(labels))
	for _, l := range labels {
		labelInfos = append(labelInfos, github.LabelInfo{
			Name:        l.Name,
			Color:       l.Color,
			Description: l.Description,
		})
	}

	for _, cb := range o.config.GetEnabledCodebases() {
		o.log("Syncing labels for %s...", cb.Repo)
		if err := o.ghClient.SyncLabels(cb.Repo, labelInfos); err != nil {
			o.log("Warning: failed to sync labels for %s: %v", cb.Repo, err)
		}
	}

	return nil
}

// handleSessionEvents processes session output and status events
func (o *Orchestrator) handleSessionEvents() {
	outputChan := o.sessionManager.OutputChan()
	statusChan := o.sessionManager.StatusChan()

	for {
		select {
		case <-o.ctx.Done():
			return

		case event := <-outputChan:
			o.sendUpdate(StateUpdate{
				Type:      UpdateSessionOutput,
				Data:      event,
				Timestamp: time.Now(),
			})

		case event := <-statusChan:
			o.sendUpdate(StateUpdate{
				Type:      UpdateSessionEnded,
				Data:      event,
				Timestamp: time.Now(),
			})

			// Update issue state
			o.mu.Lock()
			for _, cb := range o.codebases {
				for _, issue := range cb.Issues {
					if issue.SessionID == event.SessionID {
						issue.HasSession = false
						issue.SessionID = ""
						break
					}
				}
			}
			o.mu.Unlock()
		}
	}
}

// Helper functions

func (o *Orchestrator) log(format string, args ...interface{}) {
	if o.logger != nil {
		o.logger.Printf(format, args...)
	}
}

func (o *Orchestrator) sendUpdate(update StateUpdate) {
	select {
	case o.stateChan <- update:
	default:
		// Channel full, drop update
	}
}

func (o *Orchestrator) getPollInterval() time.Duration {
	if o.sessionManager.ActiveCount() > 0 {
		return time.Duration(o.config.Settings.ActivePollInterval) * time.Second
	}
	return time.Duration(o.config.Settings.PollInterval) * time.Second
}

func (o *Orchestrator) countTotalIssues() int {
	count := 0
	for _, cb := range o.codebases {
		count += len(cb.Issues)
	}
	return count
}

func (o *Orchestrator) getSessionID(issue *github.Issue) string {
	// Find codebase for this issue
	for _, cb := range o.codebases {
		for _, i := range cb.Issues {
			if i.Issue.Number == issue.Number {
				return fmt.Sprintf("%s#%d", cb.Config.Repo, issue.Number)
			}
		}
	}
	return ""
}

func (o *Orchestrator) getBranchName(issueNumber int) string {
	return git.GetBranchName(issueNumber)
}
