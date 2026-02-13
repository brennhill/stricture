# Stricture — Tree-sitter Version Pinning

> **Grammar versions are part of the correctness contract.**
>
> A tree-sitter grammar update can silently change how code is parsed, altering which violations are detected. Grammar versions must be pinned, tested, and upgraded deliberately.

---

## Pinned Versions

| Grammar | Module | Version | Last verified |
|---------|--------|---------|---------------|
| Go | `github.com/smacker/go-tree-sitter` | (native go/parser — no tree-sitter) | N/A |
| TypeScript | `github.com/smacker/go-tree-sitter/typescript` | `v0.0.0-20240827094217-dd81d9e9be82` | 2026-02-13 |
| Python | `github.com/smacker/go-tree-sitter/python` | `v0.0.0-20240827094217-dd81d9e9be82` | 2026-02-13 |
| Java | `github.com/smacker/go-tree-sitter/java` | `v0.0.0-20240827094217-dd81d9e9be82` | 2026-02-13 |

## Pinning Protocol

### When to pin

Pin the grammar version in `go.mod` after:
1. All adapter tests pass
2. All validation set fixtures for that language pass
3. Performance benchmarks are within targets

### When to upgrade

Upgrade a grammar version ONLY when:
1. A bug requires it (tree-sitter bug affecting our parsing)
2. A new language feature requires it (e.g., new syntax added to TypeScript)
3. Scheduled maintenance (quarterly review)

### Upgrade process

1. Update the version in `go.mod`
2. Run ALL tests: `make test`
3. Run ALL validation fixtures for the affected language
4. Compare violation counts before and after — any change must be investigated
5. Update the "Last verified" date in this table
6. Document any behavioral changes in the changelog

### Why this matters

Tree-sitter grammars are third-party code. A grammar change can:
- Rename AST node types → our queries stop matching → false negatives
- Add new node types → our queries match unexpected nodes → false positives
- Change tree structure → our depth calculations change → different assertion depths
- Fix grammar bugs → violations appear/disappear → users see "random" changes

By pinning, we guarantee that Stricture's behavior only changes when WE decide to change it.

---

## Query Version Tracking

Each adapter's `.scm` query files are tied to a specific grammar version. The query file header must declare which grammar version it was written for:

```scheme
;; queries/imports.scm
;; Grammar version: typescript@0.20.3
;; Last updated: 2026-03-15

(import_statement
  source: (string) @import_path)
```

When upgrading a grammar, all query files for that language must be reviewed and the version comment updated.
