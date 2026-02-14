# Fake Domain APIs (Go + Docker Compose)

This stack runs five fake APIs backed by the 50-flow use-case catalog:

- logistics (`LogisticsGateway`)
- fintech (`FintechGateway`)
- media (`MediaGateway`)
- ecommerce (`CommerceGateway`)
- governance (`GovernanceHub`)

Each service exposes:

- `GET /health`
- `GET /api/v1/flows`
- `GET /api/v1/flows/{flow_id}`
- `GET /api/v1/simulate/{flow_id}?drift=<change_type>`
- `GET /api/v1/use-cases`
- `GET /api/v1/stricture-truth`

## Run

From repository root:

```bash
docker compose -f tests/fake-apis/docker-compose.yml up --build
```

## Smoke check

Run the live endpoint smoke check used by `usecase-agent`:

```bash
./scripts/check-fake-apis-live.sh
```

Pass criteria:
- Starts all 5 services via docker-compose.
- Validates 5 endpoints per service (`/health`, `/api/v1/flows`, `/api/v1/simulate/{id}`, `/api/v1/use-cases`, `/api/v1/stricture-truth`).
- Asserts service/domain identity, flow counts, requested drift simulation, and required use-case categories.
- Prints all endpoint payloads in-order (all domains, all endpoints) so consumers can inspect every lineage output from one run.
- Prints assertion totals and duration.

This exits `0` with a `SKIP:` message when Docker daemon is unavailable.

If `golang:1.22` is missing locally, the script automatically runs `docker pull golang:1.22` before starting the stack (and fails if pull cannot complete).

## Example calls

```bash
curl -s http://localhost:18081/health | jq
curl -s http://localhost:18082/api/v1/flows | jq '.flows | length'
curl -s "http://localhost:18083/api/v1/simulate/media_01_track_metadata_unification?drift=source_version_changed" | jq
```
