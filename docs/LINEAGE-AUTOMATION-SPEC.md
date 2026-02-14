# Lineage Automation Spec (Draft)

Last updated: 2026-02-14

## Goal

Reduce annotation maintenance overhead by moving high-churn metadata to a
deterministic build/CI history pipeline while keeping canonical lineage outputs
explicit for diffing and policy gates.

## Core Idea

Use a generated repository state directory:

```text
.stricture-history/
```

as the source of truth for computed lineage versioning and delta summaries.

Build/CI computes artifacts, compares against the last promoted baseline, and
records both machine and human-readable change history.

## Proposed `.stricture-history/` Layout

```text
.stricture-history/
  baseline.json          # last promoted lineage artifact
  current.json           # artifact generated in current build
  diff.json              # deterministic diff result (base vs current)
  versions.json          # computed version state per system/field_id
  summary.md             # human + AI readable summary
  runs/
    2026-02-14T21-30-00Z/
      current.json
      diff.json
      summary.md
```

## CLI Surface (Proposed)

```bash
stricture history init
stricture history record [paths...]
stricture history diff
stricture history summarize
stricture history promote
```

Semantics:

1. `init` creates scaffolding and optional defaults file.
2. `record` runs lineage export and writes `current.json`.
3. `diff` compares `baseline.json` and `current.json`, writes `diff.json`.
4. `summarize` emits `summary.md` from diff + ownership/escalation context.
5. `promote` advances baseline after successful policy checks.

## Annotation Automation Matrix

Automation tiers:

1. **Auto**: deterministic and safe to compute.
2. **Auto+Policy**: deterministic with org policy/templates.
3. **Suggest**: inferred with confidence; requires review.
4. **Manual**: should remain explicit owner intent.

| Key | Tier | Source | Notes |
|---|---|---|---|
| `annotation_schema_version` | Auto | constant | always `1` |
| `field_id` | Auto | derive from `field` | stable slug transform |
| `field` | Auto | derive from `field_id` | fallback only |
| `source_system` | Suggest/Auto+Policy | repo/service mapping | can be defaulted from module map |
| `source_version` | Auto | `.stricture-history/versions.json` | computed build version |
| `min_supported_source_version` | Auto | source version policy | default same as `source_version` |
| `transform_type` | Auto+Policy / Suggest | defaults + static analysis hints | default `passthrough` |
| `merge_strategy` | Auto | source count | single source => `single_source`, multi => `priority` |
| `break_policy` | Auto+Policy / Manual | policy pack | default `strict`; override for compatibility windows |
| `confidence` | Auto | derivation mode | `declared`/`inferred` based on generation source |
| `data_classification` | Auto+Policy / Manual | path/classification registry | default `internal`; regulated paths policy-driven |
| `owner` | Auto+Policy | system registry | from service ownership map |
| `escalation` | Auto+Policy | system registry | from on-call routing map |
| `contract_test_id` | Auto+Policy | naming template | `ci://contracts/<service>/<field_id>` |
| `introduced_at` | Auto | first seen in history | initial fallback date when unknown |
| `sources` | Suggest / Auto+Policy | static analysis + OpenAPI/AsyncAPI/proto | confidence-scored candidates |
| `flow` | Suggest / Auto+Policy | graph path synthesis | generated from source edges |
| `note` | Auto+Policy / Suggest | summary templates | machine-generated; editable |
| `renamed_from` | Suggest | history match + similarity | require confirmation |
| `sunset_at` | Manual / Auto+Policy | deprecation policy | often ticket-driven |

## What Should Stay Human-Owned

Even with automation, these should usually require explicit review in PR:

1. `break_policy` for externally exposed contracts.
2. `data_classification` for sensitive/regulated domains.
3. `renamed_from` and deprecation windows (`sunset_at`).
4. Any auto-generated `sources` with low confidence.

## CI Workflow (Reference)

1. Generate `current.json` (`lineage-export`).
2. Diff against `baseline.json` (`lineage-diff`).
3. Compute version bumps from diff classes:
   - breaking/high -> major
   - additive/medium -> minor
   - metadata/low -> patch
4. Write `versions.json` and `summary.md`.
5. Enforce policy gates.
6. Promote baseline on successful protected branch builds.

## UX Output Contract for Summaries

Each finding summary should include:

1. Cause
2. Blast radius
3. Owner
4. Escalation path
5. Remediation steps
6. Version bump rationale

This ensures both humans and AI agents can act without reconstructing context
from raw diffs.
