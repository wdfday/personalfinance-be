package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"personalfinancedss/internal/module/calendar/event/domain"
	"personalfinancedss/internal/module/calendar/event/repository"
)

// Holiday represents a Vietnam holiday definition
type Holiday struct {
	Date time.Time
	Name string
}

// HolidayGeneratorService handles automatic generation of Vietnam holidays
type HolidayGeneratorService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewHolidayGeneratorService creates a new holiday generator service
func NewHolidayGeneratorService(repo repository.Repository, logger *zap.Logger) *HolidayGeneratorService {
	return &HolidayGeneratorService{
		repo:   repo,
		logger: logger.Named("calendar.event.holiday_generator"),
	}
}

// GenerateForNewUser generates holidays for a newly registered user
// Generates current year + next year
func (g *HolidayGeneratorService) GenerateForNewUser(ctx context.Context, userID uuid.UUID) error {
	currentYear := time.Now().Year()

	if err := g.GenerateForYear(ctx, userID, currentYear); err != nil {
		g.logger.Error("failed to generate holidays for current year",
			zap.String("user_id", userID.String()),
			zap.Int("year", currentYear),
			zap.Error(err))
		return err
	}

	if err := g.GenerateForYear(ctx, userID, currentYear+1); err != nil {
		g.logger.Error("failed to generate holidays for next year",
			zap.String("user_id", userID.String()),
			zap.Int("year", currentYear+1),
			zap.Error(err))
		return err
	}

	g.logger.Info("successfully generated holidays for new user",
		zap.String("user_id", userID.String()),
		zap.Int("current_year", currentYear),
		zap.Int("next_year", currentYear+1))

	return nil
}

// GenerateForYear generates all Vietnam holidays for a specific year and user
func (g *HolidayGeneratorService) GenerateForYear(ctx context.Context, userID uuid.UUID, year int) error {
	holidays := getVietnamHolidays(year)

	redColor := "#FF5733" // Red for holidays

	for _, h := range holidays {
		// Start at 00:00:00
		startDate := time.Date(h.Date.Year(), h.Date.Month(), h.Date.Day(), 0, 0, 0, 0, time.UTC)
		// End at 23:59:59
		endDate := time.Date(h.Date.Year(), h.Date.Month(), h.Date.Day(), 23, 59, 59, 0, time.UTC)

		event := &domain.Event{
			ID:             uuid.New(),
			UserID:         userID,
			Name:           h.Name,
			Type:           domain.EventTypeHoliday,
			Source:         domain.SourceSystemGenerated,
			StartDate:      startDate,
			EndDate:        &endDate,
			AllDay:         true,
			Color:          &redColor,
			IsRecurring:    false,
			RecurrenceType: nil,
		}

		// Check if holiday already exists to avoid duplicates
		exists, err := g.repo.CheckHolidayExists(ctx, userID, startDate, h.Name)
		if err != nil {
			g.logger.Error("failed to check holiday existence",
				zap.String("user_id", userID.String()),
				zap.String("holiday", h.Name),
				zap.Error(err))
			continue
		}

		if exists {
			g.logger.Debug("holiday already exists, skipping",
				zap.String("user_id", userID.String()),
				zap.String("holiday", h.Name),
				zap.Int("year", year))
			continue
		}

		if err := g.repo.Create(ctx, event); err != nil {
			g.logger.Error("failed to create holiday event",
				zap.String("user_id", userID.String()),
				zap.String("holiday", h.Name),
				zap.Error(err))
			continue
		}

		g.logger.Debug("created holiday event",
			zap.String("user_id", userID.String()),
			zap.String("holiday", h.Name),
			zap.Time("date", startDate))
	}

	return nil
}

// GenerateUpcomingYear generates holidays for the next year if we're approaching year end
// Should be called by a scheduler monthly or quarterly
func (g *HolidayGeneratorService) GenerateUpcomingYear(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	nextYear := now.Year() + 1

	// Only generate if we're in Q4 (October onwards)
	if now.Month() >= time.October {
		return g.GenerateForYear(ctx, userID, nextYear)
	}

	return nil
}

// getVietnamHolidays returns all Vietnam holidays for a given year
func getVietnamHolidays(year int) []Holiday {
	holidays := []Holiday{
		// Fixed holidays (Dương lịch)
		{
			Date: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			Name: "Tết Dương Lịch",
		},
		{
			Date: time.Date(year, time.April, 30, 0, 0, 0, 0, time.UTC),
			Name: "Ngày Giải Phóng Miền Nam",
		},
		{
			Date: time.Date(year, time.May, 1, 0, 0, 0, 0, time.UTC),
			Name: "Ngày Quốc Tế Lao Động",
		},
		{
			Date: time.Date(year, time.September, 2, 0, 0, 0, 0, time.UTC),
			Name: "Quốc Khánh Việt Nam",
		},
	}

	// Lunar holidays (Âm lịch) - hardcoded dates for common years
	// Note: These are approximate and should be verified for each year
	lunarHolidays := getLunarHolidays(year)
	holidays = append(holidays, lunarHolidays...)

	return holidays
}

// getLunarHolidays returns lunar calendar based holidays
// Hardcoded for 2024-2030 (most common usage period)
func getLunarHolidays(year int) []Holiday {
	var holidays []Holiday

	// Tết Nguyên Đán dates (Mùng 1 Tết)
	// Source: Vietnam official calendar
	tetDates := map[int]time.Time{
		2024: time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC),
		2025: time.Date(2025, time.January, 29, 0, 0, 0, 0, time.UTC),
		2026: time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC),
		2027: time.Date(2027, time.February, 6, 0, 0, 0, 0, time.UTC),
		2028: time.Date(2028, time.January, 26, 0, 0, 0, 0, time.UTC),
		2029: time.Date(2029, time.February, 13, 0, 0, 0, 0, time.UTC),
		2030: time.Date(2030, time.February, 3, 0, 0, 0, 0, time.UTC),
	}

	if tetDate, ok := tetDates[year]; ok {
		// Tết usually has 5 days off: 29 Tết to Mùng 3
		// Add multiple days of Tết celebration
		holidays = append(holidays, Holiday{
			Date: tetDate.AddDate(0, 0, -1), // 29 Tết
			Name: "29 Tết",
		})
		holidays = append(holidays, Holiday{
			Date: tetDate, // Mùng 1 Tết
			Name: "Tết Nguyên Đán",
		})
		holidays = append(holidays, Holiday{
			Date: tetDate.AddDate(0, 0, 1), // Mùng 2
			Name: "Mùng 2 Tết",
		})
		holidays = append(holidays, Holiday{
			Date: tetDate.AddDate(0, 0, 2), // Mùng 3
			Name: "Mùng 3 Tết",
		})
		holidays = append(holidays, Holiday{
			Date: tetDate.AddDate(0, 0, 3), // Mùng 4
			Name: "Mùng 4 Tết",
		})
	}

	// Giỗ Tổ Hùng Vương (10/3 Âm lịch)
	hungVuongDates := map[int]time.Time{
		2024: time.Date(2024, time.April, 18, 0, 0, 0, 0, time.UTC),
		2025: time.Date(2025, time.April, 7, 0, 0, 0, 0, time.UTC),
		2026: time.Date(2026, time.March, 28, 0, 0, 0, 0, time.UTC),
		2027: time.Date(2027, time.April, 16, 0, 0, 0, 0, time.UTC),
		2028: time.Date(2028, time.April, 4, 0, 0, 0, 0, time.UTC),
		2029: time.Date(2029, time.March, 24, 0, 0, 0, 0, time.UTC),
		2030: time.Date(2030, time.April, 13, 0, 0, 0, 0, time.UTC),
	}

	if hungVuongDate, ok := hungVuongDates[year]; ok {
		holidays = append(holidays, Holiday{
			Date: hungVuongDate,
			Name: "Giỗ Tổ Hùng Vương",
		})
	}

	return holidays
}

