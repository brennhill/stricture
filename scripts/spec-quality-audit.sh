#!/usr/bin/env bash
# spec-quality-audit.sh — Deep quality audit across spec, tests, and implementation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$PROJECT_ROOT/.overseer-agent"
LOG_FILE="$STATE_DIR/spec-quality-audit.log"
COVERAGE_FILE="$STATE_DIR/coverage.out"
MIN_COVERAGE="${OVERSEER_MIN_COVERAGE:-80}"
COVERAGE_PKGS="${OVERSEER_COVERAGE_PKGS:-./internal/adapter/java/... ./internal/adapter/python/... ./internal/adapter/typescript/... ./internal/config/... ./internal/lineage/... ./internal/manifest/... ./internal/rules/arch/... ./internal/rules/conv/... ./internal/rules/ctr/... ./internal/rules/tq/...}"

mkdir -p "$STATE_DIR"
: >"$LOG_FILE"

FAIL_COUNT=0

now_utc() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

log() {
    local message="$1"
    printf '[%s] %s\n' "$(now_utc)" "$message" | tee -a "$LOG_FILE"
}

pass() {
    log "PASS: $1"
}

fail() {
    log "FAIL: $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

run_check() {
    local label="$1"
    local cmd="$2"

    log "CHECK: ${label} -> ${cmd}"
    if (cd "$PROJECT_ROOT" && eval "$cmd" >>"$LOG_FILE" 2>&1); then
        pass "$label"
        return 0
    fi

    fail "$label"
    return 1
}

check_no_marker_rules() {
    local tmp
    tmp="$(mktemp)"

    if (cd "$PROJECT_ROOT" && rg -n --glob '*.go' 'stricture:fail|hasRuleMarker\(' internal/rules/arch internal/rules/tq internal/rules/ctr >"$tmp" 2>/dev/null); then
        fail "semantic-rule-quality (marker-driven rule logic still present)"
        sed 's/^/  /' "$tmp" | tee -a "$LOG_FILE"
    else
        pass "semantic-rule-quality"
    fi

    rm -f "$tmp"
}

check_no_skipped_integration_tests() {
    local tmp
    tmp="$(mktemp)"

    if (cd "$PROJECT_ROOT" && rg -n 't\.Skip\(' tests/integration >"$tmp" 2>/dev/null); then
        fail "integration-rigor (skipped integration tests still present)"
        sed 's/^/  /' "$tmp" | tee -a "$LOG_FILE"
    else
        pass "integration-rigor"
    fi

    rm -f "$tmp"
}

check_fixture_depth() {
    local missing=0
    local dir

    while IFS= read -r dir; do
        [ -z "$dir" ] && continue
        local real_count
        real_count="$(find "$dir" -type f ! -name '.gitkeep' | wc -l | tr -d ' ')"
        if [ "$real_count" -eq 0 ]; then
            fail "fixture-depth (${dir} has only placeholders)"
            missing=$((missing + 1))
        fi
    done < <(cd "$PROJECT_ROOT" && find tests/fixtures -maxdepth 1 -type d \( -name 'arch-*' -o -name 'tq-*' -o -name 'ctr-*' \) | sort)

    if [ "$missing" -eq 0 ]; then
        pass "fixture-depth"
    fi
}

check_coverage_threshold() {
    local coverage_raw

    run_check "coverage-collection" "go test -coverprofile '$COVERAGE_FILE' $COVERAGE_PKGS" || return

    coverage_raw="$(cd "$PROJECT_ROOT" && go tool cover -func "$COVERAGE_FILE" | awk '/^total:/ {print $3}' | tr -d '%')"
    if [ -z "$coverage_raw" ]; then
        fail "coverage-threshold (unable to parse total coverage)"
        return
    fi

    if awk -v actual="$coverage_raw" -v min="$MIN_COVERAGE" 'BEGIN {exit !(actual + 0 >= min + 0)}'; then
        pass "coverage-threshold (${coverage_raw}% >= ${MIN_COVERAGE}%)"
    else
        fail "coverage-threshold (${coverage_raw}% < ${MIN_COVERAGE}%)"
    fi
}

main() {
    log "spec-quality-audit start"

    run_check "phase5-tests" "make test-phase5" || true
    run_check "integration-tests" "go test -tags=integration -timeout=120s ./tests/integration/..." || true
    run_check "full-gates" "./scripts/validate-gate.sh --all" || true
    run_check "message-catalog-consistency" "./scripts/validate-error-messages.sh" || true
    run_check "rule-consistency" "./scripts/check-rule-consistency.sh" || true
    run_check "invariant-tests" "./scripts/check-invariant-tests.sh" || true
    run_check "validation-health" "./scripts/validation-health-check.sh" || true
    run_check "race-tests" "go test -race -count=1 -timeout=300s ./cmd/... ./internal/..." || true

    check_coverage_threshold
    check_no_marker_rules
    check_no_skipped_integration_tests
    check_fixture_depth

    if [ "$FAIL_COUNT" -gt 0 ]; then
        log "spec-quality-audit result: FAIL (${FAIL_COUNT} deficiencies)"
        exit 1
    fi

    log "spec-quality-audit result: PASS"
}

main "$@"
