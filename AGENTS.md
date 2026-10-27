# Stricture Agents

This repository defines autonomous execution agents for implementation flow.

## Phase Foreman Agent

Name: `phase-foreman`

Purpose:
- Own end-to-end delegation for phases 3, 4, and 5.
- Start phase `N+1` immediately after phase `N` completes.
- Avoid user interaction unless blocked by a real, non-recoverable question.

Operating rules:
- Do not ask for confirmation between phases.
- Follow strict order: `3 -> 4 -> 5`.
- A phase is complete only when its phase gate command passes.
- If blocked, surface one concise blocker with exact failing command/output.

Completion gates:
- Phase 3: `make test-phase3`
- Phase 4: `make test-phase4`
- Phase 5: `make test-phase5`

Delegation policy:
- Phase 3 delegate scope: `TQ-*` rules and required supporting engine/context work.
- Phase 4 delegate scope: `CTR-*` rules, manifest support, contract checks.
- Phase 5 delegate scope: Python/Java adapters, SARIF/JUnit reporters, changed/staged and cache hardening.

Execution entrypoint:
- Use `./scripts/phase-agent.sh run`.
- Status: `./scripts/phase-agent.sh status`.
- Reset: `./scripts/phase-agent.sh reset`.

Escalation threshold (the only time to ask user):
- Missing/contradictory product requirement.
- External dependency or credential that cannot be resolved locally.
- Destructive or policy-sensitive action not already approved.

## Overseer Agent

Name: `overseer-agent`

Purpose:
- Keep issuing a "what's next" directive until the tool is complete.
- If tests/checks fail, direct the next action to fix code quality and failing checks.
- If tests/checks pass, require a deeper spec-vs-tests-vs-code quality audit and keep raising the bar.

Operating rules:
- Runtime budget defaults to 5 hours per overseer run.
- Loop until completion or timeout, without asking the user for routine confirmations.
- Prompt the user only for truly blocking questions that cannot be resolved locally.
- Persist state so progress survives interruptions.

Completion criteria:
- `make ci` passes.
- `./scripts/spec-quality-audit.sh` passes with zero deficiencies.

What "done" means:
- Tests pass, validation gates pass, and the audit reports no spec/tests/code quality gaps.
- The final prompt still requests an independent high-bar review pass (top-1% quality objective).

Execution entrypoint:
- Run loop: `./scripts/overseer-agent.sh run`
- One cycle: `./scripts/overseer-agent.sh once`
- Status: `./scripts/overseer-agent.sh status`
- Reset: `./scripts/overseer-agent.sh reset`

## Usecase Agent

Name: `usecase-agent`

Purpose:
- Build and validate realistic, combined lineage examples across:
  - logistics
  - fintech
  - media
  - ecommerce
  - corporate governance
- Keep generated catalog, fixture annotations, and fake API data synchronized.
- Enforce at least 50 API flows with multi-use-case combinations per flow.

Operating rules:
- Regenerate artifacts from source definitions before validation.
- Require lineage parsing and drift checks to pass.
- Run fake API unit tests on every usecase-agent cycle.
- Validate docker-compose configuration when Docker is available.
- Run live docker-compose smoke checks when Docker daemon and cached base image are available.
- Treat unavailable Docker runtime as non-blocking and continue with explicit `SKIP` output.
- Avoid user interaction unless a non-recoverable local blocker occurs.

Execution entrypoint:
- Run: `./scripts/usecase-agent.sh run`
- Status: `./scripts/usecase-agent.sh status`
- Reset: `./scripts/usecase-agent.sh reset`

## Publication Agent

Name: `publication-agent`

Purpose:
- Scaffold and maintain publication-facing documentation from `PUB-ROADMAP.md`.
- Keep `docs/publication/` ready for continuous authoring with minimal setup friction.
- Avoid user prompts for routine scaffolding and state tracking tasks.

Operating rules:
- Create missing publication docs from deterministic templates.
- Never overwrite existing publication docs unless explicitly told.
- Persist run state and progress for interruption-safe execution.
- Surface blockers only when filesystem operations fail.

Execution entrypoint:
- Run: `./scripts/publication-agent.sh run`
- Status: `./scripts/publication-agent.sh status`
- Reset: `./scripts/publication-agent.sh reset`

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:970c3bf2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:
   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   bd dolt push
   git push
   git status
   ```
5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**
- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->
