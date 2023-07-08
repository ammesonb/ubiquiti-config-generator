all: lint test

.PHONY: lint
lint:
	docker-compose run lint

.PHONY: test
test:
	docker-compose run test

.PHONY: build
build:
	docker-compose run build
