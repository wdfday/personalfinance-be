package service

import (
	"context"
	"fmt"
	transactiondto "personalfinancedss/internal/module/cashflow/transaction/dto"
	transactionrepo "personalfinancedss/internal/module/cashflow/transaction/repository"
	userrepo "personalfinancedss/internal/module/identify/user/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type scheduledReportService struct {
	transactionRepo transactionrepo.Repository
	userRepo        userrepo.Repository
	emailService    EmailService
	logger          *zap.Logger
}

// NewScheduledReportService creates a new scheduled report service
func NewScheduledReportService(
	transactionRepo transactionrepo.Repository,
	userRepo userrepo.Repository,
	emailService EmailService,
	logger *zap.Logger,
) ScheduledReportService {
	return &scheduledReportService{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		emailService:    emailService,
		logger:          logger,
	}
}

func (s *scheduledReportService) GenerateDailyReport(ctx context.Context, userID string) error {
	s.logger.Info("Generating daily report", zap.String("user_id", userID))

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, userUUID.String())
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	// Calculate date range (today)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Get transaction summary for today
	query := transactiondto.ListTransactionsQuery{
		StartBookingDate: &startOfDay,
		EndBookingDate:   &endOfDay,
	}

	summary, err := s.transactionRepo.GetSummary(ctx, userUUID, query)
	if err != nil {
		s.logger.Error("Failed to get transaction summary", zap.Error(err))
		return err
	}

	// Prepare email data
	emailData := map[string]interface{}{
		"Date":             now.Format("January 2, 2006"),
		"TotalIncome":      summary.TotalCredit, // Credit = income
		"TotalExpense":     summary.TotalDebit,  // Debit = expenses
		"NetAmount":        summary.NetAmount,
		"TransactionCount": summary.Count,
	}

	// Send email (using existing monthly summary template, can be customized later)
	return s.emailService.SendCustomEmail(
		user.Email,
		fmt.Sprintf("Daily Summary - %s", now.Format("Jan 2, 2006")),
		"monthly_summary.html",
		emailData,
	)
}

func (s *scheduledReportService) GenerateWeeklyReport(ctx context.Context, userID string) error {
	s.logger.Info("Generating weekly report", zap.String("user_id", userID))

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, userUUID.String())
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	// Calculate date range (last 7 days)
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -7)

	// Get transaction summary for the week
	query := transactiondto.ListTransactionsQuery{
		StartBookingDate: &startOfWeek,
		EndBookingDate:   &now,
	}

	summary, err := s.transactionRepo.GetSummary(ctx, userUUID, query)
	if err != nil {
		s.logger.Error("Failed to get transaction summary", zap.Error(err))
		return err
	}

	// Prepare email data
	emailData := map[string]interface{}{
		"Period":           fmt.Sprintf("%s - %s", startOfWeek.Format("Jan 2"), now.Format("Jan 2, 2006")),
		"TotalIncome":      summary.TotalCredit, // Credit = income
		"TotalExpense":     summary.TotalDebit,  // Debit = expenses
		"NetAmount":        summary.NetAmount,
		"TransactionCount": summary.Count,
		"AvgDailyExpense":  float64(summary.TotalDebit) / 7,
	}

	// Send email
	return s.emailService.SendCustomEmail(
		user.Email,
		fmt.Sprintf("Weekly Summary - %s", now.Format("Jan 2, 2006")),
		"monthly_summary.html",
		emailData,
	)
}

func (s *scheduledReportService) GenerateMonthlySummary(ctx context.Context, userID string) error {
	s.logger.Info("Generating monthly summary", zap.String("user_id", userID))

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, userUUID.String())
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	// Calculate date range (current month)
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Get transaction summary for the month
	query := transactiondto.ListTransactionsQuery{
		StartBookingDate: &startOfMonth,
		EndBookingDate:   &endOfMonth,
	}

	summary, err := s.transactionRepo.GetSummary(ctx, userUUID, query)
	if err != nil {
		s.logger.Error("Failed to get transaction summary", zap.Error(err))
		return err
	}

	// Prepare summary data
	summaryData := map[string]interface{}{
		"total_income":      summary.TotalCredit, // Credit = income
		"total_expense":     summary.TotalDebit,  // Debit = expenses
		"net_amount":        summary.NetAmount,
		"transaction_count": summary.Count,
	}

	// Use existing SendMonthlySummary method
	return s.emailService.SendMonthlySummary(
		user.Email,
		user.FullName,
		now.Format("January"),
		now.Year(),
		summaryData,
	)
}

func (s *scheduledReportService) GenerateCustomReport(ctx context.Context, userID string, startDate, endDate time.Time) error {
	s.logger.Info("Generating custom report",
		zap.String("user_id", userID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, userUUID.String())
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	// Get transaction summary for the custom date range
	query := transactiondto.ListTransactionsQuery{
		StartBookingDate: &startDate,
		EndBookingDate:   &endDate,
	}

	summary, err := s.transactionRepo.GetSummary(ctx, userUUID, query)
	if err != nil {
		s.logger.Error("Failed to get transaction summary", zap.Error(err))
		return err
	}

	// Calculate number of days
	days := endDate.Sub(startDate).Hours() / 24

	// Prepare email data
	emailData := map[string]interface{}{
		"Period":           fmt.Sprintf("%s - %s", startDate.Format("Jan 2, 2006"), endDate.Format("Jan 2, 2006")),
		"TotalIncome":      summary.TotalCredit, // Credit = income
		"TotalExpense":     summary.TotalDebit,  // Debit = expenses
		"NetAmount":        summary.NetAmount,
		"TransactionCount": summary.Count,
		"Days":             int(days),
		"AvgDailyExpense":  float64(summary.TotalDebit) / days,
	}

	// Send email
	return s.emailService.SendCustomEmail(
		user.Email,
		fmt.Sprintf("Custom Report - %s to %s", startDate.Format("Jan 2"), endDate.Format("Jan 2, 2006")),
		"monthly_summary.html",
		emailData,
	)
}
