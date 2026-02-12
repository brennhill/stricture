# CLI & Tool Tests

Tests for all Stricture CLI commands, flags, and tool-specific behavior. Each test specifies input, expected output, and verification criteria. All CLI invocations assume `stricture` is on PATH (or invoked via `npx stricture`).

**Product spec reference:** Section 9 (CLI Interface), Section 10 (Output & Reporting), Section 11 (Auto-Fix System), Section 13 (Cross-Service Contracts & Stricture Manifest).

**Total rules:** 34 (10 TQ + 6 ARCH + 6 CONV + 8 CTR + 2 reserved for v0.2).

---

## 1. `stricture lint` (Default Command)

The default command when no subcommand is given. `stricture [options] [paths...]` is equivalent to `stricture lint [options] [paths...]`.

### 1.1 File Targeting

**Test ID: CLI-LINT-TARGET-01: Glob pattern matches files**

- **Input:**
```bash
stricture "src/**/*.ts"
```
- **Setup:** Directory `src/` with `a.ts`, `b.ts`, `sub/c.ts`, and `d.js`.
- **Expected:** Lints `a.ts`, `b.ts`, `sub/c.ts`. Skips `d.js`. Summary line shows 3 files checked.
- **Verification:** Output contains `3 files checked`. No mention of `d.js`.

**Test ID: CLI-LINT-TARGET-02: Multiple glob patterns**

- **Input:**
```bash
stricture "src/**/*.ts" "lib/**/*.go"
```
- **Setup:** `src/a.ts`, `lib/b.go`, `lib/c.py`.
- **Expected:** Lints `a.ts` and `b.go`. Summary shows 2 files checked.
- **Verification:** Both files appear in output if they have violations. `c.py` is not mentioned.

**Test ID: CLI-LINT-TARGET-03: --changed flag (git diff)**

- **Input:**
```bash
stricture --changed
```
- **Setup:** Git repo with `main` branch. Current branch has 2 modified files (`src/a.ts`, `src/b.ts`) and 1 unmodified (`src/c.ts`).
- **Expected:** Only `a.ts` and `b.ts` are linted. Summary shows 2 files checked.
- **Verification:** `c.ts` does not appear anywhere in output.

**Test ID: CLI-LINT-TARGET-04: --staged flag (pre-commit)**

- **Input:**
```bash
stricture --staged
```
- **Setup:** 3 files staged (`git add`), 2 files modified but unstaged.
- **Expected:** Only the 3 staged files are linted. Summary shows 3 files checked.
- **Verification:** Unstaged files excluded. Works correctly as a pre-commit hook.

**Test ID: CLI-LINT-TARGET-05: --ext filtering**

- **Input:**
```bash
stricture --ext .go src/
```
- **Setup:** `src/` contains `a.go`, `b.go`, `c.ts`, `d.js`.
- **Expected:** Only `a.go` and `b.go` linted. Summary shows 2 files checked.
- **Verification:** TypeScript and JavaScript files excluded.

**Test ID: CLI-LINT-TARGET-06: No files matched exits 0 with message**

- **Input:**
```bash
stricture "src/**/*.rs"
```
- **Setup:** No `.rs` files exist in `src/`.
- **Expected:** Exit code 0. Output contains a message like `No files matched the given patterns.`
- **Verification:** `echo $?` returns 0. Stderr or stdout contains informational message, not an error.

**Test ID: CLI-LINT-TARGET-07: Hidden files are skipped by default**

- **Input:**
```bash
stricture .
```
- **Setup:** Directory contains `.hidden.ts` and `visible.ts`.
- **Expected:** Only `visible.ts` is linted. `.hidden.ts` is skipped.
- **Verification:** Hidden file does not appear in output or summary count.

**Test ID: CLI-LINT-TARGET-08: Symlinks are followed**

- **Input:**
```bash
stricture src/
```
- **Setup:** `src/link.ts` is a symlink to `../lib/real.ts`.
- **Expected:** The symlinked file is linted. Violations report the symlink path (`src/link.ts`).
- **Verification:** Symlinked file appears in results under its symlink path.

**Test ID: CLI-LINT-TARGET-09: Binary files are skipped gracefully**

- **Input:**
```bash
stricture .
```
- **Setup:** Directory contains `image.png`, `data.bin`, and `valid.ts`.
- **Expected:** Binary files skipped without error. Only `valid.ts` linted.
- **Verification:** No parse error for binary files. No crash. Exit code reflects only `valid.ts` results.

**Test ID: CLI-LINT-TARGET-10: Directory path expands to all supported files**

- **Input:**
```bash
stricture src/
```
- **Setup:** `src/` contains `a.ts`, `b.go`, `c.py`, `d.rs`, `e.json`.
- **Expected:** Lints `a.ts` and `b.go` (supported languages). Skips unsupported languages. Skips non-code files.
- **Verification:** Summary shows 2 files checked.

**Test ID: CLI-LINT-TARGET-11: Default target is current directory**

- **Input:**
```bash
stricture
```
- **Setup:** Current directory contains `src/a.ts` and `lib/b.go`.
- **Expected:** Recursively discovers and lints all supported files from `.`.
- **Verification:** Both `src/a.ts` and `lib/b.go` appear in results.

**Test ID: CLI-LINT-TARGET-12: Gitignored files are skipped**

- **Input:**
```bash
stricture .
```
- **Setup:** `.gitignore` contains `dist/`. Directory has `dist/bundle.ts` and `src/app.ts`.
- **Expected:** `dist/bundle.ts` is skipped. Only `src/app.ts` linted.
- **Verification:** Gitignored files do not appear in output or summary count.

**Test ID: CLI-LINT-TARGET-13: Config ignore patterns respected**

- **Input:**
```bash
stricture .
```
- **Setup:** `.stricture.yml` has `settings.ignore: ["vendor/**", "*.generated.ts"]`. Directory has `vendor/dep.go` and `src/types.generated.ts`.
- **Expected:** Both ignored. Not counted in summary.
- **Verification:** Neither file appears in output.

---

### 1.2 Flag Parsing

**Test ID: CLI-LINT-FLAG-01: --rule filters to a single rule**

- **Input:**
```bash
stricture --rule TQ-no-shallow-assertions src/
```
- **Setup:** Files with violations for multiple rules (TQ-no-shallow-assertions, CONV-file-header, ARCH-max-file-lines).
- **Expected:** Only TQ-no-shallow-assertions violations reported. Other rule violations suppressed.
- **Verification:** Output contains only `TQ-no-shallow-assertions` rule IDs. No other rule IDs present.

**Test ID: CLI-LINT-FLAG-02: --rule repeated for multiple rules**

- **Input:**
```bash
stricture --rule TQ-no-shallow-assertions --rule CONV-file-header src/
```
- **Expected:** Only violations from those two rules reported.
- **Verification:** Output contains only `TQ-no-shallow-assertions` and `CONV-file-header` rule IDs.

**Test ID: CLI-LINT-FLAG-03: --category filters by category**

- **Input:**
```bash
stricture --category TQ src/
```
- **Setup:** Files with TQ, ARCH, CONV, and CTR violations.
- **Expected:** Only TQ-* violations reported.
- **Verification:** All rule IDs in output start with `TQ-`. No `ARCH-`, `CONV-`, or `CTR-` rules.

**Test ID: CLI-LINT-FLAG-04: --severity error filters out warnings**

- **Input:**
```bash
stricture --severity error src/
```
- **Setup:** Config has some rules at `warn` and some at `error`.
- **Expected:** Only error-level violations shown. Warning-level violations suppressed.
- **Verification:** All violations in output have severity `error`.

**Test ID: CLI-LINT-FLAG-05: --quiet suppresses warnings, shows errors**

- **Input:**
```bash
stricture --quiet src/
```
- **Setup:** Files with both error and warning violations.
- **Expected:** Only error-level violations in output. Warnings suppressed. Summary line still shows total.
- **Verification:** No `warn` or `warning` severity in output body.

**Test ID: CLI-LINT-FLAG-06: --verbose shows timing and debug info**

- **Input:**
```bash
stricture --verbose src/
```
- **Expected:** Output includes per-rule timing info (e.g., `TQ-no-shallow-assertions: 12ms`). May include AST parse time, file count, cache hit/miss stats.
- **Verification:** Output contains timing numbers. More lines than non-verbose output.

**Test ID: CLI-LINT-FLAG-07: --format json produces valid JSON**

- **Input:**
```bash
stricture --format json src/
```
- **Expected:** Output is valid JSON matching the schema from product spec section 10.2. Contains `version`, `timestamp`, `summary`, and `violations` array.
- **Verification:** `stricture --format json src/ | jq .` succeeds without error. `summary.filesChecked` is a number. `violations` is an array.

**Test ID: CLI-LINT-FLAG-08: --format sarif produces valid SARIF 2.1.0**

- **Input:**
```bash
stricture --format sarif src/
```
- **Expected:** Output is valid SARIF 2.1.0 JSON. Contains `$schema`, `version: "2.1.0"`, `runs` array with tool metadata, rules, and results.
- **Verification:** Parse as JSON. Validate `runs[0].tool.driver.name` is `"stricture"`. Results have `physicalLocation` with file/line/column.

**Test ID: CLI-LINT-FLAG-09: --format junit produces valid JUnit XML**

- **Input:**
```bash
stricture --format junit src/
```
- **Expected:** Output is valid JUnit XML. Each rule is a `<testsuite>`, each checked file is a `<testcase>`, violations are `<failure>` elements.
- **Verification:** XML is well-formed. Contains `<testsuites>`, `<testsuite>`, `<testcase>`, and `<failure>` tags.

**Test ID: CLI-LINT-FLAG-10: --format text (default) produces colored output**

- **Input:**
```bash
stricture src/
```
- **Expected (TTY):** Output contains ANSI color codes for rule IDs, severity, file paths, and suggestions.
- **Expected (pipe):** No ANSI codes when output is piped (e.g., `stricture src/ | cat`).
- **Verification:** Check raw bytes for ANSI escape sequences (`\033[`) in TTY mode.

**Test ID: CLI-LINT-FLAG-11: --output writes to file**

- **Input:**
```bash
stricture --format json --output results.json src/
```
- **Expected:** No output to stdout. `results.json` created with valid JSON content.
- **Verification:** `test -f results.json` succeeds. `jq . results.json` succeeds. Stdout is empty.

**Test ID: CLI-LINT-FLAG-12: --output to stdout when not specified**

- **Input:**
```bash
stricture --format json src/
```
- **Expected:** JSON written to stdout.
- **Verification:** `stricture --format json src/ > /dev/null` produces no stderr. Content appears on stdout.

**Test ID: CLI-LINT-FLAG-13: --color forces color on in pipe**

- **Input:**
```bash
stricture --color src/ | cat
```
- **Expected:** Output contains ANSI color codes even when piped.
- **Verification:** Raw output contains `\033[` sequences.

**Test ID: CLI-LINT-FLAG-14: --no-color disables color in TTY**

- **Input:**
```bash
stricture --no-color src/
```
- **Expected:** No ANSI color codes in output, even in a TTY.
- **Verification:** Raw output does not contain `\033[` sequences.

**Test ID: CLI-LINT-FLAG-15: --config uses custom config path**

- **Input:**
```bash
stricture --config configs/strict.yml src/
```
- **Setup:** `configs/strict.yml` enables all rules at error level. Project root `.stricture.yml` has most rules at warn.
- **Expected:** Uses `configs/strict.yml`. All violations are error-level.
- **Verification:** Violations match the config from `configs/strict.yml`, not root config.

**Test ID: CLI-LINT-FLAG-16: --no-config ignores .stricture.yml**

- **Input:**
```bash
stricture --no-config src/
```
- **Setup:** `.stricture.yml` disables TQ-no-shallow-assertions. File has a shallow assertion.
- **Expected:** TQ-no-shallow-assertions violation reported (defaults used, rule is error by default).
- **Verification:** Rule fires despite being disabled in config file.

**Test ID: CLI-LINT-FLAG-17: Invalid flag produces helpful error**

- **Input:**
```bash
stricture --nonexistent-flag src/
```
- **Expected:** Exit code 2. Error message identifies the unknown flag and suggests similar valid flags (if applicable).
- **Verification:** `echo $?` returns 2. Stderr contains `unknown flag` or `unrecognized option`.

**Test ID: CLI-LINT-FLAG-18: --rule with invalid rule ID**

- **Input:**
```bash
stricture --rule FAKE-nonexistent src/
```
- **Expected:** Exit code 2 or warning. Error message states the rule does not exist and lists valid rule IDs.
- **Verification:** Error message contains `FAKE-nonexistent` and suggests valid alternatives.

**Test ID: CLI-LINT-FLAG-19: --category with invalid category**

- **Input:**
```bash
stricture --category NONEXISTENT src/
```
- **Expected:** Exit code 2. Error message lists valid categories: TQ, ARCH, CONV, CTR.
- **Verification:** Stderr contains valid category names.

**Test ID: CLI-LINT-FLAG-20: --format with invalid format**

- **Input:**
```bash
stricture --format yaml src/
```
- **Expected:** Exit code 2. Error message lists valid formats: text, json, sarif, junit.
- **Verification:** Stderr contains `yaml` and the valid format names.

**Test ID: CLI-LINT-FLAG-21: Conflicting flags --quiet and --verbose**

- **Input:**
```bash
stricture --quiet --verbose src/
```
- **Expected:** Either exit code 2 with "conflicting flags" error, or one takes precedence (document which). Recommended: error.
- **Verification:** Behavior is deterministic and documented.

**Test ID: CLI-LINT-FLAG-22: --version prints version and exits**

- **Input:**
```bash
stricture --version
```
- **Expected:** Prints version string (e.g., `stricture 0.1.0`) to stdout. Exit code 0. No linting performed.
- **Verification:** Output matches semver pattern `\d+\.\d+\.\d+`. Exit code 0.

**Test ID: CLI-LINT-FLAG-23: --help prints usage and exits**

- **Input:**
```bash
stricture --help
```
- **Expected:** Prints help text including all commands (lint, fix, audit, trace, init, inspect, list-rules) and all options. Exit code 0.
- **Verification:** Output contains all 7 commands and all option flags from section 9.2.

---

### 1.3 Exit Codes

**Test ID: CLI-LINT-EXIT-01: Exit 0 when no errors (warnings OK)**

- **Input:**
```bash
stricture src/
```
- **Setup:** Files have warning-level violations only (all error-level rules pass).
- **Expected:** Exit code 0.
- **Verification:** `echo $?` returns 0. Output may contain warnings but no errors.

**Test ID: CLI-LINT-EXIT-02: Exit 1 when errors exist**

- **Input:**
```bash
stricture src/
```
- **Setup:** At least one file has an error-level violation.
- **Expected:** Exit code 1.
- **Verification:** `echo $?` returns 1.

**Test ID: CLI-LINT-EXIT-03: Exit 2 on config error**

- **Input:**
```bash
stricture --config invalid.yml src/
```
- **Setup:** `invalid.yml` has invalid YAML syntax.
- **Expected:** Exit code 2. Error message includes the config file path and describes the parse error.
- **Verification:** `echo $?` returns 2. Stderr mentions the file path.

**Test ID: CLI-LINT-EXIT-04: Exit 2 on unparseable source file (only file targeted)**

- **Input:**
```bash
stricture broken.ts
```
- **Setup:** `broken.ts` has severe syntax errors that prevent parsing.
- **Expected:** Exit code 2 if it is the only file and no other violations exist. (Alternative: exit 1 with the parse error reported as a violation. Document which behavior is chosen.)
- **Verification:** `echo $?` returns 2. Error message identifies the file and parse error.

**Test ID: CLI-LINT-EXIT-05: Exit code with --quiet still correct**

- **Input:**
```bash
stricture --quiet src/
```
- **Setup:** Files have error-level violations.
- **Expected:** Exit code 1 (errors exist) even though output may be reduced.
- **Verification:** `echo $?` returns 1 regardless of `--quiet` flag.

**Test ID: CLI-LINT-EXIT-06: Exit 0 when zero files and no errors**

- **Input:**
```bash
stricture "nonexistent-pattern/**/*.ts"
```
- **Expected:** Exit code 0. Informational message about no files matched.
- **Verification:** `echo $?` returns 0.

**Test ID: CLI-LINT-EXIT-07: Exit 1 with mixed errors and warnings**

- **Input:**
```bash
stricture src/
```
- **Setup:** 3 warnings and 1 error across files.
- **Expected:** Exit code 1 (error takes precedence over warning).
- **Verification:** `echo $?` returns 1. Summary shows both error and warning counts.

---

### 1.4 Caching

**Test ID: CLI-LINT-CACHE-01: Cache enabled by default**

- **Input:**
```bash
stricture src/
stricture src/   # second run
```
- **Setup:** No files changed between runs.
- **Expected:** Second run is significantly faster (uses cached ASTs). `.stricture-cache/` directory exists after first run.
- **Verification:** `.stricture-cache/` directory created. `--verbose` on second run shows cache hit stats.

**Test ID: CLI-LINT-CACHE-02: --no-cache disables caching**

- **Input:**
```bash
stricture --no-cache src/
```
- **Expected:** No `.stricture-cache/` directory created. Full parse on every run.
- **Verification:** `.stricture-cache/` does not exist after run. `--verbose` shows no cache hits.

**Test ID: CLI-LINT-CACHE-03: Cache hit on unchanged file**

- **Input:**
```bash
stricture --verbose src/a.ts
# do not modify a.ts
stricture --verbose src/a.ts
```
- **Expected:** Second run reports cache hit for `a.ts`. Faster parse time.
- **Verification:** Verbose output on second run contains cache hit indicator for `a.ts`.

**Test ID: CLI-LINT-CACHE-04: Cache miss on changed file**

- **Input:**
```bash
stricture --verbose src/a.ts
echo "// new comment" >> src/a.ts
stricture --verbose src/a.ts
```
- **Expected:** Second run reports cache miss for `a.ts` (file content changed). Full re-parse.
- **Verification:** Verbose output on second run shows cache miss for `a.ts`.

**Test ID: CLI-LINT-CACHE-05: Cache invalidation on adapter version change**

- **Input:**
```bash
stricture src/a.ts                    # populates cache
# simulate adapter version bump (change adapter hash)
stricture --verbose src/a.ts          # should re-parse
```
- **Expected:** Cache invalidated because cache key includes adapter version. Full re-parse.
- **Verification:** All files show cache miss despite no content changes.

**Test ID: CLI-LINT-CACHE-06: .stricture-cache/ creation in project root**

- **Input:**
```bash
stricture src/
ls -d .stricture-cache/
```
- **Expected:** `.stricture-cache/` created in project root (same directory as `.stricture.yml` or cwd).
- **Verification:** Directory exists and contains cache files.

**Test ID: CLI-LINT-CACHE-07: Same results with and without cache**

- **Input:**
```bash
stricture --format json --no-cache src/ > uncached.json
stricture --format json src/ > cached.json
```
- **Expected:** Both JSON outputs have identical `violations` arrays (same rule IDs, same files, same lines). Only `elapsed_ms` and `timestamp` differ.
- **Verification:** `jq .violations uncached.json` equals `jq .violations cached.json`.

---

### 1.5 Concurrency

**Test ID: CLI-LINT-CONC-01: Default concurrency matches CPU count**

- **Input:**
```bash
stricture --verbose src/
```
- **Expected:** Verbose output shows concurrency level matching system CPU count (e.g., `Concurrency: 8`).
- **Verification:** Reported concurrency matches `nproc` or `os.cpus().length`.

**Test ID: CLI-LINT-CONC-02: --concurrency N overrides default**

- **Input:**
```bash
stricture --concurrency 2 --verbose src/
```
- **Expected:** Verbose output shows `Concurrency: 2`. Only 2 files processed in parallel.
- **Verification:** Reported concurrency is 2.

**Test ID: CLI-LINT-CONC-03: Single-threaded produces same results as parallel**

- **Input:**
```bash
stricture --concurrency 1 --format json src/ > single.json
stricture --format json src/ > parallel.json
```
- **Expected:** Identical `violations` arrays (same count, same rule IDs, same files, same line numbers). Order may differ.
- **Verification:** Sort violations by (file, line, ruleId) and compare. Must be identical.

**Test ID: CLI-LINT-CONC-04: --concurrency 0 produces error**

- **Input:**
```bash
stricture --concurrency 0 src/
```
- **Expected:** Exit code 2. Error message: concurrency must be at least 1.
- **Verification:** `echo $?` returns 2.

**Test ID: CLI-LINT-CONC-05: --concurrency with negative number**

- **Input:**
```bash
stricture --concurrency -1 src/
```
- **Expected:** Exit code 2. Error message about invalid concurrency value.
- **Verification:** `echo $?` returns 2.

---

### 1.6 Inline Suppressions

**Test ID: CLI-LINT-SUPPRESS-01: stricture-disable-next-line suppresses one line**

- **Input file (`src/a.test.ts`):**
```typescript
// stricture-disable-next-line TQ-no-shallow-assertions
expect(result).toBeDefined();
expect(other).toBeDefined(); // no suppression
```
- **Input:** `stricture src/a.test.ts`
- **Expected:** First `expect` is suppressed (no violation). Second `expect` produces `TQ-no-shallow-assertions` violation.
- **Verification:** Exactly 1 violation reported, on the line without the suppression comment.

**Test ID: CLI-LINT-SUPPRESS-02: stricture-disable / stricture-enable block suppression**

- **Input file:**
```typescript
// stricture-disable TQ-no-shallow-assertions
expect(a).toBeDefined();
expect(b).toBeDefined();
// stricture-enable TQ-no-shallow-assertions
expect(c).toBeDefined(); // not suppressed
```
- **Input:** `stricture src/a.test.ts`
- **Expected:** Lines inside the disable/enable block are suppressed. The line after `enable` produces a violation.
- **Verification:** 1 violation on the `expect(c)` line.

**Test ID: CLI-LINT-SUPPRESS-03: Go inline suppression syntax**

- **Input file (`internal/a_test.go`):**
```go
// stricture-disable-next-line ARCH-import-boundary
import "internal/capture"
```
- **Input:** `stricture internal/a_test.go`
- **Expected:** ARCH-import-boundary violation suppressed.
- **Verification:** No violations for the suppressed import.

---

## 2. `stricture fix`

### 2.1 Fix Application

**Test ID: CLI-FIX-APPLY-01: Single fixable rule applied**

- **Input:**
```bash
stricture fix src/utils/helpers.ts
```
- **Setup:** `helpers.ts` is missing a file header (CONV-file-header violation). Config pattern: `// {filename} -- {purpose}`.
- **Expected:** File modified to include header comment. Exit code 0 or 1 depending on remaining unfixable violations.
- **Verification:** `head -1 src/utils/helpers.ts` shows the added header comment. Re-running `stricture src/utils/helpers.ts` no longer shows CONV-file-header.

**Test ID: CLI-FIX-APPLY-02: Multiple fixable rules applied in one pass**

- **Input:**
```bash
stricture fix src/
```
- **Setup:** `src/UserService.ts` has CONV-file-naming violation (should be `user-service.ts`) and CONV-file-header violation (missing header).
- **Expected:** File renamed and header added. Both fixes applied.
- **Verification:** `ls src/` shows `user-service.ts` (not `UserService.ts`). `head -1 src/user-service.ts` shows header.

**Test ID: CLI-FIX-APPLY-03: Non-fixable rule not modified**

- **Input:**
```bash
stricture fix src/
```
- **Setup:** File has TQ-no-shallow-assertions (not fixable) and CONV-file-header (fixable).
- **Expected:** Only CONV-file-header fixed. TQ-no-shallow-assertions still reported as a remaining violation.
- **Verification:** Re-run without `fix`: TQ-no-shallow-assertions still present, CONV-file-header gone.

**Test ID: CLI-FIX-APPLY-04: Mix of fixable and non-fixable rules**

- **Input:**
```bash
stricture fix --format json src/
```
- **Setup:** 5 violations total: 2 fixable (CONV-file-header, CONV-file-naming), 3 non-fixable (TQ-*, ARCH-*).
- **Expected:** JSON output shows 2 fixes applied and 3 remaining violations.
- **Verification:** JSON `summary.fixable` shows the count. Fixed violations no longer appear on re-run.

**Test ID: CLI-FIX-APPLY-05: Fix changes file content correctly**

- **Input:**
```bash
stricture fix src/helpers.ts
```
- **Setup:** `helpers.ts` lacks a file header. Expected header: `// helpers.ts -- Utility functions.`
- **Expected:** After fix, first line of file is the header comment. Remaining file content unchanged.
- **Verification:** Diff between before and after shows only the added header line. No other changes.

**Test ID: CLI-FIX-APPLY-06: --fix flag on lint command (alternative syntax)**

- **Input:**
```bash
stricture --fix src/
```
- **Expected:** Equivalent to `stricture fix src/`. Fixes applied, violations reported for non-fixable rules.
- **Verification:** Same behavior as `stricture fix src/`.

**Test ID: CLI-FIX-APPLY-07: Fix only specific rules with --rule**

- **Input:**
```bash
stricture fix --rule CONV-file-header src/
```
- **Setup:** Files have both CONV-file-header and CONV-file-naming violations.
- **Expected:** Only CONV-file-header fixes applied. CONV-file-naming violations remain.
- **Verification:** Re-run `stricture src/`: CONV-file-header gone, CONV-file-naming still present.

---

### 2.2 Fix Dry Run

**Test ID: CLI-FIX-DRY-01: --fix-dry-run shows diff without modifying files**

- **Input:**
```bash
stricture --fix-dry-run src/utils/helpers.ts
```
- **Setup:** `helpers.ts` missing file header.
- **Expected:** Output shows a diff of what would change (added header line with `+` prefix). File is NOT modified.
- **Verification:** `md5sum src/utils/helpers.ts` is identical before and after. Output contains `+` diff lines showing the header.

**Test ID: CLI-FIX-DRY-02: Dry run output format matches product spec**

- **Input:**
```bash
stricture --fix-dry-run src/
```
- **Setup:** Multiple fixable violations.
- **Expected:** Output shows each file, the rule ID, and the proposed change. Ends with `N fixes available (run with --fix to apply)`.
- **Verification:** Output matches format from product spec section 11.3.

**Test ID: CLI-FIX-DRY-03: Files are verified unchanged after dry run**

- **Input:**
```bash
md5sum src/**/*.ts > before.md5
stricture --fix-dry-run src/
md5sum src/**/*.ts > after.md5
diff before.md5 after.md5
```
- **Expected:** No difference. All file checksums identical.
- **Verification:** `diff` produces no output.

**Test ID: CLI-FIX-DRY-04: Dry run shows multiple fixes per file**

- **Input:**
```bash
stricture --fix-dry-run src/BadFile.ts
```
- **Setup:** File has CONV-file-naming (needs rename) and CONV-file-header (needs header) violations.
- **Expected:** Output shows both proposed fixes for the same file.
- **Verification:** Two fix proposals listed under the same file path.

---

### 2.3 Fix Safety

**Test ID: CLI-FIX-SAFE-01: Idempotency -- running fix twice produces same result**

- **Input:**
```bash
stricture fix src/
stricture fix src/   # second run
```
- **Expected:** Second run produces no changes (all fixable violations already resolved).
- **Verification:** `stricture --fix-dry-run src/` after second run shows `0 fixes available`. File contents identical between first and second fix run.

**Test ID: CLI-FIX-SAFE-02: Read-only file handling**

- **Input:**
```bash
chmod 444 src/helpers.ts
stricture fix src/helpers.ts
```
- **Setup:** `helpers.ts` has a fixable violation but is read-only.
- **Expected:** Error message indicating the file cannot be written. Fix not applied. Process does not crash.
- **Verification:** Error message mentions the file path and permission issue. File content unchanged. Exit code is non-zero.

**Test ID: CLI-FIX-SAFE-03: Fix does not introduce new violations**

- **Input:**
```bash
stricture --format json src/ > before.json
stricture fix src/
stricture --format json src/ > after.json
```
- **Expected:** The set of violations in `after.json` is a subset of (or equal to) `before.json`. No new rule IDs or files appear.
- **Verification:** Every violation in `after.json` also exists in `before.json` (by ruleId + file + line, allowing line shift due to fixes).

**Test ID: CLI-FIX-SAFE-04: File encoding preserved (UTF-8)**

- **Input:**
```bash
stricture fix src/unicode.ts
```
- **Setup:** `unicode.ts` contains UTF-8 characters (emoji, CJK, accented chars) and has a fixable violation.
- **Expected:** Fix applied. All non-ASCII characters preserved exactly. File encoding remains UTF-8.
- **Verification:** `file src/unicode.ts` shows UTF-8 encoding. Non-ASCII content unchanged.

**Test ID: CLI-FIX-SAFE-05: File encoding preserved (UTF-8 with BOM)**

- **Input:**
```bash
stricture fix src/bom-file.ts
```
- **Setup:** File starts with UTF-8 BOM (`EF BB BF`). Has a fixable violation.
- **Expected:** BOM preserved after fix. Fix does not corrupt the BOM marker.
- **Verification:** First 3 bytes are still `EF BB BF`.

**Test ID: CLI-FIX-SAFE-06: Trailing newline preserved**

- **Input:**
```bash
stricture fix src/helpers.ts
```
- **Setup:** File ends with a newline character. Has a fixable violation.
- **Expected:** Trailing newline preserved after fix.
- **Verification:** Last byte of file is `0x0A`.

---

## 3. `stricture audit`

### 3.1 Manifest Loading

**Test ID: CLI-AUDIT-MANIFEST-01: Load manifest from --manifest flag**

- **Input:**
```bash
stricture audit --manifest ./infra/stricture-manifest.yml
```
- **Setup:** Manifest at `./infra/stricture-manifest.yml` declares services and contracts.
- **Expected:** Audit runs using the specified manifest. Output shows contract analysis.
- **Verification:** Output contains the contract names from the manifest file.

**Test ID: CLI-AUDIT-MANIFEST-02: Auto-detect from .stricture.yml manifest.path**

- **Input:**
```bash
stricture audit
```
- **Setup:** `.stricture.yml` has `manifest.path: "../stricture-manifest/stricture-manifest.yml"`. Manifest exists at that path.
- **Expected:** Manifest auto-detected and loaded. Audit runs.
- **Verification:** Output contains contract analysis. No "manifest not found" error.

**Test ID: CLI-AUDIT-MANIFEST-03: Auto-detect from .stricture.yml manifest.url (git)**

- **Input:**
```bash
stricture audit
```
- **Setup:** `.stricture.yml` has `manifest.url: "github.com/acme/stricture-manifest"`.
- **Expected:** Manifest cloned/fetched from git URL and loaded. Audit runs.
- **Verification:** Output contains contract analysis. Manifest content matches remote repo.

**Test ID: CLI-AUDIT-MANIFEST-04: Missing manifest produces helpful error**

- **Input:**
```bash
stricture audit
```
- **Setup:** No `--manifest` flag. No `manifest` section in `.stricture.yml`. No manifest file in project.
- **Expected:** Exit code 2. Error message explains how to provide a manifest (via `--manifest` flag, or `manifest.path`/`manifest.url` in `.stricture.yml`).
- **Verification:** `echo $?` returns 2. Stderr contains actionable instructions.

**Test ID: CLI-AUDIT-MANIFEST-05: Invalid manifest YAML produces error with line number**

- **Input:**
```bash
stricture audit --manifest invalid-manifest.yml
```
- **Setup:** `invalid-manifest.yml` has a YAML syntax error on line 15.
- **Expected:** Exit code 2. Error message includes the file path and line number of the YAML error.
- **Verification:** Error message contains `invalid-manifest.yml` and `line 15` (or similar).

**Test ID: CLI-AUDIT-MANIFEST-06: Manifest with missing required fields**

- **Input:**
```bash
stricture audit --manifest incomplete.yml
```
- **Setup:** Manifest YAML is valid but missing `services` or `contracts` sections.
- **Expected:** Exit code 2. Error message identifies which required section is missing.
- **Verification:** Error message specifies the missing field.

---

### 3.2 Service Detection

**Test ID: CLI-AUDIT-SERVICE-01: --service flag overrides auto-detection**

- **Input:**
```bash
stricture audit --service billing-service
```
- **Setup:** `.stricture.yml` has `manifest.service: user-service`. Manifest declares both services.
- **Expected:** Audit runs for `billing-service`, not `user-service`.
- **Verification:** Output header shows `billing-service`. Contracts analyzed are those for `billing-service`.

**Test ID: CLI-AUDIT-SERVICE-02: Auto-detect from .stricture.yml manifest.service**

- **Input:**
```bash
stricture audit
```
- **Setup:** `.stricture.yml` has `manifest.service: user-service`.
- **Expected:** Audit runs for `user-service`.
- **Verification:** Output header shows `user-service`.

**Test ID: CLI-AUDIT-SERVICE-03: Unknown service name produces error**

- **Input:**
```bash
stricture audit --service nonexistent-service
```
- **Setup:** Manifest does not declare `nonexistent-service`.
- **Expected:** Exit code 2. Error message states service not found in manifest. Lists valid service names from manifest.
- **Verification:** Error message contains `nonexistent-service` and lists actual service names (e.g., `api-gateway, user-service, billing-service`).

**Test ID: CLI-AUDIT-SERVICE-04: No service specified and no auto-detect**

- **Input:**
```bash
stricture audit
```
- **Setup:** No `--service` flag. No `manifest.service` in `.stricture.yml`.
- **Expected:** Exit code 2. Error message explains how to specify a service.
- **Verification:** Helpful error message with both flag and config options.

---

### 3.3 Scorecard Generation

**Test ID: CLI-AUDIT-SCORE-01: Per-field strictness level calculation**

- **Input:**
```bash
stricture audit
```
- **Setup:** Manifest declares `id: { type: integer, range: [1, 2147483647], required: true }`. Code has `id: number` with range validation.
- **Expected:** Field `id` shows strictness level `strict` (type + required + format + constraints).
- **Verification:** Output table row for `id` shows strict-level status.

**Test ID: CLI-AUDIT-SCORE-02: Overall strictness percentage**

- **Input:**
```bash
stricture audit
```
- **Setup:** Manifest has 5 fields. Code validates 4 strictly, 1 minimally.
- **Expected:** Output shows percentage (e.g., `Response strictness: 80%`).
- **Verification:** Percentage matches actual validation coverage.

**Test ID: CLI-AUDIT-SCORE-03: Correct status symbols**

- **Input:**
```bash
stricture audit
```
- **Setup:** 3 fields: one strict, one basic (missing constraint), one with no validation.
- **Expected output symbols:**
  - Strict field: `✓ Strict`
  - Basic field: `⚠ Missing range check` (or similar)
  - No validation: `✗ No format validation` (or similar)
- **Verification:** Output contains exactly the correct Unicode symbols for each field status.

**Test ID: CLI-AUDIT-SCORE-04: Multiple contracts for same service**

- **Input:**
```bash
stricture audit
```
- **Setup:** Service participates in 2 contracts (`user-api` as producer, `billing-events` as consumer).
- **Expected:** Output shows separate scorecard sections for each contract, each with its own strictness percentage.
- **Verification:** Both contract names appear as section headers. Each has its own table and percentage.

**Test ID: CLI-AUDIT-SCORE-05: Producer vs consumer perspective**

- **Input:**
```bash
stricture audit --service user-service
```
- **Setup:** `user-service` is producer for `user-api` and consumer for `billing-events`.
- **Expected:** Producer section validates response fields (does code enforce constraints before sending?). Consumer section validates request/message handling (does code validate incoming data?).
- **Verification:** Output labels contracts as `(producer)` and `(consumer)`. Different validation criteria applied.

**Test ID: CLI-AUDIT-SCORE-06: Summary line with totals**

- **Input:**
```bash
stricture audit
```
- **Expected:** Final summary line matches format: `N contracts, M fields, X strict, Y basic, Z missing validation`.
- **Verification:** Numbers in summary match actual field counts from all contracts.

---

### 3.4 Field Analysis

**Test ID: CLI-AUDIT-FIELD-01: Range constraint detection in code**

- **Setup:** Manifest: `amount: { type: number, range: [0.01, 999999.99] }`. Code:
```go
if amount < 0.01 || amount > 999999.99 {
    return errors.New("amount out of range")
}
```
- **Expected:** Field `amount` shows `✓ Strict` -- range validation detected.
- **Verification:** Status column shows strict/passing indicator.

**Test ID: CLI-AUDIT-FIELD-02: Range constraint missing in code**

- **Setup:** Manifest: `amount: { type: number, range: [0.01, 999999.99] }`. Code: `amount := body["amount"].(float64)` with no range check.
- **Expected:** Field `amount` shows `✗ No range validation`.
- **Verification:** Status column shows failure indicator with clear message.

**Test ID: CLI-AUDIT-FIELD-03: Enum constraint detection in code**

- **Setup:** Manifest: `role: { type: enum, values: ["admin", "user", "viewer"] }`. Code:
```typescript
if (!["admin", "user", "viewer"].includes(role)) {
    throw new Error("invalid role");
}
```
- **Expected:** Field `role` shows `✓ Strict` -- enum validation detected.
- **Verification:** Enum values in code match manifest exactly.

**Test ID: CLI-AUDIT-FIELD-04: Partial enum coverage in code**

- **Setup:** Manifest: `role: { type: enum, values: ["admin", "user", "viewer"] }`. Code only checks `"admin"` and `"user"` in switch statement.
- **Expected:** Field `role` shows `⚠ Partial enum coverage (2/3)` or similar warning.
- **Verification:** Warning identifies which enum values are missing.

**Test ID: CLI-AUDIT-FIELD-05: Format validation detection (email)**

- **Setup:** Manifest: `email: { type: string, format: email }`. Code uses email regex or `validator.isEmail()`.
- **Expected:** Field `email` shows `✓ Strict` -- email format validation detected.
- **Verification:** Format validation recognized.

**Test ID: CLI-AUDIT-FIELD-06: Length validation detection**

- **Setup:** Manifest: `name: { type: string, minLength: 1, maxLength: 255 }`. Code: `if len(name) == 0 || len(name) > 255`.
- **Expected:** Field `name` shows `✓ Strict` -- length validation detected.
- **Verification:** Both minLength and maxLength checks found.

**Test ID: CLI-AUDIT-FIELD-07: Length validation partially missing**

- **Setup:** Manifest: `name: { type: string, minLength: 1, maxLength: 255 }`. Code only checks `if name === ""` (empty check but no max length).
- **Expected:** Field `name` shows `⚠ Missing maxLength check`.
- **Verification:** Identifies which constraint is missing.

**Test ID: CLI-AUDIT-FIELD-08: Custom validation library detection -- zod**

- **Setup:** Manifest: `email: { type: string, format: email }`. Code:
```typescript
const schema = z.object({
    email: z.string().email(),
});
```
- **Expected:** Field `email` shows `✓ Strict` -- zod email validation detected.
- **Verification:** Zod schema recognized as format validation.

**Test ID: CLI-AUDIT-FIELD-09: Custom validation library detection -- joi**

- **Setup:** Code: `Joi.object({ email: Joi.string().email().required() })`.
- **Expected:** Joi email validation detected. Field shows strict status.
- **Verification:** Joi schema recognized.

**Test ID: CLI-AUDIT-FIELD-10: Custom validation library detection -- go-playground/validator**

- **Setup:** Code:
```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
}
```
- **Expected:** Go struct tag validation detected. Field shows strict status.
- **Verification:** `validate` struct tag parsed and matched to manifest constraint.

---

### 3.5 Output Formats

**Test ID: CLI-AUDIT-FORMAT-01: Text format with table**

- **Input:**
```bash
stricture audit
```
- **Expected:** Output contains Unicode box-drawing characters (e.g., `┌─────`, `│`, `├─────`, `└─────`). Columns: Field, Manifest, Code, Status. Matches product spec section 13.7 example.
- **Verification:** Output parseable as a text table. Contains correct Unicode box characters.

**Test ID: CLI-AUDIT-FORMAT-02: JSON format**

- **Input:**
```bash
stricture audit --format json
```
- **Expected:** Valid JSON output containing:
  - `service`: string
  - `contracts`: array of contract objects
  - Each contract: `id`, `role`, `endpoints`/`messages`
  - Each endpoint: `fields` array with `name`, `manifest_spec`, `code_spec`, `status`, `strictness_level`
  - Top-level `summary` with `strictness_percentage`, `total_fields`, `strict_count`, `basic_count`, `missing_count`
- **Verification:** `jq .` succeeds. All expected fields present.

**Test ID: CLI-AUDIT-FORMAT-03: SARIF format for audit findings**

- **Input:**
```bash
stricture audit --format sarif
```
- **Expected:** Valid SARIF 2.1.0 JSON. Audit findings mapped to SARIF results. Each missing/weak validation is a result with physical location pointing to the relevant code.
- **Verification:** SARIF schema validates. Results have `ruleId`, `message`, `locations`.

---

### 3.6 Remote Validation

**Test ID: CLI-AUDIT-REMOTE-01: --remote fetches other service repos**

- **Input:**
```bash
stricture audit --remote
```
- **Setup:** Manifest declares `user-service` and `billing-service` with `repo` URLs. Current service is `user-service`.
- **Expected:** Fetches `billing-service` repo. Validates cross-service parity (e.g., consumer strictness matches producer strictness).
- **Verification:** Output includes cross-service comparison data. Remote service fields appear in analysis.

**Test ID: CLI-AUDIT-REMOTE-02: Network failure handling**

- **Input:**
```bash
stricture audit --remote
```
- **Setup:** Remote repo URL is unreachable (DNS failure, timeout).
- **Expected:** Error message about network failure. Local audit still runs. Remote validation section shows "unavailable" or skipped.
- **Verification:** Partial results still produced. Clear error about which remote service could not be fetched.

**Test ID: CLI-AUDIT-REMOTE-03: --strictness override**

- **Input:**
```bash
stricture audit --strictness exhaustive
```
- **Setup:** Manifest has `strictness.minimum: strict`. Flag overrides to `exhaustive`.
- **Expected:** Audit evaluates all fields against `exhaustive` level (the override), not `strict`.
- **Verification:** Fields that pass at `strict` but fail at `exhaustive` show as violations.

---

## 4. `stricture trace`

### 4.1 HAR Format

**Test ID: CLI-TRACE-HAR-01: Standard HAR 1.2 parsing**

- **Input:**
```bash
stricture trace traffic.har
```
- **Setup:** `traffic.har` is a standard HAR 1.2 file from Chrome DevTools with 10 entries.
- **Expected:** All 10 entries parsed. Matched to manifest contracts. Violations reported per-endpoint.
- **Verification:** Output shows matched request count. No parse errors.

**Test ID: CLI-TRACE-HAR-02: Missing optional HAR fields (graceful handling)**

- **Input:**
```bash
stricture trace minimal.har
```
- **Setup:** HAR file where entries are missing optional fields like `comment`, `timings`, `serverIPAddress`.
- **Expected:** Parsing succeeds. Only required fields needed. No errors for missing optional fields.
- **Verification:** No warnings about missing optional fields. All entries still analyzed.

**Test ID: CLI-TRACE-HAR-03: Compressed response bodies (gzip)**

- **Input:**
```bash
stricture trace compressed.har
```
- **Setup:** HAR file with gzip-compressed response bodies (`content.encoding: "gzip"`, `content.text` is base64-encoded compressed data).
- **Expected:** Bodies decompressed before validation. Field values checked correctly.
- **Verification:** No errors about unreadable bodies. Field validation runs on decompressed content.

**Test ID: CLI-TRACE-HAR-04: Compressed response bodies (brotli)**

- **Input:**
```bash
stricture trace brotli.har
```
- **Setup:** HAR with brotli-compressed bodies (`content.encoding: "br"`).
- **Expected:** Bodies decompressed correctly. Validation runs.
- **Verification:** Same as gzip test.

**Test ID: CLI-TRACE-HAR-05: Binary response bodies skipped**

- **Input:**
```bash
stricture trace mixed.har
```
- **Setup:** HAR with JSON API responses and binary responses (images, PDFs). Binary entries have `content.mimeType: "image/png"`.
- **Expected:** Binary bodies skipped for field validation. JSON bodies validated. No errors for binary entries.
- **Verification:** Binary entries counted as "unmatched" or skipped. No parse errors.

**Test ID: CLI-TRACE-HAR-06: Redirect chains (301/302)**

- **Input:**
```bash
stricture trace redirects.har
```
- **Setup:** HAR file with a 301 redirect from `/old-users` to `/api/users`, followed by the final 200 response.
- **Expected:** Redirect entries not validated for body content (no body). Final response validated against manifest.
- **Verification:** 301/302 entries not flagged for missing required fields (they are redirects). Final response validated.

**Test ID: CLI-TRACE-HAR-07: Large HAR files (1000+ entries)**

- **Input:**
```bash
stricture trace large-traffic.har
```
- **Setup:** HAR file with 1500 entries.
- **Expected:** All entries parsed and validated. Performance: completes within reasonable time (< 30s). No memory issues.
- **Verification:** Summary shows all entries accounted for (matched + unmatched = total). Process does not OOM.

**Test ID: CLI-TRACE-HAR-08: Empty HAR file (no entries)**

- **Input:**
```bash
stricture trace empty.har
```
- **Setup:** Valid HAR 1.2 structure but `log.entries` is an empty array.
- **Expected:** Exit code 0. Informational message: "No entries found in trace file."
- **Verification:** `echo $?` returns 0. No violations.

**Test ID: CLI-TRACE-HAR-09: Invalid HAR file (not valid JSON)**

- **Input:**
```bash
stricture trace not-json.har
```
- **Setup:** File contents are not valid JSON.
- **Expected:** Exit code 2. Error message: unable to parse trace file.
- **Verification:** `echo $?` returns 2. Error identifies the file and parse issue.

---

### 4.2 OpenTelemetry Format

**Test ID: CLI-TRACE-OTEL-01: JSON export format parsing**

- **Input:**
```bash
stricture trace otel-export.json --trace-format otel
```
- **Setup:** OpenTelemetry JSON export with spans for HTTP requests.
- **Expected:** Spans mapped to endpoints. Attributes extracted for field validation.
- **Verification:** Output shows matched span count. Endpoints resolved correctly.

**Test ID: CLI-TRACE-OTEL-02: Span-to-endpoint matching**

- **Setup:** Span attributes include `http.method: "GET"`, `http.target: "/api/users/42"`. Manifest declares `GET /api/users/:id`.
- **Expected:** Span matched to `GET /api/users/:id` contract endpoint.
- **Verification:** Span appears under the correct endpoint in output.

**Test ID: CLI-TRACE-OTEL-03: Attribute extraction for field validation**

- **Setup:** Span has `http.response.body` attribute with JSON user object. Manifest declares field constraints.
- **Expected:** Fields extracted from response body. Validated against manifest constraints.
- **Verification:** Per-field violations reported (if any).

**Test ID: CLI-TRACE-OTEL-04: Multi-service traces (filter by service)**

- **Input:**
```bash
stricture trace otel-export.json --trace-format otel --service user-service
```
- **Setup:** OTel export contains spans from multiple services (user-service, billing-service, api-gateway).
- **Expected:** Only spans from `user-service` analyzed. Other services' spans ignored.
- **Verification:** Output header shows `user-service`. No analysis of other services' spans.

**Test ID: CLI-TRACE-OTEL-05: OTel format auto-detection**

- **Input:**
```bash
stricture trace otel-export.json
```
- **Setup:** File has OTel JSON structure (contains `resourceSpans` top-level key).
- **Expected:** Auto-detected as OTel format. No need for `--trace-format otel`.
- **Verification:** Same results as with explicit `--trace-format otel`.

---

### 4.3 Custom JSON Format

**Test ID: CLI-TRACE-CUSTOM-01: Valid custom JSON format**

- **Input:**
```bash
stricture trace custom-traffic.json --trace-format custom
```
- **Setup:** File content:
```json
[
  {
    "method": "GET",
    "path": "/api/users/42",
    "request_body": null,
    "response_body": { "id": 42, "name": "Alice", "email": "alice@example.com", "role": "admin", "created_at": "2026-01-01T00:00:00Z" },
    "status": 200
  }
]
```
- **Expected:** Entry parsed. Matched to `GET /api/users/:id`. Response body fields validated.
- **Verification:** Output shows 1 matched request. Field validation results appear.

**Test ID: CLI-TRACE-CUSTOM-02: Partial fields (some missing)**

- **Input:**
```bash
stricture trace partial.json --trace-format custom
```
- **Setup:** Entries missing `request_body` and `status` fields.
- **Expected:** Graceful handling. Fields present are validated. Missing fields do not cause parse errors.
- **Verification:** No crash. Entries with partial data still matched where possible. Warning about missing fields (optional).

**Test ID: CLI-TRACE-CUSTOM-03: Nested request/response bodies**

- **Setup:** Response body with nested objects:
```json
{
  "user": {
    "id": 42,
    "profile": { "name": "Alice", "settings": { "theme": "dark" } }
  }
}
```
- **Expected:** Nested fields validated against manifest (if manifest declares nested constraints). Deeply nested structures traversed.
- **Verification:** Validation reaches into nested objects.

---

### 4.4 Contract Matching

**Test ID: CLI-TRACE-MATCH-01: Exact path matching**

- **Setup:** Trace entry: `GET /api/users`. Manifest: `path: "/api/users"`, `method: GET`.
- **Expected:** Exact match. Entry validated against contract.
- **Verification:** Entry appears under the correct contract endpoint.

**Test ID: CLI-TRACE-MATCH-02: Path parameter matching**

- **Setup:** Trace entry: `GET /api/users/123`. Manifest: `path: "/api/users/:id"`, `method: GET`.
- **Expected:** Path parameter `123` matched to `:id`. Entry validated.
- **Verification:** Entry grouped under `GET /api/users/:id`. `123` extracted as `id` value.

**Test ID: CLI-TRACE-MATCH-03: Multiple path parameters**

- **Setup:** Trace entry: `GET /api/orgs/5/users/42`. Manifest: `path: "/api/orgs/:orgId/users/:userId"`.
- **Expected:** Both parameters extracted. Entry matched.
- **Verification:** Entry appears under correct endpoint.

**Test ID: CLI-TRACE-MATCH-04: Method matching (GET vs POST to same path)**

- **Setup:** Trace entries: `GET /api/users` and `POST /api/users`. Manifest declares both with different contracts.
- **Expected:** GET entries matched to GET contract. POST entries matched to POST contract. Different validation for each.
- **Verification:** GET and POST entries analyzed separately with their respective field constraints.

**Test ID: CLI-TRACE-MATCH-05: Unmatched requests reported but not errors**

- **Setup:** Trace entries include `GET /healthz` and `GET /metrics`. Manifest has no contracts for these paths.
- **Expected:** Unmatched requests reported in summary section (e.g., `Unmatched requests: 57 (no manifest contract)`). These are informational, not violations.
- **Verification:** Exit code not affected by unmatched requests (unless `--strict`). Unmatched section lists paths and counts.

**Test ID: CLI-TRACE-MATCH-06: Multiple contracts for same URL pattern**

- **Setup:** Manifest declares two contracts from different services that both serve `GET /api/users/:id`. Trace is for one specific service.
- **Expected:** Matched to the correct contract based on `--service` flag or auto-detection.
- **Verification:** Correct contract's field constraints used for validation.

**Test ID: CLI-TRACE-MATCH-07: Query parameters ignored for path matching**

- **Setup:** Trace entry: `GET /api/users?page=1&limit=20`. Manifest: `path: "/api/users"`.
- **Expected:** Query parameters ignored during path matching. Entry matched to `/api/users`.
- **Verification:** Entry appears in matched results.

---

### 4.5 Field Validation

**Test ID: CLI-TRACE-FIELD-01: Required field missing in response**

- **Setup:** Manifest: `email: { type: string, format: email, required: true }`. Response body: `{ "id": 42, "name": "Alice" }` (missing `email`).
- **Expected:** Violation: required field `email` missing from response.
- **Verification:** Output contains violation message for missing `email` field.

**Test ID: CLI-TRACE-FIELD-02: Field value out of range**

- **Setup:** Manifest: `amount: { type: number, range: [0.01, 999999.99] }`. Response body: `{ "amount": 0.00 }`.
- **Expected:** Violation: field `amount` value `0.00` is below minimum `0.01`.
- **Verification:** Violation message includes the field name, actual value, and expected range.

**Test ID: CLI-TRACE-FIELD-03: Enum value not in allowed set**

- **Setup:** Manifest: `role: { type: enum, values: ["admin", "user", "viewer"] }`. Response body: `{ "role": "moderator" }`.
- **Expected:** Violation: field `role` value `"moderator"` not in allowed values `["admin", "user", "viewer"]`.
- **Verification:** Violation lists the invalid value and the allowed values.

**Test ID: CLI-TRACE-FIELD-04: String too long**

- **Setup:** Manifest: `name: { type: string, maxLength: 255 }`. Response body: `{ "name": "<256-character string>" }`.
- **Expected:** Violation: field `name` exceeds maxLength of 255.
- **Verification:** Violation includes field name, actual length, and max constraint.

**Test ID: CLI-TRACE-FIELD-05: String too short (minLength)**

- **Setup:** Manifest: `name: { type: string, minLength: 1 }`. Response body: `{ "name": "" }`.
- **Expected:** Violation: field `name` violates minLength of 1 (empty string).
- **Verification:** Violation identifies the field and the minLength constraint.

**Test ID: CLI-TRACE-FIELD-06: Format violation (invalid email)**

- **Setup:** Manifest: `email: { type: string, format: email }`. Response body: `{ "email": "not-an-email" }`.
- **Expected:** Violation: field `email` does not match email format.
- **Verification:** Violation identifies the field, actual value, and expected format.

**Test ID: CLI-TRACE-FIELD-07: Format violation (invalid UUID)**

- **Setup:** Manifest: `invoice_id: { type: string, format: uuid }`. Response body: `{ "invoice_id": "not-a-uuid" }`.
- **Expected:** Violation for invalid UUID format.
- **Verification:** Violation identifies the field and the expected UUID format.

**Test ID: CLI-TRACE-FIELD-08: Format violation (invalid ISO 8601)**

- **Setup:** Manifest: `created_at: { type: string, format: iso8601 }`. Response body: `{ "created_at": "2026/01/01" }`.
- **Expected:** Violation: `created_at` does not match ISO 8601 format.
- **Verification:** Violation identifies expected format.

**Test ID: CLI-TRACE-FIELD-09: Status code not in declared set**

- **Setup:** Manifest: `status_codes: [200, 400, 404, 500]`. Trace entry has status `503`.
- **Expected:** Warning: status code `503` not in declared set `[200, 400, 404, 500]`.
- **Verification:** Warning (not error) reported. Status code warnings are softer than field violations.

**Test ID: CLI-TRACE-FIELD-10: All fields valid -- no violations**

- **Setup:** Response body matches all manifest constraints exactly.
- **Expected:** Endpoint shows `✓ All responses had required fields` and no violations.
- **Verification:** No violation entries for this endpoint.

**Test ID: CLI-TRACE-FIELD-11: Numeric precision violation**

- **Setup:** Manifest: `amount: { type: number, range: [0.01, 999999.99], precision: 2 }`. Response body: `{ "amount": 10.123 }` (3 decimal places, contract says 2).
- **Expected:** Violation: field `amount` has precision 3, expected precision 2.
- **Verification:** Violation identifies the precision mismatch.

**Test ID: CLI-TRACE-FIELD-12: Extra fields in response (not in manifest)**

- **Setup:** Manifest declares `id`, `name`, `email`. Response body includes `{ "id": 1, "name": "Alice", "email": "a@b.com", "internal_id": "xyz" }`.
- **Expected:** Extra field `internal_id` noted but not necessarily a violation (depends on strictness mode). In `--strict` mode, this is a violation.
- **Verification:** Behavior documented. Extra fields reported if `--strict`.

**Test ID: CLI-TRACE-FIELD-13: Null value for required field**

- **Setup:** Manifest: `name: { type: string, required: true }`. Response body: `{ "name": null }`.
- **Expected:** Violation: required field `name` is null.
- **Verification:** Null treated as missing for required fields.

---

### 4.6 Trace Output

**Test ID: CLI-TRACE-OUTPUT-01: Per-endpoint violation summary**

- **Input:**
```bash
stricture trace traffic.har
```
- **Expected:** Output grouped by contract, then by endpoint. Each endpoint shows request count, passing checks, and violation details. Matches format from product spec section 13.8.
- **Verification:** Output structure matches:
```
Contract: user-api
  GET /api/users/:id (142 requests)
    ✓ All responses had required fields
    ✗ 3 responses had role="moderator" — not in enum [admin, user, viewer]
```

**Test ID: CLI-TRACE-OUTPUT-02: Unmatched request count**

- **Input:**
```bash
stricture trace traffic.har
```
- **Setup:** Some requests have no matching manifest contract.
- **Expected:** Unmatched section shows paths and counts:
```
Unmatched requests: 57 (no manifest contract)
  GET /healthz (30)
  GET /metrics (27)
```
- **Verification:** Total unmatched count matches sum of individual path counts.

**Test ID: CLI-TRACE-OUTPUT-03: Summary totals**

- **Input:**
```bash
stricture trace traffic.har
```
- **Expected:** Final summary line: `Summary: N matched, V violations, U unmatched` where N + U = total entries.
- **Verification:** Numbers are consistent. `matched + unmatched = total HAR entries`.

**Test ID: CLI-TRACE-OUTPUT-04: --strict mode (any deviation = failure)**

- **Input:**
```bash
stricture trace traffic.har --strict
```
- **Setup:** One warning-level deviation (e.g., undeclared status code 503).
- **Expected:** Exit code 1. In strict mode, even warnings become failures.
- **Verification:** `echo $?` returns 1. Without `--strict`, same trace returns 0.

**Test ID: CLI-TRACE-OUTPUT-05: JSON output format for trace**

- **Input:**
```bash
stricture trace traffic.har --format json
```
- **Expected:** Valid JSON with `contracts`, `unmatched`, and `summary` sections. Each contract has `endpoints` with `request_count`, `violations`, and `checks`.
- **Verification:** `jq .` succeeds. All expected fields present.

**Test ID: CLI-TRACE-OUTPUT-06: Trace with manifest auto-detection**

- **Input:**
```bash
stricture trace traffic.har
```
- **Setup:** No `--manifest` flag. `.stricture.yml` has `manifest.path`.
- **Expected:** Manifest auto-detected. Trace runs normally.
- **Verification:** No "manifest not found" error. Results produced.

**Test ID: CLI-TRACE-OUTPUT-07: --trace-format auto-detection**

- **Input:**
```bash
stricture trace traffic.har
```
- **Expected:** HAR format auto-detected from file extension and/or content structure (`log.version`, `log.entries`).
- **Verification:** Same results as with explicit `--trace-format har`.

**Test ID: CLI-TRACE-OUTPUT-08: Auto-detection falls back on ambiguous file**

- **Input:**
```bash
stricture trace ambiguous.json
```
- **Setup:** File extension is `.json` and content could be custom or OTel.
- **Expected:** Stricture attempts auto-detection. If ambiguous, tries each parser. If all fail, reports error asking user to specify `--trace-format`.
- **Verification:** Either auto-detects correctly or provides a clear error message.

---

## 5. `stricture init`

### 5.1 Config Generation

**Test ID: CLI-INIT-01: Creates .stricture.yml in current directory**

- **Input:**
```bash
stricture init
```
- **Setup:** No `.stricture.yml` exists.
- **Expected:** `.stricture.yml` created in current directory. Contains default configuration with all rules enabled.
- **Verification:** `test -f .stricture.yml` succeeds. File is valid YAML. Contains `rules:` section.

**Test ID: CLI-INIT-02: Auto-detects TypeScript**

- **Input:**
```bash
stricture init
```
- **Setup:** Directory contains `src/app.ts`, `src/utils.ts`.
- **Expected:** Generated config includes `languages: [typescript]` (or equivalent auto-detection section).
- **Verification:** Config references TypeScript adapter.

**Test ID: CLI-INIT-03: Auto-detects Go**

- **Input:**
```bash
stricture init
```
- **Setup:** Directory contains `main.go`, `internal/handler.go`.
- **Expected:** Generated config includes `languages: [go]`.
- **Verification:** Config references Go adapter.

**Test ID: CLI-INIT-04: Auto-detects multiple languages**

- **Input:**
```bash
stricture init
```
- **Setup:** Directory contains `.ts`, `.go`, and `.py` files.
- **Expected:** Generated config lists all detected supported languages.
- **Verification:** Config `languages` array contains at least `typescript` and `go`.

**Test ID: CLI-INIT-05: Default rules (all enabled at error level)**

- **Input:**
```bash
stricture init
```
- **Expected:** Generated config sets all 30 rules (excluding 2 manifest-only and 2 reserved) to `error` by default.
- **Verification:** Each TQ-*, ARCH-*, CONV-*, CTR-* rule (that doesn't require manifest) is set to `error`.

**Test ID: CLI-INIT-06: Does not overwrite existing config**

- **Input:**
```bash
stricture init
```
- **Setup:** `.stricture.yml` already exists with custom content.
- **Expected:** Exit with message like `".stricture.yml" already exists. Use --force to overwrite.` File not modified.
- **Verification:** File content identical before and after. Exit message is informational, not an error.

**Test ID: CLI-INIT-07: Init in empty directory**

- **Input:**
```bash
mkdir /tmp/empty-project && cd /tmp/empty-project
stricture init
```
- **Expected:** `.stricture.yml` created with generic defaults. No `languages` section (nothing to detect) or `languages: []`.
- **Verification:** File created. Valid YAML.

---

## 6. `stricture inspect`

### 6.1 TypeScript Files

**Test ID: CLI-INSPECT-TS-01: Shows parsed UnifiedFileModel for TS**

- **Input:**
```bash
stricture inspect src/services/user-service.ts
```
- **Setup:** File exports a class with 3 methods, imports from 2 modules, exports 1 type.
- **Expected:** Output shows the UnifiedFileModel fields:
  - `path`: full file path
  - `language`: `typescript`
  - `imports`: list of import sources
  - `exports`: list of exported symbols
  - `functions`: list of function/method names with signatures
  - `classes`: class names with method lists
  - `types`: type/interface definitions
- **Verification:** All structural elements of the file appear in output. Useful for debugging adapter behavior.

**Test ID: CLI-INSPECT-TS-02: Test file detection and test target mapping**

- **Input:**
```bash
stricture inspect tests/services/user-service.test.ts
```
- **Setup:** Test file imports from `../../src/services/user-service`.
- **Expected:** Output shows:
  - `isTestFile`: true
  - `testTargets`: `["src/services/user-service.ts"]`
  - `testFunctions`: list of test cases (e.g., `"should create user"`, `"should throw on invalid input"`)
- **Verification:** Test file correctly identified. Target source file resolved.

**Test ID: CLI-INSPECT-TS-03: Arrow functions and modern syntax**

- **Input:**
```bash
stricture inspect src/utils.ts
```
- **Setup:** File uses arrow functions, optional chaining, nullish coalescing, generics.
- **Expected:** All functions listed in output regardless of syntax style.
- **Verification:** Arrow functions appear in `functions` list.

---

### 6.2 Go Files

**Test ID: CLI-INSPECT-GO-01: Shows parsed UnifiedFileModel for Go**

- **Input:**
```bash
stricture inspect internal/handler/user.go
```
- **Setup:** File has 2 structs (one with JSON tags), 4 functions, 1 interface.
- **Expected:** Output shows:
  - `path`: full file path
  - `language`: `go`
  - `imports`: list of import paths
  - `exports`: exported identifiers (capital first letter)
  - `functions`: function list with signatures
  - `types`: struct and interface definitions
  - JSON tags shown for struct fields
- **Verification:** Structs include JSON tag information. All functions listed.

**Test ID: CLI-INSPECT-GO-02: Test file detection for Go**

- **Input:**
```bash
stricture inspect internal/handler/user_test.go
```
- **Expected:** Output shows:
  - `isTestFile`: true
  - `testTargets`: `["internal/handler/user.go"]` (convention: `_test.go` maps to same-name file)
  - `testFunctions`: list of `Test*` functions
- **Verification:** Go test functions correctly identified.

**Test ID: CLI-INSPECT-GO-03: Go interfaces and embedded types**

- **Input:**
```bash
stricture inspect pkg/types/interfaces.go
```
- **Setup:** File contains interfaces with embedded types and method sets.
- **Expected:** Interfaces listed with their method signatures. Embedded types noted.
- **Verification:** Interface methods appear in output.

---

### 6.3 Error Cases

**Test ID: CLI-INSPECT-ERR-01: Non-existent file**

- **Input:**
```bash
stricture inspect nonexistent.ts
```
- **Expected:** Exit code 2. Error message: file not found. Includes the path attempted.
- **Verification:** `echo $?` returns 2. Error message contains the file path.

**Test ID: CLI-INSPECT-ERR-02: Binary file**

- **Input:**
```bash
stricture inspect image.png
```
- **Expected:** Exit code 2. Error message: cannot inspect binary file.
- **Verification:** Clear error, no crash, no garbage output.

**Test ID: CLI-INSPECT-ERR-03: Unsupported language**

- **Input:**
```bash
stricture inspect script.rb
```
- **Expected:** Exit code 2. Error message: no language adapter for `.rb` files. Lists supported languages.
- **Verification:** Error message includes `.rb` and lists `typescript`, `go` as supported.

**Test ID: CLI-INSPECT-ERR-04: File with syntax errors**

- **Input:**
```bash
stricture inspect broken.ts
```
- **Setup:** File has a syntax error (unclosed brace, missing semicolon causing parse failure).
- **Expected:** Exit code 2 (or partial output with parse error location). Error identifies the syntax error location.
- **Verification:** Error includes file path and line/column of syntax error.

**Test ID: CLI-INSPECT-ERR-05: Empty file**

- **Input:**
```bash
stricture inspect empty.ts
```
- **Setup:** `empty.ts` is a 0-byte file.
- **Expected:** Output shows a valid UnifiedFileModel with empty arrays for all collections (imports: [], exports: [], functions: [], etc.).
- **Verification:** No crash. Valid output structure with empty collections.

---

## 7. `stricture list-rules`

**Test ID: CLI-LISTRULES-01: Lists all 34 rules**

- **Input:**
```bash
stricture list-rules
```
- **Expected:** Output lists all 34 rules (10 TQ + 6 ARCH + 6 CONV + 8 CTR + 2 reserved). Each rule shows:
  - Rule ID (e.g., `TQ-no-shallow-assertions`)
  - Description
  - Default severity (error/warn/off)
  - Fixable status (Yes/No/Partial)
  - Category (TQ/ARCH/CONV/CTR)
- **Verification:** Count of rule entries equals 34. All rule IDs from product spec section 15 present.

**Test ID: CLI-LISTRULES-02: Default severity shown**

- **Input:**
```bash
stricture list-rules
```
- **Expected:** Every rule shows its default severity. Per product spec, all rules default to `error`.
- **Verification:** Each rule line includes `error` as default severity.

**Test ID: CLI-LISTRULES-03: Fixable status shown**

- **Input:**
```bash
stricture list-rules
```
- **Expected output (partial):**
```
CONV-file-naming       error  Yes      Enforce file naming convention (kebab-case, etc.)
CONV-file-header       error  Yes      Require file header comments
TQ-no-shallow-assertions error No     Reject assertions that only check existence/truthiness
TQ-mock-scope          error  Partial  Mocks must be scoped and cleaned up
```
- **Verification:** Fixable rules match product spec section 15 and section 11.1: CONV-file-naming (Yes), CONV-file-header (Yes), CONV-export-naming (Yes), CONV-test-file-location (Yes), TQ-mock-scope (Partial). All others: No.

**Test ID: CLI-LISTRULES-04: Category grouping**

- **Input:**
```bash
stricture list-rules
```
- **Expected:** Rules grouped by category with category headers:
```
Test Quality (TQ) — 10 rules
  TQ-no-shallow-assertions  ...
  TQ-return-type-verified   ...
  ...

Architecture (ARCH) — 6 rules
  ARCH-dependency-direction ...
  ...

Convention (CONV) — 6 rules
  CONV-file-naming          ...
  ...

Contract (CTR) — 8 rules
  CTR-request-shape         ...
  ...
```
- **Verification:** Each category has the correct count. All rules appear under their correct category.

**Test ID: CLI-LISTRULES-05: Rule descriptions match product spec**

- **Input:**
```bash
stricture list-rules
```
- **Verification (spot check):**
  - `TQ-no-shallow-assertions` description contains "existence" or "truthiness"
  - `ARCH-dependency-direction` description contains "unidirectional"
  - `CTR-strictness-parity` description contains "producer and consumer" or "same constraints"
  - `CONV-file-header` description contains "header"

**Test ID: CLI-LISTRULES-06: Manifest-required rules indicated**

- **Input:**
```bash
stricture list-rules
```
- **Expected:** `CTR-strictness-parity` and `CTR-manifest-conformance` are marked as requiring a manifest (e.g., `(requires manifest)` annotation).
- **Verification:** Both rules have manifest requirement noted.

**Test ID: CLI-LISTRULES-07: Output is valid for piping/parsing**

- **Input:**
```bash
stricture list-rules | wc -l
```
- **Expected:** Consistent line count. No ANSI codes when piped.
- **Verification:** Line count is stable across runs. No `\033[` sequences in piped output.

---

## 8. Meta Flags (Cross-Command)

These flags apply to all commands.

**Test ID: CLI-META-01: --version on any subcommand**

- **Input:**
```bash
stricture lint --version
stricture audit --version
stricture trace --version
```
- **Expected:** All print the same version string.
- **Verification:** Output identical across all commands.

**Test ID: CLI-META-02: --help on each subcommand shows relevant options**

- **Input:**
```bash
stricture lint --help
stricture fix --help
stricture audit --help
stricture trace --help
stricture init --help
stricture inspect --help
stricture list-rules --help
```
- **Expected:** Each shows command-specific help. `audit --help` shows `--manifest`, `--service`, `--remote`, `--strictness`. `trace --help` shows `--trace-format`, `--strict`. Generic lint help shows filtering, targeting, output, and performance options.
- **Verification:** Each help output contains the flags specific to that command.

**Test ID: CLI-META-03: No subcommand defaults to lint**

- **Input:**
```bash
stricture src/
```
- **Expected:** Equivalent to `stricture lint src/`. Same output, same exit code.
- **Verification:** Compare output of `stricture src/` and `stricture lint src/` (must match).

**Test ID: CLI-META-04: Unknown subcommand produces error**

- **Input:**
```bash
stricture foobar src/
```
- **Expected:** Exit code 2. Error message: unknown command `foobar`. Lists valid commands.
- **Verification:** Stderr contains `foobar` and lists all 7 commands.

**Test ID: CLI-META-05: Stdin pipe (no TTY) detected for color default**

- **Input:**
```bash
stricture src/ | cat  # piped, no TTY
```
- **Expected:** No ANSI color codes in output (auto-detected non-TTY).
- **Verification:** Output does not contain `\033[` escape sequences.

---

## 9. Edge Cases & Error Handling

**Test ID: CLI-EDGE-01: Very long file path**

- **Input:** File path exceeding 255 characters.
- **Expected:** Either handles correctly or produces a clear OS-level error. No crash, no truncated path in output.
- **Verification:** Error message (if any) is meaningful.

**Test ID: CLI-EDGE-02: File with no newline at end**

- **Input:**
```bash
stricture src/no-trailing-newline.ts
```
- **Setup:** File does not end with a newline character.
- **Expected:** Parsed correctly. No parse errors. Line numbers correct.
- **Verification:** Violations (if any) have correct line numbers.

**Test ID: CLI-EDGE-03: File with Windows line endings (CRLF)**

- **Input:**
```bash
stricture src/windows-file.ts
```
- **Setup:** File uses `\r\n` line endings.
- **Expected:** Parsed correctly. Line numbers correct. No phantom violations from CRLF handling.
- **Verification:** Line numbers match expected. No CRLF-related false positives.

**Test ID: CLI-EDGE-04: File with mixed line endings**

- **Input:** File with some `\n` and some `\r\n` lines.
- **Expected:** Parsed without crash. Line numbers may be slightly off but no fatal error.
- **Verification:** No crash. Reasonable violation output.

**Test ID: CLI-EDGE-05: Very large file (50,000 lines)**

- **Input:**
```bash
stricture huge-file.ts
```
- **Expected:** Processes without OOM. ARCH-max-file-lines violation reported (if configured < 50000). Completes within reasonable time.
- **Verification:** Process completes. Memory stays under 500MB for single file.

**Test ID: CLI-EDGE-06: Concurrent runs on same project**

- **Input:**
```bash
stricture src/ &
stricture src/ &
wait
```
- **Expected:** Both runs complete without corrupting cache or producing incorrect results. No file locking errors.
- **Verification:** Both produce identical violation sets. Cache not corrupted.

**Test ID: CLI-EDGE-07: Config file with UTF-8 BOM**

- **Input:** `.stricture.yml` starts with UTF-8 BOM (`EF BB BF`).
- **Expected:** Config parsed correctly. BOM does not cause YAML parse error.
- **Verification:** Config loads without error. Rules applied correctly.

**Test ID: CLI-EDGE-08: Extremely deep directory nesting**

- **Input:**
```bash
stricture deeply/nested/path/with/many/levels/file.ts
```
- **Expected:** File found and linted. Path displayed correctly in output.
- **Verification:** Full path shown in violation output.

**Test ID: CLI-EDGE-09: Ctrl+C (SIGINT) during long run**

- **Input:** Start `stricture` on a large project, send SIGINT mid-run.
- **Expected:** Process exits promptly (< 1s). No orphan child processes. Cache not corrupted. Partial output may appear.
- **Verification:** Process exits cleanly. Subsequent runs work correctly.

**Test ID: CLI-EDGE-10: Disk full during --output write**

- **Input:**
```bash
stricture --format json --output /nearly-full-disk/results.json src/
```
- **Expected:** Error message about write failure. Exit code 2. No partial/corrupt file left behind (or clearly marked as incomplete).
- **Verification:** Error message is clear about the disk issue.

**Test ID: CLI-EDGE-11: Permission denied on output file**

- **Input:**
```bash
stricture --format json --output /root/results.json src/
```
- **Expected:** Exit code 2. Error message about permission denied on output path.
- **Verification:** Error message includes the path and "permission denied" or equivalent.
