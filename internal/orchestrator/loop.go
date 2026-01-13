package orchestrator

import (
	"fmt"
	"strings"
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/git"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
	"github.com/nathanbarrett/dev-swarm-go/internal/session"
)

// mainLoop is the main polling loop
func (o *Orchestrator) mainLoop() {
	// Do initial poll immediately
	o.poll()

	for {
		pollInterval := o.getPollInterval()

		select {
		case <-o.ctx.Done():
			return

		case <-time.After(pollInterval):
			o.mu.RLock()
			isPaused := o.isPaused
			o.mu.RUnlock()

			if !isPaused {
				o.poll()
			}
		}
	}
}

// poll performs a single polling cycle
func (o *Orchestrator) poll() {
	o.mu.Lock()
	o.lastPoll = time.Now()
	o.mu.Unlock()

	for _, codebase := range o.config.GetEnabledCodebases() {
		o.pollCodebase(&codebase)
	}

	// Check session status and cleanup
	o.checkSessionStatus()

	// Check CI status for active issues
	o.checkCIStatus()

	// Cleanup merged PRs
	o.cleanupMergedPRs()

	o.sendUpdate(StateUpdate{
		Type:      UpdatePollComplete,
		Timestamp: time.Now(),
	})
}

// pollCodebase polls a single codebase for issues
func (o *Orchestrator) pollCodebase(codebase *config.Codebase) {
	o.mu.Lock()
	cbState, exists := o.codebases[codebase.Name]
	if !exists {
		cbState = &CodebaseState{
			Config:    codebase,
			Issues:    make(map[int]*IssueState),
			IsHealthy: true,
		}
		o.codebases[codebase.Name] = cbState
	}
	o.mu.Unlock()

	// Get pickup labels
	pickupLabels := o.getPickupLabels()

	// Fetch issues with pickup labels
	issues, err := o.ghClient.ListIssuesWithLabels(codebase.Repo, pickupLabels)
	if err != nil {
		o.log("Error fetching issues for %s: %v", codebase.Repo, err)
		o.mu.Lock()
		cbState.IsHealthy = false
		cbState.Error = err
		o.mu.Unlock()
		return
	}

	o.mu.Lock()
	cbState.IsHealthy = true
	cbState.Error = nil
	cbState.LastPoll = time.Now()

	// Track current issue numbers
	currentIssueNums := make(map[int]bool)
	for _, issue := range issues {
		currentIssueNums[issue.Number] = true
	}

	// Remove issues that are no longer active
	for num := range cbState.Issues {
		if !currentIssueNums[num] {
			delete(cbState.Issues, num)
			o.sendUpdate(StateUpdate{
				Type:      UpdateIssueRemoved,
				Codebase:  codebase.Name,
				IssueNum:  num,
				Timestamp: time.Now(),
			})
		}
	}
	o.mu.Unlock()

	// Process each issue
	for _, issue := range issues {
		o.processIssue(codebase, cbState, issue)
	}
}

// processIssue processes a single issue
func (o *Orchestrator) processIssue(codebase *config.Codebase, cbState *CodebaseState, issue github.Issue) {
	currentLabel := o.getCurrentLabel(&issue)
	if currentLabel == "" {
		return // No dev-swarm label
	}

	labelCfg := o.getLabelConfig(currentLabel)
	if labelCfg == nil {
		return // Unknown label
	}

	sessionID := fmt.Sprintf("%s#%d", codebase.Repo, issue.Number)

	o.mu.Lock()
	issueState, exists := cbState.Issues[issue.Number]
	if !exists {
		issueState = &IssueState{
			Issue:       &issue,
			Label:       currentLabel,
			LastChecked: time.Now(),
		}
		cbState.Issues[issue.Number] = issueState

		o.sendUpdate(StateUpdate{
			Type:      UpdateIssueFound,
			Codebase:  codebase.Name,
			IssueNum:  issue.Number,
			Data:      issueState,
			Timestamp: time.Now(),
		})
	} else {
		// Update issue state
		issueState.Issue = &issue
		if issueState.Label != currentLabel {
			issueState.Label = currentLabel
			o.sendUpdate(StateUpdate{
				Type:      UpdateLabelChanged,
				Codebase:  codebase.Name,
				IssueNum:  issue.Number,
				Data:      currentLabel,
				Timestamp: time.Now(),
			})
		}
		issueState.LastChecked = time.Now()
	}
	o.mu.Unlock()

	// Check if we should pick up this issue
	if !o.ShouldPickup(&issue, labelCfg) {
		return
	}

	// For conditional pickup, we need full issue details with comments
	var fullIssue *github.Issue
	if labelCfg.AIPickup == string(config.PickupOnUserComment) {
		var err error
		fullIssue, err = o.ghClient.GetIssue(codebase.Repo, issue.Number)
		if err != nil {
			o.log("Error fetching issue details for %s#%d: %v", codebase.Repo, issue.Number, err)
			return
		}

		// Re-check with full comments
		if !o.hasNewUserComment(fullIssue) {
			// Also check PR comments for code review
			if currentLabel == o.config.Labels.CodeReview.Name {
				if !o.hasNewUserPRComment(codebase.Repo, issue.Number) {
					return
				}
			} else {
				return
			}
		}
	} else {
		fullIssue = &issue
	}

	// Spawn session
	o.log("Picking up issue %s#%d (label: %s)", codebase.Repo, issue.Number, currentLabel)

	req := session.SpawnRequest{
		Issue:        fullIssue,
		Codebase:     codebase,
		CurrentLabel: currentLabel,
		AIAction:     labelCfg.AIAction,
	}

	sess, err := o.sessionManager.SpawnSession(req, o.config.AIInstructions.General)
	if err != nil {
		o.log("Error spawning session for %s#%d: %v", codebase.Repo, issue.Number, err)
		return
	}

	o.mu.Lock()
	issueState.HasSession = true
	issueState.SessionID = sessionID
	o.mu.Unlock()

	o.sendUpdate(StateUpdate{
		Type:      UpdateSessionStarted,
		Codebase:  codebase.Name,
		IssueNum:  issue.Number,
		Data:      sess.Info(),
		Timestamp: time.Now(),
	})
}

// checkSessionStatus checks the status of all active sessions
func (o *Orchestrator) checkSessionStatus() {
	sessions := o.sessionManager.GetAllSessions()
	for _, sess := range sessions {
		if sess.IsComplete() {
			// Session finished, update issue state
			o.mu.Lock()
			for _, cb := range o.codebases {
				for _, issue := range cb.Issues {
					if issue.SessionID == sess.ID {
						issue.HasSession = false
						issue.SessionID = ""
						break
					}
				}
			}
			o.mu.Unlock()

			// Clean up session if done label was set
			if sess.Status == session.StatusCompleted {
				// Check if issue now has done label
				parts := strings.Split(sess.ID, "#")
				if len(parts) == 2 {
					repo := parts[0]
					issueNum := sess.Issue.Number
					issue, err := o.ghClient.GetIssue(repo, issueNum)
					if err == nil && issue.HasLabel(o.config.Labels.Done.Name) {
						// Clean up worktree
						codebase := o.config.GetCodebaseByRepo(repo)
						if codebase != nil {
							worktreePath := git.GetWorktreePath(config.WorktreesDir(), codebase.Name, issueNum)
							git.RemoveWorktree(codebase.LocalPath, worktreePath, false)
						}
					}
				}
			}

			// Remove from manager
			o.sessionManager.RemoveSession(sess.ID)
		}
	}
}

// checkCIStatus checks CI status for issues with implementing or code review labels
func (o *Orchestrator) checkCIStatus() {
	for _, cb := range o.codebases {
		for _, issueState := range cb.Issues {
			// Only check CI for implementing or code review states
			if issueState.Label != o.config.Labels.Implementing.Name &&
				issueState.Label != o.config.Labels.CodeReview.Name {
				continue
			}

			// Don't check if there's an active session
			if issueState.HasSession {
				continue
			}

			// Get PR for this issue
			branchName := git.GetBranchName(issueState.Issue.Number)
			pr, err := o.ghClient.GetPRForBranch(cb.Config.Repo, branchName)
			if err != nil || pr == nil {
				continue
			}

			// Check if CI failed
			failed, err := o.ghClient.PRChecksFailed(cb.Config.Repo, pr.Number)
			if err != nil {
				continue
			}

			if failed && issueState.Label != o.config.Labels.CIFailed.Name {
				// Update label to ci-failed
				err := o.ghClient.UpdateIssueLabels(
					cb.Config.Repo,
					issueState.Issue.Number,
					[]string{issueState.Label},
					[]string{o.config.Labels.CIFailed.Name},
				)
				if err != nil {
					o.log("Error updating label for %s#%d: %v", cb.Config.Repo, issueState.Issue.Number, err)
				} else {
					issueState.Label = o.config.Labels.CIFailed.Name
					o.sendUpdate(StateUpdate{
						Type:      UpdateLabelChanged,
						Codebase:  cb.Config.Name,
						IssueNum:  issueState.Issue.Number,
						Data:      o.config.Labels.CIFailed.Name,
						Timestamp: time.Now(),
					})
				}
			}
		}
	}
}

// cleanupMergedPRs cleans up worktrees for merged PRs
func (o *Orchestrator) cleanupMergedPRs() {
	for _, cb := range o.config.GetEnabledCodebases() {
		prs, err := o.ghClient.GetMergedPRs(cb.Repo)
		if err != nil {
			continue
		}

		for _, pr := range prs {
			// Check if this is a claude branch
			if !strings.HasPrefix(pr.HeadRef, "claude/issue-") {
				continue
			}

			// Extract issue number from branch name
			var issueNum int
			_, err := fmt.Sscanf(pr.HeadRef, "claude/issue-%d", &issueNum)
			if err != nil {
				continue
			}

			// Clean up worktree if it exists
			worktreePath := git.GetWorktreePath(config.WorktreesDir(), cb.Name, issueNum)
			if git.WorktreeExists(worktreePath) {
				git.RemoveWorktree(cb.LocalPath, worktreePath, true)
				o.log("Cleaned up worktree for merged PR: %s", pr.HeadRef)
			}
		}
	}
}
