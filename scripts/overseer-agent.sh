#!/usr/bin/env bash
# overseer-agent.sh — Autonomous overseer that keeps asking "what's next" until done or timed out.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$PROJECT_ROOT/.overseer-agent"
STATE_FILE="$STATE_DIR/state.env"
RUN_LOG="$STATE_DIR/run.log"
PROMPT_LOG="$STATE_DIR/prompts.log"
NEXT_PROMPT_FILE="$STATE_DIR/next-prompt.txt"

MAX_HOURS="${OVERSEER_MAX_HOURS:-5}"
POLL_SECONDS="${OVERSEER_POLL_SECONDS:-60}"

now_utc() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

now_epoch() {
    date +%s
}

sanitize_value() {
    local value="$1"
    value="${value//$'\n'/ }"
    printf '%s' "$value"
}

ensure_state_dir() {
    mkdir -p "$STATE_DIR"
}

ensure_state() {
    ensure_state_dir
    if [ ! -f "$STATE_FILE" ]; then
        cat >"$STATE_FILE" <<EOF_STATE
STARTED_AT_EPOCH=0
DEADLINE_AT_EPOCH=0
STATUS=idle
ITERATION=0
LAST_ERROR=
LAST_PROMPT=
LAST_CHECK=
UPDATED_AT=$(now_utc)
EOF_STATE
    fi
}

load_state() {
    STARTED_AT_EPOCH=0
    DEADLINE_AT_EPOCH=0
    STATUS=idle
    ITERATION=0
    LAST_ERROR=
    LAST_PROMPT=
    LAST_CHECK=
    UPDATED_AT=

    if [ -f "$STATE_FILE" ]; then
        while IFS='=' read -r key value; do
            [ -z "$key" ] && continue
            case "$key" in
                STARTED_AT_EPOCH) STARTED_AT_EPOCH="$value" ;;
                DEADLINE_AT_EPOCH) DEADLINE_AT_EPOCH="$value" ;;
                STATUS) STATUS="$value" ;;
                ITERATION) ITERATION="$value" ;;
                LAST_ERROR) LAST_ERROR="$value" ;;
                LAST_PROMPT) LAST_PROMPT="$value" ;;
                LAST_CHECK) LAST_CHECK="$value" ;;
                UPDATED_AT) UPDATED_AT="$value" ;;
                *) ;;
            esac
        done <"$STATE_FILE"
    fi
}

save_state() {
    ensure_state_dir
    cat >"$STATE_FILE" <<EOF_STATE
STARTED_AT_EPOCH=$(sanitize_value "$STARTED_AT_EPOCH")
DEADLINE_AT_EPOCH=$(sanitize_value "$DEADLINE_AT_EPOCH")
STATUS=$(sanitize_value "$STATUS")
ITERATION=$(sanitize_value "$ITERATION")
LAST_ERROR=$(sanitize_value "$LAST_ERROR")
LAST_PROMPT=$(sanitize_value "$LAST_PROMPT")
LAST_CHECK=$(sanitize_value "$LAST_CHECK")
UPDATED_AT=$(now_utc)
EOF_STATE
}

log_line() {
    local message="$1"
    ensure_state_dir
    printf '[%s] %s\n' "$(now_utc)" "$message" | tee -a "$RUN_LOG"
}

emit_prompt() {
    local message="$1"
    ensure_state_dir
    LAST_PROMPT="$message"
    save_state

    printf '%s\n' "WHAT'S NEXT: $message" | tee -a "$PROMPT_LOG"
    printf '%s\n' "$message" >"$NEXT_PROMPT_FILE"
}

run_check() {
    local label="$1"
    local cmd="$2"

    LAST_CHECK="$label"
    save_state

    log_line "check=${label} cmd=${cmd}"
    if (cd "$PROJECT_ROOT" && eval "$cmd" >>"$RUN_LOG" 2>&1); then
        log_line "check=${label} result=pass"
        return 0
    fi

    log_line "check=${label} result=fail"
    LAST_ERROR="${label} failed: ${cmd}"
    save_state
    return 1
}

run_cycle() {
    LAST_ERROR=
    save_state

    if ! run_check "ci" "make ci"; then
        STATUS="blocked"
        emit_prompt "Fix failing code and tests so 'make ci' passes, then rerun overseer."
        save_state
        return 1
    fi

    if ! run_check "spec-quality-audit" "./scripts/spec-quality-audit.sh"; then
        STATUS="blocked"
        emit_prompt "Agents must reconcile spec vs tests vs code deficiencies from spec-quality-audit and raise quality until it passes."
        save_state
        return 1
    fi

    STATUS="complete"
    LAST_ERROR=
    emit_prompt "All checks pass. Launch independent agents to review spec/tests/code for hidden deficiencies and keep raising quality bar toward top 1%."
    save_state
    return 0
}

init_window_if_needed() {
    if [ "$STARTED_AT_EPOCH" -eq 0 ] || [ "$DEADLINE_AT_EPOCH" -eq 0 ]; then
        STARTED_AT_EPOCH="$(now_epoch)"
        DEADLINE_AT_EPOCH="$((STARTED_AT_EPOCH + (MAX_HOURS * 3600)))"
        STATUS="running"
        save_state
    fi
}

run_once() {
    ensure_state
    load_state
    init_window_if_needed

    ITERATION=$((ITERATION + 1))
    STATUS="running"
    save_state

    if run_cycle; then
        printf 'COMPLETE: overseer checks passed.\n'
        return 0
    fi

    printf 'BLOCKED: %s\n' "$LAST_ERROR"
    return 1
}

run_loop() {
    ensure_state
    load_state
    init_window_if_needed

    : >"$RUN_LOG"
    log_line "overseer run start (max_hours=${MAX_HOURS}, poll_seconds=${POLL_SECONDS})"

    while [ "$(now_epoch)" -lt "$DEADLINE_AT_EPOCH" ]; do
        ITERATION=$((ITERATION + 1))
        STATUS="running"
        save_state

        if run_cycle; then
            log_line "overseer result=complete iteration=${ITERATION}"
            return 0
        fi

        log_line "overseer result=blocked iteration=${ITERATION} error=${LAST_ERROR}"
        sleep "$POLL_SECONDS"
    done

    STATUS="timed_out"
    LAST_ERROR="Overseer time budget exceeded (${MAX_HOURS}h)."
    emit_prompt "5-hour overseer window expired. Triage unresolved deficiencies and continue with a fresh overseer run."
    save_state
    log_line "overseer result=timed_out"
    return 2
}

print_status() {
    ensure_state
    load_state

    echo "Overseer Agent Status"
    echo "  status: $STATUS"
    echo "  iteration: $ITERATION"
    echo "  started_at_epoch: $STARTED_AT_EPOCH"
    echo "  deadline_at_epoch: $DEADLINE_AT_EPOCH"
    echo "  last_check: ${LAST_CHECK:-<none>}"
    echo "  last_error: ${LAST_ERROR:-<none>}"
    echo "  last_prompt: ${LAST_PROMPT:-<none>}"
    echo "  updated_at: ${UPDATED_AT:-<unknown>}"
    echo "  run_log: $RUN_LOG"
    echo "  prompt_log: $PROMPT_LOG"
    echo "  next_prompt_file: $NEXT_PROMPT_FILE"
}

reset_state() {
    rm -rf "$STATE_DIR"
    ensure_state
    echo "Overseer agent state reset."
}

usage() {
    cat <<'EOF_USAGE'
Usage: scripts/overseer-agent.sh <command>

Commands:
  run      Loop until complete or timeout (default)
  once     Execute one cycle immediately
  status   Print current overseer state
  reset    Reset overseer state
  help     Show this message

Environment:
  OVERSEER_MAX_HOURS     Max runtime window in hours (default: 5)
  OVERSEER_POLL_SECONDS  Delay between blocked retries (default: 60)
  OVERSEER_MIN_COVERAGE  Coverage threshold for spec-quality-audit (default: 80)
EOF_USAGE
}

main() {
    local cmd="${1:-run}"
    case "$cmd" in
        run) run_loop ;;
        once) run_once ;;
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
