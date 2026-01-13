package config

// DefaultSettings returns the default settings
func DefaultSettings() Settings {
	return Settings{
		PollInterval:          60,
		ActivePollInterval:    10,
		MaxConcurrentSessions: 5,
		AutoMergeOnApproval:   true,
		ApprovalKeywords: []string{
			"approved",
			"lgtm",
			"ship it",
			"merge it",
			"looks good",
		},
		OutputBufferLines: 1000,
	}
}

// DefaultLabels returns the default label configurations
func DefaultLabels() Labels {
	return Labels{
		ReadyToPlan: LabelConfig{
			Name:        "user:ready-to-plan",
			Color:       "0052CC",
			Description: "Ready for AI to create implementation plan",
			Owner:       "user",
			AIPickup:    "always",
			AIAction: `You are creating an implementation plan for this issue.

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
5. Change the label from ai:planning to user:plan-review`,
		},
		PlanReview: LabelConfig{
			Name:        "user:plan-review",
			Color:       "0052CC",
			Description: "Implementation plan ready for user review",
			Owner:       "user",
			AIPickup:    "on_user_comment",
			AIAction: `The user has commented on your implementation plan.

Check the user's comment:
- If it contains an approval keyword (approved, lgtm, ship it, etc.):
  → Change label from user:plan-review to user:ready-to-implement
- If it contains feedback, questions, or change requests:
  → Change label from user:plan-review to ai:planning
  → Revise the implementation plan based on the feedback
  → Add a new comment with the updated plan
  → Change label from ai:planning to user:plan-review`,
		},
		ReadyToImplement: LabelConfig{
			Name:        "user:ready-to-implement",
			Color:       "0052CC",
			Description: "Plan approved, ready for AI to implement",
			Owner:       "user",
			AIPickup:    "always",
			AIAction: `You are implementing an approved plan.

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
8. Change label from ai:implementing to user:code-review`,
		},
		CodeReview: LabelConfig{
			Name:        "user:code-review",
			Color:       "0052CC",
			Description: "PR created, awaiting user code review",
			Owner:       "user",
			AIPickup:    "on_user_comment",
			AIAction: `The user has commented on your Pull Request.

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
  → Reply to review comments explaining your changes`,
		},
		Blocked: LabelConfig{
			Name:        "user:blocked",
			Color:       "D93F0B",
			Description: "AI is stuck and needs human intervention",
			Owner:       "user",
			AIPickup:    "never",
		},
		Planning: LabelConfig{
			Name:        "ai:planning",
			Color:       "FBCA04",
			Description: "AI is writing the implementation plan",
			Owner:       "ai",
			AIPickup:    "never",
		},
		Implementing: LabelConfig{
			Name:        "ai:implementing",
			Color:       "FBCA04",
			Description: "AI is actively writing code",
			Owner:       "ai",
			AIPickup:    "never",
		},
		CIFailed: LabelConfig{
			Name:        "ai:ci-failed",
			Color:       "D93F0B",
			Description: "CI failed, AI will analyze and fix",
			Owner:       "ai",
			AIPickup:    "always",
			AIAction: `The CI pipeline has failed. You need to fix it.

Steps:
1. Change label from ai:ci-failed to ai:implementing
2. Fetch and analyze the CI logs using: gh run view --log-failed
3. Identify the cause of failure (test failures, build errors, lint errors)
4. Fix the issues in your code
5. Commit with a clear message: "Fix CI failures (#XX)"
6. Push the fix
7. Wait briefly for CI to start, then change label to user:code-review

Do NOT change unrelated code. Focus only on fixing the CI failure.`,
		},
		Done: LabelConfig{
			Name:        "ai:done",
			Color:       "0E8A16",
			Description: "Complete - PR merged",
			Owner:       "ai",
			AIPickup:    "never",
		},
	}
}

// DefaultAIInstructions returns the default AI instructions
func DefaultAIInstructions() AIInstructions {
	return AIInstructions{
		General: `You are an AI developer working on GitHub issues via the dev-swarm orchestrator.

## Important Guidelines

1. **Label Management**: Always update labels to reflect current state using gh CLI:
   ` + "```bash" + `
   gh issue edit {number} --remove-label "old:label" --add-label "new:label"
   ` + "```" + `

2. **Comment Markers**: Wrap ALL your comments with these markers so the system
   can distinguish AI comments from user comments:
   ` + "```" + `
   <!-- dev-swarm:ai -->
   Your comment content here
   <!-- /dev-swarm:ai -->
   ` + "```" + `

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
   - Write tests when specified in the plan`,
	}
}

// DefaultConfig returns a complete default configuration
func DefaultConfig() *Config {
	return &Config{
		Settings:       DefaultSettings(),
		Labels:         DefaultLabels(),
		AIInstructions: DefaultAIInstructions(),
		Codebases:      []Codebase{},
	}
}
