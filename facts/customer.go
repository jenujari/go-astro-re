package facts

// Customer is fact object used by GRL rules.
type Customer struct {
	Age      int
	IsAdult  bool
	Discount float64
	TaxRate  float64
}
