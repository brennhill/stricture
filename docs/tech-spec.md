# Stricture — Technical Specification

> **Implementation blueprint for LLM-driven development.**
>
> This document defines HOW to build Stricture: module boundaries, interfaces, build order, data flow, and verification strategy. Every component has a clear interface contract and acceptance criteria so that LLMs can build modules independently, verify correctness through tests, and iterate to a working product.

**Version:** 0.1.0
**Status:** Tech Spec
**Companion:** [Product Spec](product-spec.md) (WHAT to build), [Development Gates](DEVELOPMENT-GATES.md) (WHEN to proceed)

---

## 1. Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Language | Go 1.22+ | Fast startup (~2ms), single binary, native concurrency, strong LLM support |
| Go parser | `go/parser` + `go/types` (stdlib) | Full type resolution for Go files, zero dependencies |
| TS/Python/Java parser | [go-tree-sitter](https://github.com/smacker/go-tree-sitter) | Unified parsing for all non-Go languages from a single binary |
| Config format | YAML (`.stricture.yml`) | Human-readable, widely supported, `gopkg.in/yaml.v3` |
| Manifest format | YAML (`.stricture-manifest.yml`) | Same parser as config |
| Output formats | Text, JSON, SARIF 2.1.0, JUnit XML | `encoding/json`, `encoding/xml` (stdlib) |
| Testing | `testing` + `testify` (test only) | Table-driven tests, `assert`/`require` for readability |
| Build | `make` + `go build` | Simple, no external build tools |
| Linting | `golangci-lint` | Self-dogfooding: Stricture lints itself |

**Zero production dependencies.** Only the Go standard library and tree-sitter C bindings. `testify` is test-only.

---

## 2. Project Structure

```
stricture/
├── cmd/
│   └── stricture/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── adapter/                    # Language adapters
│   │   ├── adapter.go              # LanguageAdapter interface
│   │   ├── goparser/               # Go adapter (go/parser + go/types)
│   │   │   ├── goparser.go
│   │   │   └── goparser_test.go
│   │   ├── typescript/             # TypeScript adapter (tree-sitter)
│   │   │   ├── typescript.go
│   │   │   └── typescript_test.go
│   │   ├── python/                 # Python adapter (tree-sitter)
│   │   └── java/                   # Java adapter (tree-sitter)
│   ├── config/                     # Config loader
│   │   ├── config.go               # Config types + loader
│   │   ├── defaults.go             # Default rule settings
│   │   └── config_test.go
│   ├── engine/                     # Rule engine orchestrator
│   │   ├── engine.go               # Core engine
│   │   ├── context.go              # ProjectContext builder
│   │   ├── graph.go                # Dependency graph (Tarjan's SCC)
│   │   └── engine_test.go
│   ├── manifest/                   # Manifest parser
│   │   ├── manifest.go             # Manifest types + parser
│   │   ├── strictness.go           # Strictness level calculation
│   │   └── manifest_test.go
│   ├── model/                      # Core data model
│   │   ├── file.go                 # UnifiedFileModel
│   │   ├── violation.go            # Violation types
│   │   └── rule.go                 # Rule interface
│   ├── reporter/                   # Output formatters
│   │   ├── reporter.go             # Reporter interface
│   │   ├── text.go                 # Human-readable text
│   │   ├── json.go                 # JSON output
│   │   ├── sarif.go                # SARIF 2.1.0
│   │   ├── junit.go                # JUnit XML
│   │   └── reporter_test.go
│   └── rules/                      # Rule implementations
│       ├── tq/                     # Test Quality rules
│       │   ├── no_shallow.go       # TQ-no-shallow-assertions
│       │   ├── no_shallow_test.go
│       │   ├── return_type.go      # TQ-return-type-verified
│       │   └── ...
│       ├── arch/                   # Architecture rules
│       │   ├── dependency_dir.go   # ARCH-dependency-direction
│       │   └── ...
│       ├── conv/                   # Convention rules
│       │   ├── file_naming.go      # CONV-file-naming
│       │   └── ...
│       └── ctr/                    # Contract rules
│           ├── request_shape.go    # CTR-request-shape
│           └── ...
├── pkg/                            # Public API (for Go plugins)
│   └── rule/
│       ├── definition.go           # Rule definition type (plugin API)
│       └── helpers.go              # Helper functions for rule authors
├── tests/
│   ├── fixtures/                   # Golden file test fixtures
│   │   ├── tq-no-shallow-assertions/
│   │   │   ├── pass/              # PERFECT code (0 violations)
│   │   │   ├── fail-shallow/      # Code with shallow assertions
│   │   │   └── expected.json      # Expected violations
│   │   └── ...
│   ├── golden/                     # Golden output files
│   │   ├── output-text.txt         # Expected text output
│   │   ├── output.json             # Expected JSON output
│   │   └── output.sarif            # Expected SARIF output
│   └── integration/                # Integration tests
│       ├── cli_test.go             # CLI flag parsing, exit codes
│       └── incremental_test.go     # --changed / --staged modes
├── scripts/
│   ├── extract-fixtures.sh         # Extract code from validation set markdown
│   ├── run-validation-set.sh       # Run strict against validation set
│   ├── validation-health-check.sh  # Validate the validation set itself
│   └── update-golden.sh            # Regenerate golden files
├── docs/
│   ├── product-spec.md
│   ├── tech-spec.md                # This file
│   ├── DEVELOPMENT-GATES.md
│   ├── error-catalog.yml           # Message templates for all rules
│   └── test-plan/                  # Test plans and validation set
├── .stricture.yml                  # Stricture lints itself
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 3. Core Interfaces

These interfaces define module boundaries. Every module communicates through these contracts. **Define interfaces BEFORE implementation.**

### 3.1 UnifiedFileModel

> **Product spec:** §7.1 UnifiedFileModel | **Source:** `internal/model/file.go`

```go
// internal/model/file.go

// UnifiedFileModel is the language-agnostic representation of a parsed source file.
// Language adapters produce this. Rules consume this.
type UnifiedFileModel struct {
    Path        string
    Language    string       // "go", "typescript", "python", "java"
    IsTestFile  bool
    Source      string
    LineCount   int
    Imports     []ImportDecl
    Exports     []ExportDecl
    Functions   []FuncModel
    Types       []TypeModel
    Classes     []ClassModel // TS, Java only
    TestCases   []TestCase   // Only populated for test files
    TestTargets []string     // Source files this test covers
    JSONTags    []JSONTag    // Go struct field → json tag mappings
}

type ImportDecl struct {
    Path       string // Import path (e.g., "fmt", "./user-service", "express")
    Alias      string // Import alias (empty if none)
    IsRelative bool   // True for relative imports (./foo, ../bar)
    Line       int
}

type ExportDecl struct {
    Name     string
    Kind     string // "function", "type", "const", "class", "default"
    Line     int
}

type FuncModel struct {
    Name       string
    Exported   bool
    Async      bool
    Receiver   string         // Go method receiver type (empty for functions)
    Params     []ParamModel
    ReturnType *TypeModel
    ErrorExits []ErrorExit
    StartLine  int
    EndLine    int
    Body       string         // Raw function body source (for pattern matching)
}

type ParamModel struct {
    Name     string
    Type     string
    Optional bool
}

type TypeModel struct {
    Name     string
    Kind     string       // "interface", "type", "struct", "enum", "class", "record"
    Fields   []FieldModel
    Exported bool
}

type FieldModel struct {
    Name     string
    Type     string
    Optional bool
    Nullable bool
    Nested   *TypeModel
    Depth    int
    JSONTag  string   // Go: `json:"field_name"`, Java: @JsonProperty value
    Line     int
}

type TestCase struct {
    Name       string
    Kind       string   // "positive", "negative", "unknown"
    TargetFns  []string
    Assertions []Assertion
    Mocks      []Mock
    StartLine  int
    EndLine    int
}

type Assertion struct {
    Expression     string
    TargetPath     string // "result.Data.Items[0].Name"
    Kind           string // "shallow", "value", "type", "structure", "error", "matcher"
    Depth          int
    ConstrainsType bool
    Line           int
}

type Mock struct {
    Target     string
    Scope      string // "module", "describe", "test"
    HasCleanup bool
    Line       int
}

type ErrorExit struct {
    Line            int
    ErrorIdentifier string
    GuardCondition  string
}

type JSONTag struct {
    StructName string
    FieldName  string
    TagValue   string // The json tag value (e.g., "user_id,omitempty")
    Line       int
}
```

### 3.2 LanguageAdapter Interface

> **Product spec:** §7 Language Adapters | **Source:** `internal/adapter/adapter.go`

```go
// internal/adapter/adapter.go

// LanguageAdapter parses source files into UnifiedFileModel.
// One adapter per supported language.
type LanguageAdapter interface {
    // Language returns the language identifier (e.g., "go", "typescript").
    Language() string

    // Extensions returns file extensions this adapter handles (e.g., [".go"]).
    Extensions() []string

    // Parse converts source bytes into a UnifiedFileModel.
    Parse(path string, source []byte, config AdapterConfig) (*model.UnifiedFileModel, error)

    // ResolveImport resolves an import path to an absolute file path.
    // Returns empty string if unresolvable (external dependency).
    ResolveImport(importPath string, fromFile string) string

    // IsTestFile returns true if the file is a test file for this language.
    IsTestFile(path string) bool
}

// AdapterConfig contains language-specific configuration.
type AdapterConfig struct {
    TestFrameworks  []string // e.g., ["jest", "vitest"] or ["testify"]
    ModuleRoot      string   // Project root for import resolution
    TSConfigPath    string   // TypeScript: path to tsconfig.json
    GoModPath       string   // Go: path to go.mod
}
```

### 3.3 Rule Interface

> **Product spec:** §6 Rule Categories | **Source:** `internal/model/rule.go`

```go
// internal/model/rule.go

// Rule is the interface all rules implement.
type Rule interface {
    // ID returns the rule identifier (e.g., "TQ-no-shallow-assertions").
    ID() string

    // Category returns the rule category (e.g., "test-quality", "architecture", "convention", "contract").
    Category() string

    // Description returns a one-line description of the rule.
    Description() string

    // Why returns a one-sentence explanation of why this rule exists.
    Why() string

    // Severity returns the default severity ("error" or "warn").
    DefaultSeverity() string

    // NeedsProjectContext returns true if this rule requires cross-file analysis.
    NeedsProjectContext() bool

    // Match returns true if this rule applies to the given file.
    Match(file *UnifiedFileModel, config RuleConfig) bool

    // Check runs the rule against a file and returns violations.
    // context may be nil for rules where NeedsProjectContext() returns false.
    Check(file *UnifiedFileModel, context *ProjectContext, config RuleConfig) []Violation
}

// RuleConfig contains per-rule configuration from .stricture.yml.
type RuleConfig struct {
    Severity string                 // "error", "warn", "off"
    Options  map[string]interface{} // Rule-specific options
}
```

### 3.4 Violation

> **Product spec:** §10 Output | **Source:** `internal/model/violation.go`

```go
// internal/model/violation.go

// Violation represents a single rule violation.
type Violation struct {
    RuleID     string      `json:"rule_id"`     // "TQ-no-shallow-assertions"
    Category   string      `json:"category"`    // "test-quality", "architecture", "convention", "contract"
    Severity   string      `json:"severity"`    // "error" or "warn"
    File       string      `json:"file"`        // Absolute file path
    Line       int         `json:"line"`        // 1-based line number
    Column     int         `json:"column"`      // 0-based column
    Message    string      `json:"message"`     // Human-readable message
    Why        string      `json:"why"`         // Why this rule exists (one sentence)
    Suggestion string      `json:"suggestion"`  // How to fix it
    Suppress   string      `json:"suppress"`    // Suppression comment syntax
    Fixable    bool        `json:"fixable"`     // Whether auto-fix is available
    Context    ViolationContext `json:"context"` // Additional context for tooling
}

// ViolationContext provides structured data about the violation.
type ViolationContext struct {
    FunctionName      string   // Function where violation occurs
    ReturnType        string   // Return type name (for TQ rules)
    FieldsAsserted    []string // Fields that are asserted (for TQ rules)
    FieldsMissing     []string // Fields not asserted (for TQ rules)
    ImportPath        string   // Offending import (for ARCH rules)
    ManifestField     string   // Manifest field name (for CTR rules)
    ManifestConstraint string  // Expected constraint (for CTR rules)
    CodeConstraint    string   // Actual constraint in code (for CTR rules)
}
```

### 3.5 ProjectContext

> **Product spec:** §4 Architecture | **Source:** `internal/model/context.go`

```go
// internal/model/context.go

// ProjectContext holds cross-file analysis state.
// Built once per run, shared across all rules.
type ProjectContext struct {
    // All parsed files indexed by absolute path.
    Files map[string]*model.UnifiedFileModel

    // Dependency graph: file path → list of imported file paths.
    DependencyGraph map[string][]string

    // Reverse dependency graph: file path → list of files that import it.
    ReverseDeps map[string][]string

    // Module boundaries: directory path → exported symbols.
    ModuleBoundaries map[string][]string

    // Test-to-source mapping: test file → source file(s).
    TestSourceMap map[string][]string

    // Manifest (nil if no manifest configured).
    Manifest *manifest.Manifest

    // Config
    Config *config.Config
}
```

### 3.6 Reporter Interface

> **Product spec:** §10 Output | **Source:** `internal/reporter/reporter.go`

```go
// internal/reporter/reporter.go

// Reporter formats violations for output.
type Reporter interface {
    // Format returns the format name (e.g., "text", "json", "sarif", "junit").
    Format() string

    // Report formats a list of violations and writes to the writer.
    Report(w io.Writer, violations []model.Violation, summary Summary) error
}

// Summary contains aggregated statistics.
type Summary struct {
    FilesChecked int   `json:"files_checked"`
    Errors       int   `json:"errors"`
    Warnings     int   `json:"warnings"`
    Fixable      int   `json:"fixable"`
    ElapsedMs    int64 `json:"elapsed_ms"`
}
```

---

## 4. Build Phases

Each phase produces a working binary that passes all tests for that phase. **Never start Phase N+1 until all Phase N tests pass.**

### Phase 1: Foundation (CLI + Config + Go Adapter + CONV Rules)

> **Product spec:** §5 Configuration, §6.3 CONV, §9 CLI | **Source:** `cmd/strict/`, `internal/config/`, `internal/rules/conv/`

**Goal:** `strict` can lint Go files for convention violations.

**Deliverables:**
1. CLI entry point with flag parsing (`cmd/strict/main.go`)
2. Config loader (reads `.stricture.yml`, merges with defaults)
3. Go language adapter (`go/parser` + `go/types` → `UnifiedFileModel`)
4. Two CONV rules: `CONV-file-naming`, `CONV-file-header`
5. Text reporter (human-readable output)
6. JSON reporter

**Test gate:**
```bash
# All must pass before Phase 2
make test-phase1
strict --version                    # Prints version
strict .                            # Runs on current directory
strict --format json .              # JSON output
strict --rule CONV-file-naming .    # Single rule
# Golden file tests for CONV-file-naming and CONV-file-header pass
```

**Files to create:**
- `cmd/strict/main.go`
- `internal/config/config.go`, `defaults.go`
- `internal/adapter/goparser/goparser.go`
- `internal/model/file.go`, `violation.go`, `rule.go`
- `internal/engine/engine.go`
- `internal/rules/conv/file_naming.go`, `file_header.go`
- `internal/reporter/text.go`, `json.go`
- Tests for all of the above

**Estimated size:** ~2,000 LOC

---

### Phase 2: TypeScript + Remaining CONV + ARCH Rules

> **Product spec:** §6.2 ARCH, §6.3 CONV, §7.3 TS Adapter | **Source:** `internal/adapter/typescript/`, `internal/rules/arch/`, `internal/rules/conv/`

**Goal:** `strict` can lint TypeScript files and detect architectural violations.

**Deliverables:**
1. TypeScript adapter (tree-sitter → `UnifiedFileModel`)
2. Remaining CONV rules: `CONV-error-format`, `CONV-export-naming`, `CONV-test-file-location`, `CONV-required-exports`
3. ProjectContext builder (dependency graph, module boundaries)
4. All ARCH rules: `dependency-direction`, `import-boundary`, `no-circular-deps`, `max-file-lines`, `layer-violation`, `module-boundary`
5. Dependency graph with Tarjan's SCC algorithm

**Test gate:**
```bash
make test-phase2
strict src/                         # Lints TypeScript files
strict --category ARCH .            # ARCH rules only
# Golden file tests for all 6 CONV and 6 ARCH rules pass
# Validation set fixtures for 30-express-layered-app pass
```

**Files to create:**
- `internal/adapter/typescript/typescript.go`
- `internal/engine/context.go`, `graph.go`
- `internal/rules/conv/error_format.go`, `export_naming.go`, `test_location.go`, `required_exports.go`
- `internal/rules/arch/dependency_dir.go`, `import_boundary.go`, `circular_deps.go`, `max_lines.go`, `layer_violation.go`, `module_boundary.go`
- Tests for all of the above

**Estimated size:** ~3,500 LOC (cumulative: ~5,500)

---

### Phase 3: TQ Rules (Test Quality)

> **Product spec:** §6.1 Test Quality | **Source:** `internal/rules/tq/`

**Goal:** `strict` can analyze test quality — the core differentiator.

**Deliverables:**
1. Test-to-source file mapping (convention, import analysis, config)
2. Assertion classifier (shallow vs. deep)
3. All TQ rules: `no-shallow-assertions`, `return-type-verified`, `schema-conformance`, `error-path-coverage`, `assertion-depth`, `boundary-tested`, `mock-scope`, `test-isolation`, `negative-cases`, `test-naming`

**Test gate:**
```bash
make test-phase3
strict --category TQ tests/        # TQ rules only
# Golden file tests for all 10 TQ rules pass
# Validation set 40-test-quality-patterns fixtures: 20/20 violations detected
# Validation set 41-ai-generated-test-patterns fixtures: all anti-patterns caught
```

**Files to create:**
- `internal/engine/test_mapping.go` (test-to-source resolution)
- `internal/engine/assertion_classifier.go`
- `internal/rules/tq/no_shallow.go`, `return_type.go`, `schema_conformance.go`, `error_path.go`, `assertion_depth.go`, `boundary.go`, `mock_scope.go`, `test_isolation.go`, `negative_cases.go`, `test_naming.go`
- Tests for all of the above

**Estimated size:** ~4,000 LOC (cumulative: ~9,500)

---

### Phase 4: CTR Rules + Manifest Parser

> **Product spec:** §6.4 Contract, §13 Manifest | **Source:** `internal/rules/ctr/`, `internal/manifest/`

**Goal:** `strict` can validate API contracts against a manifest.

**Deliverables:**
1. Manifest parser (`.stricture-manifest.yml`)
2. Strictness level calculator
3. All CTR rules: `request-shape`, `response-shape`, `status-code-handling`, `shared-type-sync`, `json-tag-match`, `dual-test`, `strictness-parity`, `manifest-conformance`
4. `strict audit` command

**Test gate:**
```bash
make test-phase4
strict --category CTR .             # CTR rules
strict audit                        # Audit command
# Golden file tests for all 8 CTR rules pass
# Cross-language fixtures (20-24) all detect mismatches
# Manifest parsing tests pass (edge cases, validation, versioning)
```

**Files to create:**
- `internal/manifest/manifest.go`, `strictness.go`, `parser.go`
- `internal/rules/ctr/request_shape.go`, `response_shape.go`, `status_code.go`, `shared_type_sync.go`, `json_tag_match.go`, `dual_test.go`, `strictness_parity.go`, `manifest_conformance.go`
- `cmd/strict/audit.go`
- Tests for all of the above

**Estimated size:** ~5,000 LOC (cumulative: ~14,500)

---

### Phase 5: Python + Java Adapters + Advanced Output

> **Product spec:** §7.4–7.5 Adapters, §10.3–10.4 SARIF/JUnit | **Source:** `internal/adapter/python/`, `internal/adapter/java/`, `internal/reporter/`

**Goal:** Multi-language support complete. All output formats working.

**Deliverables:**
1. Python adapter (tree-sitter)
2. Java adapter (tree-sitter)
3. SARIF 2.1.0 reporter
4. JUnit XML reporter
5. `--changed` / `--staged` modes (git integration)
6. AST caching (`.stricture-cache/`)
7. `strict trace` command (runtime trace validation)

**Test gate:**
```bash
make test-phase5
strict .                            # Multi-language project
strict --format sarif .             # SARIF output
strict --changed                    # Git-based incremental
# Python/Java validation fixtures pass
# Incremental analysis scenarios pass
# Performance benchmarks within targets
```

**Estimated size:** ~4,000 LOC (cumulative: ~18,500)

---

### Phase 6: Polish + Distribution

> **Product spec:** §8 Plugins, §11 Auto-Fix | **Source:** `internal/plugins/`, `internal/fix/`

**Goal:** Production-ready release.

**Deliverables:**
1. Auto-fix engine (`--fix` for fixable rules)
2. Inline suppression parsing
3. Plugin system (YAML custom rules + Go plugins)
4. Cross-platform binary distribution (GoReleaser)
5. `strict init` command
6. `strict list-rules` command
7. `strict inspect` command (debug)

**Test gate:**
```bash
make ci                                # Full CI suite
# All 30 rules pass golden file tests
# All validation set fixtures pass
# Performance benchmarks pass
# Cross-platform builds succeed
# Self-lint passes (strict lints itself)
```

**Estimated size:** ~3,500 LOC (cumulative: ~22,000)

---

## 5. Data Flow

> **Product spec:** §4 Architecture, §9 CLI | **Source:** `internal/engine/`, `cmd/strict/`

### 5.1 Main Lint Flow

```
CLI args
  │
  ▼
Config Loader ──→ merged Config
  │
  ▼
File Discovery ──→ []filePath (glob, --changed, --staged)
  │
  ▼
┌─────────────────────────────────────────┐
│ Per-file (goroutine pool):              │
│   1. Detect language from extension     │
│   2. Check AST cache (hash match?)      │
│   3. If miss: adapter.Parse() → UFM    │
│   4. Store in cache                     │
│   5. Collect UFM                        │
└─────────────────────────────────────────┘
  │
  ▼
Build ProjectContext
  ├── Dependency graph (from all UFM imports)
  ├── Module boundary map
  ├── Test-to-source mapping
  └── Load manifest (if configured)
  │
  ▼
┌─────────────────────────────────────────┐
│ Rule execution (sequential by category):│
│   1. CONV rules (per-file, fast)        │
│   2. ARCH rules (need ProjectContext)   │
│   3. TQ rules (need test+source pairs)  │
│   4. CTR rules (need manifest+context)  │
│                                         │
│ Within each category: parallel per-file │
│ Skip files with suppression comments    │
└─────────────────────────────────────────┘
  │
  ▼
Collect []Violation
  │
  ▼
Reporter.Report() ──→ stdout or --output file
  │
  ▼
Exit code: 0 (clean) | 1 (violations) | 2 (config/parse error)
```

### 5.2 Concurrency Model

```
main goroutine
  │
  ├── Parse files: worker pool (GOMAXPROCS workers)
  │     Each worker: read file → detect language → parse → return UFM
  │     Channel-based: filePaths chan → results chan
  │
  ├── Build ProjectContext: single goroutine (needs all UFMs)
  │
  ├── Run rules: worker pool per category
  │     CONV: 1 goroutine per file (no cross-file deps)
  │     ARCH: 1 goroutine per file (reads ProjectContext, no writes)
  │     TQ:   1 goroutine per test file (reads source file + ProjectContext)
  │     CTR:  1 goroutine per contract pair (reads manifest + both files)
  │
  └── Collect violations: single goroutine (merge from channels)
```

### 5.3 Caching Strategy

```
.stricture-cache/
  ├── version              # Cache format version (invalidate on upgrade)
  ├── go/                  # Per-language cache directories
  │   ├── abc123.gob       # gob-encoded UnifiedFileModel
  │   └── ...
  ├── typescript/
  └── ...

Cache key = SHA256(file content + adapter version)
Cache hit = skip Parse(), deserialize from gob
Cache miss = Parse(), serialize to gob
Cache eviction = delete .stricture-cache/ (or --no-cache flag)
```

---

## 6. Parsing Strategy

> **Product spec:** §7 Language Adapters | **Source:** `internal/adapter/`

### 6.1 Go Adapter (Native)

Uses `go/parser.ParseFile()` + `go/types.Config.Check()` from stdlib:

```
.go file
  │
  ▼
go/parser.ParseFile() → *ast.File
  │
  ▼
go/types.Config{}.Check() → *types.Package (full type resolution)
  │
  ▼
Walk AST:
  ├── Extract imports (ast.ImportSpec)
  ├── Extract functions (ast.FuncDecl) with receivers, params, returns
  ├── Extract types (ast.TypeSpec → struct, interface)
  ├── Extract struct json tags (reflect.StructTag)
  ├── Detect test functions (Test*, Benchmark*, Example*)
  ├── Detect test assertions (t.Fatal, t.Error, assert.*, require.*)
  └── Detect error returns (multiple return with error type)
  │
  ▼
UnifiedFileModel
```

### 6.2 tree-sitter Adapters (TS, Python, Java)

Uses go-tree-sitter with language-specific grammars:

```
source file
  │
  ▼
tree-sitter.Parse() → *sitter.Tree (CST — Concrete Syntax Tree)
  │
  ▼
tree-sitter queries (S-expression patterns):
  ├── Import query: (import_statement) / (import_declaration)
  ├── Function query: (function_declaration) / (method_definition)
  ├── Type query: (interface_declaration) / (type_alias_declaration)
  ├── Class query: (class_declaration)
  ├── Export query: (export_statement)
  ├── Test query: (call_expression function: (identifier) @fn (#match? @fn "^(describe|it|test)$"))
  ├── Assertion query: (call_expression function: (member_expression) @fn (#match? @fn "expect|assert"))
  └── Mock query: (call_expression function: (member_expression) @fn (#match? @fn "mock|spy|stub"))
  │
  ▼
UnifiedFileModel
```

**Tree-sitter query library:** Each adapter ships with a set of `.scm` query files (tree-sitter query language). These are embedded in the binary via `embed.FS`.

---

## 7. Rule Implementation Pattern

> **Product spec:** §6 Rule Categories | **Source:** `internal/rules/`

Every rule follows the same pattern. This consistency enables LLMs to implement rules by following a template.

### 7.1 Rule Template

```go
// internal/rules/conv/file_naming.go

package conv

import "github.com/stricture/stricture/internal/model"

// FileNaming enforces file naming conventions.
type FileNaming struct{}

func (r *FileNaming) ID() string          { return "CONV-file-naming" }
func (r *FileNaming) Category() string    { return "convention" }
func (r *FileNaming) Description() string { return "Enforce file naming convention" }
func (r *FileNaming) Why() string {
    return "Consistent naming makes files predictable and searchable."
}
func (r *FileNaming) DefaultSeverity() string    { return "error" }
func (r *FileNaming) NeedsProjectContext() bool   { return false }

func (r *FileNaming) Match(file *model.UnifiedFileModel, config model.RuleConfig) bool {
    // Apply to all files
    return true
}

func (r *FileNaming) Check(
    file *model.UnifiedFileModel,
    context *model.ProjectContext,
    config model.RuleConfig,
) []model.Violation {
    // Implementation here
    // Use error-catalog.yml message template
    // Return violations or empty slice
}
```

### 7.2 Rule Test Template

```go
// internal/rules/conv/file_naming_test.go

package conv

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stricture/stricture/internal/model"
)

func TestFileNaming(t *testing.T) {
    rule := &FileNaming{}

    tests := []struct {
        name       string
        file       *model.UnifiedFileModel
        config     model.RuleConfig
        wantCount  int
        wantRuleID string
        wantLine   int
    }{
        {
            name: "kebab-case passes",
            file: &model.UnifiedFileModel{
                Path:     "/project/src/user-service.ts",
                Language: "typescript",
            },
            config:    model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}},
            wantCount: 0,
        },
        {
            name: "PascalCase violates kebab-case",
            file: &model.UnifiedFileModel{
                Path:     "/project/src/UserService.ts",
                Language: "typescript",
            },
            config:    model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}},
            wantCount: 1,
            wantRuleID: "CONV-file-naming",
        },
        // ... more cases from validation set
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            violations := rule.Check(tt.file, nil, tt.config)
            assert.Len(t, violations, tt.wantCount)
            if tt.wantCount > 0 {
                assert.Equal(t, tt.wantRuleID, violations[0].RuleID)
            }
        })
    }
}
```

---

## 8. Assertion Classification Algorithm

> **Product spec:** §6.1 Test Quality | **Source:** `internal/rules/tq/`

The assertion classifier is the core of TQ rules. It determines whether an assertion is "shallow" (checks existence only) or "deep" (checks actual values).

### Classification Rules

| Pattern | Kind | Depth | Example |
|---------|------|-------|---------|
| `toBeDefined()` | shallow | 0 | `expect(result).toBeDefined()` |
| `toBeTruthy()` | shallow | 0 | `expect(result).toBeTruthy()` |
| `not.toBeNull()` | shallow | 0 | `expect(result).not.toBeNull()` |
| `toBeInstanceOf(X)` | shallow | 0 | `expect(result).toBeInstanceOf(Object)` |
| `!= nil` (Go) | shallow | 0 | `if result != nil { ... }` |
| `toBe(value)` | value | 0 | `expect(result).toBe(42)` |
| `toEqual(obj)` | structure | max | `expect(result).toEqual({id: 1})` |
| `field.toBe(value)` | value | 1 | `expect(result.name).toBe("Alice")` |
| `nested.field.toBe()` | value | 2+ | `expect(result.data.id).toBe(1)` |
| `toHaveLength(n)` | value | 0 | `expect(items).toHaveLength(3)` |
| `toThrow(Error)` | error | 0 | `expect(fn).toThrow(NotFoundError)` |
| `toMatch(regex)` | matcher | 0 | `expect(id).toMatch(/^ch_/)` |

### Algorithm

```
Input: assertion AST node
Output: Assertion { Kind, Depth, ConstrainsType }

1. Extract the assertion method name (toBe, toEqual, toBeDefined, etc.)
2. If method is in SHALLOW_METHODS: return Kind="shallow", Depth=0
3. Extract the target expression chain (result.data.items[0].name)
4. Count property access depth (dots + bracket accesses)
5. If method compares against a literal value: Kind="value"
6. If method compares against an object literal: Kind="structure"
7. If method checks error type/message: Kind="error"
8. If method uses regex/matcher: Kind="matcher"
9. ConstrainsType = (method is not shallow AND expected value has a specific type)
10. Return Assertion { Kind, Depth, ConstrainsType }
```

---

## 9. Performance Targets & Measurement

| Metric | Target | How to measure | Phase |
|--------|--------|----------------|-------|
| Cold start (500 files) | < 3s | `time strict` on benchmark repo | Phase 2 |
| Cached run (500 files) | < 1s | Second `time strict` run | Phase 2 |
| Incremental (20 files) | < 2s | `time strict --changed` | Phase 5 |
| Per-file | < 50ms | `strict --verbose` timing output | Phase 1 |
| Memory (10K files) | < 500MB | `GOMEMLIMIT=500MiB strict` | Phase 5 |
| Binary size | < 50MB | `ls -la strict` (with tree-sitter) | Phase 2 |

**Benchmark infrastructure:**
- `scripts/generate-benchmark-repo.sh` generates synthetic repos at various sizes
- `make benchmark` runs all benchmarks and compares against baseline
- CI posts benchmark results as PR comments on regressions > 20%

---

## 10. Distribution

### Cross-Platform Builds

```
GOOS/GOARCH matrix:
  - linux/amd64
  - linux/arm64
  - darwin/amd64
  - darwin/arm64
  - windows/amd64

Build with CGO_ENABLED=1 (required for tree-sitter C bindings)
Use zig cc for cross-compilation (handles C cross-compile cleanly)
```

### Installation Methods

```bash
# Binary download
curl -fsSL https://github.com/stricture/stricture/releases/latest/download/strict-$(uname -s)-$(uname -m) -o /usr/local/bin/strict

# Homebrew
brew install stricture/tap/stricture

# Go install (builds from source, requires C compiler)
go install github.com/stricture/stricture/cmd/strict@latest

# Docker
docker run --rm -v "$PWD":/src stricture/stricture

# curl (downloads binary)
curl -fsSL https://get.stricture.dev | sh
```

---

## 11. Self-Verification

Stricture lints itself. The project ships with a `.stricture.yml` that enforces:

```yaml
# .stricture.yml — Stricture lints itself
rules:
  CONV-file-naming: [error, { style: "snake_case" }]  # Go convention
  CONV-file-header: [error, { pattern: "// {filename} — {purpose}" }]
  CONV-error-format: [error, { pattern: "{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}" }]
  ARCH-max-file-lines: [error, { max: 800 }]
  ARCH-no-circular-deps: error
  ARCH-dependency-direction: [error, {
    layers: ["cmd", "internal/engine", "internal/rules", "internal/adapter", "internal/model"]
  }]
```

**If `strict .` fails on the strict codebase, the CI build fails.** This is the ultimate verification: the tool validates its own code quality.
