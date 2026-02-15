# Stricture Server Storage + Auth Design

Last updated: 2026-02-15

## Objectives

1. Durable, append-only storage for lineage ingest records.
2. No mandatory relational DB for v0/v1.
3. Portable deployment:
   - simple server + filesystem
   - simple server + S3
   - Cloudflare Worker + R2
4. Deterministic keys so records can be reconstructed and replayed.

## Day 1 Operation Support Matrix

| Mode | Runtime | Storage | Status |
|---|---|---|---|
| Local process | `cmd/stricture-server` | `fs` | Implemented |
| Local process | `cmd/stricture-server` | `s3`/`r2` | Planned (required for day-1 completion) |
| Cloud function | runtime adapter + `NewHandler` | `s3`/`r2` | Planned (required for day-1 completion) |
| Hybrid mirror | process + replicator | `fs` + `s3`/`r2` | Planned (required for day-1 completion) |

## Record Model

Each ingest request becomes one immutable record envelope:

- identity: `organization`, `project`, `service`, `run_id`
- lineage payloads: `artifact`, optional `diff`, optional `summary`
- derived drift partitions:
  - `findings` (impact-gated policy candidates)
  - `change_events` (includes `self_only` drift for publication history)
- optional policy/flow context:
  - effective flow levels derived from `'strict:flows'` + `systems[].flows`
  - policy hard-block rationale when flow-criticality rules trigger
- optional service metadata context:
  - `owner_team`, `escalation`, `runbook_url`, `doc_root`
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
<prefix>/v1/org=<organization>/project=<project>/service=<service>/changes/date=<YYYY-MM-DD>.jsonl
```

Policy objects (draft):

```text
<prefix>/v1/policies/<policy_id>/versions/<version>/policy.yml
<prefix>/v1/policies/<policy_id>/latest.json
<prefix>/v1/flows/<catalog_id>/versions/<version>/flows.yml
```

Notes:

- `payload.json` is immutable.
- `latest.json` is a pointer object (best-effort convenience).
- per-day `manifest` enables query without SQL by listing one object.
- per-day `changes` stream includes publishable self-only and downstream events.

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

### Driver: `mirror` (planned day-1 completion)

- Primary write to `fs` for low-latency local durability.
- Secondary async replication to `s3` or `r2`.
- Guarantees:
  - ingest ACK after successful local write
  - replication status tracked and retryable
  - idempotent remote object keys by `run_id`

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
- Records write directly to S3 bucket (`s3` driver).
- Token auth (v0), then OIDC/JWT (v1).

### Profile B: Cloudflare Worker + R2

- Function/worker handler serves the same ingest endpoint contract.
- Writes payload directly into R2 with canonical keys.
- Cloudflare Access/JWT for auth.
- Optional queue for retries under burst load.

### Profile C: Hybrid

- Ingest to local `fs` for fast local ACK.
- Replicator mirrors to S3/R2 for ultra-durable storage.
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
2. v1/day-1 completion: `s3` and/or `r2` driver + cloud-function adapter test path.
3. v1.1: mirror driver (`fs` + object store replication queue/retry).
4. v2: optional index service + policy/routing APIs.
