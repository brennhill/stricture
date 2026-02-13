# Stricture — Development Gates

> **Mandatory checkpoints that must exist BEFORE writing implementation code.**
>
> This document defines what artifacts, data, and verification infrastructure must be in place before any feature, rule, language adapter, or protocol support begins development. LLMs building Stricture must pass through these gates — no exceptions.

**Status:** Active
**Enforced by:** CI checks + pre-implementation review

---

## How to Use This Document

1. Before starting any development task, find the matching gate category below.
2. Verify ALL prerequisites listed for that gate are satisfied.
3. If any prerequisite is missing, create it FIRST — don't skip ahead to implementation.
4. After implementation, verify ALL exit criteria are met before considering the task complete.

**This document is the law.** If the gate says "golden file tests must exist," they must exist before the first line of implementation code is written.

---

## Gate 1: New Rule Implementation

**Applies to:** Any new rule (TQ, ARCH, CONV, CTR)

### Prerequisites (must exist BEFORE writing rule code)

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Product spec entry | `docs/product-spec.md` §15 | Rule ID, category, default severity, fixable, one-line description |
| 2 | Error catalog entry | `docs/error-catalog.yml` | Message template, why text, suggestion text, suppression syntax |
| 3 | Validation set cases | `docs/test-plan/validation-set/` | PERFECT example (0 violations) + at least 5 bug cases with expected rule ID |
| 4 | Golden file fixtures | `tests/fixtures/{rule-id}/` | `pass/` directory with clean code, `fail-*/` directories with violations, `expected.json` |
| 5 | Test file stub | `internal/rules/{category}/{rule}_test.go` | Table-driven test skeleton loading golden fixtures, all tests failing (TDD red phase) |
| 6 | Rule interface stub | `internal/rules/{category}/{rule}.go` | Struct implementing `model.Rule` interface with `Check()` returning nil (TDD red phase) |

### Exit Criteria (must pass AFTER implementation)

| # | Check | Command | Pass condition |
|---|-------|---------|----------------|
| 1 | Unit tests pass | `go test ./internal/rules/{category}/...` | 0 failures |
| 2 | Golden file match | `go test -run TestGolden ./internal/rules/{category}/...` | Output matches `expected.json` exactly |
| 3 | Validation PERFECT | `stricture` on validation PERFECT code | 0 violations from this rule |
| 4 | Validation bugs caught | `stricture` on validation bug cases | Expected violation triggered for each |
| 5 | Self-lint passes | `stricture .` on Stricture's own codebase | 0 violations (or rule excluded in `.stricture.yml`) |
| 6 | No regressions | `make test` | All pre-existing tests still pass |
| 7 | Error message quality | Manual review | Message follows catalog template, includes why + suggestion |
| 8 | File size | `wc -l` on implementation file | < 800 lines |

---

## Gate 2: New Language Adapter

**Applies to:** Any new language adapter (Go, TypeScript, Python, Java, or future)

### Prerequisites

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Language spec in product spec | `docs/product-spec.md` §7 | Adapter description, test framework detection, import resolution strategy |
| 2 | Parser selection documented | `docs/tech-spec.md` §6 | Which parser to use (go/parser, tree-sitter grammar), rationale |
| 3 | tree-sitter queries (if applicable) | `internal/adapter/{lang}/queries/` | `.scm` query files for imports, functions, types, classes, tests, assertions, mocks |
| 4 | Test file convention documented | `docs/tech-spec.md` or adapter README | How to detect test files (e.g., `_test.go`, `*.test.ts`, `test_*.py`) |
| 5 | UnifiedFileModel mapping table | Adapter README or tech spec | Explicit mapping: language construct → UFM field (e.g., "Python `def` → FuncModel") |
| 6 | Fixture files per construct | `tests/fixtures/adapter-{lang}/` | Real-world source files exercising: imports, exports, functions, types, classes, test cases, assertions, mocks, JSON tags, error exits |
| 7 | Validation set examples | `docs/test-plan/validation-set/` | At least 2 validation files using this language with PERFECT + bug cases |
| 8 | Cross-language pair (if not first adapter) | `docs/test-plan/validation-set/2x-*` | Contract pair with existing language showing json tag / naming mismatches |

### Exit Criteria

| # | Check | Command | Pass condition |
|---|-------|---------|----------------|
| 1 | Adapter unit tests | `go test ./internal/adapter/{lang}/...` | All UFM fields populated correctly for fixture files |
| 2 | All rules work with this language | `go test ./internal/rules/...` | No language-specific failures |
| 3 | Parse performance | `go test -bench ./internal/adapter/{lang}/...` | < 50ms per file (800 LOC) |
| 4 | Validation set passes | Validation script | PERFECT → 0 violations, bugs → correct violations |
| 5 | Extension detection correct | Adapter test | `Extensions()` returns correct list, `IsTestFile()` works |
| 6 | Import resolution works | Adapter test | `ResolveImport()` handles relative and absolute imports |

---

## Gate 3: New CTR (Contract) Rule

**Applies to:** Contract rules that validate cross-service or cross-file contracts. Inherits ALL of Gate 1, plus additional requirements.

### Additional Prerequisites (beyond Gate 1)

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Manifest schema support | `docs/tech-spec.md` §4 Phase 4 | The manifest field/constraint this rule validates is documented |
| 2 | Cross-language fixture pair | `tests/fixtures/{rule-id}/` | At least one fixture with different languages on each side (e.g., Go server + TS client) |
| 3 | Manifest fixture | `tests/fixtures/{rule-id}/manifest.yml` | Manifest file declaring the contract this rule validates |
| 4 | Mismatch taxonomy | Error catalog or validation set | At least 5 distinct mismatch types this rule catches (e.g., missing field, wrong type, wrong format, extra field, naming mismatch) |

### Additional Exit Criteria

| # | Check | Description |
|---|-------|-------------|
| 1 | Cross-language detection | Rule catches mismatches between Go↔TS, Python↔TS, Java↔TS |
| 2 | Manifest integration | Rule reads manifest constraints and validates code against them |
| 3 | Both directions tested | If applicable, validates both producer and consumer side |

---

## Gate 4: New Output Format

**Applies to:** New reporter implementations (text, JSON, SARIF, JUnit XML, or future)

### Prerequisites

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Format specification reference | Tech spec or reporter README | Link to the format specification (e.g., SARIF 2.1.0, JUnit XML schema) |
| 2 | Golden output file | `tests/golden/output.{format}` | Expected output for a known set of violations |
| 3 | Violation fixture set | `tests/golden/violations.json` | JSON file with exactly the violations to format (shared across all reporters) |

### Exit Criteria

| # | Check | Command | Pass condition |
|---|-------|---------|----------------|
| 1 | Golden file match | `go test -run TestGolden ./internal/reporter/...` | Output matches golden file byte-for-byte |
| 2 | Format validation | External validator | Output validates against format schema (SARIF validator, JUnit XSD) |
| 3 | Empty input | Reporter test | Handles 0 violations gracefully |
| 4 | Large input | Reporter test | 10,000 violations don't OOM or take > 1s |

---

## Gate 5: New CLI Command

**Applies to:** New commands (lint, audit, trace, init, list-rules, inspect, or future)

### Prerequisites

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Product spec entry | `docs/product-spec.md` | Command name, purpose, flags, example output |
| 2 | Help text | `cmd/stricture/{command}.go` | `--help` output drafted |
| 3 | Exit code contract | Tech spec or command doc | What exit codes this command returns and when |
| 4 | Integration test plan | Test stub | Scenarios covering: success, no files, invalid config, partial failure |

### Exit Criteria

| # | Check | Command | Pass condition |
|---|-------|---------|----------------|
| 1 | Help text works | `stricture {command} --help` | Prints help, exits 0 |
| 2 | Exit codes correct | Integration tests | Each documented exit code triggers correctly |
| 3 | Stdout is parseable | Integration test | JSON output is valid JSON, SARIF is valid SARIF |
| 4 | Stderr for errors | Integration test | Errors go to stderr, results go to stdout |

---

## Gate 6: Performance Change

**Applies to:** Any change to parsing, caching, file discovery, or concurrency

### Prerequisites

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Baseline benchmark | `tests/benchmark/baseline.json` | Current performance numbers for BM-01 through BM-07 |
| 2 | Benchmark repo generated | `scripts/generate-benchmark-repo.sh` | Synthetic repos at 500, 10K file sizes exist |

### Exit Criteria

| # | Check | Command | Pass condition |
|---|-------|---------|----------------|
| 1 | No regression | `make benchmark` | All targets within 20% of baseline |
| 2 | Memory within budget | `GOMEMLIMIT=500MiB stricture` on 10K repo | No OOM |
| 3 | Benchmark updated | `tests/benchmark/baseline.json` | New numbers recorded |

---

## Gate 7: Distribution / Release

**Applies to:** Any binary release, package publish, or installation method

### Prerequisites

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Version bumped | `VERSION` / `go.mod` | Semantic version updated |
| 2 | Changelog entry | `CHANGELOG.md` | What changed, migration notes if breaking |
| 3 | Cross-platform builds succeed | CI | linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 |
| 4 | All tests pass | CI | `make ci` green on all platforms |

### Exit Criteria

| # | Check | Description |
|---|-------|-------------|
| 1 | Binary works on all platforms | Download + run `stricture --version` on each OS/arch |
| 2 | Install methods work | Homebrew, npm wrapper, pip wrapper, `go install`, curl script |
| 3 | Self-lint passes | `stricture .` on Stricture's own source from the released binary |
| 4 | Validation set passes | Full validation set run with released binary |

---

## Gate 8: Future Language Support Planning

**Applies to:** Evaluating whether to add support for a new programming language (e.g., Rust, Kotlin, Swift, C#)

### Decision Prerequisites (must answer BEFORE committing to implementation)

| # | Question | Required answer |
|---|----------|-----------------|
| 1 | tree-sitter grammar availability | Does a mature, maintained tree-sitter grammar exist? Link to it. |
| 2 | Test framework identification | What are the top 2-3 test frameworks? How are tests detected? What's the assertion pattern? |
| 3 | Import/module system | How does import resolution work? Can we resolve imports from AST alone, or do we need build config? |
| 4 | JSON/serialization tags | Does the language have struct/field tags for JSON (like Go's `json:"..."` or Java's `@JsonProperty`)? How are they represented in the AST? |
| 5 | Naming conventions | What's the standard (camelCase, snake_case, PascalCase)? How does this interact with cross-language contracts? |
| 6 | Type system depth | Can tree-sitter extract type information? Do we need type inference? What's the accuracy impact? |
| 7 | Market demand | Is there evidence that users want this language? (GitHub issues, user requests, market size) |
| 8 | Existing coverage | Do existing tools (language-specific linters) already cover our rules for this language? What's the gap? |

### Required Artifacts Before Starting

| # | Artifact | Description |
|---|----------|-------------|
| 1 | Language evaluation document | `docs/languages/{lang}-evaluation.md` answering all 8 questions above |
| 2 | UFM mapping table | Explicit mapping of language constructs → UnifiedFileModel fields |
| 3 | tree-sitter query prototypes | `.scm` queries for imports, functions, types, tests, assertions — verified against real code |
| 4 | Validation set examples | At least 3 validation files (1 API integration, 1 cross-language pair, 1 test quality) |
| 5 | Performance budget | Projected parse time per file, memory per file, impact on cold start |

---

## Gate 9: Configuration Change

**Applies to:** Any change to `.stricture.yml` schema, `.stricture-manifest.yml` schema, or default values

### Prerequisites

| # | Artifact | Location | Description |
|---|----------|----------|-------------|
| 1 | Schema documented | `docs/product-spec.md` §5 or §13 | New field/option documented with type, default, valid values |
| 2 | Migration path | Changelog or migration doc | How existing configs handle the change (backwards compatible? new field? deprecated field?) |
| 3 | Validation test | Config test | Invalid values produce clear error messages |

### Exit Criteria

| # | Check | Description |
|---|-------|-------------|
| 1 | Backwards compatible | Existing valid configs still parse without warnings |
| 2 | Clear errors | Invalid new config values produce actionable error messages |
| 3 | Documentation updated | Product spec and/or README reflect the change |

---

## Master Gate Checklist

For any development task, identify which gates apply and verify all prerequisites:

```
Before writing any implementation code:

□ Gate 1 (Rule):        Product spec + error catalog + validation cases + golden fixtures + test stub
□ Gate 2 (Adapter):     Parser selection + UFM mapping + query files + fixture files + validation examples
□ Gate 3 (CTR Rule):    Gate 1 + manifest schema + cross-language fixtures + mismatch taxonomy
□ Gate 4 (Reporter):    Format spec + golden output + violation fixture set
□ Gate 5 (CLI Command): Product spec + help text + exit codes + integration test plan
□ Gate 6 (Performance): Baseline benchmark + benchmark repos
□ Gate 7 (Release):     Version bump + changelog + cross-platform builds
□ Gate 8 (New Lang):    Evaluation doc + UFM mapping + query prototypes + validation examples
□ Gate 9 (Config):      Schema documented + migration path + validation tests

After implementation:

□ All unit tests pass (go test -race ./...)
□ All golden file tests match
□ Validation set: PERFECT → 0 violations
□ Validation set: bugs → correct violations detected
□ Self-lint passes (stricture .)
□ No regressions (make test)
□ File size < 800 LOC
□ Error messages follow catalog template
□ Errors follow GO-ERROR-HANDLING.md (wrapped with %w, sentinels where needed)
□ Invariants preserved (docs/INVARIANTS.md):
    □ Output determinism (same input → byte-identical output)
    □ Parse safety (no panics on any input)
    □ Suppression correctness (suppress comments work per-rule)
    □ Fix idempotency (--fix twice = same result)
    □ Adapter parity (equivalent code → equivalent UFM)
    □ No silent degradation (skipped files reported)
    □ Config monotonicity (adding a rule only adds violations)
    □ Cache transparency (cached = uncached output)
□ If bug fix: regression fixture added per REGRESSION-PROTOCOL.md
□ If new API surface: stability tier documented per API-STABILITY.md
```

---

## Enforcement

### CI Enforcement (automated)

These gates are enforced by CI checks in `.github/workflows/ci.yml`:

1. **`golangci-lint`** — Aggressive config (`.golangci.yml`): errcheck, exhaustive enum switches, gocritic, gosec, race detection hints, cyclomatic complexity
2. **`go test -race`** — All tests with race detector enabled, on 3 OS (Linux, macOS, Windows)
3. **Coverage threshold** — 80% minimum overall, 100% for rule `Check()` methods
4. **Fuzz tests** — 30s fuzz runs for Go adapter, TS adapter, config loader, manifest parser (on main/next only)
5. **Integration tests** — Binary exit codes, JSON output validation, self-lint, golden file comparison
6. **File header check** — Every `.go` file starts with `// filename.go —`
7. **File size check** — No source file > 800 LOC
8. **Cross-platform builds** — linux/darwin/windows x amd64/arm64
9. **Benchmark regression** — On PRs, compares against baseline
10. **Validation set** — Health check + full run against validation fixtures
11. **JSON schema validation** — Output conforms to `tests/golden/schema.json`

### Pre-commit Enforcement (local)

Install with `pip install pre-commit && pre-commit install`. Checks:

1. `go fmt` + `go imports` — Formatting
2. `go vet` + `go build` — Compilation
3. File headers — `// filename.go —` pattern
4. File sizes — Max 800 LOC
5. Secret detection — gitleaks
6. No direct commits to main

### Review Enforcement (manual)

Before approving any PR, the reviewer (human or LLM) must verify:

1. All applicable gate prerequisites are satisfied
2. Error catalog entries exist for any new violations
3. Validation set has been updated if rules changed
4. Golden files have been updated if output changed
5. Error handling follows [`GO-ERROR-HANDLING.md`](GO-ERROR-HANDLING.md) — errors wrapped with `%w`, sentinel errors used
6. If fixing a bug: regression fixture added per [`REGRESSION-PROTOCOL.md`](REGRESSION-PROTOCOL.md)
7. No Tier 1 API changes without major version bump per [`API-STABILITY.md`](API-STABILITY.md)

### LLM Development Protocol

When an LLM is tasked with implementing a feature:

1. **Read this document first.** Identify which gates apply.
2. **Check prerequisites.** Read the artifacts listed. If any are missing, create them FIRST.
3. **Write tests first (TDD).** Tests must fail before implementation begins.
4. **Implement.** Write the minimum code to pass all tests.
5. **Verify exit criteria.** Run every check listed. Fix failures before marking complete.
6. **Report.** State which gate was followed, which checks passed, and any notes.

**Template response for LLM after completing a gate:**

```
Gate: {gate number and name}
Prerequisites verified: {list}
Implementation files: {list with line counts}
Exit criteria:
  □ Unit tests: PASS (X/X)
  □ Golden files: PASS
  □ Validation PERFECT: 0 violations
  □ Validation bugs: X/Y caught
  □ Self-lint: PASS
  □ Regressions: NONE
  □ File sizes: all < 800 LOC
Notes: {any deviations or concerns}
```
