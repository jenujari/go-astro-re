# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go-based starter for a rule-driven Indian Vedic astrology quantification engine. Entrypoints live in `cmd/server`, `cmd/cli`, and `cmd/migrate`. Core business logic is under `internal/`: `domain` defines models and contracts, `application` coordinates use cases, `engine` runs rules in parallel, `rules` contains the registry and plugin rules, `interfaces` holds HTTP and CLI adapters, `infrastructure/astrology/mock` provides the mocked context builder, and `storage/postgres` owns persistence and migrations. SQL lives in `sql/migrations`, config in `config/config.yaml`, and design notes in `docs/low-level-design.md`.

## Build, Test, and Development Commands
- `make fmt`: format Go files with `gofmt`.
- `make test`: run the full test suite with `GOCACHE=/tmp/go-build go test ./...`.
- `make run-server`: start the HTTP API using `config/config.yaml`.
- `make run-cli`: run a sample CLI evaluation.
- `make migrate`: apply PostgreSQL migrations.
- `make build`: build server and CLI binaries into `bin/`.

For local infrastructure, use `docker compose up -d postgres`.

## Coding Style & Naming Conventions
Use idiomatic Go with tabs, small focused packages, and explicit ownership boundaries between domain, application, interfaces, and storage. Format all edits with `gofmt`. Keep package names lowercase. Exported symbols use `PascalCase`; internal helpers use `camelCase`. New rules should be added as isolated packages under `internal/rules/plugins/<rule_name>/rule.go` and registered without modifying the evaluator.

## Testing Guidelines
Keep unit tests next to the code they verify in `*_test.go` files. Use table-driven tests where useful. This repo also uses component tests for cross-layer flows such as service + mock builder + real registry, HTTP adapter flow, and CLI flow. Run `go test ./...` before submitting changes and add tests for new rules, concurrency-sensitive logic, and API behavior.

## Commit & Pull Request Guidelines
Use short imperative commits such as `Add Jupiter strength rule` or `Refactor evaluation repository`. Keep pull requests focused and include a concise summary, test evidence, config or migration changes, and sample request/response updates when API behavior changes.

## Configuration & Operations
Configuration is loaded from `config/config.yaml` via `viper`. PostgreSQL settings, worker count, logging level, and persistence behavior are defined there. When changing the schema, add forward and rollback SQL migrations and update seed data if rule catalog metadata changes.
