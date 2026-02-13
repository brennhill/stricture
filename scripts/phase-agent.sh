#!/usr/bin/env bash
# phase-agent.sh — Autonomous phase foreman for phases 3, 4, and 5.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$PROJECT_ROOT/.phase-agent"
STATE_FILE="$STATE_DIR/state.env"
LOG_FILE="$STATE_DIR/run.log"

PHASES=(3 4 5)

phase_commands() {
    local phase="$1"
    case "$phase" in
        3)
            cat <<'CMDS'
./scripts/validate-gate.sh --phase 3
./scripts/check-no-stubs.sh --phase 3
make test-phase3
CMDS
            ;;
        4)
            cat <<'CMDS'
./scripts/validate-gate.sh --phase 4
./scripts/check-no-stubs.sh --phase 4
make test-phase4
CMDS
            ;;
        5)
            cat <<'CMDS'
make test-phase5
go test -tags=integration -timeout=120s ./tests/integration/...
CMDS
            ;;
        *)
            return 1
            ;;
    esac
}

now_utc() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

ensure_state() {
    mkdir -p "$STATE_DIR"
    if [ ! -f "$STATE_FILE" ]; then
        cat >"$STATE_FILE" <<EOF
CURRENT_PHASE=3
STATUS=idle
COMPLETED_PHASES=
LAST_ERROR=
UPDATED_AT=$(now_utc)
EOF
    fi
}

load_state() {
    # shellcheck disable=SC1090
    source "$STATE_FILE"
    CURRENT_PHASE="${CURRENT_PHASE:-3}"
    STATUS="${STATUS:-idle}"
    COMPLETED_PHASES="${COMPLETED_PHASES:-}"
    LAST_ERROR="${LAST_ERROR:-}"
    UPDATED_AT="${UPDATED_AT:-}"
}

save_state() {
    cat >"$STATE_FILE" <<EOF
CURRENT_PHASE=${CURRENT_PHASE}
STATUS=${STATUS}
COMPLETED_PHASES=${COMPLETED_PHASES}
LAST_ERROR=${LAST_ERROR}
UPDATED_AT=$(now_utc)
EOF
}

phase_done() {
    local phase="$1"
    case ",${COMPLETED_PHASES}," in
        *,"$phase",*) return 0 ;;
        *) return 1 ;;
    esac
}

mark_phase_done() {
    local phase="$1"
    if phase_done "$phase"; then
        return
    fi
    if [ -z "$COMPLETED_PHASES" ]; then
        COMPLETED_PHASES="$phase"
    else
        COMPLETED_PHASES="${COMPLETED_PHASES},$phase"
    fi
}

next_phase_after() {
    local phase="$1"
    case "$phase" in
        3) echo 4 ;;
        4) echo 5 ;;
        5) echo 6 ;;
        *) echo "$phase" ;;
    esac
}

run_command() {
    local phase="$1"
    local cmd="$2"
    echo "[$(now_utc)] phase=${phase} run: $cmd" | tee -a "$LOG_FILE"
    if ! (cd "$PROJECT_ROOT" && bash -lc "$cmd" >>"$LOG_FILE" 2>&1); then
        LAST_ERROR="phase ${phase} failed command: ${cmd}"
        STATUS="blocked"
        save_state
        echo "BLOCKED: $LAST_ERROR"
        echo "See log: $LOG_FILE"
        return 1
    fi
    return 0
}

run_phase() {
    local phase="$1"
    STATUS="running"
    CURRENT_PHASE="$phase"
    save_state

    while IFS= read -r cmd; do
        [ -z "$cmd" ] && continue
        run_command "$phase" "$cmd"
    done < <(phase_commands "$phase")

    mark_phase_done "$phase"
    CURRENT_PHASE="$(next_phase_after "$phase")"
    STATUS="running"
    LAST_ERROR=
    save_state
    echo "DONE: phase $phase"
}

run_all() {
    ensure_state
    load_state
    : >"$LOG_FILE"
    echo "[$(now_utc)] phase-agent run start" >>"$LOG_FILE"

    for phase in "${PHASES[@]}"; do
        if phase_done "$phase"; then
            continue
        fi
        if [ "$phase" -lt "$CURRENT_PHASE" ]; then
            mark_phase_done "$phase"
            continue
        fi
        run_phase "$phase"
    done

    STATUS="complete"
    CURRENT_PHASE=6
    LAST_ERROR=
    save_state
    echo "COMPLETE: phases 3, 4, and 5 are complete."
}

print_status() {
    ensure_state
    load_state
    echo "Phase Agent Status"
    echo "  status: $STATUS"
    echo "  current_phase: $CURRENT_PHASE"
    echo "  completed_phases: ${COMPLETED_PHASES:-<none>}"
    echo "  updated_at: ${UPDATED_AT:-<unknown>}"
    if [ -n "${LAST_ERROR:-}" ]; then
        echo "  last_error: $LAST_ERROR"
    fi
    echo "  log_file: $LOG_FILE"
}

reset_state() {
    rm -rf "$STATE_DIR"
    ensure_state
    echo "Phase agent state reset."
}

usage() {
    cat <<'EOF'
Usage: scripts/phase-agent.sh <command>

Commands:
  run      Execute phases 3 -> 4 -> 5 in order
  status   Print current phase-agent state
  reset    Reset phase-agent state
EOF
}

main() {
    local cmd="${1:-run}"
    case "$cmd" in
        run) run_all ;;
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
