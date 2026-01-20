package database

import (
	"fmt"
	"time"

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
func (s *Seeder) getCategoryMap(tx *gorm.DB, userID uuid.UUID) map[string]uuid.UUID {
	categoryMap := make(map[string]uuid.UUID)

	var categories []categorydomain.Category
	// Get system default categories (user_id is uuid.Nil for system defaults)
	tx.Where("is_default = ?", true).Find(&categories)

	for _, cat := range categories {
		categoryMap[cat.Name] = cat.ID
	}

	return categoryMap
}

// getAccountID returns the first account ID for a user (or creates dummy)
func (s *Seeder) getAccountID(tx *gorm.DB, userID uuid.UUID) uuid.UUID {
	var accountIDStr string
	if err := tx.Table("accounts").Select("id").Where("user_id = ?", userID).Limit(1).Scan(&accountIDStr).Error; err != nil || accountIDStr == "" {
		return uuid.New() // Fallback
	}

	accountID, _ := uuid.Parse(accountIDStr)
	return accountID
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
