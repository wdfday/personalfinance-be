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
// FREELANCER PROFILE - Jane Smith
// Variable income, needs income smoothing
// Monthly Income: 30-70M VND (biến động lớn)
// =====================================================

func (s *Seeder) seedFreelancerProfile(tx *gorm.DB, userID uuid.UUID) error {
	now := time.Now()
	categoryMap := s.getCategoryMap(tx, userID)

	// 1. Budget Constraints (INPUT for DSS)
	homeCatID, err := s.getCategoryID(categoryMap, "Home & Utilities", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	foodCatID, err := s.getCategoryID(categoryMap, "Food & Dining", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	transportCatID, err := s.getCategoryID(categoryMap, "Transportation", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	shoppingCatID, err := s.getCategoryID(categoryMap, "Shopping", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	healthCatID, err := s.getCategoryID(categoryMap, "Health & Wellness", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	entertainmentCatID, err := s.getCategoryID(categoryMap, "Entertainment & Travel", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	financialCatID, err := s.getCategoryID(categoryMap, "Financial & Obligations", "freelancer constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}

	constraints := []*constraintdomain.BudgetConstraint{
		// Tiền nhà - FIXED
		{
			UserID: userID, CategoryID: homeCatID,
			MinimumAmount: 6000000, MaximumAmount: 6000000, IsFlexible: false,
			Priority: 1, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Thuê studio + điện nước",
		},
		// Ăn uống - FLEXIBLE (tự nấu nhiều khi ít việc)
		{
			UserID: userID, CategoryID: foodCatID,
			MinimumAmount: 6000000, MaximumAmount: 14000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Ăn uống linh hoạt theo thu nhập",
		},
		// Đi lại - FLEXIBLE
		{
			UserID: userID, CategoryID: transportCatID,
			MinimumAmount: 2000000, MaximumAmount: 4000000, IsFlexible: true,
			Priority: 3, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Grab, taxi khi gặp khách hàng",
		},
		// Công việc - FLEXIBLE (thiết bị, software, marketing)
		{
			UserID: userID, CategoryID: shoppingCatID,
			MinimumAmount: 2000000, MaximumAmount: 8000000, IsFlexible: true,
			Priority: 3, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Software, thiết bị làm việc, portfolio hosting",
		},
		// Y tế - FLEXIBLE
		{
			UserID: userID, CategoryID: healthCatID,
			MinimumAmount: 1000000, MaximumAmount: 3000000, IsFlexible: true,
			Priority: 4, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Bảo hiểm y tế tự nguyện, khám định kỳ",
		},
		// Giải trí & Networking - FLEXIBLE (quan trọng cho freelancer)
		{
			UserID: userID, CategoryID: entertainmentCatID,
			MinimumAmount: 2000000, MaximumAmount: 6000000, IsFlexible: true,
			Priority: 5, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Cafe networking, co-working space, giải trí",
		},
		// BHXH/Thuế - Reserve
		{
			UserID: userID, CategoryID: financialCatID,
			MinimumAmount: 3000000, MaximumAmount: 8000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "BHXH tự nguyện + dự trữ thuế TNCN",
		},
	}

	for _, c := range constraints {
		if err := tx.Create(c).Error; err != nil {
			return fmt.Errorf("failed to create constraint: %w", err)
		}
		s.logger.Info("✅ Created constraint", zap.Float64("min", c.MinimumAmount))
	}

	// 2. Income Profiles - Variable income, multiple sources
	activeIncomeCatID, err := s.getCategoryID(categoryMap, "Active Income", "freelancer income")
	if err != nil {
		return fmt.Errorf("failed to get category for income: %w", err)
	}
	otherIncomeCatID, err := s.getCategoryID(categoryMap, "Other Income", "freelancer income")
	if err != nil {
		return fmt.Errorf("failed to get category for income: %w", err)
	}

	incomes := []*incomedomain.IncomeProfile{
		{
			UserID: userID, CategoryID: activeIncomeCatID,
			Source: "Freelance Web Development",
			Amount: 45000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-3, 0, 0),
			IsRecurring: false,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Dự án web - biến động 25-70M/tháng",
		},
		{
			UserID: userID, CategoryID: activeIncomeCatID,
			Source: "Freelance UI/UX Design",
			Amount: 18000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-2, 0, 0),
			IsRecurring: false,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Thiết kế giao diện, branding",
		},
		{
			UserID: userID, CategoryID: otherIncomeCatID,
			Source: "Passive - Template Sales",
			Amount: 5000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-1, 0, 0),
			IsRecurring: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Bán template trên ThemeForest",
		},
	}

	for _, income := range incomes {
		if err := tx.Create(income).Error; err != nil {
			return fmt.Errorf("failed to create income: %w", err)
		}
		s.logger.Info("✅ Created income", zap.String("source", income.Source))
	}

	// 3. Goals - Focus on stability
	accountID, err := s.getAccountID(tx, userID)
	if err != nil {
		return fmt.Errorf("failed to get account for goals: %w", err)
	}
	goals := []*goaldomain.Goal{
		{
			UserID: userID, AccountID: accountID,
			Name: "Reservoir Fund", Description: strPtr("6 tháng chi tiêu cho income smoothing"),
			TargetAmount: 200000000, CurrentAmount: 75000000,
			StartDate: now.AddDate(-1, 0, 0), TargetDate: timePtr(now.AddDate(0, 8, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryEmergency,
			Priority: goaldomain.GoalPriorityCritical, Status: goaldomain.GoalStatusActive,
			Notes: strPtr("Dùng salary-to-self khi mùa thấp điểm"),
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Tax Vault - TNCN", Description: strPtr("Dự trữ 10-15% cho thuế"),
			TargetAmount: 60000000, CurrentAmount: 25000000,
			StartDate: now.AddDate(-6, 0, 0), TargetDate: timePtr(now.AddDate(0, 12, 0)),
			Behavior: goaldomain.GoalBehaviorRecurring, Category: goaldomain.GoalCategoryOther,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Upgrade Workstation", Description: strPtr("MacBook Pro M3 Max + Studio Display"),
			TargetAmount: 120000000, CurrentAmount: 45000000,
			StartDate: now.AddDate(0, -4, 0), TargetDate: timePtr(now.AddDate(0, 6, 0)),
			Behavior: goaldomain.GoalBehaviorWilling, Category: goaldomain.GoalCategoryPurchase,
			Priority: goaldomain.GoalPriorityMedium, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Quỹ khởi nghiệp", Description: strPtr("Vốn mở design agency"),
			TargetAmount: 300000000, CurrentAmount: 50000000,
			StartDate: now.AddDate(-6, 0, 0), TargetDate: timePtr(now.AddDate(2, 0, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryInvestment,
			Priority: goaldomain.GoalPriorityLow, Status: goaldomain.GoalStatusActive,
		},
	}

	for _, goal := range goals {
		goal.UpdateCalculatedFields()
		if err := tx.Create(goal).Error; err != nil {
			return fmt.Errorf("failed to create goal: %w", err)
		}
		s.logger.Info("✅ Created goal", zap.String("name", goal.Name))
	}

	// 4. Debts - Minimal, mostly tax
	debts := []*debtdomain.Debt{
		{
			UserID: userID, Name: "Thuế TNCN 2024",
			Description: strPtr("Thuế chưa quyết toán"),
			Type:        debtdomain.DebtTypeOther, Behavior: debtdomain.DebtBehaviorInstallment,
			PrincipalAmount: 25000000, CurrentBalance: 18000000,
			Currency: "VND", InterestRate: 0, MinimumPayment: 6000000,
			StartDate: now.AddDate(0, -2, 0), NextPaymentDate: timePtr(now.AddDate(0, 1, 0)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("Cục Thuế"),
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
