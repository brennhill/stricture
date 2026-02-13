# Stress Test Corpus

Edge-case files that test parser robustness. Each file targets a specific parser challenge.

| File | Challenge | Expected behavior |
|------|-----------|-------------------|
| deeply-nested.ts | 10-level nested types + optional chaining | Parse without stack overflow, correct depth counting |
| unicode-identifiers.go | Unicode in identifiers and strings | Parse correctly, extract names with Unicode chars |
| many-functions.ts | 100 exported functions | Parse in < 200ms, correct function count |
| complex-generics.ts | Complex TypeScript generics | Parse without crashing, extract type names |
| empty-constructs.py | Empty/minimal Python constructs | No crashes on edge-case constructs |

## Usage

These fixtures are used by:
1. Fuzz test seed corpus (seeded into fuzz tests)
2. Parser performance benchmarks (many-functions.ts)
3. Adapter correctness tests (all files)
