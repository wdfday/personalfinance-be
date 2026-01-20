package dto

import (
	"personalfinancedss/internal/module/cashflow/debt/domain"
	"time"

	"github.com/google/uuid"
)

// CreateDebtRequest represents a request to create a new debt
type CreateDebtRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description *string             `json:"description"`
	Type        domain.DebtType     `json:"type" binding:"required"`
	Behavior    domain.DebtBehavior `json:"behavior" binding:"required"`
	Status      *domain.DebtStatus  `json:"status"`

	PrincipalAmount float64 `json:"principal_amount" binding:"required,gt=0"`
	CurrentBalance  float64 `json:"current_balance" binding:"required,gt=0"`
	InterestRate    float64 `json:"interest_rate" binding:"gte=0,lte=100"`
	MinimumPayment  float64 `json:"minimum_payment" binding:"gte=0"`
	PaymentAmount   float64 `json:"payment_amount" binding:"gte=0"`
	Currency        string  `json:"currency" binding:"required,len=3"`

	PaymentFrequency *domain.PaymentFrequency `json:"payment_frequency"`
	NextPaymentDate  *time.Time               `json:"next_payment_date"`

	StartDate time.Time  `json:"start_date" binding:"required"`
	DueDate   *time.Time `json:"due_date"`

	CreditorName    *string    `json:"creditor_name"`
	AccountNumber   *string    `json:"account_number"`
	LinkedAccountID *uuid.UUID `json:"linked_account_id"`

	EnableReminders   bool    `json:"enable_reminders"`
	ReminderFrequency *string `json:"reminder_frequency"`

	Notes *string `json:"notes"`
	Tags  *string `json:"tags"`
}

// UpdateDebtRequest represents a request to update an existing debt
type UpdateDebtRequest struct {
	Name        *string              `json:"name"`
	Description *string              `json:"description"`
	Type        *domain.DebtType     `json:"type"`
	Behavior    *domain.DebtBehavior `json:"behavior"`
	Status      *domain.DebtStatus   `json:"status"`

	PrincipalAmount *float64 `json:"principal_amount" binding:"omitempty,gt=0"`
	CurrentBalance  *float64 `json:"current_balance" binding:"omitempty,gte=0"`
	InterestRate    *float64 `json:"interest_rate" binding:"omitempty,gte=0,lte=100"`
	MinimumPayment  *float64 `json:"minimum_payment" binding:"omitempty,gte=0"`
	PaymentAmount   *float64 `json:"payment_amount" binding:"omitempty,gte=0"`
	Currency        *string  `json:"currency" binding:"omitempty,len=3"`

	PaymentFrequency *domain.PaymentFrequency `json:"payment_frequency"`
	NextPaymentDate  *time.Time               `json:"next_payment_date"`
	LastPaymentDate  *time.Time               `json:"last_payment_date"`

	StartDate *time.Time `json:"start_date"`
	DueDate   *time.Time `json:"due_date"`

	CreditorName    *string    `json:"creditor_name"`
	AccountNumber   *string    `json:"account_number"`
	LinkedAccountID *uuid.UUID `json:"linked_account_id"`

	EnableReminders   *bool   `json:"enable_reminders"`
	ReminderFrequency *string `json:"reminder_frequency"`

	Notes *string `json:"notes"`
	Tags  *string `json:"tags"`
}

// AddPaymentRequest represents a request to add a payment to a debt
type AddPaymentRequest struct {
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Description *string    `json:"description"`
	Date        *time.Time `json:"date"`
}
