#!/usr/bin/env bash
# check-benchmark-regression.sh — Compares benchmark results against baseline thresholds.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BASELINE_FILE="$PROJECT_ROOT/tests/benchmark/baseline.json"
BENCH_PKGS=(./cmd/... ./internal/...)

if ! command -v jq >/dev/null 2>&1; then
    echo "FAIL: jq is required for benchmark regression checks." >&2
    exit 1
fi

if [ ! -f "$BASELINE_FILE" ]; then
    echo "FAIL: benchmark baseline missing at tests/benchmark/baseline.json" >&2
    echo "Run: ./scripts/update-benchmark-baseline.sh" >&2
    exit 1
fi

threshold=$(jq -r '.threshold_percent // empty' "$BASELINE_FILE")
if [ -z "$threshold" ] || [ "$threshold" = "null" ]; then
    echo "FAIL: threshold_percent missing in $BASELINE_FILE" >&2
    exit 1
fi
baseline_count=$(jq -r '.benchmarks | length' "$BASELINE_FILE")

tmp_raw="$(mktemp)"
tmp_parsed="$(mktemp)"
cleanup() {
    rm -f "$tmp_raw" "$tmp_parsed"
}
trap cleanup EXIT

go test -run='^$' -bench=. -benchmem -count=1 "${BENCH_PKGS[@]}" > "$tmp_raw"

awk '
/^Benchmark/ {
    for (i = 1; i <= NF; i++) {
        if ($i == "ns/op") {
            print $1, $(i-1)
        }
    }
}
' "$tmp_raw" > "$tmp_parsed"

if [ ! -s "$tmp_parsed" ]; then
    if [ "$baseline_count" = "0" ]; then
        echo "SKIP: no benchmark results found and baseline is empty." >&2
        echo "Add Benchmark* tests, then run ./scripts/update-benchmark-baseline.sh." >&2
        exit 0
    fi
    echo "FAIL: no benchmark results found, but baseline has $baseline_count entries." >&2
    exit 1
fi

failures=0
while IFS=' ' read -r bench ns; do
    [ -n "$bench" ] || continue
    baseline_ns=$(jq -r --arg b "$bench" '.benchmarks[$b] // empty' "$BASELINE_FILE")
    if [ -z "$baseline_ns" ] || [ "$baseline_ns" = "null" ]; then
        echo "FAIL: baseline missing for $bench"
        failures=$((failures + 1))
        continue
    fi

    if [ "$baseline_ns" = "0" ]; then
        echo "FAIL: invalid baseline value 0 for $bench"
        failures=$((failures + 1))
        continue
    fi

    delta=$(echo "scale=4; (($ns - $baseline_ns) / $baseline_ns) * 100" | bc -l)
    abs_delta=$(echo "$delta" | awk '{ if ($1 < 0) print -$1; else print $1 }')

    if echo "$abs_delta > $threshold" | bc -l | grep -q 1; then
        echo "FAIL: $bench regressed by ${delta}% (baseline=${baseline_ns}ns/op current=${ns}ns/op threshold=${threshold}%)"
        failures=$((failures + 1))
    else
        echo "PASS: $bench delta=${delta}% (baseline=${baseline_ns}ns/op current=${ns}ns/op)"
    fi
done < "$tmp_parsed"

if [ "$failures" -gt 0 ]; then
    exit 1
fi

echo "PASS: benchmark regression check passed."
