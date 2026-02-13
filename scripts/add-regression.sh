#!/usr/bin/env bash
# add-regression.sh — Automate the REGRESSION-PROTOCOL.md workflow.
set -euo pipefail

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# --- Usage ---
usage() {
    echo -e "${BOLD}Usage:${RESET} $0 <RULE_ID> <DESCRIPTION>"
    echo ""
    echo "  RULE_ID       Rule identifier (e.g., CONV-file-naming, TQ-no-shallow-assertions)"
    echo "  DESCRIPTION   Short description of the regression (quoted string)"
    echo ""
    echo -e "${BOLD}Examples:${RESET}"
    echo "  $0 CONV-file-naming \"missed kebab-case with numbers\""
    echo "  $0 TQ-no-shallow-assertions \"false positive on toEqual with nested objects\""
    exit 1
}

# --- Validate arguments ---
if [[ $# -lt 2 ]]; then
    echo -e "${RED}Error: Missing arguments.${RESET}"
    usage
fi

RULE_ID="$1"
DESCRIPTION="$2"

# --- Validate rule ID format ---
# Must match PREFIX-suffix where PREFIX is uppercase letters and suffix is lowercase/hyphens
if [[ ! "$RULE_ID" =~ ^[A-Z]+-[a-z][a-z0-9-]*$ ]]; then
    echo -e "${RED}Error: Invalid rule ID format '${RULE_ID}'.${RESET}"
    echo -e "Expected format: ${CYAN}PREFIX-kebab-name${RESET} (e.g., CONV-file-naming, TQ-error-path-coverage)"
    exit 1
fi

# --- Derive values ---
DATE="$(date +%Y-%m-%d)"

# Slugify description: lowercase, spaces to hyphens, strip special chars, collapse hyphens
SLUG="$(echo "$DESCRIPTION" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | sed 's/[^a-z0-9-]//g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')"

if [[ -z "$SLUG" ]]; then
    echo -e "${RED}Error: Description produced an empty slug after sanitization.${RESET}"
    exit 1
fi

# Determine file extension from rule prefix
PREFIX="${RULE_ID%%-*}"
case "$PREFIX" in
    CONV|ARCH|TQ|CTR)
        EXT="go"
        ;;
    TSQ)
        EXT="ts"
        ;;
    PYQ)
        EXT="py"
        ;;
    JQ)
        EXT="java"
        ;;
    *)
        EXT="go"
        ;;
esac

# --- Build paths ---
FIXTURE_DIR="tests/fixtures/regressions/${RULE_ID}/${DATE}-${SLUG}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FULL_DIR="${REPO_ROOT}/${FIXTURE_DIR}"

# --- Check for duplicates ---
if [[ -d "$FULL_DIR" ]]; then
    echo -e "${YELLOW}Warning: Directory already exists: ${FIXTURE_DIR}${RESET}"
    echo "Remove it first or use a different description."
    exit 1
fi

# --- Create directory ---
mkdir -p "$FULL_DIR"

# --- Create README.md ---
cat > "${FULL_DIR}/README.md" <<EOF
# Regression: ${RULE_ID}

**Date:** ${DATE}
**Description:** ${DESCRIPTION}
**Status:** OPEN — test should FAIL until fix is implemented

## Reproduction

<!-- Add minimal reproduction steps here -->

## Expected Behavior

<!-- What should happen -->

## Actual Behavior

<!-- What currently happens -->
EOF

# --- Create input file ---
case "$EXT" in
    go)
        cat > "${FULL_DIR}/input.${EXT}" <<EOF
package regression

// TODO: Add minimal reproduction code for ${RULE_ID}
// Description: ${DESCRIPTION}
EOF
        ;;
    ts)
        cat > "${FULL_DIR}/input.${EXT}" <<EOF
// TODO: Add minimal reproduction code for ${RULE_ID}
// Description: ${DESCRIPTION}
EOF
        ;;
    py)
        cat > "${FULL_DIR}/input.${EXT}" <<EOF
# TODO: Add minimal reproduction code for ${RULE_ID}
# Description: ${DESCRIPTION}
EOF
        ;;
    java)
        cat > "${FULL_DIR}/input.${EXT}" <<EOF
// TODO: Add minimal reproduction code for ${RULE_ID}
// Description: ${DESCRIPTION}
EOF
        ;;
esac

# --- Create expected.json ---
cat > "${FULL_DIR}/expected.json" <<EOF
{
  "rule_id": "${RULE_ID}",
  "expected_violations": 1,
  "description": "${DESCRIPTION}"
}
EOF

# --- Print success ---
echo ""
echo -e "${GREEN}${BOLD}Regression fixture created:${RESET}"
echo -e "  ${CYAN}${FIXTURE_DIR}/${RESET}"
echo ""
echo -e "${BOLD}Next steps:${RESET}"
echo -e "  1. Add reproduction code to ${CYAN}input.${EXT}${RESET}"
echo -e "  2. Write a failing test that uses this fixture"
echo -e "  3. Fix the bug"
echo -e "  4. Verify the test passes"
echo -e "  5. Update ${CYAN}README.md${RESET} status to FIXED"
