# Cross-Cutting Test Sections

Tests that span multiple rules, verify system-level behavior, and validate infrastructure.

---

## 29. Language Adapter Parity

### 29.1 Equivalent Code Produces Equivalent Violations

For each rule that applies to both Go and TypeScript, create a test case pair:

**Test ID: LAP-01 (TQ-no-shallow-assertions)**

- **TypeScript input:**
```typescript
it("creates user", () => { expect(createUser(input)).toBeDefined(); });
```
- **Go input:**
```go
func TestCreateUser(t *testing.T) {
    result := CreateUser(input)
    if result == nil { t.Fatal("nil") }
}
```
- **Expected:** Both produce TQ-no-shallow-assertions violations.

**Test ID: LAP-02 (ARCH-max-file-lines)**

- TypeScript file with 801 lines and Go file with 801 lines.
- **Expected:** Both produce ARCH-max-file-lines violations.

**Test ID: LAP-03 through LAP-20:** Repeat for all rules applicable to both languages: TQ-return-type-verified, TQ-error-path-coverage, TQ-boundary-tested, TQ-mock-scope, TQ-test-isolation, TQ-negative-cases, TQ-test-naming, ARCH-dependency-direction, ARCH-import-boundary, ARCH-no-circular-deps, ARCH-layer-violation, ARCH-module-boundary, CONV-file-header, CONV-error-format, CONV-export-naming, CTR-request-shape, CTR-response-shape, CTR-status-code-handling.

### 29.2 Go-Specific Rules Skipped for TS

**Test ID: LAP-GO-SKIP-01:** Run CTR-json-tag-match on a TypeScript file. Expected: rule is skipped, no violations, no errors.

### 29.3 TS-Specific Patterns Do Not Crash Go Adapter

**Test ID: LAP-TS-SAFE-01:** Pass a file with arrow functions, optional chaining, and nullish coalescing to the Go adapter. Expected: adapter returns an error or skips the file gracefully, no crash.

### 29.4 Adapter Error Handling

**Test ID: LAP-ERR-01:** Malformed Go file (syntax error). Expected: adapter reports parse error, other files still processed.
**Test ID: LAP-ERR-02:** TS file with syntax error. Expected: same behavior.
**Test ID: LAP-ERR-03:** Binary file passed to adapter. Expected: skipped.
**Test ID: LAP-ERR-04:** File with mixed encoding (Latin-1 in UTF-8 context). Expected: warning, skip.
**Test ID: LAP-ERR-05:** Empty file. Expected: valid UnifiedFileModel with empty arrays.

---

## 30. Config Resolution

### 30.1 Default Config Tests

**Test ID: CFG-01:** No `.stricture.yml` exists. Run `stricture src/`. Expected: all rules enabled at default settings, no errors.
**Test ID: CFG-02:** Empty `.stricture.yml` (0 bytes). Expected: defaults used.
**Test ID: CFG-03:** `.stricture.yml` with only `rules: {}`. Expected: all rules enabled with defaults.

### 30.2 Override Precedence

**Test ID: CFG-04:** CLI `--rule TQ-no-shallow-assertions` overrides config that disables it. Expected: rule runs.
**Test ID: CFG-05:** Config sets `TQ-no-shallow-assertions: off`. No CLI override. Expected: rule does not run.
**Test ID: CFG-06:** `extends` chain: base sets rule to "warn", local config sets to "error". Expected: "error" wins.
**Test ID: CFG-07:** Two-level extends: A extends B extends C. C sets rule to "off", B to "warn", A to "error". Expected: "error" (A wins).
**Test ID: CFG-08:** Three-level extends with option merging. Expected: most local option wins.

### 30.3 Error Cases

**Test ID: CFG-09:** Invalid YAML syntax (unclosed bracket). Expected: exit code 2, clear error message with line number.
**Test ID: CFG-10:** Unknown rule ID in config: `TQ-nonexistent: error`. Expected: warning, not fatal.
**Test ID: CFG-11:** Invalid option type: `maxShallowPercent: "fifty"`. Expected: exit code 2 with specific error.
**Test ID: CFG-12:** Circular extends: A extends B, B extends A. Expected: exit code 2, "circular extends" error.
**Test ID: CFG-13:** Extends target not found: `extends: ["./nonexistent.yml"]`. Expected: exit code 2.
**Test ID: CFG-14:** Config in subdirectory (`src/.stricture.yml`). Expected: not picked up (only project root).

### 30.4 Monorepo Config

**Test ID: CFG-15:** Monorepo with root config and workspace overrides. Expected: workspace config extends root.
**Test ID: CFG-16:** Multiple workspace configs with different rule settings. Expected: each workspace uses its own config.

---

## 31. Plugin System

### 31.1 Valid Plugin

**Test ID: PLG-01:** Plugin loads and runs correctly. Creates a violation. Expected: violation appears in output with plugin's rule ID.
**Test ID: PLG-02:** Plugin with options. Config passes `options: { logPrefix: "api" }`. Expected: plugin receives options.
**Test ID: PLG-03:** Plugin accesses ProjectContext. Expected: `context.files`, `context.dependencyGraph` are populated.

### 31.2 Error Handling

**Test ID: PLG-04:** Plugin with syntax error. Expected: clear error message naming the plugin file, other rules still run.
**Test ID: PLG-05:** Plugin that throws during `check()`. Expected: error reported for that plugin, process does not crash, other rules produce output.
**Test ID: PLG-06:** Plugin returns invalid violation shape (missing required fields). Expected: warning about invalid violation, violation is dropped.
**Test ID: PLG-07:** Plugin file not found. Expected: exit code 2 with file path in error.

### 31.3 Conflict Resolution

**Test ID: PLG-08:** Plugin with same ID as built-in rule (e.g., `TQ-no-shallow-assertions`). Expected: error at load time, process exits with code 2.
**Test ID: PLG-09:** Two plugins with same ID. Expected: error at load time.

### 31.4 Execution Order

**Test ID: PLG-10:** Built-in rules run, then plugins run. Verify by checking output order or timing.

### 31.5 TypeScript vs JavaScript Plugin

**Test ID: PLG-11:** JavaScript plugin (`.js`). Expected: loads and runs.
**Test ID: PLG-12:** TypeScript plugin (`.ts`). Expected: compiled and runs (or requires pre-compilation).

---

## 32. CLI Behavior

### 32.1 Targeting

**Test ID: CLI-01:** `stricture` with no args. Expected: lints current directory recursively.
**Test ID: CLI-02:** `stricture src/`. Expected: lints only files under src/.
**Test ID: CLI-03:** `stricture src/foo.ts src/bar.ts`. Expected: lints only those two files.
**Test ID: CLI-04:** `stricture --ext .ts src/`. Expected: only TypeScript files linted.
**Test ID: CLI-05:** `stricture --changed`. Expected: only files changed vs main branch. Verify with actual git repo.
**Test ID: CLI-06:** `stricture --staged`. Expected: only staged files.

### 32.2 Filtering

**Test ID: CLI-07:** `--rule TQ-no-shallow-assertions`. Expected: only that rule runs. Other rules produce no output.
**Test ID: CLI-08:** `--rule TQ-no-shallow-assertions --rule ARCH-max-file-lines`. Expected: both rules run, others do not.
**Test ID: CLI-09:** `--category TQ`. Expected: all TQ-* rules run.
**Test ID: CLI-10:** `--severity error`. Expected: only error-level violations shown, warnings hidden.

### 32.3 Output

**Test ID: CLI-11:** `--format text`. Expected: human-readable colored output.
**Test ID: CLI-12:** `--format json`. Expected: valid JSON output matching schema.
**Test ID: CLI-13:** `--format sarif`. Expected: valid SARIF 2.1.0 output.
**Test ID: CLI-14:** `--format junit`. Expected: valid JUnit XML output.
**Test ID: CLI-15:** `--output report.json`. Expected: output written to file, nothing on stdout.
**Test ID: CLI-16:** `--quiet`. Expected: warnings suppressed from output.
**Test ID: CLI-17:** `--verbose`. Expected: timing info for each rule shown.
**Test ID: CLI-18:** `--no-color`. Expected: no ANSI escape codes in output.

### 32.4 Fix

**Test ID: CLI-19:** `--fix`. Expected: fixable violations are fixed in-place.
**Test ID: CLI-20:** `--fix-dry-run`. Expected: diff output shown, no files modified.
**Test ID: CLI-21:** `--fix --rule CONV-file-header`. Expected: only file header fixes applied.

### 32.5 Config

**Test ID: CLI-22:** `--config custom.yml`. Expected: uses specified config file.
**Test ID: CLI-23:** `--no-config`. Expected: ignores .stricture.yml, uses built-in defaults.

### 32.6 Performance

**Test ID: CLI-24:** `--concurrency 1`. Expected: runs single-threaded.
**Test ID: CLI-25:** `--cache` then `--no-cache`. Expected: second run is slower without cache.

### 32.7 Meta

**Test ID: CLI-26:** `--version`. Expected: prints version string, exits 0.
**Test ID: CLI-27:** `--help`. Expected: prints help text, exits 0.
**Test ID: CLI-28:** Invalid flag `--nonexistent`. Expected: exit code 2 with error message.

### 32.8 Exit Codes

**Test ID: CLI-29:** No errors, some warnings. Expected: exit code 0.
**Test ID: CLI-30:** One or more errors. Expected: exit code 1.
**Test ID: CLI-31:** Invalid config. Expected: exit code 2.
**Test ID: CLI-32:** No files match. Expected: exit code 0, "no files to lint" message.
**Test ID: CLI-33:** Parse error on input file. Expected: warning per file, other files still processed. Exit code 1 only if violations found, exit code 0 otherwise. Exit code 2 only if ALL files are unparseable.

---

## 33. Output Format Correctness

### 33.1 JSON Format

**Test ID: OUT-JSON-01:** Validate output against JSON schema. All required fields present.
**Test ID: OUT-JSON-02:** `version` field matches tool version.
**Test ID: OUT-JSON-03:** `timestamp` is valid ISO 8601.
**Test ID: OUT-JSON-04:** `summary.filesChecked` matches actual file count.
**Test ID: OUT-JSON-05:** `violations` array contains objects with: ruleId, severity, category, file, line, column, message, why, suggestion, suppress, fixable.
**Test ID: OUT-JSON-06:** `context` object has rule-specific extra data.
**Test ID: OUT-JSON-07:** Empty violations array when no issues found.
**Test ID: OUT-JSON-08:** Large violation set (1000+) is valid JSON.

### 33.2 SARIF Format

**Test ID: OUT-SARIF-01:** Validate against SARIF 2.1.0 JSON Schema.
**Test ID: OUT-SARIF-02:** `tool.driver.name` is "Stricture".
**Test ID: OUT-SARIF-03:** `tool.driver.rules` contains all active rules with descriptions.
**Test ID: OUT-SARIF-04:** Each result has `physicalLocation` with file, line, column.
**Test ID: OUT-SARIF-05:** Fixable violations have `fix` objects.
**Test ID: OUT-SARIF-06:** Cross-file violations (ARCH rules) have `codeFlows`.
**Test ID: OUT-SARIF-07:** Empty results when no violations.

### 33.3 JUnit Format

**Test ID: OUT-JUNIT-01:** Validate against JUnit XML schema.
**Test ID: OUT-JUNIT-02:** Each rule is a `<testsuite>`.
**Test ID: OUT-JUNIT-03:** Each checked file is a `<testcase>`.
**Test ID: OUT-JUNIT-04:** Violations are `<failure>` elements with message and type.
**Test ID: OUT-JUNIT-05:** Clean files are `<testcase>` without `<failure>`.
**Test ID: OUT-JUNIT-06:** Summary attributes (tests, failures, errors, time) are correct.

### 33.4 Text Format

**Test ID: OUT-TEXT-01:** Colors render correctly with TTY mock.
**Test ID: OUT-TEXT-02:** `--no-color` strips all ANSI codes.
**Test ID: OUT-TEXT-03:** File paths are relative to project root.
**Test ID: OUT-TEXT-04:** Line numbers are correct (match actual source lines).
**Test ID: OUT-TEXT-05:** Summary line at end: "N errors, M warnings in F files (T files checked, Xs)".
**Test ID: OUT-TEXT-06:** Each violation includes: rule ID, severity, message, suggestion, why, suppress syntax.

---

## 34. Auto-Fix Safety

**Test ID: FIX-01:** `--fix` modifies files correctly for each fixable rule (CONV-file-header, CONV-file-naming, CONV-export-naming, TQ-mock-scope).
**Test ID: FIX-02:** `--fix` is idempotent. Run twice, second run produces no changes.
**Test ID: FIX-03:** `--fix` never modifies files for non-fixable rules (TQ-no-shallow-assertions, ARCH-*).
**Test ID: FIX-04:** `--fix-dry-run` produces diff output but does NOT write files. Verify via file mtime comparison before and after.
**Test ID: FIX-05:** `--fix` on read-only file. Expected: error message, no crash, other files still fixed.
**Test ID: FIX-06:** Multiple fixable violations in same file. All applied correctly, no conflicts.
**Test ID: FIX-07:** `--fix` preserves file encoding (UTF-8 BOM, LF vs CRLF line endings).
**Test ID: FIX-08:** `--fix --rule CONV-file-header`. Only header fixes applied, other fixable rules skipped.
**Test ID: FIX-09:** After fix, re-run stricture on fixed files. Expected: no violations for fixed rules.
**Test ID: FIX-10:** CONV-file-naming fix renames file AND updates all imports referencing the old name.
**Test ID: FIX-11:** CONV-export-naming fix renames symbol AND updates all references across the project.
**Test ID: FIX-12:** TQ-mock-scope fix adds `afterEach(() => { jest.restoreAllMocks(); })` -- verify correct placement.

---

## 35. Performance

**Test ID: PERF-01:** 100 files: must complete in < 1s.
**Test ID: PERF-02:** 500 files: must complete in < 3s.
**Test ID: PERF-03:** 1000 files: must complete in < 6s.
**Test ID: PERF-04:** 10,000 files: must complete in < 60s and use < 500MB RAM.
**Test ID: PERF-05:** Cache: second run on unchanged files is >2x faster than first.
**Test ID: PERF-06:** Cache invalidation: change one file, re-run. Only that file is re-parsed. Other files use cache.
**Test ID: PERF-07:** Pathological input: file with 10,000-line function. No hang or crash.
**Test ID: PERF-08:** Pathological input: 50-level nested type. No stack overflow.
**Test ID: PERF-09:** Pathological input: 500 imports in one file. No significant slowdown.
**Test ID: PERF-10:** Concurrent access: two stricture processes on same repo. No cache corruption. Both produce correct results.
**Test ID: PERF-11:** `--concurrency 1` is slower than default but produces identical results.
**Test ID: PERF-12:** Memory does not grow unboundedly with file count (no memory leak).

---

## 36. Contract Detection Heuristics

**Test ID: CDH-01:** Express.js route + fetch() client in same repo. Expected: correctly paired.
**Test ID: CDH-02:** Go Chi route + http.Get client. Expected: correctly paired.
**Test ID: CDH-03:** Fastify route + axios client. Expected: correctly paired.
**Test ID: CDH-04:** Route with path parameters (`/users/:id`) matches client URL template (`/users/${id}`).
**Test ID: CDH-05:** Route with query parameters handled correctly.
**Test ID: CDH-06:** Multiple handlers for same path, different HTTP methods. Expected: distinct pairs.
**Test ID: CDH-07:** Client calling external API (https://api.stripe.com). Expected: NOT paired with any local server.
**Test ID: CDH-08:** Server with no matching client in repo. Expected: no contract pair, no violation.
**Test ID: CDH-09:** Annotation-based pairing (`// stricture-contract: server=... client=...`). Expected: pair created.
**Test ID: CDH-10:** OpenAPI spec-based pairing. Expected: both client and server validated against spec.
**Test ID: CDH-11:** Monorepo: server in `packages/api`, client in `packages/web`. Expected: still detected across packages.
**Test ID: CDH-12:** Express Router sub-routes (`router.get("/", ...)` mounted at `/api/users`). Expected: full path resolved.
**Test ID: CDH-13:** Go http.HandleFunc with string concatenation in path. Expected: best-effort matching.
**Test ID: CDH-14:** Client using environment variable for base URL (`${BASE_URL}/api/users`). Expected: path portion still matched.
**Test ID: CDH-15:** Multiple clients for same server endpoint. Expected: each client paired separately.

---

## 37. Rule Interaction

**Test ID: RI-01:** Multiple rules flag same line. Expected: all violations reported, none suppressed by another rule.
**Test ID: RI-02:** TQ rules + CTR rules on same file. Expected: both categories run, no interference.
**Test ID: RI-03:** ARCH-max-file-lines triggers on a test file. TQ rules still run on same file.
**Test ID: RI-04:** Disable entire category via config (`TQ-*: off` or category-level setting). Other categories unaffected.
**Test ID: RI-05:** Plugin rule + built-in rule on same file. Both run and report.
**Test ID: RI-06:** Inline suppression of one rule does not suppress another rule on the same line.
**Test ID: RI-07:** `--rule` flag limits execution. Suppressed rules do not run at all (not run then filtered).
**Test ID: RI-08:** `--severity error` filters output but all rules still execute (warnings computed but not shown).
**Test ID: RI-09:** Rule that requires ProjectContext (ARCH) does not block per-file rules (CONV) from running.
**Test ID: RI-10:** If one rule throws an error, other rules on the same file still run.

---

## 38. Error Handling

**Test ID: ERR-01:** Binary file (.png, .exe) in target directory. Expected: skipped, no crash, no violation.
**Test ID: ERR-02:** Empty file (0 bytes). Expected: skipped or minimal violations (CONV-file-header only).
**Test ID: ERR-03:** File with only comments. Expected: handled gracefully. CONV-file-header may flag if header format is wrong.
**Test ID: ERR-04:** File with BOM marker (UTF-8 BOM). Expected: parsed correctly, BOM does not affect rule behavior.
**Test ID: ERR-05:** File with mixed line endings (CRLF + LF). Expected: parsed correctly, line numbers accurate.
**Test ID: ERR-06:** Symlinked file. Expected: follows symlink, reports on real path.
**Test ID: ERR-07:** Permission denied on file. Expected: skip with warning, do not exit.
**Test ID: ERR-08:** Directory with no matching files. Expected: "no files to lint", exit 0.
**Test ID: ERR-09:** File disappears between glob and read (race condition). Expected: skip with warning.
**Test ID: ERR-10:** Extremely long line (>10,000 characters). Expected: no crash or hang.
**Test ID: ERR-11:** File with null bytes. Expected: skip with warning.
**Test ID: ERR-12:** Circular symlinks. Expected: detect and skip, no infinite loop.
**Test ID: ERR-13:** Non-UTF-8 encoding (Latin-1 file). Expected: skip with warning.
**Test ID: ERR-14:** Directory named `foo.ts` (not a file). Expected: skipped.
**Test ID: ERR-15:** File path with special characters (spaces, unicode, emoji).

---

