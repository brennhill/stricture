# Testing And Quality Gates

This document is for contributors and CI operators. End-user CLI usage stays in
`README.md`.

## Common Local Commands

```bash
make build
make test
make lint
make test-coverage
```

## Phase Gates

Use these for staged implementation validation:

```bash
make test-phase1
make test-phase2
make test-phase3
make test-phase4
make test-phase5
make test-phase6
```

## High-Bar Validation

```bash
make quality-gate
make ci
```

`make quality-gate` runs build, tests, lineage checks, rule consistency, usecase
agent checks, and tool quality checks.

`make ci` runs lint + quality gate + benchmarks + validation set.

## Integration And Specialized Checks

```bash
make test-integration
make check-lineage
make check-fake-apis-live
make check-benchmarks
```

## TDD Workflow Helpers

```bash
make scaffold-rule RULE=TQ-error-path-coverage
make tdd-red RULE=CONV-error-format
make tdd-green RULE=CONV-error-format
```
