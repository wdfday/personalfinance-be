package database

import (
	"fmt"
	"time"

	constraintdomain "personalfinancedss/internal/module/cashflow/budget_profile/domain"
	debtdomain "personalfinancedss/internal/module/cashflow/debt/domain"
	goaldomain "personalfinancedss/internal/module/cashflow/goal/domain"
	incomedomain "personalfinancedss/internal/module/cashflow/income_profile/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// =====================================================
// SALARIED PROFILE - John Doe
// Stable employee with mortgage, family goals
// Monthly Income: ~30M VND (stable)
// =====================================================

func (s *Seeder) seedSalariedProfile(tx *gorm.DB, userID uuid.UUID) error {
	now := time.Now()
	categoryMap := s.getCategoryMap(tx, userID)

	// 1. Budget Constraints (INPUT for DSS)
	homeCatID, err := s.getCategoryID(categoryMap, "Home & Utilities", "salaried constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	foodCatID, err := s.getCategoryID(categoryMap, "Food & Dining", "salaried constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	transportCatID, err := s.getCategoryID(categoryMap, "Transportation", "salaried constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	healthCatID, err := s.getCategoryID(categoryMap, "Health & Wellness", "salaried constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	entertainmentCatID, err := s.getCategoryID(categoryMap, "Entertainment & Travel", "salaried constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}

	constraints := []*constraintdomain.BudgetConstraint{
		// Tiền nhà - FIXED (không thể giảm)
		{
			UserID: userID, CategoryID: homeCatID,
			MinimumAmount: 7000000, MaximumAmount: 7000000, IsFlexible: false,
			Priority: 1, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Tiền thuê nhà cố định",
		},
		// Ăn uống - FLEXIBLE
		{
			UserID: userID, CategoryID: foodCatID,
			MinimumAmount: 8000000, MaximumAmount: 15000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Chi phí ăn uống gia đình",
		},
		// Đi lại - FLEXIBLE
		{
			UserID: userID, CategoryID: transportCatID,
			MinimumAmount: 1500000, MaximumAmount: 2500000, IsFlexible: true,
			Priority: 3, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Xăng xe, Grab",
		},
		// Y tế - FLEXIBLE
		{
			UserID: userID, CategoryID: healthCatID,
			MinimumAmount: 500000, MaximumAmount: 1500000, IsFlexible: true,
			Priority: 4, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Bảo hiểm, khám bệnh",
		},
		// Giải trí - FLEXIBLE (có thể cắt giảm)
		{
			UserID: userID, CategoryID: entertainmentCatID,
			MinimumAmount: 2000000, MaximumAmount: 4000000, IsFlexible: true,
			Priority: 5, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Xem phim, cafe, giải trí",
		},
	}

	for _, c := range constraints {
		if err := tx.Create(c).Error; err != nil {
			return fmt.Errorf("failed to create constraint: %w", err)
		}
		s.logger.Info("✅ Created constraint", zap.Float64("min", c.MinimumAmount))
	}

	// 2. Income Profiles
	activeIncomeCatID, err := s.getCategoryID(categoryMap, "Active Income", "salaried income")
	if err != nil {
		return fmt.Errorf("failed to get category for income: %w", err)
	}

	incomes := []*incomedomain.IncomeProfile{
		{
			UserID: userID, CategoryID: activeIncomeCatID,
			Source: "Lương công ty ABC Corp",
			Amount: 60000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-2, 0, 0),
			IsRecurring: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Lương chính + phụ cấp + KPI",
		},
	}

	for _, income := range incomes {
		if err := tx.Create(income).Error; err != nil {
			return fmt.Errorf("failed to create income: %w", err)
		}
		s.logger.Info("✅ Created income", zap.String("source", income.Source))
	}

	// 3. Goals
	accountID, err := s.getAccountID(tx, userID)
	if err != nil {
		return fmt.Errorf("failed to get account for goals: %w", err)
	}
	goals := []*goaldomain.Goal{
		{
			UserID: userID, AccountID: accountID,
			Name: "Quỹ khẩn cấp 6 tháng", Description: strPtr("180M cho 6 tháng chi tiêu"),
			TargetAmount: 180000000, CurrentAmount: 85000000,
			StartDate: now.AddDate(-1, 0, 0), TargetDate: timePtr(now.AddDate(0, 10, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryEmergency,
			Priority: goaldomain.GoalPriorityCritical, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Tiền cọc mua nhà", Description: strPtr("30% căn hộ 3 tỷ"),
			TargetAmount: 900000000, CurrentAmount: 180000000,
			StartDate: now.AddDate(-1, 6, 0), TargetDate: timePtr(now.AddDate(4, 0, 0)),
			Behavior: goaldomain.GoalBehaviorWilling, Category: goaldomain.GoalCategoryPurchase,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Mua ô tô Honda City", Description: strPtr("Xe gia đình 600M"),
			TargetAmount: 600000000, CurrentAmount: 150000000,
			StartDate: now.AddDate(-6, 0, 0), TargetDate: timePtr(now.AddDate(2, 6, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryPurchase,
			Priority: goaldomain.GoalPriorityMedium, Status: goaldomain.GoalStatusActive,
		},
	}

	for _, goal := range goals {
		goal.UpdateCalculatedFields()
		if err := tx.Create(goal).Error; err != nil {
			return fmt.Errorf("failed to create goal: %w", err)
		}
		s.logger.Info("✅ Created goal", zap.String("name", goal.Name))
	}

	// 4. Debts
	debts := []*debtdomain.Debt{
		{
			UserID: userID, Name: "Vay mua nhà VinHomes",
			Description: strPtr("Vay Techcombank 20 năm"),
			Type:        debtdomain.DebtTypeMortgage, Behavior: debtdomain.DebtBehaviorInstallment,
			PrincipalAmount: 1800000000, CurrentBalance: 1650000000,
			Currency: "VND", InterestRate: 8.5, MinimumPayment: 18000000,
			StartDate: now.AddDate(-1, 6, 0), NextPaymentDate: timePtr(now.AddDate(0, 0, 25)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("Techcombank"),
		},
		{
			UserID: userID, Name: "Trả góp xe SH",
			Description: strPtr("Honda Finance 12 tháng 0%"),
			Type:        debtdomain.DebtTypePersonalLoan, Behavior: debtdomain.DebtBehaviorInstallment,
			PrincipalAmount: 60000000, CurrentBalance: 20000000,
			Currency: "VND", InterestRate: 0, MinimumPayment: 5000000,
			StartDate: now.AddDate(0, -8, 0), NextPaymentDate: timePtr(now.AddDate(0, 0, 5)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("Honda Finance"),
		},
	}

	for _, debt := range debts {
		debt.UpdateCalculatedFields()
		if err := tx.Create(debt).Error; err != nil {
			return fmt.Errorf("failed to create debt: %w", err)
		}
		s.logger.Info("✅ Created debt", zap.String("name", debt.Name))
	}

	return nil
}
