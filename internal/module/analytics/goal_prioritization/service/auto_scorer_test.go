package service

import (
	"testing"
	"time"

	goal_domain "personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/stretchr/testify/assert"
)

func TestAutoScorer_CalculateUrgency(t *testing.T) {
	scorer := NewAutoScorer()

	t.Run("Very urgent - 1 month deadline", func(t *testing.T) {
		targetDate := time.Now().AddDate(0, 1, 0) // 1 month from now
		goal := &GoalLike{
			TargetDate: &targetDate,
		}

		urgency := scorer.CalculateUrgency(goal)

		// 30 days / 365 ≈ 0.082, so urgency ≈ 1 - 0.082 = 0.918
		assert.InDelta(t, 0.918, urgency, 0.05, "1-month deadline should be very urgent")
		assert.Greater(t, urgency, 0.8, "Should be > 0.8")
	})

	t.Run("Moderately urgent - 6 months deadline", func(t *testing.T) {
		targetDate := time.Now().AddDate(0, 6, 0)
		goal := &GoalLike{
			TargetDate: &targetDate,
		}

		urgency := scorer.CalculateUrgency(goal)

		// 180 days / 365 ≈ 0.493, so urgency ≈ 0.507
		assert.InDelta(t, 0.507, urgency, 0.05, "6-month deadline")
		assert.Greater(t, urgency, 0.4)
		assert.Less(t, urgency, 0.6)
	})

	t.Run("Not urgent - 2 years deadline", func(t *testing.T) {
		targetDate := time.Now().AddDate(2, 0, 0)
		goal := &GoalLike{
			TargetDate: &targetDate,
		}

		urgency := scorer.CalculateUrgency(goal)

		// 730 days / 365 = 2.0, capped at 1.0, so urgency = 1 - 1 = 0.0
		assert.InDelta(t, 0.0, urgency, 0.05, "2-year deadline should have low urgency")
	})

	t.Run("Overdue goal - maximum urgency", func(t *testing.T) {
		targetDate := time.Now().AddDate(0, -1, 0) // 1 month ago
		goal := &GoalLike{
			TargetDate: &targetDate,
		}

		urgency := scorer.CalculateUrgency(goal)

		assert.Equal(t, 1.0, urgency, "Overdue should have maximum urgency")
	})

	t.Run("No deadline - low urgency", func(t *testing.T) {
		goal := &GoalLike{
			TargetDate: nil,
		}

		urgency := scorer.CalculateUrgency(goal)

		assert.Equal(t, 0.1, urgency, "No deadline should have low urgency")
	})
}

func TestAutoScorer_CalculateFeasibility(t *testing.T) {
	scorer := NewAutoScorer()

	t.Run("Very feasible - income exceeds requirement", func(t *testing.T) {
		targetDate := time.Now().AddDate(0, 10, 0) // 10 months
		goal := &GoalLike{
			TargetAmount:    20_000_000, // 20M VND
			CurrentAmount:   0,
			RemainingAmount: 20_000_000,
			TargetDate:      &targetDate,
			Status:          GoalStatusActive,
		}

		// Required: 20M / 10 months = 2M/month
		// Income: 50M
		// Feasibility: min(1, 50/2) = 1.0
		feasibility := scorer.CalculateFeasibility(goal, 50_000_000)

		assert.Equal(t, 1.0, feasibility, "Should be very feasible")
	})

	t.Run("Challenging - income below requirement", func(t *testing.T) {
		targetDate := time.Now().AddDate(1, 0, 0) // 12 months
		goal := &GoalLike{
			TargetAmount:    960_000_000, // 960M VND
			CurrentAmount:   0,
			RemainingAmount: 960_000_000,
			TargetDate:      &targetDate,
			Status:          GoalStatusActive,
		}

		// Required: 960M / 12 months = 80M/month
		// Income: 50M
		// Feasibility: min(1, 50/80) = 0.625
		feasibility := scorer.CalculateFeasibility(goal, 50_000_000)

		assert.InDelta(t, 0.625, feasibility, 0.01, "Should be challenging")
	})

	t.Run("Already completed - very feasible", func(t *testing.T) {
		goal := &GoalLike{
			TargetAmount:  20_000_000,
			CurrentAmount: 25_000_000, // Exceeded
			Status:        goal_domain.GoalStatusCompleted,
		}

		feasibility := scorer.CalculateFeasibility(goal, 50_000_000)

		assert.Equal(t, 1.0, feasibility, "Completed goal is feasible")
	})

	t.Run("No deadline - neutral feasibility", func(t *testing.T) {
		goal := &GoalLike{
			TargetAmount:  20_000_000,
			CurrentAmount: 0,
			TargetDate:    nil,
		}

		feasibility := scorer.CalculateFeasibility(goal, 50_000_000)

		assert.Equal(t, 0.5, feasibility, "No deadline = neutral")
	})

	t.Run("Invalid income - not feasible", func(t *testing.T) {
		targetDate := time.Now().AddDate(0, 10, 0)
		goal := &GoalLike{
			TargetAmount: 20_000_000,
			TargetDate:   &targetDate,
		}

		feasibility := scorer.CalculateFeasibility(goal, 0)
		assert.Equal(t, 0.0, feasibility, "Zero income")

		feasibility = scorer.CalculateFeasibility(goal, -1000)
		assert.Equal(t, 0.0, feasibility, "Negative income")
	})
}

func TestAutoScorer_CalculateImportance(t *testing.T) {
	scorer := NewAutoScorer()

	tests := []struct {
		name               string
		category           GoalCategory
		priority           GoalPriority
		expectedImportance float64
		tolerance          float64
	}{
		{
			name:               "Emergency fund - critical",
			category:           GoalCategoryEmergency,
			priority:           GoalPriorityMedium,
			expectedImportance: 1.00,
			tolerance:          0.01,
		},
		{
			name:               "Debt repayment - very important",
			category:           GoalCategoryDebt,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.95,
			tolerance:          0.01,
		},
		{
			name:               "Retirement - important",
			category:           GoalCategoryRetirement,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.90,
			tolerance:          0.01,
		},
		{
			name:               "Education - moderately important",
			category:           GoalCategoryEducation,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.80,
			tolerance:          0.01,
		},
		{
			name:               "Purchase (home) - moderate",
			category:           GoalCategoryPurchase,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.70,
			tolerance:          0.01,
		},
		{
			name:               "Investment - moderate-low",
			category:           GoalCategoryInvestment,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.65,
			tolerance:          0.01,
		},
		{
			name:               "Savings - low",
			category:           GoalCategorySavings,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.60,
			tolerance:          0.01,
		},
		{
			name:               "Travel - low",
			category:           GoalCategoryTravel,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.55,
			tolerance:          0.01,
		},
		{
			name:               "Other - lowest",
			category:           GoalCategoryOther,
			priority:           GoalPriorityMedium,
			expectedImportance: 0.50,
			tolerance:          0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goal := &GoalLike{
				Category: tt.category,
				Priority: tt.priority,
			}

			importance := scorer.CalculateImportance(goal)

			assert.InDelta(t, tt.expectedImportance, importance, tt.tolerance)
		})
	}

	t.Run("Priority adjustment - critical boost", func(t *testing.T) {
		goal := &GoalLike{
			Category: GoalCategorySavings, // Base: 0.60
			Priority: GoalPriorityCritical,
		}

		importance := scorer.CalculateImportance(goal)

		// 0.60 × 1.10 = 0.66
		assert.InDelta(t, 0.66, importance, 0.01, "Critical should boost by 10%")
	})

	t.Run("Priority adjustment - high boost", func(t *testing.T) {
		goal := &GoalLike{
			Category: GoalCategorySavings, // Base: 0.60
			Priority: GoalPriorityHigh,
		}

		importance := scorer.CalculateImportance(goal)

		// 0.60 × 1.05 = 0.63
		assert.InDelta(t, 0.63, importance, 0.01, "High should boost by 5%")
	})

	t.Run("Priority adjustment - low penalty", func(t *testing.T) {
		goal := &GoalLike{
			Category: GoalCategorySavings, // Base: 0.60
			Priority: GoalPriorityLow,
		}

		importance := scorer.CalculateImportance(goal)

		// 0.60 × 0.90 = 0.54
		assert.InDelta(t, 0.54, importance, 0.01, "Low should reduce by 10%")
	})

	t.Run("Cap at 1.0", func(t *testing.T) {
		goal := &GoalLike{
			Category: GoalCategoryEmergency, // Base: 1.00
			Priority: GoalPriorityCritical,  // +10%
		}

		importance := scorer.CalculateImportance(goal)

		// 1.00 × 1.10 = 1.10, but capped at 1.0
		assert.Equal(t, 1.0, importance, "Should be capped at 1.0")
	})
}

func TestAutoScorer_CalculateImpact(t *testing.T) {
	scorer := NewAutoScorer()

	monthlyIncome := 50_000_000.0      // 50M VND
	annualIncome := monthlyIncome * 12 // 600M VND

	t.Run("Low impact - target much smaller than annual income", func(t *testing.T) {
		goal := &GoalLike{
			TargetAmount: 100_000_000, // 100M
		}

		impact := scorer.CalculateImpact(goal, monthlyIncome)

		// 100M / 600M = 0.167
		assert.InDelta(t, 0.167, impact, 0.01, "Low impact")
	})

	t.Run("High impact - target equals annual income", func(t *testing.T) {
		goal := &GoalLike{
			TargetAmount: annualIncome, // 600M
		}

		impact := scorer.CalculateImpact(goal, monthlyIncome)

		// 600M / 600M = 1.0
		assert.Equal(t, 1.0, impact, "High impact")
	})

	t.Run("Very high impact - target exceeds annual income", func(t *testing.T) {
		goal := &GoalLike{
			TargetAmount: 1_200_000_000, // 1.2B
		}

		impact := scorer.CalculateImpact(goal, monthlyIncome)

		// 1200M / 600M = 2.0, capped at 1.0
		assert.Equal(t, 1.0, impact, "Capped at 1.0")
	})

	t.Run("Invalid income - neutral impact", func(t *testing.T) {
		goal := &GoalLike{
			TargetAmount: 100_000_000,
		}

		impact := scorer.CalculateImpact(goal, 0)
		assert.Equal(t, 0.5, impact, "Zero income = neutral")

		impact = scorer.CalculateImpact(goal, -1000)
		assert.Equal(t, 0.5, impact, "Negative income = neutral")
	})
}

func TestAutoScorer_CalculateAllCriteria(t *testing.T) {
	scorer := NewAutoScorer()

	targetDate := time.Now().AddDate(0, 6, 0) // 6 months
	goal := &GoalLike{
		Category:        GoalCategoryEmergency,
		Priority:        GoalPriorityHigh,
		TargetAmount:    50_000_000,
		CurrentAmount:   10_000_000,
		RemainingAmount: 40_000_000,
		TargetDate:      &targetDate,
		Status:          GoalStatusActive,
	}

	monthlyIncome := 50_000_000.0

	criteria := scorer.CalculateAllCriteria(goal, monthlyIncome)

	// Verify all 4 criteria are present
	assert.Len(t, criteria, 4, "Should have 4 criteria")
	assert.Contains(t, criteria, "urgency")
	assert.Contains(t, criteria, "feasibility")
	assert.Contains(t, criteria, "importance")
	assert.Contains(t, criteria, "impact")

	// Verify all scores are in valid range [0, 1]
	for criterion, score := range criteria {
		assert.GreaterOrEqual(t, score, 0.0, criterion+" should be >= 0")
		assert.LessOrEqual(t, score, 1.0, criterion+" should be <= 1")
	}

	// Log for manual inspection
	t.Logf("Criteria scores:")
	t.Logf("  Urgency:     %.3f", criteria["urgency"])
	t.Logf("  Feasibility: %.3f", criteria["feasibility"])
	t.Logf("  Importance:  %.3f", criteria["importance"])
	t.Logf("  Impact:      %.3f", criteria["impact"])
}
