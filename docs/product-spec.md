# Stricture — Product Specification

> **The linter that makes tests mean something.**
>
> Stricture is a higher-order linting tool that enforces architectural constraints, test quality, and codebase conventions — the things ESLint, golangci-lint, and Codacy don't cover. Its killer feature: deep test analysis that makes it impossible for AI (or humans) to ship weak tests that claim full coverage. If Stricture passes, the code should just work.

**Version:** 0.1.0 (initial release)
**Status:** Product Spec
**Author:** Brenn
**Date:** 2026-02-12

---

## Table of Contents

1. [Problem Statement](#1-problem-statement)
2. [Target Users](#2-target-users)
3. [Design Principles](#3-design-principles)
4. [Architecture Overview](#4-architecture-overview)
5. [Configuration System](#5-configuration-system)
6. [Rule Categories](#6-rule-categories)
   - 6.1 [Test Quality (TQ)](#61-test-quality-tq)
   - 6.2 [Architecture (ARCH)](#62-architecture-arch)
   - 6.3 [Convention (CONV)](#63-convention-conv)
   - 6.4 [Contract (CTR)](#64-contract-ctr)
7. [Language Adapters](#7-language-adapters)
8. [Plugin System](#8-plugin-system)
9. [CLI Interface](#9-cli-interface)
10. [Output & Reporting](#10-output--reporting)
11. [Auto-Fix System](#11-auto-fix-system)
12. [CI Integration](#12-ci-integration)
13. [Cross-Service Contracts & Stricture Manifest](#13-cross-service-contracts--stricture-manifest)
14. [Performance Requirements](#14-performance-requirements)
15. [Rule Reference (All Rules)](#15-rule-reference-all-rules)
16. [Glossary](#16-glossary)

---

## 1. Problem Statement

### What existing linters cover

ESLint, golangci-lint, Pylint, ShellCheck, and Codacy catch **syntax-level** problems: unused variables, missing semicolons, unreachable code, simple security patterns, cyclomatic complexity. These are solved problems.

### What nothing covers

1. **Test quality beyond coverage metrics.** A test file can have 100% line coverage while asserting nothing meaningful. AI-generated tests routinely produce tests that check `expect(result).toBeDefined()` instead of verifying the actual output shape, values, and error paths. Code coverage tools say "100%" and the tests are worthless.

2. **Architectural constraint enforcement.** No existing tool enforces "handlers must not import from the data layer directly" or "all files in `internal/` must not be imported by `cmd/`" as a CI-blocking lint rule. Teams document these in ADRs that nobody reads.

3. **Convention drift.** File naming conventions, error handling patterns, required file headers, module boundary rules — these degrade gradually and are only caught in code review, inconsistently.

### The core insight

**The gap is between "code that compiles" and "code that's correct."** Existing linters verify the former. Stricture verifies the latter.

### Why now

AI code generation (Claude, Copilot, Codex) produces syntactically correct code that passes existing linters but often has:
- Shallow test assertions that verify structure but not behavior
- Architectural violations (the AI doesn't know your module boundaries)
- Convention drift (the AI approximates but doesn't match exact patterns)

Stricture is the CI gate that catches what AI gets wrong.

---

## 2. Target Users

### Primary: Engineering teams using AI-assisted development

- Teams where AI agents generate code and tests
- Need automated quality gates beyond syntax linting
- Want to trust that "all tests pass" means "the code works"

### Secondary: Open-source maintainers

- Enforce contribution standards without manual review burden
- Codify architectural decisions as executable rules
- Reduce review fatigue from convention violations

### Tertiary: Platform/infrastructure teams

- Enforce cross-repo standards via shared config packages
- Audit codebase health with structured reports

---

## 3. Design Principles

### 3.1 Strict by default, relaxable per-rule

Every rule ships enabled at its strictest setting. Teams relax rules they don't need. This is the opposite of ESLint's "configure everything" approach. Rationale: it's better to have a team disable a rule consciously than to never discover it exists.

### 3.2 Deep over broad

Prefer fewer rules with deep analysis over many shallow rules. One rule that truly verifies assertion quality is worth more than 50 rules that count assertion lines.

### 3.3 No false negatives over no false positives

When uncertain, Stricture should flag the issue (potential false positive) rather than miss it (false negative). Users can suppress with inline comments. Rationale: the tool exists to catch things humans and AI miss. Missing something defeats the purpose.

### 3.4 Zero configuration to start

Running `npx stricture` in any Go or TypeScript project should produce useful output with zero config. The default ruleset covers the most universal concerns. Configuration is for customization, not setup.

### 3.5 Explainable violations

Every violation must include:
- What rule was violated
- Why this rule exists (one sentence)
- Where the violation is (file, line, column)
- How to fix it (concrete suggestion, not vague guidance)
- How to suppress it (inline comment syntax)

---

## 4. Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                   CLI / Entry                    │
│  npx stricture [options] [paths...]              │
├─────────────────────────────────────────────────┤
│                Config Loader                     │
│  .stricture.yml → merged config object           │
├──────────┬──────────────┬───────────────────────┤
│ Language │   Language   │     Language           │
│ Adapter: │   Adapter:   │     Adapter:           │
│ TypeScript│     Go      │   (future: Py, Rust)   │
│ (ts-morph)│ (go/parser) │                        │
├──────────┴──────────┬───┴───────────────────────┤
│              Unified File Model                  │
│  { path, language, ast, imports, exports,        │
│    testTargets, functions, classes, types }       │
├─────────────────────────────────────────────────┤
│                Rule Engine                        │
│  For each file → run matching rules → collect     │
│  violations. Rules receive UnifiedFileModel +     │
│  ProjectContext (all files, dependency graph).     │
├──────────┬──────────┬───────────┬───────────────┤
│ Built-in │ Built-in │ Built-in  │   User        │
│ Rules:   │ Rules:   │ Rules:    │   Plugins     │
│ TQ-*     │ ARCH-*   │ CONV-*    │   (.js/.ts)   │
├──────────┴──────────┴───────────┴───────────────┤
│              Reporter / Formatter                │
│  text | json | sarif | junit                     │
├─────────────────────────────────────────────────┤
│              Auto-Fix Engine                     │
│  (optional) Apply fixes → write files            │
└─────────────────────────────────────────────────┘
```

### Key Components

| Component | Responsibility | Key Decisions |
|-----------|---------------|---------------|
| **Config Loader** | Parse `.stricture.yml`, merge with defaults, resolve plugin paths | YAML format. Supports `extends` for shared configs. |
| **Language Adapter** | Parse source into `UnifiedFileModel` | One adapter per language. Adapters are npm packages. |
| **UnifiedFileModel** | Language-agnostic representation of a source file | Contains AST, imports, exports, functions, types, test metadata. Adapters produce this. Rules consume this. |
| **ProjectContext** | Cross-file analysis state | Dependency graph, file index, module boundaries. Built once, shared across rules. |
| **Rule Engine** | Orchestrate rule execution, collect violations | Runs rules in dependency order. Supports `--rule` flag for single-rule execution. |
| **Reporter** | Format violations for output | Multiple formats. SARIF for GitHub, JUnit for CI, JSON for tooling, text for humans. |
| **Auto-Fix Engine** | Apply safe code transformations | Only runs with `--fix` or `--fix-dry-run`. Never modifies files without explicit opt-in. |
| **Plugin Loader** | Load and validate user plugins | Plugins export a rule factory function. Validated at load time for required interface. |

### Data Flow

```
1. CLI parses args → determines target files (glob/git diff)
2. Config Loader reads .stricture.yml → produces merged Config
3. For each target file:
   a. Language Adapter parses → UnifiedFileModel
   b. If test file: identify test targets (functions/methods under test)
4. Build ProjectContext (dependency graph, module map)
5. Rule Engine iterates rules:
   a. CONV rules: per-file, fast, no cross-file deps
   b. ARCH rules: need ProjectContext (dependency graph)
   c. TQ rules: need test file + its target source file(s)
6. Collect all Violations
7. Reporter formats output
8. Exit code: 0 if no errors, 1 if errors, 2 if config/parse failure
```

---

## 5. Configuration System

### 5.1 Config File: `.stricture.yml`

Located at project root. Supports `extends` for shared configurations.

```yaml
# .stricture.yml — Example configuration

# Inherit from a shared config (npm package or local path)
extends:
  - "@stricture/config-recommended"
  - "./configs/team-overrides.yml"

# Language adapters to activate (auto-detected if omitted)
languages:
  - typescript
  - go

# Global settings
settings:
  # Max parallel file processing (default: os.cpus().length)
  concurrency: 8
  # Fail on first error (default: false)
  bail: false
  # Files/directories to ignore (in addition to .gitignore)
  ignore:
    - "vendor/**"
    - "dist/**"
    - "*.generated.ts"
    - "**/*.d.ts"

# Rule configuration
# Each rule can be: "error" | "warn" | "off" | ["error", { ...options }]
rules:
  # ── Test Quality ──────────────────────────────
  TQ-assertion-depth:          error
  TQ-return-type-verified:     error
  TQ-error-path-coverage:      error
  TQ-schema-conformance:       error
  TQ-no-shallow-assertions:    error
  TQ-boundary-tested:          error
  TQ-mock-scope:               error
  TQ-test-isolation:           error
  TQ-negative-cases:           error
  TQ-test-naming:              [error, { pattern: "should {verb} when {condition}" }]

  # ── Architecture ──────────────────────────────
  ARCH-dependency-direction:   error
  ARCH-import-boundary:        [error, {
    boundaries: [
      { from: "cmd/**",      deny: ["internal/capture/**"] },
      { from: "src/content/**", deny: ["src/background/**"] },
    ]
  }]
  ARCH-max-file-lines:         [error, { max: 800 }]
  ARCH-no-circular-deps:       error
  ARCH-layer-violation:        [error, {
    layers: ["handler", "service", "repository"],
    direction: "top-down"
  }]
  ARCH-module-boundary:        error

  # ── Contract ────────────────────────────────────
  CTR-request-shape:           error
  CTR-response-shape:          error
  CTR-status-code-handling:    error
  CTR-shared-type-sync:        error
  CTR-json-tag-match:          error
  CTR-dual-test:               [error, { minConfidence: 80 }]
  CTR-strictness-parity:       error   # Requires manifest section (see §13)
  CTR-manifest-conformance:    error   # Requires manifest section (see §13)

  # ── Convention ────────────────────────────────
  CONV-file-naming:            [error, { style: "kebab-case" }]
  CONV-file-header:            [error, { pattern: "// {filename} — {purpose}" }]
  CONV-error-format:           [error, { pattern: "{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}" }]
  CONV-export-naming:          [error, { public: "PascalCase", private: "camelCase" }]
  CONV-test-file-location:     [error, { strategy: "colocated" }]  # or "mirrored"
  CONV-required-exports:       [error, { patterns: ["src/features/*/index.ts"] }]
```

### 5.2 Inline Suppressions

```typescript
// stricture-disable-next-line TQ-assertion-depth
expect(result).toBeDefined();

// stricture-disable TQ-no-shallow-assertions
// ... block of code ...
// stricture-enable TQ-no-shallow-assertions
```

```go
// stricture-disable-next-line ARCH-import-boundary
import "internal/capture"
```

### 5.3 Config Resolution Order

1. CLI flags (highest priority)
2. `.stricture.yml` in project root
3. `extends` chain (resolved left-to-right, later overrides earlier)
4. Built-in defaults (lowest priority)

### 5.4 Shared Configurations

Shared configs are npm packages that export a YAML-compatible object:

```
@stricture/config-recommended   — sensible defaults for any project
@stricture/config-strict        — maximum strictness, all rules error
@stricture/config-gasoline      — Gasoline project conventions
```

---

## 6. Rule Categories

### 6.1 Test Quality (TQ)

These are the highest-value rules. They analyze test files in relation to the source files they test, verifying that tests actually prove the code works.

#### How test-to-source mapping works

Stricture must determine which source file(s) a test file covers. Resolution strategy (in order):

1. **Explicit mapping** in config: `testTargets: { "tests/foo.test.ts": "src/foo.ts" }`
2. **Convention-based**: `foo.test.ts` → `foo.ts`, `foo_test.go` → `foo.go`
3. **Import analysis**: test file imports from source file → implicit mapping
4. **Manual annotation**: `// stricture-target: src/foo.ts` in test file header

If no mapping is found, TQ rules that require cross-file analysis emit a warning and skip.

---

#### TQ-no-shallow-assertions

**Purpose:** Reject assertions that verify existence/truthiness without checking actual values.

**What it catches:**

```typescript
// VIOLATION: Shallow — only checks existence
expect(result).toBeDefined();
expect(result).toBeTruthy();
expect(result).not.toBeNull();
expect(user).toBeInstanceOf(Object);

// VIOLATION: Shallow — type check without value check
expect(typeof result).toBe("object");
expect(Array.isArray(items)).toBe(true);

// OK: Checks actual values
expect(result.status).toBe(200);
expect(result.body).toEqual({ id: 1, name: "Alice" });
expect(items).toHaveLength(3);
expect(items[0].name).toBe("first");
```

```go
// VIOLATION: Only checks nil
if result == nil {
    t.Fatal("result is nil")
}

// VIOLATION: Only checks error existence
if err != nil {
    t.Fatal(err)
}
// (no assertion on err's type, message, or code)

// OK: Checks actual values
if result.Status != 200 {
    t.Errorf("expected status 200, got %d", result.Status)
}
if !errors.Is(err, ErrNotFound) {
    t.Errorf("expected ErrNotFound, got %v", err)
}
```

**Detection algorithm:**

1. Parse test file AST
2. Find all assertion call expressions (`expect()`, `t.Fatal`, `t.Errorf`, `assert.*`, etc.)
3. Classify each assertion as "shallow" or "deep":
   - **Shallow:** Asserts only on existence, truthiness, type identity, or instance-of without value comparison
   - **Deep:** Compares against a specific expected value, pattern, or structure
4. If shallow assertion count > 0 AND the assertion's target has deeper properties available (based on source file type analysis), flag as violation
5. **Exception:** `toBeDefined()` on an optional field in a "presence test" (test name contains "exists", "present", "defined") is allowed — but only if other tests in the same file also assert the value

**Options:**
```yaml
TQ-no-shallow-assertions:
  - error
  - allowInPresenceTests: true    # Allow toBeDefined() in tests named "should exist"
    maxShallowPercent: 10          # Allow up to 10% shallow assertions per file
```

---

#### TQ-return-type-verified

**Purpose:** Ensure that when a function returns a structured type, tests verify all required fields of that type — not just one or two fields.

**What it catches:**

```typescript
// Source: src/user.ts
interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "user";
  createdAt: Date;
}
function createUser(input: CreateUserInput): User { ... }

// Test: tests/user.test.ts
// VIOLATION: Only checks 2 of 5 fields
const user = createUser(validInput);
expect(user.id).toBe(1);
expect(user.name).toBe("Alice");
// Missing: email, role, createdAt are NEVER asserted in this test or any test in the file

// OK: All required fields verified (can be across multiple test cases)
expect(user.id).toBe(1);
expect(user.name).toBe("Alice");
expect(user.email).toBe("alice@example.com");
expect(user.role).toBe("user");
expect(user.createdAt).toBeInstanceOf(Date);

// ALSO OK: Using toEqual/toMatchObject with all fields
expect(user).toEqual({
  id: 1,
  name: "Alice",
  email: "alice@example.com",
  role: "user",
  createdAt: expect.any(Date),
});
```

```go
// Source: user.go
type User struct {
    ID        int
    Name      string
    Email     string
    Role      string
    CreatedAt time.Time
}

// Test: user_test.go
// VIOLATION: Only checks 1 of 5 fields
user := createUser(validInput)
if user.Name != "Alice" {
    t.Errorf("expected Alice, got %s", user.Name)
}
// Missing: ID, Email, Role, CreatedAt never asserted anywhere in file

// OK: All fields checked
if user.ID != 1 { ... }
if user.Name != "Alice" { ... }
if user.Email != "alice@example.com" { ... }
if user.Role != "user" { ... }
if user.CreatedAt.IsZero() { ... }
```

**Detection algorithm:**

1. Identify functions under test from test-to-source mapping
2. Resolve the return type of each function (via adapter's type system)
3. Flatten the return type into a set of required fields (skip optional fields)
4. Scan all assertions in the test file for field accesses on the return value
5. Compute "field coverage": (fields asserted / fields in type) as a percentage
6. If field coverage < threshold (default 80%), flag as violation
7. **Cross-test aggregation:** Fields asserted across different `it()` / `t.Run()` blocks for the same function count toward coverage

**Options:**
```yaml
TQ-return-type-verified:
  - error
  - minFieldCoverage: 80           # Percent of return type fields that must be asserted
    ignoreOptionalFields: true     # Skip fields marked as optional (?)
    ignoreFields: ["createdAt", "updatedAt"]  # Skip specific field names
    countToEqual: true              # toEqual({...}) with all fields counts as 100%
```

---

#### TQ-schema-conformance

**Purpose:** When a function returns a type/interface/struct, verify that tests assert the return value conforms to that schema — not just that it's "an object" or "not null."

This is deeper than `TQ-return-type-verified`: it checks that assertions verify the **types** of field values, not just their existence.

**What it catches:**

```typescript
// VIOLATION: Asserts values but not types — if createUser() returns
// { id: "1" } (string instead of number), this test still passes
const user = createUser(input);
expect(user.id).toBeTruthy();
expect(user.name).toBeTruthy();

// VIOLATION: Checks some types but uses loose equality
expect(user.id).toBeTruthy();  // 0 is falsy — would miss id=0 bug
expect(user.role).toBe("admin");  // Good value check, but no type guard

// OK: Proper type-aware assertions
expect(user).toEqual(expect.objectContaining({
  id: expect.any(Number),
  name: expect.any(String),
  email: expect.stringMatching(/@/),
  role: expect.stringMatching(/^(admin|user)$/),
}));

// ALSO OK: Direct value comparison (implicitly checks type)
expect(user.id).toBe(1);           // Strict equality proves it's a number
expect(user.name).toBe("Alice");   // Strict equality proves it's a string
```

```go
// VIOLATION: Reflection-only check
if reflect.TypeOf(result).Kind() != reflect.Struct {
    t.Fatal("not a struct")
}
// No field-level assertions follow

// OK: Type assertion + field checks
user, ok := result.(*User)
if !ok {
    t.Fatal("expected *User")
}
if user.ID != 1 { ... }
```

**Detection algorithm:**

1. For each function under test, resolve its return type schema
2. Find all assertions on the return value in test code
3. For each asserted field, determine if the assertion constrains the type:
   - `toBe(1)` → constrains to number (strict equality) ✓
   - `toBeTruthy()` → does NOT constrain type ✗
   - `toEqual({...})` → constrains if expected object has typed values ✓
   - `expect.any(Number)` → explicit type constraint ✓
   - `toBeInstanceOf(Date)` → explicit type constraint ✓
4. Compute "type coverage": (fields with type-constraining assertions / total fields)
5. Flag if type coverage < threshold

**Options:**
```yaml
TQ-schema-conformance:
  - error
  - minTypeCoverage: 70             # Percent of fields needing type-constraining assertions
    strictEquality: true             # Count .toBe() as type-constraining
    treatToEqualAsComplete: true     # toEqual({...}) with all fields = 100%
```

---

#### TQ-error-path-coverage

**Purpose:** Ensure that functions which can throw/return errors have tests for those error paths.

**What it catches:**

```typescript
// Source: src/api.ts
async function fetchUser(id: string): Promise<User> {
  if (!id) throw new ValidationError("id is required");
  const res = await fetch(`/users/${id}`);
  if (!res.ok) throw new APIError(`HTTP ${res.status}`);
  const data = await res.json();
  if (!data.id) throw new DataError("invalid response");
  return data;
}

// Test: tests/api.test.ts
// VIOLATION: Only tests happy path
it("fetches user", async () => {
  const user = await fetchUser("123");
  expect(user.id).toBe("123");
});
// Missing: no test for empty id, HTTP error, invalid response

// OK: All error paths covered
it("throws on empty id", async () => {
  await expect(fetchUser("")).rejects.toThrow(ValidationError);
});
it("throws on HTTP error", async () => {
  mockFetch(500);
  await expect(fetchUser("123")).rejects.toThrow(APIError);
});
it("throws on invalid response", async () => {
  mockFetch(200, {});
  await expect(fetchUser("123")).rejects.toThrow(DataError);
});
```

```go
// Source: api.go
func FetchUser(id string) (*User, error) {
    if id == "" {
        return nil, ErrInvalidID
    }
    resp, err := http.Get("/users/" + id)
    if err != nil {
        return nil, fmt.Errorf("fetch: %w", err)
    }
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    ...
}

// Test: api_test.go
// VIOLATION: Only tests success + one error
func TestFetchUser(t *testing.T) {
    user, err := FetchUser("123")
    if err != nil { t.Fatal(err) }
    if user.ID != "123" { t.Fail() }
}
func TestFetchUser_EmptyID(t *testing.T) {
    _, err := FetchUser("")
    if err == nil { t.Fatal("expected error") }
}
// Missing: HTTP error path, decode error path
```

**Detection algorithm:**

1. Parse source function AST
2. Identify all "error exit points":
   - `throw` statements (TS)
   - `return ..., err` / `return ..., fmt.Errorf(...)` (Go)
   - `reject()` / `Promise.reject()` (TS)
3. For each error exit point, extract the error condition (the `if` guard or input that triggers it)
4. Scan test file for test cases that trigger each error condition:
   - Match by: error class name, error message substring, sentinel error variable, or input values that match the guard condition
5. Compute "error path coverage": (error exits with matching tests / total error exits)
6. Flag if < threshold

**Options:**
```yaml
TQ-error-path-coverage:
  - error
  - minErrorPathCoverage: 100       # Require 100% error path coverage
    countGenericCatch: false         # .rejects.toThrow() without specific error doesn't count
    ignoreInternalHelpers: true      # Skip unexported/private helper functions
```

---

#### TQ-assertion-depth

**Purpose:** Verify that assertions go deep enough into nested structures. A test that checks `result.data` but never checks `result.data.items[0].name` is shallow.

**What it catches:**

```typescript
// Source returns: { data: { items: [{ id, name, price }], total: number } }

// VIOLATION: Only checks top-level
expect(result.data).toBeDefined();
expect(result.data.items).toHaveLength(2);
// Never checks what's IN the items

// OK: Asserts into nested structure
expect(result.data.items).toHaveLength(2);
expect(result.data.items[0].id).toBe(1);
expect(result.data.items[0].name).toBe("Widget");
expect(result.data.items[0].price).toBe(9.99);
expect(result.data.total).toBe(19.98);
```

**Detection algorithm:**

1. Resolve return type of function under test
2. Compute "type depth" — maximum nesting level of the return type (e.g., `{ data: { items: { name: string } } }` has depth 3)
3. Find maximum "assertion depth" — deepest property chain in any assertion on the return value
4. If (assertion depth / type depth) < threshold, flag
5. **Special handling for arrays:** Checking `items[0].name` counts as depth into the array element type

**Options:**
```yaml
TQ-assertion-depth:
  - error
  - minDepthRatio: 0.6              # Assertions must reach 60% of type nesting depth
    ignoreLeafPrimitives: false     # Even primitive leaves should be asserted
```

---

#### TQ-boundary-tested

**Purpose:** Ensure that functions with numeric, string length, or collection size parameters have tests at boundary values (0, 1, max, empty string, empty array, etc.).

**What it catches:**

```typescript
// Source: function paginate(items: Item[], page: number, pageSize: number)

// VIOLATION: Only tests middle-of-range values
it("paginates items", () => {
  const result = paginate(items, 2, 10);
  expect(result).toHaveLength(10);
});

// OK: Tests boundaries
it("returns empty for page 0", () => { ... });
it("returns empty for empty array", () => { ... });
it("handles page beyond total", () => { ... });
it("handles pageSize of 1", () => { ... });
it("handles pageSize of 0", () => { ... });
```

**Detection algorithm:**

1. Identify function parameters with numeric types, string types, or collection types
2. Define boundary values per type:
   - `number`: 0, 1, -1, MAX_SAFE_INTEGER (or configurable max)
   - `string`: "", single char, very long string
   - `array/slice`: [], single element, empty
3. Scan test cases for calls to the function with boundary values
4. Flag if any boundary value category is completely untested

**Options:**
```yaml
TQ-boundary-tested:
  - error
  - requireZero: true               # Must test 0 for numeric params
    requireEmpty: true              # Must test empty for string/array params
    requireNegative: true           # Must test negative for numeric params
    customBoundaries:               # Add custom boundary values
      pageSize: [1, 100]
```

---

#### TQ-mock-scope

**Purpose:** Ensure mocks are scoped correctly — no global mocks that leak between tests, no mocks of things that should be tested for real.

**What it catches:**

```typescript
// VIOLATION: Global mock at module level (leaks across all tests in file)
jest.mock("../database");
// All tests now use the mock — none test real database integration

// VIOLATION: Mock inside test but never restored
it("fetches data", () => {
  jest.spyOn(Date, "now").mockReturnValue(1000);
  // Missing: no mockRestore() or afterEach cleanup
});

// OK: Scoped mock with cleanup
beforeEach(() => {
  jest.spyOn(Date, "now").mockReturnValue(1000);
});
afterEach(() => {
  jest.restoreAllMocks();
});

// OK: Mock in test with inline restore
it("fetches data", () => {
  const spy = jest.spyOn(Date, "now").mockReturnValue(1000);
  // ... test ...
  spy.mockRestore();
});
```

**Detection algorithm:**

1. Find all mock/spy creation calls at different scopes (module, describe, it/test, before/afterEach)
2. For module-level mocks: flag as violation (suggest moving to beforeEach or per-test)
3. For per-test mocks: verify corresponding cleanup (mockRestore, jest.restoreAllMocks in afterEach, or explicit restore in same test)
4. For Go: detect test helpers that modify global state without `t.Cleanup()`

**Options:**
```yaml
TQ-mock-scope:
  - error
  - allowGlobalMocks: []             # Specific modules allowed as global mocks
    requireCleanup: true             # Every spy must have a restore
```

---

#### TQ-test-isolation

**Purpose:** Ensure tests don't depend on execution order or shared mutable state.

**What it catches:**

```typescript
// VIOLATION: Shared mutable state between tests
let counter = 0;
it("increments", () => { counter++; expect(counter).toBe(1); });
it("increments again", () => { counter++; expect(counter).toBe(2); }); // Depends on first test!

// VIOLATION: Test depends on side effect from previous test
it("creates user", async () => {
  await createUser({ name: "Alice" });
});
it("finds user", async () => {
  const user = await findUser("Alice"); // Depends on previous test's side effect
  expect(user.name).toBe("Alice");
});
```

**Detection algorithm:**

1. Identify variables declared outside `it()`/`test()`/`t.Run()` blocks
2. Track mutations of those variables inside test blocks
3. If a variable is mutated in one test and read in another, flag
4. Also detect: file system writes without cleanup, database inserts without rollback

**Options:**
```yaml
TQ-test-isolation:
  - error
  - allowSharedSetup: true           # Allow shared state set in beforeEach (not in tests)
    checkFileSystem: true            # Detect fs writes without cleanup
```

---

#### TQ-negative-cases

**Purpose:** Ensure that for every "success" test, there's at least one corresponding "failure" test.

**What it catches:**

```typescript
// VIOLATION: Only success tests, no failure tests
describe("createUser", () => {
  it("creates a user with valid input", () => { ... });
  it("creates a user with minimal input", () => { ... });
  // Missing: no test for invalid input, duplicate email, etc.
});

// OK: Both success and failure
describe("createUser", () => {
  it("creates a user with valid input", () => { ... });
  it("rejects empty name", () => { ... });
  it("rejects duplicate email", () => { ... });
});
```

**Detection algorithm:**

1. Group tests by the function they test (via test name parsing + call analysis)
2. Classify each test as "positive" (expects success) or "negative" (expects error/rejection/failure)
3. If a function has positive tests but zero negative tests, flag
4. **Heuristic for classification:**
   - Test name contains "should throw", "rejects", "fails", "invalid", "error" → negative
   - Test body contains `.rejects`, `.toThrow`, `expect(...).not.`, error class assertions → negative
   - Otherwise → positive

**Options:**
```yaml
TQ-negative-cases:
  - error
  - minNegativeRatio: 0.3           # At least 30% of tests should be negative
    perFunction: true               # Apply per function, not per file
```

---

#### TQ-test-naming

**Purpose:** Enforce consistent, descriptive test names.

**What it catches:**

```typescript
// VIOLATION: Vague name
it("works", () => { ... });
it("test 1", () => { ... });
it("should handle edge case", () => { ... });  // What edge case?

// OK: Descriptive name
it("should return 404 when user ID does not exist", () => { ... });
it("should truncate names longer than 255 characters", () => { ... });
```

**Detection algorithm:**

1. Extract all test names from `it()`, `test()`, `t.Run()`, `Test*` function names
2. Check against configurable pattern
3. Reject names shorter than threshold
4. Reject names that are too generic (configurable blocklist: "works", "test", "basic", "simple", "edge case" without specifics)

**Options:**
```yaml
TQ-test-naming:
  - error
  - pattern: "should {verb} when {condition}"   # or "describes behavior"
    minLength: 15                                # Minimum test name length
    blockWords: ["works", "basic", "simple", "test 1", "test 2"]
```

---

### 6.2 Architecture (ARCH)

These rules enforce structural constraints that prevent architectural decay.

---

#### ARCH-dependency-direction

**Purpose:** Enforce unidirectional dependency flow between architectural layers.

**Configuration:**

```yaml
ARCH-dependency-direction:
  - error
  - layers:
      - name: handler
        patterns: ["cmd/*/handler*.go", "src/routes/**"]
      - name: service
        patterns: ["internal/service/**", "src/services/**"]
      - name: repository
        patterns: ["internal/repo/**", "src/repositories/**"]
      - name: model
        patterns: ["internal/model/**", "src/models/**"]
    direction: top-down  # handler → service → repository → model (never reverse)
```

**Detection algorithm:**

1. Classify each file into a layer based on path patterns
2. Build import graph
3. For each import, check if it violates the layer direction
4. `handler → service` is OK. `service → handler` is a violation.

---

#### ARCH-import-boundary

**Purpose:** Enforce explicit module boundaries — certain directories cannot import from certain other directories.

**Configuration:**

```yaml
ARCH-import-boundary:
  - error
  - boundaries:
      - from: "src/content/**"
        deny: ["src/background/**"]
        reason: "Content scripts cannot import background modules (MV3 process isolation)"
      - from: "cmd/**"
        deny: ["internal/capture/**"]
        reason: "CLI entry points should go through service layer"
      - from: "src/**"
        deny: ["tests/**", "**/*.test.*"]
        reason: "Source code must not import from test files"
```

---

#### ARCH-no-circular-deps

**Purpose:** Detect and reject circular import dependencies.

**Detection:** Build full import graph, run Tarjan's algorithm for strongly connected components. Any SCC with >1 node is a cycle. Report the shortest cycle path.

---

#### ARCH-max-file-lines

**Purpose:** Enforce maximum file size.

**Options:**
```yaml
ARCH-max-file-lines:
  - error
  - max: 800
    excludeComments: true
    excludeBlankLines: true
    overrides:
      "**/*.test.*": 1200   # Test files can be longer
      "**/*_test.go": 1200
```

---

#### ARCH-layer-violation

**Purpose:** Detect when code at one architectural layer performs responsibilities belonging to another layer (e.g., a handler doing direct database queries).

**Detection:** Pattern-based. Define "forbidden patterns" per layer:

```yaml
ARCH-layer-violation:
  - error
  - layers:
      handler:
        patterns: ["src/routes/**", "cmd/*/handler*.go"]
        forbiddenCalls: ["sql.*", "db.*", "*.Query", "*.Exec"]
        reason: "Handlers must not make direct database calls"
      repository:
        patterns: ["src/repositories/**", "internal/repo/**"]
        forbiddenCalls: ["http.*", "fetch", "net/http.*"]
        reason: "Repositories must not make HTTP calls"
```

---

#### ARCH-module-boundary

**Purpose:** Enforce that modules (directories with an `index.ts` or Go package) are accessed only through their public API, not by reaching into internal files.

**What it catches:**

```typescript
// VIOLATION: Reaching into module internals
import { helperFn } from "../auth/internal/token-utils";

// OK: Using module's public API
import { validateToken } from "../auth";
```

**Detection:** If a directory has an `index.ts` (or `index.js`), imports from other modules must use the directory path (resolved to index), not import specific internal files.

---

### 6.3 Convention (CONV)

These rules enforce codebase-wide consistency.

---

#### CONV-file-naming

**Purpose:** Enforce consistent file naming conventions.

**Options:**
```yaml
CONV-file-naming:
  - error
  - style: kebab-case        # or: camelCase, PascalCase, snake_case
    overrides:
      "src/components/**": PascalCase
      "**/*.test.*": kebab-case
```

---

#### CONV-file-header

**Purpose:** Require a header comment at the top of every file.

**Options:**
```yaml
CONV-file-header:
  - error
  - patterns:
      "**/*.go": "// {filename} — {description}"
      "**/*.ts": "// {filename} — {description}"
    ignore: ["**/*.d.ts", "**/index.ts"]
```

`{filename}` is auto-populated from the actual filename. `{description}` must be a non-empty string — Stricture verifies it's present but doesn't judge content.

---

#### CONV-error-format

**Purpose:** Enforce a consistent error message format across the codebase.

**Options:**
```yaml
CONV-error-format:
  - error
  - pattern: "{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}"
    # Matches: "CreateUser: email already exists. Use a different email address."
    # Rejects: "error creating user" or "something went wrong"
    applyTo:
      - "fmt.Errorf"
      - "new Error"
      - "throw new .*Error"
    minSegments: 2   # At least operation + root cause
```

---

#### CONV-export-naming

**Purpose:** Enforce naming conventions for exported/public symbols.

**Options:**
```yaml
CONV-export-naming:
  - error
  - typescript:
      exportedFunctions: camelCase
      exportedClasses: PascalCase
      exportedConstants: UPPER_SNAKE_CASE
      exportedTypes: PascalCase
    go:
      exportedFunctions: PascalCase    # Go convention
      exportedTypes: PascalCase
```

---

#### CONV-test-file-location

**Purpose:** Enforce where test files live relative to source files.

**Options:**
```yaml
CONV-test-file-location:
  - error
  - strategy: colocated   # test file must be adjacent to source file
    # or: mirrored          # tests/foo/ mirrors src/foo/
    # or: subfolder          # src/foo/__tests__/foo.test.ts
    suffixes:
      typescript: [".test.ts", ".spec.ts"]
      go: ["_test.go"]
```

---

#### CONV-required-exports

**Purpose:** Ensure that certain files or directories have required exports (e.g., every feature module must export an `index.ts` with specific symbols).

**Options:**
```yaml
CONV-required-exports:
  - error
  - patterns:
      "src/features/*/index.ts":
        required: ["default"]         # Must have a default export
      "src/services/*/index.ts":
        required: ["create*Service"]  # Must export a factory function
```

---

### 6.4 Contract (CTR)

These rules enforce **dual-contract testing** — verifying that both sides of a protocol boundary (HTTP, WebSocket, IPC, message queue) agree on the shape of data they send and receive. This catches the most insidious class of bugs: code that compiles, passes its own tests, but breaks at integration time because the client sends `userId` and the server expects `user_id`.

#### How contract pair detection works

Stricture identifies protocol boundaries by analyzing:

1. **HTTP clients and servers in the same repo** — A function that calls `fetch("/api/users")` or `http.Get("/api/users")` is a client. A route handler registered at `/api/users` is a server. If both exist in the same codebase, they form a contract pair.

2. **Shared type references** — If a client function parses a response into type `UserResponse` and a server handler serializes type `UserResponse`, they share a contract via that type.

3. **Explicit annotation** — For cases where auto-detection fails:
   ```typescript
   // stricture-contract: server=src/routes/users.ts client=src/services/api-client.ts
   ```
   ```go
   // stricture-contract: server=cmd/server/handlers.go client=pkg/client/users.go
   ```

4. **OpenAPI/Swagger/Proto files** — If the repo contains API spec files, Stricture uses them as the source of truth and verifies both client and server code conform.

#### Contract Pair Model

```typescript
interface ContractPair {
  /** Unique ID for this contract */
  id: string;
  /** How this pair was detected */
  detectedVia: "route-matching" | "shared-type" | "annotation" | "api-spec";

  /** Server side */
  server: {
    file: string;
    handler: string;        // Function name
    method: string;         // HTTP method: GET, POST, etc.
    path: string;           // Route path: /api/users/:id
    requestType: TypeModel | null;   // Expected request body type
    responseType: TypeModel | null;  // Response body type
    statusCodes: number[];           // Status codes the handler can return
    line: number;
  };

  /** Client side */
  client: {
    file: string;
    caller: string;         // Function name
    method: string;         // HTTP method used
    path: string;           // URL path called
    requestType: TypeModel | null;   // Request body being sent
    responseType: TypeModel | null;  // Type the response is parsed into
    handledStatusCodes: number[];    // Status codes the client handles
    line: number;
  };
}
```

---

#### CTR-request-shape

**Purpose:** Verify that the request body type the client sends matches what the server expects.

**What it catches:**

```typescript
// Server: src/routes/users.ts
interface CreateUserRequest {
  name: string;
  email: string;
  role: "admin" | "user";
}
app.post("/api/users", (req, res) => {
  const body: CreateUserRequest = req.body;
  // ...
});

// Client: src/services/api-client.ts
// VIOLATION: Client sends "type" but server expects "role"
async function createUser(data: { name: string; email: string; type: string }) {
  return fetch("/api/users", {
    method: "POST",
    body: JSON.stringify(data),
  });
}
```

```go
// Server: cmd/server/handlers.go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}

// Client: pkg/client/users.go
type CreateUserPayload struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    // VIOLATION: Missing "role" field — server requires it
}
```

**Detection algorithm:**

1. Identify contract pairs via route/URL matching
2. Resolve request body types on both sides
3. Flatten both types to field sets: `{ name: type, required: bool }`
4. Compare:
   - **Missing fields:** Client doesn't send a field the server requires → ERROR
   - **Extra fields:** Client sends a field the server doesn't expect → WARN (may be silently ignored)
   - **Type mismatch:** Same field name but different types → ERROR
   - **Name mismatch:** Fields that look similar but differ (e.g., `userId` vs `user_id`) → ERROR with suggestion

**Options:**
```yaml
CTR-request-shape:
  - error
  - strictExtraFields: false        # Warn (not error) on extra client fields
    fuzzyNameMatch: true             # Flag userId vs user_id as potential mismatch
    ignoreOptionalFields: true       # Don't flag optional server fields missing from client
```

---

#### CTR-response-shape

**Purpose:** Verify that the response body type the server sends matches what the client expects to receive.

**What it catches:**

```typescript
// Server: src/routes/users.ts
// Returns: { id: number, name: string, email: string, created_at: string }
app.get("/api/users/:id", (req, res) => {
  const user = await db.getUser(req.params.id);
  res.json({
    id: user.id,
    name: user.name,
    email: user.email,
    created_at: user.createdAt.toISOString(),
  });
});

// Client: src/services/api-client.ts
interface UserResponse {
  id: number;
  name: string;
  email: string;
  createdAt: string;  // VIOLATION: Client expects "createdAt" but server sends "created_at"
  avatar: string;     // VIOLATION: Client expects "avatar" but server never sends it
}
```

**Detection algorithm:**

Same as CTR-request-shape but in reverse direction. Additionally:
- If the server uses `json.Marshal` (Go) or `JSON.stringify` (TS), resolve the actual field names from struct tags or object literals, not just the type field names
- Go `json:"field_name"` tags override the Go field name — compare against the tag value, not the field name

---

#### CTR-status-code-handling

**Purpose:** Verify that the client handles all status codes the server can return.

**What it catches:**

```typescript
// Server can return: 200, 400, 404, 500
app.get("/api/users/:id", (req, res) => {
  if (!req.params.id) return res.status(400).json({ error: "Missing ID" });
  const user = await db.getUser(req.params.id);
  if (!user) return res.status(404).json({ error: "Not found" });
  res.json(user);
  // Also: unhandled errors → 500
});

// Client: VIOLATION — only handles 200
async function getUser(id: string): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  return res.json();  // No status check! 400/404/500 silently parsed as "user"
}

// OK: Handles all status codes
async function getUser(id: string): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  if (res.status === 400) throw new ValidationError(await res.text());
  if (res.status === 404) throw new NotFoundError(id);
  if (!res.ok) throw new APIError(res.status);
  return res.json();
}
```

**Detection algorithm:**

1. On server side: find all `res.status(N)`, `w.WriteHeader(N)`, and implicit 200/500 paths
2. On client side: find all status code checks (`res.status === N`, `res.ok`, `resp.StatusCode`)
3. Compute: (status codes client handles / status codes server can return)
4. Flag unhandled status codes

**Options:**
```yaml
CTR-status-code-handling:
  - error
  - requireExplicit: true            # res.ok is not enough — must handle specific codes
    ignore5xx: false                 # Even 5xx must be handled (not silently swallowed)
```

---

#### CTR-shared-type-sync

**Purpose:** When client and server reference the same type name (e.g., `UserResponse`), verify the type definitions are identical or sourced from a shared package.

**What it catches:**

```typescript
// shared/types.ts — The source of truth
export interface User { id: number; name: string; email: string; }

// Server: VIOLATION — redefines User locally with different fields
interface User { id: number; name: string; }  // Missing email!

// Client: OK — imports from shared
import { User } from "../shared/types";
```

```go
// VIOLATION: Two packages define their own "User" struct with different fields
// pkg/server/types.go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// pkg/client/types.go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    // Missing: Email
}
```

**Detection algorithm:**

1. Find all type definitions with the same name across the codebase
2. If two types with the same name are used on opposite sides of a contract pair:
   - If they're the same import → OK
   - If they're locally defined → compare field sets
   - If fields differ → ERROR with diff
3. Suggest: extract to a shared package

**Options:**
```yaml
CTR-shared-type-sync:
  - error
  - requireSharedPackage: false      # If true, duplicate type names across packages always error
    ignoreTestFiles: true            # Test-local type redefinitions are OK
```

---

#### CTR-json-tag-match

**Purpose:** (Go-specific) Verify that JSON struct tags match across contract boundaries. This catches the extremely common bug where a Go struct field is `CreatedAt` with tag `json:"created_at"` on the server, but the client struct has `CreatedAt` with tag `json:"createdAt"`.

**What it catches:**

```go
// Server
type UserResponse struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`     // snake_case
}

// Client
type UserResponse struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`      // VIOLATION: camelCase — doesn't match server's snake_case
}
```

**Detection algorithm:**

1. For each Go struct involved in a contract pair, extract `json:"..."` tags
2. Build a map: `{ jsonTagName → GoFieldName }` for each struct
3. Compare the tag maps between server and client
4. Flag mismatches: same Go field name but different JSON tags

**Options:**
```yaml
CTR-json-tag-match:
  - error
  - convention: snake_case           # Enforce a single convention across all contract types
```

---

#### CTR-dual-test

**Purpose:** The integration-level rule. Verify that contract pairs have tests on **both** sides that exercise the same scenario. If the server has a test that returns a 404 for a missing user, the client must have a test that handles a 404 from that endpoint.

**What it catches:**

```typescript
// Server test: tests/routes/users.test.ts
it("returns 404 for unknown user", async () => {
  const res = await request(app).get("/api/users/unknown");
  expect(res.status).toBe(404);
});

// Client test: tests/services/api-client.test.ts
// VIOLATION: No test for 404 handling from /api/users/:id
it("fetches user successfully", async () => {
  mock("/api/users/123", { status: 200, body: mockUser });
  const user = await getUser("123");
  expect(user.name).toBe("Alice");
});
// Missing: test for getUser("unknown") → NotFoundError
```

**Detection algorithm:**

1. For each contract pair, collect test scenarios from both sides
2. Extract "scenario fingerprints": { endpoint, method, status code, error class }
3. For each server test scenario, check if a matching client test exists (and vice versa)
4. Flag unmatched scenarios with confidence score:
   - Exact match on endpoint + status code → 100% confidence
   - Match on endpoint but different status code → 80% confidence → flag
   - No client test references the endpoint at all → 100% confidence → flag
5. Only flag violations above `minConfidence` threshold

**Options:**
```yaml
CTR-dual-test:
  - error
  - minConfidence: 80               # Only flag matches above this confidence
    requireBothDirections: true      # Both server→client and client→server
    ignoreInternalEndpoints: false   # Check all endpoints, even internal
```

---

#### CTR-strictness-parity

**Purpose:** When a cross-service manifest exists, verify that both producer and consumer enforce the same constraints for every field in a contract. If the manifest says `amount` has range `[0.01, 999999.99]`, both the emitting service and the consuming service must validate that range.

**What it catches:**

```go
// Manifest declares: amount: { type: number, range: [0.01, 999999.99] }

// Producer (billing-service): OK — validates range before sending
func createInvoice(amount float64) error {
    if amount < 0.01 || amount > 999999.99 {
        return ErrInvalidAmount
    }
    // ... serialize and send
}

// Consumer (user-service): VIOLATION — no range validation
func handleInvoice(event InvoiceEvent) error {
    // Just uses the amount directly — no validation
    db.SaveInvoice(event.InvoiceID, event.Amount)
    return nil
}
```

```typescript
// Manifest declares: role: { type: enum, values: ["admin", "user", "viewer"] }

// Producer: OK — exhaustive check
function serializeUser(user: User) {
  if (!["admin", "user", "viewer"].includes(user.role)) {
    throw new ValidationError(`Invalid role: ${user.role}`);
  }
  return JSON.stringify(user);
}

// Consumer: VIOLATION — only handles 2 of 3 enum values
function handleUser(data: UserResponse) {
  switch (data.role) {
    case "admin": return renderAdmin(data);
    case "user": return renderUser(data);
    // MISSING: "viewer" — what happens when role is "viewer"?
  }
}
```

**Detection algorithm:**

1. Load manifest contract for the service
2. For each field with constraints (range, enum values, format, minLength, maxLength, etc.):
   a. Find the producer code that constructs/serializes this field
   b. Find the consumer code that receives/deserializes this field
   c. Check both sides for validation matching the constraint
3. Compare validation completeness:
   - **Range check:** Does code compare against manifest-declared bounds?
   - **Enum check:** Does code validate against the full set of allowed values?
   - **Format check:** Does code validate the format (email, uuid, iso8601)?
   - **Length check:** Does code enforce min/max length?
4. Flag any side that accepts/emits the field without matching constraint enforcement

**Options:**
```yaml
CTR-strictness-parity:
  - error
  - requireBothSides: true            # Both producer and consumer must validate
    allowTrustedInternal: false       # Even internal services must validate
    manifestPath: "stricture-manifest.yml"  # Override manifest location
```

---

#### CTR-manifest-conformance

**Purpose:** Verify that the actual code (types, handlers, clients) conforms to the cross-service manifest declarations. The manifest is the source of truth — code must match it.

**What it catches:**

```go
// Manifest says: response fields are { id, name, email, role, created_at }

// Server handler returns a struct with an extra field not in manifest
type UserResponse struct {
    ID            int       `json:"id"`
    Name          string    `json:"name"`
    Email         string    `json:"email"`
    Role          string    `json:"role"`
    CreatedAt     time.Time `json:"created_at"`
    InternalNotes string    `json:"internal_notes"` // VIOLATION: not in manifest — leaking internal data
}
```

```typescript
// Manifest says: request fields are { name, email, role }

// Client sends extra field not in manifest
async function createUser(data: {
  name: string;
  email: string;
  role: string;
  notes: string;  // VIOLATION: not in manifest — server will ignore or reject
}) {
  return fetch("/api/users", { method: "POST", body: JSON.stringify(data) });
}
```

**Detection algorithm:**

1. Load manifest contract for the service
2. For each endpoint/message in the contract:
   a. Find the corresponding handler/consumer in code
   b. Resolve the actual type being serialized/deserialized
   c. Compare actual fields against manifest fields
3. Flag:
   - **Extra fields** (in code but not in manifest) → ERROR (potential data leak or silent failure)
   - **Missing fields** (in manifest but not in code) → ERROR (contract violation)
   - **Type mismatches** (field exists but wrong type) → ERROR
   - **Constraint mismatches** (field exists, right type, but weaker constraints than manifest) → deferred to CTR-strictness-parity

**Options:**
```yaml
CTR-manifest-conformance:
  - error
  - strictExtraFields: true           # Error on fields in code not in manifest
    allowResponseSubset: false        # Server must return ALL manifest fields, not a subset
    manifestPath: "stricture-manifest.yml"
```

---

#### Contract Detection in Configuration

Users can help Stricture find contract pairs by declaring them in config:

```yaml
# .stricture.yml
contracts:
  # Explicit pairs
  pairs:
    - server: "src/routes/users.ts"
      client: "src/services/api-client.ts"
      endpoints: ["/api/users", "/api/users/:id"]

    - server: "cmd/server/handlers.go"
      client: "pkg/client/*.go"

  # API spec as source of truth
  specs:
    - path: "api/openapi.yaml"
      serverPaths: ["src/routes/**"]
      clientPaths: ["src/services/**"]

  # Route detection patterns (helps auto-detection)
  routePatterns:
    typescript:
      - "app.{method}(\"{path}\","           # Express
      - "router.{method}(\"{path}\","        # Express Router
      - "fastify.{method}(\"{path}\","       # Fastify
    go:
      - "mux.HandleFunc(\"{path}\","         # net/http
      - "r.{Method}(\"{path}\","             # Chi
      - "e.{Method}(\"{path}\","             # Echo
```

---

## 7. Language Adapters

Adapters parse source files into the `UnifiedFileModel` that rules consume. Each adapter is an npm package implementing the `LanguageAdapter` interface.

### 7.1 UnifiedFileModel

```typescript
interface UnifiedFileModel {
  /** Absolute file path */
  path: string;

  /** Detected language */
  language: "typescript" | "go" | string;

  /** Whether this file is a test file */
  isTestFile: boolean;

  /** Raw source text */
  source: string;

  /** Line count (excluding blank lines and comments if configured) */
  lineCount: number;

  /** Import statements */
  imports: ImportDeclaration[];

  /** Export statements (TS) or exported symbols (Go) */
  exports: ExportDeclaration[];

  /** All functions/methods defined in this file */
  functions: FunctionModel[];

  /** All type/interface/struct definitions */
  types: TypeModel[];

  /** All class definitions (TS) */
  classes: ClassModel[];

  /** If test file: test cases and their metadata */
  testCases: TestCaseModel[];

  /** If test file: which source files this test covers */
  testTargets: string[];

  /** Language-specific AST (opaque to rules, available for plugins) */
  rawAST: unknown;
}

interface FunctionModel {
  name: string;
  exported: boolean;
  async: boolean;
  parameters: ParameterModel[];
  returnType: TypeModel | null;
  /** Lines where errors are thrown/returned */
  errorExits: ErrorExitPoint[];
  startLine: number;
  endLine: number;
}

interface TypeModel {
  name: string;
  kind: "interface" | "type" | "struct" | "enum" | "class";
  fields: FieldModel[];
  exported: boolean;
}

interface FieldModel {
  name: string;
  type: string;           // Type as string (e.g., "string", "number", "User")
  optional: boolean;      // TypeScript optional (?)
  nested: TypeModel | null; // If this field is a complex type, its model
  depth: number;          // Nesting depth from root
}

interface TestCaseModel {
  name: string;
  kind: "positive" | "negative" | "unknown";
  /** Function(s) being tested (resolved from calls + imports) */
  targetFunctions: string[];
  assertions: AssertionModel[];
  mocks: MockModel[];
  startLine: number;
  endLine: number;
}

interface AssertionModel {
  /** The full assertion expression as source text */
  expression: string;
  /** The property chain being asserted (e.g., "result.data.items[0].name") */
  targetPath: string;
  /** Classification */
  kind: "shallow" | "value" | "type" | "structure" | "error" | "matcher";
  /** Maximum property depth reached */
  depth: number;
  /** Whether this constrains the type of the value */
  constrainsType: boolean;
  line: number;
}

interface ErrorExitPoint {
  line: number;
  /** The error class/sentinel/message pattern */
  errorIdentifier: string;
  /** The condition that triggers this error (if in an if/guard) */
  guardCondition: string | null;
}

interface MockModel {
  /** What is being mocked */
  target: string;
  /** Scope: module-level, describe-level, test-level */
  scope: "module" | "describe" | "test";
  /** Whether cleanup (restore) is present */
  hasCleanup: boolean;
  line: number;
}
```

### 7.2 TypeScript Adapter

**Package:** `@stricture/adapter-typescript`
**Parser:** [ts-morph](https://github.com/dsherret/ts-morph) (wraps TypeScript compiler API)

Capabilities:
- Full type resolution (follows type aliases, generics, intersections)
- Test framework detection: Jest, Vitest, Mocha, Node test runner
- Import/export analysis with path resolution
- Assertion pattern recognition for all major test frameworks
- Mock/spy detection for Jest, Vitest, Sinon

### 7.3 Go Adapter

**Package:** `@stricture/adapter-go`
**Parser:** Parse `go` AST output via child process. The adapter calls `go` toolchain to produce a JSON AST representation, then maps it to `UnifiedFileModel`.

Implementation strategy:
1. Ship a small Go binary (`stricture-go-parser`) that uses `go/parser` + `go/types` to produce JSON AST
2. The Node.js adapter spawns this binary and parses its JSON output
3. Avoids trying to parse Go in JavaScript

Capabilities:
- Full type resolution via `go/types`
- Test detection: `func Test*`, `t.Run()`, `t.Helper()`
- Import path analysis (Go module-aware)
- Error return analysis (multiple return values with `error` type)
- Assertion pattern recognition: `testing`, `testify/assert`, `testify/require`

### 7.4 Future Adapters

The adapter interface is designed for community contribution:

```typescript
interface LanguageAdapter {
  /** Language identifier */
  language: string;

  /** File extensions this adapter handles */
  extensions: string[];

  /** Parse a file into UnifiedFileModel */
  parse(filePath: string, source: string, config: AdapterConfig): Promise<UnifiedFileModel>;

  /** Resolve import paths to absolute file paths */
  resolveImport(importPath: string, fromFile: string): string | null;
}
```

---

## 8. Plugin System

### 8.1 Plugin Interface

Plugins are JavaScript/TypeScript files that export rule factories.

```typescript
// Example plugin: my-custom-rule.ts
import { defineRule } from "stricture";

export default defineRule({
  id: "CUSTOM-my-rule",
  category: "custom",
  severity: "error",
  description: "Ensures all API handlers log their entry",
  why: "Observability: every API call should be traceable in logs",

  // Which files this rule applies to
  match: {
    languages: ["typescript"],
    pathPatterns: ["src/routes/**"],
    isTestFile: false,
  },

  // Rule implementation
  check(file: UnifiedFileModel, context: ProjectContext): Violation[] {
    const violations: Violation[] = [];
    for (const fn of file.functions) {
      if (fn.exported && !hasLogStatement(fn, file.rawAST)) {
        violations.push({
          line: fn.startLine,
          column: 0,
          message: `Exported handler "${fn.name}" does not contain a log statement`,
          fix: `Add "logger.info('${fn.name} called', { ... })" at the start of the function`,
          suppress: `// stricture-disable-next-line CUSTOM-my-rule`,
        });
      }
    }
    return violations;
  },

  // Optional: auto-fix function
  fix(file: UnifiedFileModel, violation: Violation): FixResult | null {
    // Return null if fix is not safe
    return {
      range: { startLine: violation.line, startCol: 0, endLine: violation.line, endCol: 0 },
      replacement: `  logger.info("${violation.data?.fnName} called");\n`,
    };
  },
});
```

### 8.2 Plugin Loading

Plugins are referenced in `.stricture.yml`:

```yaml
plugins:
  - "./plugins/my-custom-rule.ts"
  - "@my-org/stricture-plugin-api-standards"
  - name: "./plugins/conditional-rule.ts"
    options:
      logPrefix: "api"
```

### 8.3 Plugin API

Plugins receive:

| Object | Contents |
|--------|----------|
| `file: UnifiedFileModel` | The parsed file being checked |
| `context: ProjectContext` | Full project state: all files, dependency graph, config |
| `helpers` | Utility functions: `findTestsFor(fn)`, `resolveType(name)`, `getImportChain(file)` |

---

## 9. CLI Interface

### 9.1 Commands

```
stricture [options] [paths...]         Lint files (default command)
stricture fix [options] [paths...]     Auto-fix violations
stricture audit [options]              Run cross-service strictness audit (see §13.7)
stricture trace <file> [options]       Validate runtime traces against manifest (see §13.8)
stricture init                         Create .stricture.yml with defaults
stricture list-rules                   Show all available rules with descriptions
stricture inspect <file>               Show parsed UnifiedFileModel for a file (debug)
```

### 9.2 Options

```
Filtering:
  --rule <id>              Run only this rule (can be repeated)
  --category <cat>         Run only rules in this category (TQ, ARCH, CONV)
  --severity <level>       Only report violations at this level or above (error, warn)

Targeting:
  [paths...]               Files/directories to lint (default: current directory)
  --changed                Only lint files changed in current git branch (vs main)
  --staged                 Only lint staged files (useful for pre-commit hook)
  --ext <ext>              Only lint files with this extension

Output:
  --format <fmt>           Output format: text (default), json, sarif, junit
  --output <file>          Write output to file (default: stdout)
  --color / --no-color     Force color on/off (default: auto-detect TTY)
  --quiet                  Only show errors, not warnings
  --verbose                Show rule timing and debug info

Fix:
  --fix                    Apply auto-fixes for all fixable violations
  --fix-dry-run            Show what --fix would change without modifying files

Config:
  --config <path>          Use a specific config file
  --no-config              Ignore .stricture.yml, use defaults only

Performance:
  --concurrency <n>        Max parallel file processing (default: CPU count)
  --cache                  Cache parsed ASTs between runs (default: on)
  --no-cache               Disable AST cache

Audit (stricture audit):
  --manifest <path>        Path to stricture-manifest.yml (default: auto-detect)
  --service <name>         Audit only this service (default: auto-detect from .stricture.yml)
  --remote                 Fetch other services' code for cross-validation
  --strictness <level>     Override minimum strictness level (minimal, basic, standard, strict, exhaustive)

Trace (stricture trace):
  <file>                   Trace file to validate (HAR, OpenTelemetry, or custom JSON)
  --manifest <path>        Path to stricture-manifest.yml
  --trace-format <fmt>     Trace format: har, otel, custom (default: auto-detect)
  --service <name>         Which service captured this trace
  --strict                 Fail on any deviation from manifest

Meta:
  --version                Show version
  --help                   Show help
```

### 9.3 Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No errors (warnings are OK) |
| 1 | One or more error-level violations |
| 2 | Configuration or parse error (invalid config, unparseable file) |

### 9.4 Example Output (Text Format)

```
$ npx stricture src/ tests/

  src/services/user.ts
    TQ-error-path-coverage  error  Function "createUser" has 3 error exits but only 1 is tested (33%)
      → Missing tests for: ValidationError at line 12, DuplicateError at line 18
      → Add test cases that trigger these error paths
      ℹ  Why: Every error path is a contract with callers. Untested errors are undocumented behavior.

  tests/services/user.test.ts
    TQ-no-shallow-assertions  error  Line 24: expect(result).toBeDefined() — shallow assertion on typed return value
      → Assert specific fields: expect(result.id).toBe(...), expect(result.name).toBe(...)
      → Suppress: // stricture-disable-next-line TQ-no-shallow-assertions
      ℹ  Why: Existence checks don't verify correctness. The function could return { garbage: true } and this passes.

    TQ-return-type-verified  error  Function "createUser" returns User (5 fields) but only 2/5 are asserted (40%)
      → Missing assertions for: email, role, createdAt
      ℹ  Why: Unasserted fields can silently break. If createUser stops setting email, no test catches it.

  src/routes/api.ts
    ARCH-layer-violation  error  Line 45: Handler directly calls db.Query() — bypass service layer
      → Move database call to a service function and call that instead
      ℹ  Why: Handlers should delegate to services. Direct DB access in handlers makes testing and refactoring harder.

  src/utils/helpers.ts
    CONV-file-header  error  Missing file header comment
      → Add: // helpers.ts — {description of this file's purpose}
      → Auto-fixable: run with --fix

─────────────────────────────
  5 errors, 0 warnings in 4 files (147 files checked, 1.2s)
```

---

## 10. Output & Reporting

### 10.1 Formats

| Format | Flag | Use Case |
|--------|------|----------|
| **text** | `--format text` (default) | Human-readable terminal output with colors |
| **json** | `--format json` | Machine-readable for custom tooling |
| **sarif** | `--format sarif` | GitHub Code Scanning, VS Code SARIF Viewer |
| **junit** | `--format junit` | CI systems (Jenkins, GitLab CI, CircleCI) |

### 10.2 JSON Schema

```json
{
  "version": "0.1.0",
  "timestamp": "2026-02-12T12:00:00Z",
  "summary": {
    "filesChecked": 147,
    "errors": 5,
    "warnings": 0,
    "fixable": 1,
    "elapsed_ms": 1200
  },
  "violations": [
    {
      "ruleId": "TQ-no-shallow-assertions",
      "severity": "error",
      "category": "test-quality",
      "file": "tests/services/user.test.ts",
      "line": 24,
      "column": 5,
      "message": "expect(result).toBeDefined() — shallow assertion on typed return value",
      "why": "Existence checks don't verify correctness.",
      "suggestion": "Assert specific fields: expect(result.id).toBe(...)",
      "suppress": "// stricture-disable-next-line TQ-no-shallow-assertions",
      "fixable": false,
      "context": {
        "functionUnderTest": "createUser",
        "returnType": "User",
        "fieldsAsserted": ["id", "name"],
        "fieldsMissing": ["email", "role", "createdAt"]
      }
    }
  ]
}
```

### 10.3 SARIF Output

Follows [SARIF 2.1.0 specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html). Includes:
- Tool metadata (name, version, rules with full descriptions)
- Results with physical location (file, line, column)
- Code flow information for cross-file violations (ARCH rules)
- Fix suggestions as SARIF `fix` objects

### 10.4 JUnit XML

Standard JUnit XML format. Each rule is a `<testsuite>`, each checked file is a `<testcase>`. Violations are `<failure>` elements.

---

## 11. Auto-Fix System

### 11.1 Fixable Rules

Not all violations can be auto-fixed. Rules declare whether they support fixing:

| Rule | Fixable | What it fixes |
|------|---------|---------------|
| CONV-file-header | Yes | Adds header comment |
| CONV-file-naming | Yes | Renames file |
| CONV-export-naming | Yes | Renames symbol + updates all references |
| TQ-mock-scope | Partial | Adds `afterEach(() => jest.restoreAllMocks())` |
| TQ-test-naming | No | Can't guess the right name |
| TQ-no-shallow-assertions | No | Can't guess the expected value |
| ARCH-* | No | Architectural changes require human decisions |

### 11.2 Fix Modes

```bash
# Show what would change (safe — no files modified)
npx stricture --fix-dry-run

# Apply all safe fixes
npx stricture --fix

# Fix only specific rules
npx stricture --fix --rule CONV-file-header
```

### 11.3 Fix Dry-Run Output

```
$ npx stricture --fix-dry-run

  src/utils/helpers.ts
    CONV-file-header  Would add:
      + // helpers.ts — Utility functions for string and date manipulation.

  tests/services/user.test.ts
    TQ-mock-scope  Would add after line 3:
      + afterEach(() => { jest.restoreAllMocks(); });

  2 fixes available (run with --fix to apply)
```

---

## 12. CI Integration

### 12.1 GitHub Actions

```yaml
# .github/workflows/stricture.yml
name: Stricture
on: [pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npx stricture --format sarif --output stricture.sarif --changed
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: stricture.sarif
        if: always()
```

### 12.2 Pre-commit Hook

```bash
# .husky/pre-commit
npx stricture --staged --quiet
```

### 12.3 GitLab CI

```yaml
stricture:
  script:
    - npx stricture --format junit --output stricture-junit.xml
  artifacts:
    reports:
      junit: stricture-junit.xml
```

---

## 13. Cross-Service Contracts & Stricture Manifest

### 13.1 The Problem

Modern applications are rarely monoliths. A typical system has:
- Multiple microservices communicating via HTTP, gRPC, or message queues
- Frontend clients consuming backend APIs
- Third-party integrations with agreed-upon contracts
- Multi-language services (Go backend, TypeScript frontend, Python ML service)

When service A changes its response shape from `{ user_id: number }` to `{ userId: string }`, the only thing that catches it is a production incident. Existing tools (OpenAPI, Pact, Buf) validate _schema compatibility_ but not **strictness parity** — whether both sides are equally precise about constraints.

Nothing answers the question: "Does the consumer validate the `amount` field with the same range the producer enforces?" Stricture does.

### 13.2 Stricture Manifest (`.stricture-manifest.yml`)

The manifest is a cross-service configuration file that lives in its own repo (or a shared infra repo). It declares:
- Which services exist
- How they communicate
- What contracts they share
- **How strict each contract must be**

```yaml
# stricture-manifest.yml — Cross-service API contract specification

manifest_version: "1.0"
name: "acme-platform"

# ── Service Declarations ──────────────────────────────────────

services:
  api-gateway:
    repo: "github.com/acme/api-gateway"
    language: typescript
    role: producer              # Produces APIs consumed by others
    stricture_config: ".stricture.yml"

  user-service:
    repo: "github.com/acme/user-service"
    language: go
    role: both                  # Produces AND consumes
    stricture_config: ".stricture.yml"

  billing-service:
    repo: "github.com/acme/billing-service"
    language: go
    role: consumer
    stricture_config: ".stricture.yml"

  web-frontend:
    repo: "github.com/acme/web-app"
    language: typescript
    role: consumer

# ── Contract Definitions ──────────────────────────────────────

contracts:
  - id: "user-api"
    producer: user-service
    consumers: [api-gateway, web-frontend]
    protocol: http
    spec: "specs/user-api.openapi.yaml"   # Optional: OpenAPI as source of truth

    endpoints:
      - path: "/api/users/:id"
        method: GET
        response:
          type: User
          fields:
            id:         { type: integer, range: [1, 2147483647], required: true }
            name:       { type: string, minLength: 1, maxLength: 255, required: true }
            email:      { type: string, format: email, required: true }
            role:       { type: enum, values: ["admin", "user", "viewer"], required: true }
            created_at: { type: string, format: iso8601, required: true }
        status_codes: [200, 400, 404, 500]
        error_shape:
          code:    { type: string, required: true }
          message: { type: string, required: true }

      - path: "/api/users"
        method: POST
        request:
          type: CreateUserRequest
          fields:
            name:  { type: string, minLength: 1, maxLength: 255, required: true }
            email: { type: string, format: email, required: true }
            role:  { type: enum, values: ["admin", "user", "viewer"], required: true }
        response:
          type: User                # Same as GET /api/users/:id
        status_codes: [201, 400, 409, 500]

  - id: "billing-events"
    producer: billing-service
    consumers: [user-service]
    protocol: message_queue
    queue: "billing.events"

    messages:
      - event: "invoice.created"
        fields:
          invoice_id: { type: string, format: uuid, required: true }
          user_id:    { type: integer, range: [1, 2147483647], required: true }
          amount:     { type: number, range: [0.01, 999999.99], precision: 2, required: true }
          currency:   { type: enum, values: ["USD", "EUR", "GBP"], required: true }

# ── Strictness Requirements ───────────────────────────────────

strictness:
  # Global minimum strictness level
  minimum: strict

  # Per-field requirements
  rules:
    # Every numeric field must declare a range
    numeric-range-required: true
    # Every string field must declare maxLength
    string-length-required: true
    # Every enum must exhaustively list values
    enum-exhaustive: true
    # Error responses must have a declared shape
    error-shape-required: true
    # All status codes must be explicitly listed
    status-codes-exhaustive: true
```

### 13.3 Per-Service Configuration

Each service's `.stricture.yml` references the manifest:

```yaml
# .stricture.yml in user-service repo

extends:
  - "@stricture/config-recommended"

# Reference the cross-service manifest
manifest:
  url: "github.com/acme/stricture-manifest"   # Git URL
  # -- OR --
  path: "../stricture-manifest/stricture-manifest.yml"  # Local path (monorepo)

  # Which service am I?
  service: user-service

  # Which contracts do I participate in?
  contracts:
    - id: "user-api"
      role: producer
      handler_paths: ["cmd/server/handlers/**"]
      type_paths: ["pkg/types/**"]

    - id: "billing-events"
      role: consumer
      handler_paths: ["pkg/billing/consumer/**"]
      type_paths: ["pkg/billing/types/**"]
```

### 13.4 Strictness Parity

This is Stricture's killer feature for cross-service contracts. **Strictness parity** means: if one side of a contract declares a constraint, the other side must enforce it with equal precision.

**Example: Range parity**

```yaml
# Manifest says: amount is [0.01, 999999.99]
amount: { type: number, range: [0.01, 999999.99], precision: 2 }
```

Stricture validates:
- **Producer (billing-service):** Does the code that constructs `amount` ensure it's within `[0.01, 999999.99]`? Is there validation before serialization?
- **Consumer (user-service):** Does the code that receives `amount` validate the range? What happens if it receives `0.00` or `1000000.00`? Is there explicit error handling for out-of-range values?

If the producer validates but the consumer just does `amount := body["amount"].(float64)` without range checking, that's a **strictness parity violation**. The consumer is less strict than the contract requires.

**Example: Enum parity**

```yaml
role: { type: enum, values: ["admin", "user", "viewer"] }
```

Stricture validates:
- **Producer:** Does the handler emit only these three values? Does validation reject unknown roles?
- **Consumer:** Does the consumer handle all three values? Is there a `default:` or `else` case? If the consumer only handles `"admin"` and `"user"` but not `"viewer"`, that's a parity violation.

**Example: Error handling parity**

```yaml
status_codes: [200, 400, 404, 500]
```

This extends CTR-status-code-handling to cross-service scope: the manifest is the source of truth, not just the server code. If the manifest says 404 is possible, the consumer _must_ handle it even if the current server implementation never returns it (the contract allows it, and a future deploy might).

### 13.5 Strictness Levels

Each field in the manifest has an implicit strictness level based on how precisely it's specified:

| Level | What's declared | Example |
|-------|----------------|---------|
| **Minimal** | Type only | `{ type: string }` |
| **Basic** | Type + required | `{ type: string, required: true }` |
| **Standard** | Type + required + format | `{ type: string, format: email, required: true }` |
| **Strict** | Type + required + format + constraints | `{ type: string, format: email, maxLength: 255, required: true }` |
| **Exhaustive** | All constraints + error cases + edge cases | Full specification with ranges, patterns, error shapes |

The `strictness.minimum` setting in the manifest enforces a floor. At `strict` level, any field declared as just `{ type: string }` is itself a violation — the manifest author must be more specific. This forces API designers to think precisely about their contracts before writing code.

### 13.6 Cross-Service Rules

Two new CTR rules operate specifically against the manifest (see section 6.4 for full details):

| Rule | Purpose |
|------|---------|
| **CTR-strictness-parity** | Both producer and consumer must enforce the same constraints for every manifest field |
| **CTR-manifest-conformance** | Actual code types/handlers must match manifest declarations (no extra fields, no missing fields, no type mismatches) |

These rules only activate when a `manifest` section exists in `.stricture.yml`. Without a manifest, they're no-ops.

### 13.7 The `stricture audit` Command

A dedicated CLI command for cross-service strictness analysis. Unlike `stricture lint` which runs rules, `stricture audit` provides a field-by-field strictness scorecard.

**What `stricture audit` does:**

1. Reads the manifest
2. Identifies which contracts this service participates in
3. For each contract:
   a. Validates types match manifest declarations (CTR-manifest-conformance)
   b. Validates strictness parity (CTR-strictness-parity)
   c. Reports strictness level per field
4. Generates a strictness scorecard

**Example output:**

```
$ stricture audit

Stricture Audit — user-service
════════════════════════════════

Contract: user-api (producer)
  Endpoint: GET /api/users/:id
    Response strictness: 85% (strict)
    ┌─────────────┬────────────┬───────────┬────────────────────────┐
    │ Field       │ Manifest   │ Code      │ Status                 │
    ├─────────────┼────────────┼───────────┼────────────────────────┤
    │ id          │ int [1,∞)  │ int       │ ⚠ Missing range check  │
    │ name        │ str [1,255]│ str       │ ⚠ Missing length check │
    │ email       │ str email  │ str email │ ✓ Strict               │
    │ role        │ enum(3)    │ enum(3)   │ ✓ Strict               │
    │ created_at  │ iso8601    │ time.Time │ ✓ Strict               │
    └─────────────┴────────────┴───────────┴────────────────────────┘

Contract: billing-events (consumer)
  Message: invoice.created
    Consumer strictness: 40% (basic)
    ┌─────────────┬──────────────┬────────────┬─────────────────────────┐
    │ Field       │ Manifest     │ Code       │ Status                  │
    ├─────────────┼──────────────┼────────────┼─────────────────────────┤
    │ invoice_id  │ uuid         │ string     │ ✗ No format validation  │
    │ user_id     │ int [1,∞)    │ int        │ ✗ No range validation   │
    │ amount      │ num [0.01,∞) │ float64    │ ✗ No range validation   │
    │ currency    │ enum(3)      │ string     │ ✗ No enum validation    │
    └─────────────┴──────────────┴────────────┴─────────────────────────┘

  Strictness Score: 62% (needs improvement)
  2 contracts, 9 fields, 4 strict, 2 basic, 3 missing validation
```

### 13.8 The `stricture trace` Command

A runtime validation tool that checks actual traffic against manifest contracts. This closes the gap between static analysis ("does the code look right?") and runtime reality ("is the actual traffic correct?").

**Supported trace formats:**

| Format | Source | What it captures |
|--------|--------|-----------------|
| **HAR** | Browser DevTools, Gasoline, Charles Proxy | HTTP request/response pairs with full bodies |
| **OpenTelemetry** | OTel collector export | Distributed traces with spans and attributes |
| **Custom JSON** | Application logging | `[{ method, path, request_body, response_body, status }]` |

**What `stricture trace` does:**

1. Reads a runtime trace file
2. Matches each request/response to a manifest contract endpoint
3. Validates actual field values against manifest constraints:
   - Did the response include all required fields?
   - Were field values within declared ranges?
   - Were enum values from the allowed set?
   - Were status codes from the declared set?
4. Reports deviations — the delta between what the manifest promises and what actually happened

**Example output:**

```
$ stricture trace captured-traffic.har

Trace Audit — 247 requests matched to contracts
════════════════════════════════════════════════

Contract: user-api
  GET /api/users/:id (142 requests)
    ✓ All responses had required fields
    ✗ 3 responses had role="moderator" — not in enum [admin, user, viewer]
    ✗ 1 response had name="" — violates minLength: 1
    ⚠ 12 responses had status 503 — not in declared status_codes [200, 400, 404, 500]

  POST /api/users (48 requests)
    ✓ All requests had required fields
    ✗ 2 requests sent role="mod" — not in enum
    ✓ All responses matched declared shape

Unmatched requests: 57 (no manifest contract)
  GET /healthz (30)
  GET /metrics (27)

Summary: 190 matched, 6 violations, 57 unmatched
```

**The three validation modes form a complete chain:**

```
Manifest (source of truth)
    ├── stricture audit   → validates CODE matches manifest (static, per-field scorecard)
    ├── stricture lint    → validates CODE quality + contracts (static, rule-based)
    └── stricture trace   → validates RUNTIME matches manifest (dynamic, trace-based)
```

### 13.9 Backwards Compatibility Detection

> **Status: Future Development (v0.3+)**

Planned feature: detect when a manifest change would break existing consumers. This includes:

- **Field removal detection:** If the manifest removes a required field, flag all consumers that depend on it
- **Type narrowing detection:** If `amount` range changes from `[0.01, 999999.99]` to `[1.00, 99999.99]`, flag consumers that may send/expect values in the old range
- **Enum value removal:** If `role` removes `"viewer"`, flag consumers handling that value
- **Status code removal:** If an endpoint stops returning 404, flag consumers with dead 404 handling code
- **Breaking vs. non-breaking classification:** Adding an optional field is non-breaking; removing a required field is breaking; narrowing a range is breaking for producers, non-breaking for consumers

This requires manifest versioning and diff analysis between versions. It will leverage the same field-level constraint model used by `stricture audit` and CTR-strictness-parity.

### 13.10 Prior Art & Differentiation

| Tool | What it does | What Stricture adds |
|------|-------------|---------------------|
| **OpenAPI / Swagger** | Schema definition language for HTTP APIs | Stricture validates that _code_ matches the schema AND that both sides enforce constraints equally. OpenAPI defines; Stricture verifies. |
| **Pact** | Consumer-driven contract testing (runtime) | Stricture is static analysis (runs without services). Pact requires running services to generate/verify contracts. |
| **Buf** | Protobuf linting + breaking change detection | Only covers protobuf/gRPC. Stricture covers HTTP, message queues, and any protocol. Also adds strictness parity. |
| **AsyncAPI** | Schema for event-driven APIs | Similar to OpenAPI but for async. Stricture validates code conformance + strictness parity against AsyncAPI specs. |
| **Spectral** | OpenAPI/AsyncAPI linting | Lints the _spec document_ for quality. Stricture lints the _code_ for conformance to the spec. Complementary tools. |

**What no tool does today:** Validate that both sides of a contract enforce constraints with equal precision. OpenAPI can say `amount: { minimum: 0.01, maximum: 999999.99 }` but nothing verifies that the consumer's code actually checks that range. Stricture does.

---

## 14. Performance Requirements

| Metric | Target | Rationale |
|--------|--------|-----------|
| **Cold start** | < 3s for 500 files | Acceptable for CI |
| **Cached run** | < 1s for 500 files | AST cache avoids re-parsing unchanged files |
| **Incremental (--changed)** | < 2s for 20 changed files | Fast enough for pre-commit |
| **Memory** | < 500MB for 10,000 files | Reasonable for CI runners |
| **Per-file** | < 50ms per file | With all rules enabled |

### Caching Strategy

- AST cache stored in `.stricture-cache/` (gitignored)
- Cache key: file content hash + adapter version
- Cache invalidation: file content change or adapter version bump
- ProjectContext (dependency graph) rebuilt on any file change in the project

---

## 15. Rule Reference (All Rules)

### Test Quality (TQ)

| Rule ID | Default | Fixable | Description |
|---------|---------|---------|-------------|
| TQ-no-shallow-assertions | error | No | Reject assertions that only check existence/truthiness |
| TQ-return-type-verified | error | No | All fields of return types must be asserted |
| TQ-schema-conformance | error | No | Assertions must constrain value types, not just existence |
| TQ-error-path-coverage | error | No | All error exits must have corresponding test cases |
| TQ-assertion-depth | error | No | Assertions must reach into nested structures |
| TQ-boundary-tested | error | No | Boundary values (0, empty, max) must be tested |
| TQ-mock-scope | error | Partial | Mocks must be scoped and cleaned up |
| TQ-test-isolation | error | No | Tests must not depend on shared mutable state |
| TQ-negative-cases | error | No | Every function needs both positive and negative tests |
| TQ-test-naming | error | No | Test names must be descriptive and follow pattern |

### Architecture (ARCH)

| Rule ID | Default | Fixable | Description |
|---------|---------|---------|-------------|
| ARCH-dependency-direction | error | No | Enforce unidirectional dependency flow between layers |
| ARCH-import-boundary | error | No | Block imports across module boundaries |
| ARCH-no-circular-deps | error | No | Reject circular import dependencies |
| ARCH-max-file-lines | error | No | Enforce maximum file size |
| ARCH-layer-violation | error | No | Detect cross-layer responsibility violations |
| ARCH-module-boundary | error | No | Enforce access through module public APIs only |

### Convention (CONV)

| Rule ID | Default | Fixable | Description |
|---------|---------|---------|-------------|
| CONV-file-naming | error | Yes | Enforce file naming convention (kebab-case, etc.) |
| CONV-file-header | error | Yes | Require file header comments |
| CONV-error-format | error | No | Enforce error message structure |
| CONV-export-naming | error | Yes | Enforce naming conventions for exports |
| CONV-test-file-location | error | Yes | Enforce test file placement strategy |
| CONV-required-exports | error | No | Enforce required exports from modules |

### Contract (CTR)

| Rule ID | Default | Fixable | Description |
|---------|---------|---------|-------------|
| CTR-request-shape | error | No | Client request body must match server's expected type |
| CTR-response-shape | error | No | Server response body must match client's expected type |
| CTR-status-code-handling | error | No | Client must handle all status codes server can return |
| CTR-shared-type-sync | error | No | Same-named types across contract boundaries must be identical |
| CTR-json-tag-match | error | No | Go JSON struct tags must match across contract pairs |
| CTR-dual-test | error | No | Both sides of a contract must have matching test scenarios |
| CTR-strictness-parity | error | No | Both producer and consumer must enforce same field constraints (requires manifest) |
| CTR-manifest-conformance | error | No | Code types/handlers must match manifest declarations (requires manifest) |

**Total: 34 rules** (10 TQ + 6 ARCH + 6 CONV + 8 CTR + 2 reserved for v0.2)

---

## 16. Glossary

| Term | Definition |
|------|------------|
| **Shallow assertion** | An assertion that checks existence, truthiness, or type without verifying the actual value. e.g., `toBeDefined()`, `toBeTruthy()`, `!= nil` without value check. |
| **Deep assertion** | An assertion that compares against a specific expected value, constraining both type and content. e.g., `toBe(42)`, `toEqual({id: 1, name: "Alice"})`. |
| **Error exit** | A code path in a function that terminates with an error: `throw`, `return err`, `reject()`. |
| **Field coverage** | The percentage of a return type's fields that are asserted in tests. |
| **Type coverage** | The percentage of a return type's fields whose assertions constrain the value's type. |
| **UnifiedFileModel** | Stricture's language-agnostic representation of a parsed source file. |
| **ProjectContext** | Cross-file analysis state including the full dependency graph and module map. |
| **Language adapter** | A parser plugin that converts language-specific source into UnifiedFileModel. |
| **Boundary value** | An input at the edge of valid ranges: 0, 1, -1, empty string, empty array, MAX_INT. |
| **Layer** | An architectural tier (e.g., handler, service, repository) with defined dependency rules. |
| **Module boundary** | The public API surface of a directory, typically defined by its `index.ts` or package exports. |
| **Stricture manifest** | A cross-service configuration file (`.stricture-manifest.yml`) that declares services, contracts, field constraints, and strictness requirements. Lives in a shared repo. |
| **Strictness parity** | The requirement that both sides of a contract enforce field constraints with equal precision. If the producer validates a range, the consumer must too. |
| **Strictness level** | How precisely a field is specified in the manifest: minimal (type only), basic (+required), standard (+format), strict (+constraints), exhaustive (all constraints + error cases). |
| **Strictness audit** | A per-field scorecard showing how well a service's code matches the precision declared in the manifest. Run via `stricture audit`. |
| **Runtime trace audit** | Validation of actual network traffic (HAR, OpenTelemetry) against manifest contract declarations. Run via `stricture trace`. |
| **Producer** | A service that emits data (serves an API endpoint, publishes a message). |
| **Consumer** | A service that receives data (calls an API endpoint, subscribes to a message queue). |

---

## Appendix A: Prior Art & Differentiation

| Tool | What it does | What Stricture adds |
|------|-------------|---------------------|
| ESLint | Syntax, style, simple patterns | Deep test quality analysis, architectural constraints |
| golangci-lint | Go-specific static analysis | Cross-file architectural rules, test quality for Go |
| Codacy | Aggregates existing linters | Original rules that no existing linter has |
| SonarQube | Code smells, duplication, coverage | Assertion quality (not just coverage %), schema conformance |
| ArchUnit (Java) | Architectural tests | Multi-language, not Java-only. Runs as lint, not as tests. |
| dependency-cruiser | JS/TS dependency validation | Also does Go. Also covers test quality + conventions. |
| Stryker (mutation testing) | Tests test quality via mutations | Complementary. Stricture is static analysis (fast); Stryker is dynamic (slow). |
| OpenAPI / Swagger | HTTP API schema definition | Stricture validates _code_ conforms to schema + strictness parity (see §13.10) |
| Pact | Consumer-driven contract testing (runtime) | Stricture is static analysis — no running services needed |
| Buf | Protobuf linting + breaking changes | Only protobuf/gRPC. Stricture covers HTTP, MQ, any protocol + strictness parity |
| AsyncAPI | Event-driven API schemas | Stricture validates code conformance + strictness parity against specs |
| Spectral | OpenAPI/AsyncAPI document linting | Lints the spec _document_. Stricture lints the _code_ against the spec. Complementary. |

## Appendix B: Open Questions for Tech Spec Phase

1. **Go adapter binary distribution:** Should `stricture-go-parser` be a pre-compiled binary per platform, or built from source on first run? Pre-compiled is faster but adds release complexity.

2. **Monorepo support:** Should `.stricture.yml` support workspace-level configs that inherit from root? (Likely yes, design in tech spec.)

3. **IDE integration:** VS Code extension with real-time linting? Defer to v0.2 or include in v0.1 scope?

4. **Baseline / legacy mode:** For existing codebases with thousands of violations, should there be a `--baseline` mode that only reports new violations? (Likely yes.)

5. **Custom test framework support:** How should users teach Stricture about non-standard assertion libraries? Plugin-level adapter, or config-level pattern matching?

6. **Backwards compatibility detection:** Manifest diff analysis to detect breaking changes is planned for v0.3+. See §13.9. Key question: should `stricture audit --diff v1.0..v1.1` compare manifest versions, or should this be a separate command? Also: how to handle "intentionally breaking" changes (new required field with migration path)?

7. **Manifest synchronization:** When a manifest lives in its own repo, how should services pin to a version? Git tags? Commit hashes? Should `stricture audit` warn if the local manifest is stale?

8. **Runtime trace volume:** `stricture trace` may need sampling for high-volume production traces. What sampling strategy? Statistical confidence thresholds?

9. **Strictness detection heuristics:** How to detect validation in code that uses custom validation libraries (e.g., `joi`, `zod`, `go-playground/validator`)? Plugin-level adapter or built-in pattern matching?
