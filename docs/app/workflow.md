# Workflow

## Label State Machine

dev-swarm-go uses GitHub labels as a state machine. Labels indicate who owns the next action (user or AI) and what that action should be.

### Label Definitions

| Label | Owner | AI Pickup | Description |
|-------|-------|-----------|-------------|
| `user:ready-to-plan` | User | Always | Issue ready for AI to create implementation plan |
| `ai:planning` | AI | Never | AI is actively writing the implementation plan |
| `user:plan-review` | User | On Comment | Plan complete, user is reviewing |
| `user:ready-to-implement` | User | Always | Plan approved, ready for AI to code |
| `ai:implementing` | AI | Never | AI is actively writing code |
| `user:code-review` | User | On Comment | PR created, user is reviewing code |
| `ai:ci-failed` | AI | Always | CI pipeline failed, AI will fix |
| `user:blocked` | User | Never | AI is stuck, needs human intervention |
| `ai:done` | AI | Never | Complete, PR merged |

### Ownership Convention

- **`user:` prefix**: Waiting for human action (user owns the next step)
- **`ai:` prefix**: AI is working or AI owns the state

### Label Colors

| Color | Hex | Usage |
|-------|-----|-------|
| Blue | `#0052CC` | User-owned states |
| Yellow | `#FBCA04` | AI-owned active states |
| Red | `#D93F0B` | Error/blocked states |
| Green | `#0E8A16` | Done state |

## State Flow

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

## Pickup Rules

The orchestrator decides whether to spawn an AI session based on pickup rules:

| Label | Pickup Rule | Behavior |
|-------|-------------|----------|
| `user:ready-to-plan` | Always | Spawn immediately if capacity available |
| `ai:planning` | Never | Already has active session |
| `user:plan-review` | On User Comment | Only if new user comment since last AI comment |
| `user:ready-to-implement` | Always | Spawn immediately if capacity available |
| `ai:implementing` | Never | Already has active session |
| `user:code-review` | On User Comment | Only if new user comment/review since last AI comment |
| `ai:ci-failed` | Always | Spawn immediately to fix |
| `user:blocked` | Never | Requires human intervention |
| `ai:done` | Never | Work complete |

### Comment Detection

For conditional pickup labels (`user:plan-review` and `user:code-review`), the system looks for comments without the AI marker:

```
<!-- dev-swarm-go:ai -->
```

If the most recent comment lacks this marker and was created after the last AI comment, a new session is spawned.

## State Transitions

| From | To | Trigger |
|------|-----|---------|
| `user:ready-to-plan` | `ai:planning` | AI picks up issue |
| `ai:planning` | `user:plan-review` | AI completes plan |
| `user:plan-review` | `user:ready-to-implement` | User approves |
| `user:plan-review` | `ai:planning` | User provides feedback |
| `user:ready-to-implement` | `ai:implementing` | AI picks up issue |
| `ai:implementing` | `user:code-review` | AI creates PR |
| `user:code-review` | `ai:done` | User approves, PR merged |
| `user:code-review` | `ai:implementing` | User requests changes |
| `ai:ci-failed` | `ai:implementing` | AI picks up to fix |
| Any | `user:blocked` | AI cannot proceed |
| Any | `ai:ci-failed` | CI fails |

## Approval Keywords

The system recognizes these keywords as approval (case-insensitive):
- "approved"
- "lgtm"
- "ship it"
- "merge it"
- "looks good"

## Typical Development Cycle

1. **User Creates Issue**: Describe the feature, bug, or task
2. **User Adds Label**: Add `user:ready-to-plan` to start
3. **AI Plans**: Claude analyzes and writes implementation plan
4. **User Reviews Plan**: Approve or provide feedback
5. **AI Implements**: Claude writes code, creates PR
6. **User Reviews Code**: Approve, request changes, or provide feedback
7. **PR Merged**: Issue marked done, worktree cleaned up
