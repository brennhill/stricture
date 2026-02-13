#!/usr/bin/env bash
# check-tree-sitter-pinning.sh — Ensures grammar pinning doc matches go.mod.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOC_FILE="$PROJECT_ROOT/docs/TREE-SITTER-PINNING.md"
GO_MOD_FILE="$PROJECT_ROOT/go.mod"

if [ ! -f "$DOC_FILE" ]; then
    echo "FAIL: missing docs file: $DOC_FILE"
    exit 1
fi

if [ ! -f "$GO_MOD_FILE" ]; then
    echo "FAIL: missing go.mod: $GO_MOD_FILE"
    exit 1
fi

PINNED_VERSION="$(awk '/github.com\/smacker\/go-tree-sitter / {print $2; exit}' "$GO_MOD_FILE")"
if [ -z "$PINNED_VERSION" ]; then
    echo "FAIL: could not find github.com/smacker/go-tree-sitter version in go.mod"
    exit 1
fi

failed=0

if rg -q 'TBD' "$DOC_FILE"; then
    echo "FAIL: docs/TREE-SITTER-PINNING.md still contains 'TBD'"
    failed=1
fi

for language in TypeScript Python Java; do
    row="$(awk -F'|' -v lang="$language" '$0 ~ ("\\|[[:space:]]*" lang "[[:space:]]*\\|") {print; exit}' "$DOC_FILE")"
    if [ -z "$row" ]; then
        echo "FAIL: missing table row for $language in docs/TREE-SITTER-PINNING.md"
        failed=1
        continue
    fi

    version_field="$(printf '%s\n' "$row" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $4); print $4}')"
    verified_field="$(printf '%s\n' "$row" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $5); print $5}')"
    normalized_version="${version_field//\`/}"

    if [ -z "$normalized_version" ] || [ "$normalized_version" = "—" ] || [ "$normalized_version" = "-" ]; then
        echo "FAIL: $language has an empty version in docs/TREE-SITTER-PINNING.md"
        failed=1
    elif [ "$normalized_version" != "$PINNED_VERSION" ]; then
        echo "FAIL: $language version mismatch (doc=$normalized_version, go.mod=$PINNED_VERSION)"
        failed=1
    fi

    if ! [[ "$verified_field" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
        echo "FAIL: $language Last verified must be YYYY-MM-DD (got: $verified_field)"
        failed=1
    fi
done

if [ "$failed" -ne 0 ]; then
    exit 1
fi

echo "PASS: tree-sitter pinning docs are complete and match go.mod ($PINNED_VERSION)."
