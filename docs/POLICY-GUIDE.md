# Stricture Policy Guide

**What it is:** a practical guide to policy packs (fields, defaults, and best use cases).  
**What it isn’t:** the machine schema (see `docs/schemas/lineage-policy-pack.schema.json`).  
**Who it’s for:** platform teams and CI owners defining enforcement rules.

This guide explains policy packs: required vs optional fields, what each block does, and when to use it.

## File format

Policy packs are YAML files validated by `lineage-policy-pack.schema.json`.

Naming convention:
- Local file: `.stricture/strict-policy.yaml`
- Remote URL: `https://.../strict-policy.yaml`

Minimal policy:

```yaml
schema_version: 1
policy_id: production_standard # org-provided string; can be any useful value
lineage:
  findings:
    require_downstream_impact: true
```

## Top-level fields

Required:
- `schema_version` (int, must be `1`)
- `policy_id` (string, org-defined identifier)
- `lineage` (object)

Optional:
- `extends` (array of strings): compose from other policy IDs or URLs.

## `lineage` object

### `lineage.require`
Declare keys that must exist in annotations/registries/overrides.

Optional keys (arrays of strings):
- `field_keys`
- `source_query_keys`
- `system_registry_keys`
- `override_keys`

Use this to enforce stronger annotation completeness in production.

### `lineage.defaults`
Define default values for fields when omitted. Keys are annotation keys and values are strings/numbers/booleans.

Use this to keep comments short while maintaining consistent metadata.

### `lineage.severity_overrides`
Map change types to severities (`high|medium|low|info`).

Use this when certain drift types should always be treated as more severe.

### `lineage.findings`
Controls what becomes a finding and how it is classified.

Optional keys:
- `require_downstream_impact` (bool, default `true`): emit findings only if downstream impact exists.
- `unknown_impact_severity` (`high|medium|low|info`, default `low`): severity for impact-unknown cases.
- `self_only` (object):
  - `emit_finding` (bool, default `false`)
  - `severity` (`high|medium|low|info`, optional)
  - `publish_change_event` (bool, default `true`)
- `flow_criticality` (object):
  - `enabled` (bool)
  - `level_direction` (`lower_is_more_critical|higher_is_more_critical`)
  - `fail_on_level` (int or null)
  - `severity_by_level` (map level -> severity)
  - `require_service_membership` (bool)
  - `critical_flow_ids` (string[])
  - `critical_flow_block_reason` (string)

Use this when business tiers (flows) should block deployments.

### `lineage.profiles`
Free-form profiles for bundling policy subsets. No enforced schema yet.

## Best-practice patterns

- **Production**: require key fields, enable downstream-impact gating, and enable flow criticality for tier-1 flows.
- **Staging**: use `warn` mode in CI but keep the same policy pack to avoid drift in rules.
- **External providers**: require `provider_id` + `as_of` for `@external` scope and publish change events even when not blocking.

## Related docs

- `docs/POLICY-PACK-SPEC.md` (full spec)
- `docs/POLICY-CLI-CONTRACT.md` (CI enforcement)
- `docs/LINEAGE-AUTOMATION-SPEC.md` (policy integration with automation)
