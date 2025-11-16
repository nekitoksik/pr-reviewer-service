.PHONY: build lint docker-up docker-down

build:
	go build -o bin/server ./cmd/main

lint:
	golangci-lint run ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down