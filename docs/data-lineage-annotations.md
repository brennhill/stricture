<!--
SPDX-FileCopyrightText: 2026 Brenn Hill
SPDX-License-Identifier: CC-BY-4.0
-->

# Data Lineage Annotations

This document defines a strict, versioned annotation contract for tracing each API
output field back to all source systems and contracts.

For a quick-start walkthrough, see `docs/LINEAGE-QUICKSTART.md`.
For annotation quality guidance, see `docs/ANNOTATION-GUIDE.md`.
For build/CI automation, see `docs/LINEAGE-AUTOMATION-SPEC.md`.
For term definitions, see `docs/GLOSSARY.md`.

## Goals

- Detect breaking drift across services before deploy.
- Make ownership and emergency escalation explicit.
- Produce deterministic artifacts for tool-based diffing.

## Annotation Format

Stricture uses inline comments and optional YAML sidecar files for lineage
metadata. The simplest inline annotation is `strict-source: ServiceName`.

### Minimal Inline (most common)

```go
// strict-source: Identity
UserID int `json:"user_id"`
```

The field is inferred from the struct field on the next line. The source system
is `Identity`. Everything else is derived from `strict-lineage.yml` and
normalization defaults.

### Inline with Qualifiers

When defaults are insufficient, add comma-separated qualifiers:

```go
// strict-source: Identity, scope external, provider spotify
TrackID string `json:"track_id"`
```

### Multi-Line Block

For complex annotations or multiple sources:

```go
// strict-source:
//   field: response.user_profile
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
//   note: "identity payload enriched with profile DB fields"
```

### Sidecar File (`strict-lineage.yml`)

Per-package or per-project YAML that centralizes lineage metadata:

```yaml
# strict-lineage.yml
service: user-service
version: v2026.02

upstream:
  Identity:
    contract: git+https://github.com/acme/identity//openapi.yaml
    scope: cross_repo
  PaymentsCore:
    contract: git+https://github.com/acme/payments-core//openapi.yaml@v1.4.2
    scope: cross_repo
  Spotify:
    contract: https://developer.spotify.com/reference/get-track
    scope: external
    provider: spotify

fields:
  response.user_id:
    from: Identity
  response.user_name:
    from: Identity
  response.track_id:
    from: Spotify
  response.payment_status:
    from: PaymentsCore
    transform: normalize
    note: "Mapped from PSP status codes to internal enum"
```

When a sidecar file exists, inline annotations can reference it implicitly.
`strict-source: Identity` on a field resolves to the contract and scope defined
in the sidecar's `upstream.Identity` section.

## Comment Prefixes

Accepted prefixes for inline annotations:

- `// strict-source: ...`
- `# strict-source: ...`

Accepted prefixes for overrides:

- `// strict-lineage-override ...`
- `# strict-lineage-override ...`

## Required Field

Only one value is required in inline annotations:

| Field | Description | Example |
|-------|-------------|---------|
| source system token | Source system name (and optional version) after `strict-source:` | `strict-source: Identity`, `strict-source: Identity v2026.02` |

## Auto-Inferred Fields

These fields are derived automatically and should never be written manually:

| Field | Inferred From | Default |
|-------|---------------|---------|
| `field` | Struct field / variable on the line below the annotation | -- |
| `field_id` | Derived from `field` (dots/brackets to underscores) | -- |
| `source_version` | `from` qualifier or sidecar `version` | sidecar version |
| `scope` | Sidecar `upstream.*.scope` or contract URL analysis | `internal` |
| `contract` | Sidecar `upstream.*.contract` | -- |
| `provider` | Sidecar `upstream.*.provider` (external only) | -- |
| `owner` | Service registry lookup by source system | `team.<system_slug>` |
| `escalation` | Service registry lookup | `slack:#<system_slug>-oncall` |
| `annotation_schema_version` | Always `2` | `2` |
| `confidence` | `declared` for manual, `inferred` for helper-generated | `declared` |
| `data_classification` | Policy default | `internal` |
| `break_policy` | Policy default | `strict` |
| `transform` | -- | `passthrough` |
| `merge` | -- | `single_source` |
| `contract_test_id` | Org naming template | `ci://contracts/<system>/<field_id>` |
| `introduced_at` | Git blame date of the annotation | -- |
| `flow` | -- | `from @<system> mapped @self` |
| `note` | -- | `""` |

## Optional Qualifiers

Write only when overriding defaults:

| Qualifier | When to Use | Example |
|-----------|-------------|---------|
| `scope` | Override auto-detected scope | `scope external` |
| `provider` | External provider ID | `provider spotify` |
| `transform` | Non-passthrough transformation | `transform aggregate` |
| `merge` | Multiple sources with merge strategy | `merge priority` |
| `classification` | Override data classification | `classification sensitive` |
| `break_policy` | Override break policy | `break_policy additive_only` |
| `note` | Free-text context | `note "mapped via UserNormalizer"` |

## Source References

Each source in a multi-source annotation uses explicit named fields:

```yaml
sources:
  - kind: api
    target: identity.GetUser
    path: response.id
    scope: cross_repo
    contract: git+https://github.com/acme/identity//openapi.yaml@a1b2
```

Source fields:

| Field | Required | Description |
|-------|----------|-------------|
| `kind` | Yes | `api`, `db`, `event`, `file`, `cache`, `input` |
| `target` | Yes | The operation or table (e.g., `identity.GetUser`, `profiles.user`) |
| `path` | Yes | The field path in the source response (e.g., `response.id`) |
| `scope` | No | `internal` (default), `cross_repo`, `external` |
| `contract` | Yes | Reference to the contract artifact |
| `provider` | External only | Third-party provider ID |
| `as_of` | External only | Snapshot date `YYYY-MM-DD` |
| `upstream_system` | No | System ID if different from the annotation's `from` |

## Sidecar File Spec

### File Discovery

Stricture looks for lineage sidecar files in this order:

1. `strict-lineage.yml` in the same directory as the source file
2. `strict-lineage.yml` in parent directories (up to project root)
3. `strict-lineage.yml` at project root

Closer files take precedence. Fields defined in an inline annotation override
sidecar values.

### Schema

```yaml
# Required
service: string          # This service's identity

# Optional
version: string          # Current service version (used as default source_version)

# Upstream service definitions
upstream:
  <ServiceName>:
    contract: string     # Contract artifact reference (required)
    scope: string        # internal | cross_repo | external (default: cross_repo)
    provider: string     # External provider ID (external scope only)
    version: string      # Override source_version for this upstream

# Per-field lineage (alternative to inline annotations)
fields:
  <dotted.field.path>:
    from: string         # Source system name (must match an upstream key)
    transform: string    # Override transform type
    merge: string        # Override merge strategy
    classification: string  # Override data classification
    note: string         # Free-text context
    sources:             # Multiple sources (overrides simple from)
      - kind: string
        target: string
        path: string
        scope: string
        contract: string
```

### Sidecar vs. Inline: When to Use Which

| Scenario | Recommendation |
|----------|---------------|
| Simple 1:1 field mapping | Inline `strict-source: ServiceName` |
| Many fields from same upstream | Sidecar `fields:` section |
| Complex multi-source with merge | Multi-line inline block |
| Team wants centralized lineage review | Sidecar file |
| Team wants co-located lineage context | Inline annotations |
| Both exist for same field | Inline wins (override) |

## Canonical Field Ordering

When generating multi-line annotations, fields appear in this order:

```
1. field          (auto-inferred in minimal form)
2. source system token after `strict-source:` (required)
3. scope          (only if overriding default)
4. provider       (external only)
5. transform      (only if non-passthrough)
6. merge          (only if non-single_source)
7. classification (only if non-internal)
8. break_policy   (only if non-strict)
9. sources        (only for multi-source or when specifying details)
10. note          (free text, always last)
```

## System ID Hierarchy

Stricture supports service-internal topology using a single system ID with an
optional `:` suffix:

- Top-level service: `location-tracking-service`
- Internal subsystem: `location-tracking-service:tracking-api`

This convention applies to `from` values and `upstream_system` in sources.

## Business Flows and Tier Levels

Stricture supports named business flows with numeric criticality levels.
Flow criticality is modeled at the service level, not per API field.

Registry shape:

```yaml
strict_flows:
  - id: checkout
    name: Checkout
    level: 1
    owner: team.payments
    business_risk: order_loss
  - id: support
    name: Customer Support
    level: 3
    owner: team.support
    business_risk: customer_delay

systems:
  - id: commerce-gateway
    name: CommerceGateway Service
    owner_team: team.ecommerce
    runbook_url: https://runbooks.example.com/commerce-gateway
    doc_root: https://docs.example.com/commerce-gateway
    flows: [checkout]
  - id: support-console
    name: Support Console
    owner_team: team.support
    doc_root: https://docs.example.com/support-console
    flows: [support]
```

Rules:

1. `level` is numeric and organization-defined.
2. Each service can belong to zero or more flows (`systems[].flows`).
3. Flow definitions are canonical in `strict_flows`.
4. Findings derive impacted flows from affected services + lineage paths.
5. `runbook_url` and `doc_root` are optional service-level metadata.
6. `business_risk` is an organization-defined risk code token.
7. `owner_team` and flow `owner` are free-form strings.

## Namespace Convention

Stricture uses consistent prefixes:

| Context | Prefix | Example |
|---------|--------|---------|
| Source comments | `strict-` | `// strict-source:` |
| YAML config keys | `stricture_` | `stricture_policy_url` |
| Override comments | `strict-` | `// strict-lineage-override` |
| Sidecar filenames | `strict-` | `strict-lineage.yml` |

## Optional Fields

- `renamed_from`: previous `field_id` when identity is intentionally migrated.
- `sunset_at`: `YYYY-MM-DD` deprecation/removal target date.

## Examples

### Minimal Internal API Field

```go
// strict-source: Identity
UserID int `json:"user_id"`
```

### With Version Override

```go
// strict-source: Identity v2026.01
LegacyID int `json:"legacy_id"`
```

### External Provider

```go
// strict-source: Spotify, scope external, provider spotify
TrackID string `json:"track_id"`
```

### Internal Subsystem

```go
// strict-source: location-tracking-service:tracking-api
ETA float64 `json:"eta"`
```

### Multi-Source Merge

```go
// strict-source:
//   field: response.user_profile
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

### Sidecar-Only (No Inline Annotations)

```yaml
# strict-lineage.yml
service: checkout-service
version: v2026.02

upstream:
  Payments:
    contract: git+https://github.com/acme/payments//openapi.yaml
  Inventory:
    contract: git+https://github.com/acme/inventory//openapi.yaml

fields:
  response.order.total:
    from: Payments
  response.order.items:
    from: Inventory
    note: "Aggregated from inventory availability check"
  response.order.shipping_estimate:
    from: Inventory
    transform: normalize
```

## Commands

- Export normalized artifact:
  - `strict lineage-export --out tests/lineage/current.json .`
- Diff artifacts:
  - `strict lineage-diff --base tests/lineage/baseline.json --head tests/lineage/current.json --fail-on medium --mode block`
- Resolve emergency chain:
  - `strict lineage-escalate --service ServiceY --artifact tests/lineage/current.json --systems docs/config-examples/lineage-systems.yml`

`lineage-diff` mode:

- `--mode block` (default): return non-zero if non-overridden finding meets `--fail-on`.
- `--mode warn`: always return zero; prints warning when threshold is met.

By default, findings are impact-gated (downstream impact required). Self-only
drift is still recorded in diff output but does not warn/block unless policy
overrides that behavior.

## Impact-Gated Findings

Each drift item is classified as:

1. `downstream`: change can impact one or more downstream consumers.
2. `self_only`: change is isolated to the producing service.
3. `unknown`: Stricture cannot confidently determine impact scope.

Default behavior:

1. `downstream` -> emits finding with severity model and participates in gates.
2. `self_only` -> no warning/error finding by default; tracked as change event.
3. `unknown` -> emits low-severity finding unless policy overrides.

When flow criticality is enabled in policy, downstream findings also carry
derived flow context (flow IDs + effective tier) based on affected services.

## Plain-Language Finding Requirement

Findings should be understandable without lineage internals context. For every
emitted finding, Stricture provides:

1. A direct "what changed" statement.
2. A plain-language impact explanation.
3. A plain-language next step.
4. Explicit cause and impacted context when known.

## Temporary Overrides

Use time-bounded overrides when a known migration window would otherwise block CI:

```go
// strict-lineage-override field_id=response_user_id change_type=field_removed expires=2026-06-30 reason="temporary dual-write migration" ticket=INC-12345
```

Rules:

- `field_id`, `change_type`, `expires`, `reason` are required.
- `change_type=*` matches any drift change type for that field.
- Override is active through its `expires` date (UTC day semantics).
- Expired overrides are ignored.

## Drift Severity Model

The diff engine classifies changes as `high|medium|low|info`.

Examples:

- High: field removed, break policy changed, data classification relaxed,
  min supported version changed, source removed, external `as_of` rollback.
- Medium: field added, source version changed, source contract changed,
  transform or merge strategy changed, contract test ID changed.
- Low/info: owner/escalation/note updates, classification tightened,
  external `as_of` advanced.

## Validation

```bash
# Check all annotations parse and reference valid fields/systems
strict helper validate [paths...]

# Strict mode (fail on stale versions, missing contracts)
strict helper validate --strict [paths...]
```

Validation checks:

1. Annotation syntax is valid.
2. `from` references match a sidecar `upstream` entry or service registry.
3. `field` references correspond to actual struct/interface fields.
4. `contract` URLs are reachable (optional, with `--check-contracts`).
5. Sidecar `upstream` entries have valid `contract` values.
6. No duplicate field definitions (sidecar + inline for same field).
