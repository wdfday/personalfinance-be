package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/debt/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AddPayment adds a payment to a debt and updates balances
func (s *debtService) AddPayment(ctx context.Context, debtID uuid.UUID, amount float64) (*domain.Debt, error) {
	if amount <= 0 {
		return nil, errors.New("payment amount must be greater than 0")
	}

	debt, err := s.repo.FindByID(ctx, debtID)
	if err != nil {
		s.logger.Error("Failed to find debt for payment",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Record original balance for logging
	originalBalance := debt.CurrentBalance

	// Add payment using domain logic
	debt.AddPayment(amount)

	// Update next payment date if frequency is set
	if debt.PaymentFrequency != nil {
		nextDate := debt.CalculateNextPaymentDate()
		debt.NextPaymentDate = nextDate
	}

	if err := s.repo.Update(ctx, debt); err != nil {
		s.logger.Error("Failed to update debt after payment",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	s.logger.Info("Payment added to debt",
		zap.String("debt_id", debtID.String()),
		zap.Float64("amount", amount),
		zap.Float64("original_balance", originalBalance),
		zap.Float64("new_balance", debt.CurrentBalance),
		zap.Float64("total_paid", debt.TotalPaid),
	)

	return debt, nil
}

// MarkAsPaidOff marks a debt as completely paid off
func (s *debtService) MarkAsPaidOff(ctx context.Context, debtID uuid.UUID) error {
	debt, err := s.repo.FindByID(ctx, debtID)
	if err != nil {
		s.logger.Error("Failed to find debt to mark as paid off",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return err
	}

	debt.Status = domain.DebtStatusPaidOff
	debt.CurrentBalance = 0
	now := time.Now()
	debt.PaidOffDate = &now
	debt.UpdateCalculatedFields()

	if err := s.repo.Update(ctx, debt); err != nil {
		s.logger.Error("Failed to mark debt as paid off",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Debt marked as paid off",
		zap.String("debt_id", debtID.String()),
		zap.String("debt_name", debt.Name),
		zap.Float64("principal_amount", debt.PrincipalAmount),
		zap.Float64("total_paid", debt.TotalPaid),
		zap.Float64("total_interest_paid", debt.TotalInterestPaid),
	)

	return nil
}
