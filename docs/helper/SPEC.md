# Stricture Helper Spec (Draft)

Last updated: 2026-02-15
Status: Draft v0

## Purpose

`stricture-helper` reduces manual annotation work by generating, updating, and
quality-checking `stricture-source` blocks from code + contract context.

Primary objective: make lineage adoption realistic in fast-moving repos with AI
agents and frequent API changes.

## What We Learned (Captured From Demo + UX Iteration)

1. Raw annotations became too long and noisy for day-to-day authoring.
2. Humans (and agents) will not reliably maintain `source_version` by hand.
3. Findings are harder to act on when metadata does not clearly map to:
   - cause
   - blast radius
   - owner/escalation
4. Teams want to reuse existing metadata from OpenAPI/OpenTelemetry/OpenLineage
   instead of retyping context.
5. Local-first workflows matter: developers may run many agents locally and
   still need CI-grade deterministic outputs.

## v0 Goals

1. Generate compact annotation blocks with deterministic defaults.
2. Auto-populate fields that should not be manually curated every PR.
3. Produce patch-ready edits (dry-run + apply).
4. Emit confidence + quality signals so uncertain output is reviewable.
5. Integrate with `.stricture-history/` so versioning/deltas are automatic.

High-priority automation targets:

1. `source_version`: derive from contract ref revision/commit where available.
2. `sources`: infer from AST + contract artifacts, then normalize refs.
3. external `provider_id` + `as_of`: derive from provider map + run context.
4. `contract_ref`: reuse existing OpenAPI/AsyncAPI/proto pointers.
5. service registry `id`: bootstrap from repo/service identity and optionally
   register to `strict:server_url` (including optional hierarchical IDs like
   `location-tracking-service:tracking-api`).
6. service flow memberships (`systems[].flows`): infer from known topology,
   path usage, and org registry defaults (with review hints when uncertain).
7. field/source keys marked defaulted in spec: emit compact source comments and
   rely on deterministic normalization for expanded artifacts.

## Non-Goals (v0)

1. Full autonomous merge of low-confidence inferred annotations.
2. IDE plugin ecosystem (separate track).
3. Cross-repo orchestration (belongs in `stricture-server` track).

## User Personas

1. **Local dev with AI agents:** needs quick, safe annotation scaffolding.
2. **CI maintainer:** needs deterministic, non-flaky outputs.
3. **Service owner:** needs readable ownership/escalation context.

## Authoring Model

`stricture-helper` works in two layers:

1. **Compact authoring view** (minimal keys in source comments).
2. **Expanded normalized artifact** (full explicit fields in export output).

This keeps source code terse while preserving deterministic policy checks.

Reference handles in helper output/report UX SHOULD use `strict:*` names (for
example `strict:source`, `strict:systems[]`) while preserving parser-compatible
source comment syntax (`stricture-source`).

## Automation Tiers

1. **Auto:** deterministic and safe.
2. **Auto+Policy:** deterministic with org config.
3. **Suggest:** inferred, confidence-scored.
4. **Manual:** explicit human intent required.

Reference matrix: `docs/LINEAGE-AUTOMATION-SPEC.md`.

## Proposed CLI Surface

```bash
stricture helper scan [paths...]
stricture helper suggest [paths...]
stricture helper apply [paths...]
stricture helper quality [paths...]
```

### Command Semantics

1. `scan`: discover candidate source fields/functions.
2. `suggest`: build annotation proposals + confidence + rationale.
3. `apply`: write patch-ready annotation updates.
4. `quality`: score existing annotations and flag ambiguity/gaps.

## Inputs

1. Source code AST (Go/TS/JS first; others as adapters mature).
2. Existing `stricture-source` blocks.
3. Optional contract sources (OpenAPI, AsyncAPI, protobuf, etc.).
4. Optional service registry (owner/escalation defaults).
5. `.stricture-history/` baseline/current/diff for version automation.

## Outputs

1. Proposed annotation edits (unified patch + JSON report).
2. Annotation quality report:
   - completeness
   - confidence
   - ambiguity reasons
   - required human review fields
3. Delta summary oriented to findings UX:
   - cause
   - blast radius
   - owner
   - recommended remediation

## Quality Heuristics (v0)

Score each annotation 0-100 with weighted checks:

1. field/source traceability confidence
2. ownership/escalation presence
3. contract reference quality
4. flow/source clarity
5. stale-version risk markers

Low score gates can run in warn mode first, then promote to block mode.

## CI Workflow (v0)

1. `stricture helper suggest` (non-destructive).
2. `stricture helper quality` (artifact + score thresholds).
3. `stricture history record/diff/summarize`.
4. optional `stricture helper apply` in bot PR mode.

## Configuration (Proposed)

`stricture-helper.yml`:

1. default field values and policy packs
2. ownership/escalation mappings
3. contract import sources
4. confidence thresholds + block/warn behavior
5. patch application rules (safe paths only, review-required zones)
6. optional flow-tier defaults and membership hints from policy/server catalogs

## Risks + Mitigations

1. **False confidence from inference**
   - Mitigation: explicit confidence labels + review-required tags.
2. **Noisy repeated edits**
   - Mitigation: stable formatting + deterministic key ordering.
3. **Policy drift across repos**
   - Mitigation: policy pack versioning + centralized templates.

## Milestones

1. **v0-a:** scan + suggest minimal compact blocks.
2. **v0-b:** apply mode + deterministic formatting.
3. **v0-c:** quality scoring + CI summary contract.
4. **v1:** stronger contract ingestion + richer source-path inference.

## Open Questions

1. Should helper patch apply be defaulted to PR-bot only in CI?
2. What is the minimum quality score allowed for block-mode rollout?
3. Which contract format import should be prioritized first (OpenAPI vs others)?
