package session

import (
	"strings"
	"testing"
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

func TestBuildContext(t *testing.T) {
	issue := &github.Issue{
		Number: 42,
		Title:  "Test Issue",
		Body:   "This is the issue body",
		URL:    "https://github.com/owner/repo/issues/42",
		Comments: []github.Comment{
			{
				ID:        1,
				Author:    github.Author{Login: "user1"},
				Body:      "First comment",
				CreatedAt: time.Now().Add(-time.Hour),
			},
			{
				ID:        2,
				Author:    github.Author{Login: "bot"},
				Body:      AICommentMarkerStart + "\nAI response\n" + AICommentMarkerEnd,
				CreatedAt: time.Now(),
			},
		},
	}

	codebase := &config.Codebase{
		Name:          "test-repo",
		Repo:          "owner/repo",
		LocalPath:     "/path/to/repo",
		DefaultBranch: "main",
	}

	currentLabel := "user:ready-to-plan"
	aiAction := "Create an implementation plan"
	aiInstructions := "Follow the guidelines"

	ctx := BuildContext(issue, codebase, currentLabel, aiAction, aiInstructions)

	// Should contain header
	if !strings.Contains(ctx, "# dev-swarm Task") {
		t.Error("Context should contain header")
	}

	// Should contain repository info
	if !strings.Contains(ctx, "owner/repo") {
		t.Error("Context should contain repo")
	}
	if !strings.Contains(ctx, "/path/to/repo") {
		t.Error("Context should contain local path")
	}
	if !strings.Contains(ctx, "main") {
		t.Error("Context should contain default branch")
	}

	// Should contain issue info
	if !strings.Contains(ctx, "#42") {
		t.Error("Context should contain issue number")
	}
	if !strings.Contains(ctx, "Test Issue") {
		t.Error("Context should contain issue title")
	}
	if !strings.Contains(ctx, "This is the issue body") {
		t.Error("Context should contain issue body")
	}

	// Should contain comments
	if !strings.Contains(ctx, "user1") {
		t.Error("Context should contain comment author")
	}
	if !strings.Contains(ctx, "First comment") {
		t.Error("Context should contain comment body")
	}

	// Should mark AI comments
	if !strings.Contains(ctx, "(AI)") {
		t.Error("Context should mark AI comments")
	}

	// Should contain current label
	if !strings.Contains(ctx, currentLabel) {
		t.Error("Context should contain current label")
	}

	// Should contain AI action
	if !strings.Contains(ctx, aiAction) {
		t.Error("Context should contain AI action")
	}

	// Should contain AI instructions
	if !strings.Contains(ctx, aiInstructions) {
		t.Error("Context should contain AI instructions")
	}

	// Should contain guidelines
	if !strings.Contains(ctx, "Label Management") {
		t.Error("Context should contain label management guidelines")
	}
	if !strings.Contains(ctx, "Comment Markers") {
		t.Error("Context should contain comment marker guidelines")
	}
}

func TestBuildContextWithManyComments(t *testing.T) {
	// Create issue with more than 20 comments
	comments := make([]github.Comment, 25)
	for i := 0; i < 25; i++ {
		comments[i] = github.Comment{
			ID:        i,
			Author:    github.Author{Login: "user"},
			Body:      "Comment " + string(rune('A'+i)),
			CreatedAt: time.Now(),
		}
	}

	issue := &github.Issue{
		Number:   1,
		Title:    "Test",
		Body:     "Body",
		Comments: comments,
	}

	codebase := &config.Codebase{
		Repo:          "owner/repo",
		LocalPath:     "/path",
		DefaultBranch: "main",
	}

	ctx := BuildContext(issue, codebase, "label", "", "")

	// Should indicate showing last 20
	if !strings.Contains(ctx, "Showing last 20 of 25") {
		t.Error("Context should indicate showing last 20 comments")
	}

	// Should not contain first comment (index 0-4)
	if strings.Contains(ctx, "Comment A") {
		t.Error("Context should not contain first comments")
	}
}

func TestBuildContextNoComments(t *testing.T) {
	issue := &github.Issue{
		Number:   1,
		Title:    "Test",
		Body:     "Body",
		Comments: []github.Comment{},
	}

	codebase := &config.Codebase{
		Repo:          "owner/repo",
		LocalPath:     "/path",
		DefaultBranch: "main",
	}

	ctx := BuildContext(issue, codebase, "label", "", "")

	// Should not contain comments section header if no comments
	// Actually it should still have the structure but no comments listed
	if !strings.Contains(ctx, "## Issue") {
		t.Error("Context should contain issue section")
	}
}

func TestBuildContextNoAIAction(t *testing.T) {
	issue := &github.Issue{
		Number: 1,
		Title:  "Test",
		Body:   "Body",
	}

	codebase := &config.Codebase{
		Repo:          "owner/repo",
		LocalPath:     "/path",
		DefaultBranch: "main",
	}

	ctx := BuildContext(issue, codebase, "label", "", "")

	// Should not contain instructions section if no AI action
	if strings.Contains(ctx, "### Instructions") {
		t.Error("Context should not contain instructions section when AI action is empty")
	}
}

func TestIsAIComment(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "with start marker",
			body: AICommentMarkerStart + "\nContent\n" + AICommentMarkerEnd,
			want: true,
		},
		{
			name: "marker in middle",
			body: "Some text " + AICommentMarkerStart + " more text",
			want: true,
		},
		{
			name: "no marker",
			body: "Regular comment without markers",
			want: false,
		},
		{
			name: "empty string",
			body: "",
			want: false,
		},
		{
			name: "partial marker",
			body: "<!-- dev-swarm",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAIComment(tt.body)
			if got != tt.want {
				t.Errorf("IsAIComment(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestWrapAIComment(t *testing.T) {
	content := "This is the AI response"
	wrapped := WrapAIComment(content)

	if !strings.HasPrefix(wrapped, AICommentMarkerStart) {
		t.Error("Wrapped comment should start with start marker")
	}
	if !strings.HasSuffix(wrapped, AICommentMarkerEnd) {
		t.Error("Wrapped comment should end with end marker")
	}
	if !strings.Contains(wrapped, content) {
		t.Error("Wrapped comment should contain the content")
	}
}

func TestStripAIMarkers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with both markers",
			input: AICommentMarkerStart + "\nContent\n" + AICommentMarkerEnd,
			want:  "Content",
		},
		{
			name:  "with only start marker",
			input: AICommentMarkerStart + "\nContent",
			want:  "Content",
		},
		{
			name:  "no markers",
			input: "Plain content",
			want:  "Plain content",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripAIMarkers(tt.input)
			if got != tt.want {
				t.Errorf("StripAIMarkers() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAICommentMarkerConstants(t *testing.T) {
	if AICommentMarkerStart != "<!-- dev-swarm:ai -->" {
		t.Errorf("AICommentMarkerStart = %q, want %q", AICommentMarkerStart, "<!-- dev-swarm:ai -->")
	}
	if AICommentMarkerEnd != "<!-- /dev-swarm:ai -->" {
		t.Errorf("AICommentMarkerEnd = %q, want %q", AICommentMarkerEnd, "<!-- /dev-swarm:ai -->")
	}
}
