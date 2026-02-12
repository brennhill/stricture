# Architecture (ARCH) Rules — Test Cases

6 rules covering architectural constraints.

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

