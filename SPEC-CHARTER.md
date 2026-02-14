<!--
SPDX-FileCopyrightText: 2026 Brenn Hill
SPDX-License-Identifier: CC-BY-4.0
-->

# Stricture Open Standard (SOS) Charter

Status: Draft
Version: 0.1
Last Updated: 2026-02-14

## Purpose

Stricture should be presented as an open standard with a reference implementation,
not only as a single tool. The goal is to standardize field-level lineage,
architecture invariants, and deterministic drift policy outcomes across multi-repo
and multi-system API ecosystems.

This charter defines how the Stricture Open Standard (SOS) is governed, versioned, and
validated for ecosystem adoption.

## Positioning

The Stricture Open Standard (SOS) has two layers:

1. Open Standard: canonical model and evaluation semantics.
2. Reference Engine: `stricture` CLI as one conformant implementation.

The standard is the contract. The CLI is an implementation.

## Scope

In scope:

1. Field-level lineage metadata model for API outputs and event payloads.
2. Source-system version and as-of tracking semantics.
3. Architecture invariant evaluation model.
4. Deterministic drift classification and policy gating semantics.
5. Override lifecycle semantics (reason, owner, expiry).
6. Ownership and escalation metadata model.
7. Overlay/profile mapping model for OpenAPI, AsyncAPI, OpenTelemetry,
   OpenLineage, and other compatible standards.
8. Conformance test suite requirements.

Out of scope:

1. Runtime observability backend design.
2. API documentation tooling replacement.
3. Dataset warehouse governance replacement.
4. Vendor-specific deployment mechanics.

## Design Principles

1. Semantic first: metadata must represent real contract and provenance intent.
2. Deterministic evaluation: identical input yields identical output.
3. Composable with existing standards: Stricture overlays standards, does not
   force replacement.
4. Incremental adoption: warn mode and profile overlays support staged rollout.
5. Operational accountability: failures include ownership and escalation routing.

## Deliverables

The initial open-standard program includes:

1. `spec/0.1-draft.md`: normative draft.
2. Machine-readable schemas for artifacts and registry objects.
3. Conformance fixture catalog with expected findings and policy decisions.
4. Compatibility matrix for profile overlays.
5. Reference implementation notes for independent implementers.

## Governance Model

### Roles

1. Spec Maintainers: approve normative changes and publish tagged spec versions.
2. Editors: prepare draft text and reconcile RFC feedback.
3. Implementers: build engines/adapters and report interoperability issues.
4. Adopters: provide production feedback and edge cases.

### Decision Process

1. Any normative change starts as an RFC PR.
2. RFC must include motivation, compatibility impact, and conformance impact.
3. Minimum review window: 7 calendar days.
4. Approval threshold: two maintainer approvals.
5. Breaking changes require explicit major-version proposal.

### RFC Categories

1. Semantic model changes.
2. Evaluation semantics changes.
3. Overlay profile changes.
4. Conformance test changes.
5. Governance process changes.

## Versioning and Compatibility

SOS versions follow semantic versioning.

1. Major: breaking semantic/evaluation changes.
2. Minor: backward-compatible new fields, drift classes, or profiles.
3. Patch: clarifications and editorial corrections.

Compatibility rules:

1. Implementations claiming `vX.Y` conformance must pass the `vX.Y` conformance
   suite.
2. Unknown additive fields must be ignored unless marked required by the same
   spec version.
3. Deprecated fields must have removal targets in the next major version.

## Conformance Program

Conformance is required for ecosystem trust.

1. Publish canonical fixtures and expected outputs.
2. Require deterministic ordering and policy outcomes.
3. Require profile export parity checks.
4. Provide status levels:
   - Core Conformant
   - Core + Profiles Conformant
   - Experimental Extensions

## Security and Compliance Expectations

1. Spec text must support data classification metadata.
2. Sensitive fields must support explicit policy constraints.
3. Override semantics must be auditable and time-bounded.
4. Conformance suites must include negative/security drift cases.

## Intellectual Property and Licensing

1. Spec text and schemas are licensed under CC BY 4.0.
2. No patent encumbrance should be introduced via normative language.
3. If any patented claim is known, it must be disclosed in RFC discussion.

## Transparency

1. All normative discussions happen in public issues/PRs.
2. Meeting notes and decisions are published.
3. Version release notes include migration guidance and rationale.

## Initial Roadmap

Phase 1:

1. Publish `v0.1-draft` and conformance fixture seed set.
2. Validate against current reference engine outputs.

Phase 2:

1. Add independent implementation trial feedback.
2. Stabilize profile mapping semantics.

Phase 3:

1. Publish `v1.0` once interoperability and migration criteria are met.

## Success Criteria

1. At least one implementation beyond the reference engine can pass core
   conformance.
2. Adopters can map existing OpenAPI/OTel/OpenLineage metadata without duplicate
   annotations.
3. Deterministic drift outcomes hold across platforms and runners.

## Related Standards and References

1. OpenAPI: https://www.openapis.org/
2. AsyncAPI: https://www.asyncapi.com/docs
3. OpenTelemetry: https://opentelemetry.io/docs/
4. OpenLineage: https://openlineage.io/docs/
5. RFC 2119 (normative language): https://www.rfc-editor.org/rfc/rfc2119
6. RFC 8174 (normative language update): https://www.rfc-editor.org/rfc/rfc8174
