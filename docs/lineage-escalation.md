# Lineage Escalation Chains

When a service reports bad data, Stricture can resolve who to call by walking
upstream dependencies from lineage artifacts.

## Inputs

- Lineage artifact JSON: produced by `stricture lineage-export`.
- Optional system registry YAML: maps system IDs to emergency contacts.

Registry format is defined in:

- `docs/schemas/lineage-system-registry.schema.json`
- example: `docs/config-examples/lineage-systems.yml`

Registry may also include flow catalog + membership:

- `'strict:flows'`: canonical flow IDs and numeric levels
- `systems[].flows`: service membership in flow IDs

By default, registry escalation contacts are strongly recommended but not
globally required; organizations can enforce stricter requirements via policy.

System IDs may use one optional `:` suffix for internal subsystems (for
example `location-tracking-service:ingestion`). Escalation can be resolved at
either subsystem or top-level service granularity depending on registry data.

## Command

```bash
stricture lineage-escalate \
  --service ServiceY \
  --artifact tests/lineage/current.json \
  --systems docs/config-examples/lineage-systems.yml \
  --max-depth 5
```

## Output

JSON array of escalation steps in upstream order (closest service first), each
including:

- `system_id`
- `owner`
- `contacts`
- `depth`
- `reason`
- `flows` (when registry includes service flow memberships)

If a system is missing in the registry, Stricture falls back to annotation
fields (`owner`, `escalation`) where available.
