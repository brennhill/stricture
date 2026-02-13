# Makefile — Build, test, and validate Stricture.
.PHONY: build test test-phase1 test-phase2 test-phase3 test-phase4 test-phase5 lint benchmark validate ci clean install

GOFLAGS ?=

build:
	go build $(GOFLAGS) -o bin/stricture ./cmd/stricture

test:
	go test $(GOFLAGS) ./...

test-phase1:
	go test $(GOFLAGS) ./internal/config/... ./internal/adapter/goparser/... ./internal/rules/conv/... ./internal/reporter/... ./cmd/stricture/...

test-phase2: test-phase1
	go test $(GOFLAGS) ./internal/adapter/typescript/... ./internal/rules/arch/... ./internal/engine/...

test-phase3: test-phase2
	go test $(GOFLAGS) ./internal/rules/tq/...

test-phase4: test-phase3
	go test $(GOFLAGS) ./internal/manifest/... ./internal/rules/ctr/...

test-phase5: test-phase4
	go test $(GOFLAGS) ./internal/adapter/python/... ./internal/adapter/java/...

lint:
	golangci-lint run ./...

benchmark:
	go test $(GOFLAGS) -bench=. -benchmem ./...

validate:
	./scripts/run-validation-set.sh

ci: lint test benchmark validate

clean:
	rm -rf bin/
	rm -rf .stricture-cache/

install:
	go install $(GOFLAGS) ./cmd/stricture
