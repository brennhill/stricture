#!/usr/bin/env bash
# run-validation-set.sh — Automated validation runner for Stricture.
# Extracts code blocks from validation markdown files, runs Stricture on each,
# and asserts expected violations.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VALIDATION_DIR="$PROJECT_ROOT/docs/test-plan/validation-set"
TEMP_DIR="${TMPDIR:-/tmp}/stricture-validation-$$"
STRICTURE="${STRICTURE_BIN:-$PROJECT_ROOT/bin/stricture}"
SINGLE_FILE=""

# ── Colors ────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ── Counters ──────────────────────────────────────────────
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0
ERRORS=""

# ── Bug-to-Rule Mapping ──────────────────────────────────
# Maps each bug level to the Stricture rule that should detect it.
declare -A BUG_RULE_MAP=(
    [B01]="TQ-error-path-coverage"
    [B02]="CTR-status-code-handling"
    [B03]="TQ-no-shallow-assertions"
    [B04]="TQ-negative-cases"
    [B05]="CTR-request-shape"
    [B06]="CTR-response-shape"
    [B07]="CTR-manifest-conformance"
    [B08]="CTR-strictness-parity"
    [B09]="CTR-strictness-parity"
    [B10]="CTR-strictness-parity"
    [B11]="CTR-strictness-parity"
    [B12]="CTR-response-shape"
    [B13]="CTR-request-shape"
    [B14]="CTR-response-shape"
    [B15]="CTR-request-shape"
)

API_PERFECT_RULES=(
    "TQ-error-path-coverage"
    "TQ-no-shallow-assertions"
    "TQ-negative-cases"
    "CTR-status-code-handling"
    "CTR-request-shape"
    "CTR-response-shape"
    "CTR-manifest-conformance"
    "CTR-strictness-parity"
)

ARCH_PERFECT_RULES=(
    "ARCH-dependency-direction"
    "ARCH-import-boundary"
    "ARCH-layer-violation"
    "ARCH-max-file-lines"
    "ARCH-module-boundary"
    "ARCH-no-circular-deps"
)

TQ_PERFECT_RULES=(
    "TQ-error-path-coverage"
    "TQ-no-shallow-assertions"
    "TQ-negative-cases"
    "TQ-return-type-verified"
    "TQ-schema-conformance"
    "TQ-assertion-depth"
    "TQ-boundary-tested"
    "TQ-mock-scope"
    "TQ-test-isolation"
    "TQ-test-naming"
)

get_scope_for_file() {
    local md_basename="$1"
    case "$md_basename" in
        30-*|31-*) echo "arch" ;;
        40-*|41-*) echo "tq" ;;
        *) echo "api" ;;
    esac
}

rules_for_scope() {
    local scope="$1"
    case "$scope" in
        arch) printf '%s\n' "${ARCH_PERFECT_RULES[@]}" ;;
        tq) printf '%s\n' "${TQ_PERFECT_RULES[@]}" ;;
        *) printf '%s\n' "${API_PERFECT_RULES[@]}" ;;
    esac
}

# ── Helpers ───────────────────────────────────────────────
log_pass() { echo -e "  ${GREEN}PASS${NC} $1"; PASSED=$((PASSED + 1)); TOTAL=$((TOTAL + 1)); }
log_fail() { echo -e "  ${RED}FAIL${NC} $1"; FAILED=$((FAILED + 1)); TOTAL=$((TOTAL + 1)); ERRORS="$ERRORS\n  FAIL: $1"; }
log_skip() { echo -e "  ${YELLOW}SKIP${NC} $1"; SKIPPED=$((SKIPPED + 1)); }
log_info() { echo -e "${CYAN}$1${NC}"; }

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# ── Extract Code Blocks from Markdown ─────────────────────
# Extracts fenced code blocks (```language ... ```) and writes them to files.
# Uses section headers (## PERFECT, ## B01, etc.) to determine the case ID.
extract_code_blocks() {
    local md_file="$1"
    local output_dir="$2"
    local current_section=""
    local in_code_block=false
    local code_lang=""
    local code_content=""
    local block_count=0

    mkdir -p "$output_dir"

    while IFS= read -r line; do
        # Detect section headers: ## PERFECT, ## B01, ### B01, etc.
        if [[ "$line" =~ ^##[[:space:]]+(PERFECT|B[0-9]{2}) ]] || \
           [[ "$line" =~ ^###[[:space:]]+(PERFECT|B[0-9]{2}) ]]; then
            current_section="${BASH_REMATCH[1]}"
            block_count=0
        fi

        # Detect code block start
        if [[ "$line" =~ ^\`\`\`(typescript|javascript|go|python|java|ts|js|py|yaml|yml) ]] && ! $in_code_block; then
            in_code_block=true
            code_lang="${BASH_REMATCH[1]}"
            code_content=""
            continue
        fi

        # Detect code block end
        if [[ "$line" == '```' ]] && $in_code_block; then
            in_code_block=false
            block_count=$((block_count + 1))

            if [ -n "$current_section" ] && [ -n "$code_content" ]; then
                # Determine file extension
                local ext
                case "$code_lang" in
                    typescript|ts) ext="ts" ;;
                    javascript|js) ext="js" ;;
                    go) ext="go" ;;
                    python|py) ext="py" ;;
                    java) ext="java" ;;
                    yaml|yml) ext="yml" ;;
                    *) ext="txt" ;;
                esac

                local section_dir="$output_dir/$current_section"
                mkdir -p "$section_dir"
                local filename="$section_dir/block-${block_count}.${ext}"
                echo "$code_content" > "$filename"
            fi
            continue
        fi

        # Accumulate code content
        if $in_code_block; then
            if [ -z "$code_content" ]; then
                code_content="$line"
            else
                code_content="$code_content
$line"
            fi
        fi
    done < "$md_file"
}

# ── Run Stricture on Extracted Code ───────────────────────
run_validation() {
    local case_dir="$1"
    local case_id="$2"
    local expected_rule="${3:-}"
    local md_basename="$4"
    local scope="${5:-api}"

    # Find all source files (not yaml/yml)
    local source_files
    source_files=$(find "$case_dir" -type f \( -name "*.ts" -o -name "*.js" -o -name "*.go" -o -name "*.py" -o -name "*.java" \) 2>/dev/null)

    if [ -z "$source_files" ]; then
        log_skip "$md_basename/$case_id — no source files extracted"
        return
    fi

    # Find manifest if present
    local manifest_file
    manifest_file=$(find "$case_dir" -type f \( -name "*.yml" -o -name "*.yaml" \) 2>/dev/null | head -1)

    local manifest_flag=""
    if [ -n "$manifest_file" ]; then
        manifest_flag="--manifest $manifest_file"
    fi

    if [ "$case_id" = "PERFECT" ]; then
        # PERFECT: expect zero violations for the scope's rule set.
        local total_violations=0
        local parsed=true
        local rule_id output exit_code violation_count
        while IFS= read -r rule_id; do
            [ -n "$rule_id" ] || continue
            exit_code=0
            # shellcheck disable=SC2086
            output=$($STRICTURE --format json --rule "$rule_id" $manifest_flag $source_files 2>&1) || exit_code=$?
            violation_count=$(echo "$output" | jq -r '.violations | length' 2>/dev/null || echo "parse_error")
            if [ "$violation_count" = "parse_error" ]; then
                parsed=false
                break
            fi
            total_violations=$((total_violations + violation_count))
        done < <(rules_for_scope "$scope")

        if [ "$parsed" = true ] && [ "$total_violations" = "0" ]; then
            log_pass "$md_basename/PERFECT — 0 violations (false positive check)"
        elif [ "$parsed" = false ]; then
            log_skip "$md_basename/PERFECT — could not parse Stricture output"
        else
            log_fail "$md_basename/PERFECT — expected 0 violations, got $total_violations"
        fi
    else
        # Bug case: expect at least one violation matching the expected rule
        if [ -z "$expected_rule" ]; then
            log_skip "$md_basename/$case_id — no expected rule mapped"
            return
        fi

        local output exit_code
        # shellcheck disable=SC2086
        output=$($STRICTURE --format json --rule "$expected_rule" $manifest_flag $source_files 2>&1) || exit_code=$?
        exit_code=${exit_code:-0}

        local has_expected_rule
        has_expected_rule=$(echo "$output" | jq -r ".violations[]? | (.RuleID // .ruleId // .rule // .rule_id) | select(. == \"$expected_rule\")" 2>/dev/null | head -1)

        if [ "$has_expected_rule" = "$expected_rule" ]; then
            log_pass "$md_basename/$case_id — detected $expected_rule"
        elif [ "$exit_code" -ne 0 ] && [ -z "$has_expected_rule" ]; then
            # Stricture found violations but not the expected rule
            local found_rules
            found_rules=$(echo "$output" | jq -r '.violations[]? | (.RuleID // .ruleId // .rule // .rule_id)' 2>/dev/null | sort -u | tr '\n' ',' | sed 's/,$//')
            log_fail "$md_basename/$case_id — expected $expected_rule, got: ${found_rules:-none}"
        else
            log_fail "$md_basename/$case_id — expected $expected_rule, got 0 violations"
        fi
    fi
}

# ── Main ──────────────────────────────────────────────────
main() {
    echo ""
    echo -e "${BOLD}Stricture Validation Set Runner${NC}"
    echo -e "${BOLD}================================${NC}"
    echo ""

    # Check if configured binary is executable.
    if [ ! -x "$STRICTURE" ]; then
        if command -v "$STRICTURE" >/dev/null 2>&1; then
            : # Command exists in PATH.
        else
            echo -e "${RED}Error: Stricture binary not found or not executable: ${STRICTURE}${NC}" >&2
            echo -e "${YELLOW}Build it first with: make build${NC}" >&2
            echo -e "${YELLOW}Or override path: STRICTURE_BIN=/path/to/stricture ./scripts/run-validation-set.sh${NC}" >&2
            exit 1
        fi
    fi

    mkdir -p "$TEMP_DIR"

    # Process API validation files (01-12, 13-17), or a single file if requested.
    local api_files
    if [ -n "$SINGLE_FILE" ]; then
        api_files="$SINGLE_FILE"
    else
        api_files=$(find "$VALIDATION_DIR" -maxdepth 1 -name "[0-9][0-9]-*.md" | sort)
    fi

    for md_file in $api_files; do
        local basename
        basename=$(basename "$md_file" .md)
        local scope
        scope=$(get_scope_for_file "$basename")
        log_info "Processing $basename..."

        local extract_dir="$TEMP_DIR/$basename"
        extract_code_blocks "$md_file" "$extract_dir"

        # Run PERFECT case
        if [ -d "$extract_dir/PERFECT" ]; then
            run_validation "$extract_dir/PERFECT" "PERFECT" "" "$basename" "$scope"
        else
            log_skip "$basename/PERFECT — no PERFECT section found"
        fi

        # Run B01-B15 bug cases
        for bug_num in $(seq -w 1 15); do
            local bug_id="B${bug_num}"
            local expected_rule="${BUG_RULE_MAP[$bug_id]:-}"
            if [ -d "$extract_dir/$bug_id" ]; then
                run_validation "$extract_dir/$bug_id" "$bug_id" "$expected_rule" "$basename" "$scope"
            fi
        done
    done

    if [ -z "$SINGLE_FILE" ]; then
        # Process architecture validation files (30-31)
        for md_file in "$VALIDATION_DIR"/3[0-9]-*.md; do
            [ -f "$md_file" ] || continue
            local basename
            basename=$(basename "$md_file" .md)
            log_info "Processing $basename (ARCH rules)..."

            local extract_dir="$TEMP_DIR/$basename"
            extract_code_blocks "$md_file" "$extract_dir"

            if [ -d "$extract_dir/PERFECT" ]; then
                run_validation "$extract_dir/PERFECT" "PERFECT" "" "$basename" "arch"
            fi
        done

        # Process test quality files (40-41)
        for md_file in "$VALIDATION_DIR"/4[0-9]-*.md; do
            [ -f "$md_file" ] || continue
            local basename
            basename=$(basename "$md_file" .md)
            log_info "Processing $basename (TQ rules)..."

            local extract_dir="$TEMP_DIR/$basename"
            extract_code_blocks "$md_file" "$extract_dir"

            if [ -d "$extract_dir/PERFECT" ]; then
                run_validation "$extract_dir/PERFECT" "PERFECT" "" "$basename" "tq"
            fi
        done
    fi

    # ── Summary ───────────────────────────────────────────
    echo ""
    echo -e "${BOLD}Results${NC}"
    echo "======="
    echo -e "  Total:   $TOTAL"
    echo -e "  ${GREEN}Passed:  $PASSED${NC}"
    echo -e "  ${RED}Failed:  $FAILED${NC}"
    echo -e "  ${YELLOW}Skipped: $SKIPPED${NC}"

    if [ $FAILED -gt 0 ]; then
        echo ""
        echo -e "${RED}Failures:${NC}"
        echo -e "$ERRORS"
    fi

    echo ""

    # Accuracy metrics
    if [ $TOTAL -gt 0 ]; then
        local accuracy
        accuracy=$(echo "scale=1; $PASSED * 100 / ($PASSED + $FAILED)" | bc 2>/dev/null || echo "N/A")
        echo -e "${BOLD}Detection accuracy: ${accuracy}%${NC}"

        # B01-B10 accuracy (should be 100%)
        # B11-B15 accuracy (should be >90%)
        echo -e "  Target: 100% for B01-B10, >90% for B11-B15"
    fi

    echo ""

    # Exit code: 0 if all passed, 1 if any failed
    if [ $FAILED -gt 0 ]; then
        exit 1
    fi
}

# ── Parse Args ────────────────────────────────────────────
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --file FILE    Run only the specified validation file"
        echo "  --verbose      Show Stricture output for each case"
        echo "  --dry-run      Extract code blocks but don't run Stricture"
        echo "  --help         Show this help"
        echo ""
        echo "Environment:"
        echo "  STRICTURE_BIN  Path to Stricture binary (default: npx stricture)"
        exit 0
        ;;
    --file)
        if [ -n "${2:-}" ]; then
            SINGLE_FILE="$2"
            VALIDATION_DIR="$(dirname "$2")"
        fi
        ;;
esac

main "$@"
