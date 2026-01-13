package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured codebases",
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Codebases) == 0 {
		fmt.Println("No codebases configured.")
		fmt.Println("Use 'dev-swarm add --repo owner/repo --path /path/to/repo' to add one.")
		return nil
	}

	fmt.Println("Configured codebases:")
	fmt.Println()

	for _, cb := range cfg.Codebases {
		status := "enabled"
		if !cb.Enabled {
			status = "disabled"
		}

		fmt.Printf("  %s (%s)\n", cb.Name, status)
		fmt.Printf("    Repo:   %s\n", cb.Repo)
		fmt.Printf("    Path:   %s\n", cb.LocalPath)
		fmt.Printf("    Branch: %s\n", cb.DefaultBranch)
		fmt.Println()
	}

	return nil
}
