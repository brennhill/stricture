#!/usr/bin/env bash
# validation-health-check.sh — Validates the validation set itself.
# "Who watches the watchmen?" Ensures all validation files are well-formed.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VALIDATION_DIR="$PROJECT_ROOT/docs/test-plan/validation-set"

# ── Colors ────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

ERRORS=0
WARNINGS=0
CHECKS=0

check_pass() { echo -e "  ${GREEN}OK${NC}   $1"; CHECKS=$((CHECKS + 1)); }
check_fail() { echo -e "  ${RED}FAIL${NC} $1"; ERRORS=$((ERRORS + 1)); CHECKS=$((CHECKS + 1)); }
check_warn() { echo -e "  ${YELLOW}WARN${NC} $1"; WARNINGS=$((WARNINGS + 1)); CHECKS=$((CHECKS + 1)); }

echo ""
echo -e "${BOLD}Stricture Validation Set Health Check${NC}"
echo -e "${BOLD}======================================${NC}"
echo ""

# ── Check 1: API validation files (01-17) have PERFECT + B01-B15 ──
echo -e "${BOLD}1. API Validation Files — Structure${NC}"

for md_file in "$VALIDATION_DIR"/[0-1][0-9]-*.md; do
    [ -f "$md_file" ] || continue
    basename=$(basename "$md_file" .md)

    # Check for PERFECT section
    if grep -q "^##.*PERFECT\|^###.*PERFECT" "$md_file"; then
        check_pass "$basename has PERFECT section"
    else
        check_fail "$basename MISSING PERFECT section"
    fi

    # Check for bug sections (B01-B15)
    local_bugs=0
    for bug in B01 B02 B03 B04 B05 B06 B07 B08 B09 B10 B11 B12 B13 B14 B15; do
        if grep -q "^##.*$bug\|^###.*$bug" "$md_file"; then
            local_bugs=$((local_bugs + 1))
        fi
    done

    if [ "$local_bugs" -eq 15 ]; then
        check_pass "$basename has all 15 bug levels (B01-B15)"
    elif [ "$local_bugs" -gt 0 ]; then
        check_warn "$basename has $local_bugs/15 bug levels"
    else
        check_fail "$basename has NO bug levels"
    fi
done

# ── Check 2: Code blocks have valid fenced syntax ─────────
echo ""
echo -e "${BOLD}2. Code Block Syntax${NC}"

for md_file in "$VALIDATION_DIR"/[0-9][0-9]-*.md "$VALIDATION_DIR"/[0-9][0-9][0-9]-*.md; do
    [ -f "$md_file" ] || continue
    basename=$(basename "$md_file" .md)

    # Count all fence markers and require an even number.
    fence_markers=$(grep -c '^```' "$md_file" 2>/dev/null || true)
    if [ $((fence_markers % 2)) -eq 0 ]; then
        check_pass "$basename: $((fence_markers / 2)) code blocks balanced"
    else
        check_fail "$basename: odd fenced block marker count ($fence_markers)"
    fi
done

# ── Check 3: Manifest fragments are valid YAML ────────────
echo ""
echo -e "${BOLD}3. Manifest Fragment Validity${NC}"

for md_file in "$VALIDATION_DIR"/[0-1][0-9]-*.md; do
    [ -f "$md_file" ] || continue
    basename=$(basename "$md_file" .md)

    # Extract YAML blocks and check basic validity
    has_manifest=false
    if grep -q "^## Manifest\|^### Manifest\|manifest_version" "$md_file"; then
        has_manifest=true
    fi

    if $has_manifest; then
        # Check for required manifest fields
        if grep -q "contracts:" "$md_file"; then
            check_pass "$basename: manifest has 'contracts:' key"
        else
            check_warn "$basename: manifest found but no 'contracts:' key"
        fi
    else
        check_warn "$basename: no manifest fragment found"
    fi
done

# ── Check 4: Bug levels map to documented rules ──────────
echo ""
echo -e "${BOLD}4. Bug-to-Rule Mapping${NC}"

readme="$VALIDATION_DIR/README.md"
if [ -f "$readme" ]; then
    # Check that README documents the bug taxonomy
    for bug in B01 B02 B03 B04 B05 B06 B07 B08 B09 B10 B11 B12 B13 B14 B15; do
        if grep -q "$bug" "$readme"; then
            check_pass "README documents $bug"
        else
            check_fail "README missing $bug mapping"
        fi
    done
else
    check_fail "README.md not found"
fi

# ── Check 5: No duplicate contract IDs across files ───────
echo ""
echo -e "${BOLD}5. Contract ID Uniqueness${NC}"

contract_ids=$(grep -h '^[[:space:]]*-[[:space:]]*id:[[:space:]]*".*"' "$VALIDATION_DIR"/[0-1][0-9]-*.md 2>/dev/null | \
    sed 's/.*id:[[:space:]]*"\([^"]*\)".*/\1/' | sort)

duplicates=$(echo "$contract_ids" | uniq -d)

if [ -z "$duplicates" ]; then
    total_ids=$(echo "$contract_ids" | wc -l | tr -d ' ')
    check_pass "All contract IDs unique ($total_ids total)"
else
    check_fail "Duplicate contract IDs found: $duplicates"
fi

# ── Check 6: Stricture rule references are valid ──────────
echo ""
echo -e "${BOLD}6. Rule Reference Validity${NC}"

valid_rules=(
    "TQ-no-shallow-assertions" "TQ-return-type-verified" "TQ-schema-conformance"
    "TQ-error-path-coverage" "TQ-assertion-depth" "TQ-boundary-tested"
    "TQ-mock-scope" "TQ-test-isolation" "TQ-negative-cases" "TQ-test-naming"
    "ARCH-dependency-direction" "ARCH-import-boundary" "ARCH-no-circular-deps"
    "ARCH-max-file-lines" "ARCH-layer-violation" "ARCH-module-boundary"
    "CONV-file-naming" "CONV-file-header" "CONV-error-format"
    "CONV-export-naming" "CONV-test-file-location" "CONV-required-exports"
    "CTR-request-shape" "CTR-response-shape" "CTR-status-code-handling"
    "CTR-shared-type-sync" "CTR-json-tag-match" "CTR-dual-test"
    "CTR-strictness-parity" "CTR-manifest-conformance"
)

# Extract all rule references from validation files
referenced_rules=$(grep -roh '\(TQ\|ARCH\|CONV\|CTR\)-[a-z-]*' "$VALIDATION_DIR"/*.md 2>/dev/null | sort -u)

invalid_count=0
while IFS= read -r rule; do
    [ -z "$rule" ] && continue
    found=false
    for valid in "${valid_rules[@]}"; do
        if [ "$rule" = "$valid" ]; then
            found=true
            break
        fi
    done
    if ! $found; then
        check_warn "Unknown rule reference: $rule"
        invalid_count=$((invalid_count + 1))
    fi
done <<< "$referenced_rules"

if [ "$invalid_count" -eq 0 ]; then
    total_refs=$(echo "$referenced_rules" | wc -l | tr -d ' ')
    check_pass "All $total_refs rule references are valid"
fi

# ── Check 7: Language coverage ────────────────────────────
echo ""
echo -e "${BOLD}7. Language Coverage${NC}"

for lang in typescript go python java; do
    count=$(grep -rl "\`\`\`${lang}" "$VALIDATION_DIR"/*.md 2>/dev/null | wc -l | tr -d ' ')
    if [ "$count" -gt 0 ]; then
        check_pass "$lang: $count validation files"
    else
        check_warn "$lang: 0 validation files"
    fi
done

# ── Check 8: File sizes ──────────────────────────────────
echo ""
echo -e "${BOLD}8. File Size Sanity${NC}"

for md_file in "$VALIDATION_DIR"/*.md; do
    [ -f "$md_file" ] || continue
    basename=$(basename "$md_file")
    lines=$(wc -l < "$md_file" | tr -d ' ')

    if [ "$lines" -lt 50 ]; then
        check_warn "$basename: only $lines lines (suspiciously short)"
    elif [ "$lines" -gt 5000 ]; then
        check_warn "$basename: $lines lines (consider splitting)"
    else
        check_pass "$basename: $lines lines"
    fi
done

# ── Check 9: Cross-reference integrity ────────────────────
echo ""
echo -e "${BOLD}9. Cross-Reference Integrity${NC}"

if [ -f "$VALIDATION_DIR/README.md" ]; then
    # Check that README links to files that exist
    linked_files=$(grep -o '\[.*\]([^)]*\.md)' "$VALIDATION_DIR/README.md" | \
        sed 's/.*(\([^)]*\))/\1/' | sort -u)

    broken=0
    while IFS= read -r link; do
        [ -z "$link" ] && continue
        target="$VALIDATION_DIR/$link"
        if [ -f "$target" ]; then
            check_pass "README link → $link exists"
        else
            check_fail "README link → $link BROKEN (file not found)"
            broken=$((broken + 1))
        fi
    done <<< "$linked_files"
fi

# ── Check 10: Rule coverage completeness ──────────────────
echo ""
echo -e "${BOLD}10. Rule Coverage Summary${NC}"

covered=0
uncovered_list=""
for rule in "${valid_rules[@]}"; do
    if grep -rq "$rule" "$VALIDATION_DIR" 2>/dev/null; then
        covered=$((covered + 1))
    else
        uncovered_list="$uncovered_list $rule"
    fi
done

total_rules=${#valid_rules[@]}
echo -e "  Rules with validation coverage: ${GREEN}$covered/$total_rules${NC}"
if [ -n "$uncovered_list" ]; then
    echo -e "  ${YELLOW}Uncovered:${NC}$uncovered_list"
fi

# ── Summary ───────────────────────────────────────────────
echo ""
echo -e "${BOLD}Summary${NC}"
echo "======="
echo -e "  Checks:   $CHECKS"
echo -e "  ${GREEN}Passed:   $((CHECKS - ERRORS - WARNINGS))${NC}"
echo -e "  ${RED}Errors:   $ERRORS${NC}"
echo -e "  ${YELLOW}Warnings: $WARNINGS${NC}"
echo ""

if [ "$ERRORS" -gt 0 ]; then
    echo -e "${RED}Health check FAILED with $ERRORS errors.${NC}"
    exit 1
else
    echo -e "${GREEN}Health check PASSED.${NC}"
fi
