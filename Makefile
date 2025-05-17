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
	@grep -v "fakes/" ./coverage.out > coverage_filtered.out

.PHONY: show-coverage
show-coverage: test-coverage
	@echo ""
	@echo "Fully covered:"
	@go tool cover -func ./coverage_filtered.out | grep -e 100.0% | sed 's/github.com\/ammesonb\/ubiquiti-config-generator\///g'
	@echo ""
	@echo "Partial cover:"
	@go tool cover -func ./coverage_filtered.out | grep -ve 100.0% | sed 's/github.com\/ammesonb\/ubiquiti-config-generator\///g'
	@echo ""
	@make clean-coverage

.PHONY: show-uncovered
show-uncovered: test-coverage
	@go tool cover -func ./coverage_filtered.out | grep -ve 100.0% | sed 's/github.com\/ammesonb\/ubiquiti-config-generator\///g'
	@make clean-coverage

.PHONY: clean-coverage
clean-coverage:
	rm ./coverage.out ./coverage_filtered.out

.PHONY: build
build:
	docker-compose run build
