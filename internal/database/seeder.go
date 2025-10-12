package database

import (
	"context"
	"fmt"

	categorydomain "personalfinancedss/internal/module/cashflow/category/domain"
	categorydto "personalfinancedss/internal/module/cashflow/category/dto"
	userdomain "personalfinancedss/internal/module/identify/user/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PasswordHasher interface for password hashing
// This allows us to inject the password service without circular dependencies
type PasswordHasher interface {
	HashPassword(password string) (string, error)
}

// UserService interface for user operations
// This allows us to inject the user service without circular dependencies
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

	// Run seeding in transaction for atomicity
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Seed admin user
		s.logger.Info("Step 1: Seeding admin user...")
		if err := s.seedAdminUser(tx); err != nil {
			s.logger.Error("Failed to seed admin user", zap.Error(err))
			return fmt.Errorf("failed to seed admin user: %w", err)
		}

		// 2. Seed default categories
		s.logger.Info("Step 2: Seeding default categories...")
		if err := s.seedDefaultCategories(tx); err != nil {
			s.logger.Error("Failed to seed default categories", zap.Error(err))
			return fmt.Errorf("failed to seed default categories: %w", err)
		}

		s.logger.Info("✅ Database seeding completed successfully")
		return nil
	})
}

// seedAdminUser creates a default admin user with profile using UserService
func (s *Seeder) seedAdminUser(tx *gorm.DB) error {
	s.logger.Info("Checking for existing admin user...")

	// Check if admin user already exists
	var count int64
	if err := tx.Model(&userdomain.User{}).Where("role = ?", userdomain.UserRoleAdmin).Count(&count).Error; err != nil {
		s.logger.Error("Failed to check admin user count", zap.Error(err))
		return err
	}

	if count > 0 {
		s.logger.Info("ℹ️  Admin user already exists, skipping", zap.Int64("count", count))
		return nil
	}

	// Hash password from config
	s.logger.Info("Hashing admin password...")
	hashedPassword, err := s.passwordHasher.HashPassword(s.adminPassword)
	if err != nil {
		s.logger.Error("Failed to hash admin password", zap.Error(err))
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user with credentials from config
	admin := &userdomain.User{
		Email:            s.adminEmail,
		Password:         hashedPassword,
		FullName:         "System Administrator",
		Role:             userdomain.UserRoleAdmin,
		Status:           userdomain.UserStatusActive,
		EmailVerified:    true,
		AnalyticsConsent: true,
	}

	s.logger.Info("Creating admin user with profile...", zap.String("email", s.adminEmail))

	// Create user using UserService (will also create default profile)
	ctx := context.Background()
	createdUser, err := s.userService.Create(ctx, admin)
	if err != nil {
		s.logger.Error("Failed to create admin user via UserService", zap.Error(err))
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	s.logger.Info("✅ Admin user and profile created successfully",
		zap.String("email", createdUser.Email),
		zap.String("id", createdUser.ID.String()),
	)
	return nil
}

// seedDefaultCategories creates 26 default system categories
func (s *Seeder) seedDefaultCategories(tx *gorm.DB) error {
	s.logger.Info("Checking for existing default categories...")

	// Check if default categories already exist
	var count int64
	if err := tx.Model(&categorydomain.Category{}).Where("is_default = ?", true).Count(&count).Error; err != nil {
		s.logger.Error("Failed to check default categories count", zap.Error(err))
		return err
	}

	if count > 0 {
		s.logger.Info("ℹ️  Default categories already exist, skipping", zap.Int64("count", count))
		return nil
	}

	s.logger.Info("Loading default categories from domain...")

	// Get default categories from domain
	// Use uuid.Nil as userID since these are system-wide categories (no specific user)
	categories := categorydto.FromDefaultCategories(uuid.Nil, true, true)

	if len(categories) == 0 {
		s.logger.Warn("⚠️  No default categories defined in domain")
		return nil
	}

	s.logger.Info("Bulk creating categories...", zap.Int("count", len(categories)))

	// Bulk create categories
	if err := tx.Create(categories).Error; err != nil {
		s.logger.Error("Failed to create default categories", zap.Error(err))
		return fmt.Errorf("failed to create default categories: %w", err)
	}

	// Count by type for logging
	var expenseCount, incomeCount int64
	tx.Model(&categorydomain.Category{}).Where("type = ? AND is_default = ?", categorydomain.CategoryTypeExpense, true).Count(&expenseCount)
	tx.Model(&categorydomain.Category{}).Where("type = ? AND is_default = ?", categorydomain.CategoryTypeIncome, true).Count(&incomeCount)

	s.logger.Info("✅ Seeded default categories successfully",
		zap.Int("total", len(categories)),
		zap.Int64("expense", expenseCount),
		zap.Int64("income", incomeCount),
	)
	return nil
}

// SeedAdminUserOnly seeds only the admin user (useful for testing)
func (s *Seeder) SeedAdminUserOnly() error {
	s.logger.Info("🌱 Seeding admin user only...")
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.seedAdminUser(tx)
	})
}

// SeedCategoriesOnly seeds only default categories (useful for testing)
func (s *Seeder) SeedCategoriesOnly() error {
	s.logger.Info("🌱 Seeding categories only...")
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.seedDefaultCategories(tx)
	})
}
