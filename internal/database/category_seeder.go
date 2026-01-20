package database

import (
	"fmt"

	categorydomain "personalfinancedss/internal/module/cashflow/category/domain"
	categorydto "personalfinancedss/internal/module/cashflow/category/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// seedDefaultCategories creates default system categories
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

// SeedCategoriesOnly seeds only default categories (useful for testing)
func (s *Seeder) SeedCategoriesOnly() error {
	s.logger.Info("🌱 Seeding categories only...")
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.seedDefaultCategories(tx)
	})
}
