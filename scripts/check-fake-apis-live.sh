#!/usr/bin/env bash
# check-fake-apis-live.sh — Run fake API docker-compose stack and verify live endpoints.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/tests/fake-apis/docker-compose.yml"
PROJECT_NAME="stricture-usecase-smoke"
REQUIRED_IMAGE="golang:1.22"
START_EPOCH="$(date +%s)"

ASSERTIONS_TOTAL=0
ASSERTIONS_PASSED=0
OUTPUT_DUMPS=()

pass_assertion() {
	local message="$1"
	ASSERTIONS_TOTAL=$((ASSERTIONS_TOTAL + 1))
	ASSERTIONS_PASSED=$((ASSERTIONS_PASSED + 1))
	echo "PASS: ${message}"
}

fail_assertion() {
	local message="$1"
	echo "FAIL: ${message}" >&2
	exit 1
}

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

fetch_json_200() {
	local url="$1"
	local label="$2"
	local tmp_file
	local status
	local body=""
	local attempt

	tmp_file="$(mktemp)"

	for attempt in $(seq 1 30); do
		status="$(curl -sS -m 5 -o "$tmp_file" -w "%{http_code}" "$url" 2>/dev/null || true)"
		body="$(cat "$tmp_file" 2>/dev/null || true)"
		if [ "$status" = "200" ] && [ -n "$body" ]; then
			rm -f "$tmp_file"
			printf '%s' "$body"
			return 0
		fi
		sleep 1
	done

	rm -f "$tmp_file"
	fail_assertion "${label} did not return HTTP 200 with JSON payload."
}

assert_contains() {
	local body="$1"
	local needle="$2"
	local label="$3"

	if printf '%s' "$body" | grep -Fq "$needle"; then
		pass_assertion "$label"
		return 0
	fi

	fail_assertion "${label} (missing: ${needle})"
}

assert_count_equals() {
	local body="$1"
	local token="$2"
	local expected="$3"
	local label="$4"
	local count

	count="$(printf '%s' "$body" | grep -F -o "$token" | wc -l | tr -d ' ')"
	if [ "$count" = "$expected" ]; then
		pass_assertion "${label} (count=${count})"
		return 0
	fi

	fail_assertion "${label} expected count=${expected}, got count=${count}"
}

record_output_dump() {
	local label="$1"
	local body="$2"
	OUTPUT_DUMPS+=("### ${label}
${body}")
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
	echo "INFO: validating domain=${domain} service=${service} port=${port}"

	health_body="$(fetch_json_200 "$base_url/health" "${domain} /health")"
	record_output_dump "${domain} /health" "$health_body"
	assert_contains "$health_body" "\"service\":\"${service}\"" "${domain} /health has expected service"
	assert_contains "$health_body" "\"domain\":\"${domain}\"" "${domain} /health has expected domain"
	assert_contains "$health_body" "\"flowCount\":10" "${domain} /health reports 10 flows"

	flows_body="$(fetch_json_200 "$base_url/api/v1/flows" "${domain} /api/v1/flows")"
	record_output_dump "${domain} /api/v1/flows" "$flows_body"
	assert_contains "$flows_body" "\"service\":\"${service}\"" "${domain} /api/v1/flows has expected service"
	assert_contains "$flows_body" "\"domain\":\"${domain}\"" "${domain} /api/v1/flows has expected domain"
	assert_count_equals "$flows_body" "\"id\":\"${domain}_" "10" "${domain} /api/v1/flows returns 10 flow ids"

	simulate_body="$(fetch_json_200 "$base_url/api/v1/simulate/${flow_id}?drift=source_version_changed" "${domain} /api/v1/simulate")"
	record_output_dump "${domain} /api/v1/simulate/${flow_id}" "$simulate_body"
	assert_contains "$simulate_body" "\"flowId\":\"${flow_id}\"" "${domain} /api/v1/simulate uses requested flow id"
	assert_contains "$simulate_body" "\"simulatedDrift\":\"source_version_changed\"" "${domain} /api/v1/simulate applies requested drift"
	assert_contains "$simulate_body" "\"domain\":\"${domain}\"" "${domain} /api/v1/simulate has expected domain"

	use_cases_body="$(fetch_json_200 "$base_url/api/v1/use-cases" "${domain} /api/v1/use-cases")"
	record_output_dump "${domain} /api/v1/use-cases" "$use_cases_body"
	assert_count_equals "$use_cases_body" "\"useCase\":\"" "5" "${domain} /api/v1/use-cases publishes 5 use-case counters"
	assert_contains "$use_cases_body" "\"useCase\":\"drift_blocking\"" "${domain} /api/v1/use-cases includes drift_blocking"
	assert_contains "$use_cases_body" "\"useCase\":\"external_provider_drift\"" "${domain} /api/v1/use-cases includes external_provider_drift"
	assert_contains "$use_cases_body" "\"useCase\":\"escalation_chain\"" "${domain} /api/v1/use-cases includes escalation_chain"
	assert_contains "$use_cases_body" "\"useCase\":\"compliance_traceability\"" "${domain} /api/v1/use-cases includes compliance_traceability"
	assert_contains "$use_cases_body" "\"useCase\":\"multilang_contract_parity\"" "${domain} /api/v1/use-cases includes multilang_contract_parity"
done

echo "INFO: collected lineage outputs (ordered)"
for dump in "${OUTPUT_DUMPS[@]}"; do
	echo "-----"
	printf '%s\n' "$dump"
done

duration="$(( $(date +%s) - START_EPOCH ))"

echo "PASS: fake API live docker-compose smoke checks succeeded."
echo "  domains_tested: 5"
echo "  endpoints_tested_per_domain: 4"
echo "  assertions_passed: ${ASSERTIONS_PASSED}/${ASSERTIONS_TOTAL}"
echo "  duration_seconds: ${duration}"
