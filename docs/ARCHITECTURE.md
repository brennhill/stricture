# Architecture

The design target for Stricture's Go code, and how it is enforced. If this document and
`internal/archtest/policy.go` disagree, the policy file wins (it is what CI runs). Change both together.

## Principles

1. **Thin CLI over a library core.** `cmd/*/main.go` only calls into `internal/`. Behavior lives where it can be tested without `os.Args`.
2. **Dependencies point inward.** Everything depends on `internal/model`; `model` depends on nothing.
3. **Extension by addition.** A new rule, language, reporter or command is a new file or package plus one registration line. It never requires editing the CLI or another rule.
4. **Small files, simple functions.** No non-test Go file exceeds 800 lines and no function exceeds cyclomatic complexity 15.

## Layers (top may import below it)

```
cmd/*                 entry points: each calls only its own front end (cmd/strict -> internal/cli)
internal/cli          presentation: command routing, flags, usage text, color, exit codes
internal/app|engine   use cases: discover -> parse -> build context -> run rules -> filter -> report
internal/{adapter, rules, report, discovery, baseline, config, manifest, suppression, fix, lineage, plugins, server}
                      leaves: each depends on internal/model only (plus noted exceptions)
internal/model        interfaces and data types (UnifiedFileModel, Rule, Violation, errors)
```

| Package | May import (internal) |
|---|---|
| `internal/model`, `pkg/*`, `internal/archtest` | nothing |
| `internal/adapter/*` | `model`, `adapter` |
| `internal/rules`, `internal/rules/<category>` | `model`, `rules` |
| `internal/rules/all` | `model`, `rules/...` (the only place that wires every rule) |
| `internal/report/*`, `reporter`, `config`, `baseline`, `manifest`, `suppression`, `fix`, `discovery`, `testutil` | `model` |
| `internal/lineage/*`, `internal/server/*` | `model`, own family |
| `internal/plugins` | `model`, `pkg/rule` |
| `internal/engine` (rule selection and running, file loading, severity filtering), `internal/app` | leaves above (never `cli`, never `cmd`) |
| `internal/cli/*` | any `internal/...` |
| `cmd/strict`, `cmd/stricture-server`, `cmd/demo-pack` | only `internal/cli`, `internal/server`, `internal/lineage` respectively |

Consequences: rule categories cannot import each other; adapters cannot import rules; nothing below `cli` can parse flags (enforced: only `internal/cli` and `cmd/demo-pack` may import `flag`); nothing imports `cmd`.

## Extending

- **New rule:** one file `internal/rules/<cat>/<name>.go` + `_test.go`, then add one line to that package's `register.go`. `rules/all` already pulls the package in. Follow the gates in `docs/DEVELOPMENT-GATES.md`.
- **New language:** new package `internal/adapter/<lang>` implementing the adapter interface and registering by file extension.
- **New output format:** new package under `internal/report/`.
- **New command:** new file under `internal/cli/`.
- **New package:** add it to `Layers` in `internal/archtest/policy.go` and to the table above. An unclassified package fails the build on purpose.

## Enforcement

`make check-architecture` runs `internal/archtest`, which parses the imports of every non-test Go file under `cmd/`, `internal/` and `pkg/`. It also runs under `make test`, `make test-race`, `make quick-check`, and `make quality-gate`/CI. It fails on:

| Check | Fails when |
|---|---|
| Layering | a package imports something its row in `Layers` does not allow, or the package has no row |
| Rule registration | anything other than `internal/rules/all` imports an individual rule package |
| File size | a file exceeds 800 lines |
| Complexity | a function exceeds cyclomatic complexity 15 |

### Ratchets (shrink-only exceptions)

`policy.go` has three escape hatches for debt that cannot be fixed in one change: `LineCeilings`, `ComplexityCeilings` and `TransitionalRuleImporters`. A ceiling pins a file or function at its current size so it can never grow; shrinking it fails until the ceiling is lowered, and fitting the limit fails until the entry is deleted.

**All three are empty today. Never add an entry.** Split the code instead.

### How complexity is counted

1 + each `if`, `for`, `range`, non-default `case`, non-default `select` clause, `&&` and `||`. Function literals count toward the function that contains them. Test files are excluded from the file-size and complexity checks.

### Not enforced by `.stricture.yml`

Stricture's own `ARCH-*` rules are currently trigger-token stubs and do not analyze real code, so `.stricture.yml` cannot enforce this design. Once they analyze imports for real, migrate the layer table into config and keep `archtest` as the backstop.
