package service

import (
	"context"
	"fmt"

	"personalfinancedss/internal/module/calendar/month/domain"

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

	// 2. Calculate period dates
	startDate, endDate := calculateMonthBoundaries(monthStr)

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
