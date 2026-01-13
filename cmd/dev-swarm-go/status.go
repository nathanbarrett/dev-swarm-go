package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/lock"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show orchestrator status",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	lck := lock.New()
	status := lck.GetStatus()

	if status.IsLocked {
		fmt.Printf("dev-swarm is running (PID: %d)\n", status.PID)
	} else {
		fmt.Println("dev-swarm is not running")
	}

	// Show config info
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Printf("\nConfig: not found or invalid (%v)\n", err)
		return nil
	}

	enabledCount := 0
	for _, cb := range cfg.Codebases {
		if cb.Enabled {
			enabledCount++
		}
	}

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Config file: %s\n", config.ConfigFilePath())
	fmt.Printf("  Codebases: %d total, %d enabled\n", len(cfg.Codebases), enabledCount)
	fmt.Printf("  Max sessions: %d\n", cfg.Settings.MaxConcurrentSessions)
	fmt.Printf("  Poll interval: %ds (active: %ds)\n", cfg.Settings.PollInterval, cfg.Settings.ActivePollInterval)

	if len(cfg.Codebases) > 0 {
		fmt.Printf("\nCodebases:\n")
		for _, cb := range cfg.Codebases {
			status := "enabled"
			if !cb.Enabled {
				status = "disabled"
			}
			fmt.Printf("  - %s (%s): %s\n", cb.Name, cb.Repo, status)
		}
	}

	return nil
}
