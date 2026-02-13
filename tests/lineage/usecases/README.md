# Combined Real-World Use-Case Catalog

This folder contains the canonical 50-flow lineage use-case set used to stress
the top production scenarios in combination:

1. `drift_blocking`
2. `external_provider_drift`
3. `escalation_chain`
4. `compliance_traceability`
5. `multilang_contract_parity`

## Files

- `flows.json`: 50 API flow definitions across logistics, fintech, media,
  ecommerce, and governance.
- `systems.yml`: service and upstream contact metadata for escalation-chain tests.
- `baseline.json`: baseline artifact used by `usecase-agent` drift checks.
- `current.json`: latest generated artifact (updated by `usecase-agent`).

## Regeneration

Regenerate catalog + fixtures + fake API data:

```bash
go run ./scripts/generate-usecase-examples.go
```

Run full validation agent:

```bash
./scripts/usecase-agent.sh run
```
