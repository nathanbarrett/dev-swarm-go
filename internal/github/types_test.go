package github

import (
	"testing"
	"time"
)

func TestIssueID(t *testing.T) {
	issue := &Issue{Number: 42}
	expected := "#42"

	if issue.ID() != expected {
		t.Errorf("Issue.ID() = %q, want %q", issue.ID(), expected)
	}
}

func TestIssueHasLabel(t *testing.T) {
	issue := &Issue{
		Number: 1,
		Labels: []Label{
			{Name: "bug"},
			{Name: "user:ready-to-plan"},
			{Name: "priority:high"},
		},
	}

	tests := []struct {
		label string
		want  bool
	}{
		{"bug", true},
		{"user:ready-to-plan", true},
		{"priority:high", true},
		{"feature", false},
		{"ai:planning", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := issue.HasLabel(tt.label)
			if got != tt.want {
				t.Errorf("HasLabel(%q) = %v, want %v", tt.label, got, tt.want)
			}
		})
	}
}

func TestIssueGetCurrentLabel(t *testing.T) {
	tests := []struct {
		name   string
		labels []Label
		want   string
	}{
		{
			name: "user label",
			labels: []Label{
				{Name: "bug"},
				{Name: "user:ready-to-plan"},
			},
			want: "user:ready-to-plan",
		},
		{
			name: "ai label",
			labels: []Label{
				{Name: "ai:planning"},
				{Name: "bug"},
			},
			want: "ai:planning",
		},
		{
			name: "no dev-swarm label",
			labels: []Label{
				{Name: "bug"},
				{Name: "feature"},
			},
			want: "",
		},
		{
			name:   "empty labels",
			labels: []Label{},
			want:   "",
		},
		{
			name: "multiple dev-swarm labels returns first",
			labels: []Label{
				{Name: "user:ready-to-plan"},
				{Name: "ai:planning"},
			},
			want: "user:ready-to-plan",
		},
		{
			name: "short labels ignored",
			labels: []Label{
				{Name: "ai"},
				{Name: "user:ready-to-plan"},
			},
			want: "user:ready-to-plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{Labels: tt.labels}
			got := issue.GetCurrentLabel()
			if got != tt.want {
				t.Errorf("GetCurrentLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssueFields(t *testing.T) {
	now := time.Now()
	issue := &Issue{
		Number:    42,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     "open",
		URL:       "https://github.com/owner/repo/issues/42",
		Labels:    []Label{{Name: "bug"}},
		Comments:  []Comment{{ID: 1, Body: "test"}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if issue.Number != 42 {
		t.Errorf("Number = %d, want 42", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Test Issue")
	}
	if issue.State != "open" {
		t.Errorf("State = %q, want %q", issue.State, "open")
	}
}

func TestLabelFields(t *testing.T) {
	label := Label{
		Name:        "bug",
		Color:       "FF0000",
		Description: "Something isn't working",
	}

	if label.Name != "bug" {
		t.Errorf("Name = %q, want %q", label.Name, "bug")
	}
	if label.Color != "FF0000" {
		t.Errorf("Color = %q, want %q", label.Color, "FF0000")
	}
}

func TestCommentFields(t *testing.T) {
	now := time.Now()
	comment := Comment{
		ID:        123,
		Author:    Author{Login: "testuser"},
		Body:      "Test comment",
		CreatedAt: now,
	}

	if comment.ID != 123 {
		t.Errorf("ID = %d, want 123", comment.ID)
	}
	if comment.Author.Login != "testuser" {
		t.Errorf("Author.Login = %q, want %q", comment.Author.Login, "testuser")
	}
}

func TestPullRequestFields(t *testing.T) {
	pr := PullRequest{
		Number:  1,
		Title:   "Test PR",
		Body:    "Test body",
		State:   "open",
		URL:     "https://github.com/owner/repo/pull/1",
		HeadRef: "feature-branch",
		BaseRef: "main",
		Merged:  false,
	}

	if pr.Number != 1 {
		t.Errorf("Number = %d, want 1", pr.Number)
	}
	if pr.HeadRef != "feature-branch" {
		t.Errorf("HeadRef = %q, want %q", pr.HeadRef, "feature-branch")
	}
	if pr.Merged {
		t.Error("Merged should be false")
	}
}

func TestPRCheckFields(t *testing.T) {
	check := PRCheck{
		Name:       "tests",
		Status:     "completed",
		Conclusion: "success",
	}

	if check.Name != "tests" {
		t.Errorf("Name = %q, want %q", check.Name, "tests")
	}
	if check.Conclusion != "success" {
		t.Errorf("Conclusion = %q, want %q", check.Conclusion, "success")
	}
}

func TestLabelInfoFields(t *testing.T) {
	info := LabelInfo{
		Name:        "bug",
		Color:       "FF0000",
		Description: "Bug label",
	}

	if info.Name != "bug" {
		t.Errorf("Name = %q, want %q", info.Name, "bug")
	}
}
