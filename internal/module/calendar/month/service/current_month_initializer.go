package service

import (
	"context"
	"time"

	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
