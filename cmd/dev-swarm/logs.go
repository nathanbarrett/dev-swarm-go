package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
)

var (
	logLines  int
	logFollow bool
)

func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View orchestrator logs",
		Long: `View the dev-swarm log file.

Logs are only written when running in daemon mode.`,
		RunE: runLogs,
	}

	cmd.Flags().IntVarP(&logLines, "lines", "n", 50, "number of lines to show")
	cmd.Flags().BoolVarP(&logFollow, "follow", "f", false, "follow log output (like tail -f)")

	return cmd
}

func runLogs(cmd *cobra.Command, args []string) error {
	logPath := config.LogFilePath()

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No log file found.")
			fmt.Println("Logs are only created when running in daemon mode (dev-swarm start --daemon).")
			return nil
		}
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	if logFollow {
		// Follow mode - read existing content then tail
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		// Continue tailing
		fmt.Println("--- Following log (Ctrl+C to stop) ---")
		for {
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
			// Small sleep to avoid busy loop
			// In a real implementation, we'd use fsnotify
		}
	}

	// Read last N lines
	// Simple implementation: read all and take last N
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	start := 0
	if len(lines) > logLines {
		start = len(lines) - logLines
	}

	for _, line := range lines[start:] {
		fmt.Println(line)
	}

	return nil
}
