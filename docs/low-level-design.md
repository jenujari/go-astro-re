# Indian Vedic Astrology Quantification Engine: Low-Level Design

## Problem Definition
The system accepts a birth or event datetime with timezone, location, and calculation profile; builds a reusable astrology context; evaluates many independent Vedic astrology rules; aggregates positive and negative contributions; persists the evaluation for auditability and later backtesting; and returns an explainable report.

## Assumptions
- The starter uses mocked astrology facts and planet positions.
- Real astronomical or jyotish calculations will later be introduced behind the `AstrologyContextBuilder` interface.
- Rules are primarily code-defined in the first phase, with database rows serving audit, metadata, and future admin use cases.
- A single evaluation is independent and can execute rules concurrently.

## Non-Functional Requirements
- Deterministic outputs for the same input and enabled rule set.
- Explainability for every rule, including non-matches and failures.
- Safe parallel execution with bounded concurrency.
- Panic isolation so one faulty rule does not crash the request.
- Persistent audit trail suitable for backtesting and analytics.
- Clean package boundaries to support long-term rule growth.

## Architecture Overview
- `internal/domain`: business types and contracts.
- `internal/application`: evaluation orchestration and score aggregation.
- `internal/engine`: concurrent rule execution with worker pool.
- `internal/infrastructure/astrology/mock`: mocked context builder.
- `internal/rules`: registry and plugin-style rules.
- `internal/storage/postgres`: migrations and persistence.
- `internal/interfaces/http` and `internal/interfaces/cli`: adapters.
- `internal/observability`: structured logging and metrics abstraction.

## Module Responsibilities
- `AstrologyContextBuilder`: computes shared facts once per request.
- `RuleRegistry`: manages rule discovery and active filtering.
- `RuleEvaluator`: runs active rules concurrently and captures partial failures.
- `Aggregator`: produces positive, negative, net, and category totals.
- `Repository`: stores request, context snapshot, aggregate result, and per-rule outputs.
- `EvaluationService`: use-case coordinator for API and CLI.

## Execution Flow
1. Transport layer validates input and converts it into `AstrologyInput`.
2. The service calls `AstrologyContextBuilder.Build`.
3. The evaluator loads active rules from the registry.
4. A bounded worker pool runs rules concurrently with `context.Context`.
5. Each rule returns `RuleResult` with score, explanation, and facts used.
6. The aggregator computes summary totals.
7. If persistence is enabled, the repository stores normalized rows and the JSON context snapshot.
8. The service returns an `EvaluationReport`.

## Concurrency Model
- The engine uses a bounded worker pool configured per request or from static config.
- Rules are submitted as jobs and results are collected through a buffered channel.
- Final ordering is deterministic because results are sorted by category and rule ID before aggregation.
- Context cancellation stops dispatching new jobs and lets cooperative rules terminate early.
- Panic recovery wraps each rule evaluation and converts failures into structured `RuleResult` errors.

## Error Handling Strategy
- Input validation errors return HTTP 400.
- Infrastructure failures return HTTP 500.
- Rule failures are captured as partial failures without aborting the whole evaluation.
- Persistence-disabled reads return a dedicated service error.

## Logging and Observability
- Structured JSON logging is implemented with `log/slog`.
- Metrics hooks are abstracted behind a lightweight interface so Prometheus or OpenTelemetry can be added later.
- Recommended future telemetry:
  - evaluation latency histogram
  - rule execution latency histogram
  - partial failure counter by rule ID
  - persistence failures counter

## Extensibility Plan
- Add a rule by creating a new package under `internal/rules/plugins/...` and registering it in `init()`.
- Rule metadata already includes version, category, tags, status, priority, and modifier flag.
- The schema supports rule metadata and version history independently from runtime evaluation rows.

## Future Evolution Plan
- Replace the mock builder with a real ephemeris-backed Vedic engine.
- Add weighted rules and modifier rules using `weight`, `confidence_multiplier`, and `is_modifier`.
- Introduce rule dependencies or staged evaluation when truly required.
- Add backtesting endpoints over date ranges.
- Add an admin UI for rule catalog inspection and version rollout.
