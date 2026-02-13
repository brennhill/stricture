#!/usr/bin/env bash
# validate-gate.sh — Checks whether all prerequisites for a development gate are met.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# ── Colors ────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# ── Phase-to-rule mapping ────────────────────────────────
PHASE_1_RULES=(
    "CONV-file-naming"
    "CONV-file-header"
)

PHASE_2_RULES=(
    "CONV-error-format"
    "CONV-export-naming"
    "CONV-test-file-location"
    "CONV-required-exports"
    "ARCH-no-circular-deps"
    "ARCH-dependency-direction"
    "ARCH-max-file-lines"
    "ARCH-import-boundary"
    "ARCH-layer-violation"
    "ARCH-module-boundary"
)

PHASE_3_RULES=(
    "TQ-no-shallow-assertions"
    "TQ-return-type-verified"
    "TQ-schema-conformance"
    "TQ-error-path-coverage"
    "TQ-assertion-depth"
    "TQ-boundary-tested"
    "TQ-mock-scope"
    "TQ-test-isolation"
    "TQ-negative-cases"
    "TQ-test-naming"
)

PHASE_4_RULES=(
    "CTR-request-shape"
    "CTR-response-shape"
    "CTR-status-code-handling"
    "CTR-shared-type-sync"
    "CTR-json-tag-match"
    "CTR-dual-test"
    "CTR-strictness-parity"
    "CTR-manifest-conformance"
)

ALL_RULES=(
    "${PHASE_1_RULES[@]}"
    "${PHASE_2_RULES[@]}"
    "${PHASE_3_RULES[@]}"
    "${PHASE_4_RULES[@]}"
)

# ── Helper: derive file paths from a rule ID ─────────────
# CONV-file-naming -> category=conv, name_part=file-naming
# Source file: internal/rules/conv/file_naming.go
# Test file:   internal/rules/conv/file_naming_test.go
# Fixture dir: tests/fixtures/CONV-file-naming/
# Test plan:   docs/test-plan/rules/conv.md

get_category() {
    local rule_id="$1"
    echo "$rule_id" | cut -d'-' -f1 | tr '[:upper:]' '[:lower:]'
}

get_source_file() {
    local rule_id="$1"
    local category
    category=$(get_category "$rule_id")
    # Strip category prefix and first hyphen, convert remaining hyphens to underscores
    local name_part
    name_part=$(echo "$rule_id" | sed "s/^[A-Z]*-//" | tr '-' '_')
    echo "internal/rules/${category}/${name_part}.go"
}

get_test_file() {
    local rule_id="$1"
    local source
    source=$(get_source_file "$rule_id")
    echo "${source%.go}_test.go"
}

get_test_plan_file() {
    local rule_id="$1"
    local category
    category=$(get_category "$rule_id")
    echo "docs/test-plan/rules/${category}.md"
}

get_fixture_dir() {
    local rule_id="$1"
    echo "tests/fixtures/$(echo "$rule_id" | tr '[:upper:]' '[:lower:]')/"
}

# ── Check functions ──────────────────────────────────────
# Each returns 0 for pass, 1 for fail, and prints formatted output.

TOTAL_PASS=0
TOTAL_FAIL=0

check_product_spec() {
    local rule_id="$1"
    local file="$PROJECT_ROOT/docs/product-spec.md"
    if [ ! -f "$file" ]; then
        printf "  ${RED}%s${NC} Product spec entry ........... docs/product-spec.md (FILE MISSING)\n" "❌"
        return 1
    fi
    local match
    match=$(grep -n "$rule_id" "$file" 2>/dev/null | head -1) || true
    if [ -n "$match" ]; then
        local lineno
        lineno=$(echo "$match" | cut -d: -f1)
        printf "  ${GREEN}%s${NC} Product spec entry ........... docs/product-spec.md:%s\n" "✅" "$lineno"
        return 0
    else
        printf "  ${RED}%s${NC} Product spec entry ........... docs/product-spec.md (MISSING)\n" "❌"
        return 1
    fi
}

check_error_catalog() {
    local rule_id="$1"
    local file="$PROJECT_ROOT/docs/error-catalog.yml"
    if [ ! -f "$file" ]; then
        printf "  ${RED}%s${NC} Error catalog entry .......... docs/error-catalog.yml (FILE MISSING)\n" "❌"
        return 1
    fi
    local match
    match=$(grep -n "$rule_id" "$file" 2>/dev/null | head -1) || true
    if [ -n "$match" ]; then
        local lineno
        lineno=$(echo "$match" | cut -d: -f1)
        printf "  ${GREEN}%s${NC} Error catalog entry .......... docs/error-catalog.yml:%s\n" "✅" "$lineno"
        return 0
    else
        printf "  ${RED}%s${NC} Error catalog entry .......... docs/error-catalog.yml (MISSING)\n" "❌"
        return 1
    fi
}

check_test_plan() {
    local rule_id="$1"
    local rel_path
    rel_path=$(get_test_plan_file "$rule_id")
    local file="$PROJECT_ROOT/$rel_path"
    if [ ! -f "$file" ]; then
        printf "  ${RED}%s${NC} Test plan entry .............. %s (FILE MISSING)\n" "❌" "$rel_path"
        return 1
    fi
    local match
    match=$(grep -n "$rule_id" "$file" 2>/dev/null | head -1) || true
    if [ -n "$match" ]; then
        local lineno
        lineno=$(echo "$match" | cut -d: -f1)
        printf "  ${GREEN}%s${NC} Test plan entry .............. %s:%s\n" "✅" "$rel_path" "$lineno"
        return 0
    else
        printf "  ${RED}%s${NC} Test plan entry .............. %s (MISSING)\n" "❌" "$rel_path"
        return 1
    fi
}

check_rule_implementation() {
    local rule_id="$1"
    local rel_path
    rel_path=$(get_source_file "$rule_id")
    local file="$PROJECT_ROOT/$rel_path"
    if [ -f "$file" ]; then
        printf "  ${GREEN}%s${NC} Rule implementation file ..... %s\n" "✅" "$rel_path"
        return 0
    else
        printf "  ${RED}%s${NC} Rule implementation file ..... %s (MISSING)\n" "❌" "$rel_path"
        return 1
    fi
}

check_rule_test() {
    local rule_id="$1"
    local rel_path
    rel_path=$(get_test_file "$rule_id")
    local file="$PROJECT_ROOT/$rel_path"
    if [ -f "$file" ]; then
        printf "  ${GREEN}%s${NC} Rule test file ............... %s\n" "✅" "$rel_path"
        return 0
    else
        printf "  ${RED}%s${NC} Rule test file ............... %s (MISSING)\n" "❌" "$rel_path"
        return 1
    fi
}

check_fixture_dir() {
    local rule_id="$1"
    local rel_path
    rel_path=$(get_fixture_dir "$rule_id")
    local dir="$PROJECT_ROOT/$rel_path"
    if [ -d "$dir" ]; then
        printf "  ${GREEN}%s${NC} Fixture directory ............ %s\n" "✅" "$rel_path"
        return 0
    else
        printf "  ${RED}%s${NC} Fixture directory ............ %s (MISSING)\n" "❌" "$rel_path"
        return 1
    fi
}

check_rule_index() {
    local rule_id="$1"
    local file="$PROJECT_ROOT/docs/RULE-INDEX.md"
    if [ ! -f "$file" ]; then
        printf "  ${RED}%s${NC} RULE-INDEX entry ............. docs/RULE-INDEX.md (FILE MISSING)\n" "❌"
        return 1
    fi
    local match
    match=$(grep -n "$rule_id" "$file" 2>/dev/null | head -1) || true
    if [ -n "$match" ]; then
        local lineno
        lineno=$(echo "$match" | cut -d: -f1)
        printf "  ${GREEN}%s${NC} RULE-INDEX entry ............. docs/RULE-INDEX.md:%s\n" "✅" "$lineno"
        return 0
    else
        printf "  ${RED}%s${NC} RULE-INDEX entry ............. docs/RULE-INDEX.md (MISSING)\n" "❌"
        return 1
    fi
}

# ── Validate a single rule ────────────────────────────────
# Returns the score (number of checks passed) via stdout capture.

validate_rule() {
    local rule_id="$1"
    local passed=0
    local total=7

    echo ""
    echo -e "${BOLD}Gate Validation: ${rule_id}${NC}"
    echo "==================================="

    check_product_spec "$rule_id"     && passed=$((passed + 1)) || true
    check_error_catalog "$rule_id"    && passed=$((passed + 1)) || true
    check_test_plan "$rule_id"        && passed=$((passed + 1)) || true
    check_rule_implementation "$rule_id" && passed=$((passed + 1)) || true
    check_rule_test "$rule_id"        && passed=$((passed + 1)) || true
    check_fixture_dir "$rule_id"      && passed=$((passed + 1)) || true
    check_rule_index "$rule_id"       && passed=$((passed + 1)) || true

    echo ""
    local failed=$((total - passed))
    echo -e "Score: ${passed}/${total} prerequisites met"

    if [ "$passed" -eq "$total" ]; then
        echo -e "Status: ${GREEN}READY${NC}"
    elif [ "$passed" -eq 0 ]; then
        echo -e "Status: ${RED}NOT STARTED (${failed} missing)${NC}"
    else
        echo -e "Status: ${YELLOW}MOSTLY READY (${failed} missing)${NC}"
    fi

    TOTAL_PASS=$((TOTAL_PASS + passed))
    TOTAL_FAIL=$((TOTAL_FAIL + failed))

    # Store score for summary table
    _LAST_SCORE="$passed/$total"
    _LAST_PASSED="$passed"
    _LAST_TOTAL="$total"
}

# ── Get phase number for a rule ──────────────────────────
get_phase_for_rule() {
    local rule_id="$1"
    for r in "${PHASE_1_RULES[@]}"; do [ "$r" = "$rule_id" ] && echo 1 && return; done
    for r in "${PHASE_2_RULES[@]}"; do [ "$r" = "$rule_id" ] && echo 2 && return; done
    for r in "${PHASE_3_RULES[@]}"; do [ "$r" = "$rule_id" ] && echo 3 && return; done
    for r in "${PHASE_4_RULES[@]}"; do [ "$r" = "$rule_id" ] && echo 4 && return; done
    echo 0
}

# ── Get rules for a given phase ──────────────────────────
get_rules_for_phase() {
    local phase="$1"
    case "$phase" in
        1) echo "${PHASE_1_RULES[@]}" ;;
        2) echo "${PHASE_2_RULES[@]}" ;;
        3) echo "${PHASE_3_RULES[@]}" ;;
        4) echo "${PHASE_4_RULES[@]}" ;;
        *) echo "" ;;
    esac
}

# ── Print summary table for a set of rules ───────────────
print_summary() {
    local label="$1"
    shift
    local rules=("$@")
    local fully_ready=0
    local total_rules=${#rules[@]}

    echo ""
    echo ""
    echo -e "${BOLD}${label}${NC}"
    echo "================"

    for rule_id in "${rules[@]}"; do
        local score="${SCORES[$rule_id]}"
        local passed="${PASSED[$rule_id]}"
        local total="${TOTALS[$rule_id]}"

        local icon
        if [ "$passed" -eq "$total" ]; then
            icon="${GREEN}✅${NC}"
            fully_ready=$((fully_ready + 1))
        elif [ "$passed" -eq 0 ]; then
            icon="${RED}❌${NC}"
        else
            icon="${YELLOW}⚠️${NC}"
        fi

        printf "  %-30s %s  %b\n" "$rule_id" "$score" "$icon"
    done

    echo ""
    echo -e "Overall: ${BOLD}${fully_ready}/${total_rules}${NC} rules fully ready"
}

# ── Validate a rule and store its score ───────────────────
declare -A SCORES
declare -A PASSED
declare -A TOTALS

validate_and_store() {
    local rule_id="$1"
    validate_rule "$rule_id"
    SCORES[$rule_id]="$_LAST_SCORE"
    PASSED[$rule_id]="$_LAST_PASSED"
    TOTALS[$rule_id]="$_LAST_TOTAL"
}

# ── Usage ────────────────────────────────────────────────
usage() {
    echo "Usage: $(basename "$0") [OPTIONS] [RULE-ID]"
    echo ""
    echo "Check whether all prerequisites for a development gate are met."
    echo ""
    echo "Arguments:"
    echo "  RULE-ID              Check a specific rule (e.g., CONV-file-naming)"
    echo ""
    echo "Options:"
    echo "  --all                Check all 30 rules across all phases"
    echo "  --phase N            Check all rules in phase N (1-4)"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Phases:"
    echo "  1  Foundation:   CONV-file-naming, CONV-file-header"
    echo "  2  TS + ARCH:    CONV-error-format, CONV-export-naming, + 8 more"
    echo "  3  Test Quality: TQ-no-shallow-assertions, TQ-return-type-verified, + 8 more"
    echo "  4  Contracts:    CTR-request-shape, CTR-response-shape, + 6 more"
    echo ""
    echo "Examples:"
    echo "  $(basename "$0") CONV-file-naming    # Check one rule"
    echo "  $(basename "$0") --all               # Check all rules"
    echo "  $(basename "$0") --phase 1           # Check Phase 1 rules"
}

# ── Validate that a rule ID is known ─────────────────────
is_valid_rule() {
    local rule_id="$1"
    for r in "${ALL_RULES[@]}"; do
        [ "$r" = "$rule_id" ] && return 0
    done
    return 1
}

# ── Main ─────────────────────────────────────────────────
main() {
    if [ $# -eq 0 ]; then
        usage
        exit 1
    fi

    local mode=""
    local target=""
    local any_failures=0

    while [ $# -gt 0 ]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            --all)
                mode="all"
                shift
                ;;
            --phase)
                mode="phase"
                if [ $# -lt 2 ]; then
                    echo -e "${RED}Error: --phase requires a phase number (1-4)${NC}" >&2
                    exit 1
                fi
                target="$2"
                if ! [[ "$target" =~ ^[1-4]$ ]]; then
                    echo -e "${RED}Error: phase must be 1, 2, 3, or 4${NC}" >&2
                    exit 1
                fi
                shift 2
                ;;
            -*)
                echo -e "${RED}Error: unknown option: $1${NC}" >&2
                usage
                exit 1
                ;;
            *)
                mode="single"
                target="$1"
                shift
                ;;
        esac
    done

    case "$mode" in
        single)
            if ! is_valid_rule "$target"; then
                echo -e "${RED}Error: unknown rule ID: ${target}${NC}" >&2
                echo "" >&2
                echo "Valid rule IDs:" >&2
                for r in "${ALL_RULES[@]}"; do
                    echo "  $r" >&2
                done
                exit 1
            fi
            validate_and_store "$target"
            if [ "${PASSED[$target]}" -lt "${TOTALS[$target]}" ]; then
                any_failures=1
            fi
            ;;
        phase)
            local rules
            read -r -a rules <<< "$(get_rules_for_phase "$target")"
            for rule_id in "${rules[@]}"; do
                validate_and_store "$rule_id"
                if [ "${PASSED[$rule_id]}" -lt "${TOTALS[$rule_id]}" ]; then
                    any_failures=1
                fi
            done
            print_summary "Phase ${target} Summary" "${rules[@]}"
            ;;
        all)
            for phase in 1 2 3 4; do
                local rules
                read -r -a rules <<< "$(get_rules_for_phase "$phase")"
                for rule_id in "${rules[@]}"; do
                    validate_and_store "$rule_id"
                    if [ "${PASSED[$rule_id]}" -lt "${TOTALS[$rule_id]}" ]; then
                        any_failures=1
                    fi
                done
            done
            for phase in 1 2 3 4; do
                local rules
                read -r -a rules <<< "$(get_rules_for_phase "$phase")"
                print_summary "Phase ${phase} Summary" "${rules[@]}"
            done
            ;;
    esac

    exit "$any_failures"
}

main "$@"
