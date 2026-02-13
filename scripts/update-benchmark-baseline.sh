#!/usr/bin/env bash
# update-benchmark-baseline.sh — Regenerates benchmark baseline from current results.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BASELINE_FILE="$PROJECT_ROOT/tests/benchmark/baseline.json"
BENCH_PKGS=(./cmd/... ./internal/...)

if ! command -v jq >/dev/null 2>&1; then
    echo "FAIL: jq is required to update benchmark baseline." >&2
    exit 1
fi

mkdir -p "$(dirname "$BASELINE_FILE")"

tmp_raw="$(mktemp)"
tmp_json="$(mktemp)"
cleanup() {
    rm -f "$tmp_raw" "$tmp_json"
}
trap cleanup EXIT

go test -run='^$' -bench=. -benchmem -count=1 "${BENCH_PKGS[@]}" > "$tmp_raw"

if ! awk '
/^Benchmark/ {
    for (i = 1; i <= NF; i++) {
        if ($i == "ns/op") {
            print $1 "=" $(i-1)
        }
    }
}
' "$tmp_raw" | awk -F= 'NF==2 {print}' > "$tmp_json"; then
    echo "FAIL: unable to parse benchmark output." >&2
    exit 1
fi

if [ ! -s "$tmp_json" ]; then
    echo "FAIL: no benchmark lines found. Add Benchmark* tests first." >&2
    exit 1
fi

jq -Rn '
  (input | split("=") | {k: .[0], v: (.[1] | tonumber)}) as $first
  | reduce ([$first] + [inputs | split("=") | {k: .[0], v: (.[1] | tonumber)}])[] as $item
      ({threshold_percent: 20, benchmarks: {}}; .benchmarks[$item.k] = $item.v)
' < "$tmp_json" > "$BASELINE_FILE"

echo "Updated benchmark baseline: $BASELINE_FILE"
