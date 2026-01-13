package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nathanbarrett/dev-swarm-go/internal/orchestrator"
)

// Run starts the TUI application
func Run(orch *orchestrator.Orchestrator) error {
	model := NewModel(orch)

	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
