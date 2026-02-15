# Stricture Visual Demo Storyboard

## Goal

Define a production-grade interactive demo suite that shows:

- service lifecycle across container bring-up
- endpoint and field-level lineage annotations
- live drift injection and re-evaluation
- severity transitions (green to red) on API graph edges
- ownership/runbook/docs context and escalation routing for bad data incidents

This storyboard is implementation-focused and maps directly to current Stricture capabilities plus minimal demo-specific glue.

## Demo Outcomes

A successful demo makes these claims obvious to humans and tools:

- Stricture can trace every output field to one or more sources.
- Stricture can detect schema/version/contract drift quickly and deterministically.
- Stricture can explain blast radius and who must respond.
- Stricture can support both warn and block policies with flow-tier governance.
- Stricture can run across a multi-service topology, not just one API in isolation.

## Runtime Architecture

## Components

- `tests/fake-apis/docker-compose.yml`: 5 domain services, 50 total flows.
- Stricture CLI:
  - `stricture lineage-export`
  - `stricture lineage-diff`
  - `stricture lineage-escalate`
- Demo UI web app:
  - graph canvas
  - service inspector
  - field lineage inspector
  - policy/escalation panels
- Demo orchestrator service:
  - executes scenario scripts
  - injects drift
  - triggers stricture runs
  - streams results to UI (SSE or WebSocket)

## Data Contracts for Demo UI

- `GraphSnapshot`
  - nodes: services/systems
  - edges: data links by `field_id`
  - edge_status: `healthy|warning|blocked`
  - last_run_id and timestamps
- `RuleFinding`
  - rule_id, severity, service, field_id, summary, remediation
- `EscalationChain`
  - ordered contacts from affected service upstream
- `ScenarioEvent`
  - action name, params, start/end, expected result

## Visual Language

- `healthy`: green edge/node glow
- `warning`: amber pulse + badge
- `blocked`: red animated edge break + banner
- `stale external source`: orange clock badge (`as_of` age)
- `flow-tier block`: red gate chip with explicit policy reason (for example
  `checkout level 1: risk of order loss`)

## Core Pages

## Page 1: Topology Bring-Up

## Purpose

Show that Stricture sees the system like operators do: real services, dependencies, and field-bearing APIs.

## UI

- left: animated service map (containers booting gray to green)
- right: live event feed (`docker up`, endpoint discovery, first scan)
- bottom: run summary strip (`services`, `flows`, `annotated`, `coverage`)

## Interaction

- click node opens service details:
  - endpoints list
  - `stricture-truth` summary
  - flow count and annotation coverage

## Backing calls

- `GET /health`
- `GET /api/v1/flows`
- `GET /api/v1/stricture-truth`

## Expected proof

- all 5 domain services visible
- 50 flows visible (10 per domain)
- initial graph all green

## Page 2: Service and Endpoint Inspector

## Purpose

Let users inspect one service as a normal API while still exposing lineage truth data.

## UI

- endpoint tabs:
  - `/health`
  - `/api/v1/flows`
  - `/api/v1/flows/{id}`
  - `/api/v1/use-cases`
  - `/api/v1/stricture-truth`
- payload viewer with field highlights
- quick links to fields in graph

## Interaction

- select flow ID to inspect one workflow payload
- toggle between raw payload and annotation-aware view

## Expected proof

- users see this is a normal server shape
- `stricture-truth` acts as a checksum-style expectation endpoint

## Page 3: Field Lineage Explorer

## Purpose

Show per-field provenance across multiple sources, including enrichment.

## UI

- center: selected field path and `field_id`
- graph edges only for selected field (dim others)
- annotation card:
  - `sources`
  - `flow` (for example: `from @X enriched @self`)
  - note/spec link/class reference
  - source versions and min supported versions

## Interaction

- click `response.<field>` in payload to jump to lineage chain
- click source systems to inspect each upstream contract reference

## Expected proof

- multi-source lineage is explicit
- enrichment is transparent and human-readable

## Page 4: Drift Injection Lab (Real-Time)

## Purpose

Demonstrate that changes in API outputs immediately alter Stricture risk and graph status.

## UI

- left: editable mutation panel
  - field remove
  - type change
  - enum change
  - source_version bump
  - external `as_of` rollback/advance
- center: graph with animated edge status transitions
- right: findings panel grouped by severity

## Interaction

- user applies mutation
- orchestrator reruns:
  - `stricture lineage-export`
  - `stricture lineage-diff`
- graph updates in place (green to amber/red)

## Expected proof

- status transitions are fast and deterministic
- each failed edge links to specific finding and remediation

## Page 5: Policy Mode and Flow-Tier Governance

## Purpose

Show operational policy behavior: warn vs block plus flow-tier criticality and
explicit policy rationale.

## UI

- policy toggle: `warn` or `block`
- threshold selector: `high|medium|low`
- flow policy panel:
  - flow ID and level
  - effective level touched by finding
  - hard-block reason (when configured)

## Interaction

- rerun with `--mode warn` and `--mode block`
- show flow-tier escalation and rerun
- show policy reason that caused block (severity threshold vs flow tier)

## Expected proof

- users understand enforcement semantics
- flow-tier gates are understandable and attributable

## Page 6: Escalation Chain Explorer

## Purpose

Answer "data is bad at service X, who do I call first, then next?"

## UI

- incident input: service + field
- escalation chain timeline (depth 0..N)
- contact cards:
  - owner
  - runbook/docs links (when available)
  - contact channels
  - reason and dependency edge

## Backing command

- `stricture lineage-escalate --service <svc> --artifact <artifact> --systems docs/config-examples/lineage-systems.yml --max-depth 5`

## Expected proof

- escalation path is generated from lineage, not tribal knowledge
- missing registry entries degrade gracefully to annotation owner/escalation (with runbook/docs omitted)

## Page 7: Release Gate (Baseline vs Head)

## Purpose

Show deploy-time decision logic and blast-radius reporting.

## UI

- baseline artifact vs current artifact diff
- counts by severity and by domain
- affected fields/services list
- gate decision banner (`PASS` or `BLOCK`)

## Interaction

- choose compare pair (last release vs current)
- inspect exact changed sources/versions/contracts

## Expected proof

- CI/CD decision is explainable, not opaque
- teams can fix quickly with precise drift context

## Scenario Library (Initial 15, then scale to 50)

Start with 3 per domain, then add full 50-flow matrix.

## Logistics

- ETA projection: external carrier schema drift
- customs clearance status: enum change
- cold chain compliance: source version mismatch

## Fintech

- payment authorization decision: field type widening
- settlement reconciliation: missing source annotation
- sanctions signal: external provider `as_of` stale

## Media

- track metadata unification: enrichment rule change
- royalty projection: upstream contract ref change
- moderation flag resolution: source removed from merge chain

## Ecommerce

- cart pricing waterfall: tax provider version drift
- promotion eligibility: conflicting source precedence
- loyalty reconciliation: field removed during migration

## Governance

- board vote tally: source system ownership missing
- control attestation snapshot: escalation chain gap
- disclosure readiness: override expired and now blocking

## Canonical 8-Minute Live Demo Script

## Minute 0-1

- open topology
- show 5 services and 50 flows discovered

## Minute 1-2

- inspect one service as a normal API
- open `stricture-truth` endpoint and checksum

## Minute 2-3

- select one field and show multi-source lineage + enrichment note

## Minute 3-5

- inject drift (type change + external `as_of` rollback)
- rerun and show green to red edge transitions
- open findings and blast radius

## Minute 5-6

- switch from warn to block
- show flow-tier hard-block reason and effective tier context

## Minute 6-7

- run escalation chain for affected service
- show "who to call" in dependency order

## Minute 7-8

- show release gate diff and final deploy decision
- revert mutation and rerun back to green

## Implementation Plan

## Phase A: Demo Data and Orchestration

- define scenario script format (`yaml` or `json`)
- implement orchestrator endpoints for:
  - run scenario step
  - apply mutation
  - run stricture
  - stream run events

## Phase B: Graph + Inspector UI

- build topology canvas with status animation states
- add service/endpoint payload inspector
- add field lineage detail panel

## Phase C: Policy + Escalation + Gate

- implement warn/block controls
- implement flow-tier policy visibility and reason chips
- implement escalation chain viewer
- implement baseline-vs-head gate screen

## Phase D: Demo Quality Hardening

- snapshot tests for visual states
- scripted end-to-end demo replay test
- deterministic seeded mutations
- no-flake timing budget for reruns

## Acceptance Criteria

- one-click run starts full demo stack
- each page has at least one scripted failure and one scripted fix
- every red edge maps to specific finding and remediation text
- escalation chain always returns deterministic ordering
- warn/block behavior matches CLI semantics exactly
- demo script completes in under 8 minutes without manual shell intervention

## Nice-to-Have Enhancements

- timeline scrubber to replay earlier states
- architecture export as SVG/PNG with findings overlays
- side-by-side compare of two domains
- persona mode:
  - platform SRE view
  - service owner view
  - governance/compliance view

## Suggested Next Build Targets

- `make demo-stack-up`
- `make demo-seed`
- `make demo-ui`
- `make demo-scenario SMOKE=1`
- `make demo-e2e`
