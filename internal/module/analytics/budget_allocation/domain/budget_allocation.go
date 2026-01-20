package domain

// SimpleCategoryAllocation represents a simple category allocation (used by legacy service)
type SimpleCategoryAllocation struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}
