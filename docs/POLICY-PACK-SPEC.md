# Stricture Policy Pack Spec (Draft)

Last updated: 2026-02-15
Status: Draft v0

## Purpose

Policy packs let organizations tighten or customize annotation requirements
without forcing all projects to use the same global baseline.

This is the control plane for:

1. promoting recommended keys to required
2. overriding defaults
3. tuning severity for policy findings
4. enforcing different strictness by environment/profile
5. defining when drift becomes a finding vs. a tracked/published change event
6. applying business-flow tier criticality to drift gating

## Scope

This spec covers lineage-facing policy behavior for:

1. field annotations (`strict-source`)
2. source references (inline `sources:` or sidecar `fields:`)
3. service registry keys (`systems[]`)
4. lineage override keys (`strict-lineage-override`)
5. business flow tier metadata (`strict_flows`, `systems[].flows`)

System IDs in policy scope can use hierarchical form `parent:child` to model
internal subsystems without introducing new annotation keys.

## Core Model

Policy packs are standalone YAML files.

Recommended location:

```text
.stricture/strict-policy.yaml
```

Recommended identity handle:

```text
policy_id: production_standard # org-provided string; can be any useful value
```

`stricture_policy` is the config key prefix for policy-related settings.

## Policy File Shape (v0 Draft)

```yaml
schema_version: 1
policy_id: production_standard # org-provided string; can be any useful value
extends: []

lineage:
  require:
    field_keys: []
    source_keys: []
    system_registry_keys: []
    override_keys: []
  defaults: {}
  severity_overrides: {}
  findings: {}
  profiles: {}
```

## Semantics

### `lineage.require`

Promotes keys to required in the selected scope.

Example:

- `system_registry_keys: [escalation, escalation[].role]`
- `system_registry_keys: [runbook_url, doc_root]`

### `lineage.defaults`

Overrides base default values from normalization.

Example:

- `break_policy: strict`
- `escalation: pagerduty:platform-oncall`

### `lineage.severity_overrides`

Maps policy finding classes to severity.

Example:

- `missing_required_key: high`
- `invalid_default_value: medium`

### `lineage.profiles`

Environment overlays (for example `dev`, `staging`, `prod`) that merge onto the
base policy.

### `lineage.findings`

Controls impact gating behavior for drift findings.

Default model (if omitted):

```yaml
lineage:
  findings:
    require_downstream_impact: true
    unknown_impact_severity: low
    self_only:
      emit_finding: false
      publish_change_event: true
```

Interpretation:

1. Downstream-impact drift can emit findings and participate in gate decisions.
2. Self-only drift (service changes that do not impact downstream consumers)
   does not emit warnings/errors by default.
3. Self-only drift is still tracked and published as a change event for audit,
   release notes, and external customer visibility.
4. Unknown impact defaults to `low` severity unless overridden.

### `lineage.findings.flow_criticality`

Controls optional flow-tier aware gating based on service membership.

Default model (if omitted):

```yaml
lineage:
  findings:
    flow_criticality:
      enabled: false
      level_direction: lower_is_more_critical
      fail_on_level: null
      severity_by_level: {}
      require_service_membership: false
      critical_flow_ids: []
      critical_flow_block_reason: ""
```

Interpretation:

1. flow criticality is evaluated only for downstream-impact findings.
2. impacted flow set is derived from affected services (`source`, `impact`,
   and affected path nodes) by joining service IDs to `systems[].flows`.
3. effective finding level is the highest criticality level touched:
   - `lower_is_more_critical`: minimum numeric level
   - `higher_is_more_critical`: maximum numeric level
4. when `fail_on_level` is set, findings touching that level (or more critical)
   are escalated for gating.
5. `severity_by_level` can override computed severity by effective level.
6. `critical_flow_ids` forces block behavior for listed flows regardless of base
   severity threshold.
7. per-API tier tags are intentionally out of baseline policy; flow membership
   is service-level by design.

## Merge And Evaluation Order

1. Parse source annotations.
2. Normalize with base defaults.
3. Apply policy default overrides.
4. Apply selected profile overlay (if any).
5. Compute impact scope per drift (`downstream`, `self_only`, `unknown`).
6. Compute impacted flows/effective level when flow criticality is enabled.
7. Evaluate required-key checks.
8. Emit policy findings with configured severities and `lineage.findings` rules.
9. Gate via existing warn/block mode using final severities and flow overrides.

## Distribution Model (Locked Requirement)

Policy must be consumable by:

1. CI/CD pipelines
2. local developer/agent tooling
3. offline local runs (from cache)

Canonical repository binding is a single policy URL reference:

```yaml
stricture_policy_url: https://policies.example.com/stricture/strict-policy.yaml
```

Recommended naming: keep the policy file named `strict-policy.yaml` whether it is stored locally or served remotely, so the URL and the repo-local filename match.

Supported URL targets:

1. `stricture-server` policy API URL
2. GitHub raw URL or GitHub repo-path URL
3. internal HTTPS policy host

Optional integrity pin:

```yaml
stricture_policy_sha256: <sha256-hex>
```

Optional server binding for registry/bootstrap workflows:

```yaml
stricture_server_url: https://stricture.example.com
```

When flow-criticality rules are enabled, clients should also resolve flow
catalog metadata (for example `strict_flows`) from the same governed source
as policy packs (server API or pinned GitHub source).

## Resolution Order

When `stricture_policy_url` is configured, clients should resolve in this order:

1. valid local cache entry (if fresh)
2. fetch from configured policy URL
3. fallback to stale cache when network is unavailable

If network is unavailable and cache exists, run in offline mode with cached
policy and emit an informational notice.

## Local Cache Contract

Recommended cache location:

```text
.stricture-cache/policies/<policy_id>.yml
.stricture-cache/policies/<policy_id>.meta.json
```

Recommended metadata:

1. `policy_id`
2. `version`
3. `source`
4. `fetched_at`
5. `etag` or content hash

## Server API Contract (Draft)

`stricture-server` should expose:

1. `GET /v1/policies/{policy_id}` (latest)
2. `GET /v1/policies/{policy_id}/versions/{version}`
3. `GET /v1/policies/{policy_id}/metadata`

Server responses should include cache hints (ETag and last-modified style
metadata) to minimize downloads.

## Org Compliance Check (Draft)

Companies can enforce consistent policy adoption by checking all repos for the
expected `stricture_policy_url`.

Recommended rule:

1. every governed repo must declare `stricture_policy_url`
2. value must match the org-approved URL (or approved allowlist)
3. CI should fail when repo binding deviates

See CLI command contract: `docs/POLICY-CLI-CONTRACT.md`.

## Finding Classes (Draft)

1. `missing_required_key`
2. `invalid_default_value`
3. `unknown_policy_key`
4. `disallowed_value`
5. `unknown_impact_scope`
6. `unknown_flow_id`
7. `missing_flow_membership`

## Reference Strict Policy Example

```yaml
schema_version: 1
policy_id: production_standard # org-provided string; can be any useful value

lineage:
  require:
    field_keys:
      - owner
      - escalation
    source_keys:
      - contract_ref
    system_registry_keys:
      - escalation
      - escalation[].role
      - escalation[].name
      - escalation[].channel
      - runbook_url
      - doc_root
    override_keys:
      - ticket
  defaults:
    break_policy: strict
    data_classification: internal
  findings:
    require_downstream_impact: true
    unknown_impact_severity: low
    self_only:
      emit_finding: false
      publish_change_event: true
    flow_criticality:
      enabled: true
      level_direction: lower_is_more_critical
      fail_on_level: 1
      severity_by_level:
        "1": high
        "2": medium
      require_service_membership: false
      critical_flow_ids:
        - checkout
      critical_flow_block_reason: "Risk of order loss"
  severity_overrides:
    missing_required_key: high
    unknown_impact_scope: low
    unknown_flow_id: medium
```

## Compatibility Notes

1. Base schemas remain adoption-friendly.
2. Policy packs add organization strictness on top.
3. If a key is required by both base schema and policy, base schema still
   applies.
4. If policy conflicts with base parser constraints, parser constraints win and
   policy must be corrected.

## Implementation Status

Spec is locked at draft level; runtime enforcement wiring is a follow-up
implementation track in helper/server/core roadmap phases.
