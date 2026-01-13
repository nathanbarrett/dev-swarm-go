package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

var (
	syncAll bool
)

func newSyncLabelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync-labels [repo]",
		Short: "Sync labels to repositories",
		Long: `Create dev-swarm labels in GitHub repositories.

This creates any missing labels with the correct colors and descriptions.

Example:
  dev-swarm sync-labels owner/repo
  dev-swarm sync-labels --all`,
		RunE: runSyncLabels,
	}

	cmd.Flags().BoolVarP(&syncAll, "all", "a", false, "sync labels to all enabled repositories")

	return cmd
}

func runSyncLabels(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ghClient := github.NewClient()

	// Build label list
	labels := cfg.Labels.GetAllLabels()
	labelInfos := make([]github.LabelInfo, 0, len(labels))
	for _, l := range labels {
		labelInfos = append(labelInfos, github.LabelInfo{
			Name:        l.Name,
			Color:       l.Color,
			Description: l.Description,
		})
	}

	var repos []string

	if syncAll {
		for _, cb := range cfg.GetEnabledCodebases() {
			repos = append(repos, cb.Repo)
		}
	} else if len(args) > 0 {
		repos = args
	} else {
		return fmt.Errorf("specify a repo or use --all")
	}

	for _, repo := range repos {
		fmt.Printf("Syncing labels for %s...\n", repo)

		if err := ghClient.SyncLabels(repo, labelInfos); err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		fmt.Printf("  Created/updated %d labels\n", len(labelInfos))
	}

	return nil
}
