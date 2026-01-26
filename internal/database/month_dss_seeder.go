package database

import (
	"fmt"
	"time"

	accountdomain "personalfinancedss/internal/module/cashflow/account/domain"
	categorydomain "personalfinancedss/internal/module/cashflow/category/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// =====================================================
// MONTH DSS SEEDER - Main Orchestrator
// Seeds complete financial input data for 4 user profiles
// =====================================================

// UserProfile defines different financial profiles for seeding
type UserProfile string

const (
	ProfileSalaried   UserProfile = "salaried"   // John - stable salaried employee
	ProfileFreelancer UserProfile = "freelancer" // Jane - pure freelancer
	ProfileMixed      UserProfile = "mixed"      // Alice - salary + side hustle
	ProfileStudent    UserProfile = "student"    // Bob - limited income
)

// SeedMonthDSSData seeds complete INPUT data for Month DSS testing
func (s *Seeder) SeedMonthDSSData(tx *gorm.DB) error {
	s.logger.Info("🌱 Seeding Month DSS input data for sample users...")

	// First, ensure all users have accounts
	s.logger.Info("Ensuring all users have accounts...")
	if err := s.seedAccountsForUsers(tx); err != nil {
		s.logger.Error("Failed to seed accounts", zap.Error(err))
		return fmt.Errorf("failed to seed accounts: %w", err)
	}

	// User mappings
	userProfiles := map[string]UserProfile{
		"john.doe@example.com":      ProfileSalaried,
		"jane.smith@example.com":    ProfileFreelancer,
		"alice.johnson@example.com": ProfileMixed,
		"bob.wilson@example.com":    ProfileStudent,
	}

	for email, profile := range userProfiles {
		// Get user ID
		var userIDStr string
		if err := tx.Table("users").Select("id").Where("email = ?", email).Scan(&userIDStr).Error; err != nil {
			s.logger.Warn("User not found, skipping", zap.String("email", email), zap.Error(err))
			continue
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil || userID == uuid.Nil {
			s.logger.Warn("Invalid user ID, skipping", zap.String("email", email))
			continue
		}

		s.logger.Info("Seeding data for user",
			zap.String("email", email),
			zap.String("profile", string(profile)),
		)

		// Check if already seeded (check budget_constraints instead of budgets)
		var constraintCount int64
		tx.Table("budget_constraints").Where("user_id = ?", userID).Count(&constraintCount)
		if constraintCount > 0 {
			s.logger.Info("ℹ️  Data already exists, skipping", zap.String("email", email))
			continue
		}

		if err := s.seedProfileByType(tx, userID, profile); err != nil {
			s.logger.Error("Failed to seed user", zap.String("email", email), zap.Error(err))
			continue
		}

		s.logger.Info("✅ Seeded data for", zap.String("profile", string(profile)))
	}

	s.logger.Info("✅ Month DSS input data seeded successfully")
	return nil
}

func (s *Seeder) seedProfileByType(tx *gorm.DB, userID uuid.UUID, profile UserProfile) error {
	switch profile {
	case ProfileSalaried:
		return s.seedSalariedProfile(tx, userID)
	case ProfileFreelancer:
		return s.seedFreelancerProfile(tx, userID)
	case ProfileMixed:
		return s.seedMixedProfile(tx, userID)
	case ProfileStudent:
		return s.seedStudentProfile(tx, userID)
	default:
		return fmt.Errorf("unknown profile: %s", profile)
	}
}

// =====================================================
// HELPER FUNCTIONS
// =====================================================

// getCategoryMap returns a map of category name -> ID for a user
// If user doesn't have categories yet, it will automatically clone system categories for them
func (s *Seeder) getCategoryMap(tx *gorm.DB, userID uuid.UUID) map[string]uuid.UUID {
	categoryMap := make(map[string]uuid.UUID)

	// First, check if user already has categories
	var categories []categorydomain.Category
	tx.Where("user_id = ?", userID).Find(&categories)

	// If user has no categories, clone from system categories
	if len(categories) == 0 {
		s.logger.Info("User has no categories, cloning from system defaults", zap.String("user_id", userID.String()))

		// Get system categories
		var systemCategories []categorydomain.Category
		tx.Where("is_default = ? AND user_id IS NULL", true).Find(&systemCategories)

		if len(systemCategories) == 0 {
			s.logger.Warn("No system categories found to clone", zap.String("user_id", userID.String()))
			return categoryMap
		}

		// Clone categories for user
		userCategories := s.cloneCategoriesForUserInTx(tx, systemCategories, userID)

		// Bulk create user categories
		if err := tx.Create(userCategories).Error; err != nil {
			s.logger.Error("Failed to clone categories for user",
				zap.String("user_id", userID.String()),
				zap.Error(err))
			return categoryMap
		}

		s.logger.Info("Successfully cloned categories for user",
			zap.String("user_id", userID.String()),
			zap.Int("count", len(userCategories)))

		// Re-query to get the created categories
		tx.Where("user_id = ?", userID).Find(&categories)
	}

	// Build map from categories
	for _, cat := range categories {
		categoryMap[cat.Name] = cat.ID
	}

	return categoryMap
}

// getCategoryID safely gets category ID from map and returns error if not found
func (s *Seeder) getCategoryID(categoryMap map[string]uuid.UUID, categoryName string, context string) (uuid.UUID, error) {
	categoryID, exists := categoryMap[categoryName]
	if !exists || categoryID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("category '%s' not found in categoryMap (context: %s)", categoryName, context)
	}
	return categoryID, nil
}

// cloneCategoriesForUserInTx creates copies of system categories for a specific user
// It maintains the hierarchical structure by remapping parent IDs
func (s *Seeder) cloneCategoriesForUserInTx(tx *gorm.DB, systemCategories []categorydomain.Category, userID uuid.UUID) []*categorydomain.Category {
	idMap := make(map[uuid.UUID]uuid.UUID, len(systemCategories))
	userCategories := make([]*categorydomain.Category, 0, len(systemCategories))

	// First pass: clone categories and create ID mapping
	for _, systemCat := range systemCategories {
		newCat := &categorydomain.Category{
			ID:           uuid.New(),
			UserID:       &userID,
			Name:         systemCat.Name,
			Description:  systemCat.Description,
			Type:         systemCat.Type,
			Icon:         systemCat.Icon,
			Color:        systemCat.Color,
			IsDefault:    false, // User copy is not a system default
			IsActive:     true,
			Level:        systemCat.Level,
			DisplayOrder: systemCat.DisplayOrder,
			ParentID:     systemCat.ParentID, // Will be remapped in second pass
		}
		idMap[systemCat.ID] = newCat.ID
		userCategories = append(userCategories, newCat)
	}

	// Second pass: remap parent IDs to new user-specific IDs
	for _, userCat := range userCategories {
		if userCat.ParentID != nil {
			if newParentID, exists := idMap[*userCat.ParentID]; exists {
				userCat.ParentID = &newParentID
			} else {
				// Orphaned category - reset to root level
				s.logger.Warn("Parent category not found, resetting to root",
					zap.String("category_id", userCat.ID.String()))
				userCat.ParentID = nil
				userCat.Level = 0
			}
		}
	}

	return userCategories
}

// seedAccountsForUsers ensures all sample users have a debit account
// Note: Cash account is already created by UserService.CreateDefaultCashAccount
func (s *Seeder) seedAccountsForUsers(tx *gorm.DB) error {
	// User email -> monthly income (for balance calculation)
	userIncomes := map[string]float64{
		"john.doe@example.com":      50000000, // Salaried: ~50M/month
		"jane.smith@example.com":    35000000, // Freelancer: ~35M/month
		"alice.johnson@example.com": 29000000, // Mixed: 18M + 8M + 3M = 29M/month
		"bob.wilson@example.com":    8000000,  // Student: ~8M/month
	}

	for email, monthlyIncome := range userIncomes {
		// Get user ID
		var userIDStr string
		if err := tx.Table("users").Select("id").Where("email = ?", email).Scan(&userIDStr).Error; err != nil {
			s.logger.Warn("User not found, skipping account creation", zap.String("email", email), zap.Error(err))
			continue
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil || userID == uuid.Nil {
			s.logger.Warn("Invalid user ID, skipping account creation", zap.String("email", email))
			continue
		}

		// Check if user already has a debit account
		var debitAccountCount int64
		if err := tx.Table("accounts").
			Where("user_id = ? AND account_type = ?", userID, accountdomain.AccountTypeBank).
			Count(&debitAccountCount).Error; err != nil {
			s.logger.Warn("Failed to check debit account count", zap.String("email", email), zap.Error(err))
			continue
		}

		if debitAccountCount > 0 {
			s.logger.Info("User already has debit account(s), skipping", zap.String("email", email), zap.Int64("count", debitAccountCount))
			continue
		}

		// Calculate initial balance: 3-4 times monthly income
		// This represents savings accumulated over time
		initialBalance := monthlyIncome * 3.5

		// Create debit account (bank account) with initial balance
		debitAccount := &accountdomain.Account{
			ID:                  uuid.New(),
			UserID:              userID,
			AccountName:         "Techcombank",
			AccountType:         accountdomain.AccountTypeBank,
			CurrentBalance:      initialBalance,
			Currency:            accountdomain.CurrencyVND,
			IsActive:            true,
			IsPrimary:           true, // Debit account is primary
			IncludeInNetWorth:   true,
			InstitutionName:     strPtr("Techcombank"),
			AccountNumberMasked: strPtr("****1234"),
		}

		if err := tx.Create(debitAccount).Error; err != nil {
			s.logger.Error("Failed to create debit account", zap.String("email", email), zap.Error(err))
			return fmt.Errorf("failed to create debit account for user %s: %w", email, err)
		}

		s.logger.Info("✅ Created debit account for user",
			zap.String("email", email),
			zap.String("account_id", debitAccount.ID.String()),
			zap.Float64("initial_balance", initialBalance),
			zap.Float64("monthly_income", monthlyIncome))
	}

	return nil
}

// getAccountID returns the first account ID for a user
// Returns error if no account exists (fail fast)
func (s *Seeder) getAccountID(tx *gorm.DB, userID uuid.UUID) (uuid.UUID, error) {
	var accountIDStr string
	if err := tx.Table("accounts").Select("id").Where("user_id = ?", userID).Limit(1).Scan(&accountIDStr).Error; err != nil || accountIDStr == "" {
		return uuid.Nil, fmt.Errorf("no account found for user %s: %w", userID.String(), err)
	}

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse account ID '%s' for user %s: %w", accountIDStr, userID.String(), err)
	}
	return accountID, nil
}

// getDebitAccountID returns the primary debit account (bank account) ID for a user
// Returns error if no debit account exists (fail fast)
func (s *Seeder) getDebitAccountID(tx *gorm.DB, userID uuid.UUID) (uuid.UUID, error) {
	var accountIDStr string
	if err := tx.Table("accounts").
		Select("id").
		Where("user_id = ? AND account_type = ? AND is_primary = ?", userID, accountdomain.AccountTypeBank, true).
		Limit(1).
		Scan(&accountIDStr).Error; err != nil || accountIDStr == "" {
		// Fallback to any bank account
		if err := tx.Table("accounts").
			Select("id").
			Where("user_id = ? AND account_type = ?", userID, accountdomain.AccountTypeBank).
			Limit(1).
			Scan(&accountIDStr).Error; err != nil || accountIDStr == "" {
			return uuid.Nil, fmt.Errorf("no debit account found for user %s: %w", userID.String(), err)
		}
	}

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse debit account ID '%s' for user %s: %w", accountIDStr, userID.String(), err)
	}
	return accountID, nil
}

// Helper functions for pointers
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func floatPtr(f float64) *float64 {
	return &f
}
