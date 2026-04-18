package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/jenujari/go-astro-re/facts"
	rulesengine "github.com/jenujari/go-astro-re/rules-engine"
)

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

	customer := &facts.Customer{
		Age: 65,
	}

	metrics, err := ruleService.EvaluateCustomer(rulesengine.EvaluateCustomerInput{
		Customer: customer,
		Phase:    facts.DefaultPhase,
	})
	if err != nil {
		writeRuleError(w, err)
		return
	}

	fmt.Printf("Customer: %+v metrics=%+v\n", customer, metrics)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}
