package models

// Money represents a monetary value with an explicit currency.
// Amount is stored as a decimal string to avoid floating-point precision issues.
type Money struct {
	Amount   string
	Currency string
}
