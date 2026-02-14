# Stricture Marketing Site + Interactive Demo

This package contains:

- Astro pre-rendered marketing pages (`/` and `/demo`)
- Cloudflare Worker API for live demo state
- Durable Object session engine for mutations, reruns, and escalation chains
- artifact-driven scenario pack generated from Stricture lineage logic

## Why this architecture

- Landing pages are static and cheap to serve.
- Demo API is stateful and fast via Durable Objects.
- You can keep the entire deployment on Cloudflare with low fixed cost.

## Local setup

1. Install dependencies:

```bash
cd site
npm install
```

2. Build static site:

```bash
npm run build
```

`npm run build` first regenerates the demo pack via:

```bash
npm run demo-pack
```

This command executes `go run ../cmd/demo-pack` and writes:

- `site/public/demo/demo-pack.json`
- `site/worker/src/generated/demo-pack.ts`

3. Run Worker + static assets locally:

```bash
npm run worker:dev
```

Then open the local URL from Wrangler output.

## Deploy

```bash
cd site
npm run build
npm run worker:deploy
```

## API surface used by demo page

- `POST /api/session` -> create isolated demo session
- `POST /api/session/{id}/mutations` -> inject drift mutation
- `POST /api/session/{id}/run` -> recompute findings and gate result
- `POST /api/session/{id}/policy` -> update warn/block behavior and fail threshold
- `POST /api/session/{id}/override` -> add temporary override comment equivalent
- `GET /api/session/{id}/escalation?serviceId=...` -> resolve contact chain
- `GET /api/session/{id}/snapshot` -> fetch current topology state

## Notes

- Findings are generated from real `internal/lineage` diff and escalation logic via the generated demo pack.
- The Worker executes fast policy/override/session orchestration over that precomputed pack for low runtime cost.
