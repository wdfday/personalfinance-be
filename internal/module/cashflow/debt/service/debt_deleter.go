package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteDebt soft deletes a debt
func (s *debtService) DeleteDebt(ctx context.Context, debtID uuid.UUID) error {
	if err := s.repo.Delete(ctx, debtID); err != nil {
		s.logger.Error("Failed to delete debt",
			zap.String("debt_id", debtID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Debt deleted successfully",
		zap.String("debt_id", debtID.String()),
	)

	return nil
}
