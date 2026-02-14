# Stricture Deploy Guide (Cloudflare)

Last updated: 2026-02-14

This runbook deploys the marketing site + interactive demo from GitHub to Cloudflare so pushes to `main` auto-publish.

## What gets deployed

- Static site built from `site/src` (Astro) and output to `site/dist`
- Worker API from `site/worker/src/index.ts`
- Durable Object `DemoSession` defined in `site/worker/wrangler.toml`

## Prerequisites

- Cloudflare account with Workers enabled
- GitHub repo connected to Cloudflare
- Domain delegated to Cloudflare DNS: `stricture-lint.com`

## One-time Cloudflare setup (Git auto-deploy)

1. In Cloudflare Dashboard, go to `Workers & Pages` and create/select Worker `stricture-demo-site`.
2. Open `Builds` and connect the GitHub repository.
3. Configure build/deploy:

```text
Production branch: main
Build root directory: site
Build command: npm ci && npx astro build
Deploy command: npx wrangler deploy --config worker/wrangler.toml
Non-production deploy command: npx wrangler versions upload --config worker/wrangler.toml
```

4. Save and trigger initial build.

## Domain binding

1. In Worker `stricture-demo-site`, open `Settings` -> `Domains & Routes`.
2. Add custom domain:
   - `stricture-lint.com`
3. Optional:
   - Add `www.stricture-lint.com`
   - Set redirect `www` -> apex (`stricture-lint.com`)

## Required repo files already used by deploy

- `site/worker/wrangler.toml`
- `site/package.json`
- `site/public/demo/demo-pack.json`
- `site/worker/src/generated/demo-pack.ts`

## Local verification before push

From repo root:

```bash
make site-install
make site-build
make site-worker-dev
```

Or from `site/`:

```bash
npm ci
npx astro build
npx wrangler dev --config worker/wrangler.toml
```

## Deploy flow

1. Push commit to `main`.
2. Cloudflare Build runs install/build/deploy automatically.
3. Verify:
   - `https://stricture-lint.com/`
   - `https://stricture-lint.com/demo`
4. Smoke-check API:

```bash
curl -s -X POST https://stricture-lint.com/api/session | jq .
```

## Troubleshooting

- Build fails with missing dependencies:
  - Ensure `site/package-lock.json` is committed.
- Worker deploy fails on Durable Object migration:
  - Confirm `site/worker/wrangler.toml` includes migration tags and `DemoSession`.
- Domain not resolving:
  - Confirm DNS is managed by Cloudflare and custom domain is attached to the Worker.
- Demo data stale:
  - Regenerate and commit artifacts:

```bash
go run ./cmd/demo-pack \
  --artifact ./tests/lineage/usecases/current.json \
  --systems ./tests/lineage/usecases/systems.yml \
  --out-json ./site/public/demo/demo-pack.json \
  --out-ts ./site/worker/src/generated/demo-pack.ts
```

## CI/CD behavior summary

- Source of truth: GitHub `main`
- Trigger: push to `main`
- Runtime: Cloudflare Worker + Durable Objects
- Cost profile: static assets + lightweight Worker requests
