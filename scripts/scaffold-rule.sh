#!/usr/bin/env bash
# scaffold-rule.sh — Generate boilerplate files for a new Stricture rule.
set -euo pipefail

# ---------------------------------------------------------------------------
# Colors
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
RESET='\033[0m'

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
info()  { printf "${GREEN}[created]${RESET} %s\n" "$1"; }
skip()  { printf "${YELLOW}[skipped]${RESET} %s — already exists\n" "$1"; }
error() { printf "${RED}[error]${RESET}  %s\n" "$1" >&2; }
bold()  { printf "${BOLD}%s${RESET}\n" "$1"; }

# ---------------------------------------------------------------------------
# Resolve project root (directory containing go.mod)
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ---------------------------------------------------------------------------
# Validate arguments
# ---------------------------------------------------------------------------
if [[ $# -ne 1 ]]; then
    error "Usage: $0 <RULE-ID>  (e.g., CONV-error-format)"
    exit 1
fi

RULE_ID="$1"

if ! [[ "${RULE_ID}" =~ ^(TQ|ARCH|CONV|CTR)-[a-z][a-z-]+$ ]]; then
    error "Invalid rule ID: '${RULE_ID}'"
    error "Must match ^(TQ|ARCH|CONV|CTR)-[a-z][a-z-]+\$"
    error "Examples: CONV-error-format, TQ-no-shallow-assertions, ARCH-no-circular-deps"
    exit 1
fi

# ---------------------------------------------------------------------------
# Parse rule ID into components
# ---------------------------------------------------------------------------
CATEGORY_UPPER="${RULE_ID%%-*}"           # CONV
RULE_NAME="${RULE_ID#*-}"                 # error-format
PACKAGE="${CATEGORY_UPPER,,}"             # conv (lowercase)

# Convert kebab-case to snake_case for Go file name
GO_FILE="${RULE_NAME//-/_}"               # error_format

# Convert kebab-case to PascalCase for Go type name
GO_TYPE=""
IFS='-' read -ra PARTS <<< "${RULE_NAME}"
for part in "${PARTS[@]}"; do
    GO_TYPE+="$(tr '[:lower:]' '[:upper:]' <<< "${part:0:1}")${part:1}"
done
# e.g., ErrorFormat

# ---------------------------------------------------------------------------
# File paths
# ---------------------------------------------------------------------------
RULE_DIR="${PROJECT_ROOT}/internal/rules/${PACKAGE}"
RULE_FILE="${RULE_DIR}/${GO_FILE}.go"
TEST_FILE="${RULE_DIR}/${GO_FILE}_test.go"
FIXTURE_PASS="${PROJECT_ROOT}/tests/fixtures/${RULE_ID}/pass"
FIXTURE_FAIL="${PROJECT_ROOT}/tests/fixtures/${RULE_ID}/fail-b01"

# ---------------------------------------------------------------------------
# Abort if any target file already exists
# ---------------------------------------------------------------------------
ABORT=false
for f in "${RULE_FILE}" "${TEST_FILE}"; do
    if [[ -f "${f}" ]]; then
        error "File already exists: ${f}"
        ABORT=true
    fi
done
if [[ "${ABORT}" == "true" ]]; then
    error "Aborting — will not overwrite existing files."
    exit 1
fi

# ---------------------------------------------------------------------------
# Try to read description and why from error-catalog.yml
# ---------------------------------------------------------------------------
DESCRIPTION="TODO: Add description"
WHY="TODO: Add why"
SEVERITY="error"
CATALOG="${PROJECT_ROOT}/docs/error-catalog.yml"

if [[ -f "${CATALOG}" ]]; then
    # Check if the rule ID exists in the catalog
    if grep -q "^  ${RULE_ID}:" "${CATALOG}" 2>/dev/null; then
        # Extract the why field (line after the rule ID block that starts with "    why:")
        extracted_why=$(awk "/^  ${RULE_ID}:/{found=1} found && /^    why:/{gsub(/^    why: *\"?/, \"\"); gsub(/\"$/, \"\"); print; exit}" "${CATALOG}")
        if [[ -n "${extracted_why}" ]]; then
            WHY="${extracted_why}"
        fi

        # Extract severity
        extracted_severity=$(awk "/^  ${RULE_ID}:/{found=1} found && /^    severity:/{gsub(/^    severity: */, \"\"); print; exit}" "${CATALOG}")
        if [[ -n "${extracted_severity}" ]]; then
            SEVERITY="${extracted_severity}"
        fi

        # Extract message as description
        extracted_msg=$(awk "/^  ${RULE_ID}:/{found=1} found && /^    message:/{gsub(/^    message: *\"?/, \"\"); gsub(/\"$/, \"\"); print; exit}" "${CATALOG}")
        if [[ -n "${extracted_msg}" ]]; then
            DESCRIPTION="${extracted_msg}"
        fi
    fi
fi

# ---------------------------------------------------------------------------
# Create directories
# ---------------------------------------------------------------------------
mkdir -p "${RULE_DIR}"
mkdir -p "${FIXTURE_PASS}"
mkdir -p "${FIXTURE_FAIL}"

# ---------------------------------------------------------------------------
# Generate rule file
# ---------------------------------------------------------------------------
cat > "${RULE_FILE}" << GOEOF
// ${GO_FILE}.go — ${RULE_ID}: ${DESCRIPTION}.
package ${PACKAGE}

import (
	"github.com/stricture/stricture/internal/model"
)

// ${GO_TYPE} implements the ${RULE_ID} rule.
type ${GO_TYPE} struct{}

func (r *${GO_TYPE}) ID() string                { return "${RULE_ID}" }
func (r *${GO_TYPE}) Category() string          { return "${PACKAGE}" }
func (r *${GO_TYPE}) Description() string       { return "${DESCRIPTION}" }
func (r *${GO_TYPE}) Why() string               { return "${WHY}" }
func (r *${GO_TYPE}) DefaultSeverity() string   { return "${SEVERITY}" }
func (r *${GO_TYPE}) NeedsProjectContext() bool { return false }

func (r *${GO_TYPE}) Check(file *model.UnifiedFileModel, context *model.ProjectContext, config model.RuleConfig) []model.Violation {
	// TODO: Implement — currently returns nil (tests will fail, TDD red phase)
	return nil
}
GOEOF
info "${RULE_FILE}"

# ---------------------------------------------------------------------------
# Generate test file
# ---------------------------------------------------------------------------
cat > "${TEST_FILE}" << GOEOF
// ${GO_FILE}_test.go — Tests for ${RULE_ID} rule.
package ${PACKAGE}

import (
	"testing"

	"github.com/stricture/stricture/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test${GO_TYPE}(t *testing.T) {
	rule := &${GO_TYPE}{}

	// Interface compliance
	assert.Equal(t, "${RULE_ID}", rule.ID())
	assert.Equal(t, "${PACKAGE}", rule.Category())
	assert.Equal(t, "${SEVERITY}", rule.DefaultSeverity())
	assert.NotEmpty(t, rule.Description())
	assert.NotEmpty(t, rule.Why())

	tests := []struct {
		name      string
		file      *model.UnifiedFileModel
		config    model.RuleConfig
		wantCount int
	}{
		{
			name: "valid file passes",
			file: &model.UnifiedFileModel{
				Path:     "/project/src/example.go",
				Language: "go",
				Source:   []byte("// example.go — Example.\npackage example\n"),
			},
			config:    model.RuleConfig{Severity: "${SEVERITY}", Options: map[string]interface{}{}},
			wantCount: 0,
		},
		{
			name: "invalid file fails",
			file: &model.UnifiedFileModel{
				Path:     "/project/src/example.go",
				Language: "go",
				Source:   []byte("package example\n"),
			},
			config:    model.RuleConfig{Severity: "${SEVERITY}", Options: map[string]interface{}{}},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.file, nil, tt.config)
			require.Len(t, violations, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, "${RULE_ID}", violations[0].RuleID)
				assert.NotEmpty(t, violations[0].Message)
			}
		})
	}
}
GOEOF
info "${TEST_FILE}"

# ---------------------------------------------------------------------------
# Create fixture .gitkeep files
# ---------------------------------------------------------------------------
touch "${FIXTURE_PASS}/.gitkeep"
info "${FIXTURE_PASS}/.gitkeep"

touch "${FIXTURE_FAIL}/.gitkeep"
info "${FIXTURE_FAIL}/.gitkeep"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
bold "Scaffolded rule: ${RULE_ID}"
echo "  Package:    ${PACKAGE}"
echo "  Type:       ${GO_TYPE}"
echo "  Rule file:  internal/rules/${PACKAGE}/${GO_FILE}.go"
echo "  Test file:  internal/rules/${PACKAGE}/${GO_FILE}_test.go"
echo "  Fixtures:   tests/fixtures/${RULE_ID}/pass/"
echo "              tests/fixtures/${RULE_ID}/fail-b01/"
echo ""
echo "Next steps:"
echo "  1. Edit the test file to define expected behavior (TDD red phase)"
echo "  2. Run: go test ./internal/rules/${PACKAGE}/... -run Test${GO_TYPE}"
echo "  3. Implement the Check() method to make tests pass (TDD green phase)"
echo "  4. Add fixture files for integration tests"
