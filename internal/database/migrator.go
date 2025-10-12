package database

import (
	"fmt"

	accountdomain "personalfinancedss/internal/module/cashflow/account/domain"
	budgetdomain "personalfinancedss/internal/module/cashflow/budget/domain"
	categorydomain "personalfinancedss/internal/module/cashflow/category/domain"
	debtdomain "personalfinancedss/internal/module/cashflow/debt/domain"
	goaldomain "personalfinancedss/internal/module/cashflow/goal/domain"
	transactiondomain "personalfinancedss/internal/module/cashflow/transaction/domain"
	authdomain "personalfinancedss/internal/module/identify/auth/domain"
	profiledomain "personalfinancedss/internal/module/identify/profile/domain"
	userdomain "personalfinancedss/internal/module/identify/user/domain"
	assetdomain "personalfinancedss/internal/module/investment/investment_asset/domain"
	investmenttransactiondomain "personalfinancedss/internal/module/investment/investment_transaction/domain"
	snapshotdomain "personalfinancedss/internal/module/investment/portfolio_snapshot/domain"
	notificationdomain "personalfinancedss/internal/module/notification/domain"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AutoMigrate runs automatic database migrations for all entities
// Migration order is important to respect foreign key constraints
func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	log.Info("🔧 Running database migrations...")

	// 1. Enable PostgreSQL extensions
	if err := enableUUIDExtension(db, log); err != nil {
		log.Error("Failed to enable PostgreSQL extensions", zap.Error(err))
		return fmt.Errorf("failed to enable PostgreSQL extensions: %w", err)
	}

	// 2. Migrate entities in order (respecting foreign key dependencies)
	// Note: Using VARCHAR for all enum-like fields instead of PostgreSQL ENUMs for flexibility
	entities := []interface{}{
		// 1. Base tables (no foreign keys)
		&userdomain.User{},

		// 2. Tables with foreign key to User
		&profiledomain.UserProfile{},
		&authdomain.VerificationToken{},
		&authdomain.TokenBlacklist{},
	}

	log.Info("Migrating entities", zap.Int("entity_count", len(entities)))

	if err := db.AutoMigrate(entities...); err != nil {
		log.Error("Auto migration failed", zap.Error(err))
		return fmt.Errorf("auto migration failed: %w", err)
	}

	log.Info("✅ Database migrations completed successfully",
		zap.Strings("tables", []string{
			"users",
			"user_profiles",
			"verification_tokens",
			"token_blacklist",
			"periods",
			"accounts",
			"debts",
			"calendar_events",
			"notifications",
			"investment_assets",
			"portfolio_snapshots",
			"categories",
			"security_events",
			"notification_analytics",
			"transactions",
			"investment_transactions",
			"budgets",
			"goals",
		}),
	)

	return nil
}

// enableUUIDExtension enables UUID generation extension for PostgreSQL
func enableUUIDExtension(db *gorm.DB, log *zap.Logger) error {
	log.Info("Enabling required PostgreSQL extensions...")

	// 1. Enable UUID extension
	// Try uuid-ossp first (most common)
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Warn("uuid-ossp extension not available, checking for pgcrypto...", zap.Error(err))

		// Fallback to pgcrypto (alternative UUID generation)
		if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
			log.Warn("pgcrypto extension not available, using built-in gen_random_uuid()", zap.Error(err))
			// PostgreSQL 13+ has built-in gen_random_uuid(), no extension needed
		} else {
			log.Info("pgcrypto extension enabled successfully")
		}
	} else {
		log.Info("uuid-ossp extension enabled successfully")
	}

	// 2. Enable citext extension (for case-insensitive email)
	log.Info("Enabling citext extension for case-insensitive text...")
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "citext"`).Error; err != nil {
		log.Error("Failed to enable citext extension", zap.Error(err))
		return fmt.Errorf("failed to enable citext extension: %w", err)
	}
	log.Info("citext extension enabled successfully")

	return nil
}

// DropAllTables drops all tables (useful for development reset)
// WARNING: This will delete all data!
func DropAllTables(db *gorm.DB, log *zap.Logger) error {
	log.Warn("⚠️  Dropping all tables...")

	// Drop in reverse dependency order (opposite of migration order)
	entities := []interface{}{
		// Budget and Goals tables (drop first - have FKs to User, Category, Account)
		&goaldomain.Goal{},
		&budgetdomain.Budget{},

		// Tables with multiple foreign keys (drop first)
		&investmenttransactiondomain.InvestmentTransaction{},
		&transactiondomain.Transaction{},

		// Tables with FK to notifications
		&notificationdomain.NotificationAnalytics{},

		// Independent or single FK tables
		&notificationdomain.SecurityEvent{},
		&categorydomain.Category{},

		// Tables with FK to User
		&snapshotdomain.PortfolioSnapshot{},
		&assetdomain.InvestmentAsset{},
		&notificationdomain.NotificationPreference{},
		&notificationdomain.AlertRule{},
		&notificationdomain.Notification{},
		&debtdomain.Debt{},
		&accountdomain.Account{},
		&authdomain.TokenBlacklist{},
		&authdomain.VerificationToken{},
		&profiledomain.UserProfile{},

		// Base table (drop last)
		&userdomain.User{},
	}

	log.Info("Dropping tables", zap.Int("entity_count", len(entities)))

	if err := db.Migrator().DropTable(entities...); err != nil {
		log.Error("Failed to drop tables", zap.Error(err))
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	log.Info("✅ All tables dropped successfully")
	return nil
}
