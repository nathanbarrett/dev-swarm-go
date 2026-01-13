package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/pkg/version"
)

var (
	cfgFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dev-swarm",
		Short: "AI-powered development orchestration",
		Long: `dev-swarm monitors GitHub issues and automatically spawns Claude Code
sessions to plan, implement, and fix code based on label-driven workflows.

Use GitHub Issues as your control plane - just add labels to trigger AI actions.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ~/.config/dev-swarm/config.yaml)")

	// Add commands
	rootCmd.AddCommand(
		newStartCmd(),
		newInitCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newListCmd(),
		newSyncLabelsCmd(),
		newStatusCmd(),
		newLogsCmd(),
		newStopCmd(),
		newCleanupCmd(),
		newVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dev-swarm %s\n", version.Full())
		},
	}
}
