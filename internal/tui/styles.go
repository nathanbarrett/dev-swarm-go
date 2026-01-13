package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorBlue   = lipgloss.Color("#0052CC")
	ColorYellow = lipgloss.Color("#FBCA04")
	ColorRed    = lipgloss.Color("#D93F0B")
	ColorGreen  = lipgloss.Color("#0E8A16")
	ColorGray   = lipgloss.Color("#6B7280")
	ColorWhite  = lipgloss.Color("#FFFFFF")
	ColorCyan   = lipgloss.Color("#00B4D8")
	ColorDim    = lipgloss.Color("#666666")

	// Base styles
	BaseStyle = lipgloss.NewStyle()

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	// Codebase styles
	CodebaseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlue)

	CodebaseIdleStyle = lipgloss.NewStyle().
				Foreground(ColorGray).
				Italic(true)

	// Issue styles
	IssueStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	IssueNumberStyle = lipgloss.NewStyle().
				Foreground(ColorCyan)

	// Label styles
	LabelActiveStyle = lipgloss.NewStyle().
				Foreground(ColorYellow).
				Bold(true)

	LabelWaitingStyle = lipgloss.NewStyle().
				Foreground(ColorBlue)

	LabelFailedStyle = lipgloss.NewStyle().
				Foreground(ColorRed).
				Bold(true)

	LabelDoneStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	LabelBlockedStyle = lipgloss.NewStyle().
				Foreground(ColorRed)

	// Status icons
	IconActive  = lipgloss.NewStyle().Foreground(ColorYellow).Render("●")
	IconQueued  = lipgloss.NewStyle().Foreground(ColorCyan).Render("◆")
	IconWaiting = lipgloss.NewStyle().Foreground(ColorBlue).Render("○")
	IconFailed  = lipgloss.NewStyle().Foreground(ColorRed).Render("✗")
	IconDone    = lipgloss.NewStyle().Foreground(ColorGreen).Render("✓")

	// Selection style
	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333"))

	// Output panel style
	OutputPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorGray)

	OutputTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(lipgloss.Color("#2a2a2a")).
				Padding(0, 1)

	OutputLineStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	OutputTimestampStyle = lipgloss.NewStyle().
				Foreground(ColorDim)

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a1a")).
			Foreground(ColorGray).
			Padding(0, 1)

	StatusBarActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a1a1a")).
				Foreground(ColorYellow)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Bold(true)

	StatusBarValueStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	// Tree connector styles
	TreeBranch = lipgloss.NewStyle().Foreground(ColorGray).Render("├─")
	TreeLeaf   = lipgloss.NewStyle().Foreground(ColorGray).Render("└─")
	TreeLine   = lipgloss.NewStyle().Foreground(ColorGray).Render("│")
)

// GetLabelStyle returns the appropriate style for a label
func GetLabelStyle(label string) lipgloss.Style {
	switch {
	case label == "ai:planning" || label == "ai:implementing":
		return LabelActiveStyle
	case label == "ai:ci-failed" || label == "user:blocked":
		return LabelFailedStyle
	case label == "ai:done":
		return LabelDoneStyle
	default:
		return LabelWaitingStyle
	}
}

// GetStatusIcon returns the appropriate icon for a status
func GetStatusIcon(isActive bool, isFailed bool, isDone bool) string {
	switch {
	case isFailed:
		return IconFailed
	case isDone:
		return IconDone
	case isActive:
		return IconActive
	default:
		return IconWaiting
	}
}
