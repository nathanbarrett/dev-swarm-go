# Configuration

## Config File Location

All configuration is stored in a single directory:

```
~/.config/dev-swarm-go/
├── config.yaml              # Main configuration file
├── dev-swarm-go.lock        # PID lock file (created at runtime)
├── dev-swarm-go.log         # Log file (daemon mode)
└── worktrees/               # Git worktrees directory
    ├── {repo-name}/
    │   ├── issue-{number}/
    │   └── issue-{number}/
    └── {repo-name}/
        └── issue-{number}/
```

## Configuration Structure

### Settings

Global settings that control orchestrator behavior:

| Setting | Default | Description |
|---------|---------|-------------|
| `poll_interval` | 60 | Seconds between GitHub checks when idle |
| `active_poll_interval` | 10 | Seconds between checks when sessions are active |
| `max_concurrent_sessions` | 5 | Maximum simultaneous Claude sessions |
| `auto_merge_on_approval` | true | Automatically merge PR when user approves |
| `approval_keywords` | See below | Keywords that trigger approval |
| `output_buffer_lines` | 1000 | Number of output lines to keep per session |

Default approval keywords:
- "approved"
- "lgtm"
- "ship it"
- "merge it"
- "looks good"

### Labels

Labels can be customized globally or per-codebase. Each label has:

| Field | Description |
|-------|-------------|
| `name` | The label text shown on GitHub |
| `color` | Hex color code (without #) |
| `description` | Label description for GitHub |
| `owner` | Either "user" or "ai" |
| `ai_pickup` | Pickup rule: "always", "never", or "on_user_comment" |
| `ai_action` | Instructions for Claude when this label is active |

### Codebases

Each monitored repository requires:

| Field | Required | Description |
|-------|----------|-------------|
| `name` | No | Friendly name (defaults to repo name) |
| `repo` | Yes | GitHub repo in `owner/name` format |
| `local_path` | Yes | Local clone path (~ expanded) |
| `default_branch` | Yes | Branch to create PRs against |
| `enabled` | No | Set to false to disable (default: true) |
| `labels` | No | Per-repo label overrides |

### AI Instructions

Global instructions included in every Claude session context:

- Label management guidelines
- Comment marker requirements
- Commit message format
- PR linking conventions
- Code quality expectations

## Validation Rules

1. **Required fields**:
   - `codebases[].repo` - must be in `owner/name` format
   - `codebases[].local_path` - must be a valid path
   - `codebases[].default_branch` - must be specified

2. **Path expansion**: `~` is expanded to home directory

3. **Label merging**: Per-codebase labels merge with global labels. Only specified fields are overridden.

4. **Numeric limits**:
   - `poll_interval`: 1-3600 seconds
   - `max_concurrent_sessions`: 1-20

## Initialization

Running `dev-swarm-go init` creates the config directory and a default config file with:
- All default settings
- All standard labels with default colors and descriptions
- An example codebase entry (commented out)
- Global AI instructions

## Label Sync

The `sync-labels` command ensures all required labels exist on GitHub repos:
- Creates missing labels
- Does not modify existing labels
- Can target a single repo or all enabled repos
