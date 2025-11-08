package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/debt/domain"

	"go.uber.org/zap"
)

// CreateDebt creates a new debt for a user
func (s *debtService) CreateDebt(ctx context.Context, debt *domain.Debt) error {
	if err := s.validateDebt(debt); err != nil {
		return err
	}

	debt.UpdateCalculatedFields()

	// Calculate next payment date if frequency is provided
	if debt.PaymentFrequency != nil {
		nextDate := debt.CalculateNextPaymentDate()
		debt.NextPaymentDate = nextDate
	}

	if err := s.repo.Create(ctx, debt); err != nil {
		s.logger.Error("Failed to create debt",
			zap.String("debt_name", debt.Name),
			zap.String("user_id", debt.UserID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Debt created successfully",
		zap.String("debt_id", debt.ID.String()),
		zap.String("debt_name", debt.Name),
		zap.String("user_id", debt.UserID.String()),
	)

	return nil
}
