<!--
SPDX-FileCopyrightText: 2026 Brenn Hill
SPDX-License-Identifier: CC-BY-4.0
-->

# Stricture Standards Overlay (Reuse-First)

Last updated: 2026-02-14

This document defines how Stricture should reuse major ecosystem standards
instead of inventing parallel metadata where overlap already exists.

Primary goal: make Stricture easy to adopt in existing stacks by being a
superset profile, not a replacement platform.

## Reuse Policy

1. If there is total/obvious overlap, use the major standard field as-is.
2. If overlap is partial, map both directions and preserve all information.
3. Only create Stricture extensions for semantics not present in standards.
4. Keep round-trip conversion lossless between Stricture artifact and profiles.

## Compatibility Profiles

Stricture should support these profiles:

1. `openapi` profile (HTTP contract surface)
2. `asyncapi` profile (event contract surface)
3. `openlineage` profile (lineage graph/event interchange)
4. `otel` profile (runtime correlation/observability)
5. `sarif` profile (CI findings interchange)
6. `backstage` profile (owner/system registry bootstrap)

## What We Reuse Directly

## OpenAPI + JSON Schema

Reuse directly:

1. Operation and path identity from OpenAPI.
2. Request/response schema constraints from JSON Schema.
3. Property-level structure for field path alignment.

Stricture adds only:

1. Provenance (`sources`, `flow`, `note`).
2. Drift policy (`break_policy`, `min_supported_source_version`).
3. Escalation metadata (`owner`, `escalation`).

## AsyncAPI + JSON Schema

Reuse directly:

1. Channel/message identity.
2. Payload schema constraints.
3. Message trait metadata where available.

Stricture adds only:

1. Field-level provenance for emitted/consumed payloads.
2. Cross-system source references and external `as_of`.

## OpenLineage

Reuse directly:

1. Standard run/job/dataset identity model.
2. Standard lineage transport/event envelope.
3. Standard column-level lineage where representable.

Stricture adds only via custom facet payload:

1. `break_policy`
2. `min_supported_source_version`
3. `flow`
4. `note`
5. `escalation`
6. override semantics and CI gate context

Guideline:

1. Prefer standard OpenLineage facets for common lineage data.
2. Put Stricture-only fields under a namespaced custom facet with schema URL.

## OpenTelemetry

Reuse directly:

1. Trace/span correlation IDs.
2. Resource-level service identity (`service.name` and related resource attrs).
3. Existing HTTP/RPC semantic-convention attributes for request context.

Stricture adds only:

1. compact verdict markers (for example rule/drift level/result)
2. artifact digest/correlation key to join runtime signals with lineage artifacts

Guideline:

1. Do not emit full lineage graphs on spans.
2. Keep Stricture attributes namespaced (avoid reserved `otel.*`).

## SARIF

Reuse directly:

1. Findings interchange format for code scanning.
2. Rule/result metadata and location model.

Stricture adds only:

1. Stricture-specific rule taxonomy and drift classification payload details.

## Backstage Catalog (optional but high-value)

Reuse directly:

1. `spec.owner` and system/component identity as owner bootstrap.

Stricture adds only:

1. escalation chain traversal logic and fallback behavior.

## Crosswalk: Stricture Field -> Standard Reuse

| Stricture concept | Reuse source | Handling |
| --- | --- | --- |
| `field` path | OpenAPI/AsyncAPI+JSON Schema property path | Reuse path identity; normalize during export/import |
| `source_system` | OTel `service.name`, OpenLineage system identity | Reuse canonical service identity |
| `source_version` | Spec/document version tags + lineage facet/tag metadata | Map into profile metadata; retain in Stricture artifact |
| `min_supported_source_version` | No universal standard field | Stricture extension |
| `transform` | Column lineage semantics where available | Reuse when available, else extension |
| `merge` | No universal standard field | Stricture extension |
| `break_policy` | No universal standard field | Stricture extension |
| `confidence` | No universal standard field | Stricture extension |
| `data_classification` | Existing policy tags/labels in platform metadata | Reuse tags where available; keep canonical in Stricture |
| `owner` | Backstage owner metadata | Reuse owner identity |
| `escalation` | Rarely standardized | Stricture extension (optionally linked to registry) |
| `contract_test_id` | CI/test system reference | Keep as Stricture field; optionally emit as tags |
| `sources[*].as_of` | No universal standard field | Stricture extension |

## Overlay Encoding Strategy

## 1. OpenAPI Overlay

Attach Stricture metadata to existing OpenAPI properties using extension keys,
without rewriting the base API spec.

Recommended extension key: `x-strict-source`.

Example (conceptual):

```yaml
overlay: 1.0.0
actions:
  - target: "$.paths['/users/{id}'].get.responses['200'].content['application/json'].schema.properties.user_id"
    update:
      x-strict-source:
        field_id: response_user_id
        source_system: Identity
        source_version: v2026.02
        sources:
          - kind: api
            target: identity.GetUser
            path: response.id
            scope: cross_repo
            contract_ref: "git+https://github.com/acme/identity//openapi.yaml@a1b2"
        flow: "from @Identity normalized @self"
```

## 2. AsyncAPI Extension

Attach Stricture metadata to payload properties using `x-strict-source`.

## 3. OpenLineage Facet Bridge

During export, generate:

1. standard lineage edges/facets for common lineage semantics
2. custom Stricture facet containing only non-standard keys

## 4. OpenTelemetry Bridge

Emit compact attributes on relevant spans/logs/metrics to correlate:

1. run/trace identity
2. Stricture result status
3. artifact digest

## Implemented Alias/Proxy Baseline

Current implementation supports:

1. Annotation ingest synonyms (for example `field_path`, `service_name`,
   `owner_team`, `contract_test`).
2. Source query synonyms (for example `schema_ref`, `provider`,
   `upstream_service`, `asof`).
3. Conflict fail-fast behavior when canonical and synonym values disagree.
4. `lineage-export --profile <stricture|openlineage|otel|openapi|asyncapi>`
   which emits canonical fields plus profile alias keys.

## Adoption Phases

1. Phase A: Import only
   - Read OpenAPI/AsyncAPI/OpenLineage/OTel context and enrich Stricture output.
2. Phase B: Bidirectional mapping
   - Export Stricture artifact into profile-compatible payloads.
3. Phase C: Enforced compatibility
   - CI gates require profile mapping completeness and round-trip stability.

## Non-Goals

1. Replacing existing observability/lineage platforms.
2. Defining new generic standards where existing ones are sufficient.
3. Encoding large lineage graphs directly in telemetry spans.

## Immediate Implementation Checklist

1. Add `profiles` section to annotation spec docs (openapi, asyncapi, openlineage, otel).
2. Add profile validation tests for mapping round-trip:
   - Stricture -> Profile -> Stricture
3. Add profile fixtures in `tests/fixtures/lineage-usecases/`.
4. Add CLI flags:
   - `lineage-export --profile openlineage`
   - `lineage-export --profile otel`
5. Add compatibility report mode:
   - missing standard mappings
   - extension-only fields by profile
