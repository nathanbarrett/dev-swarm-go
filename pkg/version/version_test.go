package version

import (
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	info := Info()

	// Should return the Version variable
	if info != Version {
		t.Errorf("Info() = %q, want %q", info, Version)
	}

	// Should not be empty
	if info == "" {
		t.Error("Info() should not be empty")
	}
}

func TestFull(t *testing.T) {
	full := Full()

	// Should contain version
	if !strings.Contains(full, Version) {
		t.Errorf("Full() = %q, should contain Version %q", full, Version)
	}

	// Should contain commit
	if !strings.Contains(full, Commit) {
		t.Errorf("Full() = %q, should contain Commit %q", full, Commit)
	}

	// Should contain date
	if !strings.Contains(full, Date) {
		t.Errorf("Full() = %q, should contain Date %q", full, Date)
	}

	// Should have expected format
	expected := Version + " (" + Commit + ") built on " + Date
	if full != expected {
		t.Errorf("Full() = %q, want %q", full, expected)
	}
}

func TestVersionVariables(t *testing.T) {
	// These are compile-time variables, so we test their defaults
	// In production, they would be set via ldflags

	// Version should have a default
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Commit should have a default
	if Commit == "" {
		t.Error("Commit should not be empty")
	}

	// Date should have a default
	if Date == "" {
		t.Error("Date should not be empty")
	}
}
