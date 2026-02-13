#!/usr/bin/env bash
# tdd-rule.sh — Enforce rule-level TDD (red/green) with gate validation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
    cat <<'EOF'
Usage: tdd-rule.sh --rule RULE_ID --stage red|green

Examples:
  ./scripts/tdd-rule.sh --rule CONV-error-format --stage red
  ./scripts/tdd-rule.sh --rule CONV-error-format --stage green
EOF
}

RULE_ID=""
STAGE=""

while [ $# -gt 0 ]; do
    case "$1" in
        --rule)
            if [ $# -lt 2 ]; then
                echo "Error: --rule requires a value." >&2
                exit 1
            fi
            RULE_ID="$2"
            shift 2
            ;;
        --stage)
            if [ $# -lt 2 ]; then
                echo "Error: --stage requires a value (red|green)." >&2
                exit 1
            fi
            STAGE="$2"
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

if [ -z "$RULE_ID" ] || [ -z "$STAGE" ]; then
    usage
    exit 1
fi

if ! [[ "$RULE_ID" =~ ^(TQ|ARCH|CONV|CTR)-[a-z][a-z0-9-]*$ ]]; then
    echo "Error: invalid rule ID: $RULE_ID" >&2
    exit 1
fi

if [ "$STAGE" != "red" ] && [ "$STAGE" != "green" ]; then
    echo "Error: --stage must be red or green." >&2
    exit 1
fi

to_pascal() {
    local input="$1"
    local output=""
    local IFS='-'
    read -r -a parts <<< "$input"
    for part in "${parts[@]}"; do
        if [ -n "$part" ]; then
            output+=$(tr '[:lower:]' '[:upper:]' <<< "${part:0:1}")"${part:1}"
        fi
    done
    echo "$output"
}

CATEGORY=$(echo "$RULE_ID" | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]')
RULE_NAME=$(echo "$RULE_ID" | sed -E 's/^[A-Z]+-//')
TEST_PREFIX="Test$(to_pascal "$RULE_NAME")"
PKG="./internal/rules/${CATEGORY}/..."

cd "$PROJECT_ROOT"

echo "==> Gate check (spec/catalog/test-plan/fixtures): $RULE_ID"
"$PROJECT_ROOT/scripts/validate-gate.sh" "$RULE_ID"

echo ""
echo "==> Running targeted tests: $PKG -run ^${TEST_PREFIX}"

TMP_OUTPUT="$(mktemp)"
cleanup() {
    rm -f "$TMP_OUTPUT"
}
trap cleanup EXIT

if go test "$PKG" -run "^${TEST_PREFIX}" -count=1 -timeout=90s >"$TMP_OUTPUT" 2>&1; then
    if grep -q "no tests to run" "$TMP_OUTPUT"; then
        cat "$TMP_OUTPUT"
        echo ""
        echo "FAIL: No tests matched ^${TEST_PREFIX}. Add tests before proceeding." >&2
        exit 1
    fi

    if [ "$STAGE" = "red" ]; then
        cat "$TMP_OUTPUT"
        echo ""
        echo "FAIL: Red stage expected failing tests, but tests passed." >&2
        exit 1
    fi

    cat "$TMP_OUTPUT"
    echo ""
    echo "PASS: Green stage confirmed. Tests pass for $RULE_ID."
    exit 0
fi

cat "$TMP_OUTPUT"
echo ""

if grep -q "no tests to run" "$TMP_OUTPUT"; then
    echo "FAIL: No tests matched ^${TEST_PREFIX}. Add tests before proceeding." >&2
    exit 1
fi

if grep -q "setup failed" "$TMP_OUTPUT" || grep -q "\\[build failed\\]" "$TMP_OUTPUT" || grep -q "operation not permitted" "$TMP_OUTPUT"; then
    echo "FAIL: Test execution failed before assertions (setup/build/environment issue)." >&2
    exit 1
fi

if [ "$STAGE" = "red" ] && ! grep -Eq '^--- FAIL: ' "$TMP_OUTPUT"; then
    echo "FAIL: Red stage requires an actual failing test assertion." >&2
    exit 1
fi

if [ "$STAGE" = "green" ]; then
    echo "FAIL: Green stage expected passing tests, but tests failed." >&2
    exit 1
fi

echo "PASS: Red stage confirmed. Tests fail for $RULE_ID as expected."
