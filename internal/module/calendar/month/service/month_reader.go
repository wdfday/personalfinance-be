package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"personalfinancedss/internal/module/calendar/month/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetMonth retrieves an existing month - does NOT create if missing
// Use this for accessing historical/past months
func (s *monthService) GetMonth(ctx context.Context, userID uuid.UUID, monthStr string) (*dto.MonthViewResponse, error) {
	s.logger.Info("getting month",
		zap.String("user_id", userID.String()),
		zap.String("month", monthStr),
	)

	// Try to get existing month
	monthEntity, err := s.repo.GetMonth(ctx, userID, monthStr)
	if err != nil {
		// Check if it's a "not found" error (from repository)
		errMsg := err.Error()
		if strings.Contains(errMsg, "month not found") || strings.Contains(errMsg, "record not found") {
			return nil, shared.ErrNotFound.WithDetails("month", monthStr)
		}
		return nil, fmt.Errorf("failed to get month: %w", err)
	}

	monthEntity.EnsureState()
	state := monthEntity.CurrentState()

	response := &dto.MonthViewResponse{
		MonthID:      monthEntity.ID,
		UserID:       monthEntity.UserID,
		Month:        monthEntity.Month,
		StartDate:    monthEntity.StartDate,
		EndDate:      monthEntity.EndDate,
		Status:       string(monthEntity.Status),
		ToBeBudgeted: state.ToBeBudgeted,
		Categories:   []dto.CategoryLineResponse{},
		CreatedAt:    monthEntity.CreatedAt,
		UpdatedAt:    monthEntity.UpdatedAt,
	}

	var totalBudgeted, totalActivity float64
	for i := range state.CategoryStates {
		catState := &state.CategoryStates[i]
		totalBudgeted += catState.Assigned
		totalActivity += catState.Activity

		var catName string
		if cat, err := s.categoryService.GetCategory(ctx, "", catState.CategoryID.String()); err == nil && cat != nil {
			catName = cat.Name
		} else {
			catName = "Unknown"
		}

		line := dto.CategoryLineResponse{
			CategoryID: catState.CategoryID,
			Name:       catName,
			Rollover:   catState.Rollover,
			Assigned:   catState.Assigned,
			Activity:   catState.Activity,
			Available:  catState.Available,
		}

		if catState.GoalTarget != nil {
			line.GoalTarget = catState.GoalTarget
		}
		if catState.DebtMinPayment != nil {
			line.DebtMinPayment = catState.DebtMinPayment
		}
		if catState.Notes != nil {
			line.Notes = catState.Notes
		}

		response.Categories = append(response.Categories, line)
	}

	response.Income = state.ActualIncome
	response.Budgeted = totalBudgeted
	response.Activity = totalActivity

	// Extract DSS Workflow data if available
	if state.DSSWorkflow != nil {
		response.DSSWorkflow = &dto.DSSWorkflowSummary{
			CurrentStep:    state.DSSWorkflow.CurrentStep,
			CompletedSteps: state.DSSWorkflow.CompletedSteps,
			IsComplete:     state.DSSWorkflow.IsComplete(),
			LastUpdated:    state.DSSWorkflow.LastUpdated,
		}

		// Goal Prioritization summary
		if state.DSSWorkflow.GoalPrioritization != nil {
			rankings := make([]dto.GoalRankingSummary, 0, len(state.DSSWorkflow.GoalPrioritization.Rankings))
			for _, r := range state.DSSWorkflow.GoalPrioritization.Rankings {
				rankings = append(rankings, dto.GoalRankingSummary{
					GoalID:          r.GoalID.String(),
					GoalName:        r.GoalName,
					Rank:            r.Rank,
					Score:           r.Score,
					SuggestedAmount: r.SuggestedAmount,
				})
			}
			response.DSSWorkflow.GoalPrioritization = &dto.GoalPrioritizationSummary{
				Method:   state.DSSWorkflow.GoalPrioritization.Method,
				Rankings: rankings,
			}
		}

		// Debt Strategy summary
		if state.DSSWorkflow.DebtStrategy != nil {
			paymentPlan := make([]dto.DebtPaymentSummary, 0)
			if state.DSSWorkflow.DebtStrategy.PaymentPlan != nil {
				for _, plan := range state.DSSWorkflow.DebtStrategy.PaymentPlan {
					paymentPlan = append(paymentPlan, dto.DebtPaymentSummary{
						DebtID:       plan.DebtID.String(),
						DebtName:     plan.DebtName,
						MinPayment:   plan.MinPayment,
						ExtraPayment: plan.ExtraPayment,
						TotalPayment: plan.TotalPayment,
					})
				}
			}
			response.DSSWorkflow.DebtStrategy = &dto.DebtStrategySummary{
				Strategy:    state.DSSWorkflow.DebtStrategy.Strategy,
				PaymentPlan: paymentPlan,
			}
		}

		// Budget Allocation summary
		if state.DSSWorkflow.BudgetAllocation != nil {
			categoryAllocations := make(map[string]float64)
			for catID, amount := range state.DSSWorkflow.BudgetAllocation.Recommendations {
				categoryAllocations[catID.String()] = amount
			}
			response.DSSWorkflow.BudgetAllocation = &dto.BudgetAllocationSummary{
				Method:              state.DSSWorkflow.BudgetAllocation.Method,
				OptimalityScore:     state.DSSWorkflow.BudgetAllocation.OptimalityScore,
				CategoryAllocations: categoryAllocations,
			}
		}
	}

	// Extract Goal and Debt allocations from DSS Budget Allocation
	// Map category allocations to goal_id/debt_id using CategoryStates (which have GoalTarget/DebtMinPayment)
	goalAllocations := make(map[string]float64)
	debtAllocations := make(map[string]float64)

	if state.DSSWorkflow != nil && state.DSSWorkflow.BudgetAllocation != nil && state.Input != nil {
		// Build category -> goal_id mapping: find category with GoalTarget matching a goal
		categoryToGoal := make(map[uuid.UUID]uuid.UUID)
		for _, catState := range state.CategoryStates {
			if catState.GoalTarget != nil {
				// Find goal with matching target amount (within 10% tolerance)
				for _, goal := range state.Input.Goals {
					goalTarget := *catState.GoalTarget
					goalTargetAmount := goal.TargetAmount
					if goalTargetAmount > 0 && goalTarget > 0 {
						diff := goalTargetAmount - goalTarget
						if diff < 0 {
							diff = -diff
						}
						if diff < (goalTargetAmount * 0.1) {
							categoryToGoal[catState.CategoryID] = goal.GoalID
							break // One category -> one goal
						}
					}
				}
			}
		}

		// Build category -> debt_id mapping: find category with DebtMinPayment matching a debt
		categoryToDebt := make(map[uuid.UUID]uuid.UUID)
		for _, catState := range state.CategoryStates {
			if catState.DebtMinPayment != nil {
				// Find debt with matching min payment (within 1000 VND tolerance)
				for _, debt := range state.Input.Debts {
					debtMinPayment := *catState.DebtMinPayment
					debtMinPay := debt.MinPayment
					if debtMinPay > 0 && debtMinPayment > 0 {
						diff := debtMinPay - debtMinPayment
						if diff < 0 {
							diff = -diff
						}
						if diff < 1000 {
							categoryToDebt[catState.CategoryID] = debt.DebtID
							break // One category -> one debt
						}
					}
				}
			}
		}

		// Map category allocations from DSS to goals/debts
		for catID, amount := range state.DSSWorkflow.BudgetAllocation.Recommendations {
			if goalID, ok := categoryToGoal[catID]; ok {
				goalAllocations[goalID.String()] += amount
			}
			if debtID, ok := categoryToDebt[catID]; ok {
				debtAllocations[debtID.String()] += amount
			}
		}
	}

	if len(goalAllocations) > 0 {
		response.GoalAllocations = goalAllocations
	}
	if len(debtAllocations) > 0 {
		response.DebtAllocations = debtAllocations
	}

	return response, nil
}

// ListMonths retrieves all months for a user
func (s *monthService) ListMonths(ctx context.Context, userID uuid.UUID) ([]*dto.MonthResponse, error) {
	s.logger.Info("listing months", zap.String("user_id", userID.String()))

	months, err := s.repo.ListMonths(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list months: %w", err)
	}

	response := make([]*dto.MonthResponse, 0, len(months))
	for _, month := range months {
		month.EnsureState()
		state := month.CurrentState()

		response = append(response, &dto.MonthResponse{
			MonthID:      month.ID,
			UserID:       month.UserID,
			Month:        month.Month,
			StartDate:    month.StartDate,
			EndDate:      month.EndDate,
			Status:       string(month.Status),
			ToBeBudgeted: state.ToBeBudgeted,
			CreatedAt:    month.CreatedAt,
			UpdatedAt:    month.UpdatedAt,
		})
	}

	return response, nil
}

// GetOrCreateCurrentMonth gets or creates the current month based on system date
// This enforces sequential workflow - users must close old months before working on new ones
func (s *monthService) GetOrCreateCurrentMonth(ctx context.Context, userID uuid.UUID) (*dto.MonthViewResponse, error) {
	currentMonthStr := time.Now().Format("2006-01")

	s.logger.Info("getting or creating current month",
		zap.String("user_id", userID.String()),
		zap.String("current_month_str", currentMonthStr),
	)

	// STEP 1: Check if there's any OPEN month (could be old month)
	// This enforces sequential workflow - must close old months first
	months, err := s.repo.ListMonths(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, month := range months {
		if month.Status == "open" {
			s.logger.Info("found OPEN month, returning it instead of creating new",
				zap.String("open_month", month.Month),
				zap.String("current_month_str", currentMonthStr),
				zap.Bool("is_old_month", month.Month != currentMonthStr),
			)
			// Return existing OPEN month (even if it's old)
			// User must close it before moving to current month
			return s.GetMonth(ctx, userID, month.Month)
		}
	}

	// STEP 2: No OPEN months exist, try to get current month
	view, err := s.GetMonth(ctx, userID, currentMonthStr)
	if err == nil {
		// Current month exists (closed), return it
		return view, nil
	}

	// STEP 3: Current month doesn't exist, create it (lazy initialization)
	s.logger.Info("no OPEN months found, creating current month",
		zap.String("month", currentMonthStr),
	)

	_, err = s.CreateMonth(ctx, userID, currentMonthStr)
	if err != nil {
		return nil, err
	}

	return s.GetMonth(ctx, userID, currentMonthStr)
}
