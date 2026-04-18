package facts

import "testing"

func TestCustomerFactShape(t *testing.T) {
	customer := Customer{Age: 65, Discount: 0.2, TaxRate: 0.05, IsAdult: true}

	if customer.Age != 65 {
		t.Fatalf("unexpected age: %d", customer.Age)
	}
	if !customer.IsAdult {
		t.Fatal("expected IsAdult=true")
	}
	if customer.Discount != 0.2 {
		t.Fatalf("unexpected discount: %f", customer.Discount)
	}
	if customer.TaxRate != 0.05 {
		t.Fatalf("unexpected tax rate: %f", customer.TaxRate)
	}
}
