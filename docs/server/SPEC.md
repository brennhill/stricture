# Stricture Server Spec

Last updated: 2026-02-15

## Scope

`stricture-server` collects lineage outputs from many repos/services and stores
immutable run records for org-wide drift visibility and policy workflows.

It is intentionally append-only for v0 and does not require a relational DB.

## License and Contribution Terms

- Server implementation is **AGPL-3.0** (`LICENSE`, `LICENSES.md`).
- Server-specific contribution terms are in `docs/server/CONTRIBUTING.md`.

## Goals (v0)

1. Accept lineage ingest payloads from CI.
2. Persist immutable records durably with deterministic keys.
3. Support both runtime models on day 1:
   - long-running local/server process
   - stateless cloud-function execution
4. Define storage and auth contracts that map cleanly to Cloudflare/R2 and S3.
5. Preserve both:
   - impact-gated findings for CI/policy workflows
   - non-impacting change events for publication/audit workflows
6. Preserve flow-tier context (`'strict:flows'`, `systems[].flows`) for policy
   evaluation and reporting.

## Non-Goals (v0)

1. Full dashboard UI.
2. Global query/index API beyond basic ingest location response.
3. Full multi-tenant RBAC/SSO.

## API (v0)

### `GET /healthz`

- Response: `200 {"status":"ok"}`.

### `POST /v1/artifacts`

- Request body:
  - `organization` (required)
  - `project` (required)
  - `service` (required)
  - `run_id` (optional; generated when missing)
  - `commit_sha` (optional)
  - `generated_at` (optional RFC3339)
  - `artifact` (required lineage-export JSON)
  - `diff` (optional lineage-diff JSON)
  - `summary` (optional text/markdown)
  - `metadata` (optional map)
- Response: `202 {"accepted":true,"run_id":"...","location":"..."}`.

## Finding vs Change Event Contract (Default)

Server ingest stores the full diff model. Downstream consumers apply this
default interpretation:

1. `downstream` impact -> finding candidate (warn/block based on policy).
2. `self_only` impact -> track/publish change event, no warning by default.
3. `unknown` impact -> low-severity finding by default unless policy overrides.

This keeps CI noise low while still publishing provider-facing contract changes.
Published human-readable messages should follow the plain-language template
library in `docs/LINEAGE-PLAIN-LANGUAGE-LIBRARY.md`.

When policy enables flow criticality, server-side consumers should evaluate
impacted flow tiers from affected service sets and include effective flow level
in gate/audit outputs.

## Runtime + Operation Model (Day 1 Requirement)

`stricture-server` supports two execution models from day 1 using the same API
contract and payload schema.

### Mode A: Continual Process

- Deployment: VM/container/local machine.
- Entrypoint: `cmd/stricture-server`.
- Behavior: long-running HTTP service.
- Best for: local development, stable internal endpoint, simple operations.

### Mode B: Cloud Function

- Deployment: serverless handler (Cloudflare/AWS/GCP/Azure style).
- Entrypoint contract: reuse `internal/server.NewHandler(...)` in a runtime
  adapter.
- Behavior: stateless request handling per invocation.
- Best for: low-ops managed runtime and elastic ingest scale.

### Day 1 Compatibility Rule

Both modes MUST:

1. Serve identical endpoints (`GET /healthz`, `POST /v1/artifacts`).
2. Enforce identical request validation and auth behavior.
3. Persist records into the same canonical storage key layout.
4. Return equivalent status codes and response schema.

## Storage Architecture

Detailed design: `docs/server/STORAGE.md`.

### Design Principles

1. Immutable ingest objects (no in-place mutation).
2. Deterministic object keys for replay and audit.
3. No mandatory SQL DB in v0.
4. Storage-driver abstraction so deployment can choose:
   - local filesystem
   - S3-compatible object store
   - Cloudflare R2

### Current Code Status

- Implemented driver: `fs`.
- Object-store drivers (`s3`, `r2`) are spec-defined and reserved for v1.

## Auth Architecture

1. `none` (local/dev only).
2. `token` (shared bearer token) for initial CI integration.
3. Planned next: OIDC/JWT verification and scoped service tokens.

## Runtime Config

- `STRICTURE_SERVER_ADDR` (default `:8085`)
- `STRICTURE_SERVER_DATA_DIR` (default `.stricture-server-data`)
- `STRICTURE_SERVER_STORAGE_DRIVER` (default `fs`)
- `STRICTURE_SERVER_OBJECT_BUCKET` (reserved for object storage backends)
- `STRICTURE_SERVER_OBJECT_PREFIX` (default `stricture`)
- `STRICTURE_SERVER_AUTH_MODE` (`none` or `token`)
- `STRICTURE_SERVER_INGEST_TOKEN` (required when auth mode is `token`)

Client binding notes:

1. Repos typically bind policy with `'strict:policy_url'`.
2. Repos may bind server endpoints with `'strict:server_url'` for helper/policy
   discovery workflows.

## Repo Layout

```text
cmd/stricture-server/
internal/server/
docs/server/
```

## Planned Extensions

1. Object-store implementations for `s3` and `r2`.
2. Read/query APIs (`latest`, `by service`, `by commit`).
3. Optional compact index artifacts for faster queries without SQL.
4. Webhook/alerting for high-severity blast radius.
5. Tenant-aware authz and audit trail endpoints.
6. Change-feed APIs for downstream/external subscribers:
   - `GET /v1/changes/{organization}/{project}/{service}`
   - `GET /v1/changes/{organization}/{project}/{service}?impact=self_only|downstream`
7. Policy distribution APIs for CI/local clients:
   - `GET /v1/policies/{policy_id}`
   - `GET /v1/policies/{policy_id}/versions/{version}`
   - metadata endpoint for caching hints.
8. Flow-catalog distribution APIs for CI/local clients:
   - `GET /v1/flows/{catalog_id}`
   - `GET /v1/flows/{catalog_id}/versions/{version}`

## Day 1 Acceptance Criteria

The server is considered day-1 complete only when:

1. Continual-process mode is runnable in local/dev and production.
2. Cloud-function adapter path is documented and validated with integration
   tests.
3. At least one durable object-store backend (`s3` or `r2`) is production-ready.
4. Mirror mode (local + durable object store) is available for developer-first
   workflows.
