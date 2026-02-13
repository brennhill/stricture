#!/usr/bin/env bash
# check-lineage-drift.sh — Compare current lineage artifact against baseline.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

BASELINE_PATH="${LINEAGE_BASELINE:-$PROJECT_ROOT/tests/lineage/baseline.json}"
HEAD_PATH="${LINEAGE_HEAD:-/tmp/stricture-lineage-head.json}"
DIFF_PATH="${LINEAGE_DIFF:-/tmp/stricture-lineage-diff.json}"
FAIL_ON="${LINEAGE_FAIL_ON:-medium}"
STRICT_MODE="${LINEAGE_STRICT:-true}"
MODE="${LINEAGE_MODE:-block}"

if [ ! -f "$BASELINE_PATH" ]; then
    echo "FAIL: Missing lineage baseline at $BASELINE_PATH" >&2
    echo "Run: ./scripts/update-lineage-baseline.sh" >&2
    exit 1
fi

cd "$PROJECT_ROOT"
go run ./cmd/stricture lineage-export --strict="$STRICT_MODE" --out "$HEAD_PATH" .
go run ./cmd/stricture lineage-diff --base "$BASELINE_PATH" --head "$HEAD_PATH" --out "$DIFF_PATH" --fail-on "$FAIL_ON" --mode "$MODE"

echo "Lineage diff result:"
cat "$DIFF_PATH"
