# Cloudflare Hosting Plan for Stricture Marketing + Demo

## Objective

Host the marketing site and interactive demo with minimal monthly cost while retaining real-time interactivity.

## Deployment Model

- Astro pages are pre-rendered to static assets.
- Cloudflare Worker serves static assets and API endpoints from one deployable unit.
- Durable Objects hold per-session demo state.
- Demo findings are generated from real Stricture lineage artifact logic via `cmd/demo-pack`.

## Why this is cost-effective

- Static delivery is low-cost and cache-friendly.
- Worker API usage scales by request count.
- Durable Object sessions only exist when users run the demo.

## Build and deploy

```bash
make site-install
make site-build
make site-worker-deploy
```

## Local development

```bash
make site-dev
```

For end-to-end behavior with Worker routes:

```bash
make site-worker-dev
```

## Recommended environment setup

- Cloudflare account with Workers enabled.
- A custom domain attached to the Worker route.
- Optional analytics for demo conversion tracking.

## Suggested DNS/domain options

- Primary candidate: `stricture-lint.com`
- Alternatives:
  - `usestricture.com`
  - `stricturehq.com`
  - `stricture.dev`
  - `stricturelabs.com`

## Operational notes

- Keep marketing pages static for speed and reliability.
- Keep demo sessions short-lived and stateless from client perspective.
- `site-build` regenerates `site/public/demo/demo-pack.json` and `site/worker/src/generated/demo-pack.ts` from current lineage fixtures.
