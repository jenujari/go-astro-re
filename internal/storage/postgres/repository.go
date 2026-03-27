package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"local.io/go-astro-re/internal/config"
	"local.io/go-astro-re/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sql ping: %w", err)
	}
	return db, nil
}

func NewRepository(db *sql.DB) Repository {
	return Repository{db: db}
}

func (r Repository) SaveEvaluation(ctx context.Context, report domain.EvaluationReport) (requestID int64, evaluationID int64, err error) {
	if r.db == nil {
		return 0, 0, errors.New("database not configured")
	}

	contextJSON, err := json.Marshal(report.Context)
	if err != nil {
		return 0, 0, fmt.Errorf("marshal context: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO evaluation_requests (
			request_datetime, request_timezone, location_name, latitude, longitude, country_code, calculation_profile, request_metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`,
		report.Input.DateTime,
		report.Input.Timezone,
		report.Input.Location.Name,
		report.Input.Location.Latitude,
		report.Input.Location.Longitude,
		report.Input.Location.CountryCode,
		report.Input.CalculationProfile,
		mustJSON(report.Input.Metadata),
	).Scan(&requestID)
	if err != nil {
		return 0, 0, fmt.Errorf("insert evaluation_request: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO astrology_context_snapshots (
			evaluation_request_id, builder_version, context_payload
		) VALUES ($1,$2,$3)
	`,
		requestID,
		report.Context.BuilderVersion,
		contextJSON,
	); err != nil {
		return 0, 0, fmt.Errorf("insert context snapshot: %w", err)
	}

	status := "success"
	if len(report.PartialFailures) > 0 {
		status = "partial_success"
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO evaluation_results (
			evaluation_request_id, status, positive_total, negative_total, net_score, partial_failure_count
		) VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id
	`,
		requestID,
		status,
		report.Summary.PositiveTotal,
		report.Summary.NegativeTotal,
		report.Summary.NetScore,
		len(report.PartialFailures),
	).Scan(&evaluationID)
	if err != nil {
		return 0, 0, fmt.Errorf("insert evaluation_result: %w", err)
	}

	for _, item := range report.RuleResults {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO rule_results (
				evaluation_result_id, rule_id, rule_version, category, status, matched, polarity,
				raw_score, weight, confidence_multiplier, weighted_score, explanation, facts_used, error_text, duration_ms
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		`,
			evaluationID,
			item.RuleID,
			item.RuleVersion,
			item.Category,
			item.Status,
			item.Matched,
			item.Polarity,
			item.RawScore,
			item.Weight,
			item.ConfidenceMultiplier,
			item.WeightedScore,
			item.Explanation,
			mustJSON(item.FactsUsed),
			item.ErrorText,
			item.DurationMillis,
		); err != nil {
			return 0, 0, fmt.Errorf("insert rule_result %s: %w", item.RuleID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit tx: %w", err)
	}
	return requestID, evaluationID, nil
}

func (r Repository) GetEvaluation(ctx context.Context, evaluationID int64) (domain.EvaluationReport, error) {
	if r.db == nil {
		return domain.EvaluationReport{}, errors.New("database not configured")
	}

	var report domain.EvaluationReport
	var contextJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT
			er.id,
			er.evaluation_request_id,
			er.created_at,
			req.request_datetime,
			req.request_timezone,
			req.location_name,
			req.latitude,
			req.longitude,
			req.country_code,
			req.calculation_profile,
			er.positive_total,
			er.negative_total,
			er.net_score,
			acs.context_payload
		FROM evaluation_results er
		JOIN evaluation_requests req ON req.id = er.evaluation_request_id
		JOIN astrology_context_snapshots acs ON acs.evaluation_request_id = req.id
		WHERE er.id = $1
	`, evaluationID).Scan(
		&report.EvaluationID,
		&report.RequestID,
		&report.CreatedAt,
		&report.Input.DateTime,
		&report.Input.Timezone,
		&report.Input.Location.Name,
		&report.Input.Location.Latitude,
		&report.Input.Location.Longitude,
		&report.Input.Location.CountryCode,
		&report.Input.CalculationProfile,
		&report.Summary.PositiveTotal,
		&report.Summary.NegativeTotal,
		&report.Summary.NetScore,
		&contextJSON,
	)
	if err != nil {
		return domain.EvaluationReport{}, fmt.Errorf("query evaluation: %w", err)
	}

	if err := json.Unmarshal(contextJSON, &report.Context); err != nil {
		return domain.EvaluationReport{}, fmt.Errorf("unmarshal context: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			rule_id, rule_version, category, status, matched, polarity,
			raw_score, weight, confidence_multiplier, weighted_score, explanation, facts_used, error_text, duration_ms
		FROM rule_results
		WHERE evaluation_result_id = $1
		ORDER BY category, rule_id
	`, evaluationID)
	if err != nil {
		return domain.EvaluationReport{}, fmt.Errorf("query rule results: %w", err)
	}
	defer rows.Close()

	report.RuleResults = make([]domain.RuleResult, 0)
	for rows.Next() {
		var result domain.RuleResult
		var factsJSON []byte
		if err := rows.Scan(
			&result.RuleID,
			&result.RuleVersion,
			&result.Category,
			&result.Status,
			&result.Matched,
			&result.Polarity,
			&result.RawScore,
			&result.Weight,
			&result.ConfidenceMultiplier,
			&result.WeightedScore,
			&result.Explanation,
			&factsJSON,
			&result.ErrorText,
			&result.DurationMillis,
		); err != nil {
			return domain.EvaluationReport{}, fmt.Errorf("scan rule result: %w", err)
		}
		if err := json.Unmarshal(factsJSON, &result.FactsUsed); err != nil {
			return domain.EvaluationReport{}, fmt.Errorf("unmarshal facts used: %w", err)
		}
		report.RuleResults = append(report.RuleResults, result)
	}
	if err := rows.Err(); err != nil {
		return domain.EvaluationReport{}, fmt.Errorf("iterate rule results: %w", err)
	}

	report.Summary.CategoryTotals = buildCategoryTotals(report.RuleResults)
	return report, nil
}

func ApplyMigrations(ctx context.Context, db *sql.DB, dir string) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}

func buildCategoryTotals(results []domain.RuleResult) map[string]domain.CategoryScore {
	totals := make(map[string]domain.CategoryScore)
	for _, result := range results {
		item := totals[result.Category]
		if result.WeightedScore > 0 {
			item.Positive += result.WeightedScore
		}
		if result.WeightedScore < 0 {
			item.Negative += -result.WeightedScore
		}
		item.Net = item.Positive - item.Negative
		totals[result.Category] = item
	}
	return totals
}

func mustJSON(v any) []byte {
	out, _ := json.Marshal(v)
	return out
}
