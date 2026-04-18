package facts

// Customer is fact object used by GRL rules.
type Customer struct {
	Age      int     `json:"age"`
	IsAdult  bool    `json:"isAdult"`
	Discount float64 `json:"discount"`
	TaxRate  float64 `json:"taxRate"`
}
