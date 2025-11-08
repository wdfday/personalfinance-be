package service

import (
	"errors"
	"fmt"
	"personalfinancedss/internal/module/cashflow/debt/domain"
)

// validateDebt validates debt fields before create or update
func (s *debtService) validateDebt(debt *domain.Debt) error {
	if debt.PrincipalAmount <= 0 {
		return errors.New("principal amount must be greater than 0")
	}

	if debt.CurrentBalance < 0 {
		return errors.New("current balance cannot be negative")
	}

	if debt.CurrentBalance > debt.PrincipalAmount {
		return errors.New("current balance cannot exceed principal amount")
	}

	if !debt.Type.IsValid() {
		return fmt.Errorf("invalid debt type: %s", debt.Type)
	}

	if !debt.Status.IsValid() {
		return fmt.Errorf("invalid debt status: %s", debt.Status)
	}

	if debt.InterestRate < 0 || debt.InterestRate > 100 {
		return errors.New("interest rate must be between 0 and 100")
	}

	if debt.DueDate != nil && debt.DueDate.Before(debt.StartDate) {
		return errors.New("due date must be after start date")
	}

	if debt.PaymentFrequency != nil && !debt.PaymentFrequency.IsValid() {
		return fmt.Errorf("invalid payment frequency: %s", *debt.PaymentFrequency)
	}

	return nil
}
