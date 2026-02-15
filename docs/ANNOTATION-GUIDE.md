# Annotation Quality Guide

Clear annotations make Stricture findings readable and actionable. Use this checklist when tagging fields.

For CI/build automation and `.stricture-history` design, see
`docs/LINEAGE-AUTOMATION-SPEC.md`.

## Must-have fields (compact mode)
- **Field identity/path**: `field` (or explicit `field_id` if you need stable rename handling).
- **Producer**: `source_system`.
- **Producer version**: `source_version`.
- **Provenance**: `sources` with `contract_ref` (plus `provider_id` + `as_of` for external).

Everything else can be defaulted by Stricture and expanded in normalized artifacts.

Reference handle namespace in docs/tooling UX uses `strict:*` (for example
`strict:source`, `strict:systems[]`, `strict:flows[]`). Source-comment syntax
remains `stricture-source`.

## Hierarchical service IDs (for ecosystem + internals)
- Use one optional `:` in system IDs to model internals without new keys.
- Topology service: `location-tracking-service`
- Internal subsystem: `location-tracking-service:tracking-api`
- Use the same format in `source_system` and `upstream_system`.

## Flow tiers (service-level, not API-level)
- Define business flows in registry metadata (`'strict:flows'`) with numeric levels.
- Tag services with flow membership (`systems[].flows`).
- Keep service-level ops metadata in the same registry row (`owner_team`, optional `runbook_url`, optional `doc_root`, escalation).
- Do not assign criticality tiers directly on individual APIs by default.
- Let lineage path analysis determine which APIs/edges are affected at runtime.
- Keep flow tier decisions in policy packs (`lineage.findings.flow_criticality`).

## Good vs bad examples
- **Type/size**
  - Bad: `type: number`
  - Good: `type: uint8, range: 0-255, unit: count`
- **Enum**
  - Bad: `enum: [pending, success]`
  - Good: `enum: [pending, success, failed], note: PSP may add values quarterly; add guardrails before deploy`
- **Ownership**
  - Bad: owner missing
  - Good: owner: `team.payments-platform`, escalation: `pagerduty:payments-oncall`
- **Freshness**
  - Bad: as-of unspecified
  - Good: external provider as-of `2026-02-10`, max_staleness `24h`, SLA notes

## Writer prompts
- “What will break downstream if this changes?”
- “Which systems consume this field? Include them.”
- “What’s the largest/smallest value seen in prod?”
- “How fast can this provider add a new enum or widen a type?”

## Field annotation template
```
field: response.payments.status
source_system: PaymentsGateway
source_version: v2026.04
sources: api:payments-core.GetStatus#response.status@cross_repo?contract_ref=git+https://github.com/acme/payments-core//openapi.yaml@v1.4.2
```

Defaulted by tool:
- `annotation_schema_version=1`
- `min_supported_source_version=source_version`
- `transform_type=passthrough`
- `merge_strategy=single_source`
- `break_policy=strict`
- `confidence=declared`
- `data_classification=internal`
- `owner=team.<source_system_slug>`
- `escalation=slack:#<source_system_slug>-oncall`
- `contract_test_id=ci://contracts/<source_system_slug>/<field_id>`
- `introduced_at=1970-01-01`
- `flow="from @<source_system> mapped @self"`
- `note="defaulted_by=stricture"`

## When to block vs warn
- **Block**: type narrowing, enum removals, size increases crossing consumer limits, missing owner/contract for regulated data.
- **Warn**: additive enum, documentation-only changes, non-breaking optional fields with owners set.

Keep annotations specific; Stricture findings will inherit that clarity.
