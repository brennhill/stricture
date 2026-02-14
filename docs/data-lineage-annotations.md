<!--
SPDX-FileCopyrightText: 2026 Brenn Hill
SPDX-License-Identifier: CC-BY-4.0
-->

# Data Lineage Annotations for API Outputs

This document defines a strict, versioned annotation contract for tracing each API
output field back to all source systems and contracts.

For standards compatibility and reuse policy (OpenAPI/AsyncAPI/OpenLineage/OTel),
see `OVERLAY.md`.

For build/CI automation and version-history workflow design, see
`docs/LINEAGE-AUTOMATION-SPEC.md`.

## Goals

- Detect breaking drift across services before deploy.
- Make ownership and emergency escalation explicit.
- Produce deterministic artifacts for tool-based diffing.

## Annotation Marker

Use one annotation per output field:

```text
stricture-source <key=value pairs>
```

Accepted comment prefixes:

- `// stricture-source ...`
- `// stricture:source ...`
- `# stricture-source ...`
- `# stricture:source ...`

## Authoring-Minimal Keys

To keep annotations compact, only these keys are required in source comments:

- `field` or `field_id`
- `source_system`
- `source_version`
- `sources`

All other canonical keys are defaulted during parsing/normalization.

## Normalization Defaults

If omitted, Stricture fills:

- `annotation_schema_version`: `1`
- `field_id`: derived from `field` (`.` / `-` / `[` / `]` -> `_`)
- `field`: derived from `field_id` (`_` -> `.`) when needed
- `min_supported_source_version`: equals `source_version`
- `transform_type`: `passthrough`
- `merge_strategy`: `single_source` (one source) or `priority` (multi-source)
- `break_policy`: `strict`
- `confidence`: `declared`
- `data_classification`: `internal`
- `owner`: `team.<source_system_slug>`
- `escalation`: `slack:#<source_system_slug>-oncall`
- `contract_test_id`: `ci://contracts/<source_system_slug>/<field_id>`
- `introduced_at`: `1970-01-01`
- `flow`: `from @<source_system> mapped @self`
- `note`: `defaulted_by=stricture`

Normalized artifacts still emit explicit canonical fields so diffs remain
deterministic.

## Accepted Synonyms (No-Rename Adoption)

Stricture keeps canonical keys above, but accepts these synonyms to ease
interop with existing metadata conventions:

- `field_path` -> `field`
- `service_name` -> `source_system`
- `service_version` / `spec_version` -> `source_version`
- `min_source_version` / `min_supported_version` -> `min_supported_source_version`
- `owner_team` -> `owner`
- `contract_test` / `test_id` -> `contract_test_id`

If both canonical and synonym are present with different values, parsing fails.

## Optional Keys

- `renamed_from`: previous `field_id` when identity is intentionally migrated.
- `sunset_at`: `YYYY-MM-DD` deprecation/removal target date.

## Automation Guidance

In compact authoring mode, only field identity/path, source system/version, and
sources should be hand-authored by default.

Commonly auto-filled in normalized artifacts:

- owner/escalation (from system registry)
- contract_test_id (from org naming template)
- merge/transform/confidence defaults
- flow/note defaults
- introduced_at fallback

Keep `data_classification`, `break_policy`, and lifecycle migration fields under
explicit human review for high-risk domains.

## Source Ref Grammar

Each entry in `sources`:

```text
kind:target#path[@scope[!as_of]][?contract_ref=<ref>[&provider_id=<id>][&upstream_system=<id>]]
```

`kind` values:

- `api`, `input`, `db`, `event`, `file`, `cache`

`scope` values:

- `internal`: same repo/system boundary (default)
- `cross_repo`: internal but different repo/system
- `external`: third-party/public API/provider

Rules:

- `contract_ref` is required for every source.
- `provider_id` is required for `external` sources.
- `as_of` (`YYYY-MM-DD`) is required for `external` sources.
- `provider_id` must not be set on non-external sources.

Accepted source query synonyms:

- `schema_ref` / `spec_ref` / `contract_uri` / `schema_url` -> `contract_ref`
- `provider` / `external_provider` -> `provider_id`
- `upstream_service` / `upstream_source_system` -> `upstream_system`
- `asof` / `snapshot_as_of` -> `as_of`

## Examples

Compact single-source internal (recommended):

```go
// stricture-source field=response.user_id source_system=Identity source_version=v2026.02 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2
```

Expanded (all canonical fields explicit):

```go
// stricture-source annotation_schema_version=1 field_id=response_user_id field=response.user_id source_system=Identity source_version=v2026.02 min_supported_source_version=v2026.01 transform_type=normalize merge_strategy=single_source break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-01-10 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2 flow="from @Identity normalized @self" note="normalized by UserNormalizer.Apply; spec=https://specs.example.com/user-id"
```

Multi-source with merge:

```ts
// stricture-source annotation_schema_version=1 field_id=response_user_profile field=response.user_profile source_system=Profile source_version=v3 min_supported_source_version=v2 transform_type=aggregate merge_strategy=priority break_policy=strict confidence=declared data_classification=sensitive owner=team.profile escalation=pagerduty:profile contract_test_id=ci://contracts/profile-user introduced_at=2026-01-15 sources=api:identity.GetUser#response.user@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@f00d,db:profiles.user#payload@internal?contract_ref=internal://db/profiles.user flow="from @Identity enriched @self" note="identity payload is enriched with profile DB fields in UserProfileAggregator"
```

External provider:

```go
// stricture-source annotation_schema_version=1 field_id=response_track field=response.track source_system=Media source_version=v1 min_supported_source_version=v1 transform_type=passthrough merge_strategy=single_source break_policy=strict confidence=declared data_classification=public owner=team.media escalation=pagerduty:media contract_test_id=ci://contracts/media-track introduced_at=2026-02-13 sources=api:spotify.GetTrack#response.track@external!2026-02-13?provider_id=spotify&contract_ref=https://developer.spotify.com/reference/get-track flow="from @Spotify enriched @self" note="mapped in TrackMapper"
```

## Commands

- Export normalized artifact:
  - `stricture lineage-export --out tests/lineage/current.json .`
- Export with profile aliases/proxies:
  - `stricture lineage-export --profile otel --out tests/lineage/current-otel.json .`
- Diff artifacts:
  - `stricture lineage-diff --base tests/lineage/baseline.json --head tests/lineage/current.json --fail-on medium --mode block`
- CI helper (uses baseline + head artifact generation):
  - `LINEAGE_MODE=warn ./scripts/check-lineage-drift.sh`
- Resolve emergency chain for bad data at a service:
  - `stricture lineage-escalate --service ServiceY --artifact tests/lineage/current.json --systems docs/config-examples/lineage-systems.yml`

`lineage-diff` mode:

- `--mode block` (default): return non-zero if non-overridden drift meets `--fail-on`.
- `--mode warn`: always return zero; prints warning when threshold is met.

## Temporary Overrides

Use time-bounded overrides when a known migration window would otherwise block CI:

```text
stricture-lineage-override field_id=<field_id> change_type=<change_type|*> expires=YYYY-MM-DD reason="<why>" [ticket=<id>]
```

Accepted comment prefixes:

- `// stricture-lineage-override ...`
- `// stricture:lineage-override ...`
- `# stricture-lineage-override ...`
- `# stricture:lineage-override ...`

Rules:

- `field_id`, `change_type`, `expires`, `reason` are required.
- `change_type=*` matches any drift change type for that field.
- Override is active through its `expires` date (UTC day semantics).
- Expired overrides are ignored.

Example:

```go
// stricture-lineage-override field_id=response_user_id change_type=field_removed expires=2026-06-30 reason="temporary dual-write migration" ticket=INC-12345
```

## Drift Severity Model

The diff engine classifies changes as `high|medium|low|info`.

Examples:

- High: field removed, break policy changed, data classification relaxed,
  min supported version changed, source removed, external `as_of` rollback.
- Medium: field added, source version changed, source contract ref changed,
  transform or merge strategy changed, contract test ID changed.
- Low/info: owner/escalation/note updates, classification tightened,
  external `as_of` advanced.
