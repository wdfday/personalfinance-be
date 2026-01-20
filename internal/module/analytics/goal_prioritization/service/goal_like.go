package service

import (
	"time"

	goal_domain "personalfinancedss/internal/module/cashflow/goal/domain"
)

// GoalLike is a simplified interface for goals compatible with AutoScorer
// This allows auto-scoring without depending on full Goal entity
type GoalLike struct {
	Category        GoalCategory
	Priority        GoalPriority
	TargetAmount    float64
	CurrentAmount   float64
	TargetDate      *time.Time
	RemainingAmount float64
	Status          GoalStatus
}

// Type aliases for compatibility
type (
	GoalCategory = goal_domain.GoalCategory
	GoalPriority = goal_domain.GoalPriority
	GoalStatus   = goal_domain.GoalStatus
)

// Const aliases - using renamed GoalCategory values
const (
	GoalCategorySavings    = goal_domain.GoalCategorySavings
	GoalCategoryDebt       = goal_domain.GoalCategoryDebt
	GoalCategoryInvestment = goal_domain.GoalCategoryInvestment
	GoalCategoryPurchase   = goal_domain.GoalCategoryPurchase
	GoalCategoryEmergency  = goal_domain.GoalCategoryEmergency
	GoalCategoryRetirement = goal_domain.GoalCategoryRetirement
	GoalCategoryEducation  = goal_domain.GoalCategoryEducation
	GoalCategoryTravel     = goal_domain.GoalCategoryTravel
	GoalCategoryOther      = goal_domain.GoalCategoryOther

	GoalPriorityLow      = goal_domain.GoalPriorityLow
	GoalPriorityMedium   = goal_domain.GoalPriorityMedium
	GoalPriorityHigh     = goal_domain.GoalPriorityHigh
	GoalPriorityCritical = goal_domain.GoalPriorityCritical

	GoalStatusActive    = goal_domain.GoalStatusActive
	GoalStatusCompleted = goal_domain.GoalStatusCompleted
)

// IsCompleted checks if goal is completed
func (g *GoalLike) IsCompleted() bool {
	return g.CurrentAmount >= g.TargetAmount || g.Status == GoalStatusCompleted
}

// DaysRemaining calculates days until deadline
func (g *GoalLike) DaysRemaining() int {
	if g.TargetDate == nil {
		return 0
	}
	duration := time.Until(*g.TargetDate)
	return int(duration.Hours() / 24)
}

// UpdateCalculatedFields updates remaining amount (simplified)
func (g *GoalLike) UpdateCalculatedFields() {
	g.RemainingAmount = g.TargetAmount - g.CurrentAmount
	if g.RemainingAmount < 0 {
		g.RemainingAmount = 0
	}
}
