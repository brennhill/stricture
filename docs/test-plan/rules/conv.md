# Convention (CONV) Rules — Test Cases

6 rules covering codebase conventions.

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

