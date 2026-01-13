# Sessions

## Session Lifecycle

### 1. Issue Detection

The orchestrator polls GitHub for issues with actionable labels:
- Issues with pickup labels are candidates
- Conditional labels require new user comments
- Skipped if session already exists for issue

### 2. Worktree Creation

Each session gets an isolated working directory:
- Location: `~/.config/dev-swarm-go/worktrees/{repo-name}/issue-{number}/`
- Branch: `claude/issue-{number}`
- Created from default branch (usually `main`)

**Branch behavior:**
- If branch exists: worktree checks out existing branch
- If new: branch created from `origin/{default_branch}`

### 3. Context Building

Claude receives context about the task:
- Repository information (repo, local path, branches)
- Issue details (number, title, description, URL)
- Comment history (last 20 comments)
- Current label and expected action
- General guidelines for the workflow

### 4. Process Spawning

Claude CLI is invoked with:
- Working directory set to worktree
- Context passed via prompt file
- stdout/stderr captured for TUI display
- Environment variables for issue/repo info

### 5. Output Capture

Session output is:
- Streamed to TUI in real-time
- Buffered (configurable line limit)
- Timestamped per line
- Separated by stream (stdout/stderr)

### 6. Completion

When Claude process exits:
- Exit code checked (0 = success)
- Session status updated
- TUI notified
- Session tracking removed

### 7. Cleanup

Worktrees are cleaned up when:
- Issue label is `ai:done`
- PR is merged
- Manual cleanup command run

## Session Manager

The session manager coordinates all Claude sessions:

### Responsibilities
- Track active sessions
- Enforce concurrency limits
- Route output to TUI
- Handle session completion

### Concurrency Control
- Maximum sessions configurable (`max_concurrent_sessions`)
- New sessions queued if at capacity
- Oldest queued issues processed first

### Session Identification
Each session has a unique ID: `{owner}/{repo}#{issue_number}`

## Claude Context

### What Claude Receives

1. **Repository info**
   - GitHub repo path
   - Local filesystem path
   - Default branch name
   - Working branch name

2. **Issue details**
   - Issue number and URL
   - Title and description
   - Current label
   - Comment history

3. **Task instructions**
   - Label-specific actions
   - General workflow guidelines

### Comment Markers

Claude must wrap all comments with markers:
```
<!-- dev-swarm-go:ai -->
Comment content here
<!-- /dev-swarm-go:ai -->
```

This allows the system to distinguish AI comments from user comments for pickup logic.

## Git Worktrees

### Why Worktrees

Worktrees enable:
- Parallel work on multiple issues in same repo
- Isolated changes per issue
- No branch switching in main clone

### Worktree Structure

```
~/.config/dev-swarm-go/worktrees/
├── my-web-app/
│   ├── issue-42/          # Full checkout
│   │   ├── .git           # Link to main repo
│   │   ├── src/
│   │   └── ...
│   └── issue-45/
└── api-service/
    └── issue-12/
```

### Branch Naming

All AI branches follow the pattern: `claude/issue-{number}`

This makes it easy to:
- Identify AI-created branches
- Associate branches with issues
- Clean up after completion

## Error Handling

### Session Failures

| Scenario | Handling |
|----------|----------|
| Process crash | Log error, remove from tracking |
| Exit code non-zero | Mark as failed, leave issue state |
| Worktree creation fails | Skip issue this poll cycle |

### Recovery

- Failed sessions don't block other work
- Issues remain in current state for retry
- Manual intervention via `user:blocked` label

## Output Buffer

Each session maintains a rolling buffer of output:
- Configurable size (`output_buffer_lines`)
- Oldest lines dropped when full
- Timestamp and stream type preserved
- Available via TUI even after session ends
