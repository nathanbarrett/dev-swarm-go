package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/nathanbarrett/dev-swarm-go/internal/lock"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running orchestrator",
		RunE:  runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	lck := lock.New()

	status := lck.GetStatus()
	if !status.IsLocked {
		fmt.Println("dev-swarm is not running.")
		return nil
	}

	fmt.Printf("Stopping dev-swarm (PID: %d)...\n", status.PID)

	if err := lck.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	// Wait for process to stop
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		if !lck.IsLocked() {
			fmt.Println("Stopped.")
			return nil
		}
	}

	fmt.Println("Process may still be shutting down.")
	return nil
}
