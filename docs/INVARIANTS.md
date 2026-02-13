# Stricture — Invariants

> **Properties that must ALWAYS hold, regardless of input, configuration, or platform.**
>
> If any invariant is violated, it's a bug. Every invariant has a corresponding test. LLMs implementing Stricture must verify these invariants are preserved by their changes.

---

## 1. Output Determinism

**Invariant:** Given the same input files and configuration, Stricture produces byte-identical output on every run, regardless of goroutine scheduling, OS, or hardware.

**Implementation:**
- Violations are sorted before output: file path (lexicographic) → line (ascending) → column (ascending) → rule ID (lexicographic)
- This sort happens AFTER all rules complete, BEFORE reporter formats output
- Worker pool results are collected via channel, then sorted — never assumed to be ordered

**Test:**
```go
func TestOutputDeterminism(t *testing.T) {
    // Run stricture 10 times on the same input
    // All 10 outputs must be byte-identical
    var outputs []string
    for i := 0; i < 10; i++ {
        output := runStricture(t, goldenInputDir, "--format", "json")
        outputs = append(outputs, output)
    }
    for i := 1; i < len(outputs); i++ {
        if outputs[i] != outputs[0] {
            t.Errorf("Run %d produced different output than run 0", i)
        }
    }
}
```

---

## 2. Parse Safety

**Invariant:** Stricture never panics, crashes, or hangs on any input file. Malformed, empty, binary, or adversarial files produce either a valid UnifiedFileModel or a graceful error.

**Implementation:**
- Every adapter wraps parsing in a `defer recover()` block
- Parse timeouts: 5s per file (configurable), after which the file is skipped with a diagnostic
- Binary file detection: if the first 512 bytes contain a null byte, skip as binary
- Empty files return an empty UnifiedFileModel (not nil, not error)

**Test:** Fuzz tests (`internal/adapter/goparser/fuzz_test.go`, etc.) verify this continuously.

---

## 3. Suppression Correctness

**Invariant:** `// stricture-disable-next-line RULE-ID` on line N prevents violations from RULE-ID on line N+1, and ONLY line N+1, and ONLY for that specific rule.

**Implementation:**
- Suppression comments are parsed BEFORE rule execution
- The engine builds a suppression map: `map[int]map[string]bool` (line → rule IDs → suppressed)
- After rule execution, violations are filtered against the suppression map
- A suppression comment that doesn't match any violation is itself a warning (`stricture-unused-suppression`)

**Tests:**
```go
// Every rule must have a suppression test:
{
    name: "suppression comment disables violation",
    file: &model.UnifiedFileModel{
        // Source contains "// stricture-disable-next-line RULE-ID" on the line before the violation
    },
    wantCount: 0,  // Suppressed
},
{
    name: "suppression for wrong rule does not suppress",
    file: &model.UnifiedFileModel{
        // Source contains "// stricture-disable-next-line OTHER-RULE" on the line before the violation
    },
    wantCount: 1,  // Not suppressed — wrong rule
},
```

---

## 4. Fix Idempotency

**Invariant:** Running `stricture --fix` on already-fixed code produces no changes. `stricture --fix && stricture --fix` is always a no-op on the second run.

**Implementation:**
- After applying fixes, re-run affected rules on the fixed file
- If any new violations appear, abort the fix and report an error
- The `--fix` operation writes to a temp file, re-checks it, then replaces the original

**Test:**
```go
func TestFixIdempotency(t *testing.T) {
    // Apply fix
    fixed := applyFix(t, input)
    // Apply fix again
    fixedAgain := applyFix(t, fixed)
    // Must be identical
    assert.Equal(t, fixed, fixedAgain, "fix is not idempotent")
    // Must have 0 violations
    violations := check(t, fixedAgain)
    assert.Empty(t, violations, "fixed code still has violations")
}
```

---

## 5. Adapter Parity

**Invariant:** Equivalent code in different languages produces UnifiedFileModels with the same semantic content. A function named `getUser` with two parameters must produce the same `FuncModel.Name`, `len(FuncModel.Params)`, and `FuncModel.Exported` values regardless of whether the source is Go, TypeScript, Python, or Java.

**Implementation:**
- Cross-adapter parity test fixtures: the same logical code in all 4 languages
- Tests compare UFM fields (not source-specific details like syntax)

**Test fixtures:**
```
tests/fixtures/adapter-parity/
  ├── get-user.go       # Go:   func GetUser(id string) (*User, error)
  ├── get-user.ts       # TS:   export async function getUser(id: string): Promise<User>
  ├── get-user.py       # Py:   def get_user(id: str) -> User:
  ├── GetUser.java      # Java: public User getUser(String id)
  └── expected.json     # Expected UFM fields (name, param count, exported, async)
```

---

## 6. No Silent Degradation

**Invariant:** If Stricture cannot fully analyze a file (parse error, unsupported syntax, timeout), it MUST report a diagnostic. It must NEVER silently skip a file and return 0 violations, because that would be a false negative.

**Implementation:**
- The engine tracks `skippedFiles []Diagnostic` alongside `violations`
- Skipped files are reported in the summary (JSON: `"skipped"` field, text: footer line)
- Exit code is still 0 if only warnings, but `--strict` mode exits 1 on any skip

---

## 7. Configuration Monotonicity

**Invariant:** Adding a rule to `.stricture.yml` (changing from `off` to `error`) can only increase the number of violations, never decrease them. Removing a rule can only decrease.

**Why:** This prevents configuration interactions where enabling rule A somehow suppresses rule B's violations.

**Test:**
```go
func TestConfigMonotonicity(t *testing.T) {
    // Run with all rules
    allViolations := runWithConfig(t, allRulesConfig)
    // Run with one rule disabled
    lessViolations := runWithConfig(t, allRulesMinusOne)
    // Violations from remaining rules must be identical
    assert.Subset(t, lessViolations, allViolations)
}
```

---

## 8. Cache Transparency

**Invariant:** Cached runs produce byte-identical output to uncached runs. The cache is a pure performance optimization with zero behavioral impact.

**Test:**
```go
func TestCacheTransparency(t *testing.T) {
    // Run without cache
    uncached := runStricture(t, dir, "--no-cache", "--format", "json")
    // Run with cache (populates cache)
    _ = runStricture(t, dir, "--format", "json")
    // Run with cache (uses cache)
    cached := runStricture(t, dir, "--format", "json")
    // Must be identical (ignoring elapsedMs)
    assert.JSONEq(t, uncached, cached)
}
```

---

## Invariant Test Locations

| Invariant | Test File |
|-----------|-----------|
| Output determinism | `tests/integration/determinism_test.go` |
| Parse safety | `internal/adapter/*/fuzz_test.go` |
| Suppression correctness | `internal/engine/suppression_test.go` |
| Fix idempotency | `tests/integration/fix_test.go` |
| Adapter parity | `tests/integration/adapter_parity_test.go` |
| No silent degradation | `internal/engine/engine_test.go` |
| Config monotonicity | `tests/integration/config_test.go` |
| Cache transparency | `tests/integration/cache_test.go` |
