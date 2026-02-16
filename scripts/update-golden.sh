#!/usr/bin/env bash
set -euo pipefail

# update-golden.sh — Regenerates golden output files from a known violation set.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
STRICTURE="${PROJECT_ROOT}/bin/strict"
GOLDEN_DIR="${PROJECT_ROOT}/tests/golden"

# Check prerequisites
if [ ! -f "${STRICTURE}" ]; then
    echo "ERROR: Stricture binary not found at ${STRICTURE}"
    echo "Run 'make build' first."
    exit 2
fi

if [ ! -d "${GOLDEN_DIR}/input" ]; then
    echo "ERROR: Golden input directory not found at ${GOLDEN_DIR}/input"
    echo "Create golden input fixtures first."
    exit 2
fi

echo "Regenerating golden output files..."
echo "  Binary: ${STRICTURE}"
echo "  Input:  ${GOLDEN_DIR}/input"
echo "  Output: ${GOLDEN_DIR}"
echo ""

# Create output directory
mkdir -p "${GOLDEN_DIR}"

# Run stricture on golden fixture input with each format
echo "Generating text format..."
"${STRICTURE}" --format text "${GOLDEN_DIR}/input/" > "${GOLDEN_DIR}/output-text.txt"

echo "Generating JSON format..."
"${STRICTURE}" --format json "${GOLDEN_DIR}/input/" > "${GOLDEN_DIR}/output.json"

echo "Generating SARIF format..."
"${STRICTURE}" --format sarif "${GOLDEN_DIR}/input/" > "${GOLDEN_DIR}/output.sarif"

echo "Generating JUnit XML format..."
"${STRICTURE}" --format junit "${GOLDEN_DIR}/input/" > "${GOLDEN_DIR}/output.junit.xml"

echo ""
echo "Golden files updated successfully:"
echo "  - ${GOLDEN_DIR}/output-text.txt"
echo "  - ${GOLDEN_DIR}/output.json"
echo "  - ${GOLDEN_DIR}/output.sarif"
echo "  - ${GOLDEN_DIR}/output.junit.xml"
