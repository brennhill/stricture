#!/usr/bin/env bash
# check-no-stubs.sh — Blocks placeholder implementations in active phase scope.
set -euo pipefail

phase="1"

usage() {
    echo "Usage: $(basename "$0") [--phase N]"
    echo ""
    echo "Checks active implementation paths for placeholder markers:"
    echo "  - not yet implemented"
    echo "  - TODO: implement"
    echo ""
    echo "Options:"
    echo "  --phase N   Development phase scope (1-4). Default: 1"
}

while [ $# -gt 0 ]; do
    case "$1" in
        --phase)
            if [ $# -lt 2 ]; then
                echo "Error: --phase requires a value" >&2
                exit 1
            fi
            phase="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Error: unknown argument: $1" >&2
            usage
            exit 1
            ;;
    esac
done

if ! [[ "$phase" =~ ^[1-4]$ ]]; then
    echo "Error: phase must be 1, 2, 3, or 4" >&2
    exit 1
fi

paths=(
    "cmd/stricture"
    "internal/config"
    "internal/adapter/goparser"
    "internal/rules/conv"
    "internal/reporter"
)

if [ "$phase" -ge 2 ]; then
    paths+=("internal/adapter/typescript" "internal/rules/arch" "internal/engine")
fi
if [ "$phase" -ge 3 ]; then
    paths+=("internal/rules/tq")
fi
if [ "$phase" -ge 4 ]; then
    paths+=("internal/manifest" "internal/rules/ctr")
fi

existing_paths=()
for p in "${paths[@]}"; do
    if [ -d "$p" ] || [ -f "$p" ]; then
        existing_paths+=("$p")
    fi
done

if [ "${#existing_paths[@]}" -eq 0 ]; then
    echo "PASS: no scoped paths exist for phase $phase"
    exit 0
fi

matches=$(rg -n --color=never \
    -e 'not yet implemented' \
    -e 'TODO: implement' \
    -e 'TODO: Implement' \
    --glob '*.go' \
    "${existing_paths[@]}" || true)

if [ -n "$matches" ]; then
    echo "FAIL: placeholder implementations found in phase $phase scope:"
    echo "$matches" | sed 's/^/  /'
    exit 1
fi

echo "PASS: no placeholder stubs found in phase $phase scope."
