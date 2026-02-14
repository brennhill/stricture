# Makefile — Build, test, and validate Stricture.
.PHONY: build build-server test test-race test-coverage test-phase1 test-phase2 test-phase3 test-phase4 test-phase5 test-phase6 test-integration test-tool-quality test-server lint benchmark validate ci quality-gate check-rules check-stubs check-invariants check-shell-syntax check-tree-sitter-pinning check-usecase-examples check-fake-apis-live check-benchmarks lineage-export lineage-diff check-lineage update-lineage-baseline phase-agent phase-agent-status phase-agent-reset overseer-agent overseer-agent-once overseer-agent-status overseer-agent-reset usecase-agent usecase-agent-status usecase-agent-reset publication-agent publication-agent-status publication-agent-reset spec-quality-audit clean install scaffold-rule tdd-red tdd-green validate-gates progress progress-test progress-json check-messages add-regression validate-all quick-check site-install site-demo-pack site-build site-dev site-worker-dev site-worker-deploy

GOFLAGS ?=
GOCACHE ?= $(PWD)/.cache/go-build
export GOCACHE
LINEAGE_MODE ?= block
TEST_PKGS := ./cmd/... ./internal/...
BENCH_PKGS := ./cmd/... ./internal/...
PHASE1_PKGS := ./internal/config/... ./internal/adapter/goparser/... ./internal/rules/conv/... ./internal/reporter/... ./cmd/stricture/...

build:
	go build $(GOFLAGS) -o bin/stricture ./cmd/stricture

build-server:
	go build $(GOFLAGS) -o bin/stricture-server ./cmd/stricture-server

test:
	go test $(GOFLAGS) $(TEST_PKGS)

test-server:
	go test $(GOFLAGS) ./internal/server/... ./cmd/stricture-server/...

test-race:
	go test $(GOFLAGS) -race -count=1 -timeout=300s $(TEST_PKGS)

test-coverage:
	go test $(GOFLAGS) -coverprofile=coverage.out -covermode=atomic $(PHASE1_PKGS)

test-phase1:
	go test $(GOFLAGS) $(PHASE1_PKGS)

test-phase2: test-phase1
	go test $(GOFLAGS) ./internal/adapter/typescript/... ./internal/rules/arch/... ./internal/engine/...

test-phase3: test-phase2
	go test $(GOFLAGS) ./internal/rules/tq/...

test-phase4: test-phase3
	go test $(GOFLAGS) ./internal/manifest/... ./internal/rules/ctr/...

test-phase5: test-phase4
	go test $(GOFLAGS) ./internal/adapter/python/... ./internal/adapter/java/...

test-phase6: test-phase5
	go test $(GOFLAGS) ./internal/fix/... ./internal/plugins/... ./internal/suppression/...

test-integration:
	mkdir -p .cache/go-build
	GOCACHE=$(PWD)/.cache/go-build go test $(GOFLAGS) -tags=integration -count=1 -timeout=180s ./tests/integration/...

test-tool-quality:
	./scripts/check-tool-quality.sh

lint:
	mkdir -p .cache/go-build .cache/golangci-lint
	GOCACHE=$(PWD)/.cache/go-build GOLANGCI_LINT_CACHE=$(PWD)/.cache/golangci-lint golangci-lint run ./cmd/... ./internal/...

benchmark:
	go test $(GOFLAGS) -bench=. -benchmem $(BENCH_PKGS)

validate:
	VALIDATION_SET_ENFORCE_SKIP_POLICY=1 \
	VALIDATION_SET_ALLOWED_SKIPS="50-convention-patterns/PERFECT,61-event-driven/PERFECT,72-framework-patterns-java/PERFECT" \
	VALIDATION_SET_MAX_SKIPS=3 \
	VALIDATION_SET_REQUIRE_ALLOWED_SKIPS=1 \
	./scripts/run-validation-set.sh

ci: lint quality-gate benchmark validate

quality-gate:
	$(MAKE) build
	$(MAKE) test-phase6
	$(MAKE) test-integration
	./scripts/check-lineage-drift.sh
	./scripts/validate-gate.sh --phase 1
	./scripts/validate-error-messages.sh
	./scripts/check-rule-consistency.sh
	./scripts/check-no-stubs.sh --phase 6
	./scripts/check-invariant-tests.sh
	./scripts/check-bash-syntax.sh
	./scripts/check-tree-sitter-pinning.sh
	./scripts/usecase-agent.sh run
	./scripts/check-tool-quality.sh
	VALIDATION_HEALTH_FAIL_ON_WARNINGS=1 ./scripts/validation-health-check.sh

check-rules:
	./scripts/check-rule-consistency.sh

check-stubs:
	./scripts/check-no-stubs.sh --phase 6

check-invariants:
	./scripts/check-invariant-tests.sh

check-shell-syntax:
	./scripts/check-bash-syntax.sh

check-tree-sitter-pinning:
	./scripts/check-tree-sitter-pinning.sh

check-usecase-examples:
	./scripts/usecase-agent.sh run

check-fake-apis-live:
	./scripts/check-fake-apis-live.sh

check-benchmarks:
	./scripts/check-benchmark-regression.sh

lineage-export:
	./scripts/export-lineage-artifact.sh

lineage-diff:
ifndef BASE
	$(error BASE is required. Usage: make lineage-diff BASE=tests/lineage/baseline.json HEAD=tests/lineage/current.json)
endif
ifndef HEAD
	$(error HEAD is required. Usage: make lineage-diff BASE=tests/lineage/baseline.json HEAD=tests/lineage/current.json)
endif
	go run ./cmd/stricture lineage-diff --base $(BASE) --head $(HEAD) --mode $(LINEAGE_MODE)

check-lineage:
	./scripts/check-lineage-drift.sh

update-lineage-baseline:
	./scripts/update-lineage-baseline.sh

phase-agent:
	./scripts/phase-agent.sh run

phase-agent-status:
	./scripts/phase-agent.sh status

phase-agent-reset:
	./scripts/phase-agent.sh reset

overseer-agent:
	./scripts/overseer-agent.sh run

overseer-agent-once:
	./scripts/overseer-agent.sh once

overseer-agent-status:
	./scripts/overseer-agent.sh status

overseer-agent-reset:
	./scripts/overseer-agent.sh reset

usecase-agent:
	./scripts/usecase-agent.sh run

usecase-agent-status:
	./scripts/usecase-agent.sh status

usecase-agent-reset:
	./scripts/usecase-agent.sh reset

publication-agent:
	./scripts/publication-agent.sh run

publication-agent-status:
	./scripts/publication-agent.sh status

publication-agent-reset:
	./scripts/publication-agent.sh reset

spec-quality-audit:
	./scripts/spec-quality-audit.sh

clean:
	rm -rf bin/
	rm -rf .stricture-cache/

install:
	go install $(GOFLAGS) ./cmd/stricture

# --- Developer workflow targets ---

# Scaffold a new rule: make scaffold-rule RULE=TQ-error-path-coverage
scaffold-rule:
ifndef RULE
	$(error RULE is required. Usage: make scaffold-rule RULE=TQ-error-path-coverage)
endif
	./scripts/scaffold-rule.sh $(RULE)

# Enforce TDD red phase for a single rule (tests must fail).
tdd-red:
ifndef RULE
	$(error RULE is required. Usage: make tdd-red RULE=CONV-error-format)
endif
	./scripts/tdd-rule.sh --rule $(RULE) --stage red

# Enforce TDD green phase for a single rule (tests must pass).
tdd-green:
ifndef RULE
	$(error RULE is required. Usage: make tdd-green RULE=CONV-error-format)
endif
	./scripts/tdd-rule.sh --rule $(RULE) --stage green

# Validate development gate prerequisites
validate-gates:
	./scripts/validate-gate.sh --all

# Show rule implementation progress
progress:
	./scripts/track-progress.sh

# Show progress with test results (slower)
progress-test:
	./scripts/track-progress.sh --test

# Show progress as JSON
progress-json:
	./scripts/track-progress.sh --json

# Validate error messages match catalog
check-messages:
	./scripts/validate-error-messages.sh

# Add a regression fixture: make add-regression RULE=CONV-file-naming DESC="missed kebab-case"
add-regression:
ifndef RULE
	$(error RULE is required. Usage: make add-regression RULE=CONV-file-naming DESC="description")
endif
ifndef DESC
	$(error DESC is required. Usage: make add-regression RULE=CONV-file-naming DESC="description")
endif
	./scripts/add-regression.sh $(RULE) "$(DESC)"

# Run all validations (gates + messages + health check)
validate-all: validate-gates check-messages
	VALIDATION_HEALTH_FAIL_ON_WARNINGS=1 ./scripts/validation-health-check.sh

# Quick check: lint + test phase 1 + validate messages
quick-check: lint test-phase1 check-messages

# --- Marketing site + demo targets ---

site-install:
	cd site && npm install

site-demo-pack:
	go run ./cmd/demo-pack

site-build: site-demo-pack
	cd site && npm run build

site-dev:
	cd site && npm run dev

site-worker-dev: site-build
	cd site && npm run worker:dev

site-worker-deploy: site-build
	cd site && npm run worker:deploy
