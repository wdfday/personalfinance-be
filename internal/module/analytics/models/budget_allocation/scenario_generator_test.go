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
	scenarios, err := generator.GenerateScenarios(model, categoryNames)

	// Assertions
	if err != nil {
		t.Fatalf("GenerateScenarios failed: %v", err)
	}

	if len(scenarios) != 3 {
		t.Fatalf("Expected 3 scenarios, got %d", len(scenarios))
	}

	// Check scenario types
	expectedTypes := map[domain.ScenarioType]bool{
		domain.ScenarioConservative: false,
		domain.ScenarioBalanced:     false,
		domain.ScenarioAggressive:   false,
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

func TestScenarioGenerator_ConservativeVsAggressive(t *testing.T) {
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
	conservative := generator.GenerateConservativeScenario(model, categoryNames)
	aggressive := generator.GenerateAggressiveScenario(model, categoryNames)

	// Conservative should allocate less to goals than aggressive
	var conservativeGoalTotal float64
	var aggressiveGoalTotal float64

	for _, alloc := range conservative.GoalAllocations {
		conservativeGoalTotal += alloc.Amount
	}

	for _, alloc := range aggressive.GoalAllocations {
		aggressiveGoalTotal += alloc.Amount
	}

	if conservativeGoalTotal >= aggressiveGoalTotal {
		t.Errorf("Conservative goal allocation (%f) should be less than aggressive (%f)",
			conservativeGoalTotal, aggressiveGoalTotal)
	}

	// Aggressive should have higher savings rate
	if aggressive.Summary.SavingsRate <= conservative.Summary.SavingsRate {
		t.Errorf("Aggressive savings rate (%f%%) should be higher than conservative (%f%%)",
			aggressive.Summary.SavingsRate, conservative.Summary.SavingsRate)
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

	scenarios, err := generator.GenerateScenarios(model, categoryNames)

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

	// Generate conservative scenario (should prioritize emergency fund)
	conservative := generator.GenerateConservativeScenario(model, categoryNames)

	// Check if emergency fund got allocation
	var emergencyAllocation float64
	var savingsAllocation float64

	for _, alloc := range conservative.GoalAllocations {
		if alloc.GoalID == emergencyID {
			emergencyAllocation = alloc.Amount
		}
		if alloc.GoalID == savingsID {
			savingsAllocation = alloc.Amount
		}
	}

	// Emergency fund should get allocation in conservative scenario
	if emergencyAllocation == 0 {
		t.Error("Emergency fund should receive allocation in conservative scenario")
	}

	// With limited budget ($4000 - $3000 mandatory = $1000 surplus),
	// emergency fund (priority 3) should get more than vacation (priority 6)
	if emergencyAllocation < savingsAllocation {
		t.Errorf("Emergency fund allocation (%f) should be >= vacation (%f) due to higher priority",
			emergencyAllocation, savingsAllocation)
	}
}

func TestScenarioGenerator_WarningGeneration(t *testing.T) {
	generator := NewScenarioGenerator()

	model := &domain.ConstraintModel{
		TotalIncome:       10000.0,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	// Add mandatory expense
	model.MandatoryExpenses[uuid.New()] = domain.CategoryConstraint{
		Minimum: 8000.0, // Very high, leaving little surplus
	}

	// Add high-interest debt
	debtID := uuid.New()
	model.DebtPayments[debtID] = domain.DebtConstraint{
		DebtID:         debtID,
		DebtName:       "Credit Card",
		MinimumPayment: 200.0,
		InterestRate:   20.0, // High interest
	}

	categoryNames := map[uuid.UUID]string{}

	// Generate balanced scenario
	balanced := generator.GenerateBalancedScenario(model, categoryNames)

	// Should have warning about high-interest debt
	hasDebtWarning := false
	for _, warning := range balanced.Warnings {
		if warning.Category == "debt" && warning.Severity == domain.SeverityWarning {
			hasDebtWarning = true
			break
		}
	}

	if !hasDebtWarning {
		t.Error("Should have warning about high-interest debt")
	}

	// Generate aggressive scenario
	aggressive := generator.GenerateAggressiveScenario(model, categoryNames)

	// Aggressive scenario with low surplus should have warning
	hasLowSurplusWarning := false
	for _, warning := range aggressive.Warnings {
		if warning.Category == "income" || warning.Severity == domain.SeverityWarning {
			hasLowSurplusWarning = true
			break
		}
	}

	if !hasLowSurplusWarning && aggressive.Summary.Surplus < 100 {
		t.Error("Aggressive scenario with very low surplus should have warning")
	}
}

func TestScenarioGenerator_GenerateScenariosWithComparison(t *testing.T) {
	generator := NewScenarioGenerator()

	model := &domain.ConstraintModel{
		TotalIncome:       10000.0,
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	categoryID := uuid.New()
	model.MandatoryExpenses[categoryID] = domain.CategoryConstraint{
		CategoryID: categoryID,
		Minimum:    3000.0,
	}

	flexID := uuid.New()
	model.FlexibleExpenses[flexID] = domain.CategoryConstraint{
		CategoryID: flexID,
		Minimum:    500.0,
		Maximum:    1500.0,
		IsFlexible: true,
	}

	goalID := uuid.New()
	model.GoalTargets[goalID] = domain.GoalConstraint{
		GoalID:                goalID,
		GoalName:              "Emergency Fund",
		GoalType:              "emergency",
		SuggestedContribution: 500.0,
		Priority:              "high",
		PriorityWeight:        5,
		RemainingAmount:       10000.0,
	}

	categoryNames := map[uuid.UUID]string{
		categoryID: "Housing",
		flexID:     "Food",
	}

	results, err := generator.GenerateScenariosWithComparison(model, categoryNames)

	if err != nil {
		t.Fatalf("GenerateScenariosWithComparison failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 dual results, got %d", len(results))
	}

	expectedTypes := []domain.ScenarioType{
		domain.ScenarioConservative,
		domain.ScenarioBalanced,
		domain.ScenarioAggressive,
	}

	for i, result := range results {
		if result.ScenarioType != expectedTypes[i] {
			t.Errorf("Expected scenario type %s, got %s", expectedTypes[i], result.ScenarioType)
		}

		// Both scenarios should have allocations
		if len(result.PreemptiveScenario.CategoryAllocations) == 0 {
			t.Errorf("Preemptive scenario %s has no category allocations", result.ScenarioType)
		}
		if len(result.WeightedScenario.CategoryAllocations) == 0 {
			t.Errorf("Weighted scenario %s has no category allocations", result.ScenarioType)
		}

		// Comparison should have recommendation
		if result.Comparison.RecommendedSolver == "" {
			t.Errorf("Scenario %s comparison has no recommended solver", result.ScenarioType)
		}

		t.Logf("Scenario %s:", result.ScenarioType)
		t.Logf("  Preemptive feasibility: %.2f%%", result.PreemptiveScenario.FeasibilityScore)
		t.Logf("  Weighted feasibility: %.2f%%", result.WeightedScenario.FeasibilityScore)
		t.Logf("  Recommended: %s (%s)", result.Comparison.RecommendedSolver, result.Comparison.Reason)
	}
}

func TestScenarioGenerator_GenerateScenariosWithTripleComparison(t *testing.T) {
	generator := NewScenarioGenerator()

	model := &domain.ConstraintModel{
		TotalIncome:       5000.0, // Limited budget to see differences
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	categoryID := uuid.New()
	model.MandatoryExpenses[categoryID] = domain.CategoryConstraint{
		CategoryID: categoryID,
		Minimum:    2000.0,
	}

	// Add multiple goals to see different solver behaviors
	goalID1 := uuid.New()
	goalID2 := uuid.New()
	goalID3 := uuid.New()

	model.GoalTargets[goalID1] = domain.GoalConstraint{
		GoalID:                goalID1,
		GoalName:              "Emergency Fund",
		GoalType:              "emergency",
		SuggestedContribution: 1000.0,
		Priority:              "high",
		PriorityWeight:        5,
		RemainingAmount:       10000.0,
	}

	model.GoalTargets[goalID2] = domain.GoalConstraint{
		GoalID:                goalID2,
		GoalName:              "Vacation",
		GoalType:              "purchase",
		SuggestedContribution: 1000.0,
		Priority:              "medium",
		PriorityWeight:        15,
		RemainingAmount:       5000.0,
	}

	model.GoalTargets[goalID3] = domain.GoalConstraint{
		GoalID:                goalID3,
		GoalName:              "New Car",
		GoalType:              "purchase",
		SuggestedContribution: 1000.0,
		Priority:              "low",
		PriorityWeight:        30,
		RemainingAmount:       20000.0,
	}

	categoryNames := map[uuid.UUID]string{
		categoryID: "Housing",
	}

	results, err := generator.GenerateScenariosWithTripleComparison(model, categoryNames)

	if err != nil {
		t.Fatalf("GenerateScenariosWithTripleComparison failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 triple results, got %d", len(results))
	}

	expectedTypes := []domain.ScenarioType{
		domain.ScenarioConservative,
		domain.ScenarioBalanced,
		domain.ScenarioAggressive,
	}

	for i, result := range results {
		if result.ScenarioType != expectedTypes[i] {
			t.Errorf("Expected scenario type %s, got %s", expectedTypes[i], result.ScenarioType)
		}

		// All three scenarios should have allocations
		if len(result.PreemptiveScenario.GoalAllocations) == 0 {
			t.Errorf("Preemptive scenario %s has no goal allocations", result.ScenarioType)
		}
		if len(result.WeightedScenario.GoalAllocations) == 0 {
			t.Errorf("Weighted scenario %s has no goal allocations", result.ScenarioType)
		}
		if len(result.MinmaxScenario.GoalAllocations) == 0 {
			t.Errorf("Minmax scenario %s has no goal allocations", result.ScenarioType)
		}

		// Comparison should have recommendation
		if result.Comparison.RecommendedSolver == "" {
			t.Errorf("Scenario %s comparison has no recommended solver", result.ScenarioType)
		}

		t.Logf("Scenario %s:", result.ScenarioType)
		t.Logf("  Preemptive: feasibility=%.1f%%, achieved=%d",
			result.PreemptiveScenario.FeasibilityScore,
			result.Comparison.PreemptiveAchievedCount)
		t.Logf("  Weighted: feasibility=%.1f%%, achieved=%d",
			result.WeightedScenario.FeasibilityScore,
			result.Comparison.WeightedAchievedCount)
		t.Logf("  Minmax: feasibility=%.1f%%, achieved=%d, minAch=%.1f%%, balanced=%v",
			result.MinmaxScenario.FeasibilityScore,
			result.Comparison.MinmaxAchievedCount,
			result.Comparison.MinmaxMinAchievement,
			result.Comparison.MinmaxIsBalanced)
		t.Logf("  Recommended: %s (%s)", result.Comparison.RecommendedSolver, result.Comparison.Reason)
	}
}
