package database

import (
	"context"
	"fmt"

	userdomain "personalfinancedss/internal/module/identify/user/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PasswordHasher interface for password hashing
type PasswordHasher interface {
	HashPassword(password string) (string, error)
}

// UserService interface for user operations
type UserService interface {
	Create(ctx context.Context, user *userdomain.User) (*userdomain.User, error)
}

// Seeder handles database seeding operations
type Seeder struct {
	db             *gorm.DB
	passwordHasher PasswordHasher
	userService    UserService
	adminEmail     string
	adminPassword  string
	logger         *zap.Logger
}

// NewSeeder creates a new database seeder
func NewSeeder(db *gorm.DB, passwordHasher PasswordHasher, userService UserService, adminEmail, adminPassword string, logger *zap.Logger) *Seeder {
	return &Seeder{
		db:             db,
		passwordHasher: passwordHasher,
		userService:    userService,
		adminEmail:     adminEmail,
		adminPassword:  adminPassword,
		logger:         logger,
	}
}

// SeedAll runs all seeding operations
func (s *Seeder) SeedAll() error {
	s.logger.Info("🌱 Running database seeder...")

	// PHASE 1: Seed system categories FIRST (separate transaction, must commit before users)
	// This is needed because UserService.Create() uses its own DB connection
	// and won't see uncommitted categories from a parent transaction
	s.logger.Info("Phase 1: Seeding default system categories...")
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.seedDefaultCategories(tx)
	}); err != nil {
		s.logger.Error("Failed to seed default categories", zap.Error(err))
		return fmt.Errorf("failed to seed default categories: %w", err)
	}
	s.logger.Info("✅ System categories seeded and committed")

	// PHASE 2: Seed users and financial data (categories are now visible)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 2. Seed admin user (will auto-clone categories from system)
		s.logger.Info("Phase 2: Seeding admin user...")
		if err := s.seedAdminUser(tx); err != nil {
			s.logger.Error("Failed to seed admin user", zap.Error(err))
			return fmt.Errorf("failed to seed admin user: %w", err)
		}

		// 3. Seed sample users (4 different financial profiles)
		s.logger.Info("Phase 2: Seeding sample users...")
		if err := s.seedSampleUsers(tx); err != nil {
			s.logger.Error("Failed to seed sample users", zap.Error(err))
			return fmt.Errorf("failed to seed sample users: %w", err)
		}

		// 4. Seed comprehensive financial data for all user profiles
		s.logger.Info("Phase 2: Seeding financial data (budgets, income, goals, debts)...")
		if err := s.SeedMonthDSSData(tx); err != nil {
			s.logger.Error("Failed to seed financial data", zap.Error(err))
			return fmt.Errorf("failed to seed financial data: %w", err)
		}

		s.logger.Info("✅ Database seeding completed successfully")
		return nil
	})
}
