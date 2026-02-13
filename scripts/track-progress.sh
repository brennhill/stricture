#!/usr/bin/env bash
# track-progress.sh — Scan codebase and generate rule implementation progress dashboard.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# ── Paths ────────────────────────────────────────────────
PRODUCT_SPEC="$PROJECT_ROOT/docs/product-spec.md"
ERROR_CATALOG="$PROJECT_ROOT/docs/error-catalog.yml"
RULES_DIR="$PROJECT_ROOT/internal/rules"
FIXTURES_DIR="$PROJECT_ROOT/tests/fixtures"

# ── Colors ───────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
GRAY='\033[0;37m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# ── Flags ────────────────────────────────────────────────
JSON_MODE=false
RUN_TESTS=false

for arg in "$@"; do
    case "$arg" in
        --json) JSON_MODE=true ;;
        --test) RUN_TESTS=true ;;
        --help|-h)
            echo "Usage: $0 [--json] [--test]"
            echo ""
            echo "  --json    Output JSON status to stdout"
            echo "  --test    Run go test for each implemented rule (slow)"
            echo ""
            echo "Without --test, test pass/fail column shows unknown (not run)."
            exit 0
            ;;
        *)
            echo "Unknown flag: $arg" >&2
            echo "Usage: $0 [--json] [--test]" >&2
            exit 1
            ;;
    esac
done

# ── Rule definitions ─────────────────────────────────────
# Format: PHASE|RULE_ID|CATEGORY|SHORT_NAME
# CATEGORY is the subdirectory under internal/rules/ and file name prefix
# SHORT_NAME is the snake_case base for .go files
#
# The 30 rules as specified, mapped to their actual IDs in product-spec.md
# and error-catalog.yml, organized into 4 phases.

RULES=(
    # Phase 1: Foundation (2 rules)
    "1|CONV-file-naming|conv|file_naming"
    "1|CONV-file-header|conv|file_header"

    # Phase 2: Convention + Architecture (10 rules)
    "2|CONV-error-format|conv|error_format"
    "2|CONV-export-naming|conv|export_naming"
    "2|CONV-test-file-location|conv|test_file_location"
    "2|CONV-required-exports|conv|required_exports"
    "2|ARCH-no-circular-deps|arch|no_circular_deps"
    "2|ARCH-dependency-direction|arch|dependency_direction"
    "2|ARCH-max-file-lines|arch|max_file_lines"
    "2|ARCH-import-boundary|arch|import_boundary"
    "2|ARCH-layer-violation|arch|layer_violation"
    "2|ARCH-module-boundary|arch|module_boundary"

    # Phase 3: Test Quality (10 rules)
    "3|TQ-no-shallow-assertions|tq|no_shallow_assertions"
    "3|TQ-return-type-verified|tq|return_type_verified"
    "3|TQ-schema-conformance|tq|schema_conformance"
    "3|TQ-error-path-coverage|tq|error_path_coverage"
    "3|TQ-assertion-depth|tq|assertion_depth"
    "3|TQ-boundary-tested|tq|boundary_tested"
    "3|TQ-mock-scope|tq|mock_scope"
    "3|TQ-test-isolation|tq|test_isolation"
    "3|TQ-negative-cases|tq|negative_cases"
    "3|TQ-test-naming|tq|test_naming"

    # Phase 4: Contracts (8 rules)
    "4|CTR-request-shape|ctr|request_shape"
    "4|CTR-response-shape|ctr|response_shape"
    "4|CTR-status-code-handling|ctr|status_code_handling"
    "4|CTR-shared-type-sync|ctr|shared_type_sync"
    "4|CTR-json-tag-match|ctr|json_tag_match"
    "4|CTR-dual-test|ctr|dual_test"
    "4|CTR-strictness-parity|ctr|strictness_parity"
    "4|CTR-manifest-conformance|ctr|manifest_conformance"
)

PHASE_NAMES=(
    [1]="Foundation"
    [2]="Convention + Architecture"
    [3]="Test Quality"
    [4]="Contracts"
)

# ── Helper functions ─────────────────────────────────────

# Check if a rule ID appears in the product spec.
has_spec() {
    local rule_id="$1"
    grep -q "$rule_id" "$PRODUCT_SPEC" 2>/dev/null
}

# Check if a rule ID appears in the error catalog.
has_catalog() {
    local rule_id="$1"
    grep -q "$rule_id" "$ERROR_CATALOG" 2>/dev/null
}

# Check if the implementation file exists and has a real implementation
# (not just a stub that returns nil).
has_impl() {
    local category="$1"
    local short_name="$2"
    local impl_file="$RULES_DIR/$category/${short_name}.go"

    if [ ! -f "$impl_file" ]; then
        return 1
    fi

    # Check if Check() method has more than just "return nil"
    # A stub typically has only "return nil" in the Check function body.
    # A real implementation will have additional logic before return.
    local check_body
    check_body=$(sed -n '/^func.*Check(/,/^}/p' "$impl_file" 2>/dev/null || true)

    if [ -z "$check_body" ]; then
        return 1
    fi

    # If it contains "TODO" and only returns nil, it is a stub.
    if echo "$check_body" | grep -q "TODO" && \
       ! echo "$check_body" | grep -v "TODO" | grep -v "^func" | grep -v "^}" | \
         grep -v "return nil" | grep -v "^[[:space:]]*$" | grep -v "^[[:space:]]*//" | \
         grep -q "[a-zA-Z]"; then
        return 1
    fi

    # Check for a non-trivial implementation: more than just return nil
    local non_trivial
    non_trivial=$(echo "$check_body" | grep -v "^func" | grep -v "^}" | \
        grep -v "return nil" | grep -v "^[[:space:]]*$" | \
        grep -v "^[[:space:]]*//" | grep -c "[a-zA-Z]" || true)

    if [ "$non_trivial" -gt 0 ]; then
        return 0
    fi

    return 1
}

# Check if test file exists and has real tests (more than just interface compliance).
has_tests() {
    local category="$1"
    local short_name="$2"
    local test_file="$RULES_DIR/$category/${short_name}_test.go"

    if [ ! -f "$test_file" ]; then
        return 1
    fi

    # Count func Test occurrences
    local count
    count=$(grep -c "^func Test" "$test_file" 2>/dev/null || true)

    if [ "$count" -gt 0 ]; then
        return 0
    fi

    return 1
}

# Count func Test occurrences in a test file.
test_count() {
    local category="$1"
    local short_name="$2"
    local test_file="$RULES_DIR/$category/${short_name}_test.go"

    if [ ! -f "$test_file" ]; then
        echo "0"
        return
    fi

    grep -c "^func Test" "$test_file" 2>/dev/null || echo "0"
}

# Count lines of code in implementation file.
loc_count() {
    local category="$1"
    local short_name="$2"
    local impl_file="$RULES_DIR/$category/${short_name}.go"

    if [ ! -f "$impl_file" ]; then
        echo "0"
        return
    fi

    wc -l < "$impl_file" | tr -d ' '
}

# Check if fixture directories exist for this rule.
has_fixtures() {
    local rule_id="$1"
    local category="$2"
    local short_name="$3"

    # Check for fixtures named after the rule in tests/fixtures/
    # Fixture dirs can be numbered (01-validation, etc.) or named by rule
    local rule_lower
    rule_lower=$(echo "$rule_id" | tr '[:upper:]' '[:lower:]')

    if ls "$FIXTURES_DIR"/*"$short_name"* 1>/dev/null 2>&1 || \
       ls "$FIXTURES_DIR"/*"$rule_lower"* 1>/dev/null 2>&1; then
        return 0
    fi

    return 1
}

# Run go test for a rule's package. Returns 0 if tests pass, 1 if they fail.
run_test() {
    local category="$1"
    local pkg_dir="$RULES_DIR/$category"

    if [ ! -d "$pkg_dir" ]; then
        return 1
    fi

    # Check if there are any test files
    if ! ls "$pkg_dir"/*_test.go 1>/dev/null 2>&1; then
        return 1
    fi

    cd "$PROJECT_ROOT"
    if go test "./internal/rules/$category/..." -count=1 -timeout=30s >/dev/null 2>&1; then
        return 0
    fi

    return 1
}

# ── Collect data ─────────────────────────────────────────

# Associative arrays are not portable to older bash; use indexed arrays and functions.
# We store results as delimited strings: RULE_ID|has_spec|has_catalog|has_impl|has_tests|tests_pass|has_fixtures|loc|test_count

declare -a RESULTS=()
declare -A TESTED_PACKAGES=()

for entry in "${RULES[@]}"; do
    IFS='|' read -r phase rule_id category short_name <<< "$entry"

    # Spec
    spec="false"
    if has_spec "$rule_id"; then
        spec="true"
    fi

    # Catalog
    catalog="false"
    if has_catalog "$rule_id"; then
        catalog="true"
    fi

    # Implementation
    impl="false"
    if has_impl "$category" "$short_name"; then
        impl="true"
    fi

    # Tests
    tests="false"
    tc="0"
    if has_tests "$category" "$short_name"; then
        tests="true"
        tc=$(test_count "$category" "$short_name")
    fi

    # Test results
    pass="unknown"
    if [ "$RUN_TESTS" = true ] && [ "$impl" = "true" ]; then
        # Cache test results per package (all rules in same category share a package)
        if [ -z "${TESTED_PACKAGES[$category]+x}" ]; then
            if run_test "$category"; then
                TESTED_PACKAGES[$category]="pass"
            else
                TESTED_PACKAGES[$category]="fail"
            fi
        fi
        pass="${TESTED_PACKAGES[$category]}"
    fi

    # Fixtures
    fixtures="false"
    if has_fixtures "$rule_id" "$category" "$short_name"; then
        fixtures="true"
    fi

    # LOC
    loc=$(loc_count "$category" "$short_name")

    RESULTS+=("$phase|$rule_id|$category|$short_name|$spec|$catalog|$impl|$tests|$pass|$fixtures|$loc|$tc")
done

# ── JSON output ──────────────────────────────────────────

if [ "$JSON_MODE" = true ]; then
    # Compute summary
    total=0
    spec_count=0
    catalog_count=0
    impl_count=0
    tests_count=0
    passing_count=0

    for result in "${RESULTS[@]}"; do
        IFS='|' read -r phase rule_id category short_name spec catalog impl tests pass fixtures loc tc <<< "$result"
        total=$((total + 1))
        [ "$spec" = "true" ] && spec_count=$((spec_count + 1))
        [ "$catalog" = "true" ] && catalog_count=$((catalog_count + 1))
        [ "$impl" = "true" ] && impl_count=$((impl_count + 1))
        [ "$tests" = "true" ] && tests_count=$((tests_count + 1))
        [ "$pass" = "pass" ] && passing_count=$((passing_count + 1))
    done

    # Build phase data
    declare -A PHASE_COMPLETE=()
    declare -A PHASE_TOTAL=()
    for result in "${RESULTS[@]}"; do
        IFS='|' read -r phase rule_id category short_name spec catalog impl tests pass fixtures loc tc <<< "$result"
        PHASE_TOTAL[$phase]=$(( ${PHASE_TOTAL[$phase]:-0} + 1 ))
        if [ "$impl" = "true" ] && [ "$tests" = "true" ] && { [ "$pass" = "pass" ] || [ "$pass" = "unknown" ]; }; then
            PHASE_COMPLETE[$phase]=$(( ${PHASE_COMPLETE[$phase]:-0} + 1 ))
        fi
    done

    generated_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Start JSON
    printf '{\n'
    printf '  "generated_at": "%s",\n' "$generated_at"
    printf '  "summary": {\n'
    printf '    "total_rules": %d,\n' "$total"
    printf '    "spec_complete": %d,\n' "$spec_count"
    printf '    "catalog_complete": %d,\n' "$catalog_count"
    printf '    "impl_complete": %d,\n' "$impl_count"
    printf '    "tests_complete": %d,\n' "$tests_count"
    printf '    "passing": %d\n' "$passing_count"
    printf '  },\n'

    # Phases
    printf '  "phases": {\n'
    first_phase=true
    for p in 1 2 3 4; do
        if [ "$first_phase" = true ]; then
            first_phase=false
        else
            printf ',\n'
        fi
        phase_name="${PHASE_NAMES[$p]}"
        phase_complete="${PHASE_COMPLETE[$p]:-0}"
        phase_total="${PHASE_TOTAL[$p]:-0}"

        printf '    "%d": {\n' "$p"
        printf '      "name": "%s",\n' "$phase_name"
        printf '      "complete": %d,\n' "$phase_complete"
        printf '      "total": %d,\n' "$phase_total"

        # Collect rule IDs for this phase
        printf '      "rules": ['
        first_rule=true
        for result in "${RESULTS[@]}"; do
            IFS='|' read -r rp rule_id _ _ _ _ _ _ _ _ _ _ <<< "$result"
            if [ "$rp" = "$p" ]; then
                if [ "$first_rule" = true ]; then
                    first_rule=false
                else
                    printf ', '
                fi
                printf '"%s"' "$rule_id"
            fi
        done
        printf ']\n'
        printf '    }'
    done
    printf '\n  },\n'

    # Rules
    printf '  "rules": {\n'
    first_rule=true
    for result in "${RESULTS[@]}"; do
        IFS='|' read -r phase rule_id category short_name spec catalog impl tests pass fixtures loc tc <<< "$result"

        if [ "$first_rule" = true ]; then
            first_rule=false
        else
            printf ',\n'
        fi

        # Convert pass to JSON-friendly value
        tests_pass_json="null"
        if [ "$pass" = "pass" ]; then
            tests_pass_json="true"
        elif [ "$pass" = "fail" ]; then
            tests_pass_json="false"
        fi

        printf '    "%s": {\n' "$rule_id"
        printf '      "phase": %d,\n' "$phase"
        printf '      "category": "%s",\n' "$category"
        printf '      "has_spec": %s,\n' "$spec"
        printf '      "has_catalog": %s,\n' "$catalog"
        printf '      "has_impl": %s,\n' "$impl"
        printf '      "has_tests": %s,\n' "$tests"
        printf '      "tests_pass": %s,\n' "$tests_pass_json"
        printf '      "has_fixtures": %s,\n' "$fixtures"
        printf '      "loc": %d,\n' "$loc"
        printf '      "test_count": %d\n' "$tc"
        printf '    }'
    done
    printf '\n  }\n'
    printf '}\n'

    exit 0
fi

# ── Text dashboard ───────────────────────────────────────

icon_ok="${GREEN}ok${NC}"
icon_no="${RED}--${NC}"
icon_skip="${DIM}??${NC}"

check_icon() {
    local val="$1"
    if [ "$val" = "true" ]; then
        printf '%b' "$icon_ok"
    elif [ "$val" = "pass" ]; then
        printf '%b' "$icon_ok"
    elif [ "$val" = "fail" ]; then
        printf '%b' "$icon_no"
    elif [ "$val" = "unknown" ]; then
        printf '%b' "$icon_skip"
    else
        printf '%b' "$icon_no"
    fi
}

echo ""
echo -e "${BOLD}Stricture Rule Implementation Progress${NC}"
echo -e "${BOLD}========================================${NC}"
echo ""

# Summary counters
total=0
total_spec=0
total_catalog=0
total_impl=0
total_tests=0
total_passing=0

current_phase=0

for result in "${RESULTS[@]}"; do
    IFS='|' read -r phase rule_id category short_name spec catalog impl tests pass fixtures loc tc <<< "$result"

    total=$((total + 1))
    [ "$spec" = "true" ] && total_spec=$((total_spec + 1))
    [ "$catalog" = "true" ] && total_catalog=$((total_catalog + 1))
    [ "$impl" = "true" ] && total_impl=$((total_impl + 1))
    [ "$tests" = "true" ] && total_tests=$((total_tests + 1))
    [ "$pass" = "pass" ] && total_passing=$((total_passing + 1))

    # Phase header
    if [ "$phase" != "$current_phase" ]; then
        if [ "$current_phase" != "0" ]; then
            # Print phase progress for previous phase
            echo -e "  ${DIM}Progress: ${phase_impl}/${phase_total} implemented, ${phase_complete}/${phase_total} complete${NC}"
            echo ""
        fi
        current_phase="$phase"
        phase_name="${PHASE_NAMES[$phase]}"
        phase_total=0
        phase_impl=0
        phase_complete=0

        # Count rules in this phase
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r rp _ _ _ _ _ ri rt rpass _ _ _ <<< "$r"
            if [ "$rp" = "$phase" ]; then
                phase_total=$((phase_total + 1))
            fi
        done

        echo -e "${BOLD}Phase $phase: $phase_name ($phase_total rules)${NC}"
    fi

    # Track per-phase counts
    if [ "$impl" = "true" ]; then
        phase_impl=$((phase_impl + 1))
    fi
    if [ "$impl" = "true" ] && [ "$tests" = "true" ] && { [ "$pass" = "pass" ] || [ "$pass" = "unknown" ]; }; then
        phase_complete=$((phase_complete + 1))
    fi

    # LOC display
    loc_display="-- LOC"
    if [ "$loc" -gt 0 ]; then
        loc_display="$loc LOC"
    fi

    # Pad rule_id to 30 chars for alignment
    padded_id=$(printf '%-30s' "$rule_id")

    # Build status line
    spec_icon=$(check_icon "$spec")
    cat_icon=$(check_icon "$catalog")
    impl_icon=$(check_icon "$impl")
    tests_icon=$(check_icon "$tests")
    pass_icon=$(check_icon "$pass")

    printf "  %s  spec:%b  cat:%b  impl:%b  test:%b  pass:%b  [%s]\n" \
        "$padded_id" "$spec_icon" "$cat_icon" "$impl_icon" "$tests_icon" "$pass_icon" "$loc_display"
done

# Print final phase progress
if [ "$current_phase" != "0" ]; then
    echo -e "  ${DIM}Progress: ${phase_impl}/${phase_total} implemented, ${phase_complete}/${phase_total} complete${NC}"
fi

echo ""
echo -e "${BOLD}========================================${NC}"

# Calculate percentage
if [ "$total" -gt 0 ]; then
    pct=$((total_impl * 100 / total))
else
    pct=0
fi

echo -e "${BOLD}TOTAL: ${total_impl}/${total} rules implemented (${pct}%)${NC}"
echo ""
echo -e "  Spec:      ${total_spec}/${total}"
echo -e "  Catalog:   ${total_catalog}/${total}"
echo -e "  Impl:      ${total_impl}/${total}"
echo -e "  Tests:     ${total_tests}/${total}"
if [ "$RUN_TESTS" = true ]; then
    echo -e "  Passing:   ${total_passing}/${total}"
else
    echo -e "  Passing:   ${DIM}(run with --test)${NC}"
fi
echo ""
