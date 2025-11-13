package service

import (
	"context"
	userrepo "personalfinancedss/internal/module/identify/user/repository"
	"personalfinancedss/internal/module/notification/repository"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type schedulerService struct {
	cron          *cron.Cron
	reportService ScheduledReportService
	alertRuleRepo repository.AlertRuleRepository
	userRepo      userrepo.Repository
	logger        *zap.Logger
	isRunning     bool
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(
	reportService ScheduledReportService,
	alertRuleRepo repository.AlertRuleRepository,
	userRepo userrepo.Repository,
	logger *zap.Logger,
) SchedulerService {
	// Create cron instance with second precision
	c := cron.New(cron.WithSeconds())

	return &schedulerService{
		cron:          c,
		reportService: reportService,
		alertRuleRepo: alertRuleRepo,
		userRepo:      userRepo,
		logger:        logger,
		isRunning:     false,
	}
}

func (s *schedulerService) Start() {
	if s.isRunning {
		s.logger.Warn("Scheduler is already running")
		return
	}

	s.logger.Info("Starting notification scheduler")

	// Schedule daily reports at 9:00 AM every day
	_, err := s.cron.AddFunc("0 0 9 * * *", s.runDailyReports)
	if err != nil {
		s.logger.Error("Failed to schedule daily reports", zap.Error(err))
	}

	// Schedule weekly reports at 9:00 AM every Monday
	_, err = s.cron.AddFunc("0 0 9 * * 1", s.runWeeklyReports)
	if err != nil {
		s.logger.Error("Failed to schedule weekly reports", zap.Error(err))
	}

	// Schedule monthly reports at 9:00 AM on the 1st of each month
	_, err = s.cron.AddFunc("0 0 9 1 * *", s.runMonthlyReports)
	if err != nil {
		s.logger.Error("Failed to schedule monthly reports", zap.Error(err))
	}

	// Check for scheduled alert rules every 5 minutes
	_, err = s.cron.AddFunc("0 */5 * * * *", s.checkScheduledAlertRules)
	if err != nil {
		s.logger.Error("Failed to schedule alert rule checker", zap.Error(err))
	}

	// Start the cron scheduler
	s.cron.Start()
	s.isRunning = true

	s.logger.Info("Notification scheduler started successfully",
		zap.Int("total_jobs", len(s.cron.Entries())),
	)
}

func (s *schedulerService) Stop() {
	if !s.isRunning {
		s.logger.Warn("Scheduler is not running")
		return
	}

	s.logger.Info("Stopping notification scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.isRunning = false
	s.logger.Info("Notification scheduler stopped")
}

func (s *schedulerService) IsRunning() bool {
	return s.isRunning
}

// runDailyReports generates daily reports for all users who have daily report rules
func (s *schedulerService) runDailyReports() {
	s.logger.Info("Running daily reports job")
	ctx := context.Background()

	// Get all alert rules with scheduled daily reports
	rules, err := s.alertRuleRepo.ListScheduled(ctx)
	if err != nil {
		s.logger.Error("Failed to list scheduled rules", zap.Error(err))
		return
	}

	// Filter for daily report rules
	for _, rule := range rules {
		if rule.Type == "scheduled_report" {
			// Check if it's a daily report
			if schedule, ok := rule.Conditions["frequency"].(string); ok && schedule == "daily" {
				if err := s.reportService.GenerateDailyReport(ctx, rule.UserID.String()); err != nil {
					s.logger.Error("Failed to generate daily report",
						zap.String("user_id", rule.UserID.String()),
						zap.Error(err),
					)
				}
			}
		}
	}
}

// runWeeklyReports generates weekly reports for all users who have weekly report rules
func (s *schedulerService) runWeeklyReports() {
	s.logger.Info("Running weekly reports job")
	ctx := context.Background()

	// Get all alert rules with scheduled weekly reports
	rules, err := s.alertRuleRepo.ListScheduled(ctx)
	if err != nil {
		s.logger.Error("Failed to list scheduled rules", zap.Error(err))
		return
	}

	// Filter for weekly report rules
	for _, rule := range rules {
		if rule.Type == "scheduled_report" {
			// Check if it's a weekly report
			if schedule, ok := rule.Conditions["frequency"].(string); ok && schedule == "weekly" {
				if err := s.reportService.GenerateWeeklyReport(ctx, rule.UserID.String()); err != nil {
					s.logger.Error("Failed to generate weekly report",
						zap.String("user_id", rule.UserID.String()),
						zap.Error(err),
					)
				}
			}
		}
	}
}

// runMonthlyReports generates monthly reports for all users who have monthly report rules
func (s *schedulerService) runMonthlyReports() {
	s.logger.Info("Running monthly reports job")
	ctx := context.Background()

	// Get all alert rules with scheduled monthly reports
	rules, err := s.alertRuleRepo.ListScheduled(ctx)
	if err != nil {
		s.logger.Error("Failed to list scheduled rules", zap.Error(err))
		return
	}

	// Filter for monthly report rules
	for _, rule := range rules {
		if rule.Type == "scheduled_report" {
			// Check if it's a monthly report
			if schedule, ok := rule.Conditions["frequency"].(string); ok && schedule == "monthly" {
				if err := s.reportService.GenerateMonthlySummary(ctx, rule.UserID.String()); err != nil {
					s.logger.Error("Failed to generate monthly report",
						zap.String("user_id", rule.UserID.String()),
						zap.Error(err),
					)
				}
			}
		}
	}
}

// checkScheduledAlertRules checks for scheduled alert rules that need to be evaluated
func (s *schedulerService) checkScheduledAlertRules() {
	s.logger.Debug("Checking scheduled alert rules")
	ctx := context.Background()

	// Get all scheduled rules that are due
	rules, err := s.alertRuleRepo.ListScheduled(ctx)
	if err != nil {
		s.logger.Error("Failed to list scheduled rules", zap.Error(err))
		return
	}

	s.logger.Debug("Found scheduled rules to process", zap.Int("count", len(rules)))

	// Process each rule
	for _, rule := range rules {
		// Skip report-type rules (handled by dedicated report jobs)
		if rule.Type == "scheduled_report" {
			continue
		}

		s.logger.Debug("Processing scheduled rule",
			zap.String("rule_id", rule.ID.String()),
			zap.String("type", string(rule.Type)),
		)

		// TODO: Evaluate other scheduled alert rules
		// This will be implemented in AlertRuleService
	}
}
