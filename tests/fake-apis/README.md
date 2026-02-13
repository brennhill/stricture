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

## Run

From repository root:

```bash
docker compose -f tests/fake-apis/docker-compose.yml up --build
```

## Example calls

```bash
curl -s http://localhost:18081/health | jq
curl -s http://localhost:18082/api/v1/flows | jq '.flows | length'
curl -s "http://localhost:18083/api/v1/simulate/media_01_track_metadata_unification?drift=source_version_changed" | jq
```
