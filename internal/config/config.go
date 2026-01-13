package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	apperrors "github.com/nathanbarrett/dev-swarm-go/internal/errors"
)

const (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = "~/.config/dev-swarm"
	// DefaultConfigFile is the default configuration filename
	DefaultConfigFile = "config.yaml"
)

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	return expandPath(DefaultConfigDir)
}

// ConfigFilePath returns the full path to the config file
func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), DefaultConfigFile)
}

// WorktreesDir returns the worktrees directory path
func WorktreesDir() string {
	return filepath.Join(ConfigDir(), "worktrees")
}

// LockFilePath returns the lock file path
func LockFilePath() string {
	return filepath.Join(ConfigDir(), "dev-swarm.lock")
}

// LogFilePath returns the log file path
func LogFilePath() string {
	return filepath.Join(ConfigDir(), "dev-swarm.log")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	if path == "" {
		path = ConfigFilePath()
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, apperrors.ErrConfigNotFound
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Start with defaults
	cfg := DefaultConfig()

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrConfigInvalid, err)
	}

	// Apply defaults for missing values
	applyDefaults(cfg)

	// Expand paths
	expandPaths(cfg)

	// Merge per-codebase labels with global labels
	mergeLabelOverrides(cfg)

	// Validate
	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyDefaults fills in missing values with defaults
func applyDefaults(cfg *Config) {
	defaults := DefaultSettings()

	if cfg.Settings.PollInterval == 0 {
		cfg.Settings.PollInterval = defaults.PollInterval
	}
	if cfg.Settings.ActivePollInterval == 0 {
		cfg.Settings.ActivePollInterval = defaults.ActivePollInterval
	}
	if cfg.Settings.MaxConcurrentSessions == 0 {
		cfg.Settings.MaxConcurrentSessions = defaults.MaxConcurrentSessions
	}
	if len(cfg.Settings.ApprovalKeywords) == 0 {
		cfg.Settings.ApprovalKeywords = defaults.ApprovalKeywords
	}
	if cfg.Settings.OutputBufferLines == 0 {
		cfg.Settings.OutputBufferLines = defaults.OutputBufferLines
	}
}

// expandPaths expands ~ in all path configurations
func expandPaths(cfg *Config) {
	for i := range cfg.Codebases {
		cfg.Codebases[i].LocalPath = expandPath(cfg.Codebases[i].LocalPath)
	}
}

// mergeLabelOverrides merges per-codebase label configs with global labels
func mergeLabelOverrides(cfg *Config) {
	for i := range cfg.Codebases {
		if cfg.Codebases[i].Labels != nil {
			// Merge each label that has an override
			merged := cfg.Labels // Start with global
			override := cfg.Codebases[i].Labels

			if override.ReadyToPlan.Name != "" {
				mergeLabelConfig(&merged.ReadyToPlan, &override.ReadyToPlan)
			}
			if override.PlanReview.Name != "" {
				mergeLabelConfig(&merged.PlanReview, &override.PlanReview)
			}
			if override.ReadyToImplement.Name != "" {
				mergeLabelConfig(&merged.ReadyToImplement, &override.ReadyToImplement)
			}
			if override.CodeReview.Name != "" {
				mergeLabelConfig(&merged.CodeReview, &override.CodeReview)
			}
			if override.Blocked.Name != "" {
				mergeLabelConfig(&merged.Blocked, &override.Blocked)
			}
			if override.Planning.Name != "" {
				mergeLabelConfig(&merged.Planning, &override.Planning)
			}
			if override.Implementing.Name != "" {
				mergeLabelConfig(&merged.Implementing, &override.Implementing)
			}
			if override.CIFailed.Name != "" {
				mergeLabelConfig(&merged.CIFailed, &override.CIFailed)
			}
			if override.Done.Name != "" {
				mergeLabelConfig(&merged.Done, &override.Done)
			}

			cfg.Codebases[i].Labels = &merged
		} else {
			// No overrides, use global labels
			labels := cfg.Labels
			cfg.Codebases[i].Labels = &labels
		}
	}
}

// mergeLabelConfig merges an override into a base label config
func mergeLabelConfig(base, override *LabelConfig) {
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Color != "" {
		base.Color = override.Color
	}
	if override.Description != "" {
		base.Description = override.Description
	}
	if override.Owner != "" {
		base.Owner = override.Owner
	}
	if override.AIPickup != "" {
		base.AIPickup = override.AIPickup
	}
	if override.AIAction != "" {
		base.AIAction = override.AIAction
	}
}

// Validate checks the configuration for errors
func Validate(cfg *Config) error {
	// Validate settings
	if cfg.Settings.PollInterval < 1 {
		return &apperrors.ConfigError{Field: "settings.poll_interval", Message: "must be at least 1"}
	}
	if cfg.Settings.ActivePollInterval < 1 {
		return &apperrors.ConfigError{Field: "settings.active_poll_interval", Message: "must be at least 1"}
	}
	if cfg.Settings.MaxConcurrentSessions < 1 {
		return &apperrors.ConfigError{Field: "settings.max_concurrent_sessions", Message: "must be at least 1"}
	}

	// Validate codebases
	for i, cb := range cfg.Codebases {
		if cb.Repo == "" {
			return &apperrors.ConfigError{
				Field:   fmt.Sprintf("codebases[%d].repo", i),
				Message: "is required",
			}
		}
		if !strings.Contains(cb.Repo, "/") {
			return &apperrors.ConfigError{
				Field:   fmt.Sprintf("codebases[%d].repo", i),
				Message: "must be in 'owner/name' format",
			}
		}
		if cb.LocalPath == "" {
			return &apperrors.ConfigError{
				Field:   fmt.Sprintf("codebases[%d].local_path", i),
				Message: "is required",
			}
		}
		if cb.DefaultBranch == "" {
			return &apperrors.ConfigError{
				Field:   fmt.Sprintf("codebases[%d].default_branch", i),
				Message: "is required",
			}
		}
	}

	return nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	dir := ConfigDir()
	return os.MkdirAll(dir, 0755)
}

// EnsureWorktreesDir creates the worktrees directory if it doesn't exist
func EnsureWorktreesDir() error {
	dir := WorktreesDir()
	return os.MkdirAll(dir, 0755)
}

// WriteDefaultConfig writes a default configuration file
func WriteDefaultConfig() error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := DefaultConfig()

	// Add example codebase
	cfg.Codebases = append(cfg.Codebases, Codebase{
		Name:          "example",
		Repo:          "owner/example-repo",
		LocalPath:     "~/code/example-repo",
		DefaultBranch: "main",
		Enabled:       false, // Disabled by default
	})

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := ConfigFilePath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Save writes the configuration to a file
func Save(cfg *Config, path string) error {
	if path == "" {
		path = ConfigFilePath()
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetEnabledCodebases returns only enabled codebases
func (cfg *Config) GetEnabledCodebases() []Codebase {
	var enabled []Codebase
	for _, cb := range cfg.Codebases {
		if cb.Enabled {
			enabled = append(enabled, cb)
		}
	}
	return enabled
}

// GetCodebaseByName returns a codebase by name
func (cfg *Config) GetCodebaseByName(name string) *Codebase {
	for i := range cfg.Codebases {
		if cfg.Codebases[i].Name == name {
			return &cfg.Codebases[i]
		}
	}
	return nil
}

// GetCodebaseByRepo returns a codebase by repo
func (cfg *Config) GetCodebaseByRepo(repo string) *Codebase {
	for i := range cfg.Codebases {
		if cfg.Codebases[i].Repo == repo {
			return &cfg.Codebases[i]
		}
	}
	return nil
}

// IsApprovalKeyword checks if a string contains an approval keyword
func (cfg *Config) IsApprovalKeyword(text string) bool {
	lower := strings.ToLower(text)
	for _, keyword := range cfg.Settings.ApprovalKeywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
