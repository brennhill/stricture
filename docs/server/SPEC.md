# Stricture Server Spec

Last updated: 2026-02-14

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
3. Provide health + ingest API baseline for local server deployment.
4. Define storage and auth contracts that map cleanly to Cloudflare/R2 and S3.

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
