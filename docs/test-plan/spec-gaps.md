# Spec Gap Analysis

Ambiguities, missing specifications, contradictions, and recommendations identified during test plan creation.

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
