# Indian Vedic Astrology Quantification Engine

Starter codebase for a rule-based quantification engine focused on Indian Vedic astrology. It is framework-first: context building, concurrent rule execution, explainable scoring, and auditable persistence are in place so rules can be added gradually without changing the engine.

## Project Layout
```text
cmd/
  server/    HTTP server
  cli/       CLI runner
  migrate/   SQL migration runner
config/
docs/
internal/
  application/
  config/
  domain/
  engine/
  infrastructure/astrology/mock/
  interfaces/http/
  interfaces/cli/
  observability/
  rules/
  storage/postgres/
sql/migrations/
```

## Quick Start
```bash
docker compose up -d postgres
make migrate
make test
make run-server
```

## REST Endpoints
- `GET /healthz`
- `POST /v1/evaluations`
- `GET /v1/evaluations/{id}`
- `GET /v1/rules/active`

## CLI
```bash
go run ./cmd/cli -config config/config.yaml \
  -datetime 2025-01-01T12:00:00Z \
  -timezone Asia/Kolkata \
  -location-name Delhi \
  -country-code IN \
  -lat 28.6139 \
  -lon 77.2090 \
  -profile default \
  -workers 4
```

## Example Request
```json
{
  "datetime": "2025-01-01T12:00:00Z",
  "timezone": "Asia/Kolkata",
  "location": {
    "name": "Delhi",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "country_code": "IN"
  },
  "calculation_profile": "default",
  "metadata": {
    "source": "demo"
  },
  "worker_count": 4
}
```

## Notes
- `config/config.yaml` is loaded via `viper`.
- `engine.persist_results` can disable database writes for ad hoc runs.
- The current astrology builder is mocked and deterministic.
