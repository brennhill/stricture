# Stricture — Go Error Handling Contract

> **Every error in Stricture must be wrapped, categorized, and actionable.**
>
> This document defines exactly how errors are created, wrapped, returned, and displayed. LLMs implementing Stricture must follow these rules — no exceptions.

---

## Error Format

All user-facing errors follow the Stricture error format:

```
{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}
```

Examples:
```
parse Go file: syntax error at line 42. Fix the syntax error and retry.
load config: .stricture.yml not found. Run 'stricture init' to create one.
resolve import: cannot find module "github.com/foo/bar". Ensure the module is in go.mod.
```

---

## Rules

### 1. Never ignore errors

```go
// BAD — errcheck will catch this
json.Unmarshal(data, &config)

// GOOD
if err := json.Unmarshal(data, &config); err != nil {
    return fmt.Errorf("parse config: %w", err)
}
```

The ONLY exceptions (configured in `.golangci.yml`):
- `fmt.Fprintf` / `fmt.Fprintln` to stdout/stderr — can't recover from write failure
- `io.Copy` in test helpers — test will fail anyway

### 2. Always wrap with context using `%w`

```go
// BAD — loses context
if err != nil {
    return err
}

// BAD — breaks error unwrapping
if err != nil {
    return fmt.Errorf("failed: %v", err)
}

// GOOD — wraps with context, preserves chain
if err != nil {
    return fmt.Errorf("parse %s: %w", path, err)
}
```

### 3. Wrap prefix = the operation being performed

The wrap prefix should be a short verb phrase describing what was being attempted:

```go
return fmt.Errorf("read file: %w", err)
return fmt.Errorf("parse Go AST: %w", err)
return fmt.Errorf("resolve import %q: %w", importPath, err)
return fmt.Errorf("check rule %s: %w", rule.ID(), err)
return fmt.Errorf("write JSON output: %w", err)
```

### 4. Use sentinel errors for expected failure modes

Define sentinel errors for conditions callers need to match on:

```go
// internal/model/errors.go

package model

import "errors"

var (
    // ErrUnsupportedLanguage is returned when a file's language has no adapter.
    ErrUnsupportedLanguage = errors.New("unsupported language")

    // ErrConfigNotFound is returned when no .stricture.yml exists.
    ErrConfigNotFound = errors.New("config not found")

    // ErrManifestNotFound is returned when no .stricture-manifest.yml exists.
    ErrManifestNotFound = errors.New("manifest not found")

    // ErrManifestInvalid is returned when the manifest fails validation.
    ErrManifestInvalid = errors.New("manifest invalid")

    // ErrParseFailure is returned when a source file cannot be parsed.
    ErrParseFailure = errors.New("parse failure")

    // ErrCacheCorrupt is returned when the cache file is unreadable.
    ErrCacheCorrupt = errors.New("cache corrupt")
)
```

Callers check with `errors.Is()`:

```go
if errors.Is(err, model.ErrUnsupportedLanguage) {
    // Skip this file silently
    continue
}
```

### 5. Never panic in library code

`panic` is only acceptable in two places:
1. `init()` functions for programmer errors (registering duplicate rules)
2. Test helpers

Everywhere else, return an error:

```go
// BAD
func (r *Registry) Register(rule Rule) {
    if rule == nil {
        panic("nil rule")  // Crashes the user's CI
    }
}

// GOOD
func (r *Registry) Register(rule Rule) error {
    if rule == nil {
        return fmt.Errorf("register rule: rule is nil")
    }
    r.rules = append(r.rules, rule)
    return nil
}
```

### 6. Log vs. return — never both

Either log the error and handle it, or return it. Never both.

```go
// BAD — error reported twice
if err != nil {
    log.Printf("failed to parse: %v", err)
    return fmt.Errorf("parse: %w", err)
}

// GOOD — return only (caller decides what to do)
if err != nil {
    return fmt.Errorf("parse %s: %w", path, err)
}
```

### 7. Errors in rules must not crash the run

If a single rule fails on a single file, report it as a diagnostic and continue:

```go
// In the engine loop:
for _, file := range files {
    for _, rule := range rules {
        violations, err := safeCheck(rule, file, context, config)
        if err != nil {
            // Report as diagnostic, don't abort
            diagnostics = append(diagnostics, Diagnostic{
                Rule:    rule.ID(),
                File:    file.Path,
                Message: err.Error(),
            })
            continue
        }
        allViolations = append(allViolations, violations...)
    }
}
```

### 8. Recover from panics in rules

Rules might panic on unexpected input. The engine must recover:

```go
func safeCheck(rule model.Rule, file *model.UnifiedFileModel, ctx *model.ProjectContext, cfg model.RuleConfig) (violations []model.Violation, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("rule %s panicked on %s: %v", rule.ID(), file.Path, r)
        }
    }()
    return rule.Check(file, ctx, cfg), nil
}
```

---

## Exit Codes

| Code | Meaning | When |
|------|---------|------|
| 0 | Success | No violations (or `--help` / `--version`) |
| 1 | Violations found | At least one error-severity violation |
| 2 | Operational error | Config parse failure, invalid flags, file not found |

**Rule:** Never exit with code 2 for user input that is technically valid but produces violations. Code 2 is for "Stricture itself couldn't run."

---

## Error Categories

For structured error reporting (JSON output), errors are categorized:

| Category | Examples |
|----------|---------|
| `config` | Missing config, invalid YAML, unknown rule ID |
| `parse` | Syntax error in source file, tree-sitter failure |
| `manifest` | Missing manifest, invalid contract, schema violation |
| `io` | File not found, permission denied, disk full |
| `internal` | Rule panic, cache corruption, unexpected state |
