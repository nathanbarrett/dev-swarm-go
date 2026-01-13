package github

import (
	"fmt"
	"time"
)

// Issue represents a GitHub issue
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	URL       string    `json:"url"`
	Labels    []Label   `json:"labels"`
	Comments  []Comment `json:"comments"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ID returns a unique identifier for the issue
func (i *Issue) ID() string {
	return fmt.Sprintf("#%d", i.Number)
}

// HasLabel checks if the issue has a specific label
func (i *Issue) HasLabel(name string) bool {
	for _, l := range i.Labels {
		if l.Name == name {
			return true
		}
	}
	return false
}

// GetCurrentLabel returns the first dev-swarm label found
func (i *Issue) GetCurrentLabel() string {
	for _, l := range i.Labels {
		// Check for dev-swarm labels (user: or ai: prefix)
		if len(l.Name) > 3 && (l.Name[:5] == "user:" || l.Name[:3] == "ai:") {
			return l.Name
		}
	}
	return ""
}

// Label represents a GitHub label
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID        int       `json:"id"`
	Author    Author    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// Author represents a GitHub user
type Author struct {
	Login string `json:"login"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	URL       string    `json:"url"`
	HeadRef   string    `json:"headRefName"`
	BaseRef   string    `json:"baseRefName"`
	Merged    bool      `json:"merged"`
	CreatedAt time.Time `json:"createdAt"`
}

// PRCheck represents a CI check on a PR
type PRCheck struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// PRReview represents a PR review
type PRReview struct {
	ID        int       `json:"id"`
	Author    Author    `json:"author"`
	Body      string    `json:"body"`
	State     string    `json:"state"` // APPROVED, CHANGES_REQUESTED, COMMENTED, etc.
	CreatedAt time.Time `json:"submittedAt"`
}

// PRComment represents a PR review comment (inline comment)
type PRComment struct {
	ID        int       `json:"id"`
	Author    Author    `json:"author"`
	Body      string    `json:"body"`
	Path      string    `json:"path"`
	Line      int       `json:"line"`
	CreatedAt time.Time `json:"createdAt"`
}

// LabelInfo represents label information for listing
type LabelInfo struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Workflow represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID         int       `json:"databaseId"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"createdAt"`
}
