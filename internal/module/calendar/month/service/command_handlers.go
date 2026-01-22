package service

import (
	"context"
	"fmt"

	"personalfinancedss/internal/module/calendar/month/domain"
	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AssignCategory assigns money to a category
func (s *monthService) AssignCategory(ctx context.Context, req dto.AssignCategoryRequest, userID *uuid.UUID) error {
	s.logger.Info("assigning category",
		zap.String("month_id", req.MonthID.String()),
		zap.String("category_id", req.CategoryID.String()),
		zap.Float64("amount", req.Amount),
	)

	// 1. Load month
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}

	// 2. Check if month can be modified
	if !month.CanBeModified() {
		return fmt.Errorf("month is %s and cannot be modified", month.Status)
	}

	// 3. Ensure state is initialized
	month.EnsureState()
	state := month.CurrentState()

	// 4. Get or create category state
	var catState *domain.CategoryState
	for i := range state.CategoryStates {
		if state.CategoryStates[i].CategoryID == req.CategoryID {
			catState = &state.CategoryStates[i]
			break
		}
	}
	if catState == nil {
		// Create new category state
		newCat := domain.NewCategoryState(req.CategoryID, 0, 0, 0)
		state.CategoryStates = append(state.CategoryStates, *newCat)
		catState = &state.CategoryStates[len(state.CategoryStates)-1]
	}

	// 5. Update category assigned amount
	oldAssigned := catState.Assigned
	catState.AddAssignment(req.Amount)
	newAssigned := catState.Assigned

	// 6. Update To Be Budgeted (TBB decreases when money is assigned)
	state.ToBeBudgeted -= req.Amount

	// 7. Save updated month
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to update month: %w", err)
	}

	// 8. Log success
	s.logger.Info("category assigned successfully",
		zap.Float64("old_assigned", oldAssigned),
		zap.Float64("new_assigned", newAssigned),
		zap.Float64("new_tbb", state.ToBeBudgeted),
	)

	return nil
}

// MoveMoney moves money between categories
func (s *monthService) MoveMoney(ctx context.Context, req dto.MoveMoneyRequest, userID *uuid.UUID) error {
	s.logger.Info("moving money",
		zap.String("month_id", req.MonthID.String()),
		zap.String("from_category", req.FromCategoryID.String()),
		zap.String("to_category", req.ToCategoryID.String()),
		zap.Float64("amount", req.Amount),
	)

	if req.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}

	if !month.CanBeModified() {
		return fmt.Errorf("month is %s and cannot be modified", month.Status)
	}

	month.EnsureState()
	state := month.CurrentState()

	// Find source category
	var fromCat *domain.CategoryState
	for i := range state.CategoryStates {
		if state.CategoryStates[i].CategoryID == req.FromCategoryID {
			fromCat = &state.CategoryStates[i]
			break
		}
	}
	if fromCat == nil {
		return fmt.Errorf("source category not found in month")
	}

	// Find or create target category
	var toCat *domain.CategoryState
	for i := range state.CategoryStates {
		if state.CategoryStates[i].CategoryID == req.ToCategoryID {
			toCat = &state.CategoryStates[i]
			break
		}
	}
	if toCat == nil {
		newCat := domain.NewCategoryState(req.ToCategoryID, 0, 0, 0)
		state.CategoryStates = append(state.CategoryStates, *newCat)
		toCat = &state.CategoryStates[len(state.CategoryStates)-1]
	}

	fromCat.AddAssignment(-req.Amount)
	toCat.AddAssignment(req.Amount)

	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to update month: %w", err)
	}

	s.logger.Info("money moved successfully",
		zap.Float64("from_available", fromCat.Available),
		zap.Float64("to_available", toCat.Available),
	)

	return nil
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
