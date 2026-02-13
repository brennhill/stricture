#!/usr/bin/env bash
# validate-error-messages.sh — Validate rule implementations match error-catalog.yml.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CATALOG="$PROJECT_ROOT/docs/error-catalog.yml"
RULES_DIR="$PROJECT_ROOT/internal/rules"

# ── Colors ────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
FULLY_VALIDATED=0
TOTAL_RULES=0

pass() { echo -e "  ${GREEN}\xE2\x9C\x85${NC} $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { echo -e "  ${RED}\xE2\x9D\x8C${NC} $1"; FAIL_COUNT=$((FAIL_COUNT + 1)); }
warn() { echo -e "  ${YELLOW}⚠${NC}  $1"; WARN_COUNT=$((WARN_COUNT + 1)); }

# ── Validate prerequisites ────────────────────────────────
if [ ! -f "$CATALOG" ]; then
    echo -e "${RED}Error: error-catalog.yml not found at $CATALOG${NC}" >&2
    exit 1
fi

if [ ! -d "$RULES_DIR" ]; then
    echo -e "${RED}Error: rules directory not found at $RULES_DIR${NC}" >&2
    exit 1
fi

# ── Parse catalog ─────────────────────────────────────────
# Extract rule entries from the YAML catalog.
# We parse line-by-line, tracking the current rule ID and its fields.
# This avoids requiring yq or python as a dependency.

declare -A CATALOG_CATEGORY
declare -A CATALOG_SEVERITY
declare -A CATALOG_MESSAGE

current_rule=""
in_rules_block=false

while IFS= read -r line; do
    # Detect start of rules: block
    if [[ "$line" =~ ^rules: ]]; then
        in_rules_block=true
        continue
    fi

    $in_rules_block || continue

    # Detect a new rule ID (indented exactly 2 spaces, ends with colon, has a category prefix)
    if [[ "$line" =~ ^[[:space:]]{2}(TQ|ARCH|CONV|CTR)-([a-z-]+):$ ]]; then
        current_rule="${BASH_REMATCH[1]}-${BASH_REMATCH[2]}"
        continue
    fi

    [ -n "$current_rule" ] || continue

    # Extract category
    if [[ "$line" =~ ^[[:space:]]+category:[[:space:]]+(.+)$ ]]; then
        CATALOG_CATEGORY["$current_rule"]="${BASH_REMATCH[1]}"
        continue
    fi

    # Extract severity
    if [[ "$line" =~ ^[[:space:]]+severity:[[:space:]]+(.+)$ ]]; then
        CATALOG_SEVERITY["$current_rule"]="${BASH_REMATCH[1]}"
        continue
    fi

    # Extract message template (strip surrounding quotes)
    if [[ "$line" =~ ^[[:space:]]+message:[[:space:]]+\"(.+)\"$ ]]; then
        CATALOG_MESSAGE["$current_rule"]="${BASH_REMATCH[1]}"
        continue
    fi
done < "$CATALOG"

# ── Derive key phrases from a message template ───────────
# Extracts static text fragments (words outside {placeholders}) that must
# appear in the Go implementation's fmt.Sprintf format string.
# Returns a newline-separated list of key phrases (2+ words each).
extract_key_phrases() {
    local template="$1"
    # Remove placeholders like {field_name}, leaving static text
    local static
    static=$(echo "$template" | sed -E 's/\{[a-zA-Z_]+\}//g')
    # Split on runs of whitespace/punctuation to get meaningful fragments
    # We look for sequences of 2+ consecutive words
    local phrases=""
    local prev_word=""
    for word in $static; do
        # Strip leading/trailing punctuation
        word=$(echo "$word" | sed "s/^[^a-zA-Z0-9]*//;s/[^a-zA-Z0-9]*$//")
        [ -z "$word" ] && continue
        if [ -n "$prev_word" ]; then
            phrases="${phrases}${prev_word} ${word}"$'\n'
        fi
        prev_word="$word"
    done
    echo "$phrases"
}

# ── Map rule ID to implementation file ────────────────────
# CONV-file-naming -> internal/rules/conv/file_naming.go
find_rule_file() {
    local rule_id="$1"
    local prefix="${rule_id%%-*}"                          # CONV
    local suffix="${rule_id#*-}"                           # file-naming
    local category_dir
    category_dir=$(echo "$prefix" | tr '[:upper:]' '[:lower:]')  # conv
    local filename
    filename=$(echo "$suffix" | tr '-' '_')                # file_naming
    local filepath="$RULES_DIR/$category_dir/${filename}.go"

    if [ -f "$filepath" ]; then
        echo "$filepath"
        return 0
    fi
    return 1
}

# ── Extract Go method return value ────────────────────────
# Looks for func (r *Type) MethodName() string { return "value" }
# and extracts the "value" portion.
# Works with both BSD grep (macOS) and GNU grep (Linux).
extract_method_value() {
    local file="$1"
    local method="$2"
    local value=""

    # Try single-line form first: func (r *X) Method() string { return "val" }
    # Then try multi-line form: method header on one line, return on next
    value=$(grep -A2 "func.*${method}()" "$file" 2>/dev/null | \
        grep 'return' | \
        sed -E 's/.*return[[:space:]]+"([^"]+)".*/\1/' | head -1)

    # Verify we actually extracted a value.
    # When sed fails, it may return the original source line containing `return "..."`.
    if echo "$value" | grep -Eq 'return[[:space:]]*"'; then
        value=""
    fi

    echo "$value"
}

# ── Check if Check() returns nil (stub) ──────────────────
is_stub_rule() {
    local file="$1"
    # Look for a Check method that simply returns nil with no real logic
    # Pattern: the Check func body contains only "return nil" (possibly with a TODO comment)
    local check_body
    check_body=$(sed -n '/^func.*Check(/,/^func\|^}/p' "$file" 2>/dev/null)
    if echo "$check_body" | grep -q 'return nil' && \
       ! echo "$check_body" | grep -q 'Violation{'; then
        return 0
    fi
    return 1
}

# ── Check message template key phrases in Check() body ────
check_message_phrases() {
    local file="$1"
    local template="$2"
    local missing_phrases=""
    local checked=0
    local matched=0

    # Extract the key phrases from the catalog template
    local phrases
    phrases=$(extract_key_phrases "$template")

    [ -z "$phrases" ] && return 0

    # Read the Check function body for message strings
    local check_body
    check_body=$(sed -n '/^func.*Check(/,/^func [^(]\|^}$/p' "$file" 2>/dev/null)

    while IFS= read -r phrase; do
        [ -z "$phrase" ] && continue
        checked=$((checked + 1))
        # Case-insensitive search for the phrase in the Check body
        if echo "$check_body" | grep -qi "$phrase"; then
            matched=$((matched + 1))
        else
            if [ -n "$missing_phrases" ]; then
                missing_phrases="$missing_phrases, "
            fi
            missing_phrases="${missing_phrases}'${phrase}'"
        fi
    done <<< "$phrases"

    if [ "$checked" -eq 0 ]; then
        return 0
    fi

    # Consider it a match if at least half of key phrases are found
    local threshold=$(( (checked + 1) / 2 ))
    if [ "$matched" -ge "$threshold" ]; then
        return 0
    fi

    echo "$missing_phrases"
    return 1
}

# ── Check required violation fields ───────────────────────
check_violation_fields() {
    local file="$1"
    local check_body
    check_body=$(sed -n '/^func.*Check(/,/^func [^(]\|^}$/p' "$file" 2>/dev/null)
    local missing=""

    for field in RuleID Severity Message FilePath StartLine; do
        if ! echo "$check_body" | grep -q "${field}:"; then
            if [ -n "$missing" ]; then
                missing="$missing, "
            fi
            missing="${missing}${field}"
        fi
    done

    if [ -n "$missing" ]; then
        echo "$missing"
        return 1
    fi
    return 0
}

# ── Expected category from rule ID prefix ─────────────────
expected_category_from_id() {
    local rule_id="$1"
    local prefix="${rule_id%%-*}"
    echo "$prefix" | tr '[:upper:]' '[:lower:]'
}

# ── Main ──────────────────────────────────────────────────
echo ""
echo -e "${BOLD}Error Message Validation${NC}"
echo -e "${BOLD}=========================${NC}"
echo ""

# Build list of rule IDs to check
if [ $# -gt 0 ]; then
    RULE_IDS=("$@")
else
    # All rule IDs from the catalog
    RULE_IDS=()
    for key in "${!CATALOG_CATEGORY[@]}"; do
        RULE_IDS+=("$key")
    done
    # Sort for deterministic output
    IFS=$'\n' RULE_IDS=($(sort <<<"${RULE_IDS[*]}")); unset IFS
fi

for rule_id in "${RULE_IDS[@]}"; do
    # Verify rule exists in catalog
    if [ -z "${CATALOG_CATEGORY[$rule_id]+x}" ]; then
        echo -e "${BOLD}${rule_id}:${NC}"
        fail "Not found in error-catalog.yml"
        echo ""
        TOTAL_RULES=$((TOTAL_RULES + 1))
        continue
    fi

    # Find implementation file
    rule_file=""
    if ! rule_file=$(find_rule_file "$rule_id"); then
        echo -e "${BOLD}${rule_id}:${NC}"
        warn "No implementation file found (not yet implemented)"
        echo ""
        TOTAL_RULES=$((TOTAL_RULES + 1))
        continue
    fi

    echo -e "${BOLD}${rule_id}:${NC}"
    TOTAL_RULES=$((TOTAL_RULES + 1))
    rule_ok=true

    # 1. ID consistency
    impl_id=$(extract_method_value "$rule_file" "ID")
    if [ "$impl_id" = "$rule_id" ]; then
        pass "ID matches catalog entry"
    else
        fail "ID mismatch: implementation returns '${impl_id}', catalog expects '${rule_id}'"
        rule_ok=false
    fi

    # 2. Category consistency
    expected_cat=$(expected_category_from_id "$rule_id")
    catalog_cat="${CATALOG_CATEGORY[$rule_id]}"
    impl_cat=$(extract_method_value "$rule_file" "Category")
    if [ "$impl_cat" = "$catalog_cat" ] && [ "$impl_cat" = "$expected_cat" ]; then
        pass "Category: \"${impl_cat}\" matches"
    elif [ "$impl_cat" != "$catalog_cat" ]; then
        fail "Category: implementation returns '${impl_cat}', catalog expects '${catalog_cat}'"
        rule_ok=false
    else
        fail "Category: '${impl_cat}' does not match ID prefix expectation '${expected_cat}'"
        rule_ok=false
    fi

    # 3. Severity consistency
    catalog_sev="${CATALOG_SEVERITY[$rule_id]}"
    impl_sev=$(extract_method_value "$rule_file" "DefaultSeverity")
    if [ "$impl_sev" = "$catalog_sev" ]; then
        pass "Severity: \"${impl_sev}\" matches"
    else
        fail "Severity: implementation returns '${impl_sev}', catalog expects '${catalog_sev}'"
        rule_ok=false
    fi

    # 4. Message template key phrases
    if is_stub_rule "$rule_file"; then
        fail "Message: Check() returns nil (stub only)"
        rule_ok=false
    else
        catalog_msg="${CATALOG_MESSAGE[$rule_id]}"
        missing=""
        if missing=$(check_message_phrases "$rule_file" "$catalog_msg"); then
            pass "Message contains key template phrases"
        else
            fail "Message: missing key phrases from catalog template: ${missing}"
            rule_ok=false
        fi

        # 4b. Required violation fields
        missing_fields=""
        if missing_fields=$(check_violation_fields "$rule_file"); then
            pass "Violation sets all required fields"
        else
            fail "Violation missing required fields: ${missing_fields}"
            rule_ok=false
        fi
    fi

    # 5. Why() is non-empty
    impl_why=$(extract_method_value "$rule_file" "Why")
    if [ -n "$impl_why" ]; then
        pass "Why() is non-empty"
    else
        fail "Why() returns empty string"
        rule_ok=false
    fi

    # 6. Description() is non-empty
    impl_desc=$(extract_method_value "$rule_file" "Description")
    if [ -n "$impl_desc" ]; then
        pass "Description() is non-empty"
    else
        fail "Description() returns empty string"
        rule_ok=false
    fi

    if $rule_ok; then
        FULLY_VALIDATED=$((FULLY_VALIDATED + 1))
    fi

    echo ""
done

# ── Summary ───────────────────────────────────────────────
echo -e "${BOLD}Summary${NC}"
echo "======="
echo -e "  Rules checked:      $TOTAL_RULES"
echo -e "  ${GREEN}Fully validated:  $FULLY_VALIDATED${NC}"
echo -e "  ${GREEN}Checks passed:    $PASS_COUNT${NC}"
echo -e "  ${RED}Checks failed:    $FAIL_COUNT${NC}"
echo -e "  ${YELLOW}Warnings:         $WARN_COUNT${NC}"
echo ""

if [ "$TOTAL_RULES" -gt 0 ]; then
    echo -e "Summary: ${BOLD}${FULLY_VALIDATED}/${TOTAL_RULES}${NC} rules fully validated"
fi
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
    echo -e "${RED}Validation FAILED with $FAIL_COUNT mismatches.${NC}"
    exit 1
else
    echo -e "${GREEN}Validation PASSED.${NC}"
    exit 0
fi
