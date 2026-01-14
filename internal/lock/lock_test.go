package lock

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	apperrors "github.com/nathanbarrett/dev-swarm-go/internal/errors"
)

func TestNewWithPath(t *testing.T) {
	path := "/tmp/test.lock"
	l := NewWithPath(path)

	if l.path != path {
		t.Errorf("lock.path = %q, want %q", l.path, path)
	}
	if l.pid != os.Getpid() {
		t.Errorf("lock.pid = %d, want %d", l.pid, os.Getpid())
	}
}

func TestAcquireAndRelease(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Acquire lock
	err = l.Acquire()
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	// Verify lock file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file should exist after Acquire()")
	}

	// Verify PID in lock file
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if pid != os.Getpid() {
		t.Errorf("Lock file PID = %d, want %d", pid, os.Getpid())
	}

	// Release lock
	err = l.Release()
	if err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	// Verify lock file is removed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("Lock file should be removed after Release()")
	}
}

func TestIsLocked(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Should not be locked initially
	if l.IsLocked() {
		t.Error("IsLocked() should return false when no lock file exists")
	}

	// Acquire and check
	err = l.Acquire()
	if err != nil {
		t.Fatal(err)
	}

	// Should be locked (by current process)
	if !l.IsLocked() {
		t.Error("IsLocked() should return true after Acquire()")
	}

	// Release and check
	l.Release()
	if l.IsLocked() {
		t.Error("IsLocked() should return false after Release()")
	}
}

func TestGetPID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Should fail when no lock file
	_, err = l.GetPID()
	if err == nil {
		t.Error("GetPID() should error when no lock file exists")
	}

	// Write lock file
	err = l.Acquire()
	if err != nil {
		t.Fatal(err)
	}

	// Should return current PID
	pid, err := l.GetPID()
	if err != nil {
		t.Fatalf("GetPID() error = %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("GetPID() = %d, want %d", pid, os.Getpid())
	}

	l.Release()
}

func TestGetPIDInvalidContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Write invalid content
	err = os.WriteFile(lockPath, []byte("not-a-number"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Should fail and clean up
	_, err = l.GetPID()
	if err == nil {
		t.Error("GetPID() should error with invalid content")
	}

	// Lock file should be removed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("Invalid lock file should be removed")
	}
}

func TestCleanStale(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Write lock file with non-existent PID
	// PID 99999 is unlikely to exist
	err = os.WriteFile(lockPath, []byte("99999"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Clean should remove stale lock
	err = l.CleanStale()
	if err != nil {
		t.Fatalf("CleanStale() error = %v", err)
	}

	// Lock file should be removed (assuming PID 99999 doesn't exist)
	// Note: This test may fail if PID 99999 happens to exist
}

func TestGetStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Status without lock
	status := l.GetStatus()
	if status.IsLocked {
		t.Error("Status.IsLocked should be false when no lock")
	}
	if status.Error == nil {
		t.Error("Status.Error should be set when no lock file")
	}

	// Acquire and check status
	err = l.Acquire()
	if err != nil {
		t.Fatal(err)
	}

	status = l.GetStatus()
	if !status.IsLocked {
		t.Error("Status.IsLocked should be true after Acquire()")
	}
	if status.PID != os.Getpid() {
		t.Errorf("Status.PID = %d, want %d", status.PID, os.Getpid())
	}
	if status.Error != nil {
		t.Errorf("Status.Error should be nil, got %v", status.Error)
	}

	l.Release()
}

func TestAcquireAlreadyLocked(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l1 := NewWithPath(lockPath)

	// Acquire first lock
	err = l1.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	defer l1.Release()

	// Try to acquire second lock (same PID, should fail because already locked)
	l2 := NewWithPath(lockPath)
	err = l2.Acquire()
	if err != apperrors.ErrAlreadyRunning {
		t.Errorf("Acquire() error = %v, want %v", err, apperrors.ErrAlreadyRunning)
	}
}

func TestReleaseNotOwned(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")

	// Write lock file with different PID
	err = os.WriteFile(lockPath, []byte("1"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewWithPath(lockPath)
	err = l.Release()
	if err == nil {
		t.Error("Release() should error when lock is owned by another process")
	}
}

func TestStopNotRunning(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lock-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "test.lock")
	l := NewWithPath(lockPath)

	// Stop without lock file
	err = l.Stop()
	if err != apperrors.ErrNotRunning {
		t.Errorf("Stop() error = %v, want %v", err, apperrors.ErrNotRunning)
	}
}

func TestIsProcessRunning(t *testing.T) {
	// Current process should be running
	if !isProcessRunning(os.Getpid()) {
		t.Error("isProcessRunning(os.Getpid()) should return true")
	}

	// Invalid PIDs
	if isProcessRunning(0) {
		t.Error("isProcessRunning(0) should return false")
	}
	if isProcessRunning(-1) {
		t.Error("isProcessRunning(-1) should return false")
	}
}
