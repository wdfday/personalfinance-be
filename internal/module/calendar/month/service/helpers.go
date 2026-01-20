package service

import (
	"fmt"
	"time"
)

// calculateMonthBoundaries calculates start and end dates for a calendar month
// Input: "2024-02" or "2024-2"
// Output: (2024-02-01, 2024-02-29) for February 2024
func calculateMonthBoundaries(monthStr string) (time.Time, time.Time) {
	// Parse month string (YYYY-MM format)
	startDate, err := time.Parse("2006-01", monthStr)
	if err != nil {
		// Try with single digit month
		startDate, err = time.Parse("2006-1", monthStr)
		if err != nil {
			// Fallback to current month
			now := time.Now()
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		}
	}

	// Ensure it's the first day of the month
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Calculate last day of month
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second) // Last second of the month
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	return startDate, endDate
}

// calculatePayPeriod calculates start and end dates for a pay period
// This will be used when user has custom budget period settings
func calculatePayPeriod(baseDate time.Time, startDay int, lengthDays int) (time.Time, time.Time) {
	// Find the current or next period start date
	year := baseDate.Year()
	month := baseDate.Month()

	// Start from the specified day of current month
	startDate := time.Date(year, month, startDay, 0, 0, 0, 0, time.UTC)

	// If we're past the start day, move to next period
	if baseDate.After(startDate.AddDate(0, 0, lengthDays)) {
		// Move to next month's start day
		startDate = startDate.AddDate(0, 1, 0)
	}

	endDate := startDate.AddDate(0, 0, lengthDays-1)

	return startDate, endDate
}

// formatMonthDisplay generates a display name for a month period
// Examples: "2024-02", "Pay Period 2/15-3/14", "Q1 2024"
func formatMonthDisplay(startDate, endDate time.Time, periodType string) string {
	switch periodType {
	case "calendar_month":
		return startDate.Format("2006-01")
	case "pay_period":
		return fmt.Sprintf("Pay Period %s-%s",
			startDate.Format("1/2"),
			endDate.Format("1/2"))
	case "custom":
		return fmt.Sprintf("%s to %s",
			startDate.Format("Jan 2"),
			endDate.Format("Jan 2"))
	default:
		return startDate.Format("2006-01")
	}
}
