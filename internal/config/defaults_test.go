package config

import (
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	if settings.PollInterval != 60 {
		t.Errorf("PollInterval = %d, want 60", settings.PollInterval)
	}
	if settings.ActivePollInterval != 10 {
		t.Errorf("ActivePollInterval = %d, want 10", settings.ActivePollInterval)
	}
	if settings.MaxConcurrentSessions != 5 {
		t.Errorf("MaxConcurrentSessions = %d, want 5", settings.MaxConcurrentSessions)
	}
	if !settings.AutoMergeOnApproval {
		t.Error("AutoMergeOnApproval should be true by default")
	}
	if settings.OutputBufferLines != 1000 {
		t.Errorf("OutputBufferLines = %d, want 1000", settings.OutputBufferLines)
	}

	expectedKeywords := []string{"approved", "lgtm", "ship it", "merge it", "looks good"}
	if len(settings.ApprovalKeywords) != len(expectedKeywords) {
		t.Errorf("len(ApprovalKeywords) = %d, want %d", len(settings.ApprovalKeywords), len(expectedKeywords))
	}
	for i, kw := range expectedKeywords {
		if settings.ApprovalKeywords[i] != kw {
			t.Errorf("ApprovalKeywords[%d] = %q, want %q", i, settings.ApprovalKeywords[i], kw)
		}
	}
}

func TestDefaultLabels(t *testing.T) {
	labels := DefaultLabels()

	// Test ReadyToPlan
	if labels.ReadyToPlan.Name != "user:ready-to-plan" {
		t.Errorf("ReadyToPlan.Name = %q, want %q", labels.ReadyToPlan.Name, "user:ready-to-plan")
	}
	if labels.ReadyToPlan.Color != "0052CC" {
		t.Errorf("ReadyToPlan.Color = %q, want %q", labels.ReadyToPlan.Color, "0052CC")
	}
	if labels.ReadyToPlan.Owner != "user" {
		t.Errorf("ReadyToPlan.Owner = %q, want %q", labels.ReadyToPlan.Owner, "user")
	}
	if labels.ReadyToPlan.AIPickup != "always" {
		t.Errorf("ReadyToPlan.AIPickup = %q, want %q", labels.ReadyToPlan.AIPickup, "always")
	}

	// Test Planning (AI-owned)
	if labels.Planning.Name != "ai:planning" {
		t.Errorf("Planning.Name = %q, want %q", labels.Planning.Name, "ai:planning")
	}
	if labels.Planning.Owner != "ai" {
		t.Errorf("Planning.Owner = %q, want %q", labels.Planning.Owner, "ai")
	}
	if labels.Planning.AIPickup != "never" {
		t.Errorf("Planning.AIPickup = %q, want %q", labels.Planning.AIPickup, "never")
	}

	// Test PlanReview (conditional pickup)
	if labels.PlanReview.AIPickup != "on_user_comment" {
		t.Errorf("PlanReview.AIPickup = %q, want %q", labels.PlanReview.AIPickup, "on_user_comment")
	}

	// Test CIFailed
	if labels.CIFailed.Name != "ai:ci-failed" {
		t.Errorf("CIFailed.Name = %q, want %q", labels.CIFailed.Name, "ai:ci-failed")
	}
	if labels.CIFailed.Color != "D93F0B" {
		t.Errorf("CIFailed.Color = %q, want %q", labels.CIFailed.Color, "D93F0B")
	}
	if labels.CIFailed.AIPickup != "always" {
		t.Errorf("CIFailed.AIPickup = %q, want %q", labels.CIFailed.AIPickup, "always")
	}

	// Test Done
	if labels.Done.Name != "ai:done" {
		t.Errorf("Done.Name = %q, want %q", labels.Done.Name, "ai:done")
	}
	if labels.Done.Color != "0E8A16" {
		t.Errorf("Done.Color = %q, want %q", labels.Done.Color, "0E8A16")
	}
}

func TestDefaultAIInstructions(t *testing.T) {
	instructions := DefaultAIInstructions()

	if instructions.General == "" {
		t.Error("General instructions should not be empty")
	}

	// Check for key content
	if len(instructions.General) < 100 {
		t.Error("General instructions seem too short")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Should have default settings
	if cfg.Settings.PollInterval != 60 {
		t.Error("DefaultConfig should have default settings")
	}

	// Should have default labels
	if cfg.Labels.ReadyToPlan.Name != "user:ready-to-plan" {
		t.Error("DefaultConfig should have default labels")
	}

	// Should have AI instructions
	if cfg.AIInstructions.General == "" {
		t.Error("DefaultConfig should have AI instructions")
	}

	// Should have empty codebases
	if len(cfg.Codebases) != 0 {
		t.Errorf("DefaultConfig should have empty codebases, got %d", len(cfg.Codebases))
	}
}
