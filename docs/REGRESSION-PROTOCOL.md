# Stricture — Regression Protocol

> **Every bug gets a fixture before it gets a fix.**
>
> When a bug is found — whether a false positive, false negative, crash, or wrong output — the protocol is always the same: reproduce it as a test fixture FIRST, then fix it. This ensures the bug can never return.

---

## Protocol

### Step 1: Reproduce

Create a minimal test fixture that triggers the bug:

```
tests/fixtures/{rule-id}/regression-{issue-number}/
  ├── input.{ts,go,py,java}    # Minimal source that triggers the bug
  ├── expected.json              # What Stricture SHOULD report
  └── README.md                  # One paragraph: what the bug is, why it happens
```

For crashes or parser bugs:

```
tests/fixtures/adapter-{lang}/regression-{issue-number}/
  ├── input.{ts,go,py,java}    # File that causes the crash
  ├── expected.json              # Expected output (may be "no violations" or "graceful error")
  └── README.md
```

### Step 2: Write the test

Add a test case to the rule's test file:

```go
{
    name: "regression #42: nested optional chain crashes classifier",
    file: loadFixture(t, "tq-no-shallow-assertions/regression-42/input.ts"),
    wantCount: 1,  // Should report a violation, not crash
    wantRuleID: "TQ-no-shallow-assertions",
},
```

**Verify the test FAILS before fixing the bug.** This confirms the reproduction is accurate.

### Step 3: Fix the bug

Implement the fix. The test should now pass.

### Step 4: Verify no regressions

```bash
make test          # All tests pass
make validate      # Validation set still passes
```

### Step 5: Document

Add to the rule's `expected.json`:

```json
{
  "regressions": [
    {
      "issue": 42,
      "description": "Nested optional chain (a?.b?.c) crashed assertion classifier",
      "fixed_in": "0.1.3"
    }
  ]
}
```

---

## Categories of Bugs

### False Positive (Stricture reports a violation that shouldn't exist)

**Fixture requirement:** The input file should be CORRECT code. `expected.json` should show 0 violations for this rule.

**Impact:** False positives destroy trust. Users add `// stricture-disable` comments, which defeats the purpose. Treat as P0.

### False Negative (Stricture misses a violation it should catch)

**Fixture requirement:** The input file should have a SPECIFIC bug. `expected.json` should show the expected violation with correct rule ID, line, and message.

**Impact:** False negatives mean bugs in production. Treat as P0 for level 1-3 bugs, P1 for level 4-5.

### Crash / Panic

**Fixture requirement:** The input file should trigger the crash. `expected.json` should show graceful handling (skip file with diagnostic, or report parse error).

**Impact:** Crashes abort CI pipelines. Treat as P0. This is why we fuzz test parsers.

### Wrong Output

**Fixture requirement:** The input file triggers a correct violation but with wrong line number, wrong message, or wrong severity. `expected.json` should show the correct output.

**Impact:** Wrong line numbers make violations hard to find. Treat as P1.

### Performance Regression

**Fixture requirement:** A benchmark file that demonstrates the regression. Include baseline timing.

**Impact:** Slow linters get disabled. Treat as P1 if > 20% regression.

---

## Regression Test Naming Convention

```
tests/fixtures/{rule-id}/regression-{issue-number}/
```

Examples:
```
tests/fixtures/tq-no-shallow-assertions/regression-42/
tests/fixtures/arch-no-circular-deps/regression-108/
tests/fixtures/adapter-typescript/regression-55/
tests/fixtures/config/regression-71/
```

---

## CI Enforcement

The CI pipeline runs ALL regression fixtures as part of `make test`. New regressions are detected by golden file comparison — if the output changes, the test fails.

Regression tests are never deleted. They accumulate over the life of the project. A rule with 50 regression tests is a well-tested rule.
