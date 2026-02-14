# Stricture Policy CLI Contract (Draft)

Last updated: 2026-02-14
Status: Draft v0

This document locks the command contract for policy URL binding, cache usage,
and CI compliance checks.

## Primary Goal

Allow organizations to enforce one approved `'strict:policy_url'` across repos,
while local/CI runs can use cached policy when network is unavailable.

## Commands

## `stricture policy fetch`

Fetches policy from `'strict:policy_url'` and updates local cache.

```bash
stricture policy fetch \
  --config .stricture.yml \
  --cache-dir .stricture-cache/policies
```

Behavior:

1. reads `'strict:policy_url'` and optional `'strict:policy_sha256'`
2. fetches policy if cache is missing/stale
3. validates hash when provided
4. stores policy + metadata in cache

## `stricture policy resolve`

Resolves effective policy using cache-first logic.

```bash
stricture policy resolve \
  --config .stricture.yml \
  --cache-dir .stricture-cache/policies
```

Behavior:

1. uses fresh cache when available
2. otherwise fetches URL
3. if offline and stale cache exists, uses stale cache and emits info notice
4. exits non-zero if no resolvable policy exists

## `stricture policy verify-ref`

Enforces repo binding to org-approved policy URL.

```bash
stricture policy verify-ref \
  --config .stricture.yml \
  --expected-url https://policies.example.com/stricture/prod.yml
```

Behavior:

1. fails if `'strict:policy_url'` missing
2. fails if URL differs from expected (or allowlist)
3. optional: verifies `'strict:policy_sha256'` presence in strict mode

## `stricture policy lint`

Validates policy pack structure and semantics.

```bash
stricture policy lint --policy .stricture-policy.yml
```

Behavior:

1. validates against `docs/schemas/lineage-policy-pack.schema.json`
2. validates known keys and severity values
3. fails on unknown required-key paths

## Exit Codes

1. `0`: success
2. `1`: validation/policy mismatch
3. `2`: config/usage error
4. `3`: fetch/cache resolution failure

## CI Patterns

## Single-repo guard

```bash
stricture policy verify-ref \
  --config .stricture.yml \
  --expected-url "$STRICTURE_POLICY_URL"
```

## Full policy warm-up + validation

```bash
stricture policy fetch --config .stricture.yml
stricture policy lint --policy .stricture-policy.yml
stricture policy resolve --config .stricture.yml
```

## Org-wide compliance scan

Run in a central governance repo or scheduled workflow:

1. iterate governed repos
2. read each repo `.stricture.yml`
3. run `stricture policy verify-ref --expected-url ...`
4. fail/report non-conforming repos
