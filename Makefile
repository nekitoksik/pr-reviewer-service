.PHONY: build test lint e2e-test docker-up docker-down

build:
	go build -o bin/server ./cmd/main

test:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

e2e-test:
	go test -v -tags=e2e ./tests

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down