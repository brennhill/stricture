#!/usr/bin/env bash
# check-bash-syntax.sh — Validate Bash script syntax in scripts/ directory.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

failed=0

while IFS= read -r script; do
    if ! bash -n "$script"; then
        echo "FAIL: bash syntax error in $script"
        failed=1
    fi
done < <(find "$PROJECT_ROOT/scripts" -maxdepth 1 -type f -name "*.sh" | sort)

if [ "$failed" -ne 0 ]; then
    exit 1
fi

echo "PASS: bash syntax valid for all scripts in scripts/."
