package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/git"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
)

var (
	addRepo   string
	addPath   string
	addBranch string
	addName   string
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a codebase to monitor",
		Long: `Add a repository to the dev-swarm configuration.

Example:
  dev-swarm add --repo owner/repo --path ~/code/repo
  dev-swarm add --repo owner/repo --path ~/code/repo --branch main --name my-project`,
		RunE: runAdd,
	}

	cmd.Flags().StringVarP(&addRepo, "repo", "r", "", "GitHub repository (owner/repo format)")
	cmd.Flags().StringVarP(&addPath, "path", "p", "", "Local path to the repository")
	cmd.Flags().StringVarP(&addBranch, "branch", "b", "", "Default branch (auto-detected if not specified)")
	cmd.Flags().StringVarP(&addName, "name", "n", "", "Display name (defaults to repo name)")

	cmd.MarkFlagRequired("repo")
	cmd.MarkFlagRequired("path")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Validate repo format
	if !strings.Contains(addRepo, "/") {
		return fmt.Errorf("repo must be in 'owner/name' format")
	}

	// Expand and validate path
	localPath := addPath
	if strings.HasPrefix(localPath, "~/") {
		home, _ := os.UserHomeDir()
		localPath = filepath.Join(home, localPath[2:])
	}
	localPath, _ = filepath.Abs(localPath)

	// Check if path exists and is a git repo
	if !git.PathExists(localPath) {
		return fmt.Errorf("path does not exist: %s", localPath)
	}
	if !git.IsGitRepo(localPath) {
		return fmt.Errorf("path is not a git repository: %s", localPath)
	}

	// Check GitHub access
	ghClient := github.NewClient()
	if !ghClient.RepoExists(addRepo) {
		return fmt.Errorf("cannot access repository: %s", addRepo)
	}

	// Auto-detect branch if not specified
	branch := addBranch
	if branch == "" {
		var err error
		branch, err = git.GetDefaultBranch(localPath)
		if err != nil {
			branch = "main" // fallback
		}
	}

	// Generate name if not specified
	name := addName
	if name == "" {
		parts := strings.Split(addRepo, "/")
		name = parts[len(parts)-1]
	}

	// Load existing config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		// If config doesn't exist, create default
		cfg = config.DefaultConfig()
	}

	// Check if repo already exists
	for _, cb := range cfg.Codebases {
		if cb.Repo == addRepo {
			return fmt.Errorf("repository already configured: %s", addRepo)
		}
	}

	// Add codebase
	cfg.Codebases = append(cfg.Codebases, config.Codebase{
		Name:          name,
		Repo:          addRepo,
		LocalPath:     localPath,
		DefaultBranch: branch,
		Enabled:       true,
	})

	// Save config
	configPath := cfgFile
	if configPath == "" {
		configPath = config.ConfigFilePath()
	}
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added codebase '%s' (%s)\n", name, addRepo)
	fmt.Printf("  Path: %s\n", localPath)
	fmt.Printf("  Branch: %s\n", branch)

	// Offer to sync labels
	fmt.Println("\nRun 'dev-swarm sync-labels' to create labels in the repository.")

	return nil
}
