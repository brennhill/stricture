# Lineage Plain-Language Message Library (Draft)

Last updated: 2026-02-15

## Purpose

Make lineage findings understandable on first read for humans and AI agents.

Every finding should answer, in plain language:

1. what changed
2. who changed it (cause service)
3. who is impacted (impacted service or "no downstream impact detected")
4. why this matters (impact)
5. what to do next (next step)

## Output Contract

For each drift change, Stricture should emit:

1. `message`: concise plain-language narrative of the change.
2. `source`: cause context (`service`, `api`).
3. `impact`: impacted context (`service`, `api`) when known.
4. `validation`: plain-language failure mode/risk statement.
5. `suggestion`: plain-language next step.
6. `flow_context` (optional): impacted flow IDs and effective flow tier level.
7. `policy_reason` (optional): explicit hard-block rationale when policy uses
   flow criticality controls.

When no downstream impact is detected, default behavior is:

1. do not emit warn/block finding
2. still track and publish the change event
3. surface explicit "no downstream impact detected" language in UX/CLI summaries

## Library (v0)

| Change type | Plain-language "what changed" | Plain-language impact (`validation`) | Plain-language next step (`suggestion`) |
|---|---|---|---|
| `merge_strategy_changed` | Producer changed how multiple sources are combined (for example custom -> single_source). | Output values can change even if schema/type did not change. | Add merge semantic regression tests with representative multi-source fixtures. |
| `enum_changed` | Producer introduced/removed an enum value used by this field. | Consumers with stale allowlists or switch logic can fail at runtime. | Add enum compatibility tests and safe fallback handling for unknown values. |
| `source_contract_ref_changed` | Upstream contract reference changed for this field. | Enum/type compatibility can drift across service boundaries. | Run producer+consumer contract tests before rollout. |
| `source_version_changed` | Producer version changed for this field mapping. | Producer semantics may change while shape appears stable. | Run compatibility tests and communicate rollout plan. |
| `field_removed` | Producer removed this field from its contract. | Consumers can crash or silently mis-handle payloads. | Restore field or ship compatibility adapter with coordinated rollout. |
| `source_removed` | One upstream source no longer contributes to this field. | Fallback/value composition may change unexpectedly. | Validate fallback behavior and downstream assumptions. |
| `source_added` | New upstream source now contributes to this field. | Precedence and final values may change. | Validate precedence/merge expectations end-to-end. |
| `external_as_of_rollback` | External snapshot date moved backwards. | Stale provider data may reintroduce old/incompatible values. | Refresh provider snapshot and verify freshness-sensitive invariants. |

Flow-tier narrative requirements:

1. If a finding touches a named critical flow, include flow name + level in the
   message (for example `checkout (level 1)`).
2. If policy escalates/blocks due to flow tier, state that explicitly.
3. Avoid ambiguous policy text; say whether block is severity-driven or
   flow-criticality-driven.

## Language Rules

1. Prefer concrete terms over internal jargon.
2. Avoid unresolved abbreviations (for example "PSP") unless expanded ("PSP (Stripe)").
3. Use left-to-right story order: `cause -> change -> impact`.
4. Keep each sentence independently understandable.
5. If Stricture cannot infer impact, say so explicitly.

## Adoption Notes

This library is the source of truth for:

1. `stricture lineage-diff` message text/guidance fields
2. demo findings narrative
3. future server/webhook human-readable summaries

## Demo Narrative Example (Payments)

Reference scenario:

1. `Promotions` adds a new enum value (`promotion_type=stacked_cashback`).
2. `PromotionsApplication` is updated in tandem and forwards the value.
3. `CommerceGateway` forwards promotion context into checkout totals.
4. `FintechGateway` payment authorization path still handles only legacy values.
5. Stricture message should explicitly state: cause service, changed enum value,
   impacted service(s), and concrete next step (allowlist + fallback + tests).
