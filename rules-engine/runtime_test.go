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

func TestAdultRuleThreePhases(t *testing.T) {
	mgr, err := NewManager(Config{
		RuleRootDir:    testRuleRootDir,
		RuleSetName:    testRuleSetName,
		DefaultVersion: testRuleSetVersion,
	})
	if err != nil {
		t.Fatalf("init manager: %v", err)
	}

	service := NewService(mgr, DefaultDataContextBuilder{}, mgr, testRuleSetVersion)
	customer := &facts.Customer{Age: 65}

	result, err := service.EvaluateCustomerPhases(EvaluateCustomerInput{Customer: customer})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Customer.IsAdult {
		t.Errorf("expected adult true")
	}
	if result.Customer.Discount <= 0 {
		t.Errorf("expected discount to be set")
	}
	if !result.Outcome.AfterEffectsApplied {
		t.Errorf("expected after effects flag true")
	}
	if len(result.Metrics.Phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(result.Metrics.Phases))
	}
	for _, phase := range result.Metrics.Phases {
		if len(phase.FiredRules) == 0 {
			t.Fatalf("expected fired rules in phase %s", phase.Phase)
		}
	}
}

func TestTenantRuleSelection(t *testing.T) {
	mgr, err := NewManager(Config{
		RuleRootDir:    testRuleRootDir,
		RuleSetName:    testRuleSetName,
		DefaultVersion: testRuleSetVersion,
	})
	if err != nil {
		t.Fatalf("init manager: %v", err)
	}

	service := NewService(mgr, DefaultDataContextBuilder{}, mgr, testRuleSetVersion)
	customer := &facts.Customer{Age: 65}

	result, err := service.EvaluateCustomerPhases(EvaluateCustomerInput{
		TenantID: "tenant-a",
		Customer: customer,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Customer.Discount != 0.3 {
		t.Fatalf("expected tenant-a discount 0.3, got %v", result.Customer.Discount)
	}
	if result.Customer.TaxRate != 0.03 {
		t.Fatalf("expected tenant-a tax rate 0.03, got %v", result.Customer.TaxRate)
	}
}
