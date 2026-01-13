package config

// Config represents the complete dev-swarm configuration
type Config struct {
	Settings       Settings          `yaml:"settings"`
	Labels         Labels            `yaml:"labels"`
	AIInstructions AIInstructions    `yaml:"ai_instructions"`
	Codebases      []Codebase        `yaml:"codebases"`
}

// Settings contains global settings
type Settings struct {
	PollInterval          int      `yaml:"poll_interval"`
	ActivePollInterval    int      `yaml:"active_poll_interval"`
	MaxConcurrentSessions int      `yaml:"max_concurrent_sessions"`
	AutoMergeOnApproval   bool     `yaml:"auto_merge_on_approval"`
	ApprovalKeywords      []string `yaml:"approval_keywords"`
	OutputBufferLines     int      `yaml:"output_buffer_lines"`
}

// Labels contains all label configurations
type Labels struct {
	ReadyToPlan       LabelConfig `yaml:"ready_to_plan"`
	PlanReview        LabelConfig `yaml:"plan_review"`
	ReadyToImplement  LabelConfig `yaml:"ready_to_implement"`
	CodeReview        LabelConfig `yaml:"code_review"`
	Blocked           LabelConfig `yaml:"blocked"`
	Planning          LabelConfig `yaml:"planning"`
	Implementing      LabelConfig `yaml:"implementing"`
	CIFailed          LabelConfig `yaml:"ci_failed"`
	Done              LabelConfig `yaml:"done"`
}

// LabelConfig represents a single label configuration
type LabelConfig struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	Description string `yaml:"description"`
	Owner       string `yaml:"owner"`       // "user" or "ai"
	AIPickup    string `yaml:"ai_pickup"`   // "always", "never", "on_user_comment"
	AIAction    string `yaml:"ai_action"`   // Instructions for AI when this label is picked up
}

// AIInstructions contains global AI instructions
type AIInstructions struct {
	General string `yaml:"general"`
}

// Codebase represents a single repository configuration
type Codebase struct {
	Name          string  `yaml:"name"`
	Repo          string  `yaml:"repo"`           // "owner/repo" format
	LocalPath     string  `yaml:"local_path"`
	DefaultBranch string  `yaml:"default_branch"`
	Enabled       bool    `yaml:"enabled"`
	Labels        *Labels `yaml:"labels,omitempty"` // Per-codebase label overrides
}

// PickupRule defines when AI should pick up an issue
type PickupRule string

const (
	PickupAlways        PickupRule = "always"
	PickupNever         PickupRule = "never"
	PickupOnUserComment PickupRule = "on_user_comment"
)

// LabelOwner defines who owns the current state
type LabelOwner string

const (
	OwnerUser LabelOwner = "user"
	OwnerAI   LabelOwner = "ai"
)

// GetAllLabels returns all label configs as a slice for iteration
func (l *Labels) GetAllLabels() []LabelConfig {
	return []LabelConfig{
		l.ReadyToPlan,
		l.PlanReview,
		l.ReadyToImplement,
		l.CodeReview,
		l.Blocked,
		l.Planning,
		l.Implementing,
		l.CIFailed,
		l.Done,
	}
}

// GetByName returns a label config by its name
func (l *Labels) GetByName(name string) *LabelConfig {
	for _, label := range l.GetAllLabels() {
		if label.Name == name {
			return &label
		}
	}
	return nil
}

// GetPickupLabels returns labels that AI should pick up (always or on_user_comment)
func (l *Labels) GetPickupLabels() []LabelConfig {
	var result []LabelConfig
	for _, label := range l.GetAllLabels() {
		if label.AIPickup == string(PickupAlways) || label.AIPickup == string(PickupOnUserComment) {
			result = append(result, label)
		}
	}
	return result
}
