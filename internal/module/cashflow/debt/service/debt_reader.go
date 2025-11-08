package service

import (
	"context"
	"personalfinancedss/internal/module/cashflow/debt/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetDebtByID retrieves a debt by its ID
func (s *debtService) GetDebtByID(ctx context.Context, debtID uuid.UUID) (*domain.Debt, error) {
	debt, err := s.repo.FindByID(ctx, debtID)
	if err != nil {
		s.logger.Error("Failed to get debt by ID",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return debt, nil
}

// GetUserDebts retrieves all debts for a user
func (s *debtService) GetUserDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	debts, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user debts",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return debts, nil
}

// GetActiveDebts retrieves all active debts for a user
func (s *debtService) GetActiveDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	debts, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get active debts",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return debts, nil
}

// GetDebtsByType retrieves debts of a specific type
func (s *debtService) GetDebtsByType(ctx context.Context, userID uuid.UUID, debtType domain.DebtType) ([]domain.Debt, error) {
	debts, err := s.repo.FindByType(ctx, userID, debtType)
	if err != nil {
		s.logger.Error("Failed to get debts by type",
			zap.String("user_id", userID.String()),
			zap.String("debt_type", string(debtType)),
			zap.Error(err),
		)
		return nil, err
	}
	return debts, nil
}

// GetPaidOffDebts retrieves all paid off debts for a user
func (s *debtService) GetPaidOffDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	debts, err := s.repo.FindPaidOffDebts(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get paid off debts",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}
	return debts, nil
}

// GetDebtSummary calculates and returns a summary of all debts for a user
func (s *debtService) GetDebtSummary(ctx context.Context, userID uuid.UUID) (*DebtSummary, error) {
	debts, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get debts for summary",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	summary := &DebtSummary{
		DebtsByType:   make(map[string]*DebtTypeSum),
		DebtsByStatus: make(map[string]int),
	}

	var totalProgress float64

	for _, debt := range debts {
		summary.TotalDebts++
		summary.TotalPrincipalAmount += debt.PrincipalAmount
		summary.TotalCurrentBalance += debt.CurrentBalance
		summary.TotalPaid += debt.TotalPaid
		summary.TotalRemaining += debt.RemainingAmount
		summary.TotalInterestPaid += debt.TotalInterestPaid
		totalProgress += debt.PercentagePaid

		// Count by status
		switch debt.Status {
		case domain.DebtStatusActive:
			summary.ActiveDebts++
		case domain.DebtStatusPaidOff:
			summary.PaidOffDebts++
		case domain.DebtStatusDefaulted:
			summary.OverdueDebts++
		}

		// Sum by type
		typeKey := string(debt.Type)
		if summary.DebtsByType[typeKey] == nil {
			summary.DebtsByType[typeKey] = &DebtTypeSum{}
		}
		typeSum := summary.DebtsByType[typeKey]
		typeSum.Count++
		typeSum.PrincipalAmount += debt.PrincipalAmount
		typeSum.CurrentBalance += debt.CurrentBalance
		typeSum.TotalPaid += debt.TotalPaid
		if typeSum.PrincipalAmount > 0 {
			typeSum.Progress = ((typeSum.PrincipalAmount - typeSum.CurrentBalance) / typeSum.PrincipalAmount) * 100
		}

		// Count by status
		statusKey := string(debt.Status)
		summary.DebtsByStatus[statusKey]++
	}

	if summary.TotalDebts > 0 {
		summary.AverageProgress = totalProgress / float64(summary.TotalDebts)
	}

	s.logger.Info("Debt summary calculated",
		zap.String("user_id", userID.String()),
		zap.Int("total_debts", summary.TotalDebts),
		zap.Int("active_debts", summary.ActiveDebts),
	)

	return summary, nil
}
