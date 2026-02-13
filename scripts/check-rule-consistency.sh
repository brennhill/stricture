#!/usr/bin/env bash
# check-rule-consistency.sh — Ensures rule IDs stay consistent across docs and scripts.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

CATALOG="$PROJECT_ROOT/docs/error-catalog.yml"
RULE_INDEX="$PROJECT_ROOT/docs/RULE-INDEX.md"
GATE_SCRIPT="$PROJECT_ROOT/scripts/validate-gate.sh"
HEALTH_SCRIPT="$PROJECT_ROOT/scripts/validation-health-check.sh"
VALIDATION_DIR="$PROJECT_ROOT/docs/test-plan/validation-set"

TMP_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
    echo "FAIL: $1" >&2
    exit 1
}

extract_catalog_rules() {
    rg -o --no-filename '^[[:space:]]{2}(TQ|ARCH|CONV|CTR)-[a-z-]+:' "$CATALOG" |
        sed -E 's/^[[:space:]]*//; s/:$//' |
        sort -u
}

extract_rule_index_rules() {
    rg -o --no-filename '^\| (TQ|ARCH|CONV|CTR)-[a-z-]+ \|' "$RULE_INDEX" |
        sed -E 's/^\| //; s/ \|$//' |
        sort -u
}

extract_gate_rules() {
    rg -o --no-filename '"(TQ|ARCH|CONV|CTR)-[a-z-]+"' "$GATE_SCRIPT" |
        tr -d '"' |
        sort -u
}

extract_health_rules() {
    rg -o --no-filename '"(TQ|ARCH|CONV|CTR)-[a-z-]+"' "$HEALTH_SCRIPT" |
        tr -d '"' |
        sort -u
}

extract_validation_refs() {
    rg -o --no-filename '(TQ|ARCH|CONV|CTR)-[a-z-]+' "$VALIDATION_DIR" |
        sort -u
}

compare_exact_set() {
    local label="$1"
    local expected="$2"
    local actual="$3"

    local missing
    local extra
    missing=$(comm -23 "$expected" "$actual" || true)
    extra=$(comm -13 "$expected" "$actual" || true)

    if [ -n "$missing" ] || [ -n "$extra" ]; then
        echo ""
        echo "FAIL: ${label} diverges from error-catalog.yml"
        if [ -n "$missing" ]; then
            echo "  Missing in ${label}:"
            echo "$missing" | sed 's/^/    - /'
        fi
        if [ -n "$extra" ]; then
            echo "  Extra in ${label}:"
            echo "$extra" | sed 's/^/    - /'
        fi
        exit 1
    fi
}

extract_catalog_rules > "$TMP_DIR/catalog.txt"
extract_rule_index_rules > "$TMP_DIR/rule_index.txt"
extract_gate_rules > "$TMP_DIR/gate.txt"
extract_health_rules > "$TMP_DIR/health.txt"
extract_validation_refs > "$TMP_DIR/validation_refs.txt"

if [ ! -s "$TMP_DIR/catalog.txt" ]; then
    fail "No rules found in docs/error-catalog.yml"
fi

compare_exact_set "docs/RULE-INDEX.md" "$TMP_DIR/catalog.txt" "$TMP_DIR/rule_index.txt"
compare_exact_set "scripts/validate-gate.sh" "$TMP_DIR/catalog.txt" "$TMP_DIR/gate.txt"
compare_exact_set "scripts/validation-health-check.sh" "$TMP_DIR/catalog.txt" "$TMP_DIR/health.txt"

# Validation docs may mention each rule multiple times, but they must reference every known
# rule at least once and may not reference unknown rule IDs.
validation_missing=$(comm -23 "$TMP_DIR/catalog.txt" "$TMP_DIR/validation_refs.txt" || true)
validation_unknown=$(comm -13 "$TMP_DIR/catalog.txt" "$TMP_DIR/validation_refs.txt" || true)

if [ -n "$validation_missing" ] || [ -n "$validation_unknown" ]; then
    echo ""
    echo "FAIL: validation-set docs rule references are inconsistent"
    if [ -n "$validation_missing" ]; then
        echo "  Missing from validation docs:"
        echo "$validation_missing" | sed 's/^/    - /'
    fi
    if [ -n "$validation_unknown" ]; then
        echo "  Unknown in validation docs:"
        echo "$validation_unknown" | sed 's/^/    - /'
    fi
    exit 1
fi

echo "PASS: Rule IDs are consistent across catalog, index, gate scripts, and validation docs."
echo "  Total rules: $(wc -l < "$TMP_DIR/catalog.txt" | tr -d ' ')"
