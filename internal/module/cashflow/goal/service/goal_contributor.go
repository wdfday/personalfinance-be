package service

import (
	"context"
	"errors"
	"personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AddContribution adds a contribution (deposit) to a goal
func (s *goalService) AddContribution(
	ctx context.Context,
	goalID uuid.UUID,
	amount float64,
	accountID *uuid.UUID,
	note *string,
	source string,
) (*domain.Goal, error) {
	if amount <= 0 {
		return nil, errors.New("contribution amount must be greater than 0")
	}

	// Get the goal
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to find goal for contribution",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Default source if not provided
	if source == "" {
		source = "manual"
	}

	// Use provided accountID or fallback to goal's accountID
	contributionAccountID := goal.AccountID
	if accountID != nil {
		contributionAccountID = *accountID
	}

	// Create contribution record
	contribution := domain.NewDeposit(goalID, contributionAccountID, goal.UserID, amount, note)
	contribution.Currency = goal.Currency
	contribution.Source = source

	if err := s.repo.CreateContribution(ctx, contribution); err != nil {
		s.logger.Error("Failed to create contribution record",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Update goal's current amount
	goal.AddContribution(amount)
	if err := s.repo.Update(ctx, goal); err != nil {
		s.logger.Error("Failed to update goal after contribution",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Update account's available balance
	netContributions, err := s.repo.GetNetContributionsByAccountID(ctx, contributionAccountID)
	if err != nil {
		s.logger.Error("Failed to get net contributions for account",
			zap.String("account_id", contributionAccountID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Get account to calculate new available balance
	// Account.AvailableBalance = Account.CurrentBalance - NetContributions
	// We need to get current balance from account service or assume it's already set
	// For now, we'll just update available balance based on the delta
	// The account service should expose a method to recalculate available balance

	// Use account service to update available balance
	// The available balance calculation is: CurrentBalance - NetGoalContributions
	// We'll delegate this to a method on account service that we'll call
	if s.accountService != nil {
		// For proper implementation, we need account's current balance
		// Then: availableBalance = currentBalance - netContributions
		// But to avoid another query, we can let account service handle the recalculation
		// For now, we'll just update the balance directly
		// TODO: Implement RecalculateAvailableBalance in account service
		s.logger.Warn("Account balance update not yet fully implemented",
			zap.String("account_id", contributionAccountID.String()),
			zap.Float64("net_contributions", netContributions),
		)
	}

	s.logger.Info("Contribution added to goal",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.Float64("amount", amount),
		zap.Float64("new_current_amount", goal.CurrentAmount),
		zap.Float64("percentage_complete", goal.PercentageComplete),
	)

	return goal, nil
}

// WithdrawContribution withdraws from a goal (creates withdrawal record)
func (s *goalService) WithdrawContribution(
	ctx context.Context,
	goalID uuid.UUID,
	amount float64,
	note *string,
	reversingID *uuid.UUID,
) (*domain.Goal, error) {
	if amount <= 0 {
		return nil, errors.New("withdrawal amount must be greater than 0")
	}

	// Get the goal
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to find goal for withdrawal",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Check if goal has enough amount
	if goal.CurrentAmount < amount {
		return nil, errors.New("insufficient goal balance for withdrawal")
	}

	// Create withdrawal record
	contribution := domain.NewWithdrawal(goalID, goal.AccountID, goal.UserID, amount, note, reversingID)
	contribution.Currency = goal.Currency

	if err := s.repo.CreateContribution(ctx, contribution); err != nil {
		s.logger.Error("Failed to create withdrawal record",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Update goal's current amount (subtract)
	goal.CurrentAmount -= amount
	goal.RemainingAmount = goal.TargetAmount - goal.CurrentAmount
	if goal.TargetAmount > 0 {
		goal.PercentageComplete = (goal.CurrentAmount / goal.TargetAmount) * 100
	}

	if err := s.repo.Update(ctx, goal); err != nil {
		s.logger.Error("Failed to update goal after withdrawal",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Update account's available balance
	netContributions, err := s.repo.GetNetContributionsByAccountID(ctx, goal.AccountID)
	if err != nil {
		s.logger.Error("Failed to get net contributions for account",
			zap.String("account_id", goal.AccountID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	if s.accountService != nil {
		// TODO: Implement RecalculateAvailableBalance in account service
		s.logger.Warn("Account balance update not yet fully implemented",
			zap.String("account_id", goal.AccountID.String()),
			zap.Float64("net_contributions", netContributions),
		)
	}

	s.logger.Info("Withdrawal from goal",
		zap.String("goal_id", goalID.String()),
		zap.String("name", goal.Name),
		zap.Float64("amount", amount),
		zap.Float64("new_current_amount", goal.CurrentAmount),
	)

	return goal, nil
}

// GetContributions retrieves all contributions for a goal
func (s *goalService) GetContributions(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error) {
	contributions, err := s.repo.FindContributionsByGoalID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to get goal contributions",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return contributions, nil
}

// GetGoalNetContributions calculates net contributions for a goal
func (s *goalService) GetGoalNetContributions(ctx context.Context, goalID uuid.UUID) (float64, error) {
	netAmount, err := s.repo.GetNetContributionsByGoalID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to get net contributions for goal",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return 0, err
	}

	return netAmount, nil
}
