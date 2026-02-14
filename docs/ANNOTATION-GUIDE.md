# Annotation Quality Guide

Clear annotations make Stricture findings readable and actionable. Use this checklist when tagging fields.

## Must-have fields
- **Owner & escalation**: team slug and on-call channel.
- **Contract ref**: link or git ref to the source API/schema/event.
- **Data class & units**: public/internal/regulated + units (ms, USD, bytes).
- **Type shape**: exact type, ranges, enum values, nullable?, collection sizes.
- **Source version & min supported**: e.g., `source_version=v2026.02`, `min_supported_source_version=v2026.01`.
- **Example payload**: realistic example, including edge values.

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
field_id: response.payments.status
owner: team.payments-platform
escalation: pagerduty:payments-oncall
contract_ref: git https://github.com/acme/payments-api//openapi.yaml@v1.4.2
source_version: v2026.04
min_supported_source_version: v2026.02
confidence: declared
transform_type: passthrough
merge_strategy: priority
break_policy: strict
annotation_schema_version: 1
note: type=uint8 (0-255), enum=[pending, success, failed, pending_review]; PSP may add values quarterly; downstream JS consumers must guard >253.
```

## When to block vs warn
- **Block**: type narrowing, enum removals, size increases crossing consumer limits, missing owner/contract for regulated data.
- **Warn**: additive enum, documentation-only changes, non-breaking optional fields with owners set.

Keep annotations specific; Stricture findings will inherit that clarity.
