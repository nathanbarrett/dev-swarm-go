# CLI Reference

## Commands Overview

| Command | Description |
|---------|-------------|
| `start` | Start the orchestrator |
| `init` | Initialize configuration |
| `add` | Add a codebase to config |
| `remove` | Remove a codebase from config |
| `list` | List configured codebases |
| `sync-labels` | Sync labels to repositories |
| `status` | Show current status |
| `logs` | View log output |
| `stop` | Stop the running orchestrator |
| `cleanup` | Clean up orphaned worktrees |
| `version` | Show version information |

## Main Commands

### start

Start the orchestrator with TUI monitoring.

| Flag | Description |
|------|-------------|
| `--config` | Path to config file (default: ~/.config/dev-swarm-go/config.yaml) |
| `--daemon` | Run in background mode (no TUI, logs to file) |

**Process flow:**
1. Load and validate configuration
2. Check lock file (ensure not already running)
3. Verify dependencies (gh, claude, git)
4. Validate codebase paths
5. Sync labels to all enabled repos
6. Start TUI (or logging if daemon mode)
7. Begin polling loop

### stop

Stop the running orchestrator gracefully.

**Process flow:**
1. Read PID from lock file
2. Send termination signal
3. Wait for graceful shutdown
4. Sessions are allowed to complete current operations

## Setup Commands

### init

Initialize configuration interactively. Creates:
- Config directory at ~/.config/dev-swarm-go/
- Default config.yaml with all settings
- Worktrees directory

### add

Add a new codebase to the configuration.

| Flag | Required | Description |
|------|----------|-------------|
| `--repo` | Yes | GitHub repo (owner/name format) |
| `--path` | Yes | Local clone path |
| `--branch` | No | Default branch (default: main) |
| `--name` | No | Friendly name for display |

### remove

Remove a codebase from configuration.

| Argument | Description |
|----------|-------------|
| `name` | Codebase name or repo identifier |

### list

Display all configured codebases with their status:
- Name and repo
- Local path
- Default branch
- Enabled/disabled status

### sync-labels

Ensure required labels exist on GitHub repositories.

| Flag | Description |
|------|-------------|
| `--all` | Sync to all enabled repos |

Without `--all`, specify a repo: `dev-swarm-go sync-labels owner/repo`

**Process flow:**
1. Fetch existing labels from repo
2. Compare with required labels from config
3. Create any missing labels
4. Log results (created/skipped)

## Status & Monitoring

### status

Show current orchestrator status:
- Is orchestrator running?
- Active sessions (count and details)
- Pending issues
- Last poll time

### logs

View orchestrator logs.

| Flag | Description |
|------|-------------|
| `--lines` | Number of lines to show (default: 50) |
| `--follow` | Tail logs continuously |

## Maintenance

### cleanup

Clean up orphaned worktrees that are no longer associated with active issues.

**Process flow:**
1. List all worktrees in config directory
2. Check if associated issue/PR is still active
3. Remove worktrees for completed or closed issues
4. Prune git worktree references

### version

Display version information:
- Version number
- Commit hash
- Build date

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Already running |
| 4 | Not running (for stop command) |
| 5 | Missing dependency |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DEV_SWARM_CONFIG` | Override config file path |
| `DEV_SWARM_LOG_LEVEL` | Set log level (debug, info, warn, error) |
