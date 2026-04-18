package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
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
	if ruleRuntime == nil {
		writeRuleError(w, fmt.Errorf("rule runtime not initialized"))
		return
	}

	kb, err := ruleRuntime.NewKnowledgeBaseInstance()
	if err != nil {
		writeRuleError(w, fmt.Errorf("new knowledge base: %w", err))
		return
	}

	customer := &rulesengine.Customer{
		Age: 65,
	}

	dataCtx := ast.NewDataContext()
	err = dataCtx.Add("Customer", customer)
	if err != nil {
		writeRuleError(w, fmt.Errorf("bind data context: %w", err))
		return
	}

	gruleEngine := engine.NewGruleEngine()
	gruleEngine.MaxCycle = 5

	err = gruleEngine.Execute(dataCtx, kb)
	if err != nil {
		writeRuleError(w, fmt.Errorf("execute rules: %w", err))
		return
	}

	fmt.Printf("Customer: %+v\n", customer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}
