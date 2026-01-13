package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nathanbarrett/dev-swarm-go/internal/orchestrator"
	"github.com/nathanbarrett/dev-swarm-go/internal/session"
)

type CodebaseInfo = orchestrator.CodebaseInfo
type IssueInfo = orchestrator.IssueInfo

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return "Shutting down...\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Show help overlay
	if m.showHelp {
		return m.renderHelp()
	}

	// Show focused session output
	if m.focusedSession != "" {
		return m.renderFocusedOutput()
	}

	// Normal split view
	return m.renderSplitView()
}

// renderSplitView renders the normal split view
func (m Model) renderSplitView() string {
	// Calculate heights
	statusBarHeight := 1
	titleHeight := 1
	separatorHeight := 1
	outputHeight := m.height / 3
	listHeight := m.height - statusBarHeight - titleHeight - separatorHeight - outputHeight - 2

	// Title bar
	title := TitleStyle.Render(fmt.Sprintf(" dev-swarm %s ", m.Version()))
	titleLine := lipgloss.NewStyle().Width(m.width).Render(title)

	// Codebase list
	listContent := m.renderCodebaseList(listHeight)

	// Output panel
	outputContent := m.renderOutputPanel(outputHeight)

	// Status bar
	statusBar := m.renderStatusBar()

	// Combine
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		listContent,
		strings.Repeat("─", m.width),
		outputContent,
		statusBar,
	)
}

// renderCodebaseList renders the list of codebases and issues
func (m Model) renderCodebaseList(height int) string {
	var lines []string
	currentLine := 0

	for cbIdx, cb := range m.codebases {
		// Codebase header
		cbLine := m.renderCodebaseHeader(cb, cbIdx)
		if currentLine == m.selectedIdx {
			cbLine = SelectedStyle.Render(cbLine)
		}
		lines = append(lines, cbLine)
		currentLine++

		if len(cb.Issues) == 0 {
			// Show idle message
			idleLine := fmt.Sprintf("  %s (idle)", TreeLeaf)
			idleLine = CodebaseIdleStyle.Render(idleLine)
			lines = append(lines, idleLine)
		} else {
			// Render issues
			for issueIdx, issue := range cb.Issues {
				isLast := issueIdx == len(cb.Issues)-1
				issueLine := m.renderIssueLine(issue, isLast)

				if currentLine == m.selectedIdx {
					issueLine = SelectedStyle.Render(issueLine)
				}
				lines = append(lines, issueLine)
				currentLine++
			}
		}

		// Add spacing between codebases
		if cbIdx < len(m.codebases)-1 {
			lines = append(lines, "")
		}
	}

	if len(lines) == 0 {
		lines = append(lines, CodebaseIdleStyle.Render("  No codebases configured"))
	}

	// Handle scrolling
	content := strings.Join(lines, "\n")

	// Pad to fill height
	lineCount := len(lines)
	if lineCount < height {
		padding := strings.Repeat("\n", height-lineCount)
		content += padding
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(height).
		Render(content)
}

// renderCodebaseHeader renders a codebase header line
func (m Model) renderCodebaseHeader(cb CodebaseInfo, _ int) string {
	name := CodebaseStyle.Render(cb.Name)
	repo := lipgloss.NewStyle().Foreground(ColorGray).Render(" • " + cb.Repo)

	health := ""
	if !cb.IsHealthy {
		health = ErrorStyle.Render(" [error]")
	}

	return fmt.Sprintf("  %s%s%s", name, repo, health)
}

// renderIssueLine renders an issue line
func (m Model) renderIssueLine(issue IssueInfo, isLast bool) string {
	// Tree connector
	connector := TreeBranch
	if isLast {
		connector = TreeLeaf
	}

	// Issue number and title
	number := IssueNumberStyle.Render(fmt.Sprintf("#%d", issue.Number))
	title := issue.Title
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	titleStyled := IssueStyle.Render(title)

	// Label and status
	labelStyle := GetLabelStyle(issue.Label)
	label := labelStyle.Render(issue.Label)

	// Status indicator
	isActive := issue.HasSession && issue.Status == session.StatusRunning
	isFailed := issue.HasSession && issue.Status == session.StatusFailed
	isDone := strings.Contains(issue.Label, "done")
	icon := GetStatusIcon(isActive, isFailed, isDone)

	// Duration if active
	duration := ""
	if isActive {
		duration = lipgloss.NewStyle().Foreground(ColorGray).Render(
			fmt.Sprintf(" (%s)", formatDuration(issue.Duration)),
		)
	}

	return fmt.Sprintf("  %s %s %s  %s %s%s",
		connector, number, titleStyled, label, icon, duration)
}

// renderOutputPanel renders the output panel
func (m Model) renderOutputPanel(height int) string {
	// Get selected session's output
	sess := m.GetSelectedSession()
	if sess == nil {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(height).
			Foreground(ColorGray).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No active session selected")
	}

	// Get session from manager for output
	fullSession := m.orchestrator.GetSessionManager().GetSession(sess.ID)
	if fullSession == nil {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(height).
			Foreground(ColorGray).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Session not found")
	}

	// Title
	title := OutputTitleStyle.Render(
		fmt.Sprintf(" Output: #%d %s ", sess.IssueNumber, sess.IssueTitle),
	)

	// Output lines
	outputLines := fullSession.GetRecentOutput(height - 2)
	var lines []string
	for _, line := range outputLines {
		timestamp := OutputTimestampStyle.Render(
			line.Timestamp.Format("[15:04:05]"),
		)
		text := OutputLineStyle.Render(line.Text)
		lines = append(lines, fmt.Sprintf("%s %s", timestamp, text))
	}

	content := strings.Join(lines, "\n")
	if len(lines) < height-2 {
		content += strings.Repeat("\n", height-2-len(lines))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	stats := m.Stats()

	// Active count
	activeText := StatusBarValueStyle.Render(fmt.Sprintf("Active: %d", stats.ActiveSessions))
	if stats.ActiveSessions > 0 {
		activeText = StatusBarActiveStyle.Render(fmt.Sprintf("Active: %d", stats.ActiveSessions))
	}

	// Total issues
	totalText := StatusBarValueStyle.Render(fmt.Sprintf("Issues: %d", stats.TotalIssues))

	// Poll countdown
	pollText := StatusBarValueStyle.Render("Poll: --")
	if !stats.LastPoll.IsZero() {
		remaining := time.Until(stats.NextPoll)
		if remaining < 0 {
			remaining = 0
		}
		pollText = StatusBarValueStyle.Render(fmt.Sprintf("Poll: %ds", int(remaining.Seconds())))
	}

	// Paused indicator
	pausedText := ""
	if stats.IsPaused {
		pausedText = StatusBarActiveStyle.Render(" [PAUSED]")
	}

	// Help hint
	helpText := HelpStyle.Render("↑↓ Nav  r Refresh  p Pause  q Quit  ? Help")

	// Combine
	left := fmt.Sprintf("  %s  │  %s  │  %s%s", activeText, totalText, pollText, pausedText)
	right := helpText

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 0 {
		gap = 0
	}

	return StatusBarStyle.Width(m.width).Render(
		fmt.Sprintf("%s%s%s", left, strings.Repeat(" ", gap), right),
	)
}

// renderFocusedOutput renders full-screen output for a session
func (m Model) renderFocusedOutput() string {
	sess := m.GetFocusedSession()
	if sess == nil {
		m.focusedSession = ""
		return m.renderSplitView()
	}

	// Title
	title := TitleStyle.Render(
		fmt.Sprintf(" Output: #%d %s ", sess.Issue.Number, sess.Issue.Title),
	)

	// Status bar with escape hint
	status := StatusBarStyle.Render("  ESC to return  │  ↑↓ Scroll  │  q Quit")

	// Calculate content height
	contentHeight := m.height - 3

	// Get output
	outputLines := sess.GetOutput()
	var lines []string
	for _, line := range outputLines {
		timestamp := OutputTimestampStyle.Render(
			line.Timestamp.Format("[15:04:05]"),
		)
		text := OutputLineStyle.Render(line.Text)
		lines = append(lines, fmt.Sprintf("%s %s", timestamp, text))
	}

	// Apply scroll offset
	start := m.outputScroll
	if start > len(lines) {
		start = len(lines)
	}
	end := start + contentHeight
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]
	content := strings.Join(visibleLines, "\n")

	// Pad
	if len(visibleLines) < contentHeight {
		content += strings.Repeat("\n", contentHeight-len(visibleLines))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, content, status)
}

// renderHelp renders the help overlay
func (m Model) renderHelp() string {
	help := `
  dev-swarm Help
  ═══════════════

  Navigation
  ──────────
  ↑/k        Move up
  ↓/j        Move down
  Enter      Focus on session output
  Esc        Return to split view

  Actions
  ───────
  r          Force refresh (poll now)
  p          Pause/resume polling
  l          Toggle log panel

  General
  ───────
  q          Quit
  ?          Show this help

  Press any key to close...
`

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(help)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}
