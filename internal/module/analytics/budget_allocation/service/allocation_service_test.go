package service

import (
	"testing"

	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"personalfinancedss/internal/module/analytics/budget_allocation/dto"
)

// Note: Full integration tests would require mocking all repositories
// This file contains basic validation tests for the service structure

func TestGenerateAllocationRequest_Validation(t *testing.T) {
	// This test validates the request DTO structure
	tests := []struct {
		name    string
		request dto.GenerateAllocationRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: dto.GenerateAllocationRequest{
				Year:  2024,
				Month: 12,
			},
			wantErr: false,
		},
		{
			name: "Invalid year (too low)",
			request: dto.GenerateAllocationRequest{
				Year:  1999,
				Month: 12,
			},
			wantErr: true, // Would fail validation binding
		},
		{
			name: "Invalid month (too high)",
			request: dto.GenerateAllocationRequest{
				Year:  2024,
				Month: 13,
			},
			wantErr: true, // Would fail validation binding
		},
		{
			name: "Invalid month (too low)",
			request: dto.GenerateAllocationRequest{
				Year:  2024,
				Month: 0,
			},
			wantErr: true, // Would fail validation binding
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate year
			if tt.request.Year < 2000 || tt.request.Year > 2100 {
				if !tt.wantErr {
					t.Errorf("Expected valid request, but year %d is invalid", tt.request.Year)
				}
				return
			}

			// Validate month
			if tt.request.Month < 1 || tt.request.Month > 12 {
				if !tt.wantErr {
					t.Errorf("Expected valid request, but month %d is invalid", tt.request.Month)
				}
				return
			}

			if tt.wantErr {
				t.Error("Expected validation error but got none")
			}
		})
	}
}

func TestExecuteAllocationRequest_Validation(t *testing.T) {
	tests := []struct {
		name         string
		scenarioType domain.ScenarioType
		valid        bool
	}{
		{
			name:         "Conservative scenario",
			scenarioType: domain.ScenarioConservative,
			valid:        true,
		},
		{
			name:         "Balanced scenario",
			scenarioType: domain.ScenarioBalanced,
			valid:        true,
		},
		{
			name:         "Aggressive scenario",
			scenarioType: domain.ScenarioAggressive,
			valid:        true,
		},
		{
			name:         "Invalid scenario type",
			scenarioType: "invalid",
			valid:        false,
		},
	}

	validTypes := map[domain.ScenarioType]bool{
		domain.ScenarioConservative: true,
		domain.ScenarioBalanced:     true,
		domain.ScenarioAggressive:   true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validTypes[tt.scenarioType]

			if isValid != tt.valid {
				t.Errorf("Scenario type %s: expected valid=%v, got %v", tt.scenarioType, tt.valid, isValid)
			}
		})
	}
}

func TestAllocationScenario_FeasibilityScore(t *testing.T) {
	scenario := domain.NewAllocationScenario(domain.ScenarioBalanced)

	// Default feasibility score should be 100
	if scenario.FeasibilityScore != 100.0 {
		t.Errorf("Expected default feasibility score 100, got %f", scenario.FeasibilityScore)
	}

	// Test IsFeasible
	if !scenario.IsFeasible() {
		t.Error("Scenario with score 100 should be feasible")
	}

	// Set low feasibility score
	scenario.FeasibilityScore = 40.0
	if scenario.IsFeasible() {
		t.Error("Scenario with score < 50 should not be feasible")
	}

	// Edge case: exactly 50
	scenario.FeasibilityScore = 50.0
	if !scenario.IsFeasible() {
		t.Error("Scenario with score = 50 should be feasible")
	}
}

func TestAllocationScenario_CalculateSummary(t *testing.T) {
	scenario := domain.NewAllocationScenario(domain.ScenarioBalanced)

	// Add category allocations
	scenario.CategoryAllocations = []domain.CategoryAllocation{
		{Amount: 3000.0, IsFlexible: false}, // Mandatory
		{Amount: 1500.0, IsFlexible: true},  // Flexible
	}

	// Add goal allocations
	scenario.GoalAllocations = []domain.GoalAllocation{
		{Amount: 500.0},
		{Amount: 300.0},
	}

	// Add debt allocations
	scenario.DebtAllocations = []domain.DebtAllocation{
		{Amount: 200.0, MinimumPayment: 150.0, ExtraPayment: 50.0},
	}

	totalIncome := 10000.0
	scenario.CalculateSummary(totalIncome)

	// Check summary calculations
	expectedMandatory := 3000.0
	if scenario.Summary.MandatoryExpenses != expectedMandatory {
		t.Errorf("Expected mandatory %f, got %f", expectedMandatory, scenario.Summary.MandatoryExpenses)
	}

	expectedFlexible := 1500.0
	if scenario.Summary.FlexibleExpenses != expectedFlexible {
		t.Errorf("Expected flexible %f, got %f", expectedFlexible, scenario.Summary.FlexibleExpenses)
	}

	expectedGoalTotal := 800.0
	if scenario.Summary.TotalGoalContributions != expectedGoalTotal {
		t.Errorf("Expected goal total %f, got %f", expectedGoalTotal, scenario.Summary.TotalGoalContributions)
	}

	expectedDebtTotal := 200.0
	if scenario.Summary.TotalDebtPayments != expectedDebtTotal {
		t.Errorf("Expected debt total %f, got %f", expectedDebtTotal, scenario.Summary.TotalDebtPayments)
	}

	// Check total allocated
	expectedTotalAllocated := 3000.0 + 1500.0 + 800.0 + 200.0 // 5500.0
	if scenario.Summary.TotalAllocated != expectedTotalAllocated {
		t.Errorf("Expected total allocated %f, got %f", expectedTotalAllocated, scenario.Summary.TotalAllocated)
	}

	// Check surplus
	expectedSurplus := totalIncome - expectedTotalAllocated // 4500.0
	if scenario.Summary.Surplus != expectedSurplus {
		t.Errorf("Expected surplus %f, got %f", expectedSurplus, scenario.Summary.Surplus)
	}

	// Check savings rate: (goals + extra debt) / income * 100
	// = (800 + 50) / 10000 * 100 = 8.5%
	expectedSavingsRate := 8.5
	if scenario.Summary.SavingsRate != expectedSavingsRate {
		t.Errorf("Expected savings rate %f%%, got %f%%", expectedSavingsRate, scenario.Summary.SavingsRate)
	}
}

func TestAllocationScenario_AddWarning(t *testing.T) {
	scenario := domain.NewAllocationScenario(domain.ScenarioBalanced)

	// Initially no warnings
	if len(scenario.Warnings) != 0 {
		t.Errorf("Expected 0 warnings initially, got %d", len(scenario.Warnings))
	}

	// Add warning
	scenario.AddWarning(
		domain.SeverityCritical,
		"income",
		"Insufficient income",
		"Increase income",
		"Reduce expenses",
	)

	if len(scenario.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(scenario.Warnings))
	}

	warning := scenario.Warnings[0]

	if warning.Severity != domain.SeverityCritical {
		t.Errorf("Expected severity %s, got %s", domain.SeverityCritical, warning.Severity)
	}

	if warning.Category != "income" {
		t.Errorf("Expected category 'income', got '%s'", warning.Category)
	}

	if len(warning.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(warning.Suggestions))
	}

	// Add another warning
	scenario.AddWarning(domain.SeverityWarning, "debt", "High interest debt")

	if len(scenario.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(scenario.Warnings))
	}
}

func TestActionItem_Structure(t *testing.T) {
	// Test ActionItem DTO structure
	action := dto.ActionItem{
		Type:        "goal_contribution",
		Description: "Set auto-contribute for Emergency Fund",
		Amount:      500.0,
		Status:      "completed",
	}

	if action.Type != "goal_contribution" {
		t.Error("Action type mismatch")
	}

	if action.Amount != 500.0 {
		t.Error("Action amount mismatch")
	}

	if action.Status != "completed" {
		t.Error("Action status mismatch")
	}

	// Test with error
	failedAction := dto.ActionItem{
		Type:   "goal_contribution",
		Status: "failed",
		Error:  "Repository error",
	}

	if failedAction.Error != "Repository error" {
		t.Error("Action error message mismatch")
	}
}

// Integration test would look like this (requires mocks):
/*
func TestAllocationService_GenerateAllocations_Integration(t *testing.T) {
	// Create mocks for all repositories
	mockIncomeRepo := &MockIncomeProfileRepository{}
	mockBudgetRepo := &MockBudgetConstraintRepository{}
	mockGoalRepo := &MockGoalRepository{}
	mockDebtRepo := &MockDebtRepository{}
	mockCategoryRepo := &MockCategoryRepository{}

	// Setup mock expectations
	mockIncomeRepo.On("GetByUserAndPeriod", ...).Return(...)
	mockBudgetRepo.On("GetByUser", ...).Return(...)
	// etc.

	// Create service with mocks
	service := NewAllocationService(
		mockIncomeRepo,
		mockBudgetRepo,
		mockGoalRepo,
		mockDebtRepo,
		mockCategoryRepo,
	)

	// Execute test
	req := &dto.GenerateAllocationRequest{...}
	resp, err := service.GenerateAllocations(context.Background(), req)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Scenarios, 3)
	// etc.
}
*/
