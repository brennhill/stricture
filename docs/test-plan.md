# Stricture — Comprehensive Test Plan

> **Version:** 0.1.0
> **Date:** 2026-02-12
> **Status:** Test Plan (pre-implementation)
> **Companion Document:** [Product Spec](./product-spec.md)

This test plan is designed to be executed mechanically by any developer, regardless of familiarity with the product spec. Every test case includes exact inputs, expected outputs, and verification methods.

---

## Table of Contents

- [PART 1: Per-Rule Test Matrices](#part-1-per-rule-test-matrices)
  - [1. TQ-no-shallow-assertions](#1-tq-no-shallow-assertions)
  - [2. TQ-return-type-verified](#2-tq-return-type-verified)
  - [3. TQ-schema-conformance](#3-tq-schema-conformance)
  - [4. TQ-error-path-coverage](#4-tq-error-path-coverage)
  - [5. TQ-assertion-depth](#5-tq-assertion-depth)
  - [6. TQ-boundary-tested](#6-tq-boundary-tested)
  - [7. TQ-mock-scope](#7-tq-mock-scope)
  - [8. TQ-test-isolation](#8-tq-test-isolation)
  - [9. TQ-negative-cases](#9-tq-negative-cases)
  - [10. TQ-test-naming](#10-tq-test-naming)
  - [11. ARCH-dependency-direction](#11-arch-dependency-direction)
  - [12. ARCH-import-boundary](#12-arch-import-boundary)
  - [13. ARCH-no-circular-deps](#13-arch-no-circular-deps)
  - [14. ARCH-max-file-lines](#14-arch-max-file-lines)
  - [15. ARCH-layer-violation](#15-arch-layer-violation)
  - [16. ARCH-module-boundary](#16-arch-module-boundary)
  - [17. CONV-file-naming](#17-conv-file-naming)
  - [18. CONV-file-header](#18-conv-file-header)
  - [19. CONV-error-format](#19-conv-error-format)
  - [20. CONV-export-naming](#20-conv-export-naming)
  - [21. CONV-test-file-location](#21-conv-test-file-location)
  - [22. CONV-required-exports](#22-conv-required-exports)
  - [23. CTR-request-shape](#23-ctr-request-shape)
  - [24. CTR-response-shape](#24-ctr-response-shape)
  - [25. CTR-status-code-handling](#25-ctr-status-code-handling)
  - [26. CTR-shared-type-sync](#26-ctr-shared-type-sync)
  - [27. CTR-json-tag-match](#27-ctr-json-tag-match)
  - [28. CTR-dual-test](#28-ctr-dual-test)
- [PART 2: Cross-Cutting Test Sections](#part-2-cross-cutting-test-sections)
  - [29. Language Adapter Parity](#29-language-adapter-parity)
  - [30. Config Resolution](#30-config-resolution)
  - [31. Plugin System](#31-plugin-system)
  - [32. CLI Behavior](#32-cli-behavior)
  - [33. Output Format Correctness](#33-output-format-correctness)
  - [34. Auto-Fix Safety](#34-auto-fix-safety)
  - [35. Performance](#35-performance)
  - [36. Contract Detection Heuristics](#36-contract-detection-heuristics)
  - [37. Rule Interaction](#37-rule-interaction)
  - [38. Error Handling](#38-error-handling)
- [PART 3: Spec Gap Analysis](#part-3-spec-gap-analysis)

---

# PART 1: Per-Rule Test Matrices

---

## 1. TQ-no-shallow-assertions

**Rule purpose:** Reject assertions that verify existence/truthiness without checking actual values.

### 1.1 True Positive Cases

**TP-NSA-01: toBeDefined on structured return value**

- **Input file** (`tests/user.test.ts`):
```typescript
import { createUser } from "../src/user";

it("creates user", () => {
  const result = createUser({ name: "Alice", email: "a@b.com" });
  expect(result).toBeDefined();
});
```
- **Source file** (`src/user.ts`):
```typescript
interface User { id: number; name: string; email: string; }
export function createUser(input: { name: string; email: string }): User {
  return { id: 1, name: input.name, email: input.email };
}
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 5 of `tests/user.test.ts`. Message: `expect(result).toBeDefined() -- shallow assertion on typed return value`
- **Why:** The function returns a `User` with 3 fields. Checking `.toBeDefined()` verifies nothing about correctness.

**TP-NSA-02: toBeTruthy on return value**

- **Input file** (`tests/order.test.ts`):
```typescript
import { getOrder } from "../src/order";

it("gets order", () => {
  const order = getOrder("abc");
  expect(order).toBeTruthy();
});
```
- **Source file** (`src/order.ts`):
```typescript
interface Order { id: string; total: number; items: string[]; }
export function getOrder(id: string): Order { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 5. Message: shallow assertion on typed return value.
- **Why:** `toBeTruthy()` only verifies the value is not null/undefined/0/false/"". It does not check any field.

**TP-NSA-03: not.toBeNull on return value**

- **Input file** (`tests/config.test.ts`):
```typescript
import { loadConfig } from "../src/config";

it("loads config", () => {
  const cfg = loadConfig();
  expect(cfg).not.toBeNull();
});
```
- **Source file** (`src/config.ts`):
```typescript
interface Config { port: number; host: string; debug: boolean; }
export function loadConfig(): Config { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 5.
- **Why:** `.not.toBeNull()` only verifies non-null. Tells nothing about the config shape.

**TP-NSA-04: toBeInstanceOf(Object)**

- **Input file** (`tests/parser.test.ts`):
```typescript
import { parse } from "../src/parser";

it("parses data", () => {
  const result = parse("some input");
  expect(result).toBeInstanceOf(Object);
});
```
- **Source file** (`src/parser.ts`):
```typescript
interface ParseResult { tokens: string[]; errors: string[]; }
export function parse(input: string): ParseResult { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 5.
- **Why:** `toBeInstanceOf(Object)` is true for any non-null object. It proves nothing about the result's shape.

**TP-NSA-05: typeof check without value check**

- **Input file** (`tests/math.test.ts`):
```typescript
import { calculate } from "../src/math";

it("calculates", () => {
  const result = calculate(1, 2);
  expect(typeof result).toBe("object");
});
```
- **Source file** (`src/math.ts`):
```typescript
interface CalcResult { sum: number; product: number; }
export function calculate(a: number, b: number): CalcResult { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 5.
- **Why:** Checking `typeof` is `"object"` is a type-identity check, not a value assertion.

**TP-NSA-06: Array.isArray without content check**

- **Input file** (`tests/list.test.ts`):
```typescript
import { getItems } from "../src/list";

it("gets items", () => {
  const items = getItems();
  expect(Array.isArray(items)).toBe(true);
});
```
- **Source file** (`src/list.ts`):
```typescript
interface Item { id: number; name: string; }
export function getItems(): Item[] { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 5.
- **Why:** `Array.isArray` verifies the value is an array, but an empty array or array of wrong types passes.

**TP-NSA-07: Go nil-only check**

- **Input file** (`user_test.go`):
```go
package user

import "testing"

func TestCreateUser(t *testing.T) {
    result := CreateUser("Alice", "a@b.com")
    if result == nil {
        t.Fatal("result is nil")
    }
}
```
- **Source file** (`user.go`):
```go
package user

type User struct {
    ID    int
    Name  string
    Email string
}

func CreateUser(name, email string) *User {
    return &User{ID: 1, Name: name, Email: email}
}
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 7 of `user_test.go`. Message: nil check without value assertion on `*User` return type.
- **Why:** Only checks `result != nil`. Never verifies ID, Name, or Email.

**TP-NSA-08: Go error-only check without error type/message**

- **Input file** (`api_test.go`):
```go
package api

import "testing"

func TestFetchData(t *testing.T) {
    _, err := FetchData("key1")
    if err != nil {
        t.Fatal(err)
    }
}
```
- **Source file** (`api.go`):
```go
package api

import "fmt"

func FetchData(key string) (string, error) {
    if key == "" {
        return "", fmt.Errorf("FetchData: empty key")
    }
    return "value", nil
}
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 7 of `api_test.go`.
- **Why:** The test checks `err != nil` but never verifies the error type, message, or the actual return value.

**TP-NSA-09: Multiple shallow assertions in one test**

- **Input file** (`tests/service.test.ts`):
```typescript
import { processData } from "../src/service";

it("processes data", () => {
  const result = processData([1, 2, 3]);
  expect(result).toBeDefined();
  expect(result).toBeTruthy();
  expect(result).not.toBeNull();
  expect(result).not.toBeUndefined();
});
```
- **Source file** (`src/service.ts`):
```typescript
interface ProcessResult { total: number; items: number[]; status: string; }
export function processData(input: number[]): ProcessResult { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at lines 5, 6, 7, 8.
- **Why:** Four assertions, all shallow. None checks `total`, `items`, or `status`.

**TP-NSA-10: Shallow assertion buried among deep ones**

- **Input file** (`tests/report.test.ts`):
```typescript
import { generateReport } from "../src/report";

it("generates report", () => {
  const report = generateReport();
  expect(report.title).toBe("Q4 Report");
  expect(report.data).toBeDefined();
  expect(report.author).toBe("System");
});
```
- **Source file** (`src/report.ts`):
```typescript
interface Report { title: string; data: ReportData; author: string; }
interface ReportData { rows: number; columns: string[]; }
export function generateReport(): Report { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 6 (`expect(report.data).toBeDefined()`).
- **Why:** `report.data` is a complex type (`ReportData`). `.toBeDefined()` does not verify its content. The other assertions are deep and fine.

**TP-NSA-11: Vitest shallow assertion**

- **Input file** (`tests/helper.test.ts`):
```typescript
import { describe, it, expect } from "vitest";
import { transform } from "../src/helper";

it("transforms input", () => {
  const result = transform("hello");
  expect(result).toBeDefined();
});
```
- **Source file** (`src/helper.ts`):
```typescript
interface TransformResult { output: string; length: number; }
export function transform(input: string): TransformResult { /* ... */ }
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 6. Rule must detect Vitest assertions too.
- **Why:** Same shallow pattern, different test framework.

**TP-NSA-12: testify assert.NotNil in Go**

- **Input file** (`handler_test.go`):
```go
package handler

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
    result := Handle("input")
    assert.NotNil(t, result)
}
```
- **Source file** (`handler.go`):
```go
package handler

type Response struct {
    Code    int
    Message string
}

func Handle(input string) *Response {
    return &Response{Code: 200, Message: "OK"}
}
```
- **Expected violation:** `TQ-no-shallow-assertions` at line 10. Message: `assert.NotNil` is a shallow assertion on `*Response`.
- **Why:** `assert.NotNil` is the testify equivalent of a nil check.

### 1.2 True Negative Cases

**TN-NSA-01: Deep value assertion**

- **Input file** (`tests/user.test.ts`):
```typescript
import { createUser } from "../src/user";

it("creates user with correct fields", () => {
  const user = createUser({ name: "Alice", email: "a@b.com" });
  expect(user.id).toBe(1);
  expect(user.name).toBe("Alice");
  expect(user.email).toBe("a@b.com");
});
```
- **Expected:** No violation.
- **Why:** All assertions check specific field values.

**TN-NSA-02: toEqual with full object**

- **Input file** (`tests/user.test.ts`):
```typescript
it("creates user", () => {
  const user = createUser({ name: "Alice", email: "a@b.com" });
  expect(user).toEqual({ id: 1, name: "Alice", email: "a@b.com" });
});
```
- **Expected:** No violation.
- **Why:** `toEqual` with a full object is a deep assertion.

**TN-NSA-03: toHaveLength followed by element checks**

- **Input file** (`tests/list.test.ts`):
```typescript
it("returns items", () => {
  const items = getItems();
  expect(items).toHaveLength(3);
  expect(items[0].name).toBe("first");
});
```
- **Expected:** No violation.
- **Why:** `toHaveLength` is a value assertion (checks count), and elements are inspected.

**TN-NSA-04: toBeDefined in presence test with other deep tests in file**

- **Input file** (`tests/optional.test.ts`):
```typescript
it("should exist when feature flag is on", () => {
  const widget = getWidget({ featureFlag: true });
  expect(widget).toBeDefined();
});

it("has correct properties", () => {
  const widget = getWidget({ featureFlag: true });
  expect(widget.name).toBe("Sparkle");
  expect(widget.size).toBe(42);
});
```
- **Expected:** No violation (with `allowInPresenceTests: true` default).
- **Why:** Test name contains "exist", and other tests in the file assert deep values.

**TN-NSA-05: Go deep field assertion**

- **Input file** (`user_test.go`):
```go
func TestCreateUser(t *testing.T) {
    user := CreateUser("Alice", "a@b.com")
    if user.ID != 1 {
        t.Errorf("expected ID 1, got %d", user.ID)
    }
    if user.Name != "Alice" {
        t.Errorf("expected name Alice, got %s", user.Name)
    }
}
```
- **Expected:** No violation.
- **Why:** Checks specific field values, not just nil.

**TN-NSA-06: Go errors.Is check**

- **Input file** (`api_test.go`):
```go
func TestFetchData_EmptyKey(t *testing.T) {
    _, err := FetchData("")
    if !errors.Is(err, ErrEmptyKey) {
        t.Errorf("expected ErrEmptyKey, got %v", err)
    }
}
```
- **Expected:** No violation.
- **Why:** `errors.Is` checks a specific sentinel error, which is a deep assertion.

**TN-NSA-07: Primitive return type with toBeDefined**

- **Input file** (`tests/counter.test.ts`):
```typescript
it("counts items", () => {
  const count = countItems([1, 2, 3]);
  expect(count).toBe(3);
});
```
- **Source:** `function countItems(arr: number[]): number`
- **Expected:** No violation.
- **Why:** `.toBe(3)` is a deep assertion on a primitive.

### 1.3 False Positive Risks

**FP-NSA-01: toBeDefined on a genuinely optional value in a presence-focused test**

- **Pattern:**
```typescript
it("should be defined when feature is enabled", () => {
  const result = getOptionalFeature(true);
  expect(result).toBeDefined();
});
```
- **Why NOT a violation:** The test's explicit purpose is to verify the value exists (it can legitimately be undefined). The test name signals presence-checking intent.
- **Detection avoidance:** If `allowInPresenceTests: true` (default) and test name matches presence keywords ("exist", "present", "defined", "available"), AND other tests in the same file assert deep values on the same function, do not flag.

**FP-NSA-02: toBeDefined on the return of a void/undefined-returning function**

- **Pattern:**
```typescript
it("initializes without error", () => {
  const result = initialize();
  expect(result).toBeUndefined();
});
```
- **Why NOT a violation:** The function returns `void`/`undefined`. There is no deeper value to check.
- **Detection avoidance:** Resolve the function's return type. If return type is `void`, `undefined`, or `Promise<void>`, skip this rule for assertions on that return value.

**FP-NSA-03: toBeTruthy on a boolean return**

- **Pattern:**
```typescript
it("validates email", () => {
  expect(isValidEmail("a@b.com")).toBeTruthy();
});
```
- **Why NOT a violation:** The function returns `boolean`. `.toBeTruthy()` is a reasonable (if slightly imprecise) assertion. `.toBe(true)` would be better but `.toBeTruthy()` is acceptable for booleans.
- **Detection avoidance:** If return type is `boolean`, `.toBeTruthy()` and `.toBeFalsy()` are not shallow.

**FP-NSA-04: Assertion on an intermediate value, not the return value**

- **Pattern:**
```typescript
it("processes items", () => {
  const spy = jest.fn();
  processItems([1, 2, 3], spy);
  expect(spy).toHaveBeenCalled();
  expect(spy).toHaveBeenCalledWith(expect.any(Number));
});
```
- **Why NOT a violation:** The assertions are on a spy/mock, not on the return value. `toHaveBeenCalled` is appropriate for verifying side effects.
- **Detection avoidance:** Distinguish between assertions on the function's return value vs. assertions on mocks/spies. Only flag shallow assertions on the return value of the function under test.

**FP-NSA-05: toBeDefined as guard before deep assertions**

- **Pattern:**
```typescript
it("gets user", () => {
  const user = getUser("123");
  expect(user).toBeDefined();
  expect(user!.name).toBe("Alice");
  expect(user!.email).toBe("a@b.com");
});
```
- **Why NOT a violation:** The `toBeDefined()` serves as a guard assertion (preventing the test from hitting a null reference on the next line). Deep assertions follow.
- **Detection avoidance:** If a shallow assertion on variable `X` is immediately followed (within the same test block) by deep assertions on `X.field`, the shallow assertion is a guard and should not be flagged.

### 1.4 False Negative Risks

**FN-NSA-01: Snapshot assertion hiding shallowness**

- **Pattern:**
```typescript
it("creates user", () => {
  const user = createUser(input);
  expect(user).toMatchSnapshot();
});
```
- **Why it might be missed:** `toMatchSnapshot()` looks like a deep assertion because it captures the full object. However, on first run it accepts anything, and it is brittle without verifying intent.
- **Additional detection needed:** Flag `toMatchSnapshot()` as shallow unless the rule explicitly allows snapshots (option: `allowSnapshot: false` by default).
- **Severity:** Medium. Snapshots can mask meaningless tests.

**FN-NSA-02: Deep-looking assertion with any-matcher**

- **Pattern:**
```typescript
it("creates user", () => {
  const user = createUser(input);
  expect(user).toEqual(expect.anything());
});
```
- **Why it might be missed:** `toEqual` is normally deep, but `expect.anything()` matches anything. The assertion is semantically shallow.
- **Additional detection needed:** Detect `expect.anything()`, `expect.any(Object)` as shallow matchers inside `toEqual`/`toMatchObject`.
- **Severity:** High. This pattern appears deep but tests nothing.

**FN-NSA-03: Assertion through a wrapper function**

- **Pattern:**
```typescript
function assertExists(val: unknown) {
  expect(val).toBeDefined();
}

it("creates user", () => {
  const user = createUser(input);
  assertExists(user);
});
```
- **Why it might be missed:** The shallow assertion is inside a helper function, not directly in the test body. Static analysis of the test body alone sees a function call, not `toBeDefined()`.
- **Additional detection needed:** Follow one level of indirection through test helper functions defined in the same file.
- **Severity:** Medium. Common pattern in large test suites.

**FN-NSA-04: JSON.stringify comparison without field inspection**

- **Pattern:**
```typescript
it("creates user", () => {
  const user = createUser(input);
  expect(JSON.stringify(user)).toBeTruthy();
});
```
- **Why it might be missed:** `JSON.stringify` returns a string, and `.toBeTruthy()` only checks it is non-empty.
- **Additional detection needed:** Detect `JSON.stringify(returnValue)` followed by shallow assertion.
- **Severity:** Low. Unusual pattern.

**FN-NSA-05: console.log as the only "assertion"**

- **Pattern:**
```typescript
it("creates user", () => {
  const user = createUser(input);
  console.log(user);
});
```
- **Why it might be missed:** There are zero assertions. The rule looks for shallow assertions, but if there are none at all, this test escapes detection.
- **Additional detection needed:** This is arguably a different rule (TQ-no-assertion-free-tests), but TQ-no-shallow-assertions should at minimum not count zero-assertion tests as passing.
- **Severity:** High. Tests with zero assertions always pass and test nothing.

### 1.5 Edge Cases

**EC-NSA-01: Generic return type**

- **Input:**
```typescript
function getResult<T>(val: T): T { return val; }
// Test:
it("returns value", () => {
  const r = getResult({ a: 1, b: 2 });
  expect(r).toBeDefined();
});
```
- **Expected:** Violation. Even though the return type is generic, the actual call-site type is `{ a: number; b: number }` which has fields.
- **Verification:** Run `stricture --rule TQ-no-shallow-assertions tests/generic.test.ts` and confirm violation reported.

**EC-NSA-02: Type alias chain**

- **Input:**
```typescript
type ID = number;
type UserID = ID;
interface User { id: UserID; name: string; }
function getUser(): User { /* ... */ }
// Test:
it("gets user", () => {
  const u = getUser();
  expect(u).toBeTruthy();
});
```
- **Expected:** Violation. Type alias resolution should see `User` has fields.

**EC-NSA-03: Union return type**

- **Input:**
```typescript
function getResult(): { ok: true; data: string } | { ok: false; error: string } {
  return { ok: true, data: "hi" };
}
// Test:
it("gets result", () => {
  const r = getResult();
  expect(r).toBeDefined();
});
```
- **Expected:** Violation. Both branches of the union have fields.

**EC-NSA-04: Function returning Promise**

- **Input:**
```typescript
async function fetchData(): Promise<{ id: number; value: string }> { /* ... */ }
// Test:
it("fetches data", async () => {
  const data = await fetchData();
  expect(data).toBeDefined();
});
```
- **Expected:** Violation. The resolved type is `{ id: number; value: string }`.

**EC-NSA-05: Function returning null or undefined union**

- **Input:**
```typescript
function findUser(id: string): User | null { /* ... */ }
// Test:
it("finds user", () => {
  const user = findUser("123");
  expect(user).not.toBeNull();
});
```
- **Expected:** Violation. `User` has fields. `.not.toBeNull()` only proves it is not null.

**EC-NSA-06: Re-exported function**

- **Input:**
```typescript
// src/index.ts
export { createUser } from "./user";
// tests/index.test.ts
import { createUser } from "../src";
it("creates user", () => {
  expect(createUser({ name: "A" })).toBeDefined();
});
```
- **Expected:** Violation. Re-export should be followed to resolve the return type.

**EC-NSA-07: Destructured return value**

- **Input:**
```typescript
it("creates user", () => {
  const { id, name } = createUser(input);
  expect(id).toBeDefined();
  expect(name).toBeDefined();
});
```
- **Expected:** Violation on both lines. Even destructured fields with `.toBeDefined()` are shallow (should check `toBe(1)`, `toBe("Alice")`).

**EC-NSA-08: Assertion inside try/catch**

- **Input:**
```typescript
it("creates user", () => {
  try {
    const user = createUser(input);
    expect(user).toBeTruthy();
  } catch {
    // swallow
  }
});
```
- **Expected:** Violation at the `expect(user).toBeTruthy()` line.

**EC-NSA-09: Go table-driven test with shallow assertion**

- **Input:**
```go
func TestCreate(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"valid", "alice"},
        {"another", "bob"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Create(tt.input)
            if result == nil {
                t.Fatal("nil result")
            }
        })
    }
}
```
- **Expected:** Violation at `result == nil` check inside the loop body. Even in table-driven tests, shallow assertions are violations.

**EC-NSA-10: Empty test file**

- **Input:** An empty `.test.ts` file.
- **Expected:** No violation for this rule (no assertions to analyze). Other rules (TQ-negative-cases, CONV-file-header) may flag it.

**EC-NSA-11: Test file with only comments**

- **Input:**
```typescript
// This file will contain user tests
// TODO: implement
```
- **Expected:** No violation for this rule.

**EC-NSA-12: Unicode identifiers**

- **Input:**
```typescript
interface Benutzer { vorname: string; nachname: string; }
function erstelleBenutzer(): Benutzer { /* ... */ }
it("erstellt Benutzer", () => {
  const b = erstelleBenutzer();
  expect(b).toBeDefined();
});
```
- **Expected:** Violation. Unicode identifiers should not affect rule behavior.

### 1.6 Configuration Interaction

**CI-NSA-01: allowInPresenceTests = false**

- **Config:**
```yaml
TQ-no-shallow-assertions:
  - error
  - allowInPresenceTests: false
```
- **Input:** A test named `"should exist when enabled"` with `expect(result).toBeDefined()`.
- **Expected:** Violation IS reported (presence tests are no longer exempt).

**CI-NSA-02: allowInPresenceTests = true (default)**

- **Config:** Default or `allowInPresenceTests: true`.
- **Input:** Same as CI-NSA-01 plus deep assertions elsewhere in file.
- **Expected:** No violation.

**CI-NSA-03: maxShallowPercent = 0**

- **Config:**
```yaml
TQ-no-shallow-assertions:
  - error
  - maxShallowPercent: 0
```
- **Input:** A file with 10 deep assertions and 1 shallow assertion (10% shallow).
- **Expected:** Violation. Zero tolerance for shallow assertions.

**CI-NSA-04: maxShallowPercent = 50**

- **Config:**
```yaml
TQ-no-shallow-assertions:
  - error
  - maxShallowPercent: 50
```
- **Input:** A file with 5 deep and 4 shallow assertions (44% shallow).
- **Expected:** No violation (under 50% threshold).

**CI-NSA-05: Rule set to "warn"**

- **Config:**
```yaml
TQ-no-shallow-assertions: warn
```
- **Input:** File with shallow assertions.
- **Expected:** Warnings reported, not errors. Exit code is 0 (warnings do not cause exit 1).

**CI-NSA-06: Rule set to "off"**

- **Config:**
```yaml
TQ-no-shallow-assertions: off
```
- **Input:** File with shallow assertions.
- **Expected:** No violations reported.

**CI-NSA-07: Invalid option type**

- **Config:**
```yaml
TQ-no-shallow-assertions:
  - error
  - maxShallowPercent: "fifty"
```
- **Expected:** Exit code 2 with error message: invalid config for `TQ-no-shallow-assertions`: `maxShallowPercent` must be a number.

### 1.7 Inline Suppression Testing

**IS-NSA-01: Disable next line suppresses violation**

- **Input:**
```typescript
// stricture-disable-next-line TQ-no-shallow-assertions
expect(result).toBeDefined();
```
- **Expected:** No violation on the `expect` line.

**IS-NSA-02: Disable next line does not suppress the line after**

- **Input:**
```typescript
// stricture-disable-next-line TQ-no-shallow-assertions
expect(result).toBeDefined();
expect(result2).toBeDefined();
```
- **Expected:** No violation on line 2; violation on line 3.

**IS-NSA-03: Block suppression**

- **Input:**
```typescript
// stricture-disable TQ-no-shallow-assertions
expect(a).toBeDefined();
expect(b).toBeDefined();
// stricture-enable TQ-no-shallow-assertions
expect(c).toBeDefined();
```
- **Expected:** No violations on lines 2-3; violation on line 5.

**IS-NSA-04: Suppressing one rule does not suppress others**

- **Input:**
```typescript
// stricture-disable-next-line TQ-no-shallow-assertions
expect(result).toBeDefined(); // also triggers TQ-assertion-depth on same line
```
- **Expected:** `TQ-no-shallow-assertions` is suppressed. `TQ-assertion-depth` (if applicable) is NOT suppressed.

**IS-NSA-05: Wrong syntax is not treated as suppression**

- **Input:**
```typescript
// stricture-ignore TQ-no-shallow-assertions
expect(result).toBeDefined();
```
- **Expected:** Violation IS reported. `stricture-ignore` is not valid syntax (must be `stricture-disable-next-line`).

**IS-NSA-06: Nonexistent rule ID in suppression produces warning**

- **Input:**
```typescript
// stricture-disable-next-line TQ-nonexistent-rule
expect(result).toBeDefined();
```
- **Expected:** Warning: `Unknown rule ID "TQ-nonexistent-rule" in suppression comment`. The `TQ-no-shallow-assertions` violation IS still reported.

---

## 2. TQ-return-type-verified

**Rule purpose:** Ensure that when a function returns a structured type, tests verify all required fields of that type.

### 2.1 True Positive Cases

**TP-RTV-01: Only 2 of 5 fields asserted**

- **Source file** (`src/user.ts`):
```typescript
interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "user";
  createdAt: Date;
}
export function createUser(input: CreateUserInput): User { /* ... */ }
```
- **Test file** (`tests/user.test.ts`):
```typescript
it("creates user", () => {
  const user = createUser(validInput);
  expect(user.id).toBe(1);
  expect(user.name).toBe("Alice");
});
```
- **Expected violation:** `TQ-return-type-verified` on `tests/user.test.ts`. Message: `Function "createUser" returns User (5 fields) but only 2/5 are asserted (40%)`.
- **Why:** email, role, createdAt are never asserted.

**TP-RTV-02: Zero fields asserted (only shallow check)**

- **Source:** Same User type with 5 fields.
- **Test:**
```typescript
it("creates user", () => {
  const user = createUser(validInput);
  expect(user).toBeDefined();
});
```
- **Expected violation:** `TQ-return-type-verified`. Message: `0/5 fields asserted (0%)`.

**TP-RTV-03: Nested type with partial assertion**

- **Source:**
```typescript
interface Order {
  id: string;
  customer: { name: string; email: string; };
  total: number;
}
export function getOrder(id: string): Order { /* ... */ }
```
- **Test:**
```typescript
it("gets order", () => {
  const order = getOrder("1");
  expect(order.id).toBe("1");
  expect(order.total).toBe(99);
});
```
- **Expected violation:** `TQ-return-type-verified`. Missing `customer` (or `customer.name` and `customer.email`). 2/3 top-level fields asserted (67%).

**TP-RTV-04: Go struct with only 1 of 5 fields checked**

- **Source** (`user.go`):
```go
type User struct {
    ID        int
    Name      string
    Email     string
    Role      string
    CreatedAt time.Time
}
func CreateUser(input CreateUserInput) *User { /* ... */ }
```
- **Test** (`user_test.go`):
```go
func TestCreateUser(t *testing.T) {
    user := CreateUser(validInput)
    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
}
```
- **Expected violation:** `TQ-return-type-verified`. 1/5 fields (20%).

**TP-RTV-05: Interface with many fields, test uses toMatchObject with subset**

- **Source:**
```typescript
interface Config {
  port: number; host: string; debug: boolean;
  logLevel: string; maxRetries: number;
}
export function loadConfig(): Config { /* ... */ }
```
- **Test:**
```typescript
it("loads config", () => {
  const cfg = loadConfig();
  expect(cfg).toMatchObject({ port: 3000 });
});
```
- **Expected violation:** `TQ-return-type-verified`. `toMatchObject` with 1 field covers only 1/5 = 20%.

**TP-RTV-06: Multiple test cases that collectively miss a field**

- **Source:** User type with `id`, `name`, `email`, `role`.
- **Test:**
```typescript
it("has correct id", () => {
  expect(createUser(input).id).toBe(1);
});
it("has correct name", () => {
  expect(createUser(input).name).toBe("Alice");
});
it("has correct email", () => {
  expect(createUser(input).email).toBe("a@b.com");
});
// Missing: no test for "role" across any test case
```
- **Expected violation:** `TQ-return-type-verified`. 3/4 fields (75%). Below 80% default threshold.

**TP-RTV-07: Return type is an array of objects, no element fields checked**

- **Source:**
```typescript
interface Item { id: number; name: string; price: number; }
export function getItems(): Item[] { /* ... */ }
```
- **Test:**
```typescript
it("gets items", () => {
  const items = getItems();
  expect(items).toHaveLength(3);
});
```
- **Expected violation:** `TQ-return-type-verified`. The element type `Item` has 3 fields, none asserted.

**TP-RTV-08: Go method on a struct returns a type with partial coverage**

- **Source:**
```go
type Service struct{}
type Result struct {
    Code    int
    Message string
    Data    []byte
}
func (s *Service) Process(input string) *Result { /* ... */ }
```
- **Test:**
```go
func TestProcess(t *testing.T) {
    svc := &Service{}
    r := svc.Process("hello")
    if r.Code != 200 {
        t.Errorf("expected 200, got %d", r.Code)
    }
}
```
- **Expected violation:** 1/3 fields = 33%.

**TP-RTV-09: Function returns intersection type, partial coverage**

- **Source:**
```typescript
type Base = { id: number; createdAt: Date; };
type UserData = { name: string; email: string; };
type UserRecord = Base & UserData;
export function createRecord(): UserRecord { /* ... */ }
```
- **Test:**
```typescript
it("creates record", () => {
  const r = createRecord();
  expect(r.id).toBe(1);
  expect(r.name).toBe("Alice");
});
```
- **Expected violation:** 2/4 fields = 50%. Missing `createdAt`, `email`.

**TP-RTV-10: Only checks length of array field, not elements**

- **Source:**
```typescript
interface Response { items: Product[]; total: number; page: number; }
export function search(query: string): Response { /* ... */ }
```
- **Test:**
```typescript
it("searches", () => {
  const res = search("widget");
  expect(res.items).toHaveLength(10);
  expect(res.total).toBe(100);
});
```
- **Expected violation:** 2/3 fields (67%). `page` is not asserted.

### 2.2 True Negative Cases

**TN-RTV-01: All fields asserted individually**

- **Test:**
```typescript
it("creates user", () => {
  const user = createUser(validInput);
  expect(user.id).toBe(1);
  expect(user.name).toBe("Alice");
  expect(user.email).toBe("a@b.com");
  expect(user.role).toBe("user");
  expect(user.createdAt).toBeInstanceOf(Date);
});
```
- **Expected:** No violation. 5/5 fields = 100%.

**TN-RTV-02: toEqual with all fields**

- **Test:**
```typescript
it("creates user", () => {
  expect(createUser(validInput)).toEqual({
    id: 1,
    name: "Alice",
    email: "a@b.com",
    role: "user",
    createdAt: expect.any(Date),
  });
});
```
- **Expected:** No violation. `toEqual` with all fields = 100% (with `countToEqual: true` default).

**TN-RTV-03: Cross-test aggregation meets threshold**

- **Test:**
```typescript
it("has correct id", () => { expect(createUser(input).id).toBe(1); });
it("has correct name", () => { expect(createUser(input).name).toBe("Alice"); });
it("has correct email", () => { expect(createUser(input).email).toBe("a@b.com"); });
it("has correct role", () => { expect(createUser(input).role).toBe("user"); });
```
- **Expected:** No violation. 4/5 fields = 80% (meets default threshold). `createdAt` is the 5th field, but if `ignoreOptionalFields: true` and it is optional, or if `ignoreFields` includes `createdAt`, this passes.

**TN-RTV-04: Function returns primitive**

- **Source:** `function count(): number { return 42; }`
- **Test:**
```typescript
it("counts", () => {
  expect(count()).toBe(42);
});
```
- **Expected:** No violation. Primitive return types have no fields to cover.

**TN-RTV-05: Go struct with all fields checked**

- **Test:**
```go
func TestCreateUser(t *testing.T) {
    user := CreateUser(validInput)
    assert.Equal(t, 1, user.ID)
    assert.Equal(t, "Alice", user.Name)
    assert.Equal(t, "a@b.com", user.Email)
    assert.Equal(t, "user", user.Role)
    assert.False(t, user.CreatedAt.IsZero())
}
```
- **Expected:** No violation. 5/5 fields.

**TN-RTV-06: ignoreFields config excludes timestamp fields**

- **Config:**
```yaml
TQ-return-type-verified:
  - error
  - ignoreFields: ["createdAt", "updatedAt"]
```
- **Test:** Checks 3/5 fields but the two missing are `createdAt` and `updatedAt`.
- **Expected:** No violation. 3/3 non-ignored fields = 100%.

### 2.3 False Positive Risks

**FP-RTV-01: Dynamic return type (mapped/conditional) that cannot be statically resolved**

- **Pattern:**
```typescript
function transform<T>(input: T): Partial<T> { /* ... */ }
```
- **Why NOT a violation:** `Partial<T>` makes all fields optional. If the test checks a subset, that may be correct because not all fields will be present.
- **Detection avoidance:** If the return type is `Partial<T>` or a mapped type that makes fields optional, treat all fields as optional.

**FP-RTV-02: Function overloads returning different types**

- **Pattern:**
```typescript
function get(id: string): User;
function get(ids: string[]): User[];
```
- **Why risk:** If the test only tests the array overload and checks `toHaveLength`, it might be flagged for not checking `User` fields, even though the single-user overload is tested elsewhere.
- **Detection avoidance:** Resolve which overload is being called in each test.

**FP-RTV-03: Exported helper function that is not the primary function under test**

- **Pattern:** A test file imports two functions. It deeply tests `createUser` but only checks the existence of a secondary `generateId` return.
- **Detection avoidance:** Only apply the rule to the primary functions under test (identified by test-to-source mapping), not every function imported.

**FP-RTV-04: Test that correctly ignores computed/derived fields**

- **Pattern:** `User` has `fullName` computed from `firstName` + `lastName`. The test checks `firstName` and `lastName` but not `fullName`.
- **Detection avoidance:** Consider adding an `ignoreComputedFields` option or letting `ignoreFields` handle this.

**FP-RTV-05: Return type with index signature**

- **Pattern:**
```typescript
interface Config { [key: string]: unknown; port: number; }
```
- **Why risk:** The index signature means infinite possible fields. Only named fields should count.
- **Detection avoidance:** Only count explicitly named fields in the interface, not index signatures.

### 2.4 False Negative Risks

**FN-RTV-01: toEqual with hardcoded object that misses a field added later**

- **Pattern:**
```typescript
expect(user).toEqual({ id: 1, name: "Alice", email: "a@b.com" });
```
- **Why missed:** At the time of writing, this covers all fields. Later, a `role` field is added to `User`. The test still passes because `toEqual` does exact match on the expected object... but `role` is now untested.
- **Detection needed:** Compare `toEqual` object keys against current type definition.
- **Severity:** High. This is a common drift pattern.

**FN-RTV-02: Field checked in a describe block that does not actually call the function under test**

- **Pattern:**
```typescript
describe("createUser", () => {
  it("returns user", () => {
    const user = createUser(input);
    expect(user.id).toBe(1);
  });
});
describe("other function", () => {
  it("also checks a user", () => {
    const user = { name: "Alice", email: "a@b.com", role: "user", createdAt: new Date() };
    expect(user.name).toBe("Alice");
  });
});
```
- **Why missed:** The second describe block checks `name`, `email`, `role`, `createdAt` on a locally-constructed object that is NOT from `createUser`. If the tool naively counts all field references in the file, it might think all fields are covered.
- **Detection needed:** Track which assertions are on the actual return value of the function under test, not any object with similar shape.
- **Severity:** High.

**FN-RTV-03: assertion through a custom assertion helper**

- **Pattern:**
```typescript
function assertUser(user: User) {
  expect(user.id).toBeGreaterThan(0);
}
it("creates user", () => { assertUser(createUser(input)); });
```
- **Why missed:** Only `id` is checked inside the helper. If the tool does not follow into helper functions, it may miss this entirely or count it as fully covered.
- **Detection needed:** Analyze test helper functions defined in the file.
- **Severity:** Medium.

**FN-RTV-04: Assertion in afterEach or beforeEach**

- **Pattern:**
```typescript
let user: User;
beforeEach(() => { user = createUser(input); });
afterEach(() => { expect(user.id).toBeGreaterThan(0); });
it("has name", () => { expect(user.name).toBe("Alice"); });
```
- **Why missed:** The `id` assertion is in `afterEach`, which runs after every test. If the tool only scans `it()` blocks, it misses this.
- **Detection needed:** Include assertions in beforeEach/afterEach in the count.
- **Severity:** Low. Unusual pattern.

**FN-RTV-05: Spread operator hides field coverage**

- **Pattern:**
```typescript
const expected = { ...defaultUser, name: "Alice" };
expect(user).toEqual(expected);
```
- **Why missed:** The tool must resolve the spread to determine which fields are covered. `defaultUser` might cover all fields or just some.
- **Detection needed:** Resolve spread operator targets to their type definitions.
- **Severity:** Medium.

### 2.5 Edge Cases

**EC-RTV-01: Function with generic return type `Promise<T>`**

- Resolve the inner type T. If T is a struct/interface, require field coverage on T.

**EC-RTV-02: Function returns `void`/`undefined`**

- Skip this rule entirely. There are no fields to cover.

**EC-RTV-03: Function returns a class instance**

- Treat class public properties/getters as fields.

**EC-RTV-04: Go embedded struct**

- Promoted fields from embedded structs should count as fields of the outer struct.

**EC-RTV-05: TypeScript conditional type return**

- `T extends string ? StringResult : NumberResult` -- both branches should have field coverage.

**EC-RTV-06: Return type is `any` or `unknown`**

- Skip this rule. Cannot determine fields.

**EC-RTV-07: Optional fields with ignoreOptionalFields=true**

- Only count required fields toward the total.

**EC-RTV-08: Computed property names in return type**

- `{ [Symbol.iterator]: ... }` -- skip symbol-keyed properties.

**EC-RTV-09: Generated protobuf types with many fields**

- Should still enforce. Consider `ignoreFields` for generated metadata fields.

**EC-RTV-10: Recursive type**

- `interface TreeNode { value: string; children: TreeNode[]; }` -- assertion depth is tested by TQ-assertion-depth. For return-type-verified, count `value` and `children` as two fields.

**EC-RTV-11: Overloaded function with different return types per overload**

- Test each overload's return type separately.

### 2.6 Configuration Interaction

**CI-RTV-01: minFieldCoverage = 100**

- File with 4/5 fields asserted (80%) should trigger violation.

**CI-RTV-02: minFieldCoverage = 50**

- File with 3/5 fields asserted (60%) should NOT trigger violation.

**CI-RTV-03: ignoreOptionalFields = true**

- TypeScript interface with 3 required and 2 optional fields. Test checks all 3 required. No violation.

**CI-RTV-04: countToEqual = false**

- `expect(user).toEqual({...})` should NOT count as covering all fields. Each field must be individually asserted.

**CI-RTV-05: ignoreFields combined with minFieldCoverage**

- 5-field type, 2 fields in `ignoreFields`, test covers 2/3 remaining = 67%. With minFieldCoverage=60, no violation.

### 2.7 Inline Suppression Testing

Same patterns as Section 1.7 but with rule ID `TQ-return-type-verified`. Verify disable-next-line, block disable/enable, wrong syntax, and nonexistent rule ID warnings.

---

## 3. TQ-schema-conformance

**Rule purpose:** Verify that assertions constrain the types of field values, not just their existence.

### 3.1 True Positive Cases

**TP-SC-01: toBeTruthy on numeric field**

- **Source:** `interface User { id: number; name: string; }`
- **Test:**
```typescript
const user = createUser(input);
expect(user.id).toBeTruthy();
```
- **Expected violation:** `TQ-schema-conformance`. `toBeTruthy` does not constrain `id` to `number`. `id=0` would fail this assertion despite being valid.

**TP-SC-02: No type constraint on any field**

- **Test:**
```typescript
expect(user.id).toBeTruthy();
expect(user.name).toBeTruthy();
```
- **Expected violation:** 0% type coverage.

**TP-SC-03: Reflection-only check in Go**

- **Test:**
```go
if reflect.TypeOf(result).Kind() != reflect.Struct {
    t.Fatal("not a struct")
}
```
- **Expected violation:** Type check at the struct level but no field-level type constraints.

**TP-SC-04: Loose equality does not constrain type**

- **Test:**
```typescript
expect(user.id == 1).toBe(true); // == not ===
```
- **Expected violation:** Loose equality (`==`) allows type coercion. `user.id = "1"` would pass.

**TP-SC-05: Checking truthiness of a string field**

- **Test:**
```typescript
expect(user.email).toBeTruthy();
```
- **Expected violation:** Empty string `""` is a valid email format check but is falsy. This assertion does not constrain type.

**TP-SC-06: toBeGreaterThan without specific value on non-numeric field**

- **Source:** `interface Item { name: string; count: number; }`
- **Test:**
```typescript
expect(item.name).toBeTruthy();
expect(item.count).toBeTruthy();
```
- **Expected violation:** Neither assertion constrains type. `count=0` is valid but would fail `toBeTruthy`.

**TP-SC-07: Go only checks error is not nil, not error type**

- **Test:**
```go
_, err := FetchData("")
if err == nil {
    t.Fatal("expected error")
}
```
- **Expected violation:** Error checked for existence but not type/message.

**TP-SC-08: expect.anything() inside toEqual**

- **Test:**
```typescript
expect(user).toEqual({
  id: expect.anything(),
  name: expect.anything(),
  email: expect.anything(),
});
```
- **Expected violation:** `expect.anything()` does not constrain type. 0% type coverage.

**TP-SC-09: toHaveProperty without value**

- **Test:**
```typescript
expect(user).toHaveProperty("id");
expect(user).toHaveProperty("name");
```
- **Expected violation:** `toHaveProperty(key)` without second argument only checks existence, not type.

**TP-SC-10: Only checking truthiness on Date fields**

- **Test:**
```typescript
expect(user.createdAt).toBeTruthy();
```
- **Expected violation:** Does not verify it is a `Date`. Could be a string, number, or any truthy value.

### 3.2 True Negative Cases

**TN-SC-01: Strict equality with literal value**

- **Test:**
```typescript
expect(user.id).toBe(1);
expect(user.name).toBe("Alice");
```
- **Expected:** No violation. `toBe(1)` constrains to number; `toBe("Alice")` constrains to string.

**TN-SC-02: expect.any(Type)**

- **Test:**
```typescript
expect(user).toEqual(expect.objectContaining({
  id: expect.any(Number),
  name: expect.any(String),
}));
```
- **Expected:** No violation. `expect.any(Number)` is an explicit type constraint.

**TN-SC-03: toBeInstanceOf(Date)**

- **Test:**
```typescript
expect(user.createdAt).toBeInstanceOf(Date);
```
- **Expected:** No violation. Explicit type constraint.

**TN-SC-04: Go type assertion + field checks**

- **Test:**
```go
user, ok := result.(*User)
if !ok { t.Fatal("expected *User") }
if user.ID != 1 { t.Errorf("expected 1, got %d", user.ID) }
```
- **Expected:** No violation. Type assertion and value comparison.

**TN-SC-05: toHaveProperty with value argument**

- **Test:**
```typescript
expect(user).toHaveProperty("id", 1);
```
- **Expected:** No violation. Second arg constrains both existence and value.

### 3.3 False Positive Risks

**FP-SC-01: String matching that implicitly constrains type**

- **Pattern:** `expect(user.name).toMatch(/^[A-Z]/);`
- **Why NOT a violation:** `toMatch` only works on strings. It implicitly constrains the type.

**FP-SC-02: toContain on array (constrains to array type)**

- **Pattern:** `expect(items).toContain("widget");`
- **Why NOT a violation:** `toContain` implicitly proves it is iterable/array.

**FP-SC-03: Numeric comparison methods**

- **Pattern:** `expect(user.id).toBeGreaterThan(0);`
- **Why NOT a violation:** `toBeGreaterThan` works on numbers, implicitly constraining type.

**FP-SC-04: Boolean return with toBe(true)**

- **Pattern:** `expect(isValid).toBe(true);`
- **Why NOT a violation:** `toBe(true)` is strict equality with boolean literal.

**FP-SC-05: toEqual with explicit values**

- **Pattern:** `expect(user).toEqual({ id: 1, name: "Alice" });`
- **Why NOT a violation:** Explicit values in `toEqual` constrain both type and value.

### 3.4 False Negative Risks

**FN-SC-01: toBe with a variable that could be any type**

- **Pattern:** `expect(user.id).toBe(expectedId);` where `expectedId` is `any`.
- **Why missed:** If `expectedId` is typed as `any`, the assertion does not truly constrain the type.
- **Severity:** Medium.

**FN-SC-02: JSON.parse result used without type assertion**

- **Pattern:**
```typescript
const parsed = JSON.parse(response.body);
expect(parsed.id).toBe(1);
```
- **Why missed:** `parsed` is `any`. The `.toBe(1)` constrains value but the source type is unchecked.
- **Severity:** Low. This is more of a TypeScript strict-mode issue.

**FN-SC-03: Assertion through custom matcher**

- **Pattern:** `expect(user).toBeValidUser();` where `toBeValidUser` does deep checks.
- **Why missed:** Static analysis cannot know what custom matchers verify.
- **Severity:** Medium.

**FN-SC-04: Go fmt.Sprintf comparison hides type info**

- **Pattern:** `assert.Equal(t, "1", fmt.Sprintf("%v", user.ID))` -- converts to string before comparison.
- **Severity:** Low.

**FN-SC-05: Dynamic property access**

- **Pattern:** `expect(user[fieldName]).toBe(value);` where `fieldName` is a variable.
- **Why missed:** Cannot statically determine which field is being checked.
- **Severity:** Low.

### 3.5 Edge Cases

**EC-SC-01:** Union type `string | number` -- assertion must constrain to one branch.
**EC-SC-02:** Enum values -- `toBe(Status.Active)` constrains type.
**EC-SC-03:** `null` as a valid value -- `toBe(null)` constrains type to null.
**EC-SC-04:** BigInt fields -- `toBe(1n)` constrains to BigInt.
**EC-SC-05:** Tuple types -- each element position should be independently verified.
**EC-SC-06:** Go interface{}/any fields -- cannot constrain, should warn.
**EC-SC-07:** Type guard functions -- `if (isUser(result))` in test constrains type.
**EC-SC-08:** Template literal types -- `expect(url).toMatch(/^https:\/\//)` constrains.
**EC-SC-09:** Symbol properties -- skip from type coverage calculation.
**EC-SC-10:** Readonly fields -- same treatment as regular fields.
**EC-SC-11:** Map/Record types -- `Record<string, number>` field access with `expect(map["key"]).toBe(1)` constrains value type.

### 3.6 Configuration Interaction

**CI-SC-01:** `minTypeCoverage: 100` -- all fields must have type-constraining assertions.
**CI-SC-02:** `minTypeCoverage: 0` -- rule effectively disabled.
**CI-SC-03:** `strictEquality: false` -- `.toBe()` no longer counts as type-constraining.
**CI-SC-04:** `treatToEqualAsComplete: true` -- `toEqual({...})` with all fields counts as 100%.
**CI-SC-05:** `treatToEqualAsComplete: false` -- each field in `toEqual` individually assessed.

### 3.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-schema-conformance`.

---

## 4. TQ-error-path-coverage

**Rule purpose:** Ensure that functions which can throw/return errors have tests for those error paths.

### 4.1 True Positive Cases

**TP-EPC-01: Only happy path tested, 3 error exits untested**

- **Source** (`src/api.ts`):
```typescript
export async function fetchUser(id: string): Promise<User> {
  if (!id) throw new ValidationError("id is required");
  const res = await fetch(`/users/${id}`);
  if (!res.ok) throw new APIError(`HTTP ${res.status}`);
  const data = await res.json();
  if (!data.id) throw new DataError("invalid response");
  return data;
}
```
- **Test** (`tests/api.test.ts`):
```typescript
it("fetches user", async () => {
  const user = await fetchUser("123");
  expect(user.id).toBe("123");
});
```
- **Expected violation:** `TQ-error-path-coverage`. Message: `Function "fetchUser" has 3 error exits but only 0 are tested (0%)`. Missing: ValidationError (line 2), APIError (line 4), DataError (line 6).

**TP-EPC-02: One of three error paths tested**

- **Test:**
```typescript
it("fetches user", async () => { /* ... */ });
it("throws on empty id", async () => {
  await expect(fetchUser("")).rejects.toThrow(ValidationError);
});
```
- **Expected violation:** 1/3 error paths tested (33%).

**TP-EPC-03: Go function with 3 error returns, 1 tested**

- **Source** (`api.go`):
```go
func FetchUser(id string) (*User, error) {
    if id == "" { return nil, ErrInvalidID }
    resp, err := http.Get("/users/" + id)
    if err != nil { return nil, fmt.Errorf("fetch: %w", err) }
    if resp.StatusCode != 200 { return nil, fmt.Errorf("HTTP %d", resp.StatusCode) }
    // ...
}
```
- **Test** (`api_test.go`):
```go
func TestFetchUser_EmptyID(t *testing.T) {
    _, err := FetchUser("")
    if !errors.Is(err, ErrInvalidID) { t.Fatal("expected ErrInvalidID") }
}
```
- **Expected violation:** 1/3 error paths tested.

**TP-EPC-04: Generic catch without specific error class**

- **Config:** `countGenericCatch: false`
- **Test:**
```typescript
it("throws on bad input", async () => {
  await expect(fetchUser("")).rejects.toThrow();
});
```
- **Expected violation:** `.toThrow()` without argument is generic. Does not count as covering the ValidationError path.

**TP-EPC-05: Error path in nested function call not tested**

- **Source:**
```typescript
export function processOrder(order: Order): Receipt {
  const validated = validateOrder(order); // can throw ValidationError
  const priced = calculatePrice(validated); // can throw PricingError
  return generateReceipt(priced); // can throw ReceiptError
}
```
- **Test:**
```typescript
it("processes order", () => {
  const receipt = processOrder(validOrder);
  expect(receipt.total).toBe(100);
});
```
- **Expected violation:** 0/3 error paths tested.

**TP-EPC-06: Async rejection not tested**

- **Source:**
```typescript
export function save(data: Data): Promise<void> {
  return new Promise((_, reject) => {
    if (!data.valid) reject(new Error("invalid data"));
    // ...
  });
}
```
- **Test:** Only tests the success path.
- **Expected violation:** Promise.reject error path not tested.

**TP-EPC-07: Multiple throw statements in same function, only first tested**

- **Source:**
```typescript
export function parse(input: string): AST {
  if (!input) throw new EmptyInputError();
  if (input.length > 10000) throw new InputTooLargeError();
  if (!isValid(input)) throw new SyntaxError("invalid syntax");
  // ...
}
```
- **Test:**
```typescript
it("throws on empty input", () => {
  expect(() => parse("")).toThrow(EmptyInputError);
});
```
- **Expected violation:** 1/3 error paths. Missing InputTooLargeError and SyntaxError.

**TP-EPC-08: Go function returns error only in defer**

- **Source:**
```go
func WriteFile(path string, data []byte) error {
    f, err := os.Create(path)
    if err != nil { return fmt.Errorf("create: %w", err) }
    defer func() {
        if cerr := f.Close(); cerr != nil {
            err = fmt.Errorf("close: %w", cerr)
        }
    }()
    _, err = f.Write(data)
    if err != nil { return fmt.Errorf("write: %w", err) }
    return nil
}
```
- **Test:** Only tests happy path.
- **Expected violation:** 3 error paths (create, close, write), 0 tested.

**TP-EPC-09: Try/catch that re-throws**

- **Source:**
```typescript
export function transform(input: string): Output {
  try {
    return parse(input);
  } catch (e) {
    throw new TransformError(`transform failed: ${e.message}`);
  }
}
```
- **Test:** No error test for TransformError.
- **Expected violation:** 1 error exit, 0 tested.

**TP-EPC-10: Constructor that throws**

- **Source:**
```typescript
export class Connection {
  constructor(url: string) {
    if (!url.startsWith("http")) throw new Error("invalid URL");
  }
}
```
- **Test:**
```typescript
it("creates connection", () => {
  const conn = new Connection("http://localhost");
  expect(conn).toBeDefined();
});
```
- **Expected violation:** Constructor error path not tested.

### 4.2 True Negative Cases

**TN-EPC-01: All error paths tested**

- **Test:**
```typescript
it("throws on empty id", () => { expect(() => fn("")).toThrow(ValidationError); });
it("throws on HTTP error", () => { /* mock + test */ });
it("throws on invalid response", () => { /* mock + test */ });
```
- **Expected:** No violation.

**TN-EPC-02: Function with no error exits**

- **Source:** `function add(a: number, b: number): number { return a + b; }`
- **Expected:** No violation (no error exits to test).

**TN-EPC-03: Private helper function (ignoreInternalHelpers=true)**

- **Source:** Unexported function `function _validate(x: string)` with throws.
- **Expected:** No violation. Private helpers are skipped by default.

**TN-EPC-04: Go function with no error return**

- **Source:** `func Format(s string) string { return strings.ToUpper(s) }`
- **Expected:** No violation.

**TN-EPC-05: Error test using error message substring**

- **Test:**
```typescript
it("rejects bad input", () => {
  expect(() => fn("")).toThrow("is required");
});
```
- **Source:** `throw new Error("id is required");`
- **Expected:** No violation. Error message substring matches the error exit.

### 4.3 False Positive Risks

**FP-EPC-01: Error in unreachable code (dead code)**

- If an error exit is unreachable (after a return statement), it should not count.

**FP-EPC-02: Error in a logging path (non-functional)**

- `console.error("warning")` is not a throw/return error.

**FP-EPC-03: Re-thrown error in catch block that is actually tested via the outer function**

- The inner function's error path is tested through the outer function's test.

**FP-EPC-04: Generated code with many error exits (protobuf validation)**

- Users may want to exclude generated files.

**FP-EPC-05: Defensive programming: error that "can never happen"**

- `if (typeof x !== "string") throw new Error("unreachable");` -- type system guarantees this cannot happen.

### 4.4 False Negative Risks

**FN-EPC-01: Error thrown in a callback/lambda**

- `items.map(item => { if (!item.valid) throw new Error("invalid"); })` -- the throw is inside a lambda, not directly in the function body.
- **Severity:** Medium.

**FN-EPC-02: Implicit error from third-party call**

- `await db.query(sql)` can throw but there is no explicit throw in the source.
- **Severity:** High. Very common.

**FN-EPC-03: Error path via process.exit(1)**

- Not a throw/return error, but terminates the process.
- **Severity:** Low.

**FN-EPC-04: Go panic (not error return)**

- `panic("unexpected state")` is an error exit but not a `return ..., err`.
- **Severity:** Medium.

**FN-EPC-05: Conditional throw that depends on runtime state**

- `if (config.strict) throw new Error(...)` -- only throws in some configurations.
- **Severity:** Low.

### 4.5 Edge Cases

**EC-EPC-01:** Async generator function with throw inside yield loop.
**EC-EPC-02:** Function that returns `Result<T, E>` pattern (no throw, error in return type).
**EC-EPC-03:** Go function returning `(T, error)` where error is always nil (defensive return).
**EC-EPC-04:** Multiple catch blocks with different error types.
**EC-EPC-05:** Error thrown in finally block.
**EC-EPC-06:** Nested try/catch where inner catch swallows error.
**EC-EPC-07:** Go errors.Join (multiple errors combined).
**EC-EPC-08:** Error exit in switch/case default branch.
**EC-EPC-09:** Throw inside ternary expression: `x ? value : throw new Error()`.
**EC-EPC-10:** Method on a class vs standalone function -- both should be analyzed.
**EC-EPC-11:** Go test using require.Error (testify) should count as testing error path.

### 4.6 Configuration Interaction

**CI-EPC-01:** `minErrorPathCoverage: 100` -- every error exit must be tested.
**CI-EPC-02:** `minErrorPathCoverage: 50` -- at least half.
**CI-EPC-03:** `countGenericCatch: true` -- `.toThrow()` without specific error class counts.
**CI-EPC-04:** `countGenericCatch: false` (default) -- only `.toThrow(SpecificError)` counts.
**CI-EPC-05:** `ignoreInternalHelpers: false` -- unexported functions are also checked.

### 4.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-error-path-coverage`.

---

## 5. TQ-assertion-depth

**Rule purpose:** Verify that assertions go deep enough into nested structures.

### 5.1 True Positive Cases

**TP-AD-01: Only top-level check on deeply nested return**

- **Source** (`src/search.ts`):
```typescript
interface SearchResult {
  data: { items: { id: number; name: string; price: number; }[]; total: number; };
  meta: { page: number; pageSize: number; };
}
export function search(query: string): SearchResult { /* ... */ }
```
- **Test** (`tests/search.test.ts`):
```typescript
it("searches", () => {
  const result = search("widget");
  expect(result.data).toBeDefined();
});
```
- **Expected violation:** `TQ-assertion-depth`. Type depth is 3 (result.data.items[0].name). Max assertion depth is 1 (result.data). Ratio: 1/3 = 33%, below 60% default.

**TP-AD-02: Array length check without element inspection**

- **Test:**
```typescript
it("searches", () => {
  const result = search("widget");
  expect(result.data.items).toHaveLength(2);
});
```
- **Expected violation:** Assertion depth reaches `result.data.items` (depth 2), but type depth is 3 (items[0].name). Ratio: 2/3 = 67%. With default `minDepthRatio: 0.6`, this barely passes. But if `ignoreLeafPrimitives: false`, the leaf primitives (name, price) also need asserting.

**TP-AD-03: Only checks meta, ignores data subtree**

- **Test:**
```typescript
it("searches", () => {
  const result = search("widget");
  expect(result.meta.page).toBe(1);
  expect(result.meta.pageSize).toBe(10);
});
```
- **Expected violation:** Assertions reach depth 2 in `meta` but `data.items[0].*` (depth 3) is never reached.

**TP-AD-04: Go nested struct with shallow assertion**

- **Source:**
```go
type Response struct {
    Data struct {
        Items []struct {
            ID   int
            Name string
        }
        Total int
    }
}
```
- **Test:**
```go
func TestSearch(t *testing.T) {
    r := Search("widget")
    if r.Data.Total != 10 {
        t.Errorf("expected 10, got %d", r.Data.Total)
    }
}
```
- **Expected violation:** Depth 2 assertion (Data.Total) but type depth is 3 (Data.Items[0].Name).

**TP-AD-05: Only checks first level of a 4-deep type**

- **Source:**
```typescript
interface Tree {
  value: string;
  left: { value: string; left: { value: string; leaf: boolean; }; right: null; };
  right: null;
}
```
- **Test:**
```typescript
expect(tree.value).toBe("root");
```
- **Expected violation:** Depth 1 of 3. Ratio: 33%.

**TP-AD-06: toEqual on partial object (missing nested)**

- **Test:**
```typescript
expect(result).toEqual({ data: expect.any(Object), meta: { page: 1, pageSize: 10 } });
```
- **Expected violation:** `data` is checked as `expect.any(Object)` -- no depth into `data.items`.

**TP-AD-07: Checks two sibling paths but not the deepest**

- **Source:**
```typescript
interface Config { server: { host: string; port: number; ssl: { cert: string; key: string; } }; }
```
- **Test:**
```typescript
expect(config.server.host).toBe("localhost");
expect(config.server.port).toBe(3000);
```
- **Expected violation:** Depth 2, but type depth is 3 (server.ssl.cert). Ratio: 2/3 = 67%.

**TP-AD-08: Map/Record with shallow access**

- **Source:**
```typescript
interface Dashboard { widgets: Record<string, { title: string; data: number[]; }>; }
```
- **Test:**
```typescript
expect(dashboard.widgets).toBeDefined();
```
- **Expected violation:** Depth 1, type depth is 2+ (widgets["key"].title).

**TP-AD-09: Array of arrays with no inner check**

- **Source:**
```typescript
interface Matrix { rows: { cells: { value: number; formatted: string; }[] }[] }
```
- **Test:**
```typescript
expect(matrix.rows).toHaveLength(3);
```
- **Expected violation:** Depth 1, type depth is 3 (rows[0].cells[0].value).

**TP-AD-10: Only asserts existence of deeply nested optional**

- **Source:**
```typescript
interface Form { sections: { fields: { validation: { rules: string[]; } }[] }[] }
```
- **Test:**
```typescript
expect(form.sections[0].fields[0].validation).toBeDefined();
```
- **Expected violation:** Depth 3 but type depth is 4 (validation.rules). Ratio: 3/4 = 75%. Passes with default 60%, but if minDepthRatio=0.8, this fails.

### 5.2 True Negative Cases

**TN-AD-01: Full depth assertion**

- **Test:**
```typescript
expect(result.data.items[0].name).toBe("Widget");
expect(result.data.items[0].price).toBe(9.99);
expect(result.data.total).toBe(19.98);
expect(result.meta.page).toBe(1);
```
- **Expected:** No violation. Assertions reach maximum type depth.

**TN-AD-02: toEqual with fully specified nested object**

- **Test:**
```typescript
expect(result).toEqual({
  data: { items: [{ id: 1, name: "Widget", price: 9.99 }], total: 9.99 },
  meta: { page: 1, pageSize: 10 },
});
```
- **Expected:** No violation.

**TN-AD-03: Flat return type (depth 1)**

- **Source:** `interface Simple { a: number; b: string; }`
- **Test:** `expect(result.a).toBe(1);`
- **Expected:** No violation. Type depth is 1, assertion depth is 1. Ratio: 100%.

**TN-AD-04: Primitive return type**

- **Source:** `function count(): number`
- **Test:** `expect(count()).toBe(42);`
- **Expected:** No violation. No nesting to check.

**TN-AD-05: Go deep field checks**

- **Test:**
```go
assert.Equal(t, "Widget", result.Data.Items[0].Name)
assert.Equal(t, 9.99, result.Data.Items[0].Price)
```
- **Expected:** No violation. Full depth reached.

### 5.3 False Positive Risks

**FP-AD-01: Intentionally shallow test for a specific aspect**

- A test named "should return correct page count" that only checks `result.meta.page` is intentionally focused. Flagging it for not checking `result.data.items[0].name` may be noise.
- **Avoidance:** Consider the test name and whether the assertion covers the declared scope.

**FP-AD-02: Recursive type with infinite depth**

- `interface TreeNode { children: TreeNode[] }` has theoretically infinite depth.
- **Avoidance:** Cap type depth calculation at a configurable maximum (e.g., 5 levels).

**FP-AD-03: Large generated type with hundreds of nested fields**

- Asserting every leaf of a 50-field protobuf message is unreasonable.
- **Avoidance:** Respect `ignoreFields` and generated file patterns.

**FP-AD-04: Type depth inflated by optional deep paths**

- Most of the depth comes from an optional `metadata` field that is rarely used.
- **Avoidance:** If `ignoreOptionalFields` is set, exclude optional deep paths from depth calculation.

**FP-AD-05: Test that uses snapshot for deep structure**

- `toMatchSnapshot()` implicitly captures all depth.
- **Avoidance:** If snapshots are allowed, count them as full-depth.

### 5.4 False Negative Risks

**FN-AD-01: Deep property access without assertion**

- `const name = result.data.items[0].name; console.log(name);` -- accessed at full depth but never asserted.
- **Severity:** Medium.

**FN-AD-02: Assertion on wrong object at correct depth**

- Test creates a mock object at the same depth as the return type and asserts on it. Looks deep but tests nothing real.
- **Severity:** High.

**FN-AD-03: toEqual with spread that hides depth**

- `expect(result).toEqual({ ...expected })` where `expected` may or may not include deep fields.
- **Severity:** Medium.

**FN-AD-04: Dynamic property chain**

- `expect(result[path]).toBe(value)` where `path` is built dynamically.
- **Severity:** Low.

**FN-AD-05: Helper function that does deep assertion**

- `assertDeepStructure(result)` -- the helper reaches full depth but the tool cannot see inside it.
- **Severity:** Medium.

### 5.5 Edge Cases

**EC-AD-01:** Array index access `[0]` counts as entering the array element type depth.
**EC-AD-02:** Map/Record key access `["key"]` counts as entering the value type depth.
**EC-AD-03:** Optional chaining `result?.data?.items` -- same depth counting.
**EC-AD-04:** Destructured access `const { data: { items } } = result` then `expect(items[0].name)` -- depth starts from root.
**EC-AD-05:** Go pointer dereference does not add depth.
**EC-AD-06:** TypeScript `Readonly<T>` wrapper does not add depth.
**EC-AD-07:** Intersection type `A & B` -- depth is max of A and B depths.
**EC-AD-08:** Function returning `Map<string, NestedType>` -- Map access adds depth.
**EC-AD-09:** Empty nested object `{ data: {} }` -- depth 1 for data, but data has no fields.
**EC-AD-10:** Tuple type `[string, { nested: boolean }]` -- index 1 access enters object depth.

### 5.6 Configuration Interaction

**CI-AD-01:** `minDepthRatio: 1.0` -- must assert at full depth.
**CI-AD-02:** `minDepthRatio: 0.0` -- effectively disabled.
**CI-AD-03:** `ignoreLeafPrimitives: true` -- leaf string/number/boolean fields do not count toward depth.
**CI-AD-04:** `ignoreLeafPrimitives: false` -- all leaves count.

### 5.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-assertion-depth`.

---

## 6. TQ-boundary-tested

**Rule purpose:** Ensure functions with numeric, string, or collection parameters have tests at boundary values.

### 6.1 True Positive Cases

**TP-BT-01: Only mid-range values tested for pagination**

- **Source:**
```typescript
export function paginate(items: Item[], page: number, pageSize: number): Item[] { /* ... */ }
```
- **Test:**
```typescript
it("paginates items", () => {
  const result = paginate(items, 2, 10);
  expect(result).toHaveLength(10);
});
```
- **Expected violation:** `TQ-boundary-tested`. Missing boundary tests: page=0, page=1, pageSize=0, pageSize=1, empty items array.

**TP-BT-02: String parameter without empty string test**

- **Source:**
```typescript
export function validate(name: string): boolean { /* ... */ }
```
- **Test:**
```typescript
it("validates name", () => {
  expect(validate("Alice")).toBe(true);
});
```
- **Expected violation:** Missing boundary: empty string `""`.

**TP-BT-03: Array parameter without empty array test**

- **Source:**
```typescript
export function sum(numbers: number[]): number { /* ... */ }
```
- **Test:**
```typescript
it("sums numbers", () => {
  expect(sum([1, 2, 3])).toBe(6);
});
```
- **Expected violation:** Missing boundary: empty array `[]`.

**TP-BT-04: Numeric parameter without zero test**

- **Source:**
```typescript
export function divide(a: number, b: number): number { /* ... */ }
```
- **Test:**
```typescript
it("divides", () => {
  expect(divide(10, 2)).toBe(5);
});
```
- **Expected violation:** Missing boundary: b=0 (division by zero), a=0.

**TP-BT-05: Go slice parameter without nil/empty test**

- **Source:**
```go
func Average(nums []float64) float64 { /* ... */ }
```
- **Test:**
```go
func TestAverage(t *testing.T) {
    result := Average([]float64{1.0, 2.0, 3.0})
    assert.Equal(t, 2.0, result)
}
```
- **Expected violation:** Missing: nil slice, empty slice `[]float64{}`.

**TP-BT-06: Negative number not tested**

- **Source:** `function abs(n: number): number { /* ... */ }`
- **Test:** Only tests positive numbers.
- **Expected violation:** Missing boundary: n=-1, n=0.

**TP-BT-07: String length boundary not tested**

- **Source:** `function truncate(s: string, maxLen: number): string { /* ... */ }`
- **Test:** Tests with `maxLen=10` and a 20-char string.
- **Expected violation:** Missing: maxLen=0, maxLen=1, s="" (empty string), s with exactly maxLen characters.

**TP-BT-08: Map/object parameter without empty object test**

- **Source:**
```typescript
export function mergeConfigs(configs: Record<string, string>): Record<string, string> { /* ... */ }
```
- **Test:** Only tests with populated objects.
- **Expected violation:** Missing boundary: empty object `{}`.

**TP-BT-09: Go int parameter without zero and negative**

- **Source:**
```go
func Repeat(s string, n int) string { /* ... */ }
```
- **Test:**
```go
func TestRepeat(t *testing.T) {
    assert.Equal(t, "aaa", Repeat("a", 3))
}
```
- **Expected violation:** Missing: n=0, n=-1, s="".

**TP-BT-10: Multiple numeric parameters, none boundary-tested**

- **Source:**
```typescript
export function range(start: number, end: number, step: number): number[] { /* ... */ }
```
- **Test:**
```typescript
it("creates range", () => {
  expect(range(1, 10, 2)).toEqual([1, 3, 5, 7, 9]);
});
```
- **Expected violation:** Missing boundaries for all three params: start=0, end=0, step=0, step=1, step=-1, start=end.

### 6.2 True Negative Cases

**TN-BT-01: All boundaries covered**

- **Test:**
```typescript
it("paginates empty array", () => { paginate([], 1, 10); });
it("paginates page 0", () => { paginate(items, 0, 10); });
it("paginates page 1", () => { paginate(items, 1, 10); });
it("paginates pageSize 0", () => { paginate(items, 1, 0); });
it("paginates pageSize 1", () => { paginate(items, 1, 1); });
it("paginates beyond total", () => { paginate(items, 999, 10); });
```
- **Expected:** No violation.

**TN-BT-02: Function with no boundaryable parameters**

- **Source:** `function greet(): string { return "hello"; }` -- no parameters.
- **Expected:** No violation (rule does not apply).

**TN-BT-03: Boolean parameter**

- **Source:** `function toggle(flag: boolean): void`
- **Expected:** No violation. Booleans have only two values (true/false), not boundary values.

**TN-BT-04: Enum parameter**

- **Source:** `function handleStatus(s: "active" | "inactive"): void`
- **Expected:** No violation. Enum-like unions are not boundary-tested (they are finite sets).

**TN-BT-05: Go test with nil slice**

- **Test:**
```go
func TestAverage_Nil(t *testing.T) { Average(nil) }
func TestAverage_Empty(t *testing.T) { Average([]float64{}) }
func TestAverage_Single(t *testing.T) { Average([]float64{5.0}) }
func TestAverage_Normal(t *testing.T) { Average([]float64{1.0, 2.0, 3.0}) }
```
- **Expected:** No violation. Boundaries covered.

### 6.3 False Positive Risks

**FP-BT-01: Generic numeric parameter that cannot be zero**

- `function getPage(page: number)` where page is 1-indexed and 0 is explicitly invalid by design.
- **Avoidance:** The test for page=0 should exist as a negative test (testing that it throws). If TQ-negative-cases catches this, TQ-boundary-tested should not double-flag.

**FP-BT-02: String parameter that represents a specific format (UUID)**

- Empty string "" is not a meaningful boundary for a UUID parameter.
- **Avoidance:** Consider type annotations or parameter name heuristics to skip format-specific parameters.

**FP-BT-03: Array parameter with minimum length constraint**

- `function first(items: [Item, ...Item[])` -- empty array is impossible by type.
- **Avoidance:** Respect tuple types that enforce minimum length.

**FP-BT-04: Computed/derived parameter from config**

- Parameter value comes from config and is always > 0 in practice.
- **Avoidance:** Cannot determine runtime constraints statically. Flag anyway -- testing boundary is still valuable.

**FP-BT-05: Third-party type parameter**

- `function processRequest(req: express.Request)` -- cannot meaningfully boundary-test a complex third-party type.
- **Avoidance:** Only apply boundary testing to primitive types (number, string) and collections (array, slice, map).

### 6.4 False Negative Risks

**FN-BT-01: Boundary value used but not asserted**

- Test calls `paginate([], 1, 10)` but has no assertion. The boundary value is exercised but not verified.
- **Severity:** Medium.

**FN-BT-02: Boundary embedded in a variable**

- `const page = 0; paginate(items, page, 10);` -- tool must recognize `page = 0` as a boundary value.
- **Severity:** Medium.

**FN-BT-03: Boundary in table-driven test with many rows**

- Go table-driven test with 50 cases. Boundary values exist in the table but tool must scan all table entries.
- **Severity:** Low.

**FN-BT-04: Custom type alias hides numeric type**

- `type PageNumber = number;` -- tool must resolve alias to detect boundary.
- **Severity:** Medium.

**FN-BT-05: MAX_SAFE_INTEGER or MaxInt not tested**

- For numeric params, the maximum value boundary is harder to detect. Tool may only check 0, 1, -1.
- **Severity:** Low.

### 6.5 Edge Cases

**EC-BT-01:** Optional parameter -- boundary includes `undefined`.
**EC-BT-02:** Default parameter value -- `function f(n: number = 10)` -- test without argument tests the default.
**EC-BT-03:** Rest parameter -- `function f(...args: number[])` -- test with 0 args, 1 arg.
**EC-BT-04:** Go variadic parameter -- `func f(args ...int)` -- same as above.
**EC-BT-05:** Nullable parameter -- `function f(n: number | null)` -- test with null.
**EC-BT-06:** BigInt parameter -- boundaries are 0n, 1n, -1n.
**EC-BT-07:** Float parameter -- boundaries include 0.0, -0.0, NaN, Infinity.
**EC-BT-08:** Uint parameter in Go -- boundary is 0 (no negative).
**EC-BT-09:** Parameter destructured from object -- `function f({ page, size }: Opts)`.
**EC-BT-10:** Multiple functions in the same source file -- each function's params tested independently.

### 6.6 Configuration Interaction

**CI-BT-01:** `requireZero: false` -- no violation for missing zero test.
**CI-BT-02:** `requireEmpty: false` -- no violation for missing empty string/array test.
**CI-BT-03:** `requireNegative: false` -- no violation for missing negative number test.
**CI-BT-04:** `customBoundaries: { pageSize: [1, 100] }` -- must test pageSize=1 and pageSize=100.

### 6.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-boundary-tested`.

---

## 7. TQ-mock-scope

**Rule purpose:** Ensure mocks are scoped correctly -- no global mocks that leak between tests, no mocks without cleanup.

### 7.1 True Positive Cases

**TP-MS-01: Global jest.mock at module level**

- **Input:**
```typescript
jest.mock("../database");

describe("UserService", () => {
  it("creates user", () => { /* ... */ });
  it("finds user", () => { /* ... */ });
});
```
- **Expected violation:** `TQ-mock-scope` at line 1. Message: `Module-level jest.mock("../database") leaks across all tests in file`.

**TP-MS-02: jest.spyOn without restore**

- **Input:**
```typescript
it("uses current time", () => {
  jest.spyOn(Date, "now").mockReturnValue(1000);
  const result = getTimestamp();
  expect(result).toBe(1000);
  // Missing: no mockRestore()
});
```
- **Expected violation:** `TQ-mock-scope`. Mock created without cleanup.

**TP-MS-03: Multiple spies, one without cleanup**

- **Input:**
```typescript
it("processes", () => {
  const spy1 = jest.spyOn(console, "log").mockImplementation();
  const spy2 = jest.spyOn(console, "error").mockImplementation();
  // ...
  spy1.mockRestore();
  // spy2 never restored
});
```
- **Expected violation:** spy2 created without cleanup.

**TP-MS-04: Go global variable modified without t.Cleanup**

- **Input:**
```go
var globalDB *sql.DB

func TestCreate(t *testing.T) {
    globalDB = mockDB()
    // Missing: t.Cleanup(func() { globalDB = nil })
    Create(globalDB)
}
```
- **Expected violation:** Global state modified without cleanup.

**TP-MS-05: jest.mock inside describe but not in beforeEach**

- **Input:**
```typescript
describe("Service", () => {
  jest.mock("../api");
  it("test 1", () => { /* ... */ });
  it("test 2", () => { /* ... */ });
});
```
- **Expected violation:** `jest.mock` at describe level leaks to all tests in that describe block.

**TP-MS-06: vi.spyOn (Vitest) without restore**

- **Input:**
```typescript
import { vi } from "vitest";
it("test", () => {
  vi.spyOn(Math, "random").mockReturnValue(0.5);
  // no restore
});
```
- **Expected violation:** Mock without cleanup.

**TP-MS-07: sinon.stub without restore**

- **Input:**
```typescript
it("test", () => {
  const stub = sinon.stub(fs, "readFileSync").returns("data");
  // no stub.restore()
});
```
- **Expected violation:** Stub created without cleanup.

**TP-MS-08: Mock in beforeAll (runs once, leaks across all tests)**

- **Input:**
```typescript
beforeAll(() => {
  jest.spyOn(Date, "now").mockReturnValue(1000);
});
// No afterAll restore
```
- **Expected violation:** Mock in beforeAll without matching afterAll cleanup.

**TP-MS-09: Multiple jest.mock calls at module level**

- **Input:**
```typescript
jest.mock("../db");
jest.mock("../cache");
jest.mock("../logger");
```
- **Expected violation:** Three violations, one per mock.

**TP-MS-10: Go test modifying package-level variable**

- **Input:**
```go
var timeout = 30 * time.Second

func TestFast(t *testing.T) {
    timeout = 1 * time.Second
    result := doWithTimeout()
    // timeout left at 1s for subsequent tests
}
```
- **Expected violation:** Package-level variable modified without t.Cleanup.

### 7.2 True Negative Cases

**TN-MS-01: Mock in beforeEach with afterEach cleanup**

- **Input:**
```typescript
beforeEach(() => { jest.spyOn(Date, "now").mockReturnValue(1000); });
afterEach(() => { jest.restoreAllMocks(); });
```
- **Expected:** No violation.

**TN-MS-02: Mock with inline restore**

- **Input:**
```typescript
it("test", () => {
  const spy = jest.spyOn(Date, "now").mockReturnValue(1000);
  // ...
  spy.mockRestore();
});
```
- **Expected:** No violation.

**TN-MS-03: Go test with t.Cleanup**

- **Input:**
```go
func TestCreate(t *testing.T) {
    old := globalDB
    globalDB = mockDB()
    t.Cleanup(func() { globalDB = old })
}
```
- **Expected:** No violation.

**TN-MS-04: Allowed global mock via config**

- **Config:**
```yaml
TQ-mock-scope:
  - error
  - allowGlobalMocks: ["../database"]
```
- **Input:** `jest.mock("../database");` at module level.
- **Expected:** No violation (explicitly allowed).

**TN-MS-05: jest.fn() (not spying on real module)**

- **Input:**
```typescript
it("test", () => {
  const callback = jest.fn();
  processData([1, 2, 3], callback);
  expect(callback).toHaveBeenCalledTimes(3);
});
```
- **Expected:** No violation. `jest.fn()` creates a local mock, not a spy on a global.

### 7.3 False Positive Risks

**FP-MS-01: jest.mock at module level for a truly side-effect-free mock**

- Mocking a pure utility module that has no state. Leaking is harmless.
- **Avoidance:** `allowGlobalMocks` config option.

**FP-MS-02: Auto-restore enabled in jest config**

- If `jest.config.js` has `restoreMocks: true`, cleanup is automatic.
- **Avoidance:** Consider checking jest config for `restoreMocks` setting.

**FP-MS-03: beforeAll mock that is intentionally shared**

- Integration tests that set up a test server in beforeAll and tear down in afterAll.
- **Avoidance:** If beforeAll has matching afterAll cleanup, no violation.

**FP-MS-04: jest.mock with factory that creates fresh instances per test**

- `jest.mock("../db", () => ({ query: jest.fn() }))` -- each import gets a fresh mock.
- **Avoidance:** Still technically module-level. Flag but at lower confidence.

**FP-MS-05: Go t.TempDir (automatic cleanup)**

- `dir := t.TempDir()` has automatic cleanup. No need for explicit t.Cleanup.
- **Avoidance:** Recognize built-in auto-cleanup patterns.

### 7.4 False Negative Risks

**FN-MS-01: Mock restored in wrong order**

- Two spies created, restored in wrong order. State may leak.
- **Severity:** Low.

**FN-MS-02: Mock created conditionally**

- `if (process.env.CI) jest.spyOn(...)` -- only created in some environments.
- **Severity:** Low.

**FN-MS-03: Mock in helper function called from test**

- `function setupMocks() { jest.spyOn(...); }` -- tool must follow into helpers.
- **Severity:** Medium.

**FN-MS-04: Go monkey-patching via unsafe**

- Using `reflect` or `unsafe` to modify unexported variables.
- **Severity:** Low. Rare pattern.

**FN-MS-05: Timer mocks (jest.useFakeTimers) without cleanup**

- `jest.useFakeTimers()` leaks if not followed by `jest.useRealTimers()`.
- **Severity:** Medium.

### 7.5 Edge Cases

**EC-MS-01:** `jest.mock` inside `it()` block -- is this per-test or module-level? (Jest hoists it to module level.)
**EC-MS-02:** `jest.doMock` (non-hoisted) -- different scoping rules.
**EC-MS-03:** Multiple describe blocks with their own beforeEach/afterEach.
**EC-MS-04:** Nested describe with mock in outer beforeEach.
**EC-MS-05:** Go httptest.NewServer -- not a mock, a test server. Should not require t.Cleanup (server has Close()).
**EC-MS-06:** Vitest `vi.mock` vs `vi.doMock`.
**EC-MS-07:** Mocha: `sinon.sandbox.create()` with automatic cleanup.
**EC-MS-08:** Custom mock framework not recognized by Stricture.
**EC-MS-09:** Mock in a test file that is not a test (utility file).
**EC-MS-10:** Class-level mock in a test class (OOP test style).

### 7.6 Configuration Interaction

**CI-MS-01:** `allowGlobalMocks: ["../logger"]` -- global mock of logger is permitted.
**CI-MS-02:** `requireCleanup: false` -- spies without restore are permitted.
**CI-MS-03:** Rule set to "warn" -- reported as warning, not error.

### 7.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-mock-scope`.

---

## 8. TQ-test-isolation

**Rule purpose:** Ensure tests do not depend on execution order or shared mutable state.

### 8.1 True Positive Cases

**TP-TI-01: Shared mutable counter between tests**

- **Input:**
```typescript
let counter = 0;
it("increments", () => { counter++; expect(counter).toBe(1); });
it("increments again", () => { counter++; expect(counter).toBe(2); });
```
- **Expected violation:** `TQ-test-isolation`. `counter` mutated in first test and read in second test.

**TP-TI-02: Shared array mutated across tests**

- **Input:**
```typescript
const items: string[] = [];
it("adds item", () => { items.push("a"); expect(items).toHaveLength(1); });
it("adds another", () => { items.push("b"); expect(items).toHaveLength(2); });
```
- **Expected violation:** `items` mutated across tests.

**TP-TI-03: Test depends on side effect of previous test**

- **Input:**
```typescript
it("creates user", async () => { await createUser({ name: "Alice" }); });
it("finds user", async () => {
  const user = await findUser("Alice");
  expect(user.name).toBe("Alice");
});
```
- **Expected violation:** Second test depends on first test's side effect (user created in DB).

**TP-TI-04: Go shared map mutation**

- **Input:**
```go
var testData = map[string]string{}

func TestAdd(t *testing.T) {
    testData["key"] = "value"
}

func TestRead(t *testing.T) {
    if testData["key"] != "value" {
        t.Fatal("missing key")
    }
}
```
- **Expected violation:** Shared map mutated across test functions.

**TP-TI-05: File system write without cleanup**

- **Input (with `checkFileSystem: true`):**
```typescript
it("writes config", () => {
  fs.writeFileSync("/tmp/test.conf", "data");
  // No cleanup
});
it("reads config", () => {
  const data = fs.readFileSync("/tmp/test.conf", "utf8");
  expect(data).toBe("data");
});
```
- **Expected violation:** File system side effect without cleanup, second test depends on it.

**TP-TI-06: Shared object property mutation**

- **Input:**
```typescript
const config = { debug: false };
it("enables debug", () => { config.debug = true; /* ... */ });
it("logs when debug", () => { expect(config.debug).toBe(true); });
```
- **Expected violation:** Object property mutated across tests.

**TP-TI-07: Global variable reassignment**

- **Input:**
```typescript
let apiUrl = "https://api.example.com";
it("uses staging", () => { apiUrl = "https://staging.api.com"; });
it("fetches", () => { fetch(apiUrl); /* uses staging URL from previous test */ });
```
- **Expected violation:** `apiUrl` reassigned in one test, read in another.

**TP-TI-08: Go test writing to package-level var**

- **Input:**
```go
var initialized = false
func TestInit(t *testing.T) { initialized = true }
func TestRun(t *testing.T) {
    if !initialized { t.Skip("not initialized") }
}
```
- **Expected violation:** Test order dependency.

**TP-TI-09: Class instance shared between tests**

- **Input:**
```typescript
const service = new UserService();
it("creates", () => { service.create("Alice"); });
it("lists", () => { expect(service.list()).toContain("Alice"); });
```
- **Expected violation:** Stateful instance shared between tests.

**TP-TI-10: Closure variable captured and mutated**

- **Input:**
```typescript
describe("counter", () => {
  let count = 0;
  it("starts at zero", () => { expect(count).toBe(0); count = 5; });
  it("was modified", () => { expect(count).toBe(5); });
});
```
- **Expected violation:** `count` mutated in first test, depended upon in second.

### 8.2 True Negative Cases

**TN-TI-01: Shared state set in beforeEach**

- **Input (with `allowSharedSetup: true`):**
```typescript
let user: User;
beforeEach(() => { user = createUser(validInput); });
it("has name", () => { expect(user.name).toBe("Alice"); });
it("has email", () => { expect(user.email).toBe("a@b.com"); });
```
- **Expected:** No violation. State is reset before each test.

**TN-TI-02: Const variable shared (immutable)**

- **Input:**
```typescript
const API_URL = "https://api.example.com";
it("fetches", () => { /* uses API_URL */ });
it("posts", () => { /* uses API_URL */ });
```
- **Expected:** No violation. `const` primitive is immutable.

**TN-TI-03: Go t.Run with local variables**

- **Input:**
```go
func TestOperations(t *testing.T) {
    t.Run("create", func(t *testing.T) {
        user := CreateUser("Alice")
        assert.Equal(t, "Alice", user.Name)
    })
    t.Run("delete", func(t *testing.T) {
        err := DeleteUser("Bob")
        assert.Error(t, err)
    })
}
```
- **Expected:** No violation. Each subtest has local state.

**TN-TI-04: Fresh object created in each test**

- **Input:**
```typescript
it("test 1", () => { const svc = new Service(); svc.process(); });
it("test 2", () => { const svc = new Service(); svc.process(); });
```
- **Expected:** No violation. Each test creates its own instance.

**TN-TI-05: Shared frozen object**

- **Input:**
```typescript
const defaults = Object.freeze({ port: 3000, host: "localhost" });
it("uses port", () => { expect(defaults.port).toBe(3000); });
it("uses host", () => { expect(defaults.host).toBe("localhost"); });
```
- **Expected:** No violation. Frozen objects cannot be mutated.

### 8.3 False Positive Risks

**FP-TI-01: Module-level constant array (never mutated)**

- `const VALID_STATUSES = ["active", "inactive"];` -- shared but never mutated.
- **Avoidance:** Only flag variables that are actually mutated inside test blocks.

**FP-TI-02: Shared database connection (stateless reference)**

- `const db = new Database(url);` -- the reference is shared but the database state is reset in beforeEach.
- **Avoidance:** Focus on mutations of the variable itself, not side effects through it (those are harder to detect statically).

**FP-TI-03: Accumulator for test results (afterAll report)**

- `const results: TestResult[] = [];` mutated in each test for reporting in afterAll.
- **Avoidance:** If the variable is only written (pushed to) and never read in tests, it may be a reporting accumulator.

**FP-TI-04: Event emitter subscriptions**

- Subscribing to events in beforeAll/beforeEach without unsubscribing.
- **Avoidance:** Focus on variable mutation, not subscription patterns.

**FP-TI-05: Let variable that is always reassigned before read**

- `let result; it("test", () => { result = compute(); expect(result).toBe(1); });`
- **Avoidance:** If every test that reads the variable also writes it first (in the same test block), no dependency.

### 8.4 False Negative Risks

**FN-TI-01: Side effect through a method call, not variable mutation**

- `it("test1", () => { db.insert("row"); }); it("test2", () => { expect(db.count()).toBe(1); });`
- **Severity:** High. Very common in integration tests.

**FN-TI-02: Environment variable mutation**

- `process.env.NODE_ENV = "test"` in one test affects others.
- **Severity:** Medium.

**FN-TI-03: Singleton pattern**

- Service uses singleton. First test initializes it, second test uses cached instance.
- **Severity:** Medium.

**FN-TI-04: Date.now() dependency without mock**

- Tests pass in sequence but fail when run in parallel due to timing.
- **Severity:** Low. Not detectable statically.

**FN-TI-05: Import side effects**

- Importing a module executes initialization code that sets global state.
- **Severity:** Low.

### 8.5 Edge Cases

**EC-TI-01:** Destructuring from shared object -- `const { port } = config;` creates a local copy of primitive.
**EC-TI-02:** Spread into new object -- `const localConfig = { ...config };` creates a shallow copy.
**EC-TI-03:** Generator function yielding shared state.
**EC-TI-04:** Promise stored in shared variable -- `let promise; it("starts", () => { promise = fetch(...); });`
**EC-TI-05:** WeakMap/WeakSet as shared state (GC-dependent).
**EC-TI-06:** Go sync.Once in test setup.
**EC-TI-07:** SharedArrayBuffer (extremely rare in tests).
**EC-TI-08:** Test that reads process.env (global but typically immutable).
**EC-TI-09:** Nested describe blocks with their own let variables.
**EC-TI-10:** Go parallel tests with shared fixture -- `t.Parallel()` with shared state is a race condition.

### 8.6 Configuration Interaction

**CI-TI-01:** `allowSharedSetup: true` -- shared state set in beforeEach is allowed.
**CI-TI-02:** `allowSharedSetup: false` -- even beforeEach shared state is flagged.
**CI-TI-03:** `checkFileSystem: true` -- file writes without cleanup are flagged.
**CI-TI-04:** `checkFileSystem: false` -- file system side effects ignored.

### 8.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-test-isolation`.

---

## 9. TQ-negative-cases

**Rule purpose:** Ensure that for every function with positive tests, there is at least one negative test.

### 9.1 True Positive Cases

**TP-NC-01: Only success tests for a function**

- **Input:**
```typescript
describe("createUser", () => {
  it("creates a user with valid input", () => { /* success */ });
  it("creates a user with minimal input", () => { /* success */ });
});
```
- **Expected violation:** `TQ-negative-cases`. Function `createUser` has 2 positive tests but 0 negative tests.

**TP-NC-02: Multiple functions, one without negative tests**

- **Input:**
```typescript
describe("createUser", () => {
  it("creates user", () => { /* success */ });
  it("rejects empty name", () => { expect(() => createUser({})).toThrow(); });
});
describe("deleteUser", () => {
  it("deletes user", () => { /* success */ });
  // No negative tests for deleteUser
});
```
- **Expected violation:** `deleteUser` has 1 positive test, 0 negative tests.

**TP-NC-03: Go test with only success case**

- **Input:**
```go
func TestParse(t *testing.T) {
    result, err := Parse("valid input")
    if err != nil { t.Fatal(err) }
    assert.Equal(t, "expected", result)
}
// No TestParse_Invalid, TestParse_Empty, etc.
```
- **Expected violation:** `Parse` has 1 positive test, 0 negative tests.

**TP-NC-04: File with all positive tests (no negative keywords)**

- **Input:**
```typescript
it("should return user data", () => { /* ... */ });
it("should return user list", () => { /* ... */ });
it("should return filtered results", () => { /* ... */ });
```
- **Expected violation:** All tests are positive. Zero negative tests in file.

**TP-NC-05: Go table-driven test with only success cases**

- **Input:**
```go
func TestValidate(t *testing.T) {
    tests := []struct{ input string; expected bool }{
        {"valid@email.com", true},
        {"another@test.com", true},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            assert.True(t, Validate(tt.input))
        })
    }
}
```
- **Expected violation:** All table entries are positive cases.

**TP-NC-06: Test names suggest negative but body is positive**

- **Input:**
```typescript
it("should handle error gracefully", () => {
  const result = handleError(validInput); // Actually success path
  expect(result.status).toBe(200);
});
```
- **Expected violation:** Despite the name, the body shows success behavior. The heuristic should examine both name and body.

**TP-NC-07: Many tests for a function, all positive**

- **Input:** 10 `it()` blocks all testing different valid inputs for the same function.
- **Expected violation:** 10 positive, 0 negative.

**TP-NC-08: React component test with only success renders**

- **Input:**
```typescript
it("renders user card", () => { render(<UserCard user={validUser} />); });
it("renders user card with avatar", () => { render(<UserCard user={userWithAvatar} />); });
// No test for missing user, error state, loading state
```
- **Expected violation:** Only positive render tests.

**TP-NC-09: Go function with error return, only success tested**

- **Input:**
```go
func TestFetch(t *testing.T) {
    data, err := Fetch("key")
    require.NoError(t, err)
    assert.Equal(t, "value", data)
}
```
- **Expected violation:** Function returns error type but no test exercises the error path.

**TP-NC-10: Test file with beforeEach creating valid state, no negative describe block**

- **Input:**
```typescript
describe("OrderService", () => {
  beforeEach(() => { /* setup valid order */ });
  it("calculates total", () => { /* success */ });
  it("applies discount", () => { /* success */ });
  it("calculates tax", () => { /* success */ });
});
```
- **Expected violation:** No negative tests in the describe block.

### 9.2 True Negative Cases

**TN-NC-01: Both positive and negative tests present**

- **Input:**
```typescript
it("creates user", () => { /* success */ });
it("rejects empty name", () => { expect(() => createUser({})).toThrow(); });
```
- **Expected:** No violation.

**TN-NC-02: Negative test with error assertion**

- **Input:**
```typescript
it("throws on invalid input", () => {
  expect(() => validate("")).toThrow(ValidationError);
});
```
- **Expected:** No violation (negative test present).

**TN-NC-03: Go test with both success and error cases**

- **Input:**
```go
func TestParse(t *testing.T) { /* success */ }
func TestParse_Invalid(t *testing.T) {
    _, err := Parse("invalid")
    assert.Error(t, err)
}
```
- **Expected:** No violation.

**TN-NC-04: Test with `.not.` assertion (negative case)**

- **Input:**
```typescript
it("rejects invalid email", () => {
  expect(isValid("not-an-email")).not.toBe(true);
});
```
- **Expected:** No violation. `.not.` assertion indicates negative case.

**TN-NC-05: Pure utility function with no error paths**

- **Source:** `function add(a: number, b: number): number { return a + b; }` -- no way to fail.
- **Expected:** No violation if the function has no error paths. (But may still be flagged if `perFunction: false` and the file has other functions.)

### 9.3 False Positive Risks

**FP-NC-01: Pure function that genuinely cannot fail**

- `function add(a: number, b: number): number` -- there is no meaningful negative case.
- **Avoidance:** If a function has no error exits, no validation, and only primitive params, consider it exempt.

**FP-NC-02: Negative tests in a separate file**

- Positive tests in `user.test.ts`, negative tests in `user.errors.test.ts`.
- **Avoidance:** `perFunction: true` should aggregate across files that test the same function.

**FP-NC-03: Integration test file (all positive by design)**

- Integration test suite that tests the happy path end-to-end. Negative integration tests are in a separate suite.
- **Avoidance:** Config option to exclude integration test patterns.

**FP-NC-04: Test for a constant/config (no behavior to test negatively)**

- `it("has correct default port", () => { expect(DEFAULT_PORT).toBe(3000); });`
- **Avoidance:** Constants are not functions. Rule only applies to function tests.

**FP-NC-05: Test name heuristic mismatch**

- `it("should not crash when given valid input", () => { /* success test */ })` -- name contains "not" but it is positive.
- **Avoidance:** Check body (assertions, error catching) not just name.

### 9.4 False Negative Risks

**FN-NC-01: Negative test that does not actually test the error**

- `it("handles error", () => { const result = fn("bad"); expect(result).toBeDefined(); });` -- name says error, body tests success.
- **Severity:** High.

**FN-NC-02: Generic try/catch that swallows error**

- `it("handles invalid", () => { try { fn(""); } catch {} });` -- catches error but asserts nothing.
- **Severity:** High.

**FN-NC-03: Negative test for wrong function**

- File tests `createUser` and `deleteUser`. Has negative test for `deleteUser` but not `createUser`.
- **Severity:** Medium. `perFunction: true` catches this.

**FN-NC-04: Test that returns without assertion**

- `it("throws on bad input", async () => { await fn("bad"); });` -- no expect/assert at all.
- **Severity:** Medium.

**FN-NC-05: Mocked error that does not exercise real error path**

- `jest.mock("../db", () => ({ query: jest.fn().mockRejectedValue(new Error()) }));` -- mock throws but real code path is not tested.
- **Severity:** Medium.

### 9.5 Edge Cases

**EC-NC-01:** Test file for a module with only one exported function.
**EC-NC-02:** Multiple describe blocks for the same function (positive in one, negative in another).
**EC-NC-03:** Go TestMain function (setup, not a test case).
**EC-NC-04:** Test with conditional assertions based on environment.
**EC-NC-05:** Parameterized tests mixing positive and negative in same test block.
**EC-NC-06:** Vitest `it.skip` on a negative test -- should it count?
**EC-NC-07:** Test that asserts a promise resolves (positive) and another that asserts it rejects (negative).
**EC-NC-08:** Go subtest with positive and negative cases in same table.
**EC-NC-09:** Assertion that a function returns `null` on invalid input (is this positive or negative?).
**EC-NC-10:** Test for an event emitter -- "emits error event" is negative.

### 9.6 Configuration Interaction

**CI-NC-01:** `minNegativeRatio: 0.3` -- at least 30% of tests must be negative.
**CI-NC-02:** `minNegativeRatio: 0.0` -- at least 1 negative test (presence check, not ratio).
**CI-NC-03:** `perFunction: true` -- apply per function under test.
**CI-NC-04:** `perFunction: false` -- apply per file (file-level ratio).

### 9.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-negative-cases`.

---

## 10. TQ-test-naming

**Rule purpose:** Enforce consistent, descriptive test names.

### 10.1 True Positive Cases

**TP-TN-01: Vague name "works"**

- **Input:**
```typescript
it("works", () => { /* ... */ });
```
- **Expected violation:** `TQ-test-naming`. Name "works" is too vague (in blockWords).

**TP-TN-02: Generic name "test 1"**

- **Input:**
```typescript
it("test 1", () => { /* ... */ });
```
- **Expected violation:** "test 1" is in blockWords.

**TP-TN-03: Name too short**

- **Input (with `minLength: 15`):**
```typescript
it("adds user", () => { /* ... */ });
```
- **Expected violation:** "adds user" is 9 characters, below 15 minimum.

**TP-TN-04: Name does not match pattern**

- **Config:** `pattern: "should {verb} when {condition}"`
- **Input:**
```typescript
it("creates a user", () => { /* ... */ });
```
- **Expected violation:** Does not match "should ... when ..." pattern.

**TP-TN-05: Blocked word "simple"**

- **Input:**
```typescript
it("simple test", () => { /* ... */ });
```
- **Expected violation:** "simple" is in blockWords.

**TP-TN-06: Go test with non-descriptive name**

- **Input:**
```go
func TestIt(t *testing.T) { /* ... */ }
```
- **Expected violation:** "It" is too vague and short.

**TP-TN-07: Vague "edge case" without specifics**

- **Input:**
```typescript
it("should handle edge case", () => { /* ... */ });
```
- **Expected violation:** "edge case" without specifics is in blockWords.

**TP-TN-08: Duplicate test names**

- **Input:**
```typescript
it("should create user", () => { /* ... */ });
it("should create user", () => { /* ... */ });
```
- **Expected violation:** Duplicate test name.

**TP-TN-09: Name is just the function name**

- **Input:**
```typescript
it("createUser", () => { /* ... */ });
```
- **Expected violation:** Not descriptive of behavior; just repeats function name.

**TP-TN-10: Go t.Run with empty string**

- **Input:**
```go
t.Run("", func(t *testing.T) { /* ... */ })
```
- **Expected violation:** Empty test name.

### 10.2 True Negative Cases

**TN-TN-01: Descriptive name matching pattern**

- **Config:** `pattern: "should {verb} when {condition}"`
- **Input:** `it("should return 404 when user ID does not exist", () => { });`
- **Expected:** No violation.

**TN-TN-02: Long descriptive name**

- **Input:** `it("should truncate names longer than 255 characters", () => { });`
- **Expected:** No violation.

**TN-TN-03: Go test with descriptive name**

- **Input:** `func TestCreateUser_WithValidInput_ReturnsUser(t *testing.T) { }`
- **Expected:** No violation.

**TN-TN-04: Go t.Run with descriptive name**

- **Input:** `t.Run("returns error for empty input", func(t *testing.T) { })`
- **Expected:** No violation.

**TN-TN-05: Test name with specific values**

- **Input:** `it("should return 3 items when pageSize is 3 and total is 10", () => { });`
- **Expected:** No violation.

### 10.3 False Positive Risks

**FP-TN-01: Short but descriptive name for a simple function**

- `it("adds two numbers", () => { });` -- 15 chars exactly. May be too short for strict `minLength` but is perfectly descriptive.
- **Avoidance:** minLength should be reasonable (10-15).

**FP-TN-02: Pattern mismatch for negative test**

- Pattern is "should {verb} when {condition}" but negative test is `it("throws ValidationError for empty name", () => { })`. Does not start with "should".
- **Avoidance:** Pattern should be flexible or have multiple allowed patterns.

**FP-TN-03: Describe block provides context**

- `describe("createUser", () => { it("with valid input", () => { }); });` -- the full name is "createUser with valid input" which is descriptive.
- **Avoidance:** Concatenate describe + it name for validation.

**FP-TN-04: Technical test name for low-level unit**

- `it("handles null byte at position 0", () => { })` -- very specific, may not match "should...when" pattern.
- **Avoidance:** Pattern matching should be flexible.

**FP-TN-05: Non-English test names**

- Team writes test names in their native language.
- **Avoidance:** blockWords are English-specific; non-English names should not match blockWords.

### 10.4 False Negative Risks

**FN-TN-01: Long but meaningless name**

- `it("should do the thing when the other thing happens correctly", () => { })` -- long but vague.
- **Severity:** Medium. Hard to detect vagueness beyond blockWords.

**FN-TN-02: Copy-pasted name with small difference**

- `it("should create user 1", () => { }); it("should create user 2", () => { });`
- **Severity:** Low. Technically different names but no semantic difference.

**FN-TN-03: Name describes implementation, not behavior**

- `it("should call fetch with correct URL", () => { })` -- describes how, not what.
- **Severity:** Low. Hard to detect statically.

**FN-TN-04: Test name does not match test body**

- `it("should return error", () => { expect(result).toBe(200); })` -- name says error, body says success.
- **Severity:** Medium. Requires semantic analysis.

**FN-TN-05: Go test name that follows convention but is undescriptive**

- `func TestProcess(t *testing.T)` -- follows Go convention but says nothing about behavior.
- **Severity:** Medium.

### 10.5 Edge Cases

**EC-TN-01:** Template literal test name: `` it(`should handle ${type}`, () => { }) `` -- cannot statically evaluate.
**EC-TN-02:** Test name with special characters: `it("handles 'quoted' strings", () => { })`.
**EC-TN-03:** Describe block name is very long, it name is short.
**EC-TN-04:** Go test function with underscore convention: `TestCreate_InvalidInput_ReturnsError`.
**EC-TN-05:** Mocha-style `describe`/`it` vs Jest `describe`/`test`.
**EC-TN-06:** Node.js test runner `test("name", () => { })`.
**EC-TN-07:** Vitest `it.concurrent("name", () => { })`.
**EC-TN-08:** Go benchmark `func BenchmarkProcess(b *testing.B)` -- different naming convention.
**EC-TN-09:** Test name containing only numbers: `it("123", () => { })`.
**EC-TN-10:** Test inside `it.each` / parameterized: `it.each(cases)("should handle %s", (input) => { })`.

### 10.6 Configuration Interaction

**CI-TN-01:** `pattern: "should {verb} when {condition}"` -- names must match this pattern.
**CI-TN-02:** `pattern: null` -- no pattern enforcement, only minLength and blockWords.
**CI-TN-03:** `minLength: 10` -- names shorter than 10 chars are flagged.
**CI-TN-04:** `minLength: 0` -- no length requirement.
**CI-TN-05:** `blockWords: ["works", "basic"]` -- custom blocklist.
**CI-TN-06:** `blockWords: []` -- no blocklist.

### 10.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `TQ-test-naming`.

---

## 11. ARCH-dependency-direction

**Rule purpose:** Enforce unidirectional dependency flow between architectural layers.

### 11.1 True Positive Cases

**TP-DD-01: Service imports from handler**

- **Config:**
```yaml
ARCH-dependency-direction:
  - error
  - layers:
      - { name: handler, patterns: ["src/routes/**"] }
      - { name: service, patterns: ["src/services/**"] }
      - { name: repository, patterns: ["src/repositories/**"] }
    direction: top-down
```
- **Input** (`src/services/user-service.ts`):
```typescript
import { userRouter } from "../routes/users";
```
- **Expected violation:** `ARCH-dependency-direction`. Service layer imports from handler layer (reverse direction).

**TP-DD-02: Repository imports from service**

- **Input** (`src/repositories/user-repo.ts`):
```typescript
import { UserService } from "../services/user-service";
```
- **Expected violation:** Repository layer imports from service layer.

**TP-DD-03: Repository imports from handler (skip two layers)**

- **Input** (`src/repositories/user-repo.ts`):
```typescript
import { handleRequest } from "../routes/users";
```
- **Expected violation:** Repository imports from handler (two layers above).

**TP-DD-04: Go handler imports from repository directly (config-dependent)**

- **Config:** `direction: top-down` with layers handler > service > repository.
- **Input** (`cmd/server/handler.go`):
```go
import "internal/repo"
```
- **Expected violation:** If the rule enforces that handler must go through service (strict layering), this skips a layer. However, top-down only means higher layers can import lower layers. handler > repository is still top-down. Whether "skip layer" is a violation depends on interpretation.
- **Note:** This is a spec ambiguity. See Part 3.

**TP-DD-05: Model imports from handler**

- **Config:** Layers: handler > service > repository > model.
- **Input** (`src/models/user.ts`):
```typescript
import { parseRequest } from "../routes/users";
```
- **Expected violation:** Model (lowest) imports from handler (highest).

**TP-DD-06: Mutual import between layers**

- **Input:** Service imports from repository (OK) AND repository imports from service (VIOLATION).
- **Expected violation:** Only the repository-to-service import is flagged.

**TP-DD-07: Go internal package reverse import**

- **Config:** Layers: cmd > internal/service > internal/repo.
- **Input** (`internal/repo/store.go`):
```go
import "project/internal/service"
```
- **Expected violation:** Repo imports service (reverse).

**TP-DD-08: Dynamic import that violates direction**

- **Input** (`src/services/user.ts`):
```typescript
const handler = await import("../routes/users");
```
- **Expected violation:** Dynamic import still violates layer direction.

**TP-DD-09: Re-export chain that pulls in wrong layer**

- **Input** (`src/services/index.ts`):
```typescript
export { userRouter } from "../routes/users"; // re-exporting handler code
```
- **Expected violation:** Re-export creates a dependency from service to handler.

**TP-DD-10: Side-effect import from wrong layer**

- **Input** (`src/repositories/init.ts`):
```typescript
import "../routes/setup"; // side-effect import
```
- **Expected violation:** Side-effect import still creates dependency.

### 11.2 True Negative Cases

**TN-DD-01: Handler imports from service (correct direction)**

- **Input** (`src/routes/users.ts`):
```typescript
import { UserService } from "../services/user-service";
```
- **Expected:** No violation.

**TN-DD-02: Service imports from repository**

- **Input** (`src/services/user-service.ts`):
```typescript
import { UserRepo } from "../repositories/user-repo";
```
- **Expected:** No violation.

**TN-DD-03: Handler imports from model (skipping layers but correct direction)**

- **Expected:** No violation for top-down. Handler is above model.

**TN-DD-04: Intra-layer import**

- **Input** (`src/services/auth-service.ts`):
```typescript
import { hashPassword } from "./crypto-service";
```
- **Expected:** No violation. Same-layer imports are allowed.

**TN-DD-05: File not in any layer**

- **Input** (`src/utils/helpers.ts`):
```typescript
import { UserService } from "../services/user-service";
```
- **Expected:** No violation. `utils` is not in any defined layer.

### 11.3 False Positive Risks

**FP-DD-01: Shared types imported by all layers**

- A `types.ts` file imported by both handler and repository is not a violation -- it is not in any layer.
- **Avoidance:** Only classify files that match layer patterns. Unmatched files are layer-free.

**FP-DD-02: Test files importing from any layer**

- Test files should be exempt from layer rules.
- **Avoidance:** Exclude test files from layer classification.

**FP-DD-03: Type-only imports (TypeScript)**

- `import type { User } from "../routes/types";` -- imports only the type, no runtime dependency.
- **Avoidance:** Consider `import type` as exempt since it creates no runtime dependency.

**FP-DD-04: Circular layer definition in config**

- If layers are misconfigured such that A > B > A, everything is a violation.
- **Avoidance:** Validate config for circular layer definitions at startup (exit code 2).

**FP-DD-05: Monorepo package imports**

- `import { util } from "@company/shared"` -- cross-package import looks like it might violate direction.
- **Avoidance:** Layer patterns should be scoped to the project, not external packages.

### 11.4 False Negative Risks

**FN-DD-01: Indirect dependency through a shared utility**

- Handler imports util. Util imports from repo. Effectively handler depends on repo through util.
- **Severity:** Medium. Requires transitive dependency analysis.

**FN-DD-02: Dependency through event emitter/pubsub**

- Service emits event, handler subscribes. No import, but logical dependency.
- **Severity:** Low. Cannot detect without runtime analysis.

**FN-DD-03: Require() calls instead of import**

- `const handler = require("../routes/users");` -- must be detected like import.
- **Severity:** Medium.

**FN-DD-04: Go build tag conditional imports**

- `//go:build integration` may include imports not present in normal builds.
- **Severity:** Low.

**FN-DD-05: Dependency injection hiding the import**

- Service receives handler via constructor injection; the import is in the composition root, not in the service file.
- **Severity:** Low.

### 11.5 Edge Cases

**EC-DD-01:** File matching multiple layer patterns -- which layer wins? (First match? Most specific?)
**EC-DD-02:** Symlinked file that resolves to a different layer.
**EC-DD-03:** Barrel file (index.ts) that re-exports from multiple layers.
**EC-DD-04:** Go vendor/ directory imports.
**EC-DD-05:** TypeScript path aliases (e.g., `@/services/user` -> `src/services/user`).
**EC-DD-06:** Relative import with `..` that traverses through layer boundaries.
**EC-DD-07:** Generated code in a layer directory.
**EC-DD-08:** Only two layers defined (handler > repository, no service).
**EC-DD-09:** Single file in a layer directory.
**EC-DD-10:** Layer with no files matching its pattern.

### 11.6 Configuration Interaction

**CI-DD-01:** `direction: top-down` -- higher layers can import lower, not reverse.
**CI-DD-02:** `direction: bottom-up` -- (if supported) lower layers can import higher.
**CI-DD-03:** Layer order matters -- first layer is "highest."
**CI-DD-04:** Missing `direction` key -- should default to top-down or error.
**CI-DD-05:** Empty layers array -- rule is effectively disabled.

### 11.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `ARCH-dependency-direction`.

---

## 12. ARCH-import-boundary

**Rule purpose:** Enforce explicit module boundaries -- certain directories cannot import from certain other directories.

### 12.1 True Positive Cases

**TP-IB-01: Content script imports from background**

- **Config:**
```yaml
ARCH-import-boundary:
  - error
  - boundaries:
      - { from: "src/content/**", deny: ["src/background/**"] }
```
- **Input** (`src/content/inject.ts`):
```typescript
import { sendMessage } from "../background/messaging";
```
- **Expected violation:** `ARCH-import-boundary`. Content scripts cannot import background modules.

**TP-IB-02: CLI imports from internal capture**

- **Config:** `{ from: "cmd/**", deny: ["internal/capture/**"] }`
- **Input** (`cmd/main.go`):
```go
import "project/internal/capture"
```
- **Expected violation:** CLI entry points should go through service layer.

**TP-IB-03: Source imports from test files**

- **Config:** `{ from: "src/**", deny: ["tests/**", "**/*.test.*"] }`
- **Input** (`src/utils/helper.ts`):
```typescript
import { mockData } from "../../tests/fixtures";
```
- **Expected violation:** Source code must not import from test files.

**TP-IB-04: Multiple deny patterns, one matches**

- **Config:**
```yaml
boundaries:
  - from: "src/frontend/**"
    deny: ["src/backend/**", "src/database/**"]
```
- **Input** (`src/frontend/app.ts`):
```typescript
import { query } from "../database/connection";
```
- **Expected violation:** Frontend imports from database.

**TP-IB-05: Go package boundary violation**

- **Config:** `{ from: "pkg/client/**", deny: ["internal/**"] }`
- **Input** (`pkg/client/api.go`):
```go
import "project/internal/secret"
```
- **Expected violation:** Public package imports internal package.

**TP-IB-06: Deep nested file violating boundary**

- **Config:** `{ from: "src/content/**", deny: ["src/background/**"] }`
- **Input** (`src/content/deep/nested/component.ts`):
```typescript
import { bgFunc } from "../../../background/api";
```
- **Expected violation:** Deep nesting does not exempt from boundary.

**TP-IB-07: Require() call violating boundary**

- **Input** (`src/content/loader.ts`):
```typescript
const bg = require("../background/service");
```
- **Expected violation:** `require()` is also an import.

**TP-IB-08: Dynamic import violating boundary**

- **Input** (`src/content/lazy.ts`):
```typescript
const mod = await import("../background/heavy-module");
```
- **Expected violation:** Dynamic import still violates boundary.

**TP-IB-09: Re-export from denied module**

- **Input** (`src/content/index.ts`):
```typescript
export { handler } from "../background/handler";
```
- **Expected violation:** Re-export creates dependency across boundary.

**TP-IB-10: Type import from denied module**

- **Input** (`src/content/types.ts`):
```typescript
import type { MessageHandler } from "../background/types";
```
- **Expected:** This depends on config. If type-only imports are treated as boundary violations, then yes. By default, flag it -- type dependencies can indicate architectural coupling.

### 12.2 True Negative Cases

**TN-IB-01: Import from allowed module**

- **Input** (`src/content/app.ts`):
```typescript
import { util } from "../shared/utils";
```
- **Expected:** No violation. `shared` is not in deny list.

**TN-IB-02: File not in "from" pattern**

- **Input** (`src/shared/helper.ts`):
```typescript
import { bgFunc } from "../background/api";
```
- **Expected:** No violation. `src/shared/**` is not in any `from` pattern.

**TN-IB-03: Import from same directory**

- **Input** (`src/content/a.ts`):
```typescript
import { b } from "./b";
```
- **Expected:** No violation. Same directory import.

**TN-IB-04: Go test file importing from internal (test exception)**

- **Input** (`cmd/main_test.go`):
```go
import "project/internal/capture"
```
- **Expected:** Test files should be exempt from boundary rules.

**TN-IB-05: External package import**

- **Input** (`src/content/app.ts`):
```typescript
import express from "express";
```
- **Expected:** No violation. External packages are not subject to internal boundaries.

### 12.3 False Positive Risks

**FP-IB-01: Type-only import across boundary**

- Some teams consider type imports acceptable. Config should allow exempting `import type`.

**FP-IB-02: Test utility in src directory**

- A test helper file in `src/testing/` that imports from `tests/`. Technically a violation but intentional.
- **Avoidance:** Add to ignore list.

**FP-IB-03: Barrel file that re-exports from allowed sources**

- `src/content/index.ts` exports from `src/content/internal/` which is fine, but the glob pattern might be too broad.

**FP-IB-04: Monorepo package treated as internal**

- `import { util } from "../../packages/shared"` looks like it might cross a boundary.
- **Avoidance:** Only apply boundaries within the project root.

**FP-IB-05: Generated code importing across boundary**

- Protobuf-generated code may import from any package.
- **Avoidance:** Exclude generated files via `ignore` patterns.

### 12.4 False Negative Risks

**FN-IB-01: Transitive import through a shared module**

- A imports from shared. Shared imports from denied. A indirectly depends on denied.
- **Severity:** Medium.

**FN-IB-02: Global require with string concatenation**

- `require("../back" + "ground/api")` -- cannot statically resolve.
- **Severity:** Low.

**FN-IB-03: Go plugin loading at runtime**

- Plugin loaded with `plugin.Open()` bypasses import analysis.
- **Severity:** Low.

**FN-IB-04: Webpack/bundler alias that hides real path**

- `import { x } from "@bg/api"` where `@bg` maps to `src/background`.
- **Severity:** Medium. Must resolve aliases.

**FN-IB-05: Side-effect-only import of a denied module**

- `import "../background/polyfill";` -- no named imports but still creates dependency.
- **Severity:** Medium.

### 12.5 Edge Cases

**EC-IB-01:** Overlapping `from` patterns -- file matches multiple boundary rules.
**EC-IB-02:** `deny` pattern that matches the `from` pattern (deny self-imports).
**EC-IB-03:** Empty `deny` array.
**EC-IB-04:** Wildcard deny: `deny: ["**"]` blocks all imports.
**EC-IB-05:** Go internal/ convention (already enforced by Go compiler).
**EC-IB-06:** Symlinked directory that resolves to denied path.
**EC-IB-07:** `.d.ts` file importing from denied module.
**EC-IB-08:** CSS/JSON imports that cross boundary.
**EC-IB-09:** `reason` field displayed in violation message.
**EC-IB-10:** Boundary rule with regex in `from` or `deny`.

### 12.6 Configuration Interaction

**CI-IB-01:** Multiple boundary rules -- all are checked independently.
**CI-IB-02:** Empty boundaries array -- rule is effectively disabled.
**CI-IB-03:** `reason` field -- included in violation message.
**CI-IB-04:** Overlapping from/deny patterns with conflicting rules.

### 12.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `ARCH-import-boundary`.

---

## 13. ARCH-no-circular-deps

**Rule purpose:** Detect and reject circular import dependencies.

### 13.1 True Positive Cases

**TP-NCD-01: Direct circular import (A -> B -> A)**

- **Input:**
  - `src/a.ts`: `import { b } from "./b";`
  - `src/b.ts`: `import { a } from "./a";`
- **Expected violation:** `ARCH-no-circular-deps`. Cycle: a.ts -> b.ts -> a.ts.

**TP-NCD-02: Three-file cycle (A -> B -> C -> A)**

- **Input:**
  - `src/a.ts`: `import "./b";`
  - `src/b.ts`: `import "./c";`
  - `src/c.ts`: `import "./a";`
- **Expected violation:** Cycle: a.ts -> b.ts -> c.ts -> a.ts.

**TP-NCD-03: Go package cycle**

- **Input:**
  - `pkg/auth/auth.go`: `import "project/pkg/user"`
  - `pkg/user/user.go`: `import "project/pkg/auth"`
- **Expected violation:** Cycle between packages.

**TP-NCD-04: Self-import**

- **Input** (`src/a.ts`): `import { x } from "./a";`
- **Expected violation:** Self-cycle.

**TP-NCD-05: Cycle through barrel files**

- `src/features/auth/index.ts` exports from `./login`. `src/features/auth/login.ts` imports from `../user`. `src/features/user/index.ts` imports from `../auth`.
- **Expected violation:** auth -> user -> auth.

**TP-NCD-06: Long cycle (5 files)**

- A -> B -> C -> D -> E -> A.
- **Expected violation:** Reports the shortest cycle path.

**TP-NCD-07: Multiple cycles in the same graph**

- A -> B -> A, and C -> D -> C.
- **Expected violation:** Two separate violations reported.

**TP-NCD-08: Go test file creating cycle**

- If including test files: `a_test.go` imports `b`, `b.go` imports `a`.
- **Expected:** Depends on whether test files are included in cycle detection. By default, test file imports should be included.

**TP-NCD-09: Dynamic import creating cycle**

- `src/a.ts`: `const b = await import("./b");`
- `src/b.ts`: `import { a } from "./a";`
- **Expected violation:** Dynamic imports count.

**TP-NCD-10: Re-export creating cycle**

- `src/a.ts`: `export { x } from "./b";`
- `src/b.ts`: `export { y } from "./a";`
- **Expected violation:** Re-exports create dependency.

### 13.2 True Negative Cases

**TN-NCD-01: Linear dependency chain**

- A -> B -> C (no back-edge).
- **Expected:** No violation.

**TN-NCD-02: Diamond dependency (no cycle)**

- A -> B, A -> C, B -> D, C -> D.
- **Expected:** No violation.

**TN-NCD-03: External package import**

- `import lodash from "lodash";` -- external packages are not in the cycle graph.
- **Expected:** No violation.

**TN-NCD-04: Type-only import (questionable)**

- `import type { X } from "./b";` in a.ts and `import type { Y } from "./a";` in b.ts.
- **Expected:** This is debatable. Type-only imports create no runtime cycle. Consider no violation by default.

**TN-NCD-05: File with no imports**

- Cannot participate in a cycle.
- **Expected:** No violation.

### 13.3 False Positive Risks

**FP-NCD-01: Type-only circular import**

- TypeScript `import type` does not create runtime circular dependency.
- **Avoidance:** Option to exclude type-only imports from cycle detection.

**FP-NCD-02: Circular dependency that is handled correctly at runtime**

- Some circular deps work fine in Node.js (e.g., A initializes before B accesses A).
- **Avoidance:** This rule flags all cycles regardless. Users can suppress per-file.

**FP-NCD-03: Test file importing source that imports test utility**

- test/a.test.ts -> src/a.ts -> test/utils.ts -> test/a.test.ts.
- **Avoidance:** Exclude test files from cycle detection or treat them separately.

**FP-NCD-04: Declaration file (.d.ts) in cycle**

- `.d.ts` files are type-only and cannot create runtime cycles.
- **Avoidance:** Exclude `.d.ts` files.

**FP-NCD-05: Monorepo workspace packages with legitimate cross-references**

- Package A depends on Package B and vice versa (separate packages).
- **Avoidance:** Scope cycle detection to within a single package/workspace.

### 13.4 False Negative Risks

**FN-NCD-01: require() based cycle**

- `const a = require("./a")` creates a cycle not caught if only ES import is analyzed.
- **Severity:** Medium.

**FN-NCD-02: Dynamic string concatenation in require**

- `require("./" + name)` -- cannot resolve statically.
- **Severity:** Low.

**FN-NCD-03: Webpack lazy loading that hides cycle**

- Webpack code splitting may break a cycle at bundle time but the logical dependency exists.
- **Severity:** Low.

**FN-NCD-04: Go interface-based decoupling**

- Package A imports interface from shared, Package B implements it. Logically coupled but no import cycle.
- **Severity:** Low. Not a cycle.

**FN-NCD-05: CSS module import cycle**

- If CSS files are included in analysis.
- **Severity:** Low.

### 13.5 Edge Cases

**EC-NCD-01:** Very large cycle (100 files) -- report only the shortest cycle.
**EC-NCD-02:** Multiple overlapping cycles -- report each unique SCC.
**EC-NCD-03:** File that imports itself via absolute path vs relative.
**EC-NCD-04:** TypeScript path mapping that resolves differently.
**EC-NCD-05:** Go vendor imports.
**EC-NCD-06:** Conditional imports (`if (condition) import("./a")`).
**EC-NCD-07:** Import inside a function body (not top-level).
**EC-NCD-08:** Cycle involving only index files.
**EC-NCD-09:** File removed but cached -- stale cycle detection.
**EC-NCD-10:** Single-file project -- no possible cycle.

### 13.6 Configuration Interaction

Rule has no options in the spec. Test:
**CI-NCD-01:** Rule set to "error" -- exit code 1 on cycle.
**CI-NCD-02:** Rule set to "warn" -- warning only.
**CI-NCD-03:** Rule set to "off" -- no cycle detection.

### 13.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `ARCH-no-circular-deps`.

---

## 14. ARCH-max-file-lines

**Rule purpose:** Enforce maximum file size.

### 14.1 True Positive Cases

**TP-MFL-01: File exceeds 800 lines (default)**

- **Input:** A TypeScript file with 801 lines.
- **Expected violation:** `ARCH-max-file-lines`. File has 801 lines, maximum is 800.

**TP-MFL-02: Go file exceeds limit**

- **Input:** A Go file with 850 lines.
- **Expected violation:** File has 850 lines.

**TP-MFL-03: File exactly at limit + 1**

- **Input:** File with exactly 801 lines (max: 800).
- **Expected violation:** 801 > 800.

**TP-MFL-04: Test file exceeds override limit**

- **Config:** `overrides: { "**/*.test.*": 1200 }`
- **Input:** `user.test.ts` with 1201 lines.
- **Expected violation:** Test file exceeds 1200-line override limit.

**TP-MFL-05: File with many blank lines still over limit (excludeBlankLines: false)**

- **Config:** `excludeBlankLines: false`
- **Input:** File with 600 code lines and 201 blank lines = 801 total.
- **Expected violation:** 801 > 800.

**TP-MFL-06: File over limit even excluding comments**

- **Config:** `excludeComments: true`
- **Input:** File with 850 code lines and 100 comment lines = 950 total. 850 non-comment lines > 800.
- **Expected violation:** 850 non-comment lines > 800.

**TP-MFL-07: Go test file exceeds override**

- **Config:** `overrides: { "**/*_test.go": 1200 }`
- **Input:** `user_test.go` with 1250 lines.
- **Expected violation:** 1250 > 1200.

**TP-MFL-08: Very large generated file**

- **Input:** `schema.generated.ts` with 5000 lines (not in ignore list).
- **Expected violation:** 5000 > 800. Unless file matches ignore pattern.

**TP-MFL-09: File barely over after excluding blank lines**

- **Config:** `excludeBlankLines: true`
- **Input:** File with 810 total lines, 5 blank lines = 805 non-blank > 800.
- **Expected violation:** 805 > 800.

**TP-MFL-10: Multiple files over limit**

- **Input:** Three files each with 900 lines.
- **Expected:** Three separate violations.

### 14.2 True Negative Cases

**TN-MFL-01: File at exactly 800 lines**

- **Expected:** No violation. 800 <= 800.

**TN-MFL-02: File with 799 lines**

- **Expected:** No violation.

**TN-MFL-03: File over limit but excluded by ignore pattern**

- **Config:** `ignore: ["*.generated.ts"]`
- **Input:** `schema.generated.ts` with 2000 lines.
- **Expected:** No violation (file is ignored).

**TN-MFL-04: Test file at override limit**

- **Config:** `overrides: { "**/*.test.*": 1200 }`
- **Input:** `user.test.ts` with 1200 lines.
- **Expected:** No violation. 1200 <= 1200.

**TN-MFL-05: File over total but under after excluding comments**

- **Config:** `excludeComments: true`
- **Input:** File with 900 total lines, 150 comment lines = 750 non-comment.
- **Expected:** No violation. 750 <= 800.

### 14.3 False Positive Risks

**FP-MFL-01: Generated code files**

- Generated files (protobuf, OpenAPI) are often large by necessity.
- **Avoidance:** `ignore` patterns or `overrides` for generated file patterns.

**FP-MFL-02: Data files (large constant arrays)**

- A file containing a large lookup table.
- **Avoidance:** User should configure override or ignore.

**FP-MFL-03: Type definition files**

- `.d.ts` files with many type declarations.
- **Avoidance:** Ignore `.d.ts` by default (per spec's ignore defaults).

**FP-MFL-04: Test files with many test cases**

- Comprehensive test files are naturally longer.
- **Avoidance:** Override limits for test files.

**FP-MFL-05: File with many JSDoc comments inflating line count**

- If `excludeComments: false`, extensive documentation increases count unfairly.
- **Avoidance:** Enable `excludeComments: true`.

### 14.4 False Negative Risks

**FN-MFL-01: Multiple small files that should be one larger file**

- 10 files of 100 lines each that all belong together.
- **Severity:** Low. Not this rule's job.

**FN-MFL-02: Long lines instead of more lines**

- 200 lines with 500-char lines. File is "short" but unreadable.
- **Severity:** Low. Line length is a different concern.

**FN-MFL-03: File close to limit growing over time**

- 790-line file will eventually exceed. No early warning.
- **Severity:** Low. Could add a "warning at 90%" option.

**FN-MFL-04: Concatenated output file**

- Build process creates a single large file from many sources.
- **Severity:** Low. Should be in ignore list.

**FN-MFL-05: Comment-stripped count misses embedded documentation**

- `excludeComments: true` might exclude inline documentation that is effectively code.
- **Severity:** Low.

### 14.5 Edge Cases

**EC-MFL-01:** Empty file -- 0 lines. No violation.
**EC-MFL-02:** File with only comments -- 0 non-comment lines if excludeComments.
**EC-MFL-03:** File with only blank lines.
**EC-MFL-04:** File with mixed line endings (CRLF + LF) -- how are lines counted?
**EC-MFL-05:** File ending without newline -- does the last line count?
**EC-MFL-06:** Binary file accidentally included -- should be skipped.
**EC-MFL-07:** Very long single line (no newlines).
**EC-MFL-08:** File with BOM marker -- does BOM affect line count?
**EC-MFL-09:** Shebang line (`#!/usr/bin/env node`) -- counted as a line.
**EC-MFL-10:** Override pattern matching: most specific pattern wins.

### 14.6 Configuration Interaction

**CI-MFL-01:** `max: 500` -- files over 500 lines are violations.
**CI-MFL-02:** `excludeComments: true` + `excludeBlankLines: true` -- only code lines counted.
**CI-MFL-03:** `excludeComments: false` + `excludeBlankLines: false` -- all lines counted.
**CI-MFL-04:** Override for test files: `"**/*.test.*": 1200`.
**CI-MFL-05:** Multiple overrides -- most specific pattern takes precedence.

### 14.7 Inline Suppression Testing

- `// stricture-disable-next-line ARCH-max-file-lines` -- does this even make sense? The violation is on the file, not a specific line.
- **Expected behavior:** File-level rules should be suppressible with a comment at the top of the file.

---

## 15. ARCH-layer-violation

**Rule purpose:** Detect when code at one architectural layer performs responsibilities belonging to another layer.

### 15.1 True Positive Cases

**TP-LV-01: Handler directly calls database query**

- **Config:**
```yaml
ARCH-layer-violation:
  - error
  - layers:
      handler:
        patterns: ["src/routes/**"]
        forbiddenCalls: ["sql.*", "db.*", "*.Query", "*.Exec"]
```
- **Input** (`src/routes/users.ts`):
```typescript
import { db } from "../database";
app.get("/users", async (req, res) => {
  const users = await db.Query("SELECT * FROM users");
  res.json(users);
});
```
- **Expected violation:** `ARCH-layer-violation` at the `db.Query` line. Handler directly calls database.

**TP-LV-02: Handler executes raw SQL**

- **Input** (`src/routes/orders.ts`):
```typescript
const result = await sql.raw("INSERT INTO orders VALUES (...)");
```
- **Expected violation:** `sql.*` matches `sql.raw`.

**TP-LV-03: Repository makes HTTP call**

- **Config:**
```yaml
layers:
  repository:
    patterns: ["src/repositories/**"]
    forbiddenCalls: ["http.*", "fetch", "axios.*"]
```
- **Input** (`src/repositories/user-repo.ts`):
```typescript
const response = await fetch("https://api.example.com/users");
```
- **Expected violation:** Repository makes HTTP call.

**TP-LV-04: Go handler directly executes database query**

- **Input** (`cmd/server/handler.go`):
```go
func HandleGetUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT * FROM users")
    // ...
}
```
- **Expected violation:** Handler calls `db.Query`.

**TP-LV-05: Handler calls Exec (database write)**

- **Input** (`src/routes/users.ts`):
```typescript
await db.Exec("DELETE FROM users WHERE id = $1", id);
```
- **Expected violation:** `*.Exec` matches `db.Exec`.

**TP-LV-06: Repository uses net/http**

- **Input** (`internal/repo/external.go`):
```go
import "net/http"
resp, err := http.Get("https://external-api.com/data")
```
- **Expected violation:** Repository calls `net/http.*`.

**TP-LV-07: Handler creates database connection**

- **Input:**
```typescript
const pool = new sql.Pool({ connectionString: "..." });
```
- **Expected violation:** `sql.*` matches `sql.Pool`.

**TP-LV-08: Multiple forbidden calls in one function**

- **Input:**
```typescript
app.post("/users", async (req, res) => {
  await db.Query("SELECT ...");
  await db.Exec("INSERT ...");
});
```
- **Expected:** Two violations (one per forbidden call).

**TP-LV-09: Forbidden call in a helper function within handler file**

- **Input** (`src/routes/users.ts`):
```typescript
function fetchFromDB(id: string) {
  return db.Query("SELECT * FROM users WHERE id = $1", id);
}
app.get("/users/:id", async (req, res) => {
  const user = await fetchFromDB(req.params.id);
  res.json(user);
});
```
- **Expected violation:** `db.Query` in handler file, even though it is in a helper function.

**TP-LV-10: Go repository using exec.Command (shell execution)**

- **Config:** `forbiddenCalls: ["exec.Command", "os.exec.*"]`
- **Input:**
```go
cmd := exec.Command("curl", "https://api.example.com")
```
- **Expected violation:** Repository executing shell commands.

### 15.2 True Negative Cases

**TN-LV-01: Handler calls service (proper delegation)**

- **Input** (`src/routes/users.ts`):
```typescript
const users = await userService.findAll();
```
- **Expected:** No violation.

**TN-LV-02: Service calls database (proper layer)**

- **Input** (`src/services/user-service.ts`):
```typescript
const users = await db.Query("SELECT * FROM users");
```
- **Expected:** No violation (service is not in handler layer with db-forbidden rules).

**TN-LV-03: File not in any layer**

- **Input** (`src/utils/db-helper.ts`) -- not matching any layer pattern.
- **Expected:** No violation.

**TN-LV-04: Call that looks similar but is not the forbidden one**

- **Input:** `const result = items.Query(filter);` where `items` is a custom collection, not a database.
- **Expected:** This is a false positive risk. If the rule does pattern matching on method names, this could be incorrectly flagged. The rule should ideally resolve the import/type of the receiver.

**TN-LV-05: Test file in handler directory**

- **Input** (`src/routes/users.test.ts`) calling `db.Query` in a test.
- **Expected:** No violation. Test files are exempt.

### 15.3 False Positive Risks

**FP-LV-01: Method name collision with forbidden pattern**

- `items.Query(filter)` where `Query` is a method on a custom class, not a database query.
- **Avoidance:** Resolve the receiver type. If not possible, require import analysis.

**FP-LV-02: Database call in handler for health check endpoint**

- `app.get("/health", async (req, res) => { await db.Query("SELECT 1"); });`
- **Avoidance:** Allow suppression per-route or per-file.

**FP-LV-03: Repository making internal HTTP call to another microservice**

- Some architectures have repositories that aggregate from multiple data sources including HTTP APIs.
- **Avoidance:** User can customize forbidden patterns per repository.

**FP-LV-04: ORM method that wraps a query**

- `await User.findAll()` -- is this a "database call"? It uses an ORM, not raw SQL.
- **Avoidance:** Forbidden patterns must be explicit. ORM calls are not `sql.*` or `db.*` unless configured.

**FP-LV-05: Go handler using sql.NullString (type, not query)**

- `var name sql.NullString` -- uses `sql.*` package for types, not queries.
- **Avoidance:** Distinguish between type usage and method calls.

### 15.4 False Negative Risks

**FN-LV-01: Database call through abstraction layer**

- Handler calls `repository.find()` which calls `db.Query()`. The handler does not directly call db, but the architectural intent is violated if the handler bypasses the service.
- **Severity:** Medium. This rule catches direct calls only.

**FN-LV-02: eval/Function constructor executing SQL**

- `eval("db.Query('SELECT 1')")` -- not detectable statically.
- **Severity:** Low.

**FN-LV-03: Dynamic method dispatch**

- `const method = "Query"; db[method]("SELECT 1");`
- **Severity:** Low.

**FN-LV-04: Database call in a middleware (not handler, not service)**

- Middleware runs in handler layer but may not match handler patterns.
- **Severity:** Medium. Middleware files should be included in handler layer patterns.

**FN-LV-05: Third-party library wrapping forbidden call**

- `import { dbHelper } from "some-lib"; dbHelper.query("...")` -- not matching `db.*` pattern.
- **Severity:** Medium.

### 15.5 Edge Cases

**EC-LV-01:** Forbidden call in a comment -- should not be flagged.
**EC-LV-02:** Forbidden call in a string literal -- should not be flagged.
**EC-LV-03:** Go method with receiver that matches forbidden pattern.
**EC-LV-04:** Arrow function vs regular function in handler file.
**EC-LV-05:** Class method in handler file calling forbidden method.
**EC-LV-06:** Forbidden call in a try/catch (still a violation).
**EC-LV-07:** Conditional forbidden call: `if (directDB) db.Query(...)`.
**EC-LV-08:** Forbidden call in type assertion context.
**EC-LV-09:** Glob pattern matching in forbiddenCalls.
**EC-LV-10:** Empty forbiddenCalls array for a layer -- rule is no-op for that layer.

### 15.6 Configuration Interaction

**CI-LV-01:** Custom layers with custom forbidden patterns.
**CI-LV-02:** Multiple layers, each with different forbidden patterns.
**CI-LV-03:** `reason` field included in violation message.
**CI-LV-04:** Overlapping layer patterns -- file matches multiple layers.

### 15.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `ARCH-layer-violation`.

---

## 16. ARCH-module-boundary

**Rule purpose:** Enforce that modules (directories with an index.ts or Go package) are accessed only through their public API.

### 16.1 True Positive Cases

**TP-MB-01: Reaching into module internals**

- **Directory structure:**
```
src/auth/
  index.ts        (exports validateToken)
  internal/
    token-utils.ts (exports helperFn)
```
- **Input** (`src/api/handler.ts`):
```typescript
import { helperFn } from "../auth/internal/token-utils";
```
- **Expected violation:** `ARCH-module-boundary`. Import reaches into `auth/internal/` instead of using `auth/index.ts`.

**TP-MB-02: Import specific file instead of index**

- **Directory structure:**
```
src/users/
  index.ts        (exports UserService)
  user-repo.ts    (exports UserRepo)
```
- **Input** (`src/api/routes.ts`):
```typescript
import { UserRepo } from "../users/user-repo";
```
- **Expected violation:** Should import from `../users` (resolves to index.ts), not directly from `user-repo.ts`.

**TP-MB-03: Deep import into module**

- **Input:**
```typescript
import { hashFn } from "../auth/crypto/internal/hash";
```
- **Expected violation:** Reaching past `auth/index.ts`.

**TP-MB-04: Go importing unexported-by-convention package internals**

- **Directory structure:**
```
pkg/auth/
  auth.go         (package auth, exports ValidateToken)
  internal/
    tokens.go     (package internal)
```
- **Input** (`cmd/main.go`):
```go
import "project/pkg/auth/internal"
```
- **Expected violation:** Go already enforces `internal/` at build time, but this rule catches similar patterns for non-`internal` directories.

**TP-MB-05: Importing a private helper file from another module**

- **Directory structure:**
```
src/payments/
  index.ts
  _helpers.ts     (underscore prefix = private by convention)
```
- **Input:**
```typescript
import { calcFee } from "../payments/_helpers";
```
- **Expected violation:** Importing private file (not exported through index).

**TP-MB-06: Named import of non-exported symbol**

- **Directory structure:** `src/auth/index.ts` exports `validateToken` but not `refreshToken`.
- **Input:**
```typescript
import { refreshToken } from "../auth/refresh"; // bypassing index
```
- **Expected violation:** `refreshToken` should be accessed through `auth/index.ts` if at all.

**TP-MB-07: Require() reaching into internals**

- **Input:**
```typescript
const utils = require("../auth/internal/token-utils");
```
- **Expected violation:** Same as import.

**TP-MB-08: Dynamic import reaching into internals**

- **Input:**
```typescript
const mod = await import("../auth/internal/cache");
```
- **Expected violation:** Dynamic import violates boundary.

**TP-MB-09: Re-export from internal file**

- **Input** (`src/api/index.ts`):
```typescript
export { helperFn } from "../auth/internal/token-utils";
```
- **Expected violation:** Re-export from another module's internals.

**TP-MB-10: Import from sibling directory within module (cross-module)**

- **Directory structure:**
```
src/auth/index.ts
src/users/index.ts
src/users/validators/email.ts
```
- **Input** (`src/auth/login.ts`):
```typescript
import { validateEmail } from "../users/validators/email";
```
- **Expected violation:** Should import from `../users` not `../users/validators/email`.

### 16.2 True Negative Cases

**TN-MB-01: Import from module index**

- **Input:**
```typescript
import { validateToken } from "../auth";
```
- **Expected:** No violation. Resolves to `auth/index.ts`.

**TN-MB-02: Import from module index with explicit path**

- **Input:**
```typescript
import { validateToken } from "../auth/index";
```
- **Expected:** No violation.

**TN-MB-03: Intra-module import**

- **Input** (`src/auth/login.ts`):
```typescript
import { hashPassword } from "./crypto";
```
- **Expected:** No violation. Same module (auth) internal imports are fine.

**TN-MB-04: Directory without index.ts**

- **Input:** `src/utils/` has no `index.ts`. Importing `../utils/helpers.ts` is fine.
- **Expected:** No violation. Module boundary only applies to directories with index files.

**TN-MB-05: Go same-package import**

- Files in the same Go package can import each other's exported symbols.
- **Expected:** No violation.

### 16.3 False Positive Risks

**FP-MB-01: CSS/asset import from module directory**

- `import "../auth/styles.css"` -- CSS is not a module boundary concern.
- **Avoidance:** Only apply to `.ts`/`.js`/`.go` imports.

**FP-MB-02: Test file importing internals for testing**

- Test files may need to import internal utilities for white-box testing.
- **Avoidance:** Exempt test files.

**FP-MB-03: Monorepo package without index.ts**

- Some packages use `package.json` exports field instead of index.ts.
- **Avoidance:** Consider `package.json` exports as module boundary definition.

**FP-MB-04: Generated barrel file that imports everything**

- Auto-generated index.ts that re-exports all files. Everything is "public."
- **Avoidance:** Accept this as the module's public API.

**FP-MB-05: Path alias that resolves to index**

- `import { x } from "@auth"` resolves to `src/auth/index.ts` -- should not trigger.

### 16.4 False Negative Risks

**FN-MB-01: Module without index.ts that should have one**

- Directory `src/auth/` has many files but no index. Consumers import individual files. No violation because no index exists.
- **Severity:** Medium. Consider a complementary rule: "directories with >N files should have an index."

**FN-MB-02: Star import that pulls in everything**

- `import * as auth from "../auth"` -- imports everything from index. If index re-exports internals, this is fine.
- **Severity:** Low.

**FN-MB-03: Package.json exports field not recognized**

- Module boundary defined by `exports` field in package.json, not by index.ts.
- **Severity:** Medium for monorepos.

**FN-MB-04: Go package that should use internal/ but does not**

- `pkg/auth/helpers.go` exports helpers that should be internal.
- **Severity:** Low. Go convention issue, not import issue.

**FN-MB-05: Symbolic link bypassing module boundary**

- Symlink from `src/api/auth-internals` -> `src/auth/internal/`. Import resolves through symlink.
- **Severity:** Low.

### 16.5 Edge Cases

**EC-MB-01:** Nested index files: `src/auth/index.ts` and `src/auth/oauth/index.ts`.
**EC-MB-02:** Default export vs named exports in index.
**EC-MB-03:** `export * from` in index file.
**EC-MB-04:** Circular index files (auth/index imports from users/index and vice versa).
**EC-MB-05:** Go `main` package (no export concept).
**EC-MB-06:** TypeScript project references.
**EC-MB-07:** Path with `..` that resolves to same module.
**EC-MB-08:** File in module root (not in subdirectory) -- is it accessible directly?
**EC-MB-09:** Multiple index files (index.ts and index.js) in same directory.
**EC-MB-10:** Module with only type exports (all `export type`).

### 16.6 Configuration Interaction

Rule has no options in the spec beyond global enable/disable. Test:
**CI-MB-01:** Rule enabled -- all module boundary violations reported.
**CI-MB-02:** Rule disabled -- no violations.
**CI-MB-03:** Rule as "warn" -- warnings only.

### 16.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `ARCH-module-boundary`.

---

## 17. CONV-file-naming

**Rule purpose:** Enforce consistent file naming conventions.

### 17.1 True Positive Cases

**TP-FN-01: PascalCase file when kebab-case required**

- **Config:** `style: kebab-case`
- **Input file path:** `src/UserService.ts`
- **Expected violation:** `CONV-file-naming`. File "UserService.ts" should be "user-service.ts" (kebab-case).

**TP-FN-02: camelCase file when kebab-case required**

- **Input:** `src/userService.ts`
- **Expected violation:** Should be `user-service.ts`.

**TP-FN-03: snake_case file when kebab-case required**

- **Input:** `src/user_service.ts`
- **Expected violation:** Should be `user-service.ts`.

**TP-FN-04: File with spaces**

- **Input:** `src/user service.ts`
- **Expected violation:** No naming convention allows spaces.

**TP-FN-05: Go file with kebab-case when snake_case required**

- **Config:** `style: snake_case`
- **Input:** `internal/user-service.go`
- **Expected violation:** Should be `user_service.go`.

**TP-FN-06: PascalCase file not in override pattern**

- **Config:**
```yaml
style: kebab-case
overrides:
  "src/components/**": PascalCase
```
- **Input:** `src/services/UserService.ts` (not in components/).
- **Expected violation:** Not in PascalCase override zone; must be kebab-case.

**TP-FN-07: Mixed case in Go file**

- **Input:** `internal/UserRepo.go`
- **Expected violation:** Go files should be `snake_case` or `kebab-case` by convention.

**TP-FN-08: Uppercase letters in kebab-case file**

- **Input:** `src/User-Service.ts`
- **Expected violation:** Kebab-case does not allow uppercase.

**TP-FN-09: Double hyphen in kebab-case**

- **Input:** `src/user--service.ts`
- **Expected violation:** Double hyphens are not valid kebab-case.

**TP-FN-10: File starting with number (if not allowed)**

- **Input:** `src/1-user.ts`
- **Expected:** Depends on convention. Kebab-case allows leading numbers or not (spec ambiguity). Flag if convention requires alpha start.

### 17.2 True Negative Cases

**TN-FN-01: Correct kebab-case**

- **Input:** `src/user-service.ts`
- **Expected:** No violation.

**TN-FN-02: Correct PascalCase in override zone**

- **Config:** `overrides: { "src/components/**": PascalCase }`
- **Input:** `src/components/UserCard.tsx`
- **Expected:** No violation.

**TN-FN-03: Go snake_case file**

- **Config:** `style: snake_case`
- **Input:** `internal/user_service.go`
- **Expected:** No violation.

**TN-FN-04: Test file following convention**

- **Config:** `overrides: { "**/*.test.*": kebab-case }`
- **Input:** `tests/user-service.test.ts`
- **Expected:** No violation.

**TN-FN-05: Index file**

- **Input:** `src/auth/index.ts`
- **Expected:** No violation. "index" is valid kebab-case.

### 17.3 False Positive Risks

**FP-FN-01: Third-party convention files**

- `Dockerfile`, `Makefile`, `README.md` -- these have their own naming conventions.
- **Avoidance:** Exclude well-known filenames.

**FP-FN-02: Config files**

- `.stricture.yml`, `jest.config.ts`, `.eslintrc.js` -- naming dictated by tools.
- **Avoidance:** Exclude dotfiles and known config file patterns.

**FP-FN-03: Go test file suffix**

- `user_test.go` must end in `_test.go` (Go convention). If style is kebab-case, this would be `user-test.go` which breaks Go.
- **Avoidance:** Respect language-specific constraints. `_test.go` suffix is mandatory in Go.

**FP-FN-04: TypeScript declaration files**

- `global.d.ts`, `env.d.ts` -- the `.d.ts` extension has its own convention.
- **Avoidance:** Excluded by default per spec's ignore patterns.

**FP-FN-05: File with version number**

- `migration-v2.ts` -- kebab-case with version number is valid.

### 17.4 False Negative Risks

**FN-FN-01: Correct convention but misleading name**

- `src/a.ts` -- valid kebab-case but not descriptive. Not this rule's concern.
- **Severity:** Low.

**FN-FN-02: Extension mismatch**

- `user-service.tsx` in a directory where all other files are `.ts`. Not a naming issue.
- **Severity:** Low.

**FN-FN-03: Directory naming not checked**

- `src/UserModule/user-service.ts` -- file is kebab-case but directory is PascalCase.
- **Severity:** Medium. Spec does not mention directory naming.

**FN-FN-04: Hidden files (dotfiles)**

- `.env`, `.gitignore` -- not checked by convention rules.
- **Severity:** Low.

**FN-FN-05: Symlinked file with different name**

- Symlink `UserService.ts` -> `user-service.ts`. Both names exist.
- **Severity:** Low.

### 17.5 Edge Cases

**EC-FN-01:** File with no extension.
**EC-FN-02:** File with multiple extensions: `schema.generated.ts`.
**EC-FN-03:** File name is a single character: `a.ts`.
**EC-FN-04:** File with Unicode characters: `benutzer-dienst.ts`.
**EC-FN-05:** File with numbers: `v2-migration.ts`.
**EC-FN-06:** File with leading dot: `.env.local`.
**EC-FN-07:** File with trailing dot: `file..ts`.
**EC-FN-08:** Case-insensitive file system (macOS) -- `UserService.ts` and `userservice.ts` might be the same file.
**EC-FN-09:** Go file with `_` prefix: `_cgo_export.go`.
**EC-FN-10:** File in root directory vs nested directory.

### 17.6 Configuration Interaction

**CI-FN-01:** `style: kebab-case` -- all files must be kebab-case.
**CI-FN-02:** `style: camelCase` -- all files must be camelCase.
**CI-FN-03:** `style: PascalCase` -- all files must be PascalCase.
**CI-FN-04:** `style: snake_case` -- all files must be snake_case.
**CI-FN-05:** Overrides take precedence over global style.
**CI-FN-06:** Multiple overrides for different directories.
**CI-FN-07:** Invalid style value -- exit code 2.

### 17.7 Inline Suppression Testing

File-level rule. Suppression comment at top of file should suppress. `stricture-disable-next-line` does not make sense for file-level violations.

### 17.8 Auto-Fix Testing

- **Fix:** Renames file from `UserService.ts` to `user-service.ts`.
- **Verify:** File is renamed. All imports referencing old name are updated.
- **Verify:** `--fix-dry-run` shows rename without executing.
- **Verify:** Fix is idempotent (already-correct files are not modified).
- **Verify:** Fix does not rename files excluded by ignore patterns.

---

## 18. CONV-file-header

**Rule purpose:** Require a header comment at the top of every file.

### 18.1 True Positive Cases

**TP-FH-01: Missing header entirely**

- **Config:** `pattern: "// {filename} -- {description}"`
- **Input** (`src/utils/helpers.ts`):
```typescript
export function add(a: number, b: number): number { return a + b; }
```
- **Expected violation:** `CONV-file-header`. Missing file header comment.

**TP-FH-02: Header with wrong filename**

- **Input** (`src/utils/helpers.ts`):
```typescript
// utils.ts -- Utility functions
export function add(a: number, b: number): number { return a + b; }
```
- **Expected violation:** Header says `utils.ts` but file is `helpers.ts`.

**TP-FH-03: Header without description**

- **Input** (`src/utils/helpers.ts`):
```typescript
// helpers.ts --
export function add(a: number, b: number): number { return a + b; }
```
- **Expected violation:** `{description}` is empty.

**TP-FH-04: Header with wrong format**

- **Config:** `pattern: "// {filename} -- {description}"`
- **Input:**
```typescript
/* helpers.ts - Utility functions */
```
- **Expected violation:** Uses `/* */` instead of `//`, and single `-` instead of `--`.

**TP-FH-05: Go file missing header**

- **Config:** `patterns: { "**/*.go": "// {filename} -- {description}" }`
- **Input** (`internal/service/user.go`):
```go
package service

func CreateUser() {}
```
- **Expected violation:** Missing header (package declaration is not a header comment).

**TP-FH-06: Header on wrong line (not first line)**

- **Input:**
```typescript

// helpers.ts -- Utility functions
export function add() {}
```
- **Expected violation:** Header must be on line 1. Blank line before header is a violation.

**TP-FH-07: Header missing the dash separator**

- **Input:**
```typescript
// helpers.ts Utility functions
```
- **Expected violation:** Missing ` -- ` separator.

**TP-FH-08: Only a shebang line, no header**

- **Input:**
```typescript
#!/usr/bin/env node
export function main() {}
```
- **Expected violation:** Shebang is not a file header comment. (Or: header must be on line 2 after shebang.)

**TP-FH-09: Header with filename but placeholder description**

- **Input:**
```typescript
// helpers.ts -- TODO: add description
```
- **Expected:** Depends on config. `{description}` must be a non-empty string. "TODO: add description" is non-empty, so it technically passes. But some teams may want to reject TODO placeholders.

**TP-FH-10: Multiple files missing headers**

- **Input:** Three files without headers.
- **Expected:** Three separate violations.

### 18.2 True Negative Cases

**TN-FH-01: Correct header**

- **Input:**
```typescript
// helpers.ts -- Utility functions for string and date manipulation.
export function add() {}
```
- **Expected:** No violation.

**TN-FH-02: Go file with correct header**

- **Input:**
```go
// user.go -- User domain model and creation logic.
package user
```
- **Expected:** No violation.

**TN-FH-03: File in ignore list**

- **Config:** `ignore: ["**/*.d.ts", "**/index.ts"]`
- **Input:** `src/auth/index.ts` without header.
- **Expected:** No violation (file is ignored).

**TN-FH-04: Header with long description**

- **Input:**
```typescript
// helpers.ts -- A comprehensive collection of utility functions used across the application for string manipulation, date formatting, and common type conversions.
```
- **Expected:** No violation.

**TN-FH-05: Header after shebang (if allowed)**

- **Input:**
```typescript
#!/usr/bin/env node
// cli.ts -- Command line interface entry point.
```
- **Expected:** No violation if shebang+header pattern is accepted.

### 18.3 False Positive Risks

**FP-FH-01: Generated files with auto-generated headers**

- Generated files may have their own header format (`// Code generated by protoc. DO NOT EDIT.`).
- **Avoidance:** Ignore generated files via config.

**FP-FH-02: License header at top of file**

- Some projects require license headers first, then file headers.
- **Avoidance:** Allow configurable line offset for header location.

**FP-FH-03: JSDoc comment at top**

- `/** @module helpers */` -- is a documentation comment, not a file header in the required format.
- **Avoidance:** Pattern must match exactly.

**FP-FH-04: Go build constraint before header**

- `//go:build linux` must be the first line. Header would be second.
- **Avoidance:** Recognize Go build constraints and check for header after them.

**FP-FH-05: BOM marker before header**

- UTF-8 BOM at byte 0 makes the header not truly "first line."
- **Avoidance:** Strip BOM before checking.

### 18.4 False Negative Risks

**FN-FH-01: Header present but description is meaningless**

- `// helpers.ts -- asdf`
- **Severity:** Low. Tool cannot judge description quality.

**FN-FH-02: Header with wrong dash style**

- `// helpers.ts - Utility functions` (single dash instead of `--`).
- **Severity:** Medium. Pattern matching should be strict.

**FN-FH-03: Header in wrong comment style**

- Go file with `/* filename -- desc */` instead of `// filename -- desc`.
- **Severity:** Medium.

**FN-FH-04: Header after imports**

- Header exists but after import statements. Tool may not catch this if it only checks for presence, not position.
- **Severity:** Medium.

**FN-FH-05: Header with typo in filename**

- `// helperss.ts -- Utilities` (extra 's').
- **Severity:** Medium. Should compare against actual filename.

### 18.5 Edge Cases

**EC-FH-01:** Empty file -- should it require a header? (Probably not if it has no code.)
**EC-FH-02:** File with only a header comment and nothing else.
**EC-FH-03:** TypeScript file starting with `"use strict";` before header.
**EC-FH-04:** Go file with package comment (godoc) before package declaration.
**EC-FH-05:** File with BOM marker.
**EC-FH-06:** Windows line endings (CRLF) in header.
**EC-FH-07:** Header with Unicode characters in description.
**EC-FH-08:** Filename with special characters (spaces, dots).
**EC-FH-09:** Different pattern per file extension.
**EC-FH-10:** Nested directory affects filename: does `{filename}` include path?

### 18.6 Configuration Interaction

**CI-FH-01:** Single pattern for all files: `pattern: "// {filename} -- {description}"`.
**CI-FH-02:** Different patterns per glob: `"**/*.go": "// {filename} -- {description}"`.
**CI-FH-03:** `ignore` patterns exclude specific files/directories.
**CI-FH-04:** Pattern without `{filename}` placeholder -- should this be valid?
**CI-FH-05:** Pattern without `{description}` placeholder -- header is just filename.

### 18.7 Inline Suppression Testing

File-level rule. Suppression at top of file: `// stricture-disable CONV-file-header`.

### 18.8 Auto-Fix Testing

- **Fix:** Adds header comment as first line.
- **Verify:** Correct filename inserted.
- **Verify:** Description placeholder or empty string for `{description}`.
- **Verify:** Existing content is not displaced (header added before existing code).
- **Verify:** File encoding preserved.
- **Verify:** Fix is idempotent (file with correct header is not modified).
- **Verify:** Shebang/build constraint lines are preserved above header if applicable.

---

## 19. CONV-error-format

**Rule purpose:** Enforce a consistent error message format across the codebase.

### 19.1 True Positive Cases

**TP-EF-01: Error without structured format**

- **Config:** `pattern: "{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}"`, `minSegments: 2`
- **Input:**
```typescript
throw new Error("something went wrong");
```
- **Expected violation:** `CONV-error-format`. Error message does not follow format. Missing operation and structured cause.

**TP-EF-02: Error with only one segment**

- **Input:**
```typescript
throw new Error("failed");
```
- **Expected violation:** "failed" has only 1 segment, minimum is 2.

**TP-EF-03: Go fmt.Errorf without format**

- **Input:**
```go
return fmt.Errorf("error creating user")
```
- **Expected violation:** Missing structured format (no colon separator for operation).

**TP-EF-04: Lowercase error message (no operation prefix)**

- **Input:**
```typescript
throw new Error("email already exists");
```
- **Expected violation:** Missing `{OPERATION}:` prefix.

**TP-EF-05: Error with operation but no recovery action**

- **Config:** `minSegments: 3` (all three parts required)
- **Input:**
```typescript
throw new Error("CreateUser: email already exists");
```
- **Expected violation:** Missing recovery action (third segment).

**TP-EF-06: Go error with wrong format**

- **Input:**
```go
return fmt.Errorf("bad request: %v", err)
```
- **Expected violation:** "bad request" is not an operation name (not PascalCase/camelCase function name).

**TP-EF-07: Custom error class with bad message**

- **Input:**
```typescript
throw new ValidationError("oops");
```
- **Expected violation:** "oops" does not follow format.

**TP-EF-08: Error in callback**

- **Input:**
```typescript
callback(new Error("it broke"));
```
- **Expected violation:** Error message does not follow format.

**TP-EF-09: Multiple errors in one function, some invalid**

- **Input:**
```typescript
if (!name) throw new Error("missing name");
if (!email) throw new Error("CreateUser: email required. Provide a valid email.");
```
- **Expected violation:** First error is invalid. Second is valid.

**TP-EF-10: Go errors.New with bad format**

- **Input:**
```go
var ErrBad = errors.New("bad thing happened")
```
- **Expected violation:** Sentinel error without format.

### 19.2 True Negative Cases

**TN-EF-01: Correctly formatted error**

- **Input:**
```typescript
throw new Error("CreateUser: email already exists. Use a different email address.");
```
- **Expected:** No violation.

**TN-EF-02: Go correctly formatted error**

- **Input:**
```go
return fmt.Errorf("FetchUser: user not found. Verify the user ID is correct.")
```
- **Expected:** No violation.

**TN-EF-03: Two-segment error with minSegments=2**

- **Config:** `minSegments: 2`
- **Input:**
```typescript
throw new Error("CreateUser: email already exists.");
```
- **Expected:** No violation. Has operation + root cause.

**TN-EF-04: Error wrapping (Go)**

- **Input:**
```go
return fmt.Errorf("CreateUser: %w", err)
```
- **Expected:** No violation. Error wrapping with `%w` is acceptable (operation prefix present).

**TN-EF-05: Error not in applyTo list**

- **Config:** `applyTo: ["new Error", "throw new .*Error"]`
- **Input:** `console.error("bad thing");` -- console.error is not in applyTo.
- **Expected:** No violation.

### 19.3 False Positive Risks

**FP-EF-01: Third-party library error messages**

- `throw new TypeError("Cannot read property 'x' of undefined")` -- built-in error.
- **Avoidance:** Only apply to user-created errors, not built-in TypeError/RangeError.

**FP-EF-02: Error messages constructed dynamically**

- `throw new Error(\`${operation}: ${reason}\`)` -- template literal follows format but tool cannot statically verify.
- **Avoidance:** If template literal contains `:` separator, consider it potentially valid.

**FP-EF-03: Error in logging, not thrown**

- `logger.error("something went wrong")` -- not an error being thrown.
- **Avoidance:** Only check patterns in `applyTo` list.

**FP-EF-04: Test file error assertions**

- `expect(() => fn()).toThrow("CreateUser: email exists")` -- the string is in a test assertion, not error creation.
- **Avoidance:** Only check error creation patterns, not assertion strings.

**FP-EF-05: Error message in a non-error context**

- `const msg = "CreateUser: email exists.";` -- just a string, not thrown.
- **Avoidance:** Only check strings passed to error constructors or fmt.Errorf.

### 19.4 False Negative Risks

**FN-EF-01: Error message in a variable**

- `const msg = "bad thing"; throw new Error(msg);`
- **Severity:** Medium. Tool must follow variable to resolve message content.

**FN-EF-02: Error created by factory function**

- `throw createError("bad thing");` where `createError` calls `new Error(...)`.
- **Severity:** Medium. Requires cross-function analysis.

**FN-EF-03: Go error created by helper**

- `return wrapError("bad", err)` where `wrapError` calls `fmt.Errorf`.
- **Severity:** Medium.

**FN-EF-04: Error message that follows format but is misleading**

- `throw new Error("CreateUser: success. Everything is fine.")` -- follows format but is not an error.
- **Severity:** Low. Tool cannot judge semantic accuracy.

**FN-EF-05: Error in switch/case default**

- `default: throw new Error("unexpected");`
- **Severity:** Medium.

### 19.5 Edge Cases

**EC-EF-01:** Error message with newlines: `throw new Error("Op: cause.\nRecovery: do this.");`
**EC-EF-02:** Error message with special characters: colons in URLs.
**EC-EF-03:** Internationalized error messages.
**EC-EF-04:** Error message that is an empty string.
**EC-EF-05:** Go sentinel errors: `var ErrNotFound = errors.New("...")` -- should these follow format?
**EC-EF-06:** Multiple error constructor patterns in same file.
**EC-EF-07:** Error message spanning multiple lines (template literal).
**EC-EF-08:** Error with JSON-encoded message.
**EC-EF-09:** Assert/panic in Go.
**EC-EF-10:** Re-thrown error (same message, different stack).

### 19.6 Configuration Interaction

**CI-EF-01:** `pattern` defines the format template.
**CI-EF-02:** `applyTo` lists which function calls to check.
**CI-EF-03:** `minSegments: 2` -- at least operation + cause.
**CI-EF-04:** `minSegments: 3` -- all three parts required.
**CI-EF-05:** Custom applyTo: `["createAppError", "AppError"]`.

### 19.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `CONV-error-format`.

---

## 20. CONV-export-naming

**Rule purpose:** Enforce naming conventions for exported/public symbols.

### 20.1 True Positive Cases

**TP-EN-01: Exported function not in camelCase**

- **Config:** `typescript: { exportedFunctions: camelCase }`
- **Input:**
```typescript
export function CreateUser() {}
```
- **Expected violation:** `CONV-export-naming`. Exported function "CreateUser" should be camelCase ("createUser").

**TP-EN-02: Exported constant not UPPER_SNAKE_CASE**

- **Config:** `typescript: { exportedConstants: UPPER_SNAKE_CASE }`
- **Input:**
```typescript
export const maxRetries = 3;
```
- **Expected violation:** Should be `MAX_RETRIES`.

**TP-EN-03: Exported class not PascalCase**

- **Config:** `typescript: { exportedClasses: PascalCase }`
- **Input:**
```typescript
export class userService {}
```
- **Expected violation:** Should be `UserService`.

**TP-EN-04: Exported type not PascalCase**

- **Config:** `typescript: { exportedTypes: PascalCase }`
- **Input:**
```typescript
export type createUserInput = { name: string; };
```
- **Expected violation:** Should be `CreateUserInput`.

**TP-EN-05: Go exported function not PascalCase**

- **Config:** `go: { exportedFunctions: PascalCase }`
- **Input:**
```go
func create_user() {} // lowercase = unexported in Go, so this would not be exported
// More realistic:
func CreateUSER() {} // acronym not properly cased
```
- **Expected:** Go's export mechanism (uppercase first letter) means all exported functions start with uppercase. But `CreateUSER` might violate PascalCase (should be `CreateUser`).

**TP-EN-06: Exported enum not PascalCase**

- **Input:**
```typescript
export enum userRole { Admin = "admin", User = "user" }
```
- **Expected violation:** `userRole` should be `UserRole`.

**TP-EN-07: Default export with wrong naming**

- **Input:**
```typescript
const myHandler = () => {};
export default myHandler;
```
- **Expected:** Depends on config. If exported functions must be camelCase, `myHandler` is fine. If default exports have special rules, test that.

**TP-EN-08: Named export in barrel file with wrong casing**

- **Input** (`src/index.ts`):
```typescript
export { Create_User } from "./user";
```
- **Expected violation:** `Create_User` is neither camelCase nor PascalCase.

**TP-EN-09: Go exported type not PascalCase**

- **Input:**
```go
type USER_CONFIG struct {} // UPPER_SNAKE is not PascalCase
```
- **Expected violation:** Should be `UserConfig`.

**TP-EN-10: Exported interface not PascalCase**

- **Input:**
```typescript
export interface createUserInput { name: string; }
```
- **Expected violation:** Should be `CreateUserInput`.

### 20.2 True Negative Cases

**TN-EN-01: Correct camelCase function**

- **Input:** `export function createUser() {}`
- **Expected:** No violation.

**TN-EN-02: Correct UPPER_SNAKE constant**

- **Input:** `export const MAX_RETRIES = 3;`
- **Expected:** No violation.

**TN-EN-03: Correct PascalCase class**

- **Input:** `export class UserService {}`
- **Expected:** No violation.

**TN-EN-04: Non-exported symbol (private)**

- **Input:** `function _helper() {}` -- not exported.
- **Expected:** No violation. Rule only applies to exports.

**TN-EN-05: Go correct PascalCase**

- **Input:** `func CreateUser() {}`, `type UserConfig struct {}`
- **Expected:** No violation.

### 20.3 False Positive Risks

**FP-EN-01: Acronyms in names**

- `export function parseJSON()` -- is "JSON" correct PascalCase/camelCase? Some conventions say "Json."
- **Avoidance:** Config for acronym handling.

**FP-EN-02: Re-export of third-party symbol**

- `export { XMLParser } from "fast-xml-parser"` -- third-party naming.
- **Avoidance:** Only check symbols defined in the project, not re-exports of external.

**FP-EN-03: Test utility exports**

- `export function createMockUser()` -- test utility naming may differ.
- **Avoidance:** Exempt test files.

**FP-EN-04: Constants that are not truly constant**

- `export const userCache = new Map()` -- technically a const but not UPPER_SNAKE-worthy.
- **Avoidance:** Only apply UPPER_SNAKE to primitive constants.

**FP-EN-05: Single-letter exports**

- `export const x = 1;` -- valid UPPER_SNAKE is `X`. Some teams allow single-letter.
- **Avoidance:** Option for minimum name length.

### 20.4 False Negative Risks

**FN-EN-01: Export via module.exports**

- `module.exports = { Create_User: fn }` -- CommonJS export not caught by ES export analysis.
- **Severity:** Medium.

**FN-EN-02: Dynamic export name**

- `export const [name] = [value]` using computed property.
- **Severity:** Low.

**FN-EN-03: Go method names (not functions)**

- `func (s *Service) create_handler()` -- method name not PascalCase but might be unexported.
- **Severity:** Low.

**FN-EN-04: Default export not checked**

- Some configs may forget to cover default exports.
- **Severity:** Medium.

**FN-EN-05: Re-export renaming**

- `export { foo as Bar } from "./mod"` -- `Bar` is the exported name.
- **Severity:** Medium. Check the renamed export.

### 20.5 Edge Cases

**EC-EN-01:** Export with underscore prefix: `export const _internal = ...`.
**EC-EN-02:** Export with dollar sign: `export const $state = ...`.
**EC-EN-03:** Numeric constant: `export const MAX_UINT_32 = 4294967295`.
**EC-EN-04:** TypeScript `export =` syntax.
**EC-EN-05:** Go init function (special, not exported).
**EC-EN-06:** Export star: `export * from "./module"` -- names come from source.
**EC-EN-07:** Enum members: are they checked individually?
**EC-EN-08:** Namespace exports: `export namespace Utils {}`.
**EC-EN-09:** Mixed export styles in one file.
**EC-EN-10:** Overloaded function names.

### 20.6 Configuration Interaction

**CI-EN-01:** Different conventions per symbol type (function, class, constant, type).
**CI-EN-02:** Different conventions for Go vs TypeScript.
**CI-EN-03:** Invalid convention name -- exit code 2.

### 20.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `CONV-export-naming`.

### 20.8 Auto-Fix Testing

- **Fix:** Renames exported symbol to match convention.
- **Verify:** All references (imports in other files) are also updated.
- **Verify:** Fix does not rename non-exported symbols.
- **Verify:** Fix handles re-exports correctly.

---

## 21. CONV-test-file-location

**Rule purpose:** Enforce where test files live relative to source files.

### 21.1 True Positive Cases

**TP-TFL-01: Test file in wrong location (colocated strategy)**

- **Config:** `strategy: colocated`
- **Source:** `src/services/user.ts`
- **Test:** `tests/services/user.test.ts` (in separate tests/ directory)
- **Expected violation:** `CONV-test-file-location`. Test should be at `src/services/user.test.ts`.

**TP-TFL-02: Test file in wrong location (mirrored strategy)**

- **Config:** `strategy: mirrored`
- **Source:** `src/services/user.ts`
- **Test:** `src/services/user.test.ts` (colocated, not in tests/ mirror)
- **Expected violation:** Test should be at `tests/services/user.test.ts`.

**TP-TFL-03: Test file in wrong location (subfolder strategy)**

- **Config:** `strategy: subfolder`
- **Source:** `src/services/user.ts`
- **Test:** `src/services/user.test.ts` (not in __tests__/)
- **Expected violation:** Test should be at `src/services/__tests__/user.test.ts`.

**TP-TFL-04: Go test file not colocated**

- **Config:** `strategy: colocated`
- **Source:** `internal/service/user.go`
- **Test:** `test/service/user_test.go`
- **Expected violation:** Go tests must be colocated (same package directory).

**TP-TFL-05: Wrong suffix**

- **Config:** `suffixes: { typescript: [".test.ts", ".spec.ts"] }`
- **Input:** `src/services/user.tests.ts` (wrong suffix: `.tests.ts` instead of `.test.ts`)
- **Expected violation:** File suffix does not match.

**TP-TFL-06: Test file at wrong depth (mirrored)**

- **Source:** `src/deep/nested/module.ts`
- **Test:** `tests/module.test.ts` (should be `tests/deep/nested/module.test.ts`)
- **Expected violation:** Test file does not mirror source file's directory structure.

**TP-TFL-07: Test file for non-existent source file**

- **Test:** `src/services/ghost.test.ts` with no corresponding `ghost.ts`.
- **Expected:** Warning (orphan test file), not necessarily an error.

**TP-TFL-08: Multiple test suffixes, none matching**

- **Config:** `suffixes: { typescript: [".test.ts"] }`
- **Input:** `src/services/user.spec.ts`
- **Expected violation:** `.spec.ts` not in allowed suffixes.

**TP-TFL-09: Test file in root instead of nested (colocated)**

- **Source:** `src/services/user.ts`
- **Test:** `user.test.ts` (in project root)
- **Expected violation:** Should be adjacent to source.

**TP-TFL-10: Subfolder test at wrong level**

- **Config:** `strategy: subfolder`
- **Source:** `src/services/user.ts`
- **Test:** `src/__tests__/services/user.test.ts` (subfolder at wrong level)
- **Expected violation:** Should be `src/services/__tests__/user.test.ts`.

### 21.2 True Negative Cases

**TN-TFL-01:** Colocated strategy, test adjacent to source.
**TN-TFL-02:** Mirrored strategy, test in correct mirror location.
**TN-TFL-03:** Subfolder strategy, test in __tests__ subfolder.
**TN-TFL-04:** Go test file in same package directory.
**TN-TFL-05:** Correct suffix used.

### 21.3 False Positive Risks

**FP-TFL-01:** Integration test files that intentionally live in a separate directory.
**FP-TFL-02:** E2E test files not associated with specific source files.
**FP-TFL-03:** Shared test utilities (not tests themselves).
**FP-TFL-04:** Monorepo where packages have different strategies.
**FP-TFL-05:** Generated test files in a build output directory.

### 21.4 False Negative Risks

**FN-TFL-01:** Test file at correct location but testing the wrong source file.
**FN-TFL-02:** Test file correctly located but empty.
**FN-TFL-03:** Source file moved but test file not moved (now orphaned).
**FN-TFL-04:** Alias/symlink that resolves to correct location.
**FN-TFL-05:** Test file for an index.ts (does index.test.ts map to the right module?).

### 21.5 Edge Cases

**EC-TFL-01:** Source file with multiple test files.
**EC-TFL-02:** Test file testing multiple source files.
**EC-TFL-03:** Go `_test.go` suffix is mandatory -- cannot be changed.
**EC-TFL-04:** TypeScript `.spec.ts` vs `.test.ts` -- both should be configurable.
**EC-TFL-05:** Nested __tests__ directories.
**EC-TFL-06:** Source file in root directory.
**EC-TFL-07:** Test for a package index file.
**EC-TFL-08:** Monorepo with different test locations per package.
**EC-TFL-09:** File with both source and test code (single-file tests).
**EC-TFL-10:** Fixture files alongside tests.

### 21.6 Configuration Interaction

**CI-TFL-01:** `strategy: colocated` -- adjacent test files.
**CI-TFL-02:** `strategy: mirrored` -- mirror directory structure.
**CI-TFL-03:** `strategy: subfolder` -- __tests__ subdirectory.
**CI-TFL-04:** Custom suffixes per language.
**CI-TFL-05:** Invalid strategy value -- exit code 2.

### 21.7 Auto-Fix Testing

- **Fix:** Moves test file to correct location.
- **Verify:** Creates directories as needed.
- **Verify:** Updates imports in the moved test file.
- **Verify:** Does not overwrite existing file at target location.

---

## 22. CONV-required-exports

**Rule purpose:** Ensure that certain files or directories have required exports.

### 22.1 True Positive Cases

**TP-RE-01: Feature module missing default export**

- **Config:**
```yaml
patterns:
  "src/features/*/index.ts":
    required: ["default"]
```
- **Input** (`src/features/auth/index.ts`):
```typescript
export const authService = new AuthService();
// Missing: no default export
```
- **Expected violation:** `CONV-required-exports`. File missing required export "default".

**TP-RE-02: Service module missing factory function**

- **Config:**
```yaml
patterns:
  "src/services/*/index.ts":
    required: ["create*Service"]
```
- **Input** (`src/services/user/index.ts`):
```typescript
export class UserService {}
// Missing: no createUserService or create*Service export
```
- **Expected violation:** Missing required export matching `create*Service`.

**TP-RE-03: Multiple required exports, one missing**

- **Config:**
```yaml
patterns:
  "src/modules/*/index.ts":
    required: ["default", "config"]
```
- **Input** (`src/modules/auth/index.ts`):
```typescript
export default class Auth {};
// Missing: no "config" export
```
- **Expected violation:** Missing "config" export.

**TP-RE-04: File matches pattern but has no exports at all**

- **Input** (`src/features/payments/index.ts`):
```typescript
// empty file or internal only
const x = 1;
```
- **Expected violation:** No exports at all, "default" required.

**TP-RE-05: Wildcard pattern, no match**

- **Config:** `required: ["use*Hook"]`
- **Input:** File exports `useAuth` and `useUser` -- but if the pattern is `use*Hook` (requiring "Hook" suffix), neither matches.
- **Expected violation:** Missing export matching `use*Hook`.

**TP-RE-06: Go package missing required exported function**

- **Config:**
```yaml
patterns:
  "internal/*/":
    required: ["New*"]
```
- **Input:** Package `internal/cache/` exports `Get` and `Set` but no `NewCache` or `New*`.
- **Expected violation:** Missing `New*` export.

**TP-RE-07: File exists but wrong export name**

- **Config:** `required: ["createAuthService"]`
- **Input:** File exports `makeAuthService` instead.
- **Expected violation:** Missing `createAuthService`.

**TP-RE-08: Required export exists but is not exported (private)**

- **Input:**
```typescript
function createService() {} // not exported
export const other = 1;
```
- **Expected violation:** `createService` exists but is not exported.

**TP-RE-09: Required export as type-only**

- **Config:** `required: ["UserConfig"]`
- **Input:**
```typescript
export type UserConfig = { /* ... */ };
```
- **Expected:** Depends on whether type exports count. If the config requires a value export, this is a violation.

**TP-RE-10: Pattern matches multiple files, some missing export**

- **Config:** `"src/features/*/index.ts": { required: ["default"] }`
- Two features: `auth/index.ts` has default export, `payments/index.ts` does not.
- **Expected violation:** Only `payments/index.ts` is flagged.

### 22.2 True Negative Cases

**TN-RE-01:** File has all required exports.
**TN-RE-02:** File does not match any required-exports pattern.
**TN-RE-03:** Wildcard pattern matched by exported symbol.
**TN-RE-04:** Go package has required New* constructor.
**TN-RE-05:** Re-export counts as an export.

### 22.3 False Positive Risks

**FP-RE-01:** Module that intentionally has no default export (utility module).
**FP-RE-02:** Pattern matches a generated file that should not have custom exports.
**FP-RE-03:** Feature module in development (not yet complete).
**FP-RE-04:** Module with only type exports when value exports are required.
**FP-RE-05:** Barrel file that uses `export *` -- individual named exports may not be directly visible.

### 22.4 False Negative Risks

**FN-RE-01:** Export exists but with wrong type (function vs class).
**FN-RE-02:** Export is a no-op or placeholder.
**FN-RE-03:** Export name matches pattern but is deprecated.
**FN-RE-04:** Dynamic export: `module.exports[name] = fn`.
**FN-RE-05:** Conditional export: `if (env) export const x = ...`.

### 22.5 Edge Cases

**EC-RE-01:** Pattern with multiple wildcards.
**EC-RE-02:** File matching multiple patterns with different requirements.
**EC-RE-03:** Pattern matching zero files -- no violation (no files to check).
**EC-RE-04:** Go package with multiple files contributing exports.
**EC-RE-05:** TypeScript export assignments: `export = something`.
**EC-RE-06:** Default export that is anonymous: `export default function() {}`.
**EC-RE-07:** Re-export: `export { default } from "./other"`.
**EC-RE-08:** Star export: `export * from "./module"` -- does this satisfy named export requirements?
**EC-RE-09:** Conditional type exports.
**EC-RE-10:** Circular dependency between required export files.

### 22.6 Configuration Interaction

**CI-RE-01:** Pattern glob matching.
**CI-RE-02:** Required exports as exact names vs wildcard patterns.
**CI-RE-03:** Multiple pattern entries.
**CI-RE-04:** Empty required array -- effectively no check.

### 22.7 Inline Suppression Testing

File-level rule. Suppression at top of file.

---

## 23. CTR-request-shape

**Rule purpose:** Verify that the request body type the client sends matches what the server expects.

### 23.1 True Positive Cases

**TP-RS-01: Client sends wrong field name**

- **Server** (`src/routes/users.ts`):
```typescript
interface CreateUserRequest { name: string; email: string; role: "admin" | "user"; }
app.post("/api/users", (req, res) => { const body: CreateUserRequest = req.body; });
```
- **Client** (`src/services/api-client.ts`):
```typescript
async function createUser(data: { name: string; email: string; type: string }) {
  return fetch("/api/users", { method: "POST", body: JSON.stringify(data) });
}
```
- **Expected violation:** `CTR-request-shape`. Client sends "type" but server expects "role". Missing field "role", extra field "type".

**TP-RS-02: Client missing required field**

- **Server expects:** `{ name: string; email: string; role: string; }`
- **Client sends:** `{ name: string; email: string; }`
- **Expected violation:** Missing "role" field.

**TP-RS-03: Type mismatch on field**

- **Server expects:** `{ id: number; }`
- **Client sends:** `{ id: string; }`
- **Expected violation:** Type mismatch: "id" is number on server, string on client.

**TP-RS-04: Go struct missing json tag field**

- **Server** (`handlers.go`):
```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}
```
- **Client** (`client.go`):
```go
type CreateUserPayload struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```
- **Expected violation:** Client missing "role" field.

**TP-RS-05: Fuzzy name mismatch (userId vs user_id)**

- **Config:** `fuzzyNameMatch: true`
- **Server expects:** `{ user_id: string; }`
- **Client sends:** `{ userId: string; }`
- **Expected violation:** Potential name mismatch: `userId` vs `user_id`.

**TP-RS-06: Nested object shape mismatch**

- **Server expects:** `{ address: { street: string; city: string; zip: string; } }`
- **Client sends:** `{ address: { street: string; city: string; } }`
- **Expected violation:** Missing "zip" in nested address.

**TP-RS-07: Array element type mismatch**

- **Server expects:** `{ items: { id: number; qty: number; }[] }`
- **Client sends:** `{ items: { id: string; quantity: number; }[] }`
- **Expected violation:** "id" type mismatch (number vs string), field name mismatch ("qty" vs "quantity").

**TP-RS-08: Client sends extra fields (strictExtraFields=true)**

- **Config:** `strictExtraFields: true`
- **Client sends:** `{ name: string; email: string; role: string; avatar: string; }`
- **Server expects:** `{ name: string; email: string; role: string; }`
- **Expected violation:** Extra field "avatar" on client.

**TP-RS-09: Annotation-based pair with mismatch**

- Both files annotated with `// stricture-contract: server=... client=...`
- Request shapes differ.
- **Expected violation:** Same detection as auto-detected pairs.

**TP-RS-10: Optional field on server, client sends wrong type**

- **Server:** `{ name: string; nickname?: string; }` (nickname optional)
- **Client:** `{ name: string; nickname: number; }` (nickname wrong type)
- **Expected violation:** Type mismatch on nickname (string vs number).

### 23.2 True Negative Cases

**TN-RS-01:** Client and server have matching request types.
**TN-RS-02:** Client sends extra optional fields (strictExtraFields=false).
**TN-RS-03:** Client imports type from shared package.
**TN-RS-04:** Go structs with matching json tags.
**TN-RS-05:** Optional field missing from client (ignoreOptionalFields=true).

### 23.3 False Positive Risks

**FP-RS-01:** Server uses middleware to transform request before handler sees it.
**FP-RS-02:** Client sends to external API (not in repo) -- should not pair.
**FP-RS-03:** Server accepts `any` type or untyped body.
**FP-RS-04:** REST convention differences (query params vs body).
**FP-RS-05:** Multipart form data vs JSON body.

### 23.4 False Negative Risks

**FN-RS-01:** Request type is `any` or `interface{}` -- no type info to compare.
**FN-RS-02:** Request body built dynamically (`body[key] = value`).
**FN-RS-03:** Middleware adds fields to request before handler.
**FN-RS-04:** Different endpoint versions (v1 vs v2) with different shapes.
**FN-RS-05:** GraphQL mutation vs REST endpoint (different paradigms).

### 23.5 Edge Cases

**EC-RS-01:** Path parameters vs body fields.
**EC-RS-02:** Query string parameters.
**EC-RS-03:** Headers as part of contract (Authorization, Content-Type).
**EC-RS-04:** Multipart form data fields.
**EC-RS-05:** File upload endpoints.
**EC-RS-06:** Server with multiple request types per endpoint (content negotiation).
**EC-RS-07:** Go request parsed with json.Decoder vs json.Unmarshal.
**EC-RS-08:** TypeScript request with Zod schema validation.
**EC-RS-09:** Request through API gateway (path rewriting).
**EC-RS-10:** WebSocket message shapes (not HTTP).

### 23.6 Configuration Interaction

**CI-RS-01:** `strictExtraFields: true` -- extra client fields are errors.
**CI-RS-02:** `strictExtraFields: false` -- extra client fields are warnings.
**CI-RS-03:** `fuzzyNameMatch: true` -- flag potential casing mismatches.
**CI-RS-04:** `ignoreOptionalFields: true` -- optional server fields not required from client.

### 23.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `CTR-request-shape`.

---

## 24. CTR-response-shape

**Rule purpose:** Verify that the response body type the server sends matches what the client expects.

### 24.1 True Positive Cases

**TP-RESP-01: Client expects field server does not send**

- **Server sends:** `{ id: number; name: string; email: string; created_at: string; }`
- **Client expects:** `{ id: number; name: string; email: string; createdAt: string; avatar: string; }`
- **Expected violation:** Name mismatch: "created_at" vs "createdAt". Missing from server: "avatar".

**TP-RESP-02: Go json tag mismatch in response**

- **Server:**
```go
type UserResponse struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```
- **Client:**
```go
type UserResponse struct {
    ID   int    `json:"user_id"` // MISMATCH
    Name string `json:"name"`
}
```
- **Expected violation:** Server sends "id", client expects "user_id".

**TP-RESP-03: Server response has extra fields client ignores**

- Not a violation by default (server can send more than client needs).
- But if configured strictly, could warn.

**TP-RESP-04: Type mismatch on response field**

- Server sends `{ count: number }`, client expects `{ count: string }`.
- **Expected violation:** Type mismatch.

**TP-RESP-05: Nested response object mismatch**

- Server sends `{ data: { items: { id: number }[] } }`
- Client expects `{ data: { items: { id: string }[] } }`
- **Expected violation:** Nested type mismatch on items[].id.

**TP-RESP-06: Server sends literal object, client expects typed interface**

- Server: `res.json({ id: user.id, name: user.name })` (no email)
- Client: `interface UserResponse { id: number; name: string; email: string; }`
- **Expected violation:** Client expects "email" but server does not send it.

**TP-RESP-07: Date field serialization mismatch**

- Server sends `createdAt` as ISO string, client expects `Date` object.
- **Expected:** This is a type mismatch (string vs Date). Whether to flag depends on serialization analysis.

**TP-RESP-08: Enum value mismatch**

- Server can send `role: "admin" | "user" | "moderator"`
- Client expects `role: "admin" | "user"`
- **Expected violation:** Server can send "moderator" which client does not handle.

**TP-RESP-09: Annotation-based pair with response mismatch**

- Same as annotation-based request detection but for response.

**TP-RESP-10: OpenAPI spec says one thing, code does another**

- OpenAPI spec defines response as `{ id: number; name: string; }`.
- Server actually sends `{ id: number; fullName: string; }`.
- **Expected violation:** Server implementation does not match spec.

### 24.2 True Negative Cases

**TN-RESP-01:** Matching response types on both sides.
**TN-RESP-02:** Client uses shared type imported from common package.
**TN-RESP-03:** Server sends superset of what client expects (extra fields OK).
**TN-RESP-04:** Go structs with matching json tags.
**TN-RESP-05:** Response type is `any`/`unknown` on client (no type to check).

### 24.3-24.7: Same structure as CTR-request-shape (False Positives, False Negatives, Edge Cases, Config Interaction, Inline Suppression)

Key differences: direction is reversed (server sends, client receives). Go `json.Marshal` tag resolution is critical. Object literal analysis on server side needed.

---

## 25. CTR-status-code-handling

**Rule purpose:** Verify that the client handles all status codes the server can return.

### 25.1 True Positive Cases

**TP-SCH-01: Client does not check status at all**

- **Server returns:** 200, 400, 404, 500
- **Client:**
```typescript
async function getUser(id: string): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  return res.json(); // No status check!
}
```
- **Expected violation:** Client handles 0/4 status codes.

**TP-SCH-02: Client only checks res.ok**

- **Config:** `requireExplicit: true`
- **Client:**
```typescript
if (!res.ok) throw new Error("request failed");
return res.json();
```
- **Expected violation:** `res.ok` is not explicit. Must handle 400, 404, 500 individually.

**TP-SCH-03: Client handles 200 and 404 but not 400 and 500**

- **Server can return:** 200, 400, 404, 500
- **Client handles:** 200, 404
- **Expected violation:** Unhandled: 400, 500.

**TP-SCH-04: Go client ignoring non-200 status**

- **Client:**
```go
resp, err := http.Get(url)
if err != nil { return nil, err }
defer resp.Body.Close()
var user User
json.NewDecoder(resp.Body).Decode(&user)
return &user, nil // No status code check!
```
- **Expected violation:** No status code handling.

**TP-SCH-05: Server returns 201 (Created), client only checks 200**

- **Client:** `if (res.status === 200)` -- misses 201.
- **Expected violation:** Unhandled 201.

**TP-SCH-06: Server returns 204 (No Content), client tries to parse body**

- **Client:**
```typescript
const data = await res.json(); // 204 has no body -- this will throw
```
- **Expected violation:** 204 not handled (attempting to parse empty body).

**TP-SCH-07: Client ignores 5xx errors (ignore5xx=false)**

- **Config:** `ignore5xx: false`
- **Client:** Handles 200, 400, 404 but not 500.
- **Expected violation:** 500 unhandled.

**TP-SCH-08: Server uses implicit 200 and explicit error codes**

- **Server:**
```typescript
app.get("/users", (req, res) => {
  // Implicit 200 on res.json(data)
  // Explicit 404 on res.status(404).json({error: "not found"})
  // Implicit 500 on unhandled exception
});
```
- **Client handles:** Only 200.
- **Expected violation:** 404 and 500 not handled.

**TP-SCH-09: Multiple endpoints with different status codes**

- Endpoint A returns 200, 400. Endpoint B returns 200, 404, 500. Client for B only handles 200.
- **Expected violation:** Only on client B's unhandled codes.

**TP-SCH-10: Go server returning custom error codes**

- Server: `w.WriteHeader(http.StatusConflict)` (409)
- Client: Does not handle 409.
- **Expected violation:** 409 unhandled.

### 25.2 True Negative Cases

**TN-SCH-01:** Client handles all server status codes explicitly.
**TN-SCH-02:** Client uses `res.ok` check (with `requireExplicit: false`).
**TN-SCH-03:** Client has catch-all `if (!res.ok)` plus specific handlers.
**TN-SCH-04:** Go client checks `resp.StatusCode` for all possible values.
**TN-SCH-05:** Client handles 5xx with `ignore5xx: true`.

### 25.3-25.7: Same structure as previous CTR rules (FP risks include: server status codes in unreachable code, generic error middleware; FN risks include: status codes from middleware not in handler; Edge cases include redirect status codes 301/302, HEAD requests, CORS preflight 204).

### 25.6 Configuration Interaction

**CI-SCH-01:** `requireExplicit: true` -- `res.ok` alone is not enough.
**CI-SCH-02:** `requireExplicit: false` -- `res.ok` counts as handling non-2xx.
**CI-SCH-03:** `ignore5xx: true` -- 5xx codes do not need explicit handling.
**CI-SCH-04:** `ignore5xx: false` -- 5xx codes must be handled.

---

## 26. CTR-shared-type-sync

**Rule purpose:** When client and server reference the same type name, verify the definitions are identical or from a shared package.

### 26.1 True Positive Cases

**TP-STS-01: Same name, different fields (TS)**

- **Server** (`src/routes/types.ts`):
```typescript
interface User { id: number; name: string; email: string; }
```
- **Client** (`src/services/types.ts`):
```typescript
interface User { id: number; name: string; } // Missing email
```
- **Expected violation:** `CTR-shared-type-sync`. Type "User" defined in both files with different fields. Server has "email", client does not.

**TP-STS-02: Same name, different types on field (Go)**

- **Server:**
```go
type User struct { ID int; Name string; }
```
- **Client:**
```go
type User struct { ID string; Name string; } // ID is string, not int
```
- **Expected violation:** Type mismatch on "ID": int vs string.

**TP-STS-03: Extra field on one side**

- Server User has 5 fields, client User has 4.
- **Expected violation:** Field count mismatch.

**TP-STS-04: Same fields, different json tags (Go)**

- **Server:** `Name string \`json:"name"\``
- **Client:** `Name string \`json:"user_name"\``
- **Expected violation:** JSON tag mismatch.

**TP-STS-05: Type redefined locally instead of importing shared**

- **Config:** `requireSharedPackage: true`
- Both server and client define their own `UserResponse` type.
- **Expected violation:** Type should be imported from shared package.

**TP-STS-06: Deeply nested type with difference at leaf**

- Both sides define `OrderResponse` with nested `items[].product.name`. Server has `name: string`, client has `name: string | null`.
- **Expected violation:** Type mismatch at nested level.

**TP-STS-07: Enum/union type with different members**

- Server: `type Role = "admin" | "user" | "moderator"`
- Client: `type Role = "admin" | "user"`
- **Expected violation:** Client missing "moderator" variant.

**TP-STS-08: Go embedded struct difference**

- Server embeds `BaseModel` with 3 fields, client does not embed.
- **Expected violation:** Field count mismatch after flattening embedded struct.

**TP-STS-09: Optional vs required field difference**

- Server: `email: string` (required)
- Client: `email?: string` (optional)
- **Expected violation:** Optionality mismatch.

**TP-STS-10: Array vs single value**

- Server: `items: Item[]`
- Client: `items: Item`
- **Expected violation:** Type mismatch (array vs single).

### 26.2-26.7: Standard structure for TN, FP, FN, Edge Cases, Config, Inline Suppression.

Key config options:
- `requireSharedPackage: false/true` -- whether duplicate type names always error.
- `ignoreTestFiles: true` -- test-local type redefinitions are OK.

---

## 27. CTR-json-tag-match

**Rule purpose:** (Go-specific) Verify that JSON struct tags match across contract boundaries.

### 27.1 True Positive Cases

**TP-JTM-01: Snake_case vs camelCase tag mismatch**

- **Server:**
```go
type UserResponse struct {
    CreatedAt time.Time `json:"created_at"`
}
```
- **Client:**
```go
type UserResponse struct {
    CreatedAt time.Time `json:"createdAt"` // MISMATCH
}
```
- **Expected violation:** JSON tag mismatch: server uses "created_at", client uses "createdAt".

**TP-JTM-02: Missing json tag on one side**

- **Server:** `Name string \`json:"name"\``
- **Client:** `Name string` (no json tag -- defaults to "Name")
- **Expected violation:** Server sends "name", client expects "Name".

**TP-JTM-03: Omitempty on one side only**

- **Server:** `Email string \`json:"email,omitempty"\``
- **Client:** `Email string \`json:"email"\``
- **Expected:** Warning. `omitempty` means server might not send the field.

**TP-JTM-04: json:"-" on one side**

- **Server:** `Password string \`json:"-"\``
- **Client:** `Password string \`json:"password"\``
- **Expected violation:** Server never sends "password" (json:"-"), but client expects it.

**TP-JTM-05: Nested struct tag mismatch**

- Server's nested struct has `json:"user_id"`, client's nested struct has `json:"userId"`.
- **Expected violation:** Nested tag mismatch.

**TP-JTM-06: All tags use different conventions**

- Server consistently uses snake_case, client consistently uses camelCase.
- **Expected violation:** Multiple mismatches flagged, with suggestion to standardize.

**TP-JTM-07: Embedded struct with conflicting tags**

- Embedded struct promotes fields. Tag on promoted field differs.
- **Expected violation:** After flattening embedded structs, compare promoted field tags.

**TP-JTM-08: Slice element struct tag mismatch**

- Server: `Items []struct { ID int \`json:"id"\` }`
- Client: `Items []struct { ID int \`json:"item_id"\` }`
- **Expected violation:** Array element struct tag mismatch.

**TP-JTM-09: Convention enforcement**

- **Config:** `convention: snake_case`
- **Input:** Struct with `json:"userId"` (camelCase)
- **Expected violation:** Tag "userId" does not match enforced snake_case convention.

**TP-JTM-10: Map key struct tag**

- `map[string]struct{ Value int \`json:"val"\` }` on one side, `json:"value"` on other.
- **Expected violation:** Tag mismatch.

### 27.2 True Negative Cases

**TN-JTM-01:** Matching tags on both sides.
**TN-JTM-02:** Both sides import type from shared package.
**TN-JTM-03:** Tags match convention (all snake_case).
**TN-JTM-04:** Non-contract struct (internal only) -- not checked.
**TN-JTM-05:** TypeScript file -- rule is Go-only, skipped.

### 27.3-27.7: Standard structure for FP, FN, Edge Cases, Config, Inline Suppression.

Key edge cases: protobuf-generated structs, custom JSON marshalers, struct tags with multiple keys (`json:"name" xml:"name"`).

---

## 28. CTR-dual-test

**Rule purpose:** Verify that contract pairs have tests on both sides that exercise the same scenarios.

### 28.1 True Positive Cases

**TP-DT-01: Server tests 404, client has no 404 test**

- **Server test:**
```typescript
it("returns 404 for unknown user", async () => {
  const res = await request(app).get("/api/users/unknown");
  expect(res.status).toBe(404);
});
```
- **Client test:** Only tests 200 success case.
- **Expected violation:** `CTR-dual-test`. Server tests 404 for /api/users/:id, but client has no matching 404 test.

**TP-DT-02: Client tests error, server has no error test**

- **Client test:**
```typescript
it("handles network error", async () => {
  mock("/api/users/123", { status: 500 });
  await expect(getUser("123")).rejects.toThrow(ServerError);
});
```
- **Server test:** No test that returns 500.
- **Expected violation:** Client tests 500 from /api/users/:id, server has no matching 500 test.

**TP-DT-03: Server tests validation (400), client ignores validation**

- Server has test for 400 on invalid input. Client has no test for validation error handling.
- **Expected violation:** Unmatched scenario.

**TP-DT-04: Multiple endpoints, one without dual tests**

- `/api/users` has dual tests. `/api/orders` has server tests but no client tests.
- **Expected violation:** Only /api/orders flagged.

**TP-DT-05: Go server test with no matching Go client test**

- Server tests return different status codes. Client tests only cover success.
- **Expected violation:** Per unmatched scenario.

**TP-DT-06: Server test uses different path format**

- Server registers `/api/users/:id`, test uses `/api/users/123`. Client fetches `/api/users/${id}`.
- **Expected:** Should still match (both reference same endpoint pattern).

**TP-DT-07: Only success scenarios tested on both sides**

- Both sides test 200. Neither tests 400/404/500.
- **Expected violation:** If `requireBothDirections: true` and server can return 400/404/500, the missing error scenarios should be flagged.

**TP-DT-08: Client test uses mock, server test is integration**

- Both exist but test different things. Client mocks the server. Server tests real logic.
- **Expected:** This is acceptable -- both sides have tests. Scenario fingerprint matching should still work.

**TP-DT-09: Confidence below threshold**

- **Config:** `minConfidence: 80`
- Match is at 70% confidence.
- **Expected:** No violation (below threshold).

**TP-DT-10: Confidence above threshold**

- Match is at 90% confidence.
- **Expected violation:** Flagged.

### 28.2 True Negative Cases

**TN-DT-01:** Both sides have matching test scenarios for all status codes.
**TN-DT-02:** Client has error handling tests that match server error tests.
**TN-DT-03:** Both sides test the same endpoint with same scenarios.
**TN-DT-04:** Internal endpoint with `ignoreInternalEndpoints: true`.
**TN-DT-05:** No contract pairs detected -- rule does not apply.

### 28.3-28.7: Standard structure for FP, FN, Edge Cases, Config, Inline Suppression.

Key configuration:
- `minConfidence: 80` -- only flag above this confidence.
- `requireBothDirections: true` -- both server-to-client and client-to-server checked.
- `ignoreInternalEndpoints: false` -- check all endpoints.

---

# PART 2: Cross-Cutting Test Sections

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

# PART 3: Spec Gap Analysis

---

## 39. Ambiguities in the Spec

**GAP-AMB-01: ARCH-dependency-direction -- Does "top-down" allow skipping layers?**

The spec says `handler -> service -> repository -> model (never reverse)`. It is unclear whether `handler -> repository` (skipping service) is a violation. Top-down means handler is above repository, so the direction is correct. But strict layering would require going through service. The spec should clarify: is skipping layers allowed or not?

**Recommendation:** Add a `strictLayering: boolean` option. When true, each layer can only import from the immediately adjacent lower layer. When false (default), any downward import is allowed.

**GAP-AMB-02: TQ-no-shallow-assertions -- What qualifies as a "presence test"?**

The spec says tests named with "exists", "present", "defined" are presence tests. But what about "should be available", "should not be undefined", "verifies creation"? The keyword list is too narrow and subjective.

**Recommendation:** Define an exact list of keywords or a regex pattern for presence test detection. Allow users to configure it.

**GAP-AMB-03: CONV-file-header -- What counts as line 1?**

Is line 1 the first non-empty line? The first line of the file? What about BOM, shebang, Go build constraints? The spec says "top of every file" without being precise.

**Recommendation:** Define: "The header must appear within the first 3 non-blank, non-directive lines." Directives include shebang, Go build constraints, and `"use strict"`.

**GAP-AMB-04: CTR rules -- How are contract pairs identified in practice?**

The spec lists 4 detection methods but does not specify which takes priority when they conflict. If annotation says "server=A client=B" but route-matching detects "server=A client=C", which pair is used?

**Recommendation:** Define priority: annotation > api-spec > route-matching > shared-type.

**GAP-AMB-05: TQ-return-type-verified -- How are "fields asserted across multiple tests" aggregated?**

The spec says "cross-test aggregation" counts fields asserted in different `it()` blocks. But across different `describe()` blocks? Different files? What if the function is called with different inputs in different tests and the return type varies?

**Recommendation:** Aggregate within a single test file. Cross-file aggregation should be opt-in.

**GAP-AMB-06: CONV-file-naming -- How to handle filenames with multiple dots?**

Is `schema.generated.ts` valid kebab-case? The name portion is `schema.generated`, which contains a dot. Is the convention applied to the stem before the extension, or to the full filename?

**Recommendation:** Apply naming convention to the stem before the first extension (everything before `.ts`, `.go`, etc.). Multiple extensions like `.generated.ts` should be treated as compound extensions.

**GAP-AMB-07: TQ-mock-scope -- What constitutes "cleanup"?**

The spec mentions `mockRestore()` and `jest.restoreAllMocks()`. What about `jest.clearAllMocks()`, `jest.resetAllMocks()`, or manual reassignment of the original value?

**Recommendation:** Define a list of recognized cleanup patterns per test framework.

**GAP-AMB-08: ARCH-module-boundary -- When does a directory become a "module"?**

The spec says directories with `index.ts` are modules. What about directories with `index.js`? `mod.ts` (Deno convention)? `package.json` with `main` field? Go packages (every directory is a package)?

**Recommendation:** For TypeScript: `index.ts`, `index.tsx`, `index.js`, `index.jsx`. For Go: every directory with `.go` files. Allow config to specify custom module entry points.

---

## 40. Missing Specifications

**GAP-MISS-01: Baseline/legacy mode**

The spec mentions `--baseline` in Appendix B (open questions) but does not specify it. For existing codebases with thousands of violations, there is no way to adopt Stricture incrementally.

**Recommendation:** Specify `--baseline <file>` that saves current violations to a baseline file. Subsequent runs only report NEW violations not in the baseline.

**GAP-MISS-02: Watch mode**

No mention of `--watch` for development-time continuous linting.

**Recommendation:** Specify `--watch` that re-runs on file changes.

**GAP-MISS-03: Ignore comments for individual lines (not just rules)**

The spec has `stricture-disable-next-line RULE-ID` and block disable/enable. But what about disabling ALL rules for a line? Is `// stricture-disable-next-line` (no rule ID) valid?

**Recommendation:** Specify that `stricture-disable-next-line` without a rule ID disables all rules for the next line.

**GAP-MISS-04: Rule documentation command**

`stricture list-rules` shows rules, but there is no `stricture explain RULE-ID` to get detailed documentation for a specific rule.

**Recommendation:** Add `stricture explain <rule-id>` command.

**GAP-MISS-05: Caching specification**

The spec mentions `.stricture-cache/` but does not specify cache format, maximum size, or cleanup. Cache could grow unboundedly.

**Recommendation:** Specify: LRU cache with configurable max size. Default 100MB. Auto-cleanup of entries older than 7 days.

**GAP-MISS-06: How CONV-test-file-location handles test utilities**

Test utility files (`test-helpers.ts`, `fixtures.ts`) are not test files but live in test directories. The rule might incorrectly flag them as misplaced.

**Recommendation:** Specify that only files matching test suffixes (`.test.ts`, `_test.go`) are subject to this rule, not all files in test directories.

**GAP-MISS-07: Multi-file violations**

Some violations span multiple files (ARCH-no-circular-deps, CTR-shared-type-sync). The output schema shows violations per file, but a cycle involves multiple files. How is this reported?

**Recommendation:** Allow violations to have a `relatedLocations` array for multi-file issues.

**GAP-MISS-08: Exit code behavior with --fix**

The spec defines exit codes for linting. What exit code does `--fix` produce? 0 if all fixable violations are fixed? 1 if non-fixable violations remain?

**Recommendation:** After `--fix`, re-lint the fixed files. Exit code based on remaining violations.

---

## 41. Contradictions

**GAP-CONTRA-01: Section 3.3 vs Section 5.2**

Section 3.3 (Design Principles) says "When uncertain, Stricture should flag the issue (potential false positive)." But Section 5.2 (Inline Suppressions) provides suppression mechanisms. If the tool aggressively flags uncertain cases, suppression will be heavily used, defeating the "zero configuration to start" principle (3.4).

**Resolution:** This is a deliberate tension, not a true contradiction. Document it as: "Stricture is strict by default. Suppression is the release valve."

**GAP-CONTRA-02: CONV-file-naming fixable=Yes vs ARCH rules fixable=No**

CONV-file-naming can auto-fix by renaming files. But renaming a file changes import paths, which could introduce ARCH violations (if the new path crosses a boundary). The fix for one rule could create violations for another.

**Resolution:** After auto-fix, Stricture should re-lint fixed files and report any new violations introduced by fixes.

**GAP-CONTRA-03: Exit code 2 "on parse error" vs "only if all files are unparseable"**

Section 4 says exit code 2 for "parse error." Section 9.3 is ambiguous. If one file has a syntax error among 100 valid files, is the exit code 2 or 0?

**Resolution:** Exit code 2 only for configuration errors. Parse errors on individual files should produce warnings, with exit code 0 or 1 depending on whether violations are found in parseable files.

---

## 42. Scalability Concerns

**GAP-SCALE-01: Contract detection in large monorepos**

Route-matching and shared-type detection require analyzing ALL files to build contract pairs. In a 10,000-file monorepo, this could be very slow and produce many false pairs.

**Recommendation:** Limit contract detection scope via config. Allow specifying server/client directories explicitly. Cache contract pair detection results.

**GAP-SCALE-02: ARCH-no-circular-deps on large dependency graphs**

Tarjan's algorithm is O(V+E) which is fine, but reporting all cycles in a large graph with many cycles could produce overwhelming output.

**Recommendation:** Limit to top-N shortest cycles. Or report only NEW cycles not in baseline.

**GAP-SCALE-03: TQ rules requiring cross-file analysis**

TQ rules need test file + source file. For each test file, resolving the source file requires parsing imports and conventions. In a large repo, this is O(N) per test file.

**Recommendation:** Cache test-to-source mappings. Build the mapping once and reuse.

**GAP-SCALE-04: ProjectContext for every ARCH rule**

ARCH rules need the full dependency graph. Building this for 10,000 files on every run is expensive.

**Recommendation:** Incremental ProjectContext updates -- only rebuild the parts of the graph affected by changed files.

**GAP-SCALE-05: Generated code overwhelming violation counts**

A large protobuf-generated file might trigger hundreds of TQ/CONV violations, obscuring real issues.

**Recommendation:** Default ignore patterns should include common generated file patterns (`*.generated.*`, `*.pb.go`, `*.pb.ts`).

---

## 43. Security Considerations

**GAP-SEC-01: Plugin sandboxing**

Plugins are JavaScript/TypeScript files executed by the Stricture process. A malicious plugin could:
- Read/write arbitrary files
- Execute shell commands
- Exfiltrate code to external servers
- Consume unbounded resources (CPU/memory)

**Recommendation:** Document that plugins run with full process permissions. In future versions, consider running plugins in a sandboxed worker with restricted `fs` and `net` access.

**GAP-SEC-02: Config injection**

`extends` can reference npm packages. If a team extends `@evil/stricture-config`, that package could:
- Override rules to hide violations
- Include malicious plugins
- Reference local file paths for data exfiltration

**Recommendation:** Validate that `extends` targets are trusted. Log which configs are loaded. Consider a `--trust` flag for first-time use of external configs.

**GAP-SEC-03: Path traversal via symlinks**

A malicious symlink could point to sensitive files outside the project:
```
src/sensitive -> /etc/passwd
```
Stricture would read and analyze the file, potentially including its content in output.

**Recommendation:** Resolve symlinks and verify they stay within the project root. Skip symlinks that resolve outside the project.

**GAP-SEC-04: SARIF output may contain sensitive code snippets**

SARIF output includes source code context. If uploaded to GitHub Code Scanning, sensitive code (API keys, credentials) could be exposed.

**Recommendation:** Strip sensitive patterns from SARIF output. Or add `--redact` option.

**GAP-SEC-05: Cache poisoning**

If `.stricture-cache/` is writable by other processes or users, a malicious actor could inject false AST data, causing Stricture to miss violations.

**Recommendation:** Validate cache integrity via content hash. Set restrictive file permissions on cache directory.

---

## 44. Recommendations

**REC-01:** Add `--baseline` support before v0.1.0 release. Without it, adoption on existing projects is impractical.

**REC-02:** Define exact keyword lists for heuristics (presence test detection, negative test classification). Vague heuristics lead to unpredictable behavior.

**REC-03:** Add a `stricture explain <rule-id>` CLI command that prints the full rule documentation, detection algorithm, and examples.

**REC-04:** Specify cache format, size limits, and cleanup strategy.

**REC-05:** For CTR rules, require explicit contract pair declaration in config for v0.1.0. Auto-detection is complex and error-prone. Ship auto-detection in v0.2.0 after real-world testing.

**REC-06:** Add `--max-violations N` flag to fail fast after N violations. Useful in CI to avoid processing the entire repo when early failures are detected.

**REC-07:** Define behavior when `stricture init` is run in a directory that already has `.stricture.yml`. Overwrite? Merge? Error?

**REC-08:** Specify whether `--fix` creates a backup of modified files. Consider `--fix --backup` to save `.bak` files.

**REC-09:** Add file-level and directory-level violation suppression: `// stricture-disable-file RULE-ID` at the top of a file suppresses the rule for the entire file.

**REC-10:** Consider adding a `--diff` output mode that shows only the diff from baseline, making it easy to see what is new in a PR.

---

*End of test plan.*
