#!/usr/bin/env bash
set -euo pipefail

# extract-fixtures.sh — Extracts code blocks from validation set markdown into test fixtures.
#
# Usage: ./scripts/extract-fixtures.sh [output-dir]
#
# Reads markdown files from docs/test-plan/validation-set/
# Extracts fenced code blocks tagged with language (```typescript, ```go, etc.)
# Organizes them into tests/fixtures/ by rule ID
#
# Each markdown file has:
#   - A "PERFECT" section with clean code (→ pass/ directory)
#   - "B01" through "B15" sections with bugs (→ fail-b{nn}/ directory)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
VALIDATION_DIR="${PROJECT_ROOT}/docs/test-plan/validation-set"
OUTPUT_DIR="${1:-${PROJECT_ROOT}/tests/fixtures}"

# Language extension mapping
declare -A LANG_EXT=(
    [typescript]=ts
    [go]=go
    [python]=py
    [java]=java
    [yaml]=yaml
    [yml]=yml
    [json]=json
    [javascript]=js
)

# Counters
TOTAL_FILES=0
TOTAL_SECTIONS=0
TOTAL_BLOCKS=0
TOTAL_RULES=0

echo "Extracting fixtures from validation set..."
echo "  Source: ${VALIDATION_DIR}"
echo "  Output: ${OUTPUT_DIR}"
echo ""

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Process each numbered markdown file
for md_file in "${VALIDATION_DIR}"/[0-9][0-9]-*.md; do
    [ -f "${md_file}" ] || continue

    TOTAL_FILES=$((TOTAL_FILES + 1))
    filename=$(basename "${md_file}")
    rule_num=$(echo "${filename}" | sed 's/^\([0-9][0-9]\)-.*/\1/')

    echo "Processing: ${filename}"

    # Extract sections using awk
    # AWK script processes markdown sections and code blocks
    awk -v output_dir="${OUTPUT_DIR}" \
        -v rule_num="${rule_num}" \
        -v total_sections_var=0 \
        -v total_blocks_var=0 \
        '
    BEGIN {
        current_section = ""
        in_code_block = 0
        code_lang = ""
        code_content = ""
        block_counter = 0
        section_has_blocks = 0
    }

    # Detect section headers (## PERFECT, ## B01, etc.)
    /^## (PERFECT|B[0-9][0-9])/ {
        # Save previous section if it had blocks
        if (current_section != "" && !section_has_blocks) {
            print "  WARNING: Section " current_section " has no code blocks" > "/dev/stderr"
        }

        # Extract section name - use gsub to remove prefix
        current_section = $2
        total_sections_var++
        section_has_blocks = 0
        block_counter = 0
        next
    }

    # Detect code block start
    /^```[a-z]+/ {
        if (current_section == "") next

        in_code_block = 1
        code_lang = substr($0, 4)
        code_content = ""
        next
    }

    # Detect code block end
    /^```$/ {
        if (!in_code_block) next

        in_code_block = 0
        section_has_blocks = 1
        block_counter++
        total_blocks_var++

        # Determine file extension
        ext = code_lang
        if (code_lang == "typescript") ext = "ts"
        else if (code_lang == "javascript") ext = "js"
        else if (code_lang == "python") ext = "py"
        else if (code_lang == "golang" || code_lang == "go") ext = "go"

        # Determine output directory based on section
        if (current_section == "PERFECT") {
            section_dir = output_dir "/" rule_num "-validation/pass"
        } else {
            bug_num = tolower(current_section)
            section_dir = output_dir "/" rule_num "-validation/fail-" bug_num
        }

        # Create directory
        system("mkdir -p \"" section_dir "\"")

        # Determine filename (use block counter if multiple blocks)
        if (block_counter == 1) {
            if (ext == "ts" || ext == "js") {
                output_file = section_dir "/source." ext
            } else if (ext == "go") {
                output_file = section_dir "/main.go"
            } else if (ext == "py") {
                output_file = section_dir "/main.py"
            } else if (ext == "java") {
                output_file = section_dir "/Main.java"
            } else {
                output_file = section_dir "/file." ext
            }
        } else {
            output_file = section_dir "/file" block_counter "." ext
        }

        # Write code content
        print code_content > output_file
        close(output_file)

        # Print progress
        print "    → " current_section " block " block_counter " (" ext ")" > "/dev/stderr"

        code_content = ""
        next
    }

    # Accumulate code block content
    in_code_block {
        if (code_content == "") {
            code_content = $0
        } else {
            code_content = code_content "\n" $0
        }
        next
    }

    END {
        if (current_section != "" && !section_has_blocks) {
            print "  WARNING: Section " current_section " has no code blocks" > "/dev/stderr"
        }
    }
    ' "${md_file}" || true

    section_count=$(grep -c '^## PERFECT\|^## B[0-9][0-9]' "${md_file}" || echo 0)
    TOTAL_SECTIONS=$((TOTAL_SECTIONS + section_count))
    block_count=$(grep -c '^```[a-z]' "${md_file}" || echo 0)
    TOTAL_BLOCKS=$((TOTAL_BLOCKS + block_count))
done

# Count unique rule directories created
TOTAL_RULES=$(find "${OUTPUT_DIR}" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')

# Create expected.json stubs for each rule directory
echo ""
echo "Creating expected.json stubs..."
for rule_dir in "${OUTPUT_DIR}"/*/; do
    [ -d "${rule_dir}" ] || continue

    expected_file="${rule_dir}/expected.json"
    cat > "${expected_file}" <<EOF
{
  "description": "Expected violations for $(basename "${rule_dir}") validation fixtures",
  "violations": []
}
EOF
    echo "  Created: ${expected_file}"
done

echo ""
echo "========================================="
echo "Extraction complete!"
echo "  Files processed:    ${TOTAL_FILES}"
echo "  Sections found:     ${TOTAL_SECTIONS}"
echo "  Code blocks:        ${TOTAL_BLOCKS}"
echo "  Rule directories:   ${TOTAL_RULES}"
echo "  Output directory:   ${OUTPUT_DIR}"
echo "========================================="
