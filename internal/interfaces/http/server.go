package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/domain"
)

type Service interface {
	Evaluate(context.Context, domain.AstrologyInput, int) (domain.EvaluationReport, error)
	GetEvaluation(context.Context, int64) (domain.EvaluationReport, error)
	ListActiveRules() []domain.RuleMetadata
}

type Server struct {
	service Service
	logger  *slog.Logger
}

func NewServer(service Service, logger *slog.Logger) Server {
	return Server{service: service, logger: logger}
}

func (s Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("POST /v1/evaluations", s.handleEvaluate)
	mux.HandleFunc("GET /v1/evaluations/", s.handleGetEvaluation)
	mux.HandleFunc("GET /v1/rules/active", s.handleActiveRules)
	return mux
}

func (s Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type evaluatePayload struct {
	Datetime           string            `json:"datetime"`
	Timezone           string            `json:"timezone"`
	Location           domain.Location   `json:"location"`
	CalculationProfile string            `json:"calculation_profile"`
	Metadata           map[string]string `json:"metadata"`
	WorkerCount        int               `json:"worker_count"`
}

func (s Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var payload evaluatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	dt, err := parseRFC3339(payload.Datetime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "datetime must be RFC3339"})
		return
	}

	report, err := s.service.Evaluate(r.Context(), domain.AstrologyInput{
		DateTime:           dt,
		Timezone:           payload.Timezone,
		Location:           payload.Location,
		CalculationProfile: payload.CalculationProfile,
		Metadata:           payload.Metadata,
	}, payload.WorkerCount)
	if err != nil {
		s.logger.Error("evaluate request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (s Server) handleGetEvaluation(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/v1/evaluations/")
	evaluationID, err := strconv.ParseInt(idText, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid evaluation id"})
		return
	}

	report, err := s.service.GetEvaluation(r.Context(), evaluationID)
	if err != nil {
		if errors.Is(err, application.ErrPersistenceDisabled) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "persistence disabled"})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (s Server) handleActiveRules(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"rules": s.service.ListActiveRules(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
