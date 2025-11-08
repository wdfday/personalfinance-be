package dto

import (
	"personalfinancedss/internal/module/cashflow/debt/domain"
	"personalfinancedss/internal/module/cashflow/debt/service"
	"time"

	"github.com/google/uuid"
)

// DebtResponse represents a debt in API responses
type DebtResponse struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`

	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Type        domain.DebtType   `json:"type"`
	Status      domain.DebtStatus `json:"status"`

	PrincipalAmount float64 `json:"principal_amount"`
	CurrentBalance  float64 `json:"current_balance"`
	InterestRate    float64 `json:"interest_rate"`
	MinimumPayment  float64 `json:"minimum_payment"`
	PaymentAmount   float64 `json:"payment_amount"`
	Currency        string  `json:"currency"`

	PaymentFrequency  *domain.PaymentFrequency `json:"payment_frequency,omitempty"`
	NextPaymentDate   *time.Time               `json:"next_payment_date,omitempty"`
	LastPaymentDate   *time.Time               `json:"last_payment_date,omitempty"`
	LastPaymentAmount *float64                 `json:"last_payment_amount,omitempty"`

	StartDate   time.Time  `json:"start_date"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	PaidOffDate *time.Time `json:"paid_off_date,omitempty"`

	TotalPaid         float64 `json:"total_paid"`
	RemainingAmount   float64 `json:"remaining_amount"`
	PercentagePaid    float64 `json:"percentage_paid"`
	TotalInterestPaid float64 `json:"total_interest_paid"`

	CreditorName    *string    `json:"creditor_name,omitempty"`
	AccountNumber   *string    `json:"account_number,omitempty"`
	LinkedAccountID *uuid.UUID `json:"linked_account_id,omitempty"`

	EnableReminders    bool       `json:"enable_reminders"`
	ReminderFrequency  *string    `json:"reminder_frequency,omitempty"`
	LastReminderSentAt *time.Time `json:"last_reminder_sent_at,omitempty"`

	Notes *string `json:"notes,omitempty"`
	Tags  *string `json:"tags,omitempty"`

	DaysUntilNextPayment int `json:"days_until_next_payment"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// DebtSummaryResponse represents a debt summary in API responses
type DebtSummaryResponse struct {
	TotalDebts           int                             `json:"total_debts"`
	ActiveDebts          int                             `json:"active_debts"`
	PaidOffDebts         int                             `json:"paid_off_debts"`
	OverdueDebts         int                             `json:"overdue_debts"`
	TotalPrincipalAmount float64                         `json:"total_principal_amount"`
	TotalCurrentBalance  float64                         `json:"total_current_balance"`
	TotalPaid            float64                         `json:"total_paid"`
	TotalRemaining       float64                         `json:"total_remaining"`
	TotalInterestPaid    float64                         `json:"total_interest_paid"`
	AverageProgress      float64                         `json:"average_progress"`
	DebtsByType          map[string]*service.DebtTypeSum `json:"debts_by_type"`
	DebtsByStatus        map[string]int                  `json:"debts_by_status"`
}

// ToDebtResponse converts a domain debt to response DTO
func ToDebtResponse(debt *domain.Debt) *DebtResponse {
	if debt == nil {
		return nil
	}

	return &DebtResponse{
		ID:                   debt.ID,
		UserID:               debt.UserID,
		Name:                 debt.Name,
		Description:          debt.Description,
		Type:                 debt.Type,
		Status:               debt.Status,
		PrincipalAmount:      debt.PrincipalAmount,
		CurrentBalance:       debt.CurrentBalance,
		InterestRate:         debt.InterestRate,
		MinimumPayment:       debt.MinimumPayment,
		PaymentAmount:        debt.PaymentAmount,
		Currency:             debt.Currency,
		PaymentFrequency:     debt.PaymentFrequency,
		NextPaymentDate:      debt.NextPaymentDate,
		LastPaymentDate:      debt.LastPaymentDate,
		LastPaymentAmount:    debt.LastPaymentAmount,
		StartDate:            debt.StartDate,
		DueDate:              debt.DueDate,
		PaidOffDate:          debt.PaidOffDate,
		TotalPaid:            debt.TotalPaid,
		RemainingAmount:      debt.RemainingAmount,
		PercentagePaid:       debt.PercentagePaid,
		TotalInterestPaid:    debt.TotalInterestPaid,
		CreditorName:         debt.CreditorName,
		AccountNumber:        debt.AccountNumber,
		LinkedAccountID:      debt.LinkedAccountID,
		EnableReminders:      debt.EnableReminders,
		ReminderFrequency:    debt.ReminderFrequency,
		LastReminderSentAt:   debt.LastReminderSentAt,
		Notes:                debt.Notes,
		Tags:                 debt.Tags,
		DaysUntilNextPayment: debt.DaysUntilNextPayment(),
		CreatedAt:            debt.CreatedAt,
		UpdatedAt:            debt.UpdatedAt,
	}
}

// ToDebtResponseList converts a list of domain debts to response DTOs
func ToDebtResponseList(debts []domain.Debt) []*DebtResponse {
	responses := make([]*DebtResponse, len(debts))
	for i := range debts {
		responses[i] = ToDebtResponse(&debts[i])
	}
	return responses
}

// ToDebtSummaryResponse converts a service debt summary to response DTO
func ToDebtSummaryResponse(summary *service.DebtSummary) *DebtSummaryResponse {
	if summary == nil {
		return nil
	}

	return &DebtSummaryResponse{
		TotalDebts:           summary.TotalDebts,
		ActiveDebts:          summary.ActiveDebts,
		PaidOffDebts:         summary.PaidOffDebts,
		OverdueDebts:         summary.OverdueDebts,
		TotalPrincipalAmount: summary.TotalPrincipalAmount,
		TotalCurrentBalance:  summary.TotalCurrentBalance,
		TotalPaid:            summary.TotalPaid,
		TotalRemaining:       summary.TotalRemaining,
		TotalInterestPaid:    summary.TotalInterestPaid,
		AverageProgress:      summary.AverageProgress,
		DebtsByType:          summary.DebtsByType,
		DebtsByStatus:        summary.DebtsByStatus,
	}
}
