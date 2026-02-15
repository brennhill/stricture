# Stricture Demo Setup & Regeneration

This repo’s live demo is a static snapshot generated from the Go engine and served by the Cloudflare Worker. Below is the end-to-end flow and the manual steps to regenerate.

## How the demo is built (flow)
1) **Generate demo pack** (Go):
   - Runs `go run ./cmd/demo-pack --artifact ./tests/lineage/usecases/current.json --systems ./tests/lineage/usecases/systems.yml --out-json ./site/public/demo/demo-pack.json --out-ts ./site/worker/src/generated/demo-pack.ts`
   - Output: JSON + TS artifacts describing services, fields, mutation scenarios, escalation chains, policy metadata.
2) **Build site** (Astro/Vite):
   - `npm run build` from `site/`
   - Pulls the generated TS pack into the Worker bundle and the JSON into `site/public/demo/`.
   - Emits static marketing pages + demo page into `site/dist`.
3) **Deploy Worker + assets** (Wrangler):
   - `npm run worker:deploy` uploads `dist` assets and Worker code; Durable Object runs the session logic against the generated pack.

## CI/CD guarantees
- Workflow: `.github/workflows/site-build.yml`
- On `main` pushes and PRs touching site/demo-pack/go files:
  - `npm ci`
  - `npm run demo-pack` (fails build if demo pack generation fails)
  - `npm run test` (fails build on demo unit/worker/invariant regressions)
  - `npm run build`
- On `main` with `CF_DEPLOY_ENABLED=true` and `CLOUDFLARE_API_TOKEN` secret set:
  - Re-runs `npm run demo-pack`
  - Deploys via `npm run worker:deploy`
- Result: "initial run" data is always present because the pack must regenerate before build/deploy.

## Manual regeneration (local)
```bash
# From repo root
cd site
npm install           # first time
npm run demo-pack     # regenerate demo data/artifacts
npm run build         # rebuild static site and worker bundle
npm run worker:deploy # deploy to Cloudflare (requires CLOUDFLARE_API_TOKEN)
```

## Manual regeneration (CI-like, without deploy)
```bash
cd site
npm ci
npm run demo-pack
npm run test
npm run build
```

## Troubleshooting
- **Demo stuck at “Awaiting first run”**: ensure `npm run demo-pack` was executed and artifacts are fresh; rebuild and redeploy.
- **Routes not attached**: confirm Cloudflare token has Zone:Read + Workers Routes:Edit and rerun `npm run worker:deploy`.
- **Stale data**: rerun `npm run demo-pack` and redeploy.
