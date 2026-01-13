package git

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
}

// BranchInfo represents information about a git branch
type BranchInfo struct {
	Name       string
	Remote     string
	Upstream   string
	Commit     string
	IsCurrent  bool
	IsTracking bool
}
