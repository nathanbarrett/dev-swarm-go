package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

// AICommentMarkerStart is the marker for AI comment start
const AICommentMarkerStart = "<!-- dev-swarm:ai -->"

// AICommentMarkerEnd is the marker for AI comment end
const AICommentMarkerEnd = "<!-- /dev-swarm:ai -->"

// BuildContext creates the prompt context for a Claude session
func BuildContext(
	issue *github.Issue,
	codebase *config.Codebase,
	currentLabel string,
	aiAction string,
	aiInstructions string,
) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# dev-swarm Task\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format(time.RFC3339)))

	// Repository info
	sb.WriteString("## Repository\n\n")
	sb.WriteString(fmt.Sprintf("- **Repo**: %s\n", codebase.Repo))
	sb.WriteString(fmt.Sprintf("- **Local Path**: %s\n", codebase.LocalPath))
	sb.WriteString(fmt.Sprintf("- **Default Branch**: %s\n", codebase.DefaultBranch))
	sb.WriteString(fmt.Sprintf("- **Working Branch**: claude/issue-%d\n", issue.Number))
	sb.WriteString("\n")

	// Issue info
	sb.WriteString("## Issue\n\n")
	sb.WriteString(fmt.Sprintf("- **Number**: #%d\n", issue.Number))
	sb.WriteString(fmt.Sprintf("- **Title**: %s\n", issue.Title))
	sb.WriteString(fmt.Sprintf("- **URL**: %s\n", issue.URL))
	sb.WriteString(fmt.Sprintf("- **Current Label**: %s\n", currentLabel))
	sb.WriteString("\n")

	// Issue body
	sb.WriteString("### Description\n\n")
	sb.WriteString(issue.Body)
	sb.WriteString("\n\n")

	// Comments (last 20)
	if len(issue.Comments) > 0 {
		sb.WriteString("### Comments\n\n")

		start := 0
		if len(issue.Comments) > 20 {
			start = len(issue.Comments) - 20
			sb.WriteString(fmt.Sprintf("*(Showing last 20 of %d comments)*\n\n", len(issue.Comments)))
		}

		for _, comment := range issue.Comments[start:] {
			isAI := IsAIComment(comment.Body)
			author := comment.Author.Login
			if isAI {
				author = fmt.Sprintf("%s (AI)", author)
			}

			sb.WriteString(fmt.Sprintf("**%s** (%s):\n", author, comment.CreatedAt.Format("2006-01-02 15:04")))
			sb.WriteString(comment.Body)
			sb.WriteString("\n\n---\n\n")
		}
	}

	// Current task
	sb.WriteString("## Your Task\n\n")
	sb.WriteString(fmt.Sprintf("**Current State**: %s\n\n", currentLabel))

	if aiAction != "" {
		sb.WriteString("### Instructions\n\n")
		sb.WriteString(aiAction)
		sb.WriteString("\n\n")
	}

	// General instructions
	sb.WriteString("## Important Guidelines\n\n")
	sb.WriteString(fmt.Sprintf(`1. **Label Management**: Update labels using gh CLI:
`+"```bash"+`
gh issue edit %d --repo %s --remove-label "current:label" --add-label "new:label"
`+"```"+`

2. **Comment Markers**: Wrap ALL your comments with these markers:
`+"```"+`
<!-- dev-swarm:ai -->
Your comment here
<!-- /dev-swarm:ai -->
`+"```"+`

3. **Commit Messages**: Always reference the issue:
- "Add feature X (#%d)"
- "Fix bug in Y (fixes #%d)"

4. **Pull Requests**: Link to the issue:
- Include "Closes #%d" in the PR body

5. **Getting Stuck**: If you cannot proceed:
- Change label to user:blocked
- Add a comment explaining what's blocking you

6. **Code Quality**:
- Follow existing code style
- Write clear, maintainable code
- Add tests when specified
`, issue.Number, codebase.Repo, issue.Number, issue.Number, issue.Number))

	// Add custom AI instructions if provided
	if aiInstructions != "" {
		sb.WriteString("\n## Additional Instructions\n\n")
		sb.WriteString(aiInstructions)
		sb.WriteString("\n")
	}

	return sb.String()
}

// IsAIComment checks if a comment was made by the AI
func IsAIComment(body string) bool {
	return strings.Contains(body, AICommentMarkerStart)
}

// WrapAIComment wraps a comment with AI markers
func WrapAIComment(content string) string {
	return fmt.Sprintf("%s\n%s\n%s", AICommentMarkerStart, content, AICommentMarkerEnd)
}

// StripAIMarkers removes AI markers from a comment
func StripAIMarkers(body string) string {
	body = strings.ReplaceAll(body, AICommentMarkerStart, "")
	body = strings.ReplaceAll(body, AICommentMarkerEnd, "")
	return strings.TrimSpace(body)
}
