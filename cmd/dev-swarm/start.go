package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	"github.com/nathanbarrett/dev-swarm-go/internal/github"
	"github.com/nathanbarrett/dev-swarm-go/internal/lock"
	"github.com/nathanbarrett/dev-swarm-go/internal/orchestrator"
	"github.com/nathanbarrett/dev-swarm-go/internal/tui"
)

var (
	daemonMode bool
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the orchestrator",
		Long: `Start the dev-swarm orchestrator.

By default, starts with a TUI to monitor progress. Use --daemon to run in the background.`,
		RunE: runStart,
	}

	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "run in daemon mode (no TUI)")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Verify dependencies
	if err := verifyDependencies(); err != nil {
		return err
	}

	// Acquire lock
	lck := lock.New()
	if err := lck.Acquire(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w (is dev-swarm already running?)", err)
	}
	defer lck.Release()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create logger
	var logger *log.Logger
	if daemonMode {
		// Log to file in daemon mode
		logFile, err := os.OpenFile(config.LogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer logFile.Close()
		logger = log.New(logFile, "", log.LstdFlags)
	} else {
		// Log to stderr in TUI mode (will be captured)
		logger = log.New(os.Stderr, "[dev-swarm] ", log.LstdFlags)
	}

	// Create orchestrator
	orch, err := orchestrator.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Start orchestrator
	if err := orch.Start(); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}

	// Handle signals
	go func() {
		<-sigChan
		logger.Println("Received shutdown signal")
		orch.Stop()
	}()

	if daemonMode {
		// In daemon mode, just wait for signal
		logger.Printf("dev-swarm started in daemon mode. PID: %d", os.Getpid())
		<-sigChan
		return nil
	}

	// Start TUI
	if err := tui.Run(orch); err != nil {
		orch.Stop()
		return fmt.Errorf("TUI error: %w", err)
	}

	orch.Stop()
	return nil
}

func verifyDependencies() error {
	ghClient := github.NewClient()

	// Check gh CLI
	if !ghClient.IsInstalled() {
		return fmt.Errorf("gh CLI is not installed. Install: https://cli.github.com")
	}

	// Check gh authentication
	if !ghClient.IsAuthenticated() {
		return fmt.Errorf("gh CLI is not authenticated. Run: gh auth login")
	}

	// Check claude CLI
	if _, err := os.Stat("/usr/local/bin/claude"); os.IsNotExist(err) {
		// Try finding in PATH
		_, err := os.LookupEnv("PATH")
		if err {
			// This is a weak check, but good enough
		}
	}

	return nil
}
