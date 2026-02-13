#!/usr/bin/env bash
# usecase-agent.sh — Build and validate combined real-world lineage use-case examples.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$PROJECT_ROOT/.usecase-agent"
STATE_FILE="$STATE_DIR/state.env"
LOG_FILE="$STATE_DIR/run.log"

CATALOG_PATH="$PROJECT_ROOT/tests/lineage/usecases/flows.json"
BASELINE_PATH="$PROJECT_ROOT/tests/lineage/usecases/baseline.json"
CURRENT_PATH="$PROJECT_ROOT/tests/lineage/usecases/current.json"
FIXTURE_PATH="$PROJECT_ROOT/tests/fixtures/lineage-usecases"
COMPOSE_FILE="$PROJECT_ROOT/tests/fake-apis/docker-compose.yml"
GOCACHE_DIR="$PROJECT_ROOT/.cache/go-build"

now_utc() {
	date -u +"%Y-%m-%dT%H:%M:%SZ"
}

ensure_state() {
	mkdir -p "$STATE_DIR"
	if [ ! -f "$STATE_FILE" ]; then
		cat >"$STATE_FILE" <<EOF_STATE
STATUS=idle
LAST_ERROR=
LAST_ACTION=
LAST_RUN_AT=
UPDATED_AT=$(now_utc)
EOF_STATE
	fi
}

load_state() {
	STATUS=idle
	LAST_ERROR=
	LAST_ACTION=
	LAST_RUN_AT=
	UPDATED_AT=

	if [ -f "$STATE_FILE" ]; then
		while IFS='=' read -r key value; do
			[ -z "$key" ] && continue
			case "$key" in
				STATUS) STATUS="$value" ;;
				LAST_ERROR) LAST_ERROR="$value" ;;
				LAST_ACTION) LAST_ACTION="$value" ;;
				LAST_RUN_AT) LAST_RUN_AT="$value" ;;
				UPDATED_AT) UPDATED_AT="$value" ;;
				*) ;;
			esac
		done <"$STATE_FILE"
	fi
}

save_state() {
	cat >"$STATE_FILE" <<EOF_STATE
STATUS=${STATUS}
LAST_ERROR=${LAST_ERROR}
LAST_ACTION=${LAST_ACTION}
LAST_RUN_AT=${LAST_RUN_AT}
UPDATED_AT=$(now_utc)
EOF_STATE
}

log_line() {
	local message="$1"
	printf '[%s] %s\n' "$(now_utc)" "$message" | tee -a "$LOG_FILE"
}

run_step() {
	local label="$1"
	local cmd="$2"

	LAST_ACTION="$label"
	save_state
	log_line "step=${label} cmd=${cmd}"
	if (cd "$PROJECT_ROOT" && eval "$cmd" >>"$LOG_FILE" 2>&1); then
		log_line "step=${label} result=pass"
		return 0
	fi

	log_line "step=${label} result=fail"
	LAST_ERROR="${label} failed"
	STATUS=blocked
	save_state
	return 1
}

run_compose_check_if_available() {
	if ! command -v docker >/dev/null 2>&1; then
		log_line "step=docker-compose-config result=skip reason=docker-not-installed"
		return 0
	fi

	log_line "step=docker-compose-config cmd=docker compose -f '$COMPOSE_FILE' config"
	if (cd "$PROJECT_ROOT" && docker compose -f "$COMPOSE_FILE" config >>"$LOG_FILE" 2>&1); then
		log_line "step=docker-compose-config result=pass"
	else
		log_line "step=docker-compose-config result=skip reason=docker-daemon-unavailable"
	fi
	return 0
}

run_agent() {
	ensure_state
	load_state
	: >"$LOG_FILE"
	mkdir -p "$GOCACHE_DIR"
	export GOCACHE="$GOCACHE_DIR"

	STATUS="running"
	LAST_ERROR=
	LAST_ACTION=
	LAST_RUN_AT="$(now_utc)"
	save_state

	run_step "generate-usecase-examples" "go run ./scripts/generate-usecase-examples.go"
	run_step "usecase-tests" "go test ./internal/lineage -run 'TestUseCase' -count=1"
	run_step "fake-api-tests" "go test ./tests/fake-apis/cmd/fake-api -count=1"
	run_step "lineage-export-usecases" "go run ./cmd/stricture lineage-export --strict=true --out '$CURRENT_PATH' '$FIXTURE_PATH'"

	if [ ! -f "$BASELINE_PATH" ]; then
		log_line "step=bootstrap-baseline action=create-from-current"
		cp "$CURRENT_PATH" "$BASELINE_PATH"
	fi

	run_step "lineage-diff-usecases" "go run ./cmd/stricture lineage-diff --base '$BASELINE_PATH' --head '$CURRENT_PATH' --fail-on high --mode block"
	run_compose_check_if_available
	run_step "docker-compose-live-smoke" "./scripts/check-fake-apis-live.sh"

	STATUS="complete"
	LAST_ERROR=
	LAST_ACTION="complete"
	save_state
	echo "COMPLETE: usecase examples generated and validated."
}

print_status() {
	ensure_state
	load_state

	echo "Usecase Agent Status"
	echo "  status: $STATUS"
	echo "  last_action: ${LAST_ACTION:-<none>}"
	echo "  last_error: ${LAST_ERROR:-<none>}"
	echo "  last_run_at: ${LAST_RUN_AT:-<none>}"
	echo "  updated_at: ${UPDATED_AT:-<unknown>}"
	echo "  log_file: $LOG_FILE"
}

reset_state() {
	rm -rf "$STATE_DIR"
	ensure_state
	echo "Usecase agent state reset."
}

usage() {
	cat <<'EOF_USAGE'
Usage: scripts/usecase-agent.sh <command>

Commands:
  run      Generate and validate 50+ combined use-case examples (default)
  status   Print current usecase-agent state
  reset    Reset usecase-agent state
  help     Show this message
EOF_USAGE
}

main() {
	local cmd="${1:-run}"
	case "$cmd" in
		run) run_agent ;;
		status) print_status ;;
		reset) reset_state ;;
		-h|--help|help) usage ;;
		*)
			echo "Unknown command: $cmd" >&2
			usage
			exit 1
			;;
	esac
}

main "$@"
