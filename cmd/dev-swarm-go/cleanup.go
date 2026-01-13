package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/git"
)

func newCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up orphaned worktrees",
		Long: `Remove orphaned git worktrees that are no longer needed.

This can happen if dev-swarm crashes or worktrees are left over from merged PRs.`,
		RunE: runCleanup,
	}
}

func runCleanup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	worktreesDir := config.WorktreesDir()

	// Check if worktrees directory exists
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		fmt.Println("No worktrees directory found. Nothing to clean up.")
		return nil
	}

	cleaned := 0

	// Iterate through codebase directories
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		return fmt.Errorf("failed to read worktrees directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		codebaseName := entry.Name()
		codebasePath := filepath.Join(worktreesDir, codebaseName)

		// Find the codebase config
		var codebase *config.Codebase
		for _, cb := range cfg.Codebases {
			if cb.Name == codebaseName {
				codebase = &cb
				break
			}
		}

		// Read issue directories
		issueEntries, err := os.ReadDir(codebasePath)
		if err != nil {
			continue
		}

		for _, issueEntry := range issueEntries {
			if !issueEntry.IsDir() {
				continue
			}

			if !strings.HasPrefix(issueEntry.Name(), "issue-") {
				continue
			}

			worktreePath := filepath.Join(codebasePath, issueEntry.Name())

			// Check if worktree should be cleaned
			shouldClean := false

			if codebase == nil {
				// Codebase no longer configured
				shouldClean = true
			} else if !git.IsGitRepo(worktreePath) {
				// Invalid worktree
				shouldClean = true
			}

			if shouldClean {
				fmt.Printf("Removing orphaned worktree: %s\n", worktreePath)
				os.RemoveAll(worktreePath)
				cleaned++
			}
		}

		// Clean up git worktree references
		if codebase != nil {
			git.CleanupOrphanedWorktrees(codebase.LocalPath, codebasePath)
		}
	}

	if cleaned == 0 {
		fmt.Println("No orphaned worktrees found.")
	} else {
		fmt.Printf("Cleaned up %d orphaned worktrees.\n", cleaned)
	}

	return nil
}
