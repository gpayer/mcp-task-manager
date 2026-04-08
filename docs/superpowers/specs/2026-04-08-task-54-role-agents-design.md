# Task 54 Design: Repo-Local Planner, Coder, and Reviewer Agents

## Summary

Task 54 adds three repo-local agents under `agents/` and updates the `superpowers-workflow` skill to use them directly instead of relying only on generic references to `subagent-driven-development`.

The goal is not just to define these agents, but to make them the default execution path of the workflow. The workflow should prefer the repo-local role agents, degrade only through an explicit fallback chain, and require user confirmation before any fallback is used.

## Goals

- Add repo-local `planner`, `coder`, and `reviewer` agents.
- Give each agent strict role boundaries and clear task expectations.
- Update `skills/superpowers-workflow/SKILL.md` to dispatch these agents directly by phase.
- Define a fallback chain when a preferred role agent is unavailable.
- Require explicit user clarification before using any fallback.
- Update `.codex/INSTALL.md` so both repo-local skills and repo-local agents are installed for Codex discovery.

## Non-Goals

- Replacing the broader superpowers skill system.
- Removing compatibility with generic or default agents entirely.
- Redesigning the full `subagent-driven-development` skill.
- Introducing new production dependencies.

## Design

## Repository Structure

Add a top-level `agents/` directory:

- `agents/planner.md`
- `agents/coder.md`
- `agents/reviewer.md`

These files are owned by this repository and are part of the plugin installation surface, alongside the existing `skills/` directory.

## Agent Responsibilities

### Planner

The planner is responsible for decomposition and planning only.

It should:

- Read the current task and relevant project context.
- Identify missing requirements or ambiguity before planning.
- Use `writing-plans` when converting approved task intent into executable subtasks or plan structure.
- Produce implementation-ready subtasks or task descriptions.
- Keep scope aligned with the parent task.

It must not:

- Implement code changes.
- Perform code review as a substitute for planning.
- Silently invent requirements to fill gaps.
- Expand scope beyond the task without surfacing it.

### Coder

The coder is responsible for implementation only.

It should:

- Execute the assigned task within the provided spec.
- Use relevant superpowers execution skills for the work, especially `test-driven-development`, `systematic-debugging`, and `verification-before-completion` when applicable.
- Report `DONE`, `DONE_WITH_CONCERNS`, `NEEDS_CONTEXT`, or `BLOCKED` clearly.
- Stay inside the task boundary and request clarification when the spec is insufficient.

It must not:

- Re-plan the task unless blocked by missing or contradictory requirements.
- Expand the task into adjacent improvements without approval.
- Act as the final reviewer of its own work.

### Reviewer

The reviewer is responsible for validation only.

It should:

- Verify spec compliance first.
- Verify code quality second.
- Use existing superpowers review patterns and prompt structures where appropriate.
- Return concrete, evidence-based findings with severity and file references.

It must not:

- Rewrite the implementation as part of the review.
- Quietly accept spec drift.
- Turn review into a new planning pass unless it finds a genuine plan defect.

## Workflow Changes

`skills/superpowers-workflow/SKILL.md` should be updated so it no longer delegates only through generic wording such as "use subagent-driven-development flow." Instead, it should directly describe the role dispatch sequence used by this repository.

### Planning Phase

When a parent task has no subtasks:

1. Start the parent task.
2. Dispatch the repo-local `planner` agent.
3. The planner reads context, uses `writing-plans` guidance, and creates executable subtasks through the task manager.
4. The workflow verifies the subtasks exist before continuing.

### Execution Phase

When a subtask is ready to execute:

1. Start the subtask.
2. Dispatch the repo-local `coder` agent with the full subtask description and relevant task context.
3. Wait for the coder result and handle status correctly.
4. Dispatch the repo-local `reviewer` agent for spec compliance review.
5. If spec review passes, dispatch the repo-local `reviewer` agent again, or the same role in code-quality mode, for code quality review.
6. Only complete the subtask after both review stages pass.

The skill should be explicit that the local workflow controller owns phase transitions and completion, while the role agents own the work within each phase.

## Fallback Policy

The workflow should define a strict fallback ladder for each role:

1. Repo-local role agent from this repository.
2. Another available role-appropriate agent for that phase.
3. Default agent.

The workflow must not silently fall back. If the preferred repo-local agent cannot be used, the user must be asked which fallback to allow before dispatch continues.

This applies independently per role:

- planning fallback for `planner`
- implementation fallback for `coder`
- review fallback for `reviewer`

The clarification step should make the downgrade explicit so the user understands the workflow is leaving the preferred repository-specific guardrails.

## Installation Changes

`.codex/INSTALL.md` should be updated so installation exposes both:

- repo `skills/`
- repo `agents/`

The instructions should show the expected discovery paths and verification commands for both. The install flow should remain clone plus symlink/junction based, mirroring the current skills setup.

## Error Handling

### Missing Preferred Agent

If a repo-local role agent is unavailable:

- Stop before dispatch.
- Tell the user which preferred agent could not be used.
- Offer the next fallback choice.
- Continue only after the user confirms.

### Missing Context

If the planner or coder reports missing context:

- Do not force forward progress.
- Surface the missing information clearly.
- Ask the user for clarification or provide controller-level context if available.

### Review Failure

If the reviewer finds issues:

- Return the issues to the coder.
- Re-run the relevant review stage after fixes.
- Keep the current workflow rule that unresolved issues block completion.

## Testing

The implementation should include coverage for:

- install documentation updates for agents
- workflow instructions referencing role agents directly
- fallback behavior wording requiring user confirmation
- any parser or fixture expectations in tests that assume only `skills/` are installed or referenced

If automated tests exist for skill or plugin content discovery, extend them to cover `agents/`.

## Open Questions Resolved

- Agent location: use top-level `agents/`.
- Planner behavior: explicitly reference `writing-plans`.
- Fallback behavior: require user clarification before any downgrade from repo-local agents.

## Recommended Implementation Order

1. Add the repo-local agent files with role-specific instructions.
2. Refine `skills/superpowers-workflow/SKILL.md` to dispatch those agents directly.
3. Add fallback-chain instructions and explicit user-confirmation gates.
4. Update `.codex/INSTALL.md` to install both skills and agents.
5. Add or update tests for discovery and workflow documentation if such tests exist.
