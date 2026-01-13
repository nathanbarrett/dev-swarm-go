package orchestrator

import (
	"strings"
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

// ShouldPickup determines if an issue should be picked up for processing
func (o *Orchestrator) ShouldPickup(issue *github.Issue, labelCfg *config.LabelConfig) bool {
	sessionID := o.getSessionID(issue)

	// Already has active session?
	if o.sessionManager.HasSession(sessionID) {
		return false
	}

	// At max capacity?
	if !o.sessionManager.CanSpawn() {
		return false
	}

	// Check pickup rule
	switch labelCfg.AIPickup {
	case string(config.PickupAlways):
		return true

	case string(config.PickupNever):
		return false

	case string(config.PickupOnUserComment):
		return o.hasNewUserComment(issue)

	default:
		return false
	}
}

// hasNewUserComment checks if there's a new user comment since the last AI comment
func (o *Orchestrator) hasNewUserComment(issue *github.Issue) bool {
	comments := issue.Comments
	if len(comments) == 0 {
		return false
	}

	var lastAICommentTime time.Time
	var lastUserCommentTime time.Time

	for _, comment := range comments {
		isAI := strings.Contains(comment.Body, "<!-- dev-swarm:ai -->")

		if isAI {
			if comment.CreatedAt.After(lastAICommentTime) {
				lastAICommentTime = comment.CreatedAt
			}
		} else {
			if comment.CreatedAt.After(lastUserCommentTime) {
				lastUserCommentTime = comment.CreatedAt
			}
		}
	}

	// New user comment exists if it's after the last AI comment
	// (or if there's never been an AI comment but there are user comments)
	if lastUserCommentTime.IsZero() {
		return false
	}

	return lastUserCommentTime.After(lastAICommentTime)
}

// hasNewUserPRComment checks if there's a new user comment on the PR since the last AI comment
func (o *Orchestrator) hasNewUserPRComment(repo string, issueNumber int) bool {
	branchName := o.getBranchName(issueNumber)

	// Get PR for this issue
	pr, err := o.ghClient.GetPRForBranch(repo, branchName)
	if err != nil || pr == nil {
		return false
	}

	// Get PR comments
	comments, err := o.ghClient.GetPRComments(repo, pr.Number)
	if err != nil {
		return false
	}

	// Get PR reviews
	reviews, err := o.ghClient.GetPRReviews(repo, pr.Number)
	if err != nil {
		return false
	}

	var lastAITime time.Time
	var lastUserTime time.Time

	// Check comments
	for _, comment := range comments {
		isAI := strings.Contains(comment.Body, "<!-- dev-swarm:ai -->")
		if isAI {
			if comment.CreatedAt.After(lastAITime) {
				lastAITime = comment.CreatedAt
			}
		} else {
			if comment.CreatedAt.After(lastUserTime) {
				lastUserTime = comment.CreatedAt
			}
		}
	}

	// Check reviews
	for _, review := range reviews {
		isAI := strings.Contains(review.Body, "<!-- dev-swarm:ai -->")
		if isAI {
			if review.CreatedAt.After(lastAITime) {
				lastAITime = review.CreatedAt
			}
		} else {
			if review.CreatedAt.After(lastUserTime) {
				lastUserTime = review.CreatedAt
			}
		}
	}

	if lastUserTime.IsZero() {
		return false
	}

	return lastUserTime.After(lastAITime)
}

// getPickupLabels returns all label names that should be polled
func (o *Orchestrator) getPickupLabels() []string {
	pickupLabels := o.config.Labels.GetPickupLabels()
	names := make([]string, 0, len(pickupLabels))
	for _, label := range pickupLabels {
		names = append(names, label.Name)
	}
	return names
}

// getLabelConfig returns the label config for a given label name
func (o *Orchestrator) getLabelConfig(name string) *config.LabelConfig {
	return o.config.Labels.GetByName(name)
}

// getCurrentLabel returns the current dev-swarm label for an issue
func (o *Orchestrator) getCurrentLabel(issue *github.Issue) string {
	for _, label := range issue.Labels {
		// Check for dev-swarm labels
		if strings.HasPrefix(label.Name, "user:") || strings.HasPrefix(label.Name, "ai:") {
			return label.Name
		}
	}
	return ""
}

// getAIAction returns the AI action for a label
func (o *Orchestrator) getAIAction(labelName string) string {
	labelCfg := o.getLabelConfig(labelName)
	if labelCfg != nil {
		return labelCfg.AIAction
	}
	return ""
}
