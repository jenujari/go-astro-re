package rulesengine

import (
	"testing"

	"github.com/jenujari/go-astro-re/facts"
)

const (
	testRuleRootDir    = "../rules"
	testRuleSetName    = "example"
	testRuleSetVersion = "0.0.1"
)

func TestAdultRule(t *testing.T) {
	mgr, err := NewManager(Config{
		RuleRootDir:    testRuleRootDir,
		RuleSetName:    testRuleSetName,
		DefaultVersion: testRuleSetVersion,
	})
	if err != nil {
		t.Fatalf("init manager: %v", err)
	}

	service := NewService(mgr, DefaultDataContextBuilder{}, mgr)
	customer := &facts.Customer{Age: 20}

	_, err = service.EvaluateCustomer(EvaluateCustomerInput{Customer: customer})
	if err != nil {
		t.Fatal(err)
	}

	if !customer.IsAdult {
		t.Errorf("Expected IsAdult = true")
	}
}
