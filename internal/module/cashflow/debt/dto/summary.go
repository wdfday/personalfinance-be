package dto

import (
	"time"

	"github.com/google/uuid"
)

// DebtMonthlySummary represents aggregated debt payment data for a specific month
type DebtMonthlySummary struct {
	DebtID       uuid.UUID `json:"debt_id"`
	Name         string    `json:"name"`
	TotalPaid    float64   `json:"total_paid"`    // Total amount paid (principal + interest)
	PaymentCount int       `json:"payment_count"` // Number of payments made
}

// DebtAllTimeSummary represents aggregated debt payment data from inception to present
type DebtAllTimeSummary struct {
	DebtID         uuid.UUID  `json:"debt_id"`
	Name           string     `json:"name"`
	TotalPaid      float64    `json:"total_paid"`      // Total amount paid
	TotalPrincipal float64    `json:"total_principal"` // Total principal paid
	TotalInterest  float64    `json:"total_interest"`  // Total interest paid
	PaymentCount   int        `json:"payment_count"`
	FirstPayment   *time.Time `json:"first_payment,omitempty"`
	LastPayment    *time.Time `json:"last_payment,omitempty"`
}
