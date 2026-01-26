package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteBudgetForUser deletes a budget with ownership verification (soft delete)
func (s *budgetService) DeleteBudgetForUser(ctx context.Context, budgetID, userID uuid.UUID) error {
	s.logger.Info("Deleting budget for user",
		zap.String("budget_id", budgetID.String()),
		zap.String("user_id", userID.String()),
	)

	return s.repo.DeleteByIDAndUserID(ctx, budgetID, userID)
}
