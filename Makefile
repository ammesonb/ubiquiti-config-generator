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

.PHONY: build
build:
	docker-compose run build
