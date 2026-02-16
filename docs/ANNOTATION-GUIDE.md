# Annotation Quality Guide

**What it is:** guidance for writing compact, high-signal annotations that make findings readable.
**What it isn't:** a full schema reference (see `docs/data-lineage-annotations.md`).
**Who it's for:** engineers and reviewers writing `strict-source` comments or sidecar metadata.

Clear annotations make Stricture findings readable and actionable. Use this checklist when tagging fields.

## Required field

Only one value is required: the source system name after `strict-source:`.

```go
// strict-source: Identity
UserID int `json:"user_id"`
```

Everything else is defaulted by Stricture and expanded in normalized artifacts.
See `docs/data-lineage-annotations.md` for the full auto-inferred fields table.

## Sidecar setup (one-time)

Create `strict-lineage.yml` at your project root to declare upstream services:

```yaml
service: my-service
version: v2026.02

upstream:
  Identity:
    contract: git+https://github.com/acme/identity//openapi.yaml
    scope: cross_repo
```

Once this exists, `Identity` resolves contract and scope automatically.

## When to add qualifiers

Only add qualifiers when overriding defaults:

```go
// strict-source: Spotify, scope external, provider spotify
TrackID string `json:"track_id"`
```

Available qualifiers: `scope`, `provider`, `transform`, `merge`, `classification`, `break_policy`, `note`.

## Multi-line block (complex cases)

Use multi-line when a field has multiple sources or needs detailed metadata:

```go
// strict-source:
//   from: Profile v3
//   transform: aggregate
//   merge: priority
//   sources:
//     - kind: api
//       target: identity.GetUser
//       path: response.user
//       scope: cross_repo
//       contract: git+https://github.com/acme/identity//openapi.yaml@f00d
//     - kind: db
//       target: profiles.user
//       path: payload
//       scope: internal
//       contract: internal://db/profiles.user
```

## Hierarchical service IDs

Use one optional `:` to model internal subsystems without extra keys:

- Topology service: `location-tracking-service`
- Internal subsystem: `location-tracking-service:tracking-api`

## Good vs bad examples

- **Annotation format**
  - Bad: `// strict-source field=response.user_id source_system=Identity source_version=v2026.02 sources=api:identity.GetUser#response.id@cross_repo`
  - Good: `// strict-source: Identity`
- **Type/size**
  - Bad: `type: number`
  - Good: `type: uint8, range: 0-255, unit: count`
- **Enum**
  - Bad: `enum: [pending, success]`
  - Good: `enum: [pending, success, failed], note: PSP may add values quarterly`
- **Ownership**
  - Bad: owner missing
  - Good: owner in sidecar or registry: `team.payments-platform`, escalation: `pagerduty:payments-oncall`
- **Freshness**
  - Bad: as-of unspecified for external
  - Good: external provider as-of `2026-02-10`, max_staleness `24h`

## Writer prompts

- "What will break downstream if this changes?"
- "Which systems consume this field? Include them."
- "What's the largest/smallest value seen in prod?"
- "How fast can this provider add a new enum or widen a type?"

## Field annotation template

Minimal (recommended):

```go
// strict-source: PaymentsGateway
Status string `json:"status"`
```

With qualifiers (only when needed):

```go
// strict-source: PaymentsGateway, transform normalize
Status string `json:"status"`
```

Everything else is defaulted:

- `field` = inferred from struct field below
- `source_version` = from sidecar `version` or `upstream.*.version`
- `scope` = from sidecar `upstream.*.scope`
- `contract` = from sidecar `upstream.*.contract`
- `owner` = from service registry
- `escalation` = from service registry
- `transform` = `passthrough`
- `merge` = `single_source`
- `break_policy` = `strict`
- `data_classification` = `internal`

## When to block vs warn

- **Block**: type narrowing, enum removals, size increases crossing consumer limits, missing owner/contract for regulated data.
- **Warn**: additive enum, documentation-only changes, non-breaking optional fields with owners set.

Keep annotations specific; Stricture findings will inherit that clarity.

## Related docs

- `docs/LINEAGE-QUICKSTART.md` (getting started)
- `docs/data-lineage-annotations.md` (full spec)
- `docs/GLOSSARY.md` (term definitions)
