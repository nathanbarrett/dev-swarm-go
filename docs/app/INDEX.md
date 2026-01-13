# dev-swarm-go Documentation

**AI-powered development orchestration using GitHub Issues as the control plane.**

## What is dev-swarm-go?

dev-swarm-go is a local daemon that monitors GitHub issues across multiple repositories and automatically spawns Claude Code sessions to plan, implement, and fix code based on label-driven workflows.

## Key Concepts

- **GitHub as Control Plane**: Issues and labels are the source of truth for work state
- **Labels as State Machine**: Each label represents a state in the development workflow
- **Local Execution**: All AI sessions run on your machine using Claude Code CLI
- **Git Worktrees**: Enable parallel work on multiple issues in the same repository

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](./architecture.md) | System components and how they interact |
| [Workflow](./workflow.md) | The label state machine and development flow |
| [Configuration](./configuration.md) | Config file structure and settings |
| [CLI Reference](./cli.md) | Command-line interface commands |
| [TUI Guide](./tui.md) | Terminal user interface and keyboard shortcuts |
| [Sessions](./sessions.md) | How Claude sessions are managed |
| [Distribution](./distribution.md) | npm packaging and release process |

## Quick Start Flow

```
1. Install: npm install -g dev-swarm-go
2. Initialize: dev-swarm-go init
3. Add repo: dev-swarm-go add --repo owner/repo --path ~/code/repo
4. Sync labels: dev-swarm-go sync-labels --all
5. Start: dev-swarm-go start
```

Then create issues in your repos and add labels to trigger AI work.

## The Basic Cycle

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  User creates   │────▶│  AI plans the   │────▶│  User reviews   │
│  issue + label  │     │  implementation │     │  the plan       │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                         │
                        ┌─────────────────┐              │ approve
                        │  PR merged,     │◀─────────────┘
                        │  issue done     │     ┌─────────────────┐
                        └─────────────────┘     │  AI implements  │
                                 ▲              │  and creates PR │
                                 │              └────────┬────────┘
                        ┌────────┴────────┐              │
                        │  User reviews   │◀─────────────┘
                        │  code + approves│
                        └─────────────────┘
```

## Prerequisites

- **gh CLI**: GitHub CLI, installed and authenticated
- **claude CLI**: Claude Code CLI, installed and authenticated
- **git**: For repository operations and worktrees
