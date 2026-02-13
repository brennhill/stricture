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
