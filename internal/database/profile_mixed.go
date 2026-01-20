package database

import (
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
// MIXED PROFILE - Alice Johnson
// Part-time salary + Side hustle
// Monthly Income: ~28M VND (18M fixed + 10M variable)
// =====================================================

func (s *Seeder) seedMixedProfile(tx *gorm.DB, userID uuid.UUID) error {
	now := time.Now()
	categoryMap := s.getCategoryMap(tx, userID)

	// 1. Budget Constraints (INPUT for DSS)
	constraints := []*constraintdomain.BudgetConstraint{
		// Tiền nhà - FIXED
		{
			UserID: userID, CategoryID: categoryMap["Home & Utilities"],
			MinimumAmount: 4500000, MaximumAmount: 4500000, IsFlexible: false,
			Priority: 1, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Thuê phòng studio",
		},
		// Ăn uống - FLEXIBLE
		{
			UserID: userID, CategoryID: categoryMap["Food & Dining"],
			MinimumAmount: 3500000, MaximumAmount: 5000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Ăn uống cá nhân",
		},
		// Đi lại - FLEXIBLE
		{
			UserID: userID, CategoryID: categoryMap["Transportation"],
			MinimumAmount: 1000000, MaximumAmount: 2000000, IsFlexible: true,
			Priority: 3, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly",
		},
		// Học tập - FLEXIBLE (MBA)
		{
			UserID: userID, CategoryID: categoryMap["Education & Self-Dev"],
			MinimumAmount: 5000000, MaximumAmount: 8000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Trả góp MBA + sách vở",
		},
	}

	for _, c := range constraints {
		if c.CategoryID != uuid.Nil {
			if err := tx.Create(c).Error; err != nil {
				return err
			}
			s.logger.Info("✅ Created constraint", zap.Float64("min", c.MinimumAmount))
		}
	}

	// 2. Income Profiles - Mixed stable + variable
	incomes := []*incomedomain.IncomeProfile{
		{
			UserID: userID, CategoryID: categoryMap["Active Income"],
			Source: "Lương chính - Marketing",
			Amount: 18000000, Currency: "VND", Frequency: "monthly",
			StartDate:  now.AddDate(-1, 6, 0),
			BaseSalary: 15000000, Allowance: 3000000,
			IsRecurring: true, IsVerified: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Marketing Executive tại agency",
		},
		{
			UserID: userID, CategoryID: categoryMap["Active Income"],
			Source: "Side hustle - Content Writing",
			Amount: 8000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-1, 0, 0),
			Commission:  8000000,
			IsRecurring: false, IsVerified: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Viết content, 5-10 bài/tháng",
		},
		{
			UserID: userID, CategoryID: categoryMap["Other Income"],
			Source: "Affiliate Marketing",
			Amount: 3000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-6, 0, 0),
			Commission:  3000000,
			IsRecurring: false, IsVerified: false,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Commission giới thiệu sản phẩm",
		},
	}

	for _, income := range incomes {
		if err := tx.Create(income).Error; err != nil {
			return err
		}
		s.logger.Info("✅ Created income", zap.String("source", income.Source))
	}

	// 3. Goals
	accountID := s.getAccountID(tx, userID)
	goals := []*goaldomain.Goal{
		{
			UserID: userID, AccountID: accountID,
			Name: "Quỹ dự phòng 3 tháng", Description: strPtr("75M cho 3 tháng chi tiêu"),
			TargetAmount: 75000000, CurrentAmount: 35000000,
			StartDate: now.AddDate(-6, 0, 0), TargetDate: timePtr(now.AddDate(0, 6, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryEmergency,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Du lịch Hàn Quốc", Description: strPtr("Trip 7 ngày Q4"),
			TargetAmount: 40000000, CurrentAmount: 18000000,
			StartDate: now.AddDate(0, -2, 0), TargetDate: timePtr(now.AddDate(0, 8, 0)),
			Behavior: goaldomain.GoalBehaviorWilling, Category: goaldomain.GoalCategoryTravel,
			Priority: goaldomain.GoalPriorityMedium, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Học MBA Part-time", Description: strPtr("Học phí MBA 2 năm tại RMIT"),
			TargetAmount: 500000000, CurrentAmount: 80000000,
			StartDate: now.AddDate(-1, 0, 0), TargetDate: timePtr(now.AddDate(2, 6, 0)),
			Behavior: goaldomain.GoalBehaviorRecurring, Category: goaldomain.GoalCategoryEducation,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
	}

	for _, goal := range goals {
		goal.UpdateCalculatedFields()
		if err := tx.Create(goal).Error; err != nil {
			return err
		}
		s.logger.Info("✅ Created goal", zap.String("name", goal.Name))
	}

	// 4. Debts
	debts := []*debtdomain.Debt{
		{
			UserID: userID, Name: "Credit Card VIB",
			Description: strPtr("Thẻ tín dụng chi tiêu"),
			Type:        debtdomain.DebtTypeCreditCard, Behavior: debtdomain.DebtBehaviorRevolving,
			PrincipalAmount: 40000000, CurrentBalance: 18000000,
			Currency: "VND", InterestRate: 22, MinimumPayment: 2000000,
			StartDate: now.AddDate(-1, 0, 0), NextPaymentDate: timePtr(now.AddDate(0, 0, 15)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("VIB Bank"),
		},
		{
			UserID: userID, Name: "Vay học MBA",
			Description: strPtr("Student loan VPBank"),
			Type:        debtdomain.DebtTypePersonalLoan, Behavior: debtdomain.DebtBehaviorInterestOnly,
			PrincipalAmount: 200000000, CurrentBalance: 200000000,
			Currency: "VND", InterestRate: 12, MinimumPayment: 2000000,
			StartDate: now.AddDate(-6, 0, 0), NextPaymentDate: timePtr(now.AddDate(0, 1, 0)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("VPBank"),
		},
	}

	for _, debt := range debts {
		debt.UpdateCalculatedFields()
		if err := tx.Create(debt).Error; err != nil {
			return err
		}
		s.logger.Info("✅ Created debt", zap.String("name", debt.Name))
	}

	return nil
}
