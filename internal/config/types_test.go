package config

import (
	"testing"
)

func TestLabelsGetAllLabels(t *testing.T) {
	labels := DefaultLabels()
	all := labels.GetAllLabels()

	if len(all) != 9 {
		t.Errorf("GetAllLabels() returned %d labels, want 9", len(all))
	}

	// Verify all labels are present
	expectedNames := []string{
		"user:ready-to-plan",
		"user:plan-review",
		"user:ready-to-implement",
		"user:code-review",
		"user:blocked",
		"ai:planning",
		"ai:implementing",
		"ai:ci-failed",
		"ai:done",
	}

	for i, name := range expectedNames {
		if all[i].Name != name {
			t.Errorf("GetAllLabels()[%d].Name = %q, want %q", i, all[i].Name, name)
		}
	}
}

func TestLabelsGetByName(t *testing.T) {
	labels := DefaultLabels()

	tests := []struct {
		name     string
		wantNil  bool
		wantName string
	}{
		{"user:ready-to-plan", false, "user:ready-to-plan"},
		{"ai:planning", false, "ai:planning"},
		{"ai:done", false, "ai:done"},
		{"nonexistent", true, ""},
		{"", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labels.GetByName(tt.name)
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetByName(%q) = %v, want nil", tt.name, got)
				}
			} else {
				if got == nil {
					t.Fatalf("GetByName(%q) = nil, want non-nil", tt.name)
				}
				if got.Name != tt.wantName {
					t.Errorf("GetByName(%q).Name = %q, want %q", tt.name, got.Name, tt.wantName)
				}
			}
		})
	}
}

func TestLabelsGetPickupLabels(t *testing.T) {
	labels := DefaultLabels()
	pickup := labels.GetPickupLabels()

	// Should return labels with "always" or "on_user_comment" pickup
	// Expected: ready-to-plan (always), plan-review (on_user_comment),
	// ready-to-implement (always), code-review (on_user_comment), ci-failed (always)
	expectedCount := 5
	if len(pickup) != expectedCount {
		t.Errorf("GetPickupLabels() returned %d labels, want %d", len(pickup), expectedCount)
	}

	// Verify none have "never" pickup
	for _, label := range pickup {
		if label.AIPickup == string(PickupNever) {
			t.Errorf("GetPickupLabels() returned label %q with AIPickup=never", label.Name)
		}
	}
}

func TestPickupRuleConstants(t *testing.T) {
	if PickupAlways != "always" {
		t.Errorf("PickupAlways = %q, want %q", PickupAlways, "always")
	}
	if PickupNever != "never" {
		t.Errorf("PickupNever = %q, want %q", PickupNever, "never")
	}
	if PickupOnUserComment != "on_user_comment" {
		t.Errorf("PickupOnUserComment = %q, want %q", PickupOnUserComment, "on_user_comment")
	}
}

func TestLabelOwnerConstants(t *testing.T) {
	if OwnerUser != "user" {
		t.Errorf("OwnerUser = %q, want %q", OwnerUser, "user")
	}
	if OwnerAI != "ai" {
		t.Errorf("OwnerAI = %q, want %q", OwnerAI, "ai")
	}
}
