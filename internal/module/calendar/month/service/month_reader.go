package service

import (
	"context"
	"fmt"

	"personalfinancedss/internal/module/calendar/month/dto"

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
		return nil, fmt.Errorf("month not found: %w", err)
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
	for categoryID, catState := range state.CategoryStates {
		totalBudgeted += catState.Assigned
		totalActivity += catState.Activity

		var catName string
		if cat, err := s.categoryService.GetCategory(ctx, "", categoryID.String()); err == nil && cat != nil {
			catName = cat.Name
		} else {
			catName = "Unknown"
		}

		line := dto.CategoryLineResponse{
			CategoryID: categoryID,
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
