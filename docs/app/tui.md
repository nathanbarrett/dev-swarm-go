# TUI Guide

## Layout

The terminal interface uses a split-view layout:

```
╭─ dev-swarm-go v0.1.0 ────────────────────────────────────────────────────────╮
│                                                                              │
│  my-web-app • github.com/owner/my-web-app                                    │
│  ├─ #42 Add dark mode support                                                │
│  │   └─ ai:implementing ● active (3m 12s)                     ◀── selected  │
│  └─ #45 Fix Safari login bug                                                 │
│      └─ user:plan-review ○ waiting                                           │
│                                                                              │
│  api-service • github.com/owner/api-service                                  │
│  ├─ #12 Add rate limiting                                                    │
│  │   └─ user:code-review ○ waiting                                           │
│  └─ (no other active issues)                                                 │
│                                                                              │
│  mobile-app • github.com/owner/mobile-app                                    │
│  └─ (idle)                                                                   │
│                                                                              │
├─ Output: #42 Add dark mode support ──────────────────────────────────────────┤
│                                                                              │
│  [14:32:01] Starting session for issue #42                                   │
│  [14:32:01] Worktree: ~/.config/dev-swarm-go/worktrees/my-web-app/issue-42   │
│  [14:32:02] Reading issue details...                                         │
│  [14:32:03] Found implementation plan in comments                            │
│  [14:32:05] Creating branch claude/issue-42 from main...                     │
│  [14:32:06] ✓ Branch created                                                 │
│  [14:32:08] Implementing dark mode support...                                │
│  [14:32:30] █                                                                │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│  Active: 1   Queued: 0   Waiting: 2   │  Poll: 47s  │  ↑↓ Nav  Q Quit       │
╰──────────────────────────────────────────────────────────────────────────────╯
```

## Panels

### Top Panel - Issue List

Displays all codebases and their active issues:
- Codebase name and GitHub URL
- Issue numbers and titles (tree structure)
- Current label and session status
- Selected item highlighted

### Middle Panel - Output

Shows Claude session output for the selected issue:
- Timestamped log lines
- stdout and stderr combined
- Auto-scrolls to latest output
- Scrollable history

### Bottom Panel - Status Bar

Quick reference information:
- Session counts (Active, Queued, Waiting)
- Next poll countdown
- Keyboard shortcut hints

## Status Icons

| Icon | Meaning | Color |
|------|---------|-------|
| `●` | Active session (Claude running) | Yellow |
| `◆` | Queued (about to start) | Cyan |
| `○` | Waiting (needs user action) | Blue |
| `✗` | Failed/Blocked | Red |
| `✓` | Done | Green |

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Focus selected session output (full screen) |
| `Esc` | Return to split view |

### Controls

| Key | Action |
|-----|--------|
| `r` | Force refresh (poll now) |
| `p` | Pause/resume polling |
| `l` | Toggle log panel visibility |
| `q` | Quit |
| `?` | Show help |

### Output Panel

| Key | Action |
|-----|--------|
| `Page Up` | Scroll output up |
| `Page Down` | Scroll output down |
| `Home` | Jump to top of output |
| `End` | Jump to bottom of output |

## Color Scheme

The TUI uses colors that match the GitHub label colors:

| Color | Usage |
|-------|-------|
| Blue (#0052CC) | User-owned labels, codebase names |
| Yellow (#FBCA04) | AI-active states |
| Red (#D93F0B) | Error states, blocked |
| Green (#0E8A16) | Done states, success messages |
| Gray (#6B7280) | Borders, secondary text |
| White | Primary text, issue titles |

## Session States Display

| State | Display |
|-------|---------|
| Pending | Label only, no icon |
| Running | `● active (duration)` |
| Completed | `✓ done` |
| Failed | `✗ failed` |
| Queued | `◆ queued` |
| Waiting | `○ waiting` |

## Polling Indicator

The status bar shows time until next GitHub poll:
- Updates every second
- Resets after manual refresh (`r`)
- Shows "Paused" when polling is disabled
- Uses shorter interval when sessions are active
