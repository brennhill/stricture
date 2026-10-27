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

1. Scan source and infer candidate `strict-source` blocks.
2. Reuse OpenAPI/AsyncAPI/protobuf metadata where available.
3. Fill defaults and suggest owner/escalation plus optional runbook/docs links from registry.
4. Emit patch-ready comments (dry-run + apply modes).
5. Produce confidence labels and "needs human review" markers.
6. Use traces to suggest missing service edges and candidate `sources` (optional enrichment).

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
2. Build org-wide dependency graph and service ownership/runbook/docs map.
3. Run continuous drift/invariant evaluation and alert routing.
4. Provide tenant/project/service dashboards and incident timelines.
5. Expose APIs/webhooks for CI, chatops, ticketing, and SIEM pipelines.
6. Support policy packs at org/team/service levels.
7. Serve policy packs to CI/local tooling with offline cache support and GitHub fallback refs.
8. Add org compliance check to ensure repos reference approved `stricture_policy_url` values.
9. Serve flow-tier catalogs (`strict_flows`) and policy snapshots to local/CI clients.
10. Record deployment events and show flow-level deployment recency.
11. Deliver admin Web UI (inventory + policy + tokens).
12. Add OAuth/OIDC (Okta) for admin access.

Planned phases:

1. **v0 (ingest):** receive artifacts, store history, org-level diff API.
2. **v1 (monitor):** rules engine scheduling + alerting + escalation routing + deployment ledger.
3. **v1.5 (admin):** admin Web UI + OAuth/OIDC login (Okta).
3. **v2 (platform):** multi-tenant RBAC, SSO, audit trails, policy governance.
4. **v3 (assist):** auto-remediation suggestions and AI copilot workflows.

### Track E — Trace Audit + Drift Validation

Goal: cross-check runtime traces (B3/W3C/OTel) against Stricture annotations to detect
unexpected edges, missing lineage, and "dishonest" or stale annotations.

Scope:

1. Ingest trace data (B3, W3C Trace Context, OpenTelemetry exports).
2. Compare observed service/field flows against annotated `sources`.
3. Flag missing edges, unexpected edges, and field provenance mismatches.
4. Provide coverage scores per service and per flow tier.
5. Emit findings + suggested annotation fixes or contract updates.

Planned phases:

1. **v0 (audit):** offline validator that checks traces vs annotations.
2. **v1 (assist):** helper hints to auto-suggest missing `sources` from traces.
3. **v2 (monitor):** server-side continuous audit with alerting and dashboards.

### Track F — Annotation Ergonomics + LLM Integration

Goal: make Stricture annotations trivially easy to write, self-evident to read,
and reliable for LLMs to generate and maintain.

Spec: `docs/data-lineage-annotations.md`, `docs/design/LLM-INTEGRATION.md`

Scope:

1. Redesigned `strict-source` format: multi-line block syntax, minimal
   `from ServiceName` shorthand, decomposed source ref grammar.
2. `strict-lineage.yml` sidecar files as primary authoring surface for
   lineage metadata (alternative to inline comments).
3. Expanded `strict helper` with `init-lineage`, `import-openapi`,
   `validate`, and `migrate` commands.
4. LLM prompt templates and autocomplete-friendly annotation patterns.
5. IDE extension spec for annotation validation, autocomplete, and coverage
   gutter indicators.
6. Canonical field ordering and synonym deprecation for consistency.

Design constraints:

- A team should adopt linter rules in under 5 minutes with zero annotations.
- Lineage annotations should require at most 2 manual fields per annotation.
- Every annotation must be writable by an LLM from code context alone.
- The expanded 18-field canonical form must never appear in source code.

Planned phases:

1. **v0 (format):** multi-line syntax, `from` shorthand, `strict-lineage.yml`.
2. **v1 (tooling):** expanded helper commands, CI staleness checks.
3. **v2 (intelligence):** LLM prompt templates, IDE extension, autocomplete.

### Track G — Contract Foundation (Pact-compatible)

Goal: make contract checking the foundation that lineage, drift and audit sit on.
Today the `CTR-*` rules are placeholders and lineage edges come from hand-written
annotations that nothing verifies. Lineage should be a graph over *verified*
contract edges.

Approach: reuse Pact's pact file format (specification v4) as the contract-fact model
instead of inventing one. Pact has no source annotations; the standard to adopt is the
file format, its matching rules, generators and provider states.

Scope:

1. Contract-fact model in `internal/model` that maps onto a Pact v4 interaction
   (HTTP, asynchronous message, synchronous message) with matching rules.
2. Reader for pact files v2 through v4 and writer for v4, so existing pacts can be
   verified statically and Stricture output can feed a Pact Broker or PactFlow.
3. Static extraction of interactions from client call sites and server handlers
   (Go and TypeScript first, since the adapters exist), replacing the runtime mock server.
4. **Derived consumer-driven contracts.** Pact only covers what the consumer's tests
   exercised, and the consumer declares the fields it needs by hand. Static extraction
   can record which response fields the consumer's code actually reads, and use that
   set as the consumer-driven contract. This is Pact's idea derived from code instead
   of declared, and it covers paths no test exercised.
5. Real implementations of `CTR-request-shape`, `CTR-response-shape` and
   `CTR-status-code-handling`, replacing the trigger-token stubs. Matching rules map
   onto these rules; strictness parity (`CTR-strictness-parity`) has no Pact equivalent
   and stays a Stricture differentiator.
6. Annotation conformance: `strict-source` annotations become claims verified against
   extracted facts (a mismatch is a `CTR-manifest-conformance` violation) and stay
   authoritative only for metadata such as owner, escalation and tier.
7. Rebase `lineage-export` and `lineage-diff` onto contract facts, so a drift finding
   means a verified edge changed.
8. Broker-equivalent in `stricture-server` (Track D): versioned contracts, branches and
   environments, and a `can-i-deploy` style check. In Pact this is a query over a
   matrix of consumer and provider versions and their verification results, asking
   whether the version being deployed has a passing verification against each version
   already deployed in the target environment. Pending and work-in-progress contracts
   keep new contracts from failing the provider's main build. Map `pending` onto
   Stricture's baseline and soak modes.

Planned phases:

1. **v0 (model):** fact model plus pact v2-v4 reader and v4 writer, with tests.
2. **v1 (extract):** Go and TypeScript extraction, derived field-usage contracts,
   real request and response shape rules.
3. **v2 (verify):** annotation conformance, lineage rebased on facts, static
   verification of imported pacts against provider code.
4. **v3 (broker):** versioned contracts, environments and `can-i-deploy` in the server;
   async and message contracts, provider-state equivalents.

Research findings and decisions (Pact docs reviewed Oct 2026):

- **Pact plugins.** A plugin is a separate process with a manifest (`pact-plugin.json`:
  `manifestVersion`, `pluginInterfaceVersion`, `name`, `version`, `executableType`,
  `entryPoint`, `dependencies`) found under `$PACT_PLUGIN_DIR` or `~/.pact/plugins`.
  The driver starts it, reads `{port, serverKey}` from its stdout, and talks gRPC
  (`InitPlugin`, `UpdateCatalogue`, `CompareContents`, `ConfigureInteraction`,
  `GenerateContent`, mock-server and verification RPCs). Plugins advertise
  catalogue entries (`content-matcher`, `content-generator`, `matcher`, `transport`) keyed by
  content type. The protocol exists to run mock servers and replay requests, which static
  analysis does not need. Decision: do not build a plugin host in v0-v2. Instead
  (a) round-trip `plugins` metadata and `pluginConfiguration` losslessly through the reader
  and writer, (b) ship built-in static analyzers for protobuf/gRPC and parse `.proto`
  files directly, and (c) in v3 consider a client for the single relevant RPC
  (`CompareContents`) for formats Stricture cannot parse. Not yet reviewed: the v2 plugin
  proto (`plugin_v2.proto`).
- **can-i-deploy.** Inputs are pacticipant plus version, a target (`--to-environment`, or
  `--latest`, `--branch`, `--main-branch`, `--to` tag), `--ignore`, `--retry-while-unknown`
  and `--retry-interval`, `--dry-run` and `--output json|table`. Exit code is 0 or 1.
  Prerequisites are environments and recorded deployments (`record-deployment`,
  `record-release`). It queries the matrix, with `latestby` set to `cvp` for one selector
  and `cvpv` for two. The docs do not state how rows are combined; the usual reading is
  that every row must be successful and an unverified combination counts as unknown.
  Decision: the Stricture server verifies statically when a contract is published, so
  results are never unknown and the retry machinery is unnecessary. Ship `strict can-i-deploy`
  with the same flags and exit codes so Pact users can switch, plus a blast-radius
  section in the output.
- **Provider states.** In Pact a state (name plus params) makes the provider's data
  match so a replayed request is reproducible. Statically there is no data to set up, but
  the state still determines which response variant the consumer expects (for example
  "user does not exist" implies a 404 with an error body). Decision: treat a state as a
  scenario label that selects a provider response variant (status code and body shape).
  Imported states are kept in the model and verified by matching the expected status to an
  extracted variant; unmapped states are reported as informational, not failures.
  Extracted contracts get synthetic state names from guard conditions in the handler
  (for example a not-found branch). An optional provider-side annotation maps a Pact state
  name to a handler branch where inference cannot.

## Sequencing

0. **Track G (contract foundation) is the prerequisite for lineage drift, audit and policy gates.** Start with its v0 model and pact reader/writer.
1. Finish defaults implementation and compact UX.
2. **Deliver annotation format v2 spec and parser support.**
3. Deliver `stricture-helper` v0 (expanded scope).
4. Deliver `stricture-visualizer` v0.
5. Deliver `stricture-server` v0 ingest + org diff API.
6. **Deliver LLM integration utilities and IDE extension.**
7. Integrate helper/visualizer/server into CI templates and docs quickstart.
