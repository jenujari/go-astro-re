package postgres_test

import (
	"context"
	"testing"

	"local.io/go-astro-re/internal/domain"
	"local.io/go-astro-re/internal/storage/postgres"
)

func TestRepositoryWithoutDatabase(t *testing.T) {
	repo := postgres.NewRepository(nil)
	_, _, err := repo.SaveEvaluation(context.Background(), domain.EvaluationReport{})
	if err == nil {
		t.Fatal("expected error when database is nil")
	}
}
