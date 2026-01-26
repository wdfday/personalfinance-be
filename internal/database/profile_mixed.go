package database

import (
	"fmt"
	"time"

	monthdomain "personalfinancedss/internal/module/calendar/month/domain"
	accountdomain "personalfinancedss/internal/module/cashflow/account/domain"
	budgetdomain "personalfinancedss/internal/module/cashflow/budget/domain"
	constraintdomain "personalfinancedss/internal/module/cashflow/budget_profile/domain"
	debtdomain "personalfinancedss/internal/module/cashflow/debt/domain"
	goaldomain "personalfinancedss/internal/module/cashflow/goal/domain"
	incomedomain "personalfinancedss/internal/module/cashflow/income_profile/domain"
	transactiondomain "personalfinancedss/internal/module/cashflow/transaction/domain"

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
	homeCatID, err := s.getCategoryID(categoryMap, "Home & Utilities", "mixed constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	foodCatID, err := s.getCategoryID(categoryMap, "Food & Dining", "mixed constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	transportCatID, err := s.getCategoryID(categoryMap, "Transportation", "mixed constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}
	educationCatID, err := s.getCategoryID(categoryMap, "Education & Self-Dev", "mixed constraint")
	if err != nil {
		return fmt.Errorf("failed to get category for constraints: %w", err)
	}

	constraints := []*constraintdomain.BudgetConstraint{
		// Tiền nhà - FIXED
		{
			UserID: userID, CategoryID: homeCatID,
			MinimumAmount: 4500000, MaximumAmount: 4500000, IsFlexible: false,
			Priority: 1, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Thuê phòng studio",
		},
		// Ăn uống - FLEXIBLE
		{
			UserID: userID, CategoryID: foodCatID,
			MinimumAmount: 3500000, MaximumAmount: 5000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Ăn uống cá nhân",
		},
		// Đi lại - FLEXIBLE
		{
			UserID: userID, CategoryID: transportCatID,
			MinimumAmount: 1000000, MaximumAmount: 2000000, IsFlexible: true,
			Priority: 3, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Xe Grab, taxi",
		},
		// Học tập - FLEXIBLE (MBA)
		{
			UserID: userID, CategoryID: educationCatID,
			MinimumAmount: 5000000, MaximumAmount: 8000000, IsFlexible: true,
			Priority: 2, Status: constraintdomain.ConstraintStatusActive,
			StartDate: now, Period: "monthly", Description: "Mua sắm",
		},
	}

	for _, c := range constraints {
		if err := tx.Create(c).Error; err != nil {
			return fmt.Errorf("failed to create constraint: %w", err)
		}
		s.logger.Info("✅ Created constraint", zap.Float64("min", c.MinimumAmount))
	}

	// 2. Income Profiles - Mixed stable + variable
	activeIncomeCatID, err := s.getCategoryID(categoryMap, "Active Income", "mixed income")
	if err != nil {
		return fmt.Errorf("failed to get category for income: %w", err)
	}
	otherIncomeCatID, err := s.getCategoryID(categoryMap, "Other Income", "mixed income")
	if err != nil {
		return fmt.Errorf("failed to get category for income: %w", err)
	}

	incomes := []*incomedomain.IncomeProfile{
		{
			UserID: userID, CategoryID: activeIncomeCatID,
			Source: "Lương chính - Marketing",
			Amount: 22000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-1, 6, 0),
			IsRecurring: true,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Marketing Executive tại agency",
		},
		{
			UserID: userID, CategoryID: activeIncomeCatID,
			Source: "Side hustle - Content Writing",
			Amount: 12000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-1, 0, 0),
			IsRecurring: false,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Viết content, 5-10 bài/tháng",
		},
		{
			UserID: userID, CategoryID: otherIncomeCatID,
			Source: "Affiliate Marketing",
			Amount: 5000000, Currency: "VND", Frequency: "monthly",
			StartDate:   now.AddDate(-6, 0, 0),
			IsRecurring: false,
			Status:      incomedomain.IncomeStatusActive,
			Description: "Commission giới thiệu sản phẩm",
		},
	}

	for _, income := range incomes {
		if err := tx.Create(income).Error; err != nil {
			return fmt.Errorf("failed to create income: %w", err)
		}
		s.logger.Info("✅ Created income", zap.String("source", income.Source))
	}

	// 3. Get debit account (bank account) - should be primary
	accountID, err := s.getDebitAccountID(tx, userID)
	if err != nil {
		return fmt.Errorf("failed to get debit account for goals: %w", err)
	}

	// 4. Goals
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
			Name: "Mua xe", Description: strPtr("Mua xe ..."),
			TargetAmount: 500000000, CurrentAmount: 80000000,
			StartDate: now.AddDate(-1, 0, 0), TargetDate: timePtr(now.AddDate(2, 6, 0)),
			Behavior: goaldomain.GoalBehaviorRecurring, Category: goaldomain.GoalCategoryEducation,
			Priority: goaldomain.GoalPriorityHigh, Status: goaldomain.GoalStatusActive,
		},
	}

	for _, goal := range goals {
		goal.UpdateCalculatedFields()
		if err := tx.Create(goal).Error; err != nil {
			return fmt.Errorf("failed to create goal: %w", err)
		}
		s.logger.Info("✅ Created goal", zap.String("name", goal.Name))
	}

	// 5. Debts
	debts := []*debtdomain.Debt{
		{
			UserID: userID, Name: "Credit Card VIB",
			Description: strPtr("Thẻ tín dụng chi tiêu"),
			Type:        debtdomain.DebtTypeCreditCard, Behavior: debtdomain.DebtBehaviorRevolving,
			PrincipalAmount: 25000000, CurrentBalance: 20000000,
			Currency: "VND", InterestRate: 24, MinimumPayment: 2000000,
			StartDate: now.AddDate(-1, 0, 0), NextPaymentDate: timePtr(now.AddDate(0, 0, 15)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("VIB Bank"),
		},
		{
			UserID: userID, Name: "Credit Card Techcombank",
			Description: strPtr("Thẻ tín dụng mua sắm online"),
			Type:        debtdomain.DebtTypeCreditCard, Behavior: debtdomain.DebtBehaviorRevolving,
			PrincipalAmount: 15000000, CurrentBalance: 8000000,
			Currency: "VND", InterestRate: 22, MinimumPayment: 1000000,
			StartDate: now.AddDate(0, -8, 0), NextPaymentDate: timePtr(now.AddDate(0, 0, 20)),
			Status: debtdomain.DebtStatusActive, CreditorName: strPtr("Techcombank"),
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
			return fmt.Errorf("failed to create debt: %w", err)
		}
		s.logger.Info("✅ Created debt", zap.String("name", debt.Name))
	}

	// 6. Seed Budget và Month cho tháng trước
	lastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	lastMonthEnd := lastMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
	lastMonthStr := fmt.Sprintf("%04d-%02d", lastMonth.Year(), lastMonth.Month())

	// Tạo Month cho tháng trước
	lastMonthEntity := &monthdomain.Month{
		ID:        uuid.New(),
		UserID:    userID,
		Month:     lastMonthStr,
		StartDate: lastMonth,
		EndDate:   lastMonthEnd,
		Status:    monthdomain.StatusClosed, // Tháng trước đã đóng
		States:    []monthdomain.MonthState{},
		Version:   1,
	}
	if err := tx.Create(lastMonthEntity).Error; err != nil {
		return fmt.Errorf("failed to create last month: %w", err)
	}
	s.logger.Info("✅ Created last month", zap.String("month", lastMonthStr))

	// Tạo Budgets cho tháng trước dựa trên constraints
	var budgetIDs []uuid.UUID
	var budgets []*budgetdomain.Budget
	for _, constraint := range constraints {
		budgetID := uuid.New()
		budget := &budgetdomain.Budget{
			ID:           budgetID,
			UserID:       userID,
			CategoryID:   &constraint.CategoryID,
			ConstraintID: &constraint.ID,
			Name:         fmt.Sprintf("Budget %s - %s", lastMonthStr, constraint.Description),
			Amount:       constraint.MinimumAmount, // Sử dụng minimum amount
			Currency:     "VND",
			Period:       budgetdomain.BudgetPeriodMonthly,
			StartDate:    lastMonth,
			EndDate:      &lastMonthEnd,
			Status:       budgetdomain.BudgetStatusActive, // Sẽ được cập nhật sau khi tính toán
			EnableAlerts: true,
		}
		if err := tx.Create(budget).Error; err != nil {
			return fmt.Errorf("failed to create budget: %w", err)
		}
		budgetIDs = append(budgetIDs, budgetID)
		budgets = append(budgets, budget)
		s.logger.Info("✅ Created budget", zap.String("name", budget.Name))
	}

	// 7. Seed Transactions tháng trước (đầy đủ thông tin)
	// Lấy goal IDs và debt IDs để tạo links
	var goalIDs []uuid.UUID
	var debtIDs []uuid.UUID
	for _, goal := range goals {
		goalIDs = append(goalIDs, goal.ID)
	}
	for _, debt := range debts {
		debtIDs = append(debtIDs, debt.ID)
	}

	// Transactions tháng trước - đầy đủ thông tin
	lastMonthTransactions := []*transactiondomain.Transaction{
		// Income transactions
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionCredit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         18000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 5), // Ngày 5
			ValueDate:      lastMonth.AddDate(0, 0, 5),
			Description:    "Lương chính - Marketing",
			UserNote:       "Lương tháng trước",
			UserCategoryID: &activeIncomeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Agency ABC",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkIncomeProfile, ID: incomes[0].ID.String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionCredit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         8000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 10),
			ValueDate:      lastMonth.AddDate(0, 0, 10),
			Description:    "Side hustle - Content Writing",
			UserNote:       "5 bài content",
			UserCategoryID: &activeIncomeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Client XYZ",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkIncomeProfile, ID: incomes[1].ID.String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionCredit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         3000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 15),
			ValueDate:      lastMonth.AddDate(0, 0, 15),
			Description:    "Affiliate Marketing Commission",
			UserCategoryID: &otherIncomeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Affiliate Network",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkIncomeProfile, ID: incomes[2].ID.String()},
			},
		},
		// Expense transactions - Home & Utilities
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         4500000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 1),
			ValueDate:      lastMonth.AddDate(0, 0, 1),
			Description:    "Tiền thuê phòng studio",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Chủ nhà",
				Type: "PERSON",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[0].String()},
			},
		},
		// Expense transactions - Food & Dining
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         120000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 2),
			ValueDate:      lastMonth.AddDate(0, 0, 2),
			Description:    "Cơm trưa văn phòng",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng ABC",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         85000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 3),
			ValueDate:      lastMonth.AddDate(0, 0, 3),
			Description:    "Cà phê sáng",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Highlands Coffee",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         250000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 5),
			ValueDate:      lastMonth.AddDate(0, 0, 5),
			Description:    "Đi ăn tối với bạn",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng XYZ",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         180000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 8),
			ValueDate:      lastMonth.AddDate(0, 0, 8),
			Description:    "GrabFood",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         95000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 12),
			ValueDate:      lastMonth.AddDate(0, 0, 12),
			Description:    "Bánh mì sáng",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Bánh mì ABC",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         150000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 15),
			ValueDate:      lastMonth.AddDate(0, 0, 15),
			Description:    "Đi chợ mua đồ",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Siêu thị Coopmart",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         200000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 20),
			ValueDate:      lastMonth.AddDate(0, 0, 20),
			Description:    "Đi ăn buffet",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Buffet ABC",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         110000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 22),
			ValueDate:      lastMonth.AddDate(0, 0, 22),
			Description:    "Cơm tối",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng DEF",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         75000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 25),
			ValueDate:      lastMonth.AddDate(0, 0, 25),
			Description:    "Trà sữa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Gong Cha",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         130000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 28),
			ValueDate:      lastMonth.AddDate(0, 0, 28),
			Description:    "Bữa trưa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng GHI",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		// Expense transactions - Transportation
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         35000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 1),
			ValueDate:      lastMonth.AddDate(0, 0, 1),
			Description:    "Grab đi làm",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         45000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 3),
			ValueDate:      lastMonth.AddDate(0, 0, 3),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         500000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 5),
			ValueDate:      lastMonth.AddDate(0, 0, 5),
			Description:    "Nạp tiền xe máy",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Trạm xăng",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         40000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 7),
			ValueDate:      lastMonth.AddDate(0, 0, 7),
			Description:    "Grab đi họp",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         30000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 10),
			ValueDate:      lastMonth.AddDate(0, 0, 10),
			Description:    "Grab đi chơi",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         550000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 15),
			ValueDate:      lastMonth.AddDate(0, 0, 15),
			Description:    "Nạp xăng lần 2",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Trạm xăng",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         38000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 18),
			ValueDate:      lastMonth.AddDate(0, 0, 18),
			Description:    "Grab đi siêu thị",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         42000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 22),
			ValueDate:      lastMonth.AddDate(0, 0, 22),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         48000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 25),
			ValueDate:      lastMonth.AddDate(0, 0, 25),
			Description:    "Grab đi học",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         500000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 28),
			ValueDate:      lastMonth.AddDate(0, 0, 28),
			Description:    "Nạp xăng lần 3",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Trạm xăng",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		// Expense transactions - Education
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelWeb,
			Amount:         6000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 1),
			ValueDate:      lastMonth.AddDate(0, 0, 1),
			Description:    "Học phí MBA tháng này",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "RMIT University",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         250000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 5),
			ValueDate:      lastMonth.AddDate(0, 0, 5),
			Description:    "Mua sách MBA",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Fahasa",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         150000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 12),
			ValueDate:      lastMonth.AddDate(0, 0, 12),
			Description:    "Mua tài liệu online",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Amazon",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		// Goal contributions
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         5000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 6),
			ValueDate:      lastMonth.AddDate(0, 0, 6),
			Description:    "Tiết kiệm quỹ dự phòng",
			UserCategoryID: &homeCatID, // Tạm dùng category khác
			Counterparty: &transactiondomain.Counterparty{
				Name: "Internal Transfer",
				Type: "INTERNAL",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkGoal, ID: goalIDs[0].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         2000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 10),
			ValueDate:      lastMonth.AddDate(0, 0, 10),
			Description:    "Tiết kiệm du lịch Hàn Quốc",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Internal Transfer",
				Type: "INTERNAL",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkGoal, ID: goalIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         3000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 15),
			ValueDate:      lastMonth.AddDate(0, 0, 15),
			Description:    "Tiết kiệm học MBA",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Internal Transfer",
				Type: "INTERNAL",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkGoal, ID: goalIDs[2].String()},
			},
		},
		// Debt payments
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelWeb,
			Amount:         2000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 15),
			ValueDate:      lastMonth.AddDate(0, 0, 15),
			Description:    "Thanh toán thẻ tín dụng VIB",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "VIB Bank",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkDebt, ID: debtIDs[0].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelWeb,
			Amount:         2000000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 1),
			ValueDate:      lastMonth.AddDate(0, 0, 1),
			Description:    "Trả lãi vay học MBA",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "VPBank",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkDebt, ID: debtIDs[1].String()},
			},
		},
		// Thêm nhiều transactions nhỏ để đạt ~70 transactions/tháng
		// Food & Dining - thêm transactions để đạt ~4.2M (trong khoảng 3.5M-5M)
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         90000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 4),
			ValueDate:      lastMonth.AddDate(0, 0, 4),
			Description:    "Bánh mì trưa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Bánh mì DEF",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         140000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 6),
			ValueDate:      lastMonth.AddDate(0, 0, 6),
			Description:    "Cơm tối",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng GHI",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         80000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 7),
			ValueDate:      lastMonth.AddDate(0, 0, 7),
			Description:    "Trà sữa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Tocotoco",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         160000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 9),
			ValueDate:      lastMonth.AddDate(0, 0, 9),
			Description:    "Bữa trưa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng JKL",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         220000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 11),
			ValueDate:      lastMonth.AddDate(0, 0, 11),
			Description:    "Đi ăn với đồng nghiệp",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng MNO",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         105000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 13),
			ValueDate:      lastMonth.AddDate(0, 0, 13),
			Description:    "Cơm tối",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng PQR",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         125000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 14),
			ValueDate:      lastMonth.AddDate(0, 0, 14),
			Description:    "Bữa sáng",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Quán cơm STU",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         190000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 16),
			ValueDate:      lastMonth.AddDate(0, 0, 16),
			Description:    "GrabFood",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         165000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 17),
			ValueDate:      lastMonth.AddDate(0, 0, 17),
			Description:    "Đi chợ",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "VinMart",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         115000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 19),
			ValueDate:      lastMonth.AddDate(0, 0, 19),
			Description:    "Cơm trưa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng VWX",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         135000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 21),
			ValueDate:      lastMonth.AddDate(0, 0, 21),
			Description:    "Bữa tối",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng YZA",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         100000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 23),
			ValueDate:      lastMonth.AddDate(0, 0, 23),
			Description:    "Cà phê chiều",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Starbucks",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         175000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 24),
			ValueDate:      lastMonth.AddDate(0, 0, 24),
			Description:    "Đi ăn tối",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng BCD",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         145000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 26),
			ValueDate:      lastMonth.AddDate(0, 0, 26),
			Description:    "Bữa trưa",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng EFG",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         120000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 27),
			ValueDate:      lastMonth.AddDate(0, 0, 27),
			Description:    "Cơm tối",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà hàng HIJ",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         90000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 29),
			ValueDate:      lastMonth.AddDate(0, 0, 29),
			Description:    "Bánh mì sáng",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Bánh mì KLM",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         155000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 30),
			ValueDate:      lastMonth.AddDate(0, 0, 30),
			Description:    "Đi chợ cuối tháng",
			UserCategoryID: &foodCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Big C",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[1].String()},
			},
		},
		// Transportation - thêm transactions để đạt ~1.8M (trong khoảng 1M-2M)
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         36000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 2),
			ValueDate:      lastMonth.AddDate(0, 0, 2),
			Description:    "Grab đi làm",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         41000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 4),
			ValueDate:      lastMonth.AddDate(0, 0, 4),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         33000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 6),
			ValueDate:      lastMonth.AddDate(0, 0, 6),
			Description:    "Grab đi họp",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         39000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 8),
			ValueDate:      lastMonth.AddDate(0, 0, 8),
			Description:    "Grab đi chơi",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         37000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 9),
			ValueDate:      lastMonth.AddDate(0, 0, 9),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         34000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 11),
			ValueDate:      lastMonth.AddDate(0, 0, 11),
			Description:    "Grab đi làm",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         43000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 13),
			ValueDate:      lastMonth.AddDate(0, 0, 13),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         35000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 14),
			ValueDate:      lastMonth.AddDate(0, 0, 14),
			Description:    "Grab đi học",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         44000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 16),
			ValueDate:      lastMonth.AddDate(0, 0, 16),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         32000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 17),
			ValueDate:      lastMonth.AddDate(0, 0, 17),
			Description:    "Grab đi làm",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         46000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 19),
			ValueDate:      lastMonth.AddDate(0, 0, 19),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         38000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 20),
			ValueDate:      lastMonth.AddDate(0, 0, 20),
			Description:    "Grab đi chơi",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         36000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 21),
			ValueDate:      lastMonth.AddDate(0, 0, 21),
			Description:    "Grab đi làm",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         40000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 23),
			ValueDate:      lastMonth.AddDate(0, 0, 23),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         37000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 24),
			ValueDate:      lastMonth.AddDate(0, 0, 24),
			Description:    "Grab đi học",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         35000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 26),
			ValueDate:      lastMonth.AddDate(0, 0, 26),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         39000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 27),
			ValueDate:      lastMonth.AddDate(0, 0, 27),
			Description:    "Grab đi làm",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         41000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 29),
			ValueDate:      lastMonth.AddDate(0, 0, 29),
			Description:    "Grab về nhà",
			UserCategoryID: &transportCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Grab",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[2].String()},
			},
		},
		// Education - thêm transactions để đạt ~7M (trong khoảng 5M-8M)
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         200000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 8),
			ValueDate:      lastMonth.AddDate(0, 0, 8),
			Description:    "Mua sách tham khảo",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Tiki",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         300000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 14),
			ValueDate:      lastMonth.AddDate(0, 0, 14),
			Description:    "Đăng ký khóa học online",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Coursera",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         180000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 18),
			ValueDate:      lastMonth.AddDate(0, 0, 18),
			Description:    "Mua tài liệu in",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Tiệm photo",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         120000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 20),
			ValueDate:      lastMonth.AddDate(0, 0, 20),
			Description:    "Mua bút viết",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà sách",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         220000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 24),
			ValueDate:      lastMonth.AddDate(0, 0, 24),
			Description:    "Mua sách giáo trình",
			UserCategoryID: &educationCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Fahasa",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[3].String()},
			},
		},
		// Thêm các transactions khác (shopping, entertainment, etc.) để đạt ~70 transactions
		// Shopping & Personal Care
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         350000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 6),
			ValueDate:      lastMonth.AddDate(0, 0, 6),
			Description:    "Mua quần áo",
			UserCategoryID: &homeCatID, // Tạm dùng category khác
			Counterparty: &transactiondomain.Counterparty{
				Name: "Uniqlo",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         250000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 11),
			ValueDate:      lastMonth.AddDate(0, 0, 11),
			Description:    "Mua mỹ phẩm",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Sephora",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         180000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 16),
			ValueDate:      lastMonth.AddDate(0, 0, 16),
			Description:    "Cắt tóc",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Salon ABC",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         200000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 21),
			ValueDate:      lastMonth.AddDate(0, 0, 21),
			Description:    "Mua đồ dùng cá nhân",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Guardian",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         150000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 26),
			ValueDate:      lastMonth.AddDate(0, 0, 26),
			Description:    "Mua thuốc",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Nhà thuốc",
				Type: "MERCHANT",
			},
		},
		// Utilities & Bills
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         500000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 3),
			ValueDate:      lastMonth.AddDate(0, 0, 3),
			Description:    "Tiền điện",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "EVN",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[0].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         200000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 4),
			ValueDate:      lastMonth.AddDate(0, 0, 4),
			Description:    "Tiền nước",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "SAWACO",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[0].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         300000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 5),
			ValueDate:      lastMonth.AddDate(0, 0, 5),
			Description:    "Internet",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "VNPT",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[0].String()},
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentBankAccount,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         150000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 6),
			ValueDate:      lastMonth.AddDate(0, 0, 6),
			Description:    "Điện thoại",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Viettel",
				Type: "MERCHANT",
			},
			Links: &transactiondomain.TransactionLinks{
				{Type: transactiondomain.LinkBudget, ID: budgetIDs[0].String()},
			},
		},
		// Entertainment
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         300000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 7),
			ValueDate:      lastMonth.AddDate(0, 0, 7),
			Description:    "Xem phim",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "CGV",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         400000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 13),
			ValueDate:      lastMonth.AddDate(0, 0, 13),
			Description:    "Đi cà phê với bạn",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "The Coffee House",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         250000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 19),
			ValueDate:      lastMonth.AddDate(0, 0, 19),
			Description:    "Karaoke",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Karaoke ABC",
				Type: "MERCHANT",
			},
		},
		{
			ID:             uuid.New(),
			UserID:         userID,
			AccountID:      accountID,
			Direction:      transactiondomain.DirectionDebit,
			Source:         transactiondomain.SourceManual,
			Instrument:     transactiondomain.InstrumentEWallet,
			Channel:        transactiondomain.ChannelMobile,
			Amount:         350000,
			Currency:       "VND",
			BookingDate:    lastMonth.AddDate(0, 0, 25),
			ValueDate:      lastMonth.AddDate(0, 0, 25),
			Description:    "Đi chơi cuối tuần",
			UserCategoryID: &homeCatID,
			Counterparty: &transactiondomain.Counterparty{
				Name: "Khu vui chơi",
				Type: "MERCHANT",
			},
		},
	}

	// Tạo transactions tháng trước
	for _, txn := range lastMonthTransactions {
		if err := tx.Create(txn).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
		s.logger.Info("✅ Created transaction", zap.String("description", txn.Description))
	}

	// Tính toán và cập nhật budgets tháng trước sau khi tạo transactions
	// Budget tháng trước đã kết thúc nên cần tính đầy đủ spent_amount, remaining_amount, percentage_spent
	for _, budget := range budgets {
		var spentAmount float64

		// Tính spent_amount từ transactions đã tạo có link đến budget này
		// Tính trực tiếp từ lastMonthTransactions thay vì query lại từ DB để tránh lỗi transaction
		for _, txn := range lastMonthTransactions {
			// Chỉ tính DEBIT transactions
			if txn.Direction != transactiondomain.DirectionDebit {
				continue
			}

			// Kiểm tra booking_date trong khoảng thời gian budget
			if txn.BookingDate.Before(budget.StartDate) || txn.BookingDate.After(*budget.EndDate) {
				continue
			}

			// Kiểm tra category nếu budget có category
			if budget.CategoryID != nil {
				if txn.UserCategoryID == nil || *txn.UserCategoryID != *budget.CategoryID {
					continue
				}
			}

			// Kiểm tra link BUDGET với budget ID này
			hasBudgetLink := false
			if txn.Links != nil {
				for _, link := range *txn.Links {
					if link.Type == transactiondomain.LinkBudget && link.ID == budget.ID.String() {
						hasBudgetLink = true
						break
					}
				}
			}

			if hasBudgetLink {
				spentAmount += float64(txn.Amount)
			}
		}

		// Cập nhật các trường tracking
		budget.SpentAmount = spentAmount
		budget.RemainingAmount = budget.Amount - spentAmount
		if budget.Amount > 0 {
			budget.PercentageSpent = (spentAmount / budget.Amount) * 100
		} else {
			budget.PercentageSpent = 0
		}

		// Budget tháng trước đã kết thúc nên set status = "ended"
		budget.Status = budgetdomain.BudgetStatusEnded

		// Set LastCalculatedAt
		now := time.Now()
		budget.LastCalculatedAt = &now

		// Cập nhật budget
		if err := tx.Save(budget).Error; err != nil {
			s.logger.Warn("Failed to update budget tracking fields",
				zap.String("budget_id", budget.ID.String()),
				zap.Error(err))
		} else {
			s.logger.Info("✅ Updated budget tracking",
				zap.String("budget_name", budget.Name),
				zap.Float64("spent_amount", spentAmount),
				zap.Float64("remaining_amount", budget.RemainingAmount),
				zap.Float64("percentage_spent", budget.PercentageSpent),
				zap.String("status", string(budget.Status)))
		}
	}

	// 8. Seed GoalContributions cho tháng trước
	// 8.1. Từ transactions có link GOAL
	for _, txn := range lastMonthTransactions {
		if txn.Links != nil {
			for _, link := range *txn.Links {
				if link.Type == transactiondomain.LinkGoal {
					goalID, err := uuid.Parse(link.ID)
					if err != nil {
						continue
					}
					contribution := goaldomain.NewDeposit(
						goalID,
						accountID,
						userID,
						float64(txn.Amount), // Amount is already in VND (smallest unit), convert to float
						strPtr(fmt.Sprintf("Contribution from transaction: %s", txn.Description)),
					)
					if err := tx.Create(contribution).Error; err != nil {
						s.logger.Warn("Failed to create goal contribution", zap.Error(err))
						// Don't fail, just log
					} else {
						s.logger.Info("✅ Created goal contribution", zap.String("goal_id", goalID.String()))
					}
				}
			}
		}
	}

	// 8.2. Thêm các goal contributions trực tiếp (không qua transactions)
	// Quỹ dự phòng - thêm contributions nhỏ hơn trong tháng
	goalContributions := []*goaldomain.GoalContribution{
		// Quỹ dự phòng 3 tháng (goalIDs[0])
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[0],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    2000000, // 2M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm đầu tháng"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 3),
		},
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[0],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    1500000, // 1.5M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm giữa tháng"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 15),
		},
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[0],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    1500000, // 1.5M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm cuối tháng"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 28),
		},
		// Du lịch Hàn Quốc (goalIDs[1])
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[1],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    1000000, // 1M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm du lịch"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 7),
		},
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[1],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    1000000, // 1M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm du lịch"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 21),
		},
		// Học MBA (goalIDs[2])
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[2],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    1500000, // 1.5M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm học phí MBA"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 8),
		},
		{
			ID:        uuid.New(),
			GoalID:    goalIDs[2],
			AccountID: accountID,
			UserID:    userID,
			Type:      goaldomain.ContributionTypeDeposit,
			Amount:    1500000, // 1.5M
			Currency:  "VND",
			Note:      strPtr("Tiết kiệm học phí MBA"),
			Source:    "manual",
			CreatedAt: lastMonth.AddDate(0, 0, 22),
		},
	}

	// Tạo goal contributions
	for _, contribution := range goalContributions {
		if err := tx.Create(contribution).Error; err != nil {
			s.logger.Warn("Failed to create goal contribution", zap.Error(err))
			// Don't fail, just log
		} else {
			s.logger.Info("✅ Created goal contribution",
				zap.String("goal_id", contribution.GoalID.String()),
				zap.Float64("amount", contribution.Amount))
		}
	}

	// 9. Seed Transactions tháng này (chưa có Category và TransactionLink)
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thisMonthTransactions := []*transactiondomain.Transaction{
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionCredit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      18000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 5),
			ValueDate:   thisMonthStart.AddDate(0, 0, 5),
			Description: "Lương chính - Marketing",
			UserNote:    "Lương tháng này",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      120000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 1),
			ValueDate:   thisMonthStart.AddDate(0, 0, 1),
			Description: "Cơm trưa",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      85000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 2),
			ValueDate:   thisMonthStart.AddDate(0, 0, 2),
			Description: "Cà phê sáng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      35000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 1),
			ValueDate:   thisMonthStart.AddDate(0, 0, 1),
			Description: "Grab đi làm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      45000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 3),
			ValueDate:   thisMonthStart.AddDate(0, 0, 3),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      4500000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 1),
			ValueDate:   thisMonthStart.AddDate(0, 0, 1),
			Description: "Tiền thuê phòng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      180000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 4),
			ValueDate:   thisMonthStart.AddDate(0, 0, 4),
			Description: "GrabFood",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      250000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 5),
			ValueDate:   thisMonthStart.AddDate(0, 0, 5),
			Description: "Đi ăn tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      500000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 6),
			ValueDate:   thisMonthStart.AddDate(0, 0, 6),
			Description: "Nạp xăng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      95000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 7),
			ValueDate:   thisMonthStart.AddDate(0, 0, 7),
			Description: "Bánh mì sáng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      150000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 8),
			ValueDate:   thisMonthStart.AddDate(0, 0, 8),
			Description: "Đi chợ",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      110000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 9),
			ValueDate:   thisMonthStart.AddDate(0, 0, 9),
			Description: "Cơm tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      75000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Trà sữa",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      40000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 11),
			ValueDate:   thisMonthStart.AddDate(0, 0, 11),
			Description: "Grab đi chơi",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      130000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 12),
			ValueDate:   thisMonthStart.AddDate(0, 0, 12),
			Description: "Bữa trưa",
			// No Category, No Links
		},
		// Thêm nhiều transactions để đạt ~70 transactions tháng này
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      140000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 3),
			ValueDate:   thisMonthStart.AddDate(0, 0, 3),
			Description: "Cơm tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      80000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 4),
			ValueDate:   thisMonthStart.AddDate(0, 0, 4),
			Description: "Trà sữa",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      160000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 5),
			ValueDate:   thisMonthStart.AddDate(0, 0, 5),
			Description: "Bữa trưa",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      220000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 6),
			ValueDate:   thisMonthStart.AddDate(0, 0, 6),
			Description: "Đi ăn với đồng nghiệp",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      105000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 7),
			ValueDate:   thisMonthStart.AddDate(0, 0, 7),
			Description: "Cơm tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      125000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 8),
			ValueDate:   thisMonthStart.AddDate(0, 0, 8),
			Description: "Bữa sáng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      190000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 9),
			ValueDate:   thisMonthStart.AddDate(0, 0, 9),
			Description: "GrabFood",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      165000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Đi chợ",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      115000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 11),
			ValueDate:   thisMonthStart.AddDate(0, 0, 11),
			Description: "Cơm trưa",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      135000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 13),
			ValueDate:   thisMonthStart.AddDate(0, 0, 13),
			Description: "Bữa tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      100000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 14),
			ValueDate:   thisMonthStart.AddDate(0, 0, 14),
			Description: "Cà phê chiều",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      175000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 15),
			ValueDate:   thisMonthStart.AddDate(0, 0, 15),
			Description: "Đi ăn tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      145000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 16),
			ValueDate:   thisMonthStart.AddDate(0, 0, 16),
			Description: "Bữa trưa",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      120000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 17),
			ValueDate:   thisMonthStart.AddDate(0, 0, 17),
			Description: "Cơm tối",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      90000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 18),
			ValueDate:   thisMonthStart.AddDate(0, 0, 18),
			Description: "Bánh mì sáng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      155000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 19),
			ValueDate:   thisMonthStart.AddDate(0, 0, 19),
			Description: "Đi chợ",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      36000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 2),
			ValueDate:   thisMonthStart.AddDate(0, 0, 2),
			Description: "Grab đi làm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      41000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 4),
			ValueDate:   thisMonthStart.AddDate(0, 0, 4),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      33000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 6),
			ValueDate:   thisMonthStart.AddDate(0, 0, 6),
			Description: "Grab đi họp",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      39000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 7),
			ValueDate:   thisMonthStart.AddDate(0, 0, 7),
			Description: "Grab đi chơi",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      37000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 8),
			ValueDate:   thisMonthStart.AddDate(0, 0, 8),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      34000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 9),
			ValueDate:   thisMonthStart.AddDate(0, 0, 9),
			Description: "Grab đi làm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      43000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      35000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 11),
			ValueDate:   thisMonthStart.AddDate(0, 0, 11),
			Description: "Grab đi học",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      44000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 12),
			ValueDate:   thisMonthStart.AddDate(0, 0, 12),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      32000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 13),
			ValueDate:   thisMonthStart.AddDate(0, 0, 13),
			Description: "Grab đi làm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      46000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 14),
			ValueDate:   thisMonthStart.AddDate(0, 0, 14),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      38000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 15),
			ValueDate:   thisMonthStart.AddDate(0, 0, 15),
			Description: "Grab đi chơi",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      36000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 16),
			ValueDate:   thisMonthStart.AddDate(0, 0, 16),
			Description: "Grab đi làm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      40000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 17),
			ValueDate:   thisMonthStart.AddDate(0, 0, 17),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      37000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 18),
			ValueDate:   thisMonthStart.AddDate(0, 0, 18),
			Description: "Grab đi học",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      35000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 19),
			ValueDate:   thisMonthStart.AddDate(0, 0, 19),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      39000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 20),
			ValueDate:   thisMonthStart.AddDate(0, 0, 20),
			Description: "Grab đi làm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      41000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 21),
			ValueDate:   thisMonthStart.AddDate(0, 0, 21),
			Description: "Grab về nhà",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      550000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 7),
			ValueDate:   thisMonthStart.AddDate(0, 0, 7),
			Description: "Nạp xăng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      500000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 14),
			ValueDate:   thisMonthStart.AddDate(0, 0, 14),
			Description: "Nạp xăng lần 2",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      500000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 2),
			ValueDate:   thisMonthStart.AddDate(0, 0, 2),
			Description: "Tiền điện",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      200000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 3),
			ValueDate:   thisMonthStart.AddDate(0, 0, 3),
			Description: "Tiền nước",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      300000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 4),
			ValueDate:   thisMonthStart.AddDate(0, 0, 4),
			Description: "Internet",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      150000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 5),
			ValueDate:   thisMonthStart.AddDate(0, 0, 5),
			Description: "Điện thoại",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      350000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 8),
			ValueDate:   thisMonthStart.AddDate(0, 0, 8),
			Description: "Mua quần áo",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      250000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 9),
			ValueDate:   thisMonthStart.AddDate(0, 0, 9),
			Description: "Mua mỹ phẩm",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      300000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Xem phim",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      400000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 11),
			ValueDate:   thisMonthStart.AddDate(0, 0, 11),
			Description: "Đi cà phê với bạn",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      250000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 12),
			ValueDate:   thisMonthStart.AddDate(0, 0, 12),
			Description: "Karaoke",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      6000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 1),
			ValueDate:   thisMonthStart.AddDate(0, 0, 1),
			Description: "Học phí MBA",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      250000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 6),
			ValueDate:   thisMonthStart.AddDate(0, 0, 6),
			Description: "Mua sách MBA",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      150000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 7),
			ValueDate:   thisMonthStart.AddDate(0, 0, 7),
			Description: "Mua tài liệu online",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      2000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 15),
			ValueDate:   thisMonthStart.AddDate(0, 0, 15),
			Description: "Thanh toán thẻ tín dụng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      2000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 1),
			ValueDate:   thisMonthStart.AddDate(0, 0, 1),
			Description: "Trả lãi vay học MBA",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      5000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 6),
			ValueDate:   thisMonthStart.AddDate(0, 0, 6),
			Description: "Tiết kiệm quỹ dự phòng",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      2000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Tiết kiệm du lịch",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      3000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 15),
			ValueDate:   thisMonthStart.AddDate(0, 0, 15),
			Description: "Tiết kiệm học MBA",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      180000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 8),
			ValueDate:   thisMonthStart.AddDate(0, 0, 8),
			Description: "Cắt tóc",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      200000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 9),
			ValueDate:   thisMonthStart.AddDate(0, 0, 9),
			Description: "Mua đồ dùng cá nhân",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      150000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Mua thuốc",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      350000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 11),
			ValueDate:   thisMonthStart.AddDate(0, 0, 11),
			Description: "Đi chơi cuối tuần",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      200000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 12),
			ValueDate:   thisMonthStart.AddDate(0, 0, 12),
			Description: "Mua sách tham khảo",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      300000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 13),
			ValueDate:   thisMonthStart.AddDate(0, 0, 13),
			Description: "Đăng ký khóa học online",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      180000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 14),
			ValueDate:   thisMonthStart.AddDate(0, 0, 14),
			Description: "Mua tài liệu in",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      120000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 15),
			ValueDate:   thisMonthStart.AddDate(0, 0, 15),
			Description: "Mua bút viết",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      220000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 16),
			ValueDate:   thisMonthStart.AddDate(0, 0, 16),
			Description: "Mua sách giáo trình",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionDebit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentEWallet,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      8000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 10),
			ValueDate:   thisMonthStart.AddDate(0, 0, 10),
			Description: "Side hustle - Content Writing",
			// No Category, No Links
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			AccountID:   accountID,
			Direction:   transactiondomain.DirectionCredit,
			Source:      transactiondomain.SourceManual,
			Instrument:  transactiondomain.InstrumentBankAccount,
			Channel:     transactiondomain.ChannelMobile,
			Amount:      3000000,
			Currency:    "VND",
			BookingDate: thisMonthStart.AddDate(0, 0, 15),
			ValueDate:   thisMonthStart.AddDate(0, 0, 15),
			Description: "Affiliate Marketing Commission",
			// No Category, No Links
		},
	}

	// Tạo transactions tháng này
	for _, txn := range thisMonthTransactions {
		if err := tx.Create(txn).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
		s.logger.Info("✅ Created transaction (this month, no category)", zap.String("description", txn.Description))
	}

	// 10. Calculate and update account balance based on all transactions
	// Starting balance was set in seedAccountsForUsers (3.5x monthly income = ~101.5M)
	// Calculate net flow from transactions and update final balance

	var totalCredit int64 = 0
	var totalDebit int64 = 0

	// Last month transactions
	for _, txn := range lastMonthTransactions {
		if txn.Direction == transactiondomain.DirectionCredit {
			totalCredit += txn.Amount
		} else {
			totalDebit += txn.Amount
		}
	}

	// This month transactions
	for _, txn := range thisMonthTransactions {
		if txn.Direction == transactiondomain.DirectionCredit {
			totalCredit += txn.Amount
		} else {
			totalDebit += txn.Amount
		}
	}

	// Get current balance from account
	var currentBalance float64
	if err := tx.Table("accounts").Select("current_balance").Where("id = ?", accountID).Scan(&currentBalance).Error; err != nil {
		s.logger.Warn("Failed to get current account balance", zap.Error(err))
		currentBalance = 0
	}

	// Calculate final balance: starting balance + net flow from transactions
	netFlow := float64(totalCredit - totalDebit)
	finalBalance := currentBalance + netFlow

	// Update account balance
	if err := tx.Model(&accountdomain.Account{}).
		Where("id = ?", accountID).
		Update("current_balance", finalBalance).Error; err != nil {
		s.logger.Warn("Failed to update account balance", zap.Error(err))
		// Don't fail, just log
	} else {
		s.logger.Info("✅ Updated account balance",
			zap.String("account_id", accountID.String()),
			zap.Float64("starting_balance", currentBalance),
			zap.Float64("net_flow", netFlow),
			zap.Float64("final_balance", finalBalance),
			zap.Int64("total_credit", totalCredit),
			zap.Int64("total_debit", totalDebit))
	}

	s.logger.Info("✅ Completed seeding mixed profile with comprehensive data")
	return nil
}
