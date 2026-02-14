# Stricture Policy Pack Spec (Draft)

Last updated: 2026-02-14
Status: Draft v0

## Purpose

Policy packs let organizations tighten or customize annotation requirements
without forcing all projects to use the same global baseline.

This is the control plane for:

1. promoting recommended keys to required
2. overriding defaults
3. tuning severity for policy findings
4. enforcing different strictness by environment/profile

## Scope

This spec covers lineage-facing policy behavior for:

1. field annotations (`stricture-source`)
2. source edge query keys (`sources=...?...`)
3. service registry keys (`systems[]`)
4. lineage override keys (`stricture-lineage-override`)

## Core Model

Policy packs are standalone YAML files.

Recommended location:

```text
.stricture-policy.yml
```

Recommended identity handle:

```text
policy_id: strict:policy
```

`strict:policy` is a reference handle (not a parser token in source code).

## Policy File Shape (v0 Draft)

```yaml
schema_version: 1
policy_id: strict:policy
extends: []

lineage:
  require:
    field_keys: []
    source_query_keys: []
    system_registry_keys: []
    override_keys: []
  defaults: {}
  severity_overrides: {}
  profiles: {}
```

## Semantics

### `lineage.require`

Promotes keys to required in the selected scope.

Example:

- `system_registry_keys: [escalation, escalation[].role]`

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

## Merge And Evaluation Order

1. Parse source annotations.
2. Normalize with base defaults.
3. Apply policy default overrides.
4. Apply selected profile overlay (if any).
5. Evaluate required-key checks.
6. Emit policy findings with configured severities.
7. Gate via existing warn/block mode using final severities.

## Distribution Model (Locked Requirement)

Policy must be consumable by:

1. CI/CD pipelines
2. local developer/agent tooling
3. offline local runs (from cache)

Canonical repository binding is a single policy URL reference:

```yaml
'strict:policy_url': https://policies.example.com/stricture/prod.yml
```

Supported URL targets:

1. `stricture-server` policy API URL
2. GitHub raw URL or GitHub repo-path URL
3. internal HTTPS policy host

Optional integrity pin:

```yaml
'strict:policy_sha256': <sha256-hex>
```

Optional server binding for registry/bootstrap workflows:

```yaml
'strict:server_url': https://stricture.example.com
```

## Resolution Order

When `'strict:policy_url'` is configured, clients should resolve in this order:

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
expected `'strict:policy_url'`.

Recommended rule:

1. every governed repo must declare `'strict:policy_url'`
2. value must match the org-approved URL (or approved allowlist)
3. CI should fail when repo binding deviates

See CLI command contract: `docs/POLICY-CLI-CONTRACT.md`.

## Finding Classes (Draft)

1. `missing_required_key`
2. `invalid_default_value`
3. `unknown_policy_key`
4. `disallowed_value`

## Reference Strict Policy Example

```yaml
schema_version: 1
policy_id: strict:policy

lineage:
  require:
    field_keys:
      - owner
      - escalation
    source_query_keys:
      - contract_ref
    system_registry_keys:
      - escalation
      - escalation[].role
      - escalation[].name
      - escalation[].channel
    override_keys:
      - ticket
  defaults:
    break_policy: strict
    data_classification: internal
  severity_overrides:
    missing_required_key: high
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
