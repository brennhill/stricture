# Stricture Server Storage + Auth Design

Last updated: 2026-02-14

## Objectives

1. Durable, append-only storage for lineage ingest records.
2. No mandatory relational DB for v0/v1.
3. Portable deployment:
   - simple server + filesystem
   - simple server + S3
   - Cloudflare Worker + R2
4. Deterministic keys so records can be reconstructed and replayed.

## Record Model

Each ingest request becomes one immutable record envelope:

- identity: `organization`, `project`, `service`, `run_id`
- lineage payloads: `artifact`, optional `diff`, optional `summary`
- metadata: commit SHA, timestamps, arbitrary metadata map
- server fields: `received_at`

## Canonical Key Layout (Object Store)

```text
<prefix>/v1/org=<organization>/project=<project>/service=<service>/date=<YYYY-MM-DD>/run=<run_id>/payload.json
```

Optional helper objects:

```text
<prefix>/v1/org=<organization>/project=<project>/service=<service>/latest.json
<prefix>/v1/org=<organization>/project=<project>/service=<service>/manifests/date=<YYYY-MM-DD>.jsonl
```

Notes:

- `payload.json` is immutable.
- `latest.json` is a pointer object (best-effort convenience).
- per-day `manifest` enables query without SQL by listing one object.

## Driver Strategy

### Driver: `fs` (implemented)

- Writes records to local path:

```text
<data_dir>/<organization>/<project>/<service>/<run_id>/payload.json
```

- Good for local/dev and single-node deployments.

### Driver: `s3` (planned v1)

- Writes canonical keys to S3-compatible bucket.
- Uses conditional write semantics where supported.
- Emits idempotent writes (`run_id` collision => reject or no-op by policy).

### Driver: `r2` (planned v1)

- Same key strategy as S3.
- Targets Cloudflare R2 via S3-compatible API or Worker binding layer.

## Query Without Database

v0/v1 query path avoids DB by combining:

1. deterministic key prefix listing
2. daily manifest objects (`jsonl`)
3. optional `latest.json` pointers

This supports:

- latest run for service
- runs by service/date range
- replaying historical diffs

When query volume grows, add optional index backends (SQLite/Postgres) as a
cache/accelerator, not as source of truth.

## Durability + Consistency

1. Object store is source of truth (S3/R2 durability).
2. Writes are append-only and idempotent by `run_id`.
3. Avoid in-place edits to historical records.
4. Keep metadata small; large artifacts can be chunked later if needed.

## Deployment Profiles

### Profile A: Simple Server + S3

- `stricture-server` binary handles ingest API.
- Records write directly to S3 bucket.
- Token auth (v0), then OIDC/JWT (v1).

### Profile B: Cloudflare Worker + R2

- Worker handles ingest endpoint.
- Writes payload directly into R2 with same canonical keys.
- Cloudflare Access/JWT for auth.
- Optional queue for retries under burst load.

### Profile C: Hybrid

- Ingest to local `fs` for low latency.
- Replicator mirrors to S3/R2 for durability.
- Server can continue from local cache during object-store incidents.

## Auth Design

### v0: Shared Bearer Token

- Mode: `STRICTURE_SERVER_AUTH_MODE=token`
- Secret: `STRICTURE_SERVER_INGEST_TOKEN`
- Suitable for internal CI only.

### v1: Federated Identity

1. GitHub Actions OIDC JWT verification.
2. Cloudflare Access JWT verification.
3. Short-lived service tokens minted by org identity provider.

Token claims should include service identity and allowed org/project scopes.

## Security Baseline

1. TLS required in all non-local environments.
2. Reject unknown JSON fields to prevent schema drift.
3. Cap request body size (10MB default).
4. Audit log all ingest acceptance/rejections.

## Migration Plan

1. v0: `fs` driver + token auth (implemented now).
2. v1: `s3` and `r2` drivers + OIDC/JWT auth.
3. v2: optional index service + policy/routing APIs.
