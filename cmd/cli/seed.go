package cmd

import (
	"context"
	"log"
	"os"
	"strings"

	"personalfinancedss/internal/database"
	categorydomain "personalfinancedss/internal/module/cashflow/category/domain"
	categorydto "personalfinancedss/internal/module/cashflow/category/dto"
	userdomain "personalfinancedss/internal/module/identify/user/domain"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed database with complete test data",
	Long:  `Seeds the database with categories, users, and financial data for all 4 user profiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFullSeed()
	},
}

var seedCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Seed only default categories",
	Run: func(cmd *cobra.Command, args []string) {
		runSeedCategories()
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.AddCommand(seedCategoriesCmd)
}

// CLISeeder is a standalone seeder that doesn't need service dependencies
type CLISeeder struct {
	db     *gorm.DB
	logger *zap.Logger
}

func runFullSeed() {
	// Load .env file
	_ = loadEnvFile()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log.Println("🌱 Running complete database seeding...")

	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	seeder := &CLISeeder{db: db, logger: logger}

	// Run in transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Seed categories
		log.Println("Step 1: Seeding system categories...")
		if err := seeder.seedCategories(tx); err != nil {
			return err
		}

		// 2. Seed users
		log.Println("Step 2: Seeding users...")
		if err := seeder.seedUsers(tx); err != nil {
			return err
		}

		// 3. Seed financial data using the existing seeder
		log.Println("Step 3: Seeding financial data for user profiles...")
		mainSeeder := database.NewSeeder(tx, &bcryptHasher{}, &noopUserService{}, getEnvCLI("ADMIN_EMAIL", "admin@example.com"), getEnvCLI("ADMIN_PASSWORD", "Admin@123"), logger)
		if err := mainSeeder.SeedMonthDSSData(tx); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Complete database seeding finished!")
}

func runSeedCategories() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log.Println("🌱 Seeding default categories...")

	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	seeder := &CLISeeder{db: db, logger: logger}
	if err := seeder.seedCategories(db); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Default categories seeded successfully!")
}

func (s *CLISeeder) seedCategories(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&categorydomain.Category{}).Where("is_default = ?", true).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		s.logger.Info("ℹ️  Default categories already exist, skipping", zap.Int64("count", count))
		return nil
	}

	categories := categorydto.FromDefaultCategories(uuid.Nil, true, true)
	if len(categories) == 0 {
		s.logger.Warn("⚠️  No default categories defined")
		return nil
	}

	if err := tx.Create(categories).Error; err != nil {
		return err
	}

	s.logger.Info("✅ Created default categories", zap.Int("count", len(categories)))
	return nil
}

func (s *CLISeeder) seedUsers(tx *gorm.DB) error {
	// Check if users already exist
	var count int64
	if err := tx.Model(&userdomain.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		s.logger.Info("ℹ️  Users already exist, skipping", zap.Int64("count", count))
		return nil
	}

	// Hash password
	adminEmail := getEnvCLI("ADMIN_EMAIL", "admin@example.com")
	adminPassword := getEnvCLI("ADMIN_PASSWORD", "Admin@123")
	defaultPassword := "Password123!"

	hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashedDefaultPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []*userdomain.User{
		{
			Email: adminEmail, Password: string(hashedAdminPassword),
			FullName: "System Admin", Role: userdomain.UserRoleAdmin,
			Status: userdomain.UserStatusActive, EmailVerified: true, AnalyticsConsent: true,
		},
		{
			Email: "john.doe@example.com", Password: string(hashedDefaultPassword),
			FullName: "John Doe", Role: userdomain.UserRoleUser,
			Status: userdomain.UserStatusActive, EmailVerified: true, AnalyticsConsent: true,
		},
		{
			Email: "jane.smith@example.com", Password: string(hashedDefaultPassword),
			FullName: "Jane Smith", Role: userdomain.UserRoleUser,
			Status: userdomain.UserStatusActive, EmailVerified: true, AnalyticsConsent: true,
		},
		{
			Email: "alice.johnson@example.com", Password: string(hashedDefaultPassword),
			FullName: "Alice Johnson", Role: userdomain.UserRoleUser,
			Status: userdomain.UserStatusActive, EmailVerified: true,
		},
		{
			Email: "bob.wilson@example.com", Password: string(hashedDefaultPassword),
			FullName: "Bob Wilson", Role: userdomain.UserRoleUser,
			Status: userdomain.UserStatusPendingVerification, EmailVerified: false,
		},
	}

	for _, user := range users {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		s.logger.Info("✅ Created user", zap.String("email", user.Email))
	}

	return nil
}

// Helpers for CLI that don't need full service dependencies
type bcryptHasher struct{}

func (b *bcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

type noopUserService struct{}

func (n *noopUserService) Create(ctx context.Context, user *userdomain.User) (*userdomain.User, error) {
	return user, nil
}

func getEnvCLI(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// loadEnvFile loads .env file from common locations
func loadEnvFile() error {
	envPaths := []string{
		"deploy/.env",
		".env",
		"../.env",
	}

	for _, path := range envPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				os.Setenv(key, value)
			}
		}
		return nil
	}
	return nil
}
