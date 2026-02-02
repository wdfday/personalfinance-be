package budget_allocation

import (
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"testing"

	"github.com/google/uuid"
)

func TestScenarioGenerator_GenerateScenarios(t *testing.T) {
	generator := NewScenarioGenerator()

	// Create test constraint model
	model := &domain.ConstraintModel{
		TotalIncome:       10000.0,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	categoryID1 := uuid.New()
	categoryID2 := uuid.New()

	// Add mandatory expense
	model.MandatoryExpenses[categoryID1] = domain.CategoryConstraint{
		CategoryID: categoryID1,
		Minimum:    3000.0,
		IsFlexible: false,
		Priority:   1,
	}

	// Add flexible expense
	model.FlexibleExpenses[categoryID2] = domain.CategoryConstraint{
		CategoryID: categoryID2,
		Minimum:    1000.0,
		Maximum:    2000.0,
		IsFlexible: true,
		Priority:   10,
	}

	// Add debt
	debtID := uuid.New()
	model.DebtPayments[debtID] = domain.DebtConstraint{
		DebtID:         debtID,
		DebtName:       "Credit Card",
		MinimumPayment: 200.0,
		CurrentBalance: 5000.0,
		InterestRate:   18.0,
		Priority:       1,
	}

	// Add goal
	goalID := uuid.New()
	model.GoalTargets[goalID] = domain.GoalConstraint{
		GoalID:                goalID,
		GoalName:              "Emergency Fund",
		GoalType:              "emergency",
		SuggestedContribution: 500.0,
		Priority:              "high",
		PriorityWeight:        10,
		RemainingAmount:       5000.0,
	}

	// Category names
	categoryNames := map[uuid.UUID]string{
		categoryID1: "Housing",
		categoryID2: "Food",
	}

	// Generate scenarios
	scenarios, err := generator.GenerateScenarios(model, categoryNames, nil)

	// Assertions
	if err != nil {
		t.Fatalf("GenerateScenarios failed: %v", err)
	}

	if len(scenarios) != 3 {
		t.Fatalf("Expected 2 scenarios, got %d", len(scenarios))
	}

	// Check scenario types
	expectedTypes := map[domain.ScenarioType]bool{
		domain.ScenarioSafe:     false,
		domain.ScenarioBalanced: false,
	}

	for _, scenario := range scenarios {
		if _, exists := expectedTypes[scenario.ScenarioType]; !exists {
			t.Errorf("Unexpected scenario type: %s", scenario.ScenarioType)
		}
		expectedTypes[scenario.ScenarioType] = true

		// Check that all scenarios have allocations
		if len(scenario.CategoryAllocations) == 0 {
			t.Errorf("Scenario %s has no category allocations", scenario.ScenarioType)
		}

		// Check that summary is calculated
		if scenario.Summary.TotalIncome != model.TotalIncome {
			t.Errorf("Scenario %s: expected income %f, got %f",
				scenario.ScenarioType, model.TotalIncome, scenario.Summary.TotalIncome)
		}
	}

	// Verify all types were generated
	for scenarioType, found := range expectedTypes {
		if !found {
			t.Errorf("Scenario type %s was not generated", scenarioType)
		}
	}
}

func TestScenarioGenerator_SafeVsBalanced(t *testing.T) {
	generator := NewScenarioGenerator()

	model := &domain.ConstraintModel{
		TotalIncome:       10000.0,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	// Add mandatory expense
	categoryID := uuid.New()
	model.MandatoryExpenses[categoryID] = domain.CategoryConstraint{
		CategoryID: categoryID,
		Minimum:    3000.0,
	}

	// Add flexible expense
	flexID := uuid.New()
	model.FlexibleExpenses[flexID] = domain.CategoryConstraint{
		CategoryID: flexID,
		Minimum:    1000.0,
		Maximum:    3000.0,
		IsFlexible: true,
	}

	// Add goal
	goalID := uuid.New()
	model.GoalTargets[goalID] = domain.GoalConstraint{
		GoalID:                goalID,
		GoalName:              "Savings",
		SuggestedContribution: 1000.0,
		Priority:              "high",
		PriorityWeight:        10,
		RemainingAmount:       10000.0,
	}

	categoryNames := map[uuid.UUID]string{
		categoryID: "Housing",
		flexID:     "Food",
	}

	// Generate scenarios
	safe := generator.GenerateSafeScenario(model, categoryNames)
	balanced := generator.GenerateBalancedScenario(model, categoryNames)

	// Safe should allocate less to goals than balanced
	var safeGoalTotal float64
	var balancedGoalTotal float64

	for _, alloc := range safe.GoalAllocations {
		safeGoalTotal += alloc.Amount
	}

	for _, alloc := range balanced.GoalAllocations {
		balancedGoalTotal += alloc.Amount
	}

	if safeGoalTotal >= balancedGoalTotal {
		t.Errorf("Safe goal allocation (%f) should be less than balanced (%f)",
			safeGoalTotal, balancedGoalTotal)
	}

	// Balanced should have higher savings rate
	if balanced.Summary.SavingsRate <= safe.Summary.SavingsRate {
		t.Errorf("Balanced savings rate (%f%%) should be higher than safe (%f%%)",
			balanced.Summary.SavingsRate, safe.Summary.SavingsRate)
	}
}

func TestScenarioGenerator_InsufficientIncome(t *testing.T) {
	generator := NewScenarioGenerator()

	model := &domain.ConstraintModel{
		TotalIncome:       3000.0, // Low income
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
	}

	// Add mandatory expense that exceeds income
	categoryID := uuid.New()
	model.MandatoryExpenses[categoryID] = domain.CategoryConstraint{
		CategoryID: categoryID,
		Minimum:    2500.0,
	}

	// Add debt payment
	debtID := uuid.New()
	model.DebtPayments[debtID] = domain.DebtConstraint{
		DebtID:         debtID,
		DebtName:       "Loan",
		MinimumPayment: 600.0, // Total mandatory = 3100 > 3000
	}

	categoryNames := map[uuid.UUID]string{
		categoryID: "Housing",
	}

	scenarios, err := generator.GenerateScenarios(model, categoryNames, nil)

	if err != nil {
		t.Fatalf("GenerateScenarios failed: %v", err)
	}

	// All scenarios should have critical warnings
	for _, scenario := range scenarios {
		hasCriticalWarning := false
		for _, warning := range scenario.Warnings {
			if warning.Severity == domain.SeverityCritical {
				hasCriticalWarning = true
				break
			}
		}

		if !hasCriticalWarning {
			t.Errorf("Scenario %s should have critical warning for insufficient income", scenario.ScenarioType)
		}

		// Feasibility score should be 0
		if scenario.FeasibilityScore != 0 {
			t.Errorf("Scenario %s: expected feasibility score 0, got %f",
				scenario.ScenarioType, scenario.FeasibilityScore)
		}
	}
}

func TestScenarioGenerator_EmergencyFundPriority(t *testing.T) {
	generator := NewScenarioGenerator()

	// Use limited budget to test priority ordering
	model := &domain.ConstraintModel{
		TotalIncome:       4000.0, // Limited budget to force priority choices
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	// Add mandatory expense
	model.MandatoryExpenses[uuid.New()] = domain.CategoryConstraint{
		Minimum: 3000.0,
	}

	// Add emergency fund goal (priority 3 in preemptive GP)
	emergencyID := uuid.New()
	model.GoalTargets[emergencyID] = domain.GoalConstraint{
		GoalID:                emergencyID,
		GoalName:              "Emergency Fund",
		GoalType:              "emergency",
		SuggestedContribution: 800.0, // Higher target
		Priority:              "critical",
		PriorityWeight:        1,
		RemainingAmount:       5000.0,
	}

	// Add other goal (priority 6 in preemptive GP - lower priority)
	savingsID := uuid.New()
	model.GoalTargets[savingsID] = domain.GoalConstraint{
		GoalID:                savingsID,
		GoalName:              "Vacation",
		GoalType:              "purchase",
		SuggestedContribution: 800.0, // Same target
		Priority:              "low",
		PriorityWeight:        30, // > 10, so goes to priority 6
		RemainingAmount:       3000.0,
	}

	categoryNames := map[uuid.UUID]string{}

	// Generate safe scenario (should prioritize emergency fund)
	safe := generator.GenerateSafeScenario(model, categoryNames)

	// Check if emergency fund got allocation
	var emergencyAllocation float64
	var savingsAllocation float64

	for _, alloc := range safe.GoalAllocations {
		if alloc.GoalID == emergencyID {
			emergencyAllocation = alloc.Amount
		}
		if alloc.GoalID == savingsID {
			savingsAllocation = alloc.Amount
		}
	}

	// Emergency fund should get allocation in safe scenario
	if emergencyAllocation == 0 {
		t.Error("Emergency fund should receive allocation in safe scenario")
	}

	// With limited budget ($4000 - $3000 mandatory = $1000 surplus),
	// emergency fund (priority 3) should get more than vacation (priority 6)
	if emergencyAllocation < savingsAllocation {
		t.Errorf("Emergency fund allocation (%f) should be >= vacation (%f) due to higher priority",
			emergencyAllocation, savingsAllocation)
	}
}
