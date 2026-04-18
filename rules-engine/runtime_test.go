package rulesengine

import "testing"

const (
	testRuleRootDir    = "../rules"
	testRuleSetName    = "example"
	testRuleSetVersion = "0.0.1"
)

func TestAdultRule(t *testing.T) {
	runtime, err := NewRuntimeFromDir(testRuleRootDir, testRuleSetName, testRuleSetVersion)
	if err != nil {
		t.Fatalf("init runtime: %v", err)
	}

	customer := &Customer{Age: 20}

	err = runtime.ExecuteCustomerRules(customer)
	if err != nil {
		t.Fatal(err)
	}

	if !customer.IsAdult {
		t.Errorf("Expected IsAdult = true")
	}
}
