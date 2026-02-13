#!/usr/bin/env bash
# check-invariant-tests.sh — Verifies invariant test files referenced in docs exist.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INVARIANTS_DOC="$PROJECT_ROOT/docs/INVARIANTS.md"

if [ ! -f "$INVARIANTS_DOC" ]; then
    echo "FAIL: docs/INVARIANTS.md not found" >&2
    exit 1
fi

TMP_FILE="$(mktemp)"
cleanup() {
    rm -f "$TMP_FILE"
}
trap cleanup EXIT

awk -F'`' '/\.go`/ {
    for (i = 2; i <= NF; i += 2) {
        if ($i ~ /\.go$/ || $i ~ /\.go\*$/ || $i ~ /\*\/fuzz_test\.go$/) {
            print $i
        }
    }
}' "$INVARIANTS_DOC" | sort -u > "$TMP_FILE"

if [ ! -s "$TMP_FILE" ]; then
    echo "FAIL: no invariant test paths found in docs/INVARIANTS.md" >&2
    exit 1
fi

missing=0
while IFS= read -r rel; do
    [ -n "$rel" ] || continue

    # Handle wildcard paths in the invariant table.
    if [[ "$rel" == *"*"* || "$rel" == *"?"* || "$rel" == *"["* ]]; then
        if ! compgen -G "$PROJECT_ROOT/$rel" >/dev/null; then
            echo "MISSING: $rel"
            missing=1
        fi
        continue
    fi

    if [ ! -f "$PROJECT_ROOT/$rel" ]; then
        echo "MISSING: $rel"
        missing=1
    fi
done < "$TMP_FILE"

if [ "$missing" -ne 0 ]; then
    echo "FAIL: invariant test file coverage is incomplete." >&2
    exit 1
fi

echo "PASS: all invariant test paths in docs/INVARIANTS.md exist."
