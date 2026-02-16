#!/usr/bin/env bash
# update-lineage-baseline.sh — Refresh committed lineage drift baseline.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

BASELINE_PATH="${1:-$PROJECT_ROOT/tests/lineage/baseline.json}"
STRICT_MODE="${LINEAGE_STRICT:-true}"

cd "$PROJECT_ROOT"
go run ./cmd/strict lineage-export --strict="$STRICT_MODE" --out "$BASELINE_PATH" .
echo "Updated lineage baseline: $BASELINE_PATH"
