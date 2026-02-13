#!/usr/bin/env bash
# export-lineage-artifact.sh — Export normalized lineage artifact from source tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

OUT_PATH="${1:-$PROJECT_ROOT/tests/lineage/current.json}"
STRICT_MODE="${LINEAGE_STRICT:-true}"
shift || true

cd "$PROJECT_ROOT"
go run ./cmd/stricture lineage-export --strict="$STRICT_MODE" --out "$OUT_PATH" . "$@"

echo "Wrote lineage artifact: $OUT_PATH"
