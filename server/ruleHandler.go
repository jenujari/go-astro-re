package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/jenujari/go-astro-re/facts"
	rulesengine "github.com/jenujari/go-astro-re/rules-engine"
)

type ruleRequest struct {
	TenantID string `json:"tenantId"`
	Version  string `json:"version"`
	Age      int    `json:"age"`
}

func writeRuleError(w http.ResponseWriter, err error) {
	if reporter, ok := err.(*pkg.GruleErrorReporter); ok {
		for i, parseErr := range reporter.Errors {
			log.Printf("rule parse error #%d: %v", i+1, parseErr)
		}
	}
	log.Printf("rule handler error: %v", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ruleHandler(w http.ResponseWriter, r *http.Request) {
	if ruleService == nil {
		writeRuleError(w, fmt.Errorf("rule runtime not initialized"))
		return
	}

	req := parseRuleRequest(r)
	customer := &facts.Customer{Age: req.Age}

	result, err := ruleService.EvaluateCustomerPhases(rulesengine.EvaluateCustomerInput{
		TenantID: req.TenantID,
		Version:  req.Version,
		Customer: customer,
	})
	if err != nil {
		writeRuleError(w, err)
		return
	}

	fmt.Printf("tenant=%s customer=%+v metrics=%+v\n", result.Metrics.TenantID, result.Customer, result.Metrics)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func parseRuleRequest(r *http.Request) ruleRequest {
	result := ruleRequest{Age: 65}

	if tenantFromHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID")); tenantFromHeader != "" {
		result.TenantID = tenantFromHeader
	}
	if versionFromHeader := strings.TrimSpace(r.Header.Get("X-Rule-Version")); versionFromHeader != "" {
		result.Version = versionFromHeader
	}

	query := r.URL.Query()
	if tenantFromQuery := strings.TrimSpace(query.Get("tenantId")); tenantFromQuery != "" {
		result.TenantID = tenantFromQuery
	}
	if versionFromQuery := strings.TrimSpace(query.Get("version")); versionFromQuery != "" {
		result.Version = versionFromQuery
	}
	if ageFromQuery := strings.TrimSpace(query.Get("age")); ageFromQuery != "" {
		if parsedAge, err := strconv.Atoi(ageFromQuery); err == nil {
			result.Age = parsedAge
		}
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var body ruleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if strings.TrimSpace(body.TenantID) != "" {
				result.TenantID = strings.TrimSpace(body.TenantID)
			}
			if strings.TrimSpace(body.Version) != "" {
				result.Version = strings.TrimSpace(body.Version)
			}
			if body.Age > 0 {
				result.Age = body.Age
			}
		}
	}

	return result
}
