#!/usr/bin/env bash
# check-fake-apis-live.sh — Run fake API docker-compose stack and verify live endpoints.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/tests/fake-apis/docker-compose.yml"
PROJECT_NAME="stricture-usecase-smoke"
REQUIRED_IMAGE="golang:1.22"

if ! command -v docker >/dev/null 2>&1; then
	echo "SKIP: docker is not installed."
	exit 0
fi

if ! command -v curl >/dev/null 2>&1; then
	echo "SKIP: curl is not installed."
	exit 0
fi

if ! docker info >/dev/null 2>&1; then
	echo "SKIP: docker daemon is unavailable."
	exit 0
fi

if ! docker image inspect "$REQUIRED_IMAGE" >/dev/null 2>&1; then
	echo "SKIP: required image '$REQUIRED_IMAGE' is not cached locally."
	exit 0
fi

compose() {
	docker compose -p "$PROJECT_NAME" -f "$COMPOSE_FILE" "$@"
}

cleanup() {
	compose down --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

wait_for_contains() {
	local url="$1"
	local needle="$2"
	local label="$3"
	local body=""
	local attempt

	for attempt in $(seq 1 30); do
		body="$(curl -fsS "$url" 2>/dev/null || true)"
		if [ -n "$body" ] && printf '%s' "$body" | grep -Fq "$needle"; then
			return 0
		fi
		sleep 1
	done

	echo "FAIL: ${label} did not match expected payload." >&2
	return 1
}

compose up -d >/dev/null

checks=(
	"logistics|18081|LogisticsGateway|logistics_01_shipment_eta_projection"
	"fintech|18082|FintechGateway|fintech_01_payment_authorization_decision"
	"media|18083|MediaGateway|media_01_track_metadata_unification"
	"ecommerce|18084|CommerceGateway|ecommerce_01_cart_pricing_waterfall"
	"governance|18085|GovernanceHub|governance_01_board_vote_tally"
)

for check in "${checks[@]}"; do
	IFS='|' read -r domain port service flow_id <<<"$check"
	base_url="http://127.0.0.1:${port}"

	wait_for_contains "$base_url/health" "\"service\":\"${service}\"" "${domain} /health"
	wait_for_contains "$base_url/api/v1/flows" "\"domain\":\"${domain}\"" "${domain} /api/v1/flows"
	wait_for_contains "$base_url/api/v1/simulate/${flow_id}?drift=source_version_changed" "\"simulatedDrift\":\"source_version_changed\"" "${domain} /api/v1/simulate"
	wait_for_contains "$base_url/api/v1/use-cases" "\"summary\":[" "${domain} /api/v1/use-cases"
done

echo "PASS: fake API live docker-compose smoke checks succeeded."
