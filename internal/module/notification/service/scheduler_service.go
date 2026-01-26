package service

import (
	"context"
	userrepo "personalfinancedss/internal/module/identify/user/repository"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type schedulerService struct {
	cron          *cron.Cron
	reportService ScheduledReportService
	userRepo      userrepo.Repository
	logger        *zap.Logger
	isRunning     bool
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(
	reportService ScheduledReportService,
	userRepo userrepo.Repository,
	logger *zap.Logger,
) SchedulerService {
	// Create cron instance with second precision
	c := cron.New(cron.WithSeconds())

	return &schedulerService{
		cron:          c,
		reportService: reportService,
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

// runDailyReports generates daily reports for all users
// TODO: Implement user-specific daily report scheduling
func (s *schedulerService) runDailyReports() {
	s.logger.Info("Running daily reports job")
	ctx := context.Background()
	// TODO: Get list of users who want daily reports and generate reports
	_ = ctx
}

// runWeeklyReports generates weekly reports for all users
// TODO: Implement user-specific weekly report scheduling
func (s *schedulerService) runWeeklyReports() {
	s.logger.Info("Running weekly reports job")
	ctx := context.Background()
	// TODO: Get list of users who want weekly reports and generate reports
	_ = ctx
}

// runMonthlyReports generates monthly reports for all users
// TODO: Implement user-specific monthly report scheduling
func (s *schedulerService) runMonthlyReports() {
	s.logger.Info("Running monthly reports job")
	ctx := context.Background()
	// TODO: Get list of users who want monthly reports and generate reports
	_ = ctx
}
