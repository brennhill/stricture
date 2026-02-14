#!/usr/bin/env bash
# check-tool-quality.sh — Reliability gate for the Stricture tool itself.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

GOCACHE_DIR="$PROJECT_ROOT/.cache/go-build"
CORE_PROFILE="$PROJECT_ROOT/.cache/coverage_tool_core.out"
CMD_PROFILE="$PROJECT_ROOT/.cache/coverage_tool_cmd.out"

STRICTURE_CORE_MIN_COVERAGE="${STRICTURE_CORE_MIN_COVERAGE:-86}"
STRICTURE_CMD_MIN_COVERAGE="${STRICTURE_CMD_MIN_COVERAGE:-20}"
STRICTURE_REPEAT_RUNS="${STRICTURE_REPEAT_RUNS:-3}"
STRICTURE_ENABLE_RACE="${STRICTURE_ENABLE_RACE:-1}"

CORE_COVERAGE_PKGS="${CORE_COVERAGE_PKGS:-./internal/adapter/java/... ./internal/adapter/python/... ./internal/adapter/typescript/... ./internal/config/... ./internal/fix/... ./internal/lineage/... ./internal/manifest/... ./internal/plugins/... ./internal/rules/arch/... ./internal/rules/conv/... ./internal/rules/ctr/... ./internal/rules/tq/... ./internal/suppression/...}"

mkdir -p "$GOCACHE_DIR" "$PROJECT_ROOT/.cache"
export GOCACHE="$GOCACHE_DIR"

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

pass() {
	echo "PASS: $*"
}

extract_coverage() {
	local profile="$1"
	local total
	total="$(go tool cover -func="$profile" | awk '/^total:/ {print $3}' | tr -d '%')"
	if [ -z "$total" ]; then
		fail "unable to parse coverage total from $profile"
	fi
	echo "$total"
}

assert_min_coverage() {
	local actual="$1"
	local min="$2"
	local label="$3"
	if awk -v actual="$actual" -v min="$min" 'BEGIN {exit !(actual + 0 >= min + 0)}'; then
		pass "$label coverage ${actual}% >= ${min}%"
		return
	fi
	fail "$label coverage ${actual}% < ${min}%"
}

echo "==> Core analyzer coverage gate"
go test -count=1 -coverprofile="$CORE_PROFILE" -covermode=atomic $CORE_COVERAGE_PKGS >/dev/null
core_cov="$(extract_coverage "$CORE_PROFILE")"
assert_min_coverage "$core_cov" "$STRICTURE_CORE_MIN_COVERAGE" "core analyzer"

echo "==> CLI unit coverage gate"
go test -count=1 -coverprofile="$CMD_PROFILE" -covermode=atomic ./cmd/stricture >/dev/null
cmd_cov="$(extract_coverage "$CMD_PROFILE")"
assert_min_coverage "$cmd_cov" "$STRICTURE_CMD_MIN_COVERAGE" "cmd/stricture"

echo "==> Integration CLI suite"
go test -tags=integration -count=1 -timeout=180s ./tests/integration/...
pass "integration CLI suite"

echo "==> Repeatability smoke (runs=${STRICTURE_REPEAT_RUNS})"
for run in $(seq 1 "$STRICTURE_REPEAT_RUNS"); do
	echo "  -> repeat run $run/$STRICTURE_REPEAT_RUNS"
	go test -tags=integration -count=1 -run 'TestVersionOutputDeterminism|TestGoldenTextOutput|TestGoldenJSONOutput|TestRuleFlagCanBeRepeated|TestSelfLint' ./tests/integration/...
done
pass "repeatability smoke"

if [ "$STRICTURE_ENABLE_RACE" = "1" ]; then
	echo "==> Race detector"
	go test -race -count=1 -timeout=300s ./cmd/... ./internal/...
	pass "race detector"
fi

echo "PASS: tool quality gate passed."
