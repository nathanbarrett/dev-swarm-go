# dev-swarm

AI-powered development orchestration using GitHub Issues as the control plane.

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

## Installation

### Via npm (recommended)

```bash
npm install -g dev-swarm
```

### From source

```bash
git clone https://github.com/nathanbarrett/dev-swarm-go.git
cd dev-swarm-go
./scripts/build.sh
```

## Prerequisites

- [GitHub CLI](https://cli.github.com) (`gh`) - installed and authenticated
- [Claude Code](https://claude.ai/claude-code) (`claude`) - installed and authenticated
- Git

## Quick Start

```bash
# Initialize configuration
dev-swarm init

# Add a repository
dev-swarm add --repo owner/my-repo --path ~/code/my-repo

# Sync labels to the repository
dev-swarm sync-labels owner/my-repo

# Start monitoring
dev-swarm start
```

## Configuration

Configuration is stored at `~/.config/dev-swarm/config.yaml`.

```yaml
settings:
  poll_interval: 60              # Seconds between checks when idle
  active_poll_interval: 10       # Seconds between checks when sessions active
  max_concurrent_sessions: 5     # Maximum simultaneous Claude sessions
  auto_merge_on_approval: true   # Automatically merge PRs on approval

codebases:
  - name: my-web-app
    repo: owner/my-web-app
    local_path: ~/code/my-web-app
    default_branch: main
    enabled: true
```

## CLI Commands

```bash
dev-swarm start           # Start with TUI
dev-swarm start --daemon  # Start in background
dev-swarm stop            # Stop the orchestrator
dev-swarm status          # Show current status
dev-swarm logs            # View log file

dev-swarm init            # Create config file
dev-swarm add             # Add a repository
dev-swarm remove          # Remove a repository
dev-swarm list            # List repositories
dev-swarm sync-labels     # Create labels in repos
dev-swarm cleanup         # Remove orphaned worktrees
```

## Label Workflow

| Label | Description |
|-------|-------------|
| `user:ready-to-plan` | Issue ready for AI to create implementation plan |
| `ai:planning` | AI is actively writing the plan |
| `user:plan-review` | Plan complete, waiting for user review |
| `user:ready-to-implement` | Plan approved, ready for AI to code |
| `ai:implementing` | AI is actively writing code |
| `user:code-review` | PR created, waiting for code review |
| `ai:ci-failed` | CI failed, AI will analyze and fix |
| `user:blocked` | AI is stuck, needs human help |
| `ai:done` | Complete, PR merged |

## TUI Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `Enter` | Focus session output |
| `Esc` | Return to split view |
| `r` | Force refresh |
| `p` | Pause/resume |
| `q` | Quit |
| `?` | Help |

## License

MIT License - see [LICENSE](LICENSE)
