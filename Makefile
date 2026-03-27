.PHONY: fmt test run-server run-cli migrate build

fmt:
	gofmt -w ./cmd ./internal

test:
	GOCACHE=/tmp/go-build go test ./...

run-server:
	go run ./cmd/server config/config.yaml

run-cli:
	go run ./cmd/cli -config config/config.yaml -datetime 2025-01-01T12:00:00Z

migrate:
	go run ./cmd/migrate config/config.yaml

infra-up:
	podman compose -f docker-compose.yml up -d

build:
	mkdir -p bin
	go build -o bin/vedic-astro-server ./cmd/server
	go build -o bin/vedic-astro-cli ./cmd/cli
