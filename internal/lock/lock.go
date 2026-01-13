package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/nathanbarrett/dev-swarm-go/internal/config"
	apperrors "github.com/nathanbarrett/dev-swarm-go/internal/errors"
)

// Lock represents a process lock
type Lock struct {
	path string
	pid  int
}

// New creates a new lock
func New() *Lock {
	return &Lock{
		path: config.LockFilePath(),
		pid:  os.Getpid(),
	}
}

// NewWithPath creates a new lock with a custom path
func NewWithPath(path string) *Lock {
	return &Lock{
		path: path,
		pid:  os.Getpid(),
	}
}

// Acquire attempts to acquire the lock
func (l *Lock) Acquire() error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(l.path), 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Check if lock file exists
	if l.IsLocked() {
		return apperrors.ErrAlreadyRunning
	}

	// Write PID to lock file
	if err := os.WriteFile(l.path, []byte(strconv.Itoa(l.pid)), 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// Release releases the lock
func (l *Lock) Release() error {
	// Only remove if we own the lock
	existingPID, err := l.GetPID()
	if err != nil {
		return nil // Lock file doesn't exist
	}

	if existingPID != l.pid {
		return fmt.Errorf("lock is owned by another process (PID: %d)", existingPID)
	}

	return os.Remove(l.path)
}

// IsLocked checks if the lock is held by a running process
func (l *Lock) IsLocked() bool {
	pid, err := l.GetPID()
	if err != nil {
		return false
	}

	return isProcessRunning(pid)
}

// GetPID returns the PID from the lock file
func (l *Lock) GetPID() (int, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("lock file does not exist")
		}
		return 0, fmt.Errorf("failed to read lock file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		// Invalid PID in lock file, treat as stale
		os.Remove(l.path)
		return 0, fmt.Errorf("invalid PID in lock file")
	}

	return pid, nil
}

// CleanStale removes stale lock files
func (l *Lock) CleanStale() error {
	pid, err := l.GetPID()
	if err != nil {
		return nil // No lock file or invalid
	}

	if !isProcessRunning(pid) {
		return os.Remove(l.path)
	}

	return nil
}

// isProcessRunning checks if a process with the given PID is running
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0
	// to check if the process actually exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Status represents the lock status
type Status struct {
	IsLocked bool
	PID      int
	Error    error
}

// GetStatus returns the current lock status
func (l *Lock) GetStatus() Status {
	pid, err := l.GetPID()
	if err != nil {
		return Status{
			IsLocked: false,
			PID:      0,
			Error:    err,
		}
	}

	isRunning := isProcessRunning(pid)
	return Status{
		IsLocked: isRunning,
		PID:      pid,
		Error:    nil,
	}
}

// Stop sends a termination signal to the locked process
func (l *Lock) Stop() error {
	pid, err := l.GetPID()
	if err != nil {
		return apperrors.ErrNotRunning
	}

	if !isProcessRunning(pid) {
		// Clean up stale lock
		os.Remove(l.path)
		return apperrors.ErrNotRunning
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	return nil
}
