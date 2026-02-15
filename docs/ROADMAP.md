# Stricture Product Roadmap

Last updated: 2026-02-15

## Objective

Ship a production-ready lineage/drift workflow that is easy to adopt in real repos:

1. Minimal annotation authoring burden.
2. Deterministic CI gates.
3. Clear, visual blast-radius understanding.
4. Service-level flow-tier policy decisions with plain-language outputs.

## Priority Tracks

### Track A — Compact Authoring + Defaults

Goal: reduce manual annotation payload while preserving normalized artifact quality.

1. Define authoring-minimal keys and deterministic defaults.
2. Keep normalized artifact explicit for stable diffs/exports.
3. Provide compact + expanded annotation views in UI/docs.
4. Introduce `.stricture-history/` for computed versions and delta summaries.

Status: In progress.

### Track B — `stricture-helper` Utility

Goal: auto-decorate codebases using static analysis and existing contracts.

Spec: `docs/helper/SPEC.md` (draft)

Scope:

1. Scan source and infer candidate `stricture-source` blocks.
2. Reuse OpenAPI/AsyncAPI/protobuf metadata where available.
3. Fill defaults and suggest owner/escalation from registry.
4. Emit patch-ready comments (dry-run + apply modes).
5. Produce confidence labels and "needs human review" markers.

Planned phases:

1. **v0 (bootstrap):** generate minimal compact annotations.
2. **v1 (quality):** infer richer `sources` and contract refs.
3. **v2 (governance):** org policy packs, approval workflows, CI bot mode.

### Track C — `stricture-visualizer` Utility

Goal: render CLI/artifact output into an interactive topology + findings experience
equivalent to the live demo.

Scope:

1. Ingest `lineage-export` and `lineage-diff` outputs.
2. Render service graph and affected data paths.
3. Highlight source-of-problem and downstream blast radius.
4. Show findings narrative (`cause`, `impact`, `owner`, `remediation`).
5. Support local HTML export and CI artifact publishing.

Planned phases:

1. **v0 (reader):** static HTML report from JSON outputs.
2. **v1 (interactive):** filters, step replay, zoom-to-fit, path focus.
3. **v2 (ops):** history overlays and regression comparisons across runs.

### Track D — `stricture-server`

Goal: centralize lineage/drift signals across many company services and provide
continuous monitoring + flagging.

Scope:

1. Collect artifacts from many repositories/services (push + scheduled pull).
2. Build org-wide dependency graph and service ownership map.
3. Run continuous drift/invariant evaluation and alert routing.
4. Provide tenant/project/service dashboards and incident timelines.
5. Expose APIs/webhooks for CI, chatops, ticketing, and SIEM pipelines.
6. Support policy packs at org/team/service levels.
7. Serve policy packs to CI/local tooling with offline cache support and GitHub fallback refs.
8. Add org compliance check to ensure repos reference approved `strict:policy_url` values.
9. Serve flow-tier catalogs (`'strict:flows'`) and policy snapshots to local/CI clients.

Planned phases:

1. **v0 (ingest):** receive artifacts, store history, org-level diff API.
2. **v1 (monitor):** rules engine scheduling + alerting + escalation routing.
3. **v2 (platform):** multi-tenant RBAC, SSO, audit trails, policy governance.
4. **v3 (assist):** auto-remediation suggestions and AI copilot workflows.

## Sequencing

1. Finish defaults implementation and compact UX.
2. Deliver `stricture-helper` v0.
3. Deliver `stricture-visualizer` v0.
4. Deliver `stricture-server` v0 ingest + org diff API.
5. Integrate helper/visualizer/server into CI templates and docs quickstart.
