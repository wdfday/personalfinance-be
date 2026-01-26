package database

import (
	"context"
	"fmt"

	userdomain "personalfinancedss/internal/module/identify/user/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

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

// seedSampleUsers creates sample users for testing (4 users with different profiles)
func (s *Seeder) seedSampleUsers(tx *gorm.DB) error {
	s.logger.Info("Checking for existing sample users...")

	// Check if sample users already exist
	var count int64
	if err := tx.Model(&userdomain.User{}).Where("email IN (?)", []string{
		"john.doe@example.com",
		"jane.smith@example.com",
		"alice.johnson@example.com",
		"bob.wilson@example.com",
	}).Count(&count).Error; err != nil {
		s.logger.Error("Failed to check sample users count", zap.Error(err))
		return err
	}

	if count > 0 {
		s.logger.Info("ℹ️  Sample users already exist, skipping", zap.Int64("count", count))
		return nil
	}

	// Default password for all sample users: "Password123!"
	defaultPassword := "Password123!"
	hashedPassword, err := s.passwordHasher.HashPassword(defaultPassword)
	if err != nil {
		s.logger.Error("Failed to hash sample user password", zap.Error(err))
		return fmt.Errorf("failed to hash sample user password: %w", err)
	}

	// Define sample users with different financial profiles:
	// - John Doe: Salaried employee (stable income, mortgage, big goals)
	// - Jane Smith: Freelancer (variable income, tax vault)
	// - Alice Johnson: Mixed income (salary + side hustle)
	// - Bob Wilson: Student (limited income, small goals)
	sampleUsers := []*userdomain.User{
		{
			Email:            "john.doe@example.com",
			Password:         hashedPassword,
			FullName:         "John Doe",
			Role:             userdomain.UserRoleUser,
			Status:           userdomain.UserStatusActive,
			EmailVerified:    true,
			AnalyticsConsent: true,
		},
		{
			Email:            "jane.smith@example.com",
			Password:         hashedPassword,
			FullName:         "Jane Smith",
			Role:             userdomain.UserRoleUser,
			Status:           userdomain.UserStatusActive,
			EmailVerified:    true,
			AnalyticsConsent: true,
		},
		{
			Email:            "alice.johnson@example.com",
			Password:         hashedPassword,
			FullName:         "Alice Johnson",
			Role:             userdomain.UserRoleUser,
			Status:           userdomain.UserStatusActive,
			EmailVerified:    true,
			AnalyticsConsent: false,
		},
		{
			Email:         "bob.wilson@example.com",
			Password:      hashedPassword,
			FullName:      "Bob Wilson",
			Role:          userdomain.UserRoleUser,
			Status:        userdomain.UserStatusPendingVerification,
			EmailVerified: true,
		},
	}

	ctx := context.Background()
	for _, user := range sampleUsers {
		s.logger.Info("Creating sample user...", zap.String("email", user.Email))
		createdUser, err := s.userService.Create(ctx, user)
		if err != nil {
			s.logger.Error("Failed to create sample user", zap.String("email", user.Email), zap.Error(err))
			return fmt.Errorf("failed to create sample user %s: %w", user.Email, err)
		}
		s.logger.Info("✅ Sample user created", zap.String("email", createdUser.Email), zap.String("id", createdUser.ID.String()))
	}

	s.logger.Info("✅ Seeded sample users successfully", zap.Int("count", len(sampleUsers)))
	return nil
}

// SeedAdminUserOnly seeds only the admin user (useful for testing)
func (s *Seeder) SeedAdminUserOnly() error {
	s.logger.Info("🌱 Seeding admin user only...")
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.seedAdminUser(tx)
	})
}
