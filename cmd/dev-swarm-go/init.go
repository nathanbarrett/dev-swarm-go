package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/config"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long: `Initialize the dev-swarm configuration file.

Creates ~/.config/dev-swarm/config.yaml with default settings.`,
		RunE: runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := config.ConfigFilePath()

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists at %s\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Create default config
	if err := config.WriteDefaultConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Created config file at %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the config file to add your repositories")
	fmt.Println("2. Run 'dev-swarm add --repo owner/repo --path ~/code/repo' to add a codebase")
	fmt.Println("3. Run 'dev-swarm start' to begin monitoring")

	return nil
}
