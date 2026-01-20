package service

import (
	"context"
	"fmt"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/dto"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetMonthContributions retrieves all contributions for a goal within a date range
func (s *goalService) GetMonthContributions(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]domain.GoalContribution, error) {
	s.logger.Info("getting goal contributions for month",
		zap.String("goal_id", goalID.String()),
		zap.Time("start", startDate),
		zap.Time("end", endDate),
	)

	contributions, err := s.repo.GetContributionsByDateRange(ctx, goalID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions by date range: %w", err)
	}

	return contributions, nil
}

// GetMonthSummary calculates summary statistics for a goal's contributions in a month
func (s *goalService) GetMonthSummary(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) (*dto.GoalMonthlySummary, error) {
	s.logger.Info("calculating goal month summary",
		zap.String("goal_id", goalID.String()),
		zap.Time("start", startDate),
		zap.Time("end", endDate),
	)

	// Get goal details
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	// Get contributions in date range
	contributions, err := s.repo.GetContributionsByDateRange(ctx, goalID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}

	// Calculate totals
	totalDeposits := 0.0
	totalWithdrawals := 0.0

	for _, c := range contributions {
		if c.Type == domain.ContributionTypeDeposit {
			totalDeposits += c.Amount
		} else {
			totalWithdrawals += c.Amount
		}
	}

	netContributed := totalDeposits - totalWithdrawals

	s.logger.Info("goal month summary calculated",
		zap.String("goal_id", goalID.String()),
		zap.Float64("total_contributed", netContributed),
		zap.Int("contribution_count", len(contributions)),
	)

	return &dto.GoalMonthlySummary{
		GoalID:            goalID,
		Name:              goal.Name,
		TotalContributed:  netContributed,
		ContributionCount: len(contributions),
	}, nil
}

// GetAllTimeSummary calculates all-time statistics for a goal from inception to present
func (s *goalService) GetAllTimeSummary(ctx context.Context, goalID uuid.UUID) (*dto.GoalAllTimeSummary, error) {
	s.logger.Info("calculating goal all-time summary",
		zap.String("goal_id", goalID.String()),
	)

	// Get goal details
	goal, err := s.repo.FindByID(ctx, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	// Get all contributions (no date filter)
	contributions, err := s.repo.FindContributionsByGoalID(ctx, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all contributions: %w", err)
	}

	// Calculate totals
	totalDeposits := 0.0
	totalWithdrawals := 0.0
	var firstContribution, lastContribution *time.Time

	for i, c := range contributions {
		if c.Type == domain.ContributionTypeDeposit {
			totalDeposits += c.Amount
		} else {
			totalWithdrawals += c.Amount
		}

		// Track first and last contribution dates
		if i == len(contributions)-1 { // Contributions are ordered DESC, so last item is first chronologically
			firstContribution = &c.CreatedAt
		}
		if i == 0 { // First item is most recent
			lastContribution = &c.CreatedAt
		}
	}

	netContributed := totalDeposits - totalWithdrawals

	s.logger.Info("goal all-time summary calculated",
		zap.String("goal_id", goalID.String()),
		zap.Float64("net_contributed", netContributed),
		zap.Int("total_contributions", len(contributions)),
	)

	return &dto.GoalAllTimeSummary{
		GoalID:            goalID,
		Name:              goal.Name,
		TotalContributed:  totalDeposits,
		TotalWithdrawn:    totalWithdrawals,
		NetContributed:    netContributed,
		ContributionCount: len(contributions),
		FirstContribution: firstContribution,
		LastContribution:  lastContribution,
	}, nil
}
