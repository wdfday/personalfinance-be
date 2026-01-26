package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/calendar/month/domain"
	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateMonth creates a new month with initial state
func (s *monthService) CreateMonth(ctx context.Context, userID uuid.UUID, monthStr string) (*domain.Month, error) {
	s.logger.Info("creating new month",
		zap.String("user_id", userID.String()),
		zap.String("month", monthStr),
	)

	// 1. Check if month already exists
	existing, err := s.repo.GetMonth(ctx, userID, monthStr)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("month already exists: %s", monthStr)
	}

	// 2. Calculate period dates from monthStr (format: "2006-01")
	t, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid month format: %s", monthStr)
	}
	startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	// 3. Create Month entity with EMPTY States array
	// States will be populated by SaveDSS after frontend collects input and runs DSS
	newMonth := &domain.Month{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		Month:     monthStr,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    domain.StatusOpen,
		States:    []domain.MonthState{}, // Empty - filled by SaveDSS
		Version:   1,
	}

	// 4. Save to database
	if err := s.repo.CreateMonth(ctx, newMonth); err != nil {
		return nil, fmt.Errorf("failed to create month: %w", err)
	}

	s.logger.Info("month created successfully (empty state, waiting for DSS)",
		zap.String("month_id", newMonth.ID.String()),
		zap.String("month", monthStr),
	)

	return newMonth, nil
}

// ReceiveIncome adds income to To Be Budgeted
func (s *monthService) ReceiveIncome(ctx context.Context, req dto.IncomeReceivedRequest, userID *uuid.UUID) error {
	s.logger.Info("receiving income",
		zap.String("month_id", req.MonthID.String()),
		zap.Float64("amount", req.Amount),
	)

	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}

	if !month.CanBeModified() {
		return fmt.Errorf("month is %s and cannot be modified", month.Status)
	}

	month.EnsureState()
	state := month.CurrentState()
	state.ToBeBudgeted += req.Amount
	state.ActualIncome += req.Amount

	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to update month: %w", err)
	}

	s.logger.Info("income received successfully",
		zap.Float64("new_tbb", state.ToBeBudgeted),
	)

	return nil
}

// CloseMonth freezes the month for reporting and calculates closed snapshot
func (s *monthService) CloseMonth(ctx context.Context, userID uuid.UUID, monthStr string) error {
	s.logger.Info("closing month",
		zap.String("user_id", userID.String()),
		zap.String("month", monthStr),
	)

	// Get month by user + month string
	month, err := s.repo.GetMonth(ctx, userID, monthStr)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}

	// Use the new CloseWithSnapshot method that calculates and stores snapshot
	snapshot, err := month.CloseWithSnapshot(&userID, false)
	if err != nil {
		return fmt.Errorf("failed to close month: %w", err)
	}

	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to update month: %w", err)
	}

	s.logger.Info("month closed successfully",
		zap.Float64("final_tbb", snapshot.FinalTBB),
		zap.Float64("total_income", snapshot.TotalIncomeReceived),
		zap.Float64("total_spent", snapshot.TotalSpent),
		zap.Float64("net_savings", snapshot.NetSavings),
		zap.Int("overspent_count", len(snapshot.OverspentCategories)),
	)

	// Auto-create next month with carryover data
	if err := s.autoCreateNextMonth(ctx, month, &userID); err != nil {
		s.logger.Warn("failed to auto-create next month",
			zap.Error(err),
			zap.String("month", month.Month),
		)
	}

	return nil
}

// autoCreateNextMonth automatically creates the next month after closing
// It calculates the next month string and delegates to CreateMonth
func (s *monthService) autoCreateNextMonth(ctx context.Context, closedMonth *domain.Month, userID *uuid.UUID) error {
	nextPeriodStart := closedMonth.EndDate.AddDate(0, 0, 1)
	nextMonthStr := nextPeriodStart.Format("2006-01")

	s.logger.Info("auto-creating next month after close",
		zap.String("closed_month", closedMonth.Month),
		zap.String("next_month", nextMonthStr),
	)

	// CreateMonth handles duplicate check and returns error if already exists
	// We ignore "already exists" errors as it's safe
	_, err := s.CreateMonth(ctx, *userID, nextMonthStr)
	if err != nil {
		// Check if error is "already exists" - this is safe to ignore
		if err.Error() == fmt.Sprintf("month already exists: %s", nextMonthStr) {
			s.logger.Info("next month already exists, skipping",
				zap.String("next_month", nextMonthStr),
			)
			return nil
		}
		return fmt.Errorf("failed to create next month: %w", err)
	}

	s.logger.Info("next month auto-created successfully",
		zap.String("next_month", nextMonthStr),
	)

	return nil
}
