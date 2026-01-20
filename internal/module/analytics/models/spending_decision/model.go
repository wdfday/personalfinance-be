package spending_decision

import (
	"context"
	"errors"

	"personalfinancedss/internal/module/analytics/spending_decision/domain"
	"personalfinancedss/internal/module/analytics/spending_decision/dto"
)

// LargePurchaseModel implements large purchase analysis with budget impact
type LargePurchaseModel struct {
	name        string
	description string
}

// NewSpendingDecisionModel creates new model instance
// Note: keeping constructor name for backward compatibility
func NewSpendingDecisionModel() *LargePurchaseModel {
	return &LargePurchaseModel{
		name:        "large_purchase_analysis",
		description: "Large purchase analysis with budget allocation what-if integration",
	}
}

func (m *LargePurchaseModel) Name() string {
	return m.name
}

func (m *LargePurchaseModel) Description() string {
	return m.description
}

func (m *LargePurchaseModel) Dependencies() []string {
	return []string{"budget_allocation"} // Calls budget allocation for what-if analysis
}

func (m *LargePurchaseModel) Validate(ctx context.Context, input interface{}) error {
	lpInput, ok := input.(*dto.LargePurchaseInput)
	if !ok {
		return errors.New("input must be *dto.LargePurchaseInput")
	}

	if lpInput.PurchaseAmount <= 0 {
		return errors.New("purchase amount must be positive")
	}
	if lpInput.ItemType == "" {
		return errors.New("item type is required")
	}

	return nil
}

func (m *LargePurchaseModel) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	lpInput := input.(*dto.LargePurchaseInput)
	return m.executeLargePurchase(ctx, lpInput)
}

// executeLargePurchase handles large purchase analysis
func (m *LargePurchaseModel) executeLargePurchase(ctx context.Context, input *dto.LargePurchaseInput) (interface{}, error) {
	output := &dto.LargePurchaseOutput{
		FundingOptions:   make([]domain.FundingOption, 0),
		Alternatives:     make([]domain.Alternative, 0),
		BehavioralNudges: make([]string, 0),
	}

	// Step 1: Analyze funding options
	output.FundingOptions = m.analyzeFundingOptions(input)

	// Step 2: Calculate true cost
	output.TrueCost = m.calculateTrueCost(input)

	// Step 3: Budget reallocation plan (if requested)
	if input.PreferredFundingSource == "budget_realloc" {
		output.ReallocationPlan = m.generateReallocationPlan(input)
	}

	// Step 4: Long-term impact
	output.LongTermImpact = m.projectLongTermImpact(input)

	// Step 5: Alternatives
	output.Alternatives = m.generateLargePurchaseAlternatives(input)

	// Step 6: Recommendation
	output.Recommendation = m.makeLargePurchaseRecommendation(input, output)

	// Step 7: Behavioral nudges
	output.BehavioralNudges = m.generateBehavioralNudges(input, output)

	return output, nil
}
