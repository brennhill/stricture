# Test Quality (TQ) Rules — Test Cases

10 rules covering test quality analysis.

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

