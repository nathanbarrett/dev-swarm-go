# dev-swarm Technical Specification

**AI-powered development orchestration using GitHub Issues as the control plane.**

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Label State Machine](#label-state-machine)
4. [Configuration](#configuration)
5. [Project Structure](#project-structure)
6. [CLI Commands](#cli-commands)
7. [TUI Design](#tui-design)
8. [Core Components](#core-components)
9. [Startup Sequence](#startup-sequence)
10. [Main Loop](#main-loop)
11. [Session Management](#session-management)
12. [Claude Integration](#claude-integration)
13. [GitHub Operations](#github-operations)
14. [Git Worktree Management](#git-worktree-management)
15. [npm Distribution](#npm-distribution)
16. [Error Handling](#error-handling)
17. [Dependencies](#dependencies)

---

## Overview

`dev-swarm` is a local daemon that monitors GitHub issues across multiple repositories, automatically spawning Claude Code sessions to plan, implement, and fix code based on label-driven workflows.

### Key Features

- **Label-based state machine**: No Kanban boards needed, just labels on issues
- **Multi-repo support**: Monitor and work on multiple codebases simultaneously
- **Concurrent sessions**: Git worktrees enable parallel work on the same repo
- **Real-time TUI**: Watch AI progress across all active sessions
- **Auto-merge**: Optionally merge PRs when user approves
- **Cross-platform**: Works on macOS and Linux

### How It Works

1. User creates a GitHub issue and adds the `user:ready-to-plan` label
2. `dev-swarm` detects the label and spawns a Claude session
3. Claude analyzes the issue, writes an implementation plan, updates the label
4. User reviews, provides feedback or approves
5. Claude implements the solution, creates a PR
6. User reviews code, approves
7. PR is merged, issue marked done

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   GitHub (Control Plane)                   Local Machine (Execution Plane)   │
│                                                                              │
│   ┌─────────────────────┐                 ┌─────────────────────────────┐   │
│   │  Issues + Labels    │◀───gh cli──────▶│        dev-swarm            │   │
│   │  PR Comments        │                 │                             │   │
│   │  CI Status          │                 │  ┌───────────────────────┐  │   │
│   └─────────────────────┘                 │  │    Orchestrator       │  │   │
│                                           │  │    - Poller           │  │   │
│   Labels = State                          │  │    - State Machine    │  │   │
│   Comments = Communication                │  │    - Session Manager  │  │   │
│                                           │  └───────────┬───────────┘  │   │
│                                           │              │              │   │
│                                           │  ┌───────────▼───────────┐  │   │
│                                           │  │   Claude Sessions     │  │   │
│                                           │  │  ┌─────┐ ┌─────┐     │  │   │
│                                           │  │  │ #42 │ │ #45 │ ... │  │   │
│                                           │  │  └─────┘ └─────┘     │  │   │
│                                           │  └───────────────────────┘  │   │
│                                           │              │              │   │
│                                           │  ┌───────────▼───────────┐  │   │
│                                           │  │         TUI           │  │   │
│                                           │  │   Status + Output     │  │   │
│                                           │  └───────────────────────┘  │   │
│                                           └─────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Label State Machine

### Label Definitions

| Label | Owner | AI Pickup | Color | Hex | Description |
|-------|-------|-----------|-------|-----|-------------|
| `user:ready-to-plan` | User | ✅ Yes | Blue | `#0052CC` | Issue ready for AI to create implementation plan |
| `ai:planning` | AI | ❌ No | Yellow | `#FBCA04` | AI is actively writing the implementation plan |
| `user:plan-review` | User | ⚡ Conditional | Blue | `#0052CC` | Plan complete, user is reviewing |
| `user:ready-to-implement` | User | ✅ Yes | Blue | `#0052CC` | Plan approved, ready for AI to code |
| `ai:implementing` | AI | ❌ No | Yellow | `#FBCA04` | AI is actively writing code |
| `user:code-review` | User | ⚡ Conditional | Blue | `#0052CC` | PR created, user is reviewing code |
| `ai:ci-failed` | AI | ✅ Yes | Red | `#D93F0B` | CI pipeline failed, AI will fix |
| `user:blocked` | User | ❌ No | Red | `#D93F0B` | AI is stuck, needs human intervention |
| `ai:done` | AI | ❌ No | Green | `#0E8A16` | Complete, PR merged |

### Ownership Convention

- **`user:` prefix**: Waiting for human action (user owns the next step)
- **`ai:` prefix**: AI is working or AI owns the state

### State Flow Diagram

```
                              User adds label
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │ user:ready-to-plan  │
                         └──────────┬──────────┘
                                    │ AI picks up
                                    ▼
                         ┌─────────────────────┐
                         │    ai:planning      │
                         └──────────┬──────────┘
                                    │ AI completes plan
                                    ▼
                         ┌─────────────────────┐        User provides feedback
            ┌───────────▶│  user:plan-review   │◀─────────────────┐
            │            └──────────┬──────────┘                  │
            │                       │                             │
            │            User comments "approved"                 │
            │                       │                             │
            │                       ▼                             │
            │  ┌───────────────────────────────────┐              │
            │  │     user:ready-to-implement       │              │
            │  └────────────────┬──────────────────┘              │
            │                   │ AI picks up                     │
            │                   ▼                                 │
            │  ┌───────────────────────────────────┐              │
            │  │         ai:implementing           │──────────────┘
            │  └────────────────┬──────────────────┘  (AI revises plan)
            │                   │ AI creates PR
            │                   ▼
            │  ┌───────────────────────────────────┐
            │  │        user:code-review           │◀────────────────┐
            │  └────────────────┬──────────────────┘                 │
            │                   │                                    │
            │        User comments "approved"              User requests changes
            │                   │                                    │
            │                   ▼                                    │
            │  ┌───────────────────────────────────┐                │
            │  │            ai:done                │                │
            │  └───────────────────────────────────┘                │
            │                                                        │
            │  ┌───────────────────────────────────┐                │
            └──│         ai:implementing           │────────────────┘
               └───────────────────────────────────┘
                              ▲
                              │ AI picks up, fixes
               ┌───────────────────────────────────┐
               │          ai:ci-failed             │◀── CI fails at any point
               └───────────────────────────────────┘


               ┌───────────────────────────────────┐
               │          user:blocked             │◀── AI gets stuck
               └───────────────────────────────────┘
```

### Pickup Rules

```go
type PickupRule string

const (
    PickupAlways      PickupRule = "always"       // Always pick up if no active session
    PickupNever       PickupRule = "never"        // Never pick up
    PickupOnUserComment PickupRule = "on_user_comment" // Only if new user comment
)
```

| Label | Pickup Rule | Condition |
|-------|-------------|-----------|
| `user:ready-to-plan` | `always` | — |
| `ai:planning` | `never` | — |
| `user:plan-review` | `on_user_comment` | New comment without `<!-- dev-swarm:ai -->` marker since last AI comment |
| `user:ready-to-implement` | `always` | — |
| `ai:implementing` | `never` | — |
| `user:code-review` | `on_user_comment` | New PR comment/review without AI marker since last AI comment |
| `ai:ci-failed` | `always` | — |
| `user:blocked` | `never` | — |
| `ai:done` | `never` | — |

### State Transitions

| From | To | Trigger |
|------|-----|---------|
| `user:ready-to-plan` | `ai:planning` | AI picks up issue |
| `ai:planning` | `user:plan-review` | AI completes plan |
| `user:plan-review` | `user:ready-to-implement` | User approves (comments approval keyword) |
| `user:plan-review` | `ai:planning` | User provides feedback (AI picks up, revises) |
| `user:ready-to-implement` | `ai:implementing` | AI picks up issue |
| `ai:implementing` | `user:code-review` | AI creates PR |
| `user:code-review` | `ai:done` | User approves, PR merged |
| `user:code-review` | `ai:implementing` | User requests changes |
| `ai:ci-failed` | `ai:implementing` | AI picks up to fix |
| `ai:implementing` | `user:code-review` | AI pushes fix, CI passes |
| Any | `user:blocked` | AI cannot proceed |
| Any | `ai:ci-failed` | CI fails (detected via GitHub) |

---

## Configuration

### Config File Location

```
~/.config/dev-swarm/
├── config.yaml              # Main configuration file
├── dev-swarm.lock           # PID lock file (created at runtime)
├── dev-swarm.log            # Log file
└── worktrees/               # Git worktrees directory
    ├── {repo-name}/
    │   ├── issue-{number}/
    │   └── issue-{number}/
    └── {repo-name}/
        └── issue-{number}/
```

### Full Config File

```yaml
# ~/.config/dev-swarm/config.yaml
# dev-swarm configuration file

# =============================================================================
# SETTINGS
# =============================================================================
settings:
  # Polling intervals (in seconds)
  poll_interval: 60              # Time between GitHub checks when idle
  active_poll_interval: 10       # Time between checks when sessions are active
  
  # Session limits
  max_concurrent_sessions: 5     # Maximum simultaneous Claude sessions (total)
  
  # PR merge behavior
  auto_merge_on_approval: true   # Automatically merge PR when user approves
  approval_keywords:             # Keywords that trigger approval (case-insensitive)
    - "approved"
    - "lgtm"
    - "ship it"
    - "merge it"
    - "looks good"
  
  # Output buffer
  output_buffer_lines: 1000      # Number of lines to keep per session

# =============================================================================
# LABELS (Global Defaults)
# =============================================================================
# These labels are created on each repo if they don't exist.
# Codebases can override individual labels.

labels:
  # ---------------------------------------------------------------------------
  # USER-OWNED LABELS (waiting for human action)
  # ---------------------------------------------------------------------------
  ready_to_plan:
    name: "user:ready-to-plan"
    color: "0052CC"
    description: "Ready for AI to create implementation plan"
    owner: user
    ai_pickup: always
    ai_action: |
      You are creating an implementation plan for this issue.
      
      Steps:
      1. Change the label from user:ready-to-plan to ai:planning
      2. Analyze the issue requirements thoroughly
      3. Write a detailed implementation plan as a comment including:
         - **Summary**: Brief overview of the approach
         - **Files to Modify/Create**: List each file and what changes are needed
         - **Implementation Steps**: Detailed step-by-step instructions
         - **Edge Cases**: Potential issues to handle
         - **Testing Strategy**: How to verify the implementation
      4. Wrap your comment with AI markers (see below)
      5. Change the label from ai:planning to user:plan-review
    
  plan_review:
    name: "user:plan-review"
    color: "0052CC"
    description: "Implementation plan ready for user review"
    owner: user
    ai_pickup: on_user_comment
    ai_action: |
      The user has commented on your implementation plan.
      
      Check the user's comment:
      - If it contains an approval keyword (approved, lgtm, ship it, etc.):
        → Change label from user:plan-review to user:ready-to-implement
      - If it contains feedback, questions, or change requests:
        → Change label from user:plan-review to ai:planning
        → Revise the implementation plan based on the feedback
        → Add a new comment with the updated plan
        → Change label from ai:planning to user:plan-review
    
  ready_to_implement:
    name: "user:ready-to-implement"
    color: "0052CC"
    description: "Plan approved, ready for AI to implement"
    owner: user
    ai_pickup: always
    ai_action: |
      You are implementing an approved plan.
      
      Steps:
      1. Change label from user:ready-to-implement to ai:implementing
      2. Read the implementation plan from the issue comments
      3. Create clean, well-documented code following the plan
      4. Follow existing code style and patterns in the repository
      5. Write tests as specified in the plan
      6. Make atomic commits with clear messages that reference the issue:
         - "Add dark mode context provider (fixes #42)"
         - "Add theme toggle component (#42)"
      7. Create a Pull Request with:
         - Title: Clear description referencing issue number
         - Body: Summary of changes, link to issue (Closes #XX)
      8. Change label from ai:implementing to user:code-review
    
  code_review:
    name: "user:code-review"
    color: "0052CC"
    description: "PR created, awaiting user code review"
    owner: user
    ai_pickup: on_user_comment
    ai_action: |
      The user has commented on your Pull Request.
      
      Check the user's comment or review:
      - If it contains an approval keyword (approved, lgtm, ship it, etc.):
        → Merge the PR using: gh pr merge {number} --merge --delete-branch
        → Change label from user:code-review to ai:done
        → Add a closing comment summarizing what was implemented
      - If it contains change requests or feedback:
        → Change label from user:code-review to ai:implementing
        → Address each review comment specifically
        → Push new commits with clear messages
        → Change label from ai:implementing to user:code-review
        → Reply to review comments explaining your changes
    
  blocked:
    name: "user:blocked"
    color: "D93F0B"
    description: "AI is stuck and needs human intervention"
    owner: user
    ai_pickup: never

  # ---------------------------------------------------------------------------
  # AI-OWNED LABELS (AI is working)
  # ---------------------------------------------------------------------------
  planning:
    name: "ai:planning"
    color: "FBCA04"
    description: "AI is writing the implementation plan"
    owner: ai
    ai_pickup: never
    
  implementing:
    name: "ai:implementing"
    color: "FBCA04"
    description: "AI is actively writing code"
    owner: ai
    ai_pickup: never
    
  ci_failed:
    name: "ai:ci-failed"
    color: "D93F0B"
    description: "CI failed, AI will analyze and fix"
    owner: ai
    ai_pickup: always
    ai_action: |
      The CI pipeline has failed. You need to fix it.
      
      Steps:
      1. Change label from ai:ci-failed to ai:implementing
      2. Fetch and analyze the CI logs using: gh run view --log-failed
      3. Identify the cause of failure (test failures, build errors, lint errors)
      4. Fix the issues in your code
      5. Commit with a clear message: "Fix CI failures (#XX)"
      6. Push the fix
      7. Wait briefly for CI to start, then change label to user:code-review
      
      Do NOT change unrelated code. Focus only on fixing the CI failure.
    
  done:
    name: "ai:done"
    color: "0E8A16"
    description: "Complete - PR merged"
    owner: ai
    ai_pickup: never

# =============================================================================
# AI INSTRUCTIONS (Global)
# =============================================================================
# These instructions are included in every Claude session context

ai_instructions:
  general: |
    You are an AI developer working on GitHub issues via the dev-swarm orchestrator.
    
    ## Important Guidelines
    
    1. **Label Management**: Always update labels to reflect current state using gh CLI:
       ```bash
       gh issue edit {number} --remove-label "old:label" --add-label "new:label"
       ```
    
    2. **Comment Markers**: Wrap ALL your comments with these markers so the system 
       can distinguish AI comments from user comments:
       ```
       <!-- dev-swarm:ai -->
       Your comment content here
       <!-- /dev-swarm:ai -->
       ```
    
    3. **Commit Messages**: Always reference the issue number:
       - "Add feature X (#42)"
       - "Fix bug in Y (fixes #42)"
    
    4. **Pull Requests**: Link to the issue in the PR body:
       - "Closes #42" or "Fixes #42"
    
    5. **Getting Stuck**: If you cannot proceed for any reason:
       - Change label to user:blocked
       - Add a comment explaining what's blocking you
       - Be specific about what you need from the user
    
    6. **Code Quality**:
       - Follow existing code style in the repository
       - Write clear, maintainable code
       - Add comments for complex logic
       - Write tests when specified in the plan

# =============================================================================
# CODEBASES
# =============================================================================
# List of repositories to monitor

codebases:
  # Example codebase (uses all global defaults)
  - name: my-web-app
    repo: owner/my-web-app           # GitHub repo (owner/name format)
    local_path: ~/code/my-web-app    # Local clone path
    default_branch: main             # Branch to create PRs against
    enabled: true                    # Set to false to disable monitoring
    
  # Example with custom settings
  # - name: api-service
  #   repo: owner/api-service
  #   local_path: ~/code/api-service
  #   default_branch: main
  #   enabled: true
  #   # Override specific labels for this repo
  #   labels:
  #     ready_to_plan:
  #       name: "needs-plan"         # Different label name
  #       color: "1D76DB"            # Different color
  #     planning:
  #       name: "planning"
```

### Config Validation Rules

1. **Required fields**:
   - `codebases[].repo` - must be in `owner/name` format
   - `codebases[].local_path` - must be a valid path
   - `codebases[].default_branch` - must be specified

2. **Path expansion**: `~` should be expanded to home directory

3. **Label merging**: Per-codebase labels are merged with global labels. Only specified fields are overridden.

4. **Defaults**:
   - `poll_interval`: 60
   - `active_poll_interval`: 10
   - `max_concurrent_sessions`: 5
   - `auto_merge_on_approval`: true
   - `output_buffer_lines`: 1000

---

## Project Structure

```
dev-swarm/
├── cmd/
│   └── dev-swarm/
│       └── main.go                     # CLI entry point
│
├── internal/
│   ├── config/
│   │   ├── config.go                   # Config loading and validation
│   │   ├── labels.go                   # Label config with merge logic
│   │   ├── defaults.go                 # Default configuration values
│   │   └── types.go                    # Config type definitions
│   │
│   ├── github/
│   │   ├── client.go                   # gh CLI wrapper
│   │   ├── issues.go                   # Issue operations (list, get, edit)
│   │   ├── labels.go                   # Label operations (list, create, sync)
│   │   ├── comments.go                 # Comment operations
│   │   ├── prs.go                      # PR operations (create, merge, list)
│   │   ├── ci.go                       # CI status checking
│   │   └── types.go                    # GitHub type definitions
│   │
│   ├── git/
│   │   ├── repo.go                     # Repository operations
│   │   ├── worktree.go                 # Worktree create/remove
│   │   ├── branch.go                   # Branch operations
│   │   └── types.go                    # Git type definitions
│   │
│   ├── orchestrator/
│   │   ├── orchestrator.go             # Main orchestrator struct and lifecycle
│   │   ├── loop.go                     # Main polling loop
│   │   ├── poller.go                   # GitHub polling logic
│   │   ├── pickup.go                   # Issue pickup decision logic
│   │   ├── transitions.go              # Label state transition logic
│   │   └── types.go                    # Orchestrator type definitions
│   │
│   ├── session/
│   │   ├── manager.go                  # Session manager (tracks all sessions)
│   │   ├── session.go                  # Individual session struct
│   │   ├── spawner.go                  # Claude process spawning
│   │   ├── output.go                   # Output capture and buffering
│   │   ├── context.go                  # Build Claude context/prompt
│   │   └── types.go                    # Session type definitions
│   │
│   ├── lock/
│   │   └── lock.go                     # PID lock file management
│   │
│   └── tui/
│       ├── app.go                      # Bubbletea application setup
│       ├── model.go                    # TUI state model
│       ├── update.go                   # Event/message handlers
│       ├── view.go                     # Main render function
│       ├── keymap.go                   # Keyboard shortcuts
│       ├── styles.go                   # Lipgloss styles
│       └── components/
│           ├── statusbar.go            # Bottom status bar
│           ├── codebase.go             # Codebase list item
│           ├── session.go              # Session list item
│           └── output.go               # Output panel
│
├── pkg/
│   └── version/
│       └── version.go                  # Version info (injected at build)
│
├── configs/
│   └── default.yaml                    # Default config template
│
├── scripts/
│   ├── build.sh                        # Local build script
│   └── release.sh                      # Release automation
│
├── npm/
│   ├── package.json                    # npm package definition
│   ├── install.js                      # Post-install binary downloader
│   └── bin/
│       └── dev-swarm                   # Shell wrapper script
│
├── .goreleaser.yaml                    # GoReleaser configuration
├── go.mod                              # Go module definition
├── go.sum                              # Go dependencies
├── LICENSE                             # MIT License
└── README.md                           # Documentation
```

---

## CLI Commands

### Command Reference

```bash
# =============================================================================
# MAIN COMMANDS
# =============================================================================

# Start the orchestrator with TUI
dev-swarm start

# Start with custom config file
dev-swarm start --config /path/to/config.yaml

# Start in daemon mode (background, no TUI, logs to file)
dev-swarm start --daemon

# =============================================================================
# SETUP COMMANDS
# =============================================================================

# Initialize configuration (interactive)
# Creates ~/.config/dev-swarm/config.yaml with prompts
dev-swarm init

# Add a codebase to config
dev-swarm add --repo owner/repo --path ~/code/repo
dev-swarm add --repo owner/repo --path ~/code/repo --branch main --name my-repo

# Remove a codebase from config
dev-swarm remove my-repo
dev-swarm remove owner/repo

# List all configured codebases
dev-swarm list

# Sync labels to a specific repo (creates missing labels)
dev-swarm sync-labels owner/repo

# Sync labels to all enabled repos
dev-swarm sync-labels --all

# =============================================================================
# STATUS & MONITORING
# =============================================================================

# Show current status
# - Is orchestrator running?
# - Active sessions
# - Pending issues
dev-swarm status

# View logs
dev-swarm logs                  # Last 50 lines
dev-swarm logs --lines 100      # Last 100 lines
dev-swarm logs --follow         # Tail logs (like tail -f)

# =============================================================================
# CONTROL COMMANDS
# =============================================================================

# Stop the running orchestrator gracefully
dev-swarm stop

# Clean up orphaned worktrees
dev-swarm cleanup

# =============================================================================
# INFO COMMANDS
# =============================================================================

# Show version
dev-swarm version

# Show help
dev-swarm help
dev-swarm help start
dev-swarm help add
```

### CLI Implementation Details

Use **Cobra** for CLI framework:

```go
// cmd/dev-swarm/main.go

package main

import (
    "os"
    
    "github.com/spf13/cobra"
    "dev-swarm/internal/config"
    "dev-swarm/internal/orchestrator"
    "dev-swarm/pkg/version"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "dev-swarm",
        Short: "AI-powered development orchestration",
        Long:  `dev-swarm monitors GitHub issues and automatically spawns Claude Code sessions to plan, implement, and fix code.`,
    }
    
    // Add commands
    rootCmd.AddCommand(
        newStartCmd(),
        newInitCmd(),
        newAddCmd(),
        newRemoveCmd(),
        newListCmd(),
        newSyncLabelsCmd(),
        newStatusCmd(),
        newLogsCmd(),
        newStopCmd(),
        newCleanupCmd(),
        newVersionCmd(),
    )
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

---

## TUI Design

### Split View Layout

```
╭─ dev-swarm v0.1.0 ───────────────────────────────────────────────────────────╮
│                                                                              │
│  my-web-app • github.com/nathan/my-web-app                                   │
│  ├─ #42 Add dark mode support                                                │
│  │   └─ ai:implementing ● active (3m 12s)                     ◀── selected  │
│  └─ #45 Fix Safari login bug                                                 │
│      └─ user:plan-review ○ waiting                                           │
│                                                                              │
│  api-service • github.com/nathan/api-service                                 │
│  ├─ #12 Add rate limiting                                                    │
│  │   └─ user:code-review ○ waiting                                           │
│  └─ (no other active issues)                                                 │
│                                                                              │
│  mobile-app • github.com/nathan/mobile-app                                   │
│  └─ (idle)                                                                   │
│                                                                              │
├─ Output: #42 Add dark mode support ──────────────────────────────────────────┤
│                                                                              │
│  [14:32:01] Starting session for issue #42                                   │
│  [14:32:01] Worktree: ~/.config/dev-swarm/worktrees/my-web-app/issue-42     │
│  [14:32:02] Reading issue details...                                         │
│  [14:32:03] Found implementation plan in comments                            │
│  [14:32:03]                                                                  │
│  [14:32:05] Creating branch claude/issue-42 from main...                     │
│  [14:32:06] ✓ Branch created                                                 │
│  [14:32:06]                                                                  │
│  [14:32:08] Implementing dark mode support...                                │
│  [14:32:08] → Modifying src/contexts/ThemeContext.tsx                        │
│  [14:32:15] → Creating src/styles/dark-theme.css                             │
│  [14:32:22] → Updating src/components/Settings/ThemeToggle.tsx               │
│  [14:32:30] █                                                                │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│  Active: 1   Queued: 0   Waiting: 2   │  Poll: 47s  │  ↑↓ Nav  Q Quit       │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Status Icons

| Icon | Meaning | Color |
|------|---------|-------|
| `●` | Active session (Claude running) | Yellow |
| `◆` | Queued (about to start) | Cyan |
| `○` | Waiting (needs user action) | Blue |
| `✗` | Failed/Blocked | Red |
| `✓` | Done | Green |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Focus selected session output |
| `Esc` | Return to split view |
| `r` | Force refresh (poll now) |
| `p` | Pause/resume polling |
| `l` | Toggle log panel |
| `q` | Quit |
| `?` | Show help |

### TUI Model

```go
// internal/tui/model.go

package tui

import (
    "dev-swarm/internal/orchestrator"
    "dev-swarm/internal/session"
)

type Model struct {
    // State from orchestrator
    codebases    []CodebaseState
    sessions     map[string]*session.Session
    
    // UI state
    selectedIdx  int
    selectedType SelectionType  // Codebase or Session
    scrollOffset int
    outputScroll int
    
    // Status
    lastPoll     time.Time
    nextPoll     time.Time
    isPaused     bool
    
    // Dimensions
    width        int
    height       int
    
    // Channels for updates
    updateChan   chan orchestrator.StateUpdate
}

type CodebaseState struct {
    Name     string
    Repo     string
    Issues   []IssueState
    IsIdle   bool
}

type IssueState struct {
    Number    int
    Title     string
    Label     string
    Status    SessionStatus
    Duration  time.Duration
    HasOutput bool
}

type SessionStatus int

const (
    StatusQueued SessionStatus = iota
    StatusActive
    StatusWaiting
    StatusFailed
    StatusDone
)
```

### Styles

```go
// internal/tui/styles.go

package tui

import "github.com/charmbracelet/lipgloss"

var (
    // Colors
    ColorBlue   = lipgloss.Color("#0052CC")
    ColorYellow = lipgloss.Color("#FBCA04")
    ColorRed    = lipgloss.Color("#D93F0B")
    ColorGreen  = lipgloss.Color("#0E8A16")
    ColorGray   = lipgloss.Color("#6B7280")
    ColorWhite  = lipgloss.Color("#FFFFFF")
    
    // Styles
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorWhite)
    
    CodebaseStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorBlue)
    
    IssueStyle = lipgloss.NewStyle().
        Foreground(ColorWhite)
    
    LabelActiveStyle = lipgloss.NewStyle().
        Foreground(ColorYellow)
    
    LabelWaitingStyle = lipgloss.NewStyle().
        Foreground(ColorBlue)
    
    LabelFailedStyle = lipgloss.NewStyle().
        Foreground(ColorRed)
    
    LabelDoneStyle = lipgloss.NewStyle().
        Foreground(ColorGreen)
    
    SelectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#333333"))
    
    OutputPanelStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorGray)
    
    StatusBarStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#1a1a1a")).
        Foreground(ColorGray).
        Padding(0, 1)
)
```

---

## Core Components

### Orchestrator

```go
// internal/orchestrator/orchestrator.go

package orchestrator

import (
    "context"
    "sync"
    "time"
    
    "dev-swarm/internal/config"
    "dev-swarm/internal/github"
    "dev-swarm/internal/session"
)

type Orchestrator struct {
    config         *config.Config
    ghClient       *github.Client
    sessionManager *session.Manager
    
    // State
    mu             sync.RWMutex
    codebases      map[string]*CodebaseState
    
    // Control
    ctx            context.Context
    cancel         context.CancelFunc
    isPaused       bool
    
    // Communication
    stateChan      chan StateUpdate
    
    // Timing
    lastPoll       time.Time
    pollInterval   time.Duration
}

type CodebaseState struct {
    Config    *config.Codebase
    Issues    map[int]*IssueState
    LastPoll  time.Time
}

type IssueState struct {
    Issue       *github.Issue
    Label       string
    HasSession  bool
    SessionID   string
    LastChecked time.Time
}

type StateUpdate struct {
    Type      UpdateType
    Codebase  string
    IssueNum  int
    Data      interface{}
}

type UpdateType int

const (
    UpdateIssueFound UpdateType = iota
    UpdateSessionStarted
    UpdateSessionOutput
    UpdateSessionEnded
    UpdateLabelChanged
    UpdatePollComplete
)

func New(cfg *config.Config) (*Orchestrator, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Orchestrator{
        config:         cfg,
        ghClient:       github.NewClient(),
        sessionManager: session.NewManager(cfg.Settings.MaxConcurrentSessions),
        codebases:      make(map[string]*CodebaseState),
        ctx:            ctx,
        cancel:         cancel,
        stateChan:      make(chan StateUpdate, 100),
        pollInterval:   time.Duration(cfg.Settings.PollInterval) * time.Second,
    }, nil
}

func (o *Orchestrator) Start() error {
    // Validate codebases exist
    // Sync labels
    // Start polling loop
}

func (o *Orchestrator) Stop() {
    o.cancel()
    o.sessionManager.StopAll()
}

func (o *Orchestrator) StateChan() <-chan StateUpdate {
    return o.stateChan
}

func (o *Orchestrator) Pause() {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.isPaused = true
}

func (o *Orchestrator) Resume() {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.isPaused = false
}

func (o *Orchestrator) ForceRefresh() {
    // Trigger immediate poll
}
```

### Session Manager

```go
// internal/session/manager.go

package session

import (
    "sync"
    
    "dev-swarm/internal/config"
    "dev-swarm/internal/github"
)

type Manager struct {
    sessions     map[string]*Session  // sessionID -> Session
    maxActive    int
    mu           sync.RWMutex
    
    // Communication
    outputChan   chan OutputEvent
    statusChan   chan StatusEvent
}

type OutputEvent struct {
    SessionID string
    Line      OutputLine
}

type StatusEvent struct {
    SessionID string
    Status    Status
    ExitCode  *int
    Error     error
}

func NewManager(maxActive int) *Manager {
    return &Manager{
        sessions:   make(map[string]*Session),
        maxActive:  maxActive,
        outputChan: make(chan OutputEvent, 1000),
        statusChan: make(chan StatusEvent, 100),
    }
}

func (m *Manager) CanSpawn() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    activeCount := 0
    for _, s := range m.sessions {
        if s.Status == StatusRunning {
            activeCount++
        }
    }
    return activeCount < m.maxActive
}

func (m *Manager) HasSession(issueID string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    _, exists := m.sessions[issueID]
    return exists
}

func (m *Manager) SpawnSession(
    issue *github.Issue,
    codebase *config.Codebase,
    labels *config.LabelConfig,
    action string,
) (*Session, error) {
    // 1. Create worktree
    // 2. Build context
    // 3. Spawn Claude process
    // 4. Start output capture
    // 5. Track session
}

func (m *Manager) GetSession(sessionID string) *Session {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.sessions[sessionID]
}

func (m *Manager) GetAllSessions() []*Session {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    result := make([]*Session, 0, len(m.sessions))
    for _, s := range m.sessions {
        result = append(result, s)
    }
    return result
}

func (m *Manager) RemoveSession(sessionID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.sessions, sessionID)
}

func (m *Manager) StopAll() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for _, s := range m.sessions {
        s.Stop()
    }
}

func (m *Manager) OutputChan() <-chan OutputEvent {
    return m.outputChan
}

func (m *Manager) StatusChan() <-chan StatusEvent {
    return m.statusChan
}
```

### Individual Session

```go
// internal/session/session.go

package session

import (
    "bufio"
    "io"
    "os/exec"
    "sync"
    "time"
    
    "dev-swarm/internal/config"
    "dev-swarm/internal/github"
)

type Session struct {
    ID           string              // "owner/repo#123"
    Issue        *github.Issue
    Codebase     *config.Codebase
    WorktreePath string
    BranchName   string
    
    // Process
    cmd          *exec.Cmd
    stdin        io.WriteCloser
    stdout       io.ReadCloser
    stderr       io.ReadCloser
    
    // Output
    output       *OutputBuffer
    
    // Status
    Status       Status
    StartedAt    time.Time
    CompletedAt  *time.Time
    ExitCode     *int
    Error        error
    
    // Control
    mu           sync.RWMutex
    stopChan     chan struct{}
}

type Status int

const (
    StatusPending Status = iota
    StatusRunning
    StatusCompleted
    StatusFailed
)

type OutputBuffer struct {
    lines    []OutputLine
    maxLines int
    mu       sync.RWMutex
}

type OutputLine struct {
    Timestamp time.Time
    Text      string
    Stream    string  // "stdout" or "stderr"
}

func (s *Session) Start(outputChan chan<- OutputEvent, statusChan chan<- StatusEvent) error {
    s.mu.Lock()
    s.Status = StatusRunning
    s.StartedAt = time.Now()
    s.mu.Unlock()
    
    // Start output capture goroutines
    go s.captureOutput(s.stdout, "stdout", outputChan)
    go s.captureOutput(s.stderr, "stderr", outputChan)
    
    // Wait for completion
    go s.waitForCompletion(statusChan)
    
    return s.cmd.Start()
}

func (s *Session) captureOutput(reader io.Reader, stream string, outputChan chan<- OutputEvent) {
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        line := OutputLine{
            Timestamp: time.Now(),
            Text:      scanner.Text(),
            Stream:    stream,
        }
        
        s.output.Append(line)
        
        outputChan <- OutputEvent{
            SessionID: s.ID,
            Line:      line,
        }
    }
}

func (s *Session) waitForCompletion(statusChan chan<- StatusEvent) {
    err := s.cmd.Wait()
    
    s.mu.Lock()
    now := time.Now()
    s.CompletedAt = &now
    
    if err != nil {
        s.Status = StatusFailed
        s.Error = err
        if exitErr, ok := err.(*exec.ExitError); ok {
            code := exitErr.ExitCode()
            s.ExitCode = &code
        }
    } else {
        s.Status = StatusCompleted
        code := 0
        s.ExitCode = &code
    }
    s.mu.Unlock()
    
    statusChan <- StatusEvent{
        SessionID: s.ID,
        Status:    s.Status,
        ExitCode:  s.ExitCode,
        Error:     s.Error,
    }
}

func (s *Session) Stop() {
    if s.cmd != nil && s.cmd.Process != nil {
        s.cmd.Process.Kill()
    }
}

func (s *Session) GetOutput() []OutputLine {
    return s.output.GetAll()
}

func (s *Session) GetRecentOutput(n int) []OutputLine {
    return s.output.GetRecent(n)
}

func (s *Session) Duration() time.Duration {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    if s.CompletedAt != nil {
        return s.CompletedAt.Sub(s.StartedAt)
    }
    return time.Since(s.StartedAt)
}

// OutputBuffer methods

func NewOutputBuffer(maxLines int) *OutputBuffer {
    return &OutputBuffer{
        lines:    make([]OutputLine, 0, maxLines),
        maxLines: maxLines,
    }
}

func (b *OutputBuffer) Append(line OutputLine) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if len(b.lines) >= b.maxLines {
        // Remove oldest line
        b.lines = b.lines[1:]
    }
    b.lines = append(b.lines, line)
}

func (b *OutputBuffer) GetAll() []OutputLine {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    result := make([]OutputLine, len(b.lines))
    copy(result, b.lines)
    return result
}

func (b *OutputBuffer) GetRecent(n int) []OutputLine {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    if n >= len(b.lines) {
        result := make([]OutputLine, len(b.lines))
        copy(result, b.lines)
        return result
    }
    
    start := len(b.lines) - n
    result := make([]OutputLine, n)
    copy(result, b.lines[start:])
    return result
}
```

---

## Startup Sequence

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                             STARTUP SEQUENCE                                 │
└─────────────────────────────────────────────────────────────────────────────┘

1. PARSE CLI FLAGS
   └─ --config: path to config file (default: ~/.config/dev-swarm/config.yaml)
   └─ --daemon: run in background mode

2. LOAD CONFIGURATION
   └─ Read config file
   └─ Validate required fields
   └─ Apply defaults for optional fields
   └─ Expand paths (~/ → /home/user/)
   └─ Merge per-codebase label overrides with global labels
   
   On error: Exit with descriptive message

3. CHECK LOCK FILE
   └─ Lock file: ~/.config/dev-swarm/dev-swarm.lock
   
   If lock file exists:
      └─ Read PID from file
      └─ Check if process is running (kill -0 $PID)
      └─ If running: Exit with "Already running (PID: X). Use 'dev-swarm stop' first."
      └─ If not running: Remove stale lock file, continue
   
   Create lock file with current PID
   Register signal handlers (SIGINT, SIGTERM) to:
      └─ Clean up lock file
      └─ Stop all sessions gracefully
      └─ Exit cleanly

4. VERIFY DEPENDENCIES
   └─ Check `gh` CLI is installed:
      $ which gh
      On error: Exit with "gh CLI not found. Install: https://cli.github.com"
   
   └─ Check `gh` is authenticated:
      $ gh auth status
      On error: Exit with "gh not authenticated. Run: gh auth login"
   
   └─ Check `claude` CLI is available:
      $ which claude
      On error: Exit with "Claude CLI not found. Install: npm install -g @anthropic-ai/claude-code"

5. VALIDATE CODEBASES
   For each enabled codebase:
      └─ Verify local_path exists:
         $ test -d {local_path}
         On error: Warn (don't fail) "Path not found: {path}"
      
      └─ Verify local_path is a git repo:
         $ git -C {local_path} rev-parse --git-dir
         On error: Warn "Not a git repo: {path}"
      
      └─ Verify GitHub access:
         $ gh repo view {owner/repo}
         On error: Warn "Cannot access repo: {repo}"
      
      └─ Verify default branch exists:
         $ git -C {local_path} rev-parse --verify {branch}
         On error: Warn "Branch not found: {branch}"

6. SYNC LABELS
   For each enabled codebase:
      └─ Get existing labels:
         $ gh label list --repo {repo} --json name,color,description
      
      └─ Get required labels from config (global + overrides)
      
      └─ For each required label:
         If not exists:
            $ gh label create "{name}" --repo {repo} \
                --color "{color}" \
                --description "{description}"
            Log: "Created label: {name} on {repo}"
         
         If exists:
            Do nothing (don't update existing labels)
      
      Log: "Labels synced for {repo}"

7. CREATE WORKTREES DIRECTORY
   $ mkdir -p ~/.config/dev-swarm/worktrees

8. INITIALIZE ORCHESTRATOR
   └─ Create orchestrator instance
   └─ Initialize session manager
   └─ Initialize GitHub client

9. START TUI (if not daemon mode)
   └─ Initialize Bubbletea application
   └─ Start render loop
   └─ Connect orchestrator state channel to TUI

10. START POLLING LOOP
    └─ Begin main orchestration loop
    └─ Log: "dev-swarm started. Monitoring {N} codebases."
```

---

## Main Loop

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MAIN POLLING LOOP                               │
└─────────────────────────────────────────────────────────────────────────────┘

LOOP:
    Determine poll interval:
        If any active sessions: use active_poll_interval (default: 10s)
        Else: use poll_interval (default: 60s)
    
    Sleep for poll interval (interruptible)
    
    If paused: continue to next iteration
    
    ┌─ FOR EACH enabled codebase ─────────────────────────────────────────────┐
    │                                                                          │
    │  1. FETCH ISSUES WITH PICKUP LABELS                                      │
    │     ┌────────────────────────────────────────────────────────────────┐  │
    │     │ Direct pickup (always):                                         │  │
    │     │   $ gh issue list --repo {repo} --label "user:ready-to-plan"   │  │
    │     │   $ gh issue list --repo {repo} --label "user:ready-to-implement"│ │
    │     │   $ gh issue list --repo {repo} --label "ai:ci-failed"         │  │
    │     │                                                                 │  │
    │     │ Conditional pickup (check for user comments):                   │  │
    │     │   $ gh issue list --repo {repo} --label "user:plan-review"     │  │
    │     │   $ gh issue list --repo {repo} --label "user:code-review"     │  │
    │     └────────────────────────────────────────────────────────────────┘  │
    │                                                                          │
    │  2. FOR EACH issue found:                                                │
    │     ┌────────────────────────────────────────────────────────────────┐  │
    │     │ a. Skip if session already exists for this issue               │  │
    │     │                                                                 │  │
    │     │ b. Skip if max_concurrent_sessions reached                     │  │
    │     │                                                                 │  │
    │     │ c. For conditional pickup labels:                              │  │
    │     │    - Fetch issue comments                                      │  │
    │     │    - Check if latest non-AI comment is newer than last AI     │  │
    │     │      comment (comments without <!-- dev-swarm:ai --> marker)  │  │
    │     │    - If no new user comment: skip                             │  │
    │     │                                                                 │  │
    │     │ d. Get full issue details:                                     │  │
    │     │    $ gh issue view {number} --repo {repo} \                   │  │
    │     │        --json number,title,body,labels,comments               │  │
    │     │                                                                 │  │
    │     │ e. Determine action based on current label                     │  │
    │     │                                                                 │  │
    │     │ f. SPAWN SESSION (see Session Management)                      │  │
    │     └────────────────────────────────────────────────────────────────┘  │
    │                                                                          │
    │  3. CHECK CI STATUS FOR ACTIVE ISSUES                                    │
    │     ┌────────────────────────────────────────────────────────────────┐  │
    │     │ For issues with ai:implementing or user:code-review:           │  │
    │     │   $ gh pr list --repo {repo} --head "claude/issue-{N}"        │  │
    │     │   $ gh pr checks {pr_number} --repo {repo}                    │  │
    │     │                                                                 │  │
    │     │ If CI failed and label is not ai:ci-failed:                   │  │
    │     │   $ gh issue edit {number} --repo {repo} \                    │  │
    │     │       --remove-label "{current}" --add-label "ai:ci-failed"   │  │
    │     └────────────────────────────────────────────────────────────────┘  │
    │                                                                          │
    │  4. UPDATE TUI STATE                                                     │
    │     └─ Send StateUpdate to state channel                                │
    │                                                                          │
    └──────────────────────────────────────────────────────────────────────────┘
    
    5. CHECK SESSION STATUS
       ┌────────────────────────────────────────────────────────────────────┐
       │ For each tracked session:                                          │
       │   - Check if process has exited                                    │
       │   - If exited:                                                     │
       │     - If exit code 0: session completed successfully               │
       │     - If exit code != 0: session failed                           │
       │     - Check if label was updated (Claude's responsibility)        │
       │     - If label is ai:done: clean up worktree immediately         │
       │     - Remove session from tracking                                │
       └────────────────────────────────────────────────────────────────────┘
    
    6. CLEAN UP MERGED PRS
       ┌────────────────────────────────────────────────────────────────────┐
       │ $ gh pr list --repo {repo} --state merged --json number,headRefName│
       │                                                                    │
       │ For each merged PR with branch "claude/issue-{N}":                │
       │   - If worktree exists for issue-{N}: remove it                  │
       │   - If local branch exists: delete it                            │
       └────────────────────────────────────────────────────────────────────┘

CONTINUE LOOP
```

### Pickup Logic Detail

```go
// internal/orchestrator/pickup.go

package orchestrator

import (
    "strings"
    "time"
    
    "dev-swarm/internal/config"
    "dev-swarm/internal/github"
)

// ShouldPickup determines if an issue should be picked up for processing
func (o *Orchestrator) ShouldPickup(issue *github.Issue, labelCfg *config.LabelConfig) bool {
    // Already has active session?
    if o.sessionManager.HasSession(issue.ID()) {
        return false
    }
    
    // At max capacity?
    if !o.sessionManager.CanSpawn() {
        return false
    }
    
    // Check pickup rule
    switch labelCfg.AIPickup {
    case "always":
        return true
        
    case "never":
        return false
        
    case "on_user_comment":
        return o.hasNewUserComment(issue)
        
    default:
        return false
    }
}

// hasNewUserComment checks if there's a new user comment since the last AI comment
func (o *Orchestrator) hasNewUserComment(issue *github.Issue) bool {
    comments := issue.Comments
    if len(comments) == 0 {
        return false
    }
    
    var lastAICommentTime time.Time
    var lastUserCommentTime time.Time
    
    for _, comment := range comments {
        isAI := strings.Contains(comment.Body, "<!-- dev-swarm:ai -->")
        
        if isAI {
            if comment.CreatedAt.After(lastAICommentTime) {
                lastAICommentTime = comment.CreatedAt
            }
        } else {
            if comment.CreatedAt.After(lastUserCommentTime) {
                lastUserCommentTime = comment.CreatedAt
            }
        }
    }
    
    // New user comment exists if it's after the last AI comment
    // (or if there's never been an AI comment but there are user comments)
    if lastUserCommentTime.IsZero() {
        return false
    }
    
    return lastUserCommentTime.After(lastAICommentTime)
}

// For code review, also check PR comments
func (o *Orchestrator) hasNewUserPRComment(issue *github.Issue, repo string) bool {
    // Get PR for this issue
    pr, err := o.ghClient.GetPRForBranch(repo, fmt.Sprintf("claude/issue-%d", issue.Number))
    if err != nil || pr == nil {
        return false
    }
    
    // Check PR comments and reviews
    comments, _ := o.ghClient.GetPRComments(repo, pr.Number)
    reviews, _ := o.ghClient.GetPRReviews(repo, pr.Number)
    
    // Similar logic: find latest AI vs user activity
    // ...
    
    return hasNewUserActivity
}
```

---

## Session Management

### Spawning a Session

```go
// internal/session/spawner.go

package session

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    
    "dev-swarm/internal/config"
    "dev-swarm/internal/git"
    "dev-swarm/internal/github"
)

func (m *Manager) SpawnSession(
    issue *github.Issue,
    codebase *config.Codebase,
    labels *config.Labels,
    currentLabel string,
) (*Session, error) {
    sessionID := fmt.Sprintf("%s#%d", codebase.Repo, issue.Number)
    branchName := fmt.Sprintf("claude/issue-%d", issue.Number)
    
    // 1. Create worktree
    worktreePath := filepath.Join(
        os.ExpandEnv("$HOME/.config/dev-swarm/worktrees"),
        codebase.Name,
        fmt.Sprintf("issue-%d", issue.Number),
    )
    
    err := git.CreateWorktree(codebase.LocalPath, worktreePath, branchName, codebase.DefaultBranch)
    if err != nil {
        return nil, fmt.Errorf("failed to create worktree: %w", err)
    }
    
    // 2. Build context for Claude
    context := BuildContext(issue, codebase, labels, currentLabel)
    
    // 3. Write context to prompt file
    promptFile := filepath.Join(worktreePath, ".dev-swarm-prompt.md")
    err = os.WriteFile(promptFile, []byte(context), 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to write prompt file: %w", err)
    }
    
    // 4. Create Claude command
    cmd := exec.Command("claude",
        "--print",
        "--prompt-file", promptFile,
    )
    cmd.Dir = worktreePath
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("DEV_SWARM_ISSUE=%d", issue.Number),
        fmt.Sprintf("DEV_SWARM_REPO=%s", codebase.Repo),
    )
    
    // 5. Set up pipes
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
    }
    
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
    }
    
    // 6. Create session
    session := &Session{
        ID:           sessionID,
        Issue:        issue,
        Codebase:     codebase,
        WorktreePath: worktreePath,
        BranchName:   branchName,
        cmd:          cmd,
        stdout:       stdout,
        stderr:       stderr,
        output:       NewOutputBuffer(m.outputBufferLines),
        Status:       StatusPending,
        stopChan:     make(chan struct{}),
    }
    
    // 7. Track session
    m.mu.Lock()
    m.sessions[sessionID] = session
    m.mu.Unlock()
    
    // 8. Start session
    err = session.Start(m.outputChan, m.statusChan)
    if err != nil {
        m.mu.Lock()
        delete(m.sessions, sessionID)
        m.mu.Unlock()
        return nil, fmt.Errorf("failed to start session: %w", err)
    }
    
    return session, nil
}
```

---

## Claude Integration

### Context Template

```go
// internal/session/context.go

package session

import (
    "fmt"
    "strings"
    "time"
    
    "dev-swarm/internal/config"
    "dev-swarm/internal/github"
)

func BuildContext(
    issue *github.Issue,
    codebase *config.Codebase,
    labels *config.Labels,
    currentLabel string,
) string {
    // Get the label config for current label
    labelCfg := labels.GetByName(currentLabel)
    
    var sb strings.Builder
    
    // Header
    sb.WriteString(fmt.Sprintf("# dev-swarm Task\n\n"))
    sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format(time.RFC3339)))
    
    // Repository info
    sb.WriteString("## Repository\n\n")
    sb.WriteString(fmt.Sprintf("- **Repo**: %s\n", codebase.Repo))
    sb.WriteString(fmt.Sprintf("- **Local Path**: %s\n", codebase.LocalPath))
    sb.WriteString(fmt.Sprintf("- **Default Branch**: %s\n", codebase.DefaultBranch))
    sb.WriteString(fmt.Sprintf("- **Working Branch**: claude/issue-%d\n", issue.Number))
    sb.WriteString("\n")
    
    // Issue info
    sb.WriteString("## Issue\n\n")
    sb.WriteString(fmt.Sprintf("- **Number**: #%d\n", issue.Number))
    sb.WriteString(fmt.Sprintf("- **Title**: %s\n", issue.Title))
    sb.WriteString(fmt.Sprintf("- **URL**: %s\n", issue.URL))
    sb.WriteString(fmt.Sprintf("- **Current Label**: %s\n", currentLabel))
    sb.WriteString("\n")
    
    // Issue body
    sb.WriteString("### Description\n\n")
    sb.WriteString(issue.Body)
    sb.WriteString("\n\n")
    
    // Comments (last 20)
    if len(issue.Comments) > 0 {
        sb.WriteString("### Comments\n\n")
        
        start := 0
        if len(issue.Comments) > 20 {
            start = len(issue.Comments) - 20
            sb.WriteString(fmt.Sprintf("*(Showing last 20 of %d comments)*\n\n", len(issue.Comments)))
        }
        
        for _, comment := range issue.Comments[start:] {
            isAI := strings.Contains(comment.Body, "<!-- dev-swarm:ai -->")
            author := comment.Author
            if isAI {
                author = fmt.Sprintf("%s (AI)", author)
            }
            
            sb.WriteString(fmt.Sprintf("**%s** (%s):\n", author, comment.CreatedAt.Format("2006-01-02 15:04")))
            sb.WriteString(comment.Body)
            sb.WriteString("\n\n---\n\n")
        }
    }
    
    // Current task
    sb.WriteString("## Your Task\n\n")
    sb.WriteString(fmt.Sprintf("**Current State**: %s\n\n", currentLabel))
    
    if labelCfg != nil && labelCfg.AIAction != "" {
        sb.WriteString("### Instructions\n\n")
        sb.WriteString(labelCfg.AIAction)
        sb.WriteString("\n\n")
    }
    
    // General instructions
    sb.WriteString("## Important Guidelines\n\n")
    sb.WriteString(`1. **Label Management**: Update labels using gh CLI:
   ` + "```bash" + `
   gh issue edit ` + fmt.Sprintf("%d", issue.Number) + ` --repo ` + codebase.Repo + ` --remove-label "current:label" --add-label "new:label"
   ` + "```" + `

2. **Comment Markers**: Wrap ALL your comments with these markers:
   ` + "```" + `
   <!-- dev-swarm:ai -->
   Your comment here
   <!-- /dev-swarm:ai -->
   ` + "```" + `

3. **Commit Messages**: Always reference the issue:
   - "Add feature X (#` + fmt.Sprintf("%d", issue.Number) + `)"
   - "Fix bug in Y (fixes #` + fmt.Sprintf("%d", issue.Number) + `)"

4. **Pull Requests**: Link to the issue:
   - Include "Closes #` + fmt.Sprintf("%d", issue.Number) + `" in the PR body

5. **Getting Stuck**: If you cannot proceed:
   - Change label to user:blocked
   - Add a comment explaining what's blocking you

6. **Code Quality**:
   - Follow existing code style
   - Write clear, maintainable code
   - Add tests when specified
`)
    
    return sb.String()
}
```

### Comment Marker Format

```markdown
<!-- dev-swarm:ai -->
🤖 **AI Response**

Your actual comment content goes here.

This can be multiple paragraphs.

- Bullet points
- Are fine

<!-- /dev-swarm:ai -->
```

The orchestrator detects AI comments by looking for `<!-- dev-swarm:ai -->` anywhere in the comment body.

---

## GitHub Operations

### Client Wrapper

```go
// internal/github/client.go

package github

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
)

type Client struct{}

func NewClient() *Client {
    return &Client{}
}

// Run executes a gh command and returns stdout
func (c *Client) Run(args ...string) (string, error) {
    cmd := exec.Command("gh", args...)
    output, err := cmd.Output()
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return "", fmt.Errorf("gh command failed: %s", string(exitErr.Stderr))
        }
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// RunJSON executes a gh command and parses JSON output
func (c *Client) RunJSON(result interface{}, args ...string) error {
    output, err := c.Run(args...)
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(output), result)
}
```

### Issue Operations

```go
// internal/github/issues.go

package github

import (
    "fmt"
    "time"
)

type Issue struct {
    Number    int        `json:"number"`
    Title     string     `json:"title"`
    Body      string     `json:"body"`
    State     string     `json:"state"`
    URL       string     `json:"url"`
    Labels    []Label    `json:"labels"`
    Comments  []Comment  `json:"comments"`
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt time.Time  `json:"updatedAt"`
}

type Label struct {
    Name  string `json:"name"`
    Color string `json:"color"`
}

type Comment struct {
    ID        int       `json:"id"`
    Author    string    `json:"author"`
    Body      string    `json:"body"`
    CreatedAt time.Time `json:"createdAt"`
}

func (i *Issue) ID() string {
    return fmt.Sprintf("#%d", i.Number)
}

func (i *Issue) HasLabel(name string) bool {
    for _, l := range i.Labels {
        if l.Name == name {
            return true
        }
    }
    return false
}

// ListIssuesWithLabel returns all open issues with a specific label
func (c *Client) ListIssuesWithLabel(repo, label string) ([]Issue, error) {
    var issues []Issue
    err := c.RunJSON(&issues,
        "issue", "list",
        "--repo", repo,
        "--label", label,
        "--state", "open",
        "--json", "number,title,body,state,url,labels,createdAt,updatedAt",
        "--limit", "100",
    )
    return issues, err
}

// GetIssue returns full issue details including comments
func (c *Client) GetIssue(repo string, number int) (*Issue, error) {
    var issue Issue
    err := c.RunJSON(&issue,
        "issue", "view", fmt.Sprintf("%d", number),
        "--repo", repo,
        "--json", "number,title,body,state,url,labels,comments,createdAt,updatedAt",
    )
    if err != nil {
        return nil, err
    }
    return &issue, nil
}

// UpdateIssueLabels changes labels on an issue
func (c *Client) UpdateIssueLabels(repo string, number int, removeLabels, addLabels []string) error {
    args := []string{"issue", "edit", fmt.Sprintf("%d", number), "--repo", repo}
    
    for _, label := range removeLabels {
        args = append(args, "--remove-label", label)
    }
    for _, label := range addLabels {
        args = append(args, "--add-label", label)
    }
    
    _, err := c.Run(args...)
    return err
}

// AddIssueComment adds a comment to an issue
func (c *Client) AddIssueComment(repo string, number int, body string) error {
    _, err := c.Run(
        "issue", "comment", fmt.Sprintf("%d", number),
        "--repo", repo,
        "--body", body,
    )
    return err
}
```

### Label Operations

```go
// internal/github/labels.go

package github

type LabelInfo struct {
    Name        string `json:"name"`
    Color       string `json:"color"`
    Description string `json:"description"`
}

// ListLabels returns all labels in a repo
func (c *Client) ListLabels(repo string) ([]LabelInfo, error) {
    var labels []LabelInfo
    err := c.RunJSON(&labels,
        "label", "list",
        "--repo", repo,
        "--json", "name,color,description",
        "--limit", "100",
    )
    return labels, err
}

// CreateLabel creates a new label
func (c *Client) CreateLabel(repo, name, color, description string) error {
    _, err := c.Run(
        "label", "create", name,
        "--repo", repo,
        "--color", color,
        "--description", description,
    )
    return err
}

// LabelExists checks if a label exists
func (c *Client) LabelExists(repo, name string) (bool, error) {
    labels, err := c.ListLabels(repo)
    if err != nil {
        return false, err
    }
    
    for _, l := range labels {
        if l.Name == name {
            return true, nil
        }
    }
    return false, nil
}
```

### PR Operations

```go
// internal/github/prs.go

package github

import (
    "fmt"
    "time"
)

type PullRequest struct {
    Number    int       `json:"number"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    State     string    `json:"state"`
    URL       string    `json:"url"`
    HeadRef   string    `json:"headRefName"`
    BaseRef   string    `json:"baseRefName"`
    Merged    bool      `json:"merged"`
    CreatedAt time.Time `json:"createdAt"`
}

type PRCheck struct {
    Name       string `json:"name"`
    Status     string `json:"status"`
    Conclusion string `json:"conclusion"`
}

// GetPRForBranch finds a PR for a specific branch
func (c *Client) GetPRForBranch(repo, branch string) (*PullRequest, error) {
    var prs []PullRequest
    err := c.RunJSON(&prs,
        "pr", "list",
        "--repo", repo,
        "--head", branch,
        "--state", "open",
        "--json", "number,title,body,state,url,headRefName,baseRefName,merged,createdAt",
    )
    if err != nil {
        return nil, err
    }
    
    if len(prs) == 0 {
        return nil, nil
    }
    return &prs[0], nil
}

// CreatePR creates a new pull request
func (c *Client) CreatePR(repo, title, body, head, base string) (*PullRequest, error) {
    output, err := c.Run(
        "pr", "create",
        "--repo", repo,
        "--title", title,
        "--body", body,
        "--head", head,
        "--base", base,
    )
    if err != nil {
        return nil, err
    }
    
    // Output is the PR URL, need to fetch details
    // This is simplified - in practice, parse the URL or use --json
    _ = output
    return c.GetPRForBranch(repo, head)
}

// MergePR merges a pull request
func (c *Client) MergePR(repo string, number int, deleteRemoteBranch bool) error {
    args := []string{
        "pr", "merge", fmt.Sprintf("%d", number),
        "--repo", repo,
        "--merge",
    }
    
    if deleteRemoteBranch {
        args = append(args, "--delete-branch")
    }
    
    _, err := c.Run(args...)
    return err
}

// GetPRChecks returns CI check status for a PR
func (c *Client) GetPRChecks(repo string, number int) ([]PRCheck, error) {
    var checks []PRCheck
    err := c.RunJSON(&checks,
        "pr", "checks", fmt.Sprintf("%d", number),
        "--repo", repo,
        "--json", "name,status,conclusion",
    )
    return checks, err
}

// PRChecksPassing returns true if all checks have passed
func (c *Client) PRChecksPassing(repo string, number int) (bool, error) {
    checks, err := c.GetPRChecks(repo, number)
    if err != nil {
        return false, err
    }
    
    if len(checks) == 0 {
        return true, nil // No checks configured
    }
    
    for _, check := range checks {
        if check.Status != "completed" {
            return false, nil // Still running
        }
        if check.Conclusion != "success" && check.Conclusion != "skipped" {
            return false, nil // Failed
        }
    }
    
    return true, nil
}

// GetMergedPRs returns recently merged PRs
func (c *Client) GetMergedPRs(repo string) ([]PullRequest, error) {
    var prs []PullRequest
    err := c.RunJSON(&prs,
        "pr", "list",
        "--repo", repo,
        "--state", "merged",
        "--json", "number,title,headRefName,merged",
        "--limit", "50",
    )
    return prs, err
}
```

---

## Git Worktree Management

```go
// internal/git/worktree.go

package git

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// CreateWorktree creates a new git worktree
func CreateWorktree(repoPath, worktreePath, branchName, baseBranch string) error {
    // Ensure parent directory exists
    err := os.MkdirAll(filepath.Dir(worktreePath), 0755)
    if err != nil {
        return fmt.Errorf("failed to create worktree parent directory: %w", err)
    }
    
    // Fetch latest from remote
    cmd := exec.Command("git", "fetch", "origin", baseBranch)
    cmd.Dir = repoPath
    if err := cmd.Run(); err != nil {
        // Non-fatal, continue anyway
    }
    
    // Check if branch already exists
    cmd = exec.Command("git", "rev-parse", "--verify", branchName)
    cmd.Dir = repoPath
    branchExists := cmd.Run() == nil
    
    if branchExists {
        // Branch exists, create worktree for existing branch
        cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
    } else {
        // Create new branch from base
        cmd = exec.Command("git", "worktree", "add", "-b", branchName, worktreePath, fmt.Sprintf("origin/%s", baseBranch))
    }
    
    cmd.Dir = repoPath
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed to create worktree: %s: %w", string(output), err)
    }
    
    return nil
}

// RemoveWorktree removes a git worktree and optionally the branch
func RemoveWorktree(repoPath, worktreePath string, deleteBranch bool) error {
    // Get branch name before removing worktree
    var branchName string
    if deleteBranch {
        cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
        cmd.Dir = worktreePath
        output, err := cmd.Output()
        if err == nil {
            branchName = strings.TrimSpace(string(output))
        }
    }
    
    // Remove the worktree
    cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
    cmd.Dir = repoPath
    if err := cmd.Run(); err != nil {
        // Try removing the directory directly if worktree remove fails
        os.RemoveAll(worktreePath)
    }
    
    // Prune worktree references
    cmd = exec.Command("git", "worktree", "prune")
    cmd.Dir = repoPath
    cmd.Run() // Ignore errors
    
    // Delete the branch if requested
    if deleteBranch && branchName != "" && branchName != "HEAD" {
        cmd = exec.Command("git", "branch", "-D", branchName)
        cmd.Dir = repoPath
        cmd.Run() // Ignore errors - branch might not exist locally
    }
    
    return nil
}

// WorktreeExists checks if a worktree exists
func WorktreeExists(worktreePath string) bool {
    info, err := os.Stat(worktreePath)
    if err != nil {
        return false
    }
    return info.IsDir()
}

// ListWorktrees returns all worktrees for a repo
func ListWorktrees(repoPath string) ([]string, error) {
    cmd := exec.Command("git", "worktree", "list", "--porcelain")
    cmd.Dir = repoPath
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var worktrees []string
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "worktree ") {
            path := strings.TrimPrefix(line, "worktree ")
            worktrees = append(worktrees, path)
        }
    }
    
    return worktrees, nil
}
```

---

## npm Distribution

### package.json

```json
{
  "name": "dev-swarm",
  "version": "0.1.0",
  "description": "AI-powered development orchestration using GitHub Issues",
  "keywords": [
    "ai",
    "claude",
    "github",
    "automation",
    "development",
    "orchestration"
  ],
  "homepage": "https://github.com/yourusername/dev-swarm",
  "bugs": {
    "url": "https://github.com/yourusername/dev-swarm/issues"
  },
  "license": "MIT",
  "author": "Your Name",
  "repository": {
    "type": "git",
    "url": "https://github.com/yourusername/dev-swarm.git"
  },
  "bin": {
    "dev-swarm": "./bin/dev-swarm"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "engines": {
    "node": ">=14"
  },
  "os": [
    "darwin",
    "linux"
  ],
  "cpu": [
    "x64",
    "arm64"
  ]
}
```

### install.js

```javascript
#!/usr/bin/env node

const os = require('os');
const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const VERSION = require('./package.json').version;
const REPO = 'yourusername/dev-swarm';

function getPlatform() {
  const platform = os.platform();
  const arch = os.arch();
  
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
  };
  
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64',
  };
  
  const p = platformMap[platform];
  const a = archMap[arch];
  
  if (!p || !a) {
    console.error(`Unsupported platform: ${platform}-${arch}`);
    process.exit(1);
  }
  
  return { platform: p, arch: a };
}

function getBinaryName() {
  const { platform, arch } = getPlatform();
  return `dev-swarm_${VERSION}_${platform}_${arch}.tar.gz`;
}

function getDownloadUrl() {
  const binaryName = getBinaryName();
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;
}

async function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    
    https.get(url, (response) => {
      // Handle redirects
      if (response.statusCode === 302 || response.statusCode === 301) {
        download(response.headers.location, dest).then(resolve).catch(reject);
        return;
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Download failed: ${response.statusCode}`));
        return;
      }
      
      response.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

async function install() {
  const binDir = path.join(__dirname, 'bin');
  const binaryPath = path.join(binDir, 'dev-swarm-binary');
  const tarPath = path.join(__dirname, 'dev-swarm.tar.gz');
  
  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }
  
  // Download binary
  console.log(`Downloading dev-swarm v${VERSION}...`);
  const url = getDownloadUrl();
  
  try {
    await download(url, tarPath);
  } catch (err) {
    console.error(`Failed to download: ${err.message}`);
    console.error(`URL: ${url}`);
    process.exit(1);
  }
  
  // Extract binary
  console.log('Extracting...');
  try {
    execSync(`tar -xzf "${tarPath}" -C "${binDir}"`, { stdio: 'inherit' });
    
    // Rename extracted binary
    const extractedName = `dev-swarm`;
    const extractedPath = path.join(binDir, extractedName);
    
    if (fs.existsSync(extractedPath)) {
      fs.renameSync(extractedPath, binaryPath);
    }
    
    // Make executable
    fs.chmodSync(binaryPath, '755');
    
    // Clean up
    fs.unlinkSync(tarPath);
    
    console.log('dev-swarm installed successfully!');
  } catch (err) {
    console.error(`Failed to extract: ${err.message}`);
    process.exit(1);
  }
}

install();
```

### bin/dev-swarm (shell wrapper)

```bash
#!/bin/sh

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$SCRIPT_DIR/dev-swarm-binary"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
  echo "Error: dev-swarm binary not found at $BINARY"
  echo "Try reinstalling: npm install -g dev-swarm"
  exit 1
fi

# Execute the binary with all arguments
exec "$BINARY" "$@"
```

### GoReleaser Config

```yaml
# .goreleaser.yaml

project_name: dev-swarm

before:
  hooks:
    - go mod tidy

builds:
  - id: dev-swarm
    main: ./cmd/dev-swarm
    binary: dev-swarm
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X dev-swarm/pkg/version.Version={{.Version}}
      - -X dev-swarm/pkg/version.Commit={{.Commit}}
      - -X dev-swarm/pkg/version.Date={{.Date}}

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'

release:
  github:
    owner: yourusername
    name: dev-swarm
  draft: false
  prerelease: auto
```

---

## Error Handling

### Error Types

```go
// internal/errors/errors.go

package errors

import (
    "errors"
    "fmt"
)

var (
    ErrConfigNotFound     = errors.New("config file not found")
    ErrConfigInvalid      = errors.New("config file is invalid")
    ErrAlreadyRunning     = errors.New("dev-swarm is already running")
    ErrNotRunning         = errors.New("dev-swarm is not running")
    ErrGHNotInstalled     = errors.New("gh CLI is not installed")
    ErrGHNotAuthenticated = errors.New("gh CLI is not authenticated")
    ErrClaudeNotInstalled = errors.New("claude CLI is not installed")
    ErrRepoNotFound       = errors.New("repository not found")
    ErrRepoNotAccessible  = errors.New("repository is not accessible")
    ErrPathNotFound       = errors.New("local path not found")
    ErrNotGitRepo         = errors.New("path is not a git repository")
    ErrWorktreeExists     = errors.New("worktree already exists")
    ErrWorktreeNotFound   = errors.New("worktree not found")
    ErrSessionExists      = errors.New("session already exists for this issue")
    ErrMaxSessionsReached = errors.New("maximum concurrent sessions reached")
)

type ConfigError struct {
    Field   string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config error: %s: %s", e.Field, e.Message)
}

type GitHubError struct {
    Operation string
    Repo      string
    Err       error
}

func (e *GitHubError) Error() string {
    return fmt.Sprintf("github error: %s on %s: %v", e.Operation, e.Repo, e.Err)
}

func (e *GitHubError) Unwrap() error {
    return e.Err
}
```

### Recovery Strategies

| Error | Recovery |
|-------|----------|
| Session process crashes | Log error, remove session from tracking, leave issue in current state |
| GitHub API rate limit | Back off exponentially, log warning |
| Network error | Retry with backoff, continue polling |
| Worktree creation fails | Log error, skip issue this cycle |
| Config file missing | Exit with helpful message |
| gh CLI not found | Exit with installation instructions |
| Label creation fails | Log warning, continue (label might exist with different settings) |

---

## Dependencies

### Go Modules

```go
// go.mod

module dev-swarm

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    gopkg.in/yaml.v3 v3.0.1
)
```

### External Dependencies

| Dependency | Purpose | Required |
|------------|---------|----------|
| `gh` | GitHub CLI for all GitHub operations | Yes |
| `claude` | Claude Code CLI for AI sessions | Yes |
| `git` | Git operations (worktrees, branches) | Yes |

---

## Build & Release

### Local Build

```bash
#!/bin/bash
# scripts/build.sh

set -e

VERSION=${1:-"dev"}
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X dev-swarm/pkg/version.Version=$VERSION"
LDFLAGS="$LDFLAGS -X dev-swarm/pkg/version.Commit=$COMMIT"
LDFLAGS="$LDFLAGS -X dev-swarm/pkg/version.Date=$DATE"

echo "Building dev-swarm $VERSION..."

go build -ldflags "$LDFLAGS" -o bin/dev-swarm ./cmd/dev-swarm

echo "Done: bin/dev-swarm"
```

### Release Process

```bash
#!/bin/bash
# scripts/release.sh

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 0.1.0"
    exit 1
fi

echo "Releasing v$VERSION..."

# Create git tag
git tag -a "v$VERSION" -m "Release v$VERSION"
git push origin "v$VERSION"

# GoReleaser will handle the rest via GitHub Actions
echo "Tag pushed. GitHub Actions will build and release."

# Update npm package version
cd npm
npm version "$VERSION" --no-git-tag-version
npm publish
cd ..

echo "npm package published."
echo "Release complete!"
```

---

## Summary

This specification defines `dev-swarm`, a tool that:

1. **Monitors GitHub issues** across multiple repositories using labels as state
2. **Spawns Claude Code sessions** to plan, implement, and fix code
3. **Uses git worktrees** for concurrent work on the same repository
4. **Provides a TUI** for real-time monitoring of all activity
5. **Auto-merges PRs** when users approve
6. **Distributes via npm** as a Go binary

The label-based workflow provides clear visibility into AI activity directly in GitHub, while the local orchestrator handles all the coordination without requiring any server infrastructure.
