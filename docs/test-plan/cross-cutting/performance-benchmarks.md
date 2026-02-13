# Performance Benchmarks

Benchmark specifications for validating Stricture's performance requirements from the product spec.

---

## Performance Targets

| Benchmark | Target | Measurement |
|-----------|--------|-------------|
| Cold start (500 files) | < 3s | `time stricture` on a 500-file repo |
| Cached run (500 files) | < 1s | Second `time stricture` run |
| Incremental (20 files) | < 2s | `time stricture --changed` with 20 modifications |
| Per-file (single) | < 50ms | `time stricture single-file.ts` |
| Memory (10,000 files) | < 500MB | `node --max-old-space-size=500 stricture` |

---

## Benchmark Suite

### BM-01: Cold Start (500 files)

**Setup:** Generate a synthetic repo with 500 TypeScript files:
- 300 source files (avg 200 LOC each)
- 150 test files (avg 300 LOC each)
- 50 type definition files (avg 100 LOC each)
- 1 `.stricture-manifest.yml` with 10 contracts
- 1 `.stricture.yml` config with all rules enabled

**Run:**
```bash
rm -rf .stricture-cache/
time stricture --format json > /dev/null
```

**Target:** < 3 seconds on GitHub Actions `ubuntu-latest` (2-core)

**What's measured:** File discovery, parsing (ts-morph/go-parser), AST construction, all rule execution, output formatting.

### BM-02: Cached Run (500 files)

**Setup:** Same repo as BM-01, after first run (cache populated).

**Run:**
```bash
time stricture --format json > /dev/null
```

**Target:** < 1 second

**What's measured:** Cache hit rate, file modification check, rule execution on unchanged files (should be skipped).

### BM-03: Incremental Analysis (20 files changed)

**Setup:** Same 500-file repo. Modify 20 files (touch timestamp + add a comment line).

**Run:**
```bash
time stricture --changed --format json > /dev/null
```

**Target:** < 2 seconds

**What's measured:** Changed file detection, context loading, rule execution on 20 files only.

### BM-04: Single File Analysis

**Setup:** One TypeScript file at 800 LOC (max recommended size) with:
- 30 functions
- 10 types/interfaces
- Complex import graph (20 imports)
- 5 API client methods with fetch()

**Run:**
```bash
time stricture single-file.ts --format json > /dev/null
```

**Target:** < 50ms

**What's measured:** Single-file parsing, all rule execution, output formatting.

### BM-05: Memory Usage (10,000 files)

**Setup:** Generate a large synthetic repo with 10,000 files:
- 7,000 TypeScript source files
- 3,000 test files
- Realistic import graphs (each file imports 5-10 others)

**Run:**
```bash
node --max-old-space-size=500 dist/cli.js --format json > /dev/null
# If it exits with FATAL ERROR: HEAP out of memory, the test fails
```

**Target:** Peak memory < 500MB

**What's measured:** Memory efficiency of AST storage, dependency graph, violation collection.

### BM-06: Per-Rule Timing

**Setup:** Same 500-file repo as BM-01.

**Run:**
```bash
stricture --format json --timing > timing-report.json
```

**Target:** No single rule takes > 500ms on the 500-file repo.

**What's measured:** Rule execution time per rule category. Identifies slow rules for optimization.

### BM-07: Large File Stress Test

**Setup:** Single TypeScript file at 2,000 LOC (above recommended max):
- 100 functions
- 50 types
- Deeply nested objects (5 levels)
- 200 test assertions

**Run:**
```bash
time stricture large-file.ts --format json > /dev/null
```

**Target:** < 200ms (graceful degradation, not crash)

**What's measured:** Parser performance on oversized files, rule execution on large ASTs.

---

## Synthetic Repo Generator

Script to generate benchmark repos:

```bash
#!/usr/bin/env bash
# scripts/generate-benchmark-repo.sh
set -euo pipefail

SIZE="${1:-500}"  # Number of files
OUTPUT="${2:-/tmp/stricture-benchmark}"

mkdir -p "$OUTPUT/src" "$OUTPUT/tests"

echo "Generating $SIZE-file benchmark repo at $OUTPUT..."

for i in $(seq 1 "$SIZE"); do
    category=$((i % 3))
    case $category in
        0)
            # Source file
            file="$OUTPUT/src/module-${i}.ts"
            cat > "$file" << 'TSEOF'
// module-${i}.ts — Auto-generated benchmark file.

export interface Entity${i} {
  id: string;
  name: string;
  createdAt: Date;
  metadata: Record<string, string>;
}

export async function getEntity${i}(id: string): Promise<Entity${i}> {
  const response = await fetch(\`/api/entities/${i}/\${id}\`);
  if (!response.ok) {
    throw new Error(\`Failed to fetch entity: \${response.status}\`);
  }
  return response.json();
}

export async function createEntity${i}(input: Omit<Entity${i}, 'id' | 'createdAt'>): Promise<Entity${i}> {
  const response = await fetch('/api/entities/${i}', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }
  return response.json();
}
TSEOF
            ;;
        1)
            # Test file
            file="$OUTPUT/tests/module-${i}.test.ts"
            cat > "$file" << 'TSEOF'
// module-${i}.test.ts — Auto-generated benchmark test.
import { getEntity${i}, createEntity${i} } from '../src/module-${i}';

describe('Entity${i}', () => {
  it('should get entity by id', async () => {
    const entity = await getEntity${i}('test-id');
    expect(entity.id).toBe('test-id');
    expect(entity.name).toBeDefined();
    expect(entity.createdAt).toBeInstanceOf(Date);
  });

  it('should create entity', async () => {
    const input = { name: 'Test', metadata: {} };
    const entity = await createEntity${i}(input);
    expect(entity.id).toMatch(/^[a-f0-9-]+$/);
    expect(entity.name).toBe('Test');
  });

  it('should handle not found', async () => {
    await expect(getEntity${i}('nonexistent')).rejects.toThrow('Failed to fetch');
  });
});
TSEOF
            ;;
        2)
            # Type file
            file="$OUTPUT/src/types-${i}.ts"
            cat > "$file" << 'TSEOF'
// types-${i}.ts — Auto-generated type definitions.

export type Status${i} = 'active' | 'inactive' | 'pending' | 'archived';

export interface Config${i} {
  enabled: boolean;
  threshold: number;
  tags: string[];
  nested: {
    level1: {
      level2: {
        value: string;
      };
    };
  };
}
TSEOF
            ;;
    esac
done

# Generate manifest
cat > "$OUTPUT/.stricture-manifest.yml" << 'YAMLEOF'
manifest_version: "1.0"
contracts:
  - id: "benchmark-api"
    producer: benchmark-server
    consumers: [benchmark-client]
    protocol: http
    endpoints:
      - path: "/api/entities/:id"
        method: GET
        response:
          fields:
            id: { type: string, required: true }
            name: { type: string, required: true }
            createdAt: { type: string, format: iso8601, required: true }
YAMLEOF

echo "Generated $SIZE files in $OUTPUT"
```

---

## CI Integration

Performance benchmarks run on every PR that touches `src/rules/**` or `src/engine/**`. Results are posted as a PR comment with comparison to the `main` branch baseline.

```yaml
# In .github/workflows/validation-set.yml (performance-benchmark job)
- name: Compare with baseline
  run: |
    # Download baseline from main branch artifact
    # Compare and flag regressions > 20%
    if [ "$COLD_START_MS" -gt "$((BASELINE_COLD_START * 120 / 100))" ]; then
      echo "REGRESSION: Cold start $COLD_START_MS ms vs baseline $BASELINE_COLD_START ms"
      exit 1
    fi
```

---

## Tracking

Store benchmark results per release:

| Version | Cold Start | Cached | Incremental | Per-File | Memory |
|---------|-----------|--------|-------------|---------|--------|
| 0.1.0   | TBD       | TBD    | TBD         | TBD     | TBD    |

Update this table after each release with actual measurements.
