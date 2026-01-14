package config

import (
	"os"
	"path/filepath"
	"testing"

	apperrors "github.com/nathanbarrett/dev-swarm-go/internal/errors"
)

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expands tilde",
			input:    "~/test/path",
			expected: filepath.Join(home, "test/path"),
		},
		{
			name:     "no tilde unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "tilde in middle unchanged",
			input:    "/path/~/test",
			expected: "/path/~/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 5,
				},
				Codebases: []Codebase{
					{
						Repo:          "owner/repo",
						LocalPath:     "/path/to/repo",
						DefaultBranch: "main",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid poll interval",
			config: &Config{
				Settings: Settings{
					PollInterval:          0,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 5,
				},
			},
			wantErr: true,
			errMsg:  "poll_interval",
		},
		{
			name: "invalid active poll interval",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    0,
					MaxConcurrentSessions: 5,
				},
			},
			wantErr: true,
			errMsg:  "active_poll_interval",
		},
		{
			name: "invalid max concurrent sessions",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 0,
				},
			},
			wantErr: true,
			errMsg:  "max_concurrent_sessions",
		},
		{
			name: "missing repo",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 5,
				},
				Codebases: []Codebase{
					{
						LocalPath:     "/path",
						DefaultBranch: "main",
					},
				},
			},
			wantErr: true,
			errMsg:  "repo",
		},
		{
			name: "invalid repo format",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 5,
				},
				Codebases: []Codebase{
					{
						Repo:          "invalid-format",
						LocalPath:     "/path",
						DefaultBranch: "main",
					},
				},
			},
			wantErr: true,
			errMsg:  "owner/name",
		},
		{
			name: "missing local path",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 5,
				},
				Codebases: []Codebase{
					{
						Repo:          "owner/repo",
						DefaultBranch: "main",
					},
				},
			},
			wantErr: true,
			errMsg:  "local_path",
		},
		{
			name: "missing default branch",
			config: &Config{
				Settings: Settings{
					PollInterval:          60,
					ActivePollInterval:    10,
					MaxConcurrentSessions: 5,
				},
				Codebases: []Codebase{
					{
						Repo:      "owner/repo",
						LocalPath: "/path",
					},
				},
			},
			wantErr: true,
			errMsg:  "default_branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if configErr, ok := err.(*apperrors.ConfigError); ok {
					if tt.errMsg != "" && configErr.Field == "" {
						t.Errorf("Expected error field containing %q", tt.errMsg)
					}
				}
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	defaults := DefaultSettings()

	if cfg.Settings.PollInterval != defaults.PollInterval {
		t.Errorf("PollInterval = %d, want %d", cfg.Settings.PollInterval, defaults.PollInterval)
	}
	if cfg.Settings.ActivePollInterval != defaults.ActivePollInterval {
		t.Errorf("ActivePollInterval = %d, want %d", cfg.Settings.ActivePollInterval, defaults.ActivePollInterval)
	}
	if cfg.Settings.MaxConcurrentSessions != defaults.MaxConcurrentSessions {
		t.Errorf("MaxConcurrentSessions = %d, want %d", cfg.Settings.MaxConcurrentSessions, defaults.MaxConcurrentSessions)
	}
	if cfg.Settings.OutputBufferLines != defaults.OutputBufferLines {
		t.Errorf("OutputBufferLines = %d, want %d", cfg.Settings.OutputBufferLines, defaults.OutputBufferLines)
	}
}

func TestExpandPaths(t *testing.T) {
	home, _ := os.UserHomeDir()

	cfg := &Config{
		Codebases: []Codebase{
			{LocalPath: "~/code/repo1"},
			{LocalPath: "/absolute/path"},
		},
	}

	expandPaths(cfg)

	expected0 := filepath.Join(home, "code/repo1")
	if cfg.Codebases[0].LocalPath != expected0 {
		t.Errorf("Codebase[0].LocalPath = %q, want %q", cfg.Codebases[0].LocalPath, expected0)
	}
	if cfg.Codebases[1].LocalPath != "/absolute/path" {
		t.Errorf("Codebase[1].LocalPath = %q, want %q", cfg.Codebases[1].LocalPath, "/absolute/path")
	}
}

func TestMergeLabelConfig(t *testing.T) {
	base := &LabelConfig{
		Name:        "original-name",
		Color:       "FF0000",
		Description: "original desc",
		Owner:       "user",
		AIPickup:    "always",
		AIAction:    "original action",
	}

	override := &LabelConfig{
		Name:  "new-name",
		Color: "00FF00",
	}

	mergeLabelConfig(base, override)

	if base.Name != "new-name" {
		t.Errorf("Name = %q, want %q", base.Name, "new-name")
	}
	if base.Color != "00FF00" {
		t.Errorf("Color = %q, want %q", base.Color, "00FF00")
	}
	// These should remain unchanged
	if base.Description != "original desc" {
		t.Errorf("Description = %q, want %q", base.Description, "original desc")
	}
	if base.Owner != "user" {
		t.Errorf("Owner = %q, want %q", base.Owner, "user")
	}
}

func TestGetEnabledCodebases(t *testing.T) {
	cfg := &Config{
		Codebases: []Codebase{
			{Name: "enabled1", Enabled: true},
			{Name: "disabled", Enabled: false},
			{Name: "enabled2", Enabled: true},
		},
	}

	enabled := cfg.GetEnabledCodebases()

	if len(enabled) != 2 {
		t.Errorf("len(enabled) = %d, want 2", len(enabled))
	}
	if enabled[0].Name != "enabled1" {
		t.Errorf("enabled[0].Name = %q, want %q", enabled[0].Name, "enabled1")
	}
	if enabled[1].Name != "enabled2" {
		t.Errorf("enabled[1].Name = %q, want %q", enabled[1].Name, "enabled2")
	}
}

func TestGetCodebaseByName(t *testing.T) {
	cfg := &Config{
		Codebases: []Codebase{
			{Name: "repo1", Repo: "owner/repo1"},
			{Name: "repo2", Repo: "owner/repo2"},
		},
	}

	cb := cfg.GetCodebaseByName("repo2")
	if cb == nil {
		t.Fatal("GetCodebaseByName returned nil")
	}
	if cb.Repo != "owner/repo2" {
		t.Errorf("cb.Repo = %q, want %q", cb.Repo, "owner/repo2")
	}

	cb = cfg.GetCodebaseByName("nonexistent")
	if cb != nil {
		t.Error("GetCodebaseByName should return nil for nonexistent name")
	}
}

func TestGetCodebaseByRepo(t *testing.T) {
	cfg := &Config{
		Codebases: []Codebase{
			{Name: "repo1", Repo: "owner/repo1"},
			{Name: "repo2", Repo: "owner/repo2"},
		},
	}

	cb := cfg.GetCodebaseByRepo("owner/repo2")
	if cb == nil {
		t.Fatal("GetCodebaseByRepo returned nil")
	}
	if cb.Name != "repo2" {
		t.Errorf("cb.Name = %q, want %q", cb.Name, "repo2")
	}

	cb = cfg.GetCodebaseByRepo("nonexistent/repo")
	if cb != nil {
		t.Error("GetCodebaseByRepo should return nil for nonexistent repo")
	}
}

func TestIsApprovalKeyword(t *testing.T) {
	cfg := &Config{
		Settings: Settings{
			ApprovalKeywords: []string{"approved", "lgtm", "ship it"},
		},
	}

	tests := []struct {
		text string
		want bool
	}{
		{"approved", true},
		{"APPROVED", true},
		{"Looks approved to me", true},
		{"lgtm", true},
		{"LGTM!", true},
		{"ship it", true},
		{"Ship It Now!", true},
		{"rejected", false},
		{"needs changes", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := cfg.IsApprovalKeyword(tt.text)
			if got != tt.want {
				t.Errorf("IsApprovalKeyword(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err != apperrors.ErrConfigNotFound {
		t.Errorf("Load() error = %v, want %v", err, apperrors.ErrConfigNotFound)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config
	cfg := DefaultConfig()
	cfg.Codebases = append(cfg.Codebases, Codebase{
		Name:          "test-repo",
		Repo:          "owner/test-repo",
		LocalPath:     "/tmp/test-repo",
		DefaultBranch: "main",
		Enabled:       true,
	})

	// Save
	err = Save(cfg, configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify
	if len(loaded.Codebases) != 1 {
		t.Errorf("len(Codebases) = %d, want 1", len(loaded.Codebases))
	}
	if loaded.Codebases[0].Name != "test-repo" {
		t.Errorf("Codebases[0].Name = %q, want %q", loaded.Codebases[0].Name, "test-repo")
	}
}
