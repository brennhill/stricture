# Stricture

![Stricture Splash](docs/assets/stricture-splash.svg)

> **v0.1-alpha:** Stricture is in active development. CLI shape, annotation profiles, and policy schema are still being refined.

## Stop API Drift Before Deploy

Stricture is a lineage-first drift control and architectural enforcement layer for API ecosystems.
It tells you **what changed, who is impacted, and whether deploy should proceed**.

## Why Teams Use It

- Detect drift traditional linters miss: field contracts, source versions, and cross-service invariants.
- Get actionable findings: cause service, impacted services, blast radius, and owner/escalation.
- Enforce policy in CI/CD with deterministic warn/block decisions.
- Reuse existing metadata via overlays (OpenAPI, OpenTelemetry, OpenLineage, AsyncAPI).

## How It Works (60 Seconds)

1. Add lightweight source comments where fields are emitted or transformed.
2. Keep service metadata in a systems registry (`docs/config-examples/lineage-systems.yml` pattern).
3. Export lineage, diff against baseline, and apply policy gates in CI.

Minimal annotation:

```go
// stricture-source: from PromotionsConfig
PromotionType string `json:"promotion_type"`
```

Namespace convention:

- Source comments use `stricture-` prefixes (for example `stricture-source`).
- YAML config keys use `stricture_` prefixes (for example `stricture_policy_url`).

## Quickstart

```bash
# Build CLI
make build

# Initialize repo config
./bin/strict init

# Run static checks
./bin/strict lint .

# Export current lineage snapshot
./bin/strict lineage-export --out .stricture/current-lineage.json .

# Compare against baseline and gate release
./bin/strict lineage-diff \
  --base .stricture/baseline-lineage.json \
  --head .stricture/current-lineage.json \
  --mode block
```

First run bootstrap:

```bash
mkdir -p .stricture
cp .stricture/current-lineage.json .stricture/baseline-lineage.json
```

## Common Commands

```bash
strict list-rules
strict explain --rule ARCH-dependency-direction
strict inspect-lineage path/to/file.go
strict lineage-escalate --service ServiceY --artifact .stricture/current-lineage.json --systems docs/config-examples/lineage-systems.yml
```

## Open Standard (SOS)

Stricture Open Standard defines portable lineage, drift, and policy semantics.

- Charter: [SPEC-CHARTER.md](SPEC-CHARTER.md)
- Draft spec: [spec/0.1-draft.md](spec/0.1-draft.md)
- License split: [LICENSES.md](LICENSES.md)

## Documentation

- Product spec: [docs/product-spec.md](docs/product-spec.md)
- Annotation reference: [docs/data-lineage-annotations.md](docs/data-lineage-annotations.md)
- Annotation quality: [docs/ANNOTATION-GUIDE.md](docs/ANNOTATION-GUIDE.md)
- Policy guide: [docs/POLICY-GUIDE.md](docs/POLICY-GUIDE.md)
- Helper draft spec: [docs/helper/SPEC.md](docs/helper/SPEC.md)
- Server spec: [docs/server/SPEC.md](docs/server/SPEC.md)
- Server storage/auth design: [docs/server/STORAGE.md](docs/server/STORAGE.md)
- Testing and quality gates: [TESTING.md](TESTING.md)
- Site deploy: [DEPLOY.md](DEPLOY.md)

## Website

- Home: https://stricture-lint.com/
- What is Stricture: https://stricture-lint.com/what-is-stricture
- Examples: https://stricture-lint.com/examples
- Live demo: https://stricture-lint.com/demo

## License

- Implementation (CLI, engine, adapters, scripts): [AGPL-3.0](LICENSE)
- Open Standard docs and schemas: [CC BY 4.0](LICENSES/CC-BY-4.0.md)
- Path-level mapping: [LICENSES.md](LICENSES.md)
