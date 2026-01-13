# Architecture

## System Overview

dev-swarm-go operates as a bridge between GitHub (the control plane) and your local machine (the execution plane).

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│   GITHUB (Control Plane)              LOCAL MACHINE (Execution Plane)    │
│                                                                          │
│   ┌───────────────────┐              ┌─────────────────────────────┐    │
│   │  Issues + Labels  │◀────────────▶│        dev-swarm-go         │    │
│   │  PR Comments      │   gh CLI     │                             │    │
│   │  CI Status        │              │  ┌───────────────────────┐  │    │
│   └───────────────────┘              │  │    Orchestrator       │  │    │
│                                      │  │    - Poller           │  │    │
│   Labels = State                     │  │    - State Machine    │  │    │
│   Comments = Communication           │  │    - Session Manager  │  │    │
│                                      │  └───────────┬───────────┘  │    │
│                                      │              │              │    │
│                                      │  ┌───────────▼───────────┐  │    │
│                                      │  │   Claude Sessions     │  │    │
│                                      │  │   (one per issue)     │  │    │
│                                      │  └───────────────────────┘  │    │
│                                      │              │              │    │
│                                      │  ┌───────────▼───────────┐  │    │
│                                      │  │         TUI           │  │    │
│                                      │  │   (monitoring)        │  │    │
│                                      │  └───────────────────────┘  │    │
│                                      └─────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### Orchestrator

The central coordinator that ties everything together:

- **Poller**: Periodically checks GitHub for issues with actionable labels
- **Pickup Logic**: Decides which issues to pick up based on label rules
- **State Tracker**: Maintains current state of all monitored issues
- **Event Dispatcher**: Sends updates to TUI and handles session events

### Session Manager

Manages the lifecycle of Claude Code sessions:

- **Spawning**: Creates new sessions with proper context
- **Tracking**: Monitors active sessions and their status
- **Output Capture**: Collects stdout/stderr from Claude processes
- **Cleanup**: Removes completed sessions and worktrees

### GitHub Client

Wrapper around the `gh` CLI for all GitHub operations:

- **Issues**: List, get, update labels, add comments
- **Labels**: Create, sync, check existence
- **Pull Requests**: Create, merge, check CI status
- **Comments**: Add with AI markers for identification

### Git Manager

Handles local git operations:

- **Worktrees**: Create isolated working directories per issue
- **Branches**: Create, checkout, delete feature branches
- **Sync**: Fetch latest changes from remote

### TUI (Terminal User Interface)

Real-time monitoring interface:

- **Codebase List**: Shows all monitored repositories
- **Issue Status**: Displays current label and session state
- **Output Panel**: Shows Claude session output
- **Status Bar**: Displays stats and controls

## Data Flow

### Polling Cycle

```
1. Poll Timer Fires
       │
       ▼
2. For Each Codebase:
   └─▶ Fetch issues with pickup labels
       │
       ▼
3. For Each Issue:
   └─▶ Check pickup rules
       └─▶ If should pickup:
           └─▶ Spawn Claude session
               │
               ▼
4. Check Active Sessions
   └─▶ Process completed sessions
       └─▶ Clean up if done
           │
           ▼
5. Check CI Status
   └─▶ Update labels if CI failed
       │
       ▼
6. Update TUI
```

### Session Lifecycle

```
1. Issue Detected (has pickup label)
       │
       ▼
2. Create Git Worktree
   └─▶ Branch: claude/issue-{number}
       │
       ▼
3. Build Claude Context
   └─▶ Issue details + comments
   └─▶ Label-specific instructions
       │
       ▼
4. Spawn Claude Process
   └─▶ Working dir = worktree
       │
       ▼
5. Capture Output
   └─▶ Stream to TUI
       │
       ▼
6. Session Completes
   └─▶ Check exit code
   └─▶ Update internal state
       │
       ▼
7. Cleanup (if issue done)
   └─▶ Remove worktree
```

## File Locations

| Path | Purpose |
|------|---------|
| `~/.config/dev-swarm-go/config.yaml` | Main configuration |
| `~/.config/dev-swarm-go/dev-swarm-go.lock` | Process lock file |
| `~/.config/dev-swarm-go/dev-swarm-go.log` | Log file (daemon mode) |
| `~/.config/dev-swarm-go/worktrees/` | Git worktrees directory |

## Concurrency Model

- **Single Orchestrator**: One polling loop, serialized GitHub checks
- **Parallel Sessions**: Multiple Claude sessions can run simultaneously
- **Thread-Safe State**: Mutex-protected access to shared state
- **Buffered Channels**: Non-blocking communication between components
