package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
)

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name or repo]",
		Short: "Remove a codebase from monitoring",
		Long: `Remove a repository from the dev-swarm configuration.

You can specify either the codebase name or the full repo identifier.

Example:
  dev-swarm remove my-project
  dev-swarm remove owner/repo`,
		Args: cobra.ExactArgs(1),
		RunE: runRemove,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find and remove codebase
	found := false
	var newCodebases []config.Codebase
	for _, cb := range cfg.Codebases {
		if cb.Name == identifier || cb.Repo == identifier {
			found = true
			fmt.Printf("Removed codebase '%s' (%s)\n", cb.Name, cb.Repo)
		} else {
			newCodebases = append(newCodebases, cb)
		}
	}

	if !found {
		return fmt.Errorf("codebase not found: %s", identifier)
	}

	cfg.Codebases = newCodebases

	// Save config
	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigFilePath()
	}
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
