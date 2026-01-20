package constraint

import (
	"fmt"

	budgetprofile "personalfinancedss/internal/module/cashflow/budget_profile/domain"
	debtdomain "personalfinancedss/internal/module/cashflow/debt/domain"
	goaldomain "personalfinancedss/internal/module/cashflow/goal/domain"

	"github.com/google/uuid"

	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
)

// ConstraintBuilder builds constraint model from domain entities
type ConstraintBuilder struct{}

// NewConstraintBuilder creates a new constraint builder
func NewConstraintBuilder() *ConstraintBuilder {
	return &ConstraintBuilder{}
}

// BuildConstraints builds a constraint model from domain data
func (cb *ConstraintBuilder) BuildConstraints(
	income float64,
	budgetConstraints budgetprofile.BudgetConstraints,
	goals []*goaldomain.Goal,
	debts []*debtdomain.Debt,
) (*domain.ConstraintModel, error) {

	model := &domain.ConstraintModel{
		TotalIncome:       income,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	// Process budget constraints
	for _, bc := range budgetConstraints {
		constraint := domain.CategoryConstraint{
			CategoryID: bc.CategoryID,
			Minimum:    bc.MinimumAmount,
			Maximum:    bc.MaximumAmount,
			IsFlexible: bc.IsFlexible,
			Priority:   bc.Priority,
		}

		if bc.IsFixed() {
			model.MandatoryExpenses[bc.CategoryID] = constraint
		} else {
			model.FlexibleExpenses[bc.CategoryID] = constraint
		}
	}

	// Process debts
	for _, debt := range debts {
		if debt.Status == debtdomain.DebtStatusActive && !debt.IsPaidOff() {
			// Calculate priority based on interest rate (higher interest = higher priority)
			priority := cb.calculateDebtPriority(debt)

			model.DebtPayments[debt.ID] = domain.DebtConstraint{
				DebtID:         debt.ID,
				DebtName:       debt.Name,
				MinimumPayment: debt.MinimumPayment,
				CurrentBalance: debt.CurrentBalance,
				InterestRate:   debt.InterestRate,
				Priority:       priority,
			}
		}
	}

	// Process goals
	for _, goal := range goals {
		if goal.Status == goaldomain.GoalStatusActive && !goal.IsCompleted() {
			// Calculate suggested contribution if not set
			suggestedContribution := cb.calculateGoalContribution(goal)

			model.GoalTargets[goal.ID] = domain.GoalConstraint{
				GoalID:                goal.ID,
				GoalName:              goal.Name,
				GoalType:              string(goal.Category),
				SuggestedContribution: suggestedContribution,
				Priority:              string(goal.Priority),
				PriorityWeight:        cb.goalPriorityToWeight(goal.Priority),
				RemainingAmount:       goal.RemainingAmount,
			}
		}
	}

	return model, nil
}

// CheckFeasibility checks if the allocation is feasible (income >= mandatory expenses + debt minimums)
func (cb *ConstraintBuilder) CheckFeasibility(model *domain.ConstraintModel) (bool, float64) {
	var totalMandatory float64

	// Sum mandatory expenses
	for _, cat := range model.MandatoryExpenses {
		totalMandatory += cat.Minimum
	}

	// Sum minimum debt payments
	for _, debt := range model.DebtPayments {
		totalMandatory += debt.MinimumPayment
	}

	deficit := totalMandatory - model.TotalIncome
	isFeasible := model.TotalIncome >= totalMandatory

	return isFeasible, deficit
}

// CalculateSurplus calculates available income after mandatory expenses and minimum debt payments
func (cb *ConstraintBuilder) CalculateSurplus(model *domain.ConstraintModel) float64 {
	var totalMandatory float64

	// Sum mandatory expenses
	for _, cat := range model.MandatoryExpenses {
		totalMandatory += cat.Minimum
	}

	// Sum minimum debt payments
	for _, debt := range model.DebtPayments {
		totalMandatory += debt.MinimumPayment
	}

	surplus := model.TotalIncome - totalMandatory
	if surplus < 0 {
		surplus = 0
	}

	return surplus
}

// calculateDebtPriority calculates debt priority based on interest rate
// Higher interest rate = higher priority (pay off first to save on interest)
func (cb *ConstraintBuilder) calculateDebtPriority(debt *debtdomain.Debt) int {
	// Priority ranges from 1 (highest) to 99 (lowest)
	// Based on interest rate:
	// > 20%: priority 1 (critical - credit cards, payday loans)
	// 10-20%: priority 10
	// 5-10%: priority 20
	// < 5%: priority 30

	switch {
	case debt.InterestRate >= 20:
		return 1
	case debt.InterestRate >= 10:
		return 10
	case debt.InterestRate >= 5:
		return 20
	default:
		return 30
	}
}

// calculateGoalContribution calculates suggested monthly contribution for a goal
func (cb *ConstraintBuilder) calculateGoalContribution(goal *goaldomain.Goal) float64 {
	// If user has set a suggested contribution, use that
	if goal.SuggestedContribution != nil && *goal.SuggestedContribution > 0 {
		return *goal.SuggestedContribution
	}

	// If goal has auto-contribute amount, use that
	if goal.AutoContributeAmount != nil && *goal.AutoContributeAmount > 0 {
		return *goal.AutoContributeAmount
	}

	// Calculate based on remaining amount and time remaining
	remaining := goal.RemainingAmount
	if remaining <= 0 {
		return 0
	}

	// If there's a target date, calculate monthly contribution needed
	if goal.TargetDate != nil {
		daysRemaining := goal.DaysRemaining()
		if daysRemaining > 0 {
			monthsRemaining := float64(daysRemaining) / 30.0
			if monthsRemaining > 0 {
				return remaining / monthsRemaining
			}
		}
	}

	// Default: aim to complete in 12 months
	return remaining / 12.0
}

// goalPriorityToWeight converts goal priority string to numerical weight
func (cb *ConstraintBuilder) goalPriorityToWeight(priority goaldomain.GoalPriority) int {
	switch priority {
	case goaldomain.GoalPriorityCritical:
		return 1
	case goaldomain.GoalPriorityHigh:
		return 10
	case goaldomain.GoalPriorityMedium:
		return 20
	case goaldomain.GoalPriorityLow:
		return 30
	default:
		return 99
	}
}

// GetSuggestionsForDeficit generates suggestions when income is insufficient
func (cb *ConstraintBuilder) GetSuggestionsForDeficit(model *domain.ConstraintModel, deficit float64) []string {
	suggestions := []string{
		fmt.Sprintf("Your income is %.2f short of covering mandatory expenses and minimum debt payments", deficit),
	}

	// Suggest reducing flexible expenses if any exist
	if len(model.FlexibleExpenses) > 0 {
		suggestions = append(suggestions, "Consider adjusting these flexible expense categories:")
		for _, cat := range model.FlexibleExpenses {
			if cat.Maximum > cat.Minimum {
				savings := cat.Maximum - cat.Minimum
				suggestions = append(suggestions, fmt.Sprintf("- Reduce flexible spending by up to %.2f", savings))
			}
		}
	}

	// Suggest debt restructuring for high-interest debts
	hasHighInterestDebt := false
	for _, debt := range model.DebtPayments {
		if debt.InterestRate >= 15 {
			hasHighInterestDebt = true
			break
		}
	}

	if hasHighInterestDebt {
		suggestions = append(suggestions, "Consider debt consolidation or refinancing for high-interest debts")
	}

	// Suggest increasing income
	suggestions = append(suggestions, "Explore opportunities to increase income through:")
	suggestions = append(suggestions, "- Side gigs or freelance work")
	suggestions = append(suggestions, "- Selling unused items")
	suggestions = append(suggestions, "- Negotiating a raise")

	return suggestions
}
