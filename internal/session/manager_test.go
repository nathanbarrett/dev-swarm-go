package session

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager(5, 1000, "/tmp/worktrees")

	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.maxActive != 5 {
		t.Errorf("maxActive = %d, want 5", m.maxActive)
	}
	if m.outputBufferLines != 1000 {
		t.Errorf("outputBufferLines = %d, want 1000", m.outputBufferLines)
	}
	if m.worktreesDir != "/tmp/worktrees" {
		t.Errorf("worktreesDir = %q, want %q", m.worktreesDir, "/tmp/worktrees")
	}
	if m.sessions == nil {
		t.Error("sessions map should be initialized")
	}
	if m.outputChan == nil {
		t.Error("outputChan should be initialized")
	}
	if m.statusChan == nil {
		t.Error("statusChan should be initialized")
	}
}

func TestManagerCanSpawn(t *testing.T) {
	m := NewManager(2, 100, "/tmp")

	// Initially should be able to spawn
	if !m.CanSpawn() {
		t.Error("CanSpawn should return true when no sessions exist")
	}

	// Add a pending session (shouldn't count toward limit)
	m.mu.Lock()
	m.sessions["test#1"] = &Session{Status: StatusPending}
	m.mu.Unlock()

	if !m.CanSpawn() {
		t.Error("CanSpawn should return true with only pending sessions")
	}

	// Add a running session
	m.mu.Lock()
	m.sessions["test#2"] = &Session{Status: StatusRunning}
	m.mu.Unlock()

	if !m.CanSpawn() {
		t.Error("CanSpawn should return true with 1 running (max=2)")
	}

	// Add another running session
	m.mu.Lock()
	m.sessions["test#3"] = &Session{Status: StatusRunning}
	m.mu.Unlock()

	if m.CanSpawn() {
		t.Error("CanSpawn should return false at max capacity")
	}

	// Change one to completed
	m.mu.Lock()
	m.sessions["test#2"].Status = StatusCompleted
	m.mu.Unlock()

	if !m.CanSpawn() {
		t.Error("CanSpawn should return true after one completes")
	}
}

func TestManagerHasSession(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Should not have session initially
	if m.HasSession("test#1") {
		t.Error("HasSession should return false when session doesn't exist")
	}

	// Add session
	m.mu.Lock()
	m.sessions["test#1"] = &Session{}
	m.mu.Unlock()

	// Should have session now
	if !m.HasSession("test#1") {
		t.Error("HasSession should return true after adding session")
	}

	// Different ID should not exist
	if m.HasSession("test#2") {
		t.Error("HasSession should return false for different ID")
	}
}

func TestManagerActiveCount(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Initially zero
	if m.ActiveCount() != 0 {
		t.Errorf("ActiveCount = %d, want 0", m.ActiveCount())
	}

	// Add sessions with various statuses
	m.mu.Lock()
	m.sessions["test#1"] = &Session{Status: StatusPending}
	m.sessions["test#2"] = &Session{Status: StatusRunning}
	m.sessions["test#3"] = &Session{Status: StatusRunning}
	m.sessions["test#4"] = &Session{Status: StatusCompleted}
	m.sessions["test#5"] = &Session{Status: StatusFailed}
	m.mu.Unlock()

	// Should only count running
	if m.ActiveCount() != 2 {
		t.Errorf("ActiveCount = %d, want 2", m.ActiveCount())
	}
}

func TestManagerGetSession(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Add session
	expected := &Session{ID: "test#1"}
	m.mu.Lock()
	m.sessions["test#1"] = expected
	m.mu.Unlock()

	// Should get session
	got := m.GetSession("test#1")
	if got != expected {
		t.Error("GetSession should return the session")
	}

	// Nonexistent should return nil
	if m.GetSession("nonexistent") != nil {
		t.Error("GetSession should return nil for nonexistent ID")
	}
}

func TestManagerGetAllSessions(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Initially empty
	all := m.GetAllSessions()
	if len(all) != 0 {
		t.Errorf("len(GetAllSessions) = %d, want 0", len(all))
	}

	// Add sessions
	m.mu.Lock()
	m.sessions["test#1"] = &Session{ID: "test#1"}
	m.sessions["test#2"] = &Session{ID: "test#2"}
	m.mu.Unlock()

	all = m.GetAllSessions()
	if len(all) != 2 {
		t.Errorf("len(GetAllSessions) = %d, want 2", len(all))
	}
}

func TestManagerGetSessionForIssue(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Add session
	expected := &Session{ID: "owner/repo#42"}
	m.mu.Lock()
	m.sessions["owner/repo#42"] = expected
	m.mu.Unlock()

	// Should find by repo and issue number
	got := m.GetSessionForIssue("owner/repo", 42)
	if got != expected {
		t.Error("GetSessionForIssue should return the session")
	}

	// Different issue should return nil
	if m.GetSessionForIssue("owner/repo", 99) != nil {
		t.Error("GetSessionForIssue should return nil for different issue")
	}

	// Different repo should return nil
	if m.GetSessionForIssue("other/repo", 42) != nil {
		t.Error("GetSessionForIssue should return nil for different repo")
	}
}

func TestManagerRemoveSession(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Add session
	m.mu.Lock()
	m.sessions["test#1"] = &Session{}
	m.mu.Unlock()

	if !m.HasSession("test#1") {
		t.Fatal("Session should exist before remove")
	}

	// Remove session
	m.RemoveSession("test#1")

	if m.HasSession("test#1") {
		t.Error("Session should not exist after remove")
	}

	// Removing nonexistent should not panic
	m.RemoveSession("nonexistent")
}

func TestManagerCleanupCompleted(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	// Add sessions with various statuses
	m.mu.Lock()
	m.sessions["test#1"] = &Session{Status: StatusRunning}
	m.sessions["test#2"] = &Session{Status: StatusCompleted}
	m.sessions["test#3"] = &Session{Status: StatusFailed}
	m.sessions["test#4"] = &Session{Status: StatusPending}
	m.mu.Unlock()

	// Cleanup
	m.CleanupCompleted()

	// Running and pending should remain
	if !m.HasSession("test#1") {
		t.Error("Running session should remain")
	}
	if !m.HasSession("test#4") {
		t.Error("Pending session should remain")
	}

	// Completed and failed should be removed
	if m.HasSession("test#2") {
		t.Error("Completed session should be removed")
	}
	if m.HasSession("test#3") {
		t.Error("Failed session should be removed")
	}
}

func TestManagerOutputChan(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	ch := m.OutputChan()
	if ch == nil {
		t.Error("OutputChan should return non-nil channel")
	}
}

func TestManagerStatusChan(t *testing.T) {
	m := NewManager(5, 100, "/tmp")

	ch := m.StatusChan()
	if ch == nil {
		t.Error("StatusChan should return non-nil channel")
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{Status(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
