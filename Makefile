all: lint test

.PHONY: precommit
precommit:
	gofumpt -w **/*.go
	go vet
	goimports -w **/*.go
	golangci-lint run
	go test ./...
	go mod tidy

.PHONY: lint
lint:
	docker-compose run lint

.PHONY: test
test:
	docker-compose run test

.PHONY: test-coverage
test-coverage:
	go test ./... -coverpkg=./... -coverprofile ./coverage.out
	@echo ""
	@echo "Fully covered:"
	@go tool cover -func ./coverage.out | grep -e 100.0%
	@echo ""
	@echo "Partial cover:"
	@go tool cover -func ./coverage.out | grep -ve 100.0%
	@echo ""
	rm ./coverage.out

.PHONY: build
build:
	docker-compose run build
