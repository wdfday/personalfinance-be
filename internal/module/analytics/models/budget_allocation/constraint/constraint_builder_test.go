package constraint

import (
	"testing"
	"time"

	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	budgetprofile "personalfinancedss/internal/module/cashflow/budget_profile/domain"
	debtdomain "personalfinancedss/internal/module/cashflow/debt/domain"
	goaldomain "personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"
)

func TestConstraintBuilder_BuildConstraints(t *testing.T) {
	builder := NewConstraintBuilder()

	// Setup test data
	income := 10000.0
	userID := uuid.New()
	categoryID1 := uuid.New()
	categoryID2 := uuid.New()
	goalID := uuid.New()
	debtID := uuid.New()

	// Create budget constraints
	budgetConstraints := budgetprofile.BudgetConstraints{
		{
			ID:            uuid.New(),
			UserID:        userID,
			CategoryID:    categoryID1,
			MinimumAmount: 3000.0,
			IsFlexible:    false,
			Priority:      1,
		},
		{
			ID:            uuid.New(),
			UserID:        userID,
			CategoryID:    categoryID2,
			MinimumAmount: 1000.0,
			MaximumAmount: 2000.0,
			IsFlexible:    true,
			Priority:      10,
		},
	}

	// Create goals
	now := time.Now()
	targetDate := now.AddDate(0, 6, 0) // 6 months from now
	suggestedContribution := 500.0
	goals := []*goaldomain.Goal{
		{
			ID:                    goalID,
			UserID:                userID,
			AccountID:             uuid.New(),
			Name:                  "Emergency Fund",
			Behavior:              goaldomain.GoalBehaviorFlexible,
			Category:              goaldomain.GoalCategoryEmergency,
			Priority:              goaldomain.GoalPriorityHigh,
			TargetAmount:          10000.0,
			CurrentAmount:         5000.0,
			RemainingAmount:       5000.0,
			Status:                goaldomain.GoalStatusActive,
			SuggestedContribution: &suggestedContribution,
			TargetDate:            &targetDate,
		},
	}

	// Create debts
	debts := []*debtdomain.Debt{
		{
			ID:              debtID,
			UserID:          userID,
			Name:            "Credit Card",
			Type:            debtdomain.DebtTypeCreditCard,
			Status:          debtdomain.DebtStatusActive,
			PrincipalAmount: 5000.0,
			CurrentBalance:  5000.0,
			InterestRate:    18.0, // High interest
			MinimumPayment:  200.0,
		},
	}

	// Build constraints
	model, err := builder.BuildConstraints(income, budgetConstraints, goals, debts)

	// Assertions
	if err != nil {
		t.Fatalf("BuildConstraints failed: %v", err)
	}

	if model.TotalIncome != income {
		t.Errorf("Expected income %f, got %f", income, model.TotalIncome)
	}

	// Check mandatory expenses
	if len(model.MandatoryExpenses) != 1 {
		t.Errorf("Expected 1 mandatory expense, got %d", len(model.MandatoryExpenses))
	}

	// Check flexible expenses
	if len(model.FlexibleExpenses) != 1 {
		t.Errorf("Expected 1 flexible expense, got %d", len(model.FlexibleExpenses))
	}

	// Check debt payments
	if len(model.DebtPayments) != 1 {
		t.Errorf("Expected 1 debt payment, got %d", len(model.DebtPayments))
	}

	// Check goal targets
	if len(model.GoalTargets) != 1 {
		t.Errorf("Expected 1 goal target, got %d", len(model.GoalTargets))
	}

	// Verify debt priority (high interest = high priority)
	debtConstraint := model.DebtPayments[debtID]
	if debtConstraint.Priority != 10 {
		t.Errorf("Expected debt priority 10 for 18%% interest, got %d", debtConstraint.Priority)
	}
}

func TestConstraintBuilder_CheckFeasibility(t *testing.T) {
	builder := NewConstraintBuilder()

	tests := []struct {
		name             string
		income           float64
		mandatoryAmount  float64
		debtMinimum      float64
		expectedFeasible bool
	}{
		{
			name:             "Sufficient income",
			income:           10000.0,
			mandatoryAmount:  3000.0,
			debtMinimum:      200.0,
			expectedFeasible: true,
		},
		{
			name:             "Insufficient income",
			income:           3000.0,
			mandatoryAmount:  3000.0,
			debtMinimum:      200.0,
			expectedFeasible: false,
		},
		{
			name:             "Exact match",
			income:           3200.0,
			mandatoryAmount:  3000.0,
			debtMinimum:      200.0,
			expectedFeasible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &domain.ConstraintModel{
				TotalIncome:       tt.income,
				MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
				DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
			}

			// Add mandatory expense
			model.MandatoryExpenses[uuid.New()] = domain.CategoryConstraint{
				Minimum: tt.mandatoryAmount,
			}

			// Add debt
			model.DebtPayments[uuid.New()] = domain.DebtConstraint{
				MinimumPayment: tt.debtMinimum,
			}

			isFeasible, deficit := builder.CheckFeasibility(model)

			if isFeasible != tt.expectedFeasible {
				t.Errorf("Expected feasible=%v, got %v (deficit: %f)", tt.expectedFeasible, isFeasible, deficit)
			}
		})
	}
}

func TestConstraintBuilder_CalculateSurplus(t *testing.T) {
	builder := NewConstraintBuilder()

	model := &domain.ConstraintModel{
		TotalIncome:       10000.0,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
	}

	// Add mandatory expenses (3000)
	model.MandatoryExpenses[uuid.New()] = domain.CategoryConstraint{
		Minimum: 3000.0,
	}

	// Add debt payments (200)
	model.DebtPayments[uuid.New()] = domain.DebtConstraint{
		MinimumPayment: 200.0,
	}

	surplus := builder.CalculateSurplus(model)

	expectedSurplus := 10000.0 - 3000.0 - 200.0 // 6800.0
	if surplus != expectedSurplus {
		t.Errorf("Expected surplus %f, got %f", expectedSurplus, surplus)
	}
}

func TestConstraintBuilder_CalculateDebtPriority(t *testing.T) {
	builder := NewConstraintBuilder()

	tests := []struct {
		name             string
		interestRate     float64
		expectedPriority int
	}{
		{"Very high interest (credit card)", 25.0, 1},
		{"High interest", 18.0, 10}, // >= 10% = priority 10
		{"Medium interest", 12.0, 10},
		{"Low interest", 7.0, 20},
		{"Very low interest (mortgage)", 3.0, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debt := &debtdomain.Debt{
				InterestRate: tt.interestRate,
			}

			priority := builder.calculateDebtPriority(debt)

			if priority != tt.expectedPriority {
				t.Errorf("For %f%% interest, expected priority %d, got %d",
					tt.interestRate, tt.expectedPriority, priority)
			}
		})
	}
}

func TestConstraintBuilder_CalculateGoalContribution(t *testing.T) {
	builder := NewConstraintBuilder()

	t.Run("Uses suggested contribution if set", func(t *testing.T) {
		suggested := 500.0
		goal := &goaldomain.Goal{
			SuggestedContribution: &suggested,
			RemainingAmount:       5000.0,
		}

		contribution := builder.calculateGoalContribution(goal)

		if contribution != suggested {
			t.Errorf("Expected %f, got %f", suggested, contribution)
		}
	})

	t.Run("Uses auto-contribute amount if set", func(t *testing.T) {
		autoAmount := 300.0
		goal := &goaldomain.Goal{
			AutoContributeAmount: &autoAmount,
			RemainingAmount:      5000.0,
		}

		contribution := builder.calculateGoalContribution(goal)

		if contribution != autoAmount {
			t.Errorf("Expected %f, got %f", autoAmount, contribution)
		}
	})

	t.Run("Calculates based on target date", func(t *testing.T) {
		now := time.Now()
		targetDate := now.AddDate(0, 6, 0) // 6 months
		goal := &goaldomain.Goal{
			RemainingAmount: 6000.0,
			TargetDate:      &targetDate,
		}

		contribution := builder.calculateGoalContribution(goal)

		// Should be approximately 6000 / 6 = 1000
		expected := 1000.0
		tolerance := 100.0 // Allow some tolerance due to days calculation

		if contribution < expected-tolerance || contribution > expected+tolerance {
			t.Errorf("Expected approximately %f, got %f", expected, contribution)
		}
	})

	t.Run("Defaults to 12 months if no target date", func(t *testing.T) {
		goal := &goaldomain.Goal{
			RemainingAmount: 12000.0,
		}

		contribution := builder.calculateGoalContribution(goal)

		expected := 1000.0 // 12000 / 12
		if contribution != expected {
			t.Errorf("Expected %f, got %f", expected, contribution)
		}
	})
}

func TestConstraintBuilder_GetSuggestionsForDeficit(t *testing.T) {
	builder := NewConstraintBuilder()

	model := &domain.ConstraintModel{
		TotalIncome:       5000.0,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
	}

	// Add high-interest debt
	model.DebtPayments[uuid.New()] = domain.DebtConstraint{
		MinimumPayment: 200.0,
		InterestRate:   20.0,
	}

	// Add flexible expense
	model.FlexibleExpenses[uuid.New()] = domain.CategoryConstraint{
		Minimum: 1000.0,
		Maximum: 2000.0,
	}

	deficit := 500.0
	suggestions := builder.GetSuggestionsForDeficit(model, deficit)

	// Should have multiple suggestions
	if len(suggestions) < 3 {
		t.Errorf("Expected at least 3 suggestions, got %d", len(suggestions))
	}

	// First suggestion should mention the deficit amount
	if len(suggestions[0]) == 0 {
		t.Error("First suggestion should not be empty")
	}

	// Should suggest reducing flexible expenses
	hasFlexibleSuggestion := false
	for _, s := range suggestions {
		if len(s) > 0 && (s == "Consider adjusting these flexible expense categories:") {
			hasFlexibleSuggestion = true
			break
		}
	}
	if !hasFlexibleSuggestion {
		t.Error("Should suggest adjusting flexible expenses")
	}

	// Should suggest debt consolidation for high-interest debt
	hasDebtSuggestion := false
	for _, s := range suggestions {
		if len(s) > 0 && s == "Consider debt consolidation or refinancing for high-interest debts" {
			hasDebtSuggestion = true
			break
		}
	}
	if !hasDebtSuggestion {
		t.Error("Should suggest debt consolidation for high-interest debt")
	}
}
