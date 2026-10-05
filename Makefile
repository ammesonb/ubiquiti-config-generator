.DEFAULT_GOAL := help

GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: help build run fmt lint vet tidy tidy-check check

help:
	@echo make build       - Build the application into bin/
	@echo make run         - Display the scaffold usage
	@echo make fmt         - Apply Go formatting and organize imports
	@echo make lint        - Check formatting, errors, and static analysis
	@echo make vet         - Run go vet
	@echo make tidy        - Update module metadata
	@echo make tidy-check  - Check module metadata without changing it
	@echo make check       - Build, vet, lint, and check module metadata

build:
	$(GO) build -trimpath -o bin/ ./cmd/ubiquiti-config-generator

run:
	$(GO) run ./cmd/ubiquiti-config-generator

fmt:
	$(GOLANGCI_LINT) fmt ./...

lint:
	$(GOLANGCI_LINT) run ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

tidy-check:
	$(GO) mod tidy -diff

check: build vet lint tidy-check
