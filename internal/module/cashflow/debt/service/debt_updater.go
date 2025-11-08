package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateDebt updates an existing debt
func (s *debtService) UpdateDebt(ctx context.Context, debt *domain.Debt) error {
	if err := s.validateDebt(debt); err != nil {
		return err
	}

	debt.UpdateCalculatedFields()

	// Recalculate next payment date if needed
	if debt.PaymentFrequency != nil {
		nextDate := debt.CalculateNextPaymentDate()
		if debt.NextPaymentDate == nil || (nextDate != nil && nextDate.After(*debt.NextPaymentDate)) {
			debt.NextPaymentDate = nextDate
		}
	}

	if err := s.repo.Update(ctx, debt); err != nil {
		s.logger.Error("Failed to update debt",
			zap.String("debt_id", debt.ID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Debt updated successfully",
		zap.String("debt_id", debt.ID.String()),
		zap.String("debt_name", debt.Name),
	)

	return nil
}

// CalculateProgress recalculates progress fields for a debt
func (s *debtService) CalculateProgress(ctx context.Context, debtID uuid.UUID) error {
	debt, err := s.repo.FindByID(ctx, debtID)
	if err != nil {
		s.logger.Error("Failed to find debt for progress calculation",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return err
	}

	debt.UpdateCalculatedFields()

	if err := s.repo.Update(ctx, debt); err != nil {
		s.logger.Error("Failed to update debt progress",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Debt progress calculated",
		zap.String("debt_id", debtID.String()),
		zap.Float64("percentage_paid", debt.PercentagePaid),
	)

	return nil
}

// CheckOverdueDebts checks for overdue debts and updates their status
func (s *debtService) CheckOverdueDebts(ctx context.Context, userID uuid.UUID) error {
	overdueDebts, err := s.repo.FindOverdueDebts(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find overdue debts",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return err
	}

	for _, debt := range overdueDebts {
		debt.Status = domain.DebtStatusDefaulted
		if err := s.repo.Update(ctx, &debt); err != nil {
			s.logger.Error("Failed to mark debt as overdue",
				zap.String("debt_id", debt.ID.String()),
				zap.Error(err),
			)
			// Continue processing other debts even if one fails
			continue
		}

		s.logger.Info("Debt marked as overdue",
			zap.String("debt_id", debt.ID.String()),
			zap.String("debt_name", debt.Name),
		)
	}

	s.logger.Info("Overdue debts check completed",
		zap.String("user_id", userID.String()),
		zap.Int("overdue_count", len(overdueDebts)),
	)

	return nil
}
