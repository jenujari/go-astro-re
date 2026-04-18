package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// Fact struct (data the rules will operate on)
type Customer struct {
	Age      int
	IsAdult  bool
	Discount float64
}

func Rules() pkg.Resource {
	// 1️⃣ Define rule in GRL (Grule Rule Language)
	rule := `
	rule CheckAdult "Determine if customer is adult" salience 10 {
		when
			Customer.Age >= 18 && Customer.IsAdult == false
		then
			Customer.IsAdult = true;
	}

	rule SeniorDiscount "Give discount to seniors" salience 5 {
		when
			Customer.Age >= 60 && Customer.Discount == 0.0
		then
			Customer.Discount = 0.2;
	}
	`
	return pkg.NewBytesResource([]byte(rule))
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

	// 2️⃣ Create Knowledge Library (holds rules)
	lib := ast.NewKnowledgeLibrary()

	// 3️⃣ Builder parses GRL into executable rules
	ruleBuilder := builder.NewRuleBuilder(lib)

	// 4️⃣ Add rules to library
	err := ruleBuilder.BuildRuleFromResource("example", "0.0.1", Rules())
	if err != nil {
		writeRuleError(w, fmt.Errorf("build rules: %w", err))
		return
	}

	// 5️⃣ Create Knowledge Base (runtime instance of rules)
	kb, err := lib.NewKnowledgeBaseInstance("example", "0.0.1")
	if err != nil {
		writeRuleError(w, fmt.Errorf("new knowledge base: %w", err))
		return
	}

	// 6️⃣ Create data (facts)
	customer := &Customer{
		Age: 65,
	}

	// 7️⃣ Create Data Context (bind Go structs to rules)
	dataCtx := ast.NewDataContext()
	err = dataCtx.Add("Customer", customer)
	if err != nil {
		writeRuleError(w, fmt.Errorf("bind data context: %w", err))
		return
	}

	// 8️⃣ Create rule engine
	gruleEngine := engine.NewGruleEngine()

	// 9️⃣ Execute rules
	err = gruleEngine.Execute(dataCtx, kb)
	if err != nil {
		writeRuleError(w, fmt.Errorf("execute rules: %w", err))
		return
	}

	// 🔟 Check results
	fmt.Printf("Customer: %+v\n", customer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}
