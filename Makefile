.DEFAULT_GOAL := help

GO ?= go
GOLANGCI_LINT ?= golangci-lint
DOCKER ?= docker
TEST_IMAGE ?= ubq-go-tests:local

.PHONY: help build run fmt lint vet tidy tidy-check check test test-race test-docker smoke smoke-docker

help:
	@echo make build        - Build the application into bin/
	@echo make run          - Display the scaffold usage
	@echo make fmt          - Apply Go formatting and organize imports
	@echo make lint         - Check all Go code including opt-in integration tests
	@echo make vet          - Run go vet including integration tests
	@echo make tidy         - Update module metadata
	@echo make tidy-check   - Check module metadata without changing it
	@echo make test         - Run ordinary Go tests without a router
	@echo make test-race    - Run ordinary tests with the race detector
	@echo make test-docker  - Build and run ordinary tests in Docker
	@echo make smoke        - Read-only smoke test against an explicit lab endpoint
	@echo make smoke-docker - Provision a disposable local VyOS lab and smoke-test it
	@echo make check        - Build, vet, lint, test, and check module metadata

build:
	$(GO) build -trimpath -o bin/ ./cmd/ubiquiti-config-generator

run:
	$(GO) run ./cmd/ubiquiti-config-generator

fmt:
	$(GOLANGCI_LINT) fmt ./...

lint:
	$(GOLANGCI_LINT) run --build-tags=integration ./...

vet:
	$(GO) vet -tags=integration ./...

tidy:
	$(GO) mod tidy

tidy-check:
	$(GO) mod tidy -diff

test:
	$(GO) test -count=1 ./...

test-race:
	$(GO) test -race -count=1 ./...

test-docker:
	$(DOCKER) build -f Dockerfile.test -t $(TEST_IMAGE) .
	$(DOCKER) run --rm $(TEST_IMAGE)

smoke:
	$(GO) test -tags=integration -count=1 -timeout=6m -v ./integration

smoke-docker:
	$(GO) test -tags=integration -count=1 -timeout=6m -v ./integration -args -docker-lab

check: build vet lint tidy-check test
