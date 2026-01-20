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
// STUDENT PROFILE - Bob Wilson
// Limited income, small goals, learning to budget
// Monthly Income: ~8M VND (allowance + part-time)
// =====================================================

func (s *Seeder) seedStudentProfile(tx *gorm.DB, userID uuid.UUID) error {
	now := time.Now()
	categoryMap := s.getCategoryMap(tx, userID)

	// 1. Budget Constraints (INPUT for DSS) - Tight budget
	constraints := []*constraintdomain.BudgetConstraint{
		// Tiền phòng - FIXED (ở ký túc xá hoặc thuê chung)
		{
			UserID: userID, CategoryID: categoryMap["Home & Utilities"],
			MinimumAmount: 2000000, MaximumAmount: 2000000, IsFlexible: false,
			Priority: 1, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Tiền phòng KTX/thuê chung",
		},
		// Ăn uống - FLEXIBLE (tự nấu/căng tin)
		{
			UserID: userID, CategoryID: categoryMap["Food & Dining"],
			MinimumAmount: 2500000, MaximumAmount: 3500000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Ăn căng tin, nấu ăn",
		},
		// Học tập - FLEXIBLE
		{
			UserID: userID, CategoryID: categoryMap["Education & Self-Dev"],
			MinimumAmount: 500000, MaximumAmount: 1500000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Sách vở, tài liệu, photo",
		},
		// Đi lại - FLEXIBLE (xe buýt)
		{
			UserID: userID, CategoryID: categoryMap["Transportation"],
			MinimumAmount: 300000, MaximumAmount: 600000, IsFlexible: true,
			Priority: 3, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Vé xe buýt, grab khi cần",
		},
		// Giải trí - CÓ THỂ CẮT
		{
			UserID: userID, CategoryID: categoryMap["Entertainment & Travel"],
			MinimumAmount: 0, MaximumAmount: 500000, IsFlexible: true,
			Priority: 5, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Cafe với bạn, xem phim",
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

	// 2. Income Profiles - Limited income
	incomes := []*incomedomain.IncomeProfile{
		{
			UserID: userID, CategoryID: categoryMap["Other Income"],
			Source: "Phụ cấp từ bố mẹ",
			Amount: 5000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-3, 0, 0),
			OtherIncome: 5000000,
			IsRecurring: true, IsVerified: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Tiền sinh hoạt phí từ gia đình",
		},
		{
			UserID: userID, CategoryID: categoryMap["Active Income"],
			Source: "Gia sư môn Toán",
			Amount: 3500000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-1, 0, 0),
			OtherIncome: 3500000,
			IsRecurring: false, IsVerified: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Dạy kèm 2 học sinh lớp 10",
		},
		{
			UserID: userID, CategoryID: categoryMap["Other Income"],
			Source: "Học bổng khuyến khích",
			Amount: 2000000, Currency: "VND", Frequency: "quarterly",
			StartDate:   now.AddDate(-6, 0, 0),
			OtherIncome: 2000000,
			IsRecurring: true, IsVerified: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Học bổng GPA 3.5+",
		},
	}

	for _, income := range incomes {
		if err := tx.Create(income).Error; err != nil {
			return err
		}
		s.logger.Info("✅ Created income", zap.String("source", income.Source))
	}

	// 3. Goals - Small, achievable
	accountID := s.getAccountID(tx, userID)
	goals := []*goaldomain.Goal{
		{
			UserID: userID, AccountID: accountID,
			Name: "Mua laptop mới", Description: strPtr("MacBook Air M2 cho đồ án"),
			TargetAmount: 28000000, CurrentAmount: 12000000,
			StartDate: now.AddDate(0, -3, 0), TargetDate: timePtr(now.AddDate(0, 4, 0)),
			Behavior: goaldomain.GoalBehaviorWilling, Category: goaldomain.GoalCategoryEducation,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Khóa IELTS 7.0", Description: strPtr("Luyện thi IELTS 6 tháng"),
			TargetAmount: 25000000, CurrentAmount: 8000000,
			StartDate: now.AddDate(0, -2, 0), TargetDate: timePtr(now.AddDate(0, 6, 0)),
			Behavior: goaldomain.GoalBehaviorRecurring, Category: goaldomain.GoalCategoryEducation,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Quỹ dự phòng nhỏ", Description: strPtr("1 tháng chi tiêu"),
			TargetAmount: 8000000, CurrentAmount: 3500000,
			StartDate: now.AddDate(0, -5, 0), TargetDate: timePtr(now.AddDate(0, 3, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryEmergency,
			Priority: goaldomain.GoalPriorityMedium, Status: goaldomain.GoalStatusActive,
		},
		{
			UserID: userID, AccountID: accountID,
			Name: "Quỹ đi phượt", Description: strPtr("Gap year Đông Nam Á"),
			TargetAmount: 30000000, CurrentAmount: 5000000,
			StartDate: now.AddDate(0, -4, 0), TargetDate: timePtr(now.AddDate(1, 0, 0)),
			Behavior: goaldomain.GoalBehaviorFlexible, Category: goaldomain.GoalCategoryTravel,
			Priority: goaldomain.GoalPriorityLow, Status: goaldomain.GoalStatusActive,
		},
	}

	for _, goal := range goals {
		goal.UpdateCalculatedFields()
		if err := tx.Create(goal).Error; err != nil {
			return err
		}
		s.logger.Info("✅ Created goal", zap.String("name", goal.Name))
	}

	// 4. Debts - Small personal loans
	debts := []*debtdomain.Debt{
		{
			UserID: userID, Name: "Nợ bạn Minh",
			Description: strPtr("Mượn mua sách"),
			Type:        debtdomain.DebtTypePersonalLoan, Behavior: debtdomain.DebtBehaviorInstallment,
			PrincipalAmount: 3000000, CurrentBalance: 1500000,
			Currency: "VND", InterestRate: 0, MinimumPayment: 500000,
			StartDate: now.AddDate(0, -2, 0), NextPaymentDate: timePtr(now.AddDate(0, 1, 0)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("Minh (bạn cùng lớp)"),
		},
		{
			UserID: userID, Name: "Trả góp iPhone 15",
			Description: strPtr("TGDD 12 tháng 0%"),
			Type:        debtdomain.DebtTypePersonalLoan, Behavior: debtdomain.DebtBehaviorInstallment,
			PrincipalAmount: 22000000, CurrentBalance: 11000000,
			Currency: "VND", InterestRate: 0, MinimumPayment: 1833333,
			StartDate: now.AddDate(0, -6, 0), NextPaymentDate: timePtr(now.AddDate(0, 0, 10)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("Thế Giới Di Động"),
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
