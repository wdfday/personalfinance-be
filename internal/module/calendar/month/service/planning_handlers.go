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

// RecalculatePlanning creates a new planning iteration by collecting fresh snapshots
// This APPENDS a new state to Month.States (does not modify existing states)
func (s *monthService) RecalculatePlanning(ctx context.Context, req dto.RecalculatePlanningRequest, userID *uuid.UUID) (*dto.PlanningIterationResponse, error) {
	s.logger.Info("recalculating planning",
		zap.String("month_id", req.MonthID.String()),
	)

	// 1. Load month
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}

	if !month.CanBeModified() {
		return nil, fmt.Errorf("month is %s and cannot be modified", month.Status)
	}

	month.EnsureState()
	currentState := month.CurrentState()

	// 2. Create new InputSnapshot by collecting fresh data
	inputSnapshot := s.collectInputSnapshot(ctx, month.UserID, req)

	// 3. Create new MonthState (iteration)
	newState := domain.MonthState{
		Version:        len(month.States) + 1,
		CreatedAt:      time.Now(),
		IsApplied:      false,
		ToBeBudgeted:   currentState.ToBeBudgeted,   // Carry over TBB
		ActualIncome:   currentState.ActualIncome,   // Carry over actual income
		CategoryStates: currentState.CategoryStates, // Carry over category states
		Input:          inputSnapshot,
	}

	// 4. Append new state to States array
	month.States = append(month.States, newState)

	// 5. Save to database
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return nil, fmt.Errorf("failed to save planning iteration: %w", err)
	}

	s.logger.Info("planning iteration created",
		zap.Int("version", newState.Version),
		zap.Int("total_iterations", len(month.States)),
	)

	// 6. Build response
	var totals domain.InputTotals
	if inputSnapshot != nil {
		inputSnapshot.CalculateTotals()
		totals = inputSnapshot.Totals
	}

	return &dto.PlanningIterationResponse{
		MonthID:          month.ID,
		Version:          newState.Version,
		Total:            len(month.States),
		IsLatest:         true,
		ProjectedIncome:  totals.ProjectedIncome,
		TotalConstraints: totals.TotalConstraints,
		TotalGoalTargets: totals.TotalGoalTargets,
		TotalDebtMinimum: totals.TotalDebtMinimum,
		TotalAdHoc:       totals.TotalAdHoc,
		Disposable:       totals.Disposable,
		ToBeBudgeted:     newState.ToBeBudgeted,
		CreatedAt:        newState.CreatedAt,
	}, nil
}

// collectInputSnapshot gathers fresh data from external services
func (s *monthService) collectInputSnapshot(ctx context.Context, userID uuid.UUID, req dto.RecalculatePlanningRequest) *domain.InputSnapshot {
	snapshot := domain.NewInputSnapshot()
	snapshot.CapturedAt = time.Now()

	// TODO: Query IncomeProfileService to get recurring income profiles
	// For now, use placeholder or override
	if req.ProjectedIncomeOverride != nil {
		snapshot.IncomeProfiles = append(snapshot.IncomeProfiles, domain.IncomeSnapshot{
			Name:        "Override",
			Amount:      *req.ProjectedIncomeOverride,
			IsRecurring: true,
			SourceType:  "override",
		})
	}

	// TODO: Query ConstraintService to get budget constraints
	// TODO: Query GoalService to get goals (filtered by req.SelectedGoalIDs)
	// TODO: Query DebtService to get debts (filtered by req.SelectedDebtIDs)

	// Process ad-hoc expenses from request
	for _, adhoc := range req.AdHocExpenses {
		snapshot.AdHocExpenses = append(snapshot.AdHocExpenses, domain.AdHocExpense{
			ID:           uuid.New(),
			Name:         adhoc.Name,
			Amount:       adhoc.Amount,
			CategoryID:   adhoc.CategoryID,
			CategoryHint: adhoc.CategoryHint,
			Notes:        adhoc.Notes,
		})
	}

	// Process ad-hoc income from request
	for _, adhoc := range req.AdHocIncome {
		snapshot.AdHocIncome = append(snapshot.AdHocIncome, domain.AdHocIncome{
			ID:     uuid.New(),
			Name:   adhoc.Name,
			Amount: adhoc.Amount,
			Notes:  adhoc.Notes,
		})
	}

	// Calculate totals
	snapshot.CalculateTotals()

	return snapshot
}
