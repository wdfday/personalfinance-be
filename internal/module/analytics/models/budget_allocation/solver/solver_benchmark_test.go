package solver

import (
	"fmt"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/lp"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/preemptive"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Benchmark Tests để so sánh các LP/GP Solver
// =============================================================================

// createTestConstraintModel tạo dữ liệu test
func createTestConstraintModel(numCategories, numGoals, numDebts int) *domain.ConstraintModel {
	model := &domain.ConstraintModel{
		TotalIncome:       100_000_000, // 100 triệu VNĐ
		MandatoryExpenses: make(map[uuid.UUID]domain.CategoryConstraint),
		FlexibleExpenses:  make(map[uuid.UUID]domain.CategoryConstraint),
		DebtPayments:      make(map[uuid.UUID]domain.DebtConstraint),
		GoalTargets:       make(map[uuid.UUID]domain.GoalConstraint),
	}

	// Mandatory expenses
	for i := 0; i < numCategories/2; i++ {
		id := uuid.New()
		model.MandatoryExpenses[id] = domain.CategoryConstraint{
			CategoryID: id,
			Minimum:    float64(5_000_000 + i*1_000_000),
			Maximum:    float64(5_000_000 + i*1_000_000),
			IsFlexible: false,
			Priority:   1,
		}
	}

	// Flexible expenses
	for i := 0; i < numCategories/2; i++ {
		id := uuid.New()
		model.FlexibleExpenses[id] = domain.CategoryConstraint{
			CategoryID: id,
			Minimum:    float64(2_000_000),
			Maximum:    float64(5_000_000 + i*500_000),
			IsFlexible: true,
			Priority:   5,
		}
	}

	// Goals
	for i := 0; i < numGoals; i++ {
		id := uuid.New()
		goalType := "savings"
		if i == 0 {
			goalType = "emergency"
		}
		model.GoalTargets[id] = domain.GoalConstraint{
			GoalID:                id,
			GoalName:              fmt.Sprintf("Goal_%d", i),
			GoalType:              goalType,
			SuggestedContribution: float64(5_000_000 + i*1_000_000),
			Priority:              "medium",
			PriorityWeight:        i + 1,
			RemainingAmount:       float64(40_000_000 + i*10_000_000),
		}
	}

	// Debts
	for i := 0; i < numDebts; i++ {
		id := uuid.New()
		model.DebtPayments[id] = domain.DebtConstraint{
			DebtID:         id,
			DebtName:       fmt.Sprintf("Debt_%d", i),
			MinimumPayment: float64(2_000_000 + i*500_000),
			CurrentBalance: float64(50_000_000 + i*20_000_000),
			InterestRate:   float64(15 + i*5),
			Priority:       i + 1,
		}
	}

	return model
}

// TestSolverComparison so sánh kết quả giữa các solver
func TestSolverComparison(t *testing.T) {
	model := createTestConstraintModel(4, 3, 2)
	params := domain.ScenarioParameters{
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40,
			DebtExtraPercent:     0.30,
			GoalsPercent:         0.20,
			FlexiblePercent:      0.10,
		},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SO SÁNH CÁC SOLVER")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Tổng thu nhập: %.0f VNĐ\n", model.TotalIncome)
	fmt.Printf("Số biến: %d categories + %d goals + %d debts\n",
		len(model.MandatoryExpenses)+len(model.FlexibleExpenses),
		len(model.GoalTargets),
		len(model.DebtPayments))
	fmt.Println()

	// 1. Heuristic Preemptive GP
	fmt.Println("1. HEURISTIC PREEMPTIVE GP")
	fmt.Println(strings.Repeat("-", 40))
	start := time.Now()
	heuristicSolver := preemptive.BuildPreemptiveGPFromConstraintModel(model, params)
	heuristicResult, _ := heuristicSolver.Solve()
	heuristicTime := time.Since(start)
	printPreemptiveGPResult("Heuristic", heuristicResult, heuristicTime)

	// 2. LP-based GP (Pure Go Simplex)
	fmt.Println("\n2. LP-BASED GP (Pure Go Simplex)")
	fmt.Println(strings.Repeat("-", 40))
	start = time.Now()
	simplexSolver := lp.BuildLPGPFromConstraintModel(model, params, "simplex")
	simplexResult, _ := simplexSolver.Solve()
	simplexTime := time.Since(start)
	printLPGPResult("Simplex", simplexResult, simplexTime)

	// 3. LP-based GP (golp) - nếu có
	fmt.Println("\n3. LP-BASED GP (golp/lp_solve)")
	fmt.Println(strings.Repeat("-", 40))
	start = time.Now()
	golpSolver := lp.BuildLPGPFromConstraintModel(model, params, "golp")
	golpResult, err := golpSolver.Solve()
	golpTime := time.Since(start)
	if err != nil {
		fmt.Printf("   golp không khả dụng: %v\n", err)
	} else {
		printLPGPResult("golp", golpResult, golpTime)
	}

	// So sánh
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("BẢNG SO SÁNH")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-20s | %-15s | %-15s | %-15s | %-10s\n",
		"Solver", "Total Deviation", "Achieved", "Unachieved", "Time")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-20s | %15.2f | %15d | %15d | %10s\n",
		"Heuristic", heuristicResult.TotalDeviation,
		len(heuristicResult.AchievedGoals), len(heuristicResult.UnachievedGoals),
		heuristicTime)
	fmt.Printf("%-20s | %15.2f | %15d | %15d | %10s\n",
		"Simplex", simplexResult.TotalDeviation,
		len(simplexResult.AchievedGoals), len(simplexResult.UnachievedGoals),
		simplexTime)
	if golpResult != nil {
		fmt.Printf("%-20s | %15.2f | %15d | %15d | %10s\n",
			"golp", golpResult.TotalDeviation,
			len(golpResult.AchievedGoals), len(golpResult.UnachievedGoals),
			golpTime)
	}
}

func printPreemptiveGPResult(name string, result *preemptive.GPResult, duration time.Duration) {
	fmt.Printf("   Solver: %s\n", name)
	fmt.Printf("   Feasible: %v\n", result.IsFeasible)
	fmt.Printf("   Total Deviation: %.2f\n", result.TotalDeviation)
	fmt.Printf("   Achieved Goals: %d\n", len(result.AchievedGoals))
	fmt.Printf("   Unachieved Goals: %d\n", len(result.UnachievedGoals))
	fmt.Printf("   Iterations: %d\n", result.SolverIterations)
	fmt.Printf("   Time: %s\n", duration)

	// Print allocations
	fmt.Println("   Allocations:")
	total := 0.0
	for id, value := range result.VariableValues {
		if value > 0 {
			fmt.Printf("      %s: %.0f\n", id.String()[:8], value)
			total += value
		}
	}
	fmt.Printf("   Total Allocated: %.0f\n", total)
}

func printLPGPResult(name string, result *lp.GPResult, duration time.Duration) {
	fmt.Printf("   Solver: %s\n", name)
	fmt.Printf("   Feasible: %v\n", result.IsFeasible)
	fmt.Printf("   Total Deviation: %.2f\n", result.TotalDeviation)
	fmt.Printf("   Achieved Goals: %d\n", len(result.AchievedGoals))
	fmt.Printf("   Unachieved Goals: %d\n", len(result.UnachievedGoals))
	fmt.Printf("   Iterations: %d\n", result.SolverIterations)
	fmt.Printf("   Time: %s\n", duration)

	// Print allocations
	fmt.Println("   Allocations:")
	total := 0.0
	for id, value := range result.VariableValues {
		if value > 0 {
			fmt.Printf("      %s: %.0f\n", id.String()[:8], value)
			total += value
		}
	}
	fmt.Printf("   Total Allocated: %.0f\n", total)
}

// BenchmarkHeuristicGP benchmark cho Heuristic GP
func BenchmarkHeuristicGP(b *testing.B) {
	model := createTestConstraintModel(6, 5, 3)
	params := domain.ScenarioParameters{
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40,
			DebtExtraPercent:     0.30,
			GoalsPercent:         0.20,
			FlexiblePercent:      0.10,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver := preemptive.BuildPreemptiveGPFromConstraintModel(model, params)
		solver.Solve()
	}
}

// BenchmarkSimplexGP benchmark cho Simplex GP
func BenchmarkSimplexGP(b *testing.B) {
	model := createTestConstraintModel(6, 5, 3)
	params := domain.ScenarioParameters{
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40,
			DebtExtraPercent:     0.30,
			GoalsPercent:         0.20,
			FlexiblePercent:      0.10,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver := lp.BuildLPGPFromConstraintModel(model, params, "simplex")
		solver.Solve()
	}
}

// BenchmarkGolpGP benchmark cho golp GP
func BenchmarkGolpGP(b *testing.B) {
	model := createTestConstraintModel(6, 5, 3)
	params := domain.ScenarioParameters{
		GoalContributionFactor: 1.0,
		FlexibleSpendingLevel:  0.5,
		SurplusAllocation: domain.SurplusAllocation{
			EmergencyFundPercent: 0.40,
			DebtExtraPercent:     0.30,
			GoalsPercent:         0.20,
			FlexiblePercent:      0.10,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver := lp.BuildLPGPFromConstraintModel(model, params, "golp")
		solver.Solve()
	}
}

// TestLPSolverBasic test cơ bản cho LP solver
func TestLPSolverBasic(t *testing.T) {
	// Simple LP problem:
	// minimize: -x1 - x2
	// subject to: x1 + x2 <= 10
	//             x1 <= 6
	//             x2 <= 6
	//             x1, x2 >= 0

	solver := lp.NewPureGoSimplexSolver(2)
	solver.SetObjective([]float64{-1, -1}, false) // minimize -x1 - x2 = maximize x1 + x2
	solver.AddConstraint([]float64{1, 1}, "<=", 10)
	solver.AddConstraint([]float64{1, 0}, "<=", 6)
	solver.AddConstraint([]float64{0, 1}, "<=", 6)
	solver.SetBounds(0, 0, 100)
	solver.SetBounds(1, 0, 100)

	result, err := solver.Solve()
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	t.Logf("Status: %s", result.Status)
	t.Logf("Solution: x1=%.2f, x2=%.2f", result.Solution[0], result.Solution[1])
	t.Logf("Objective: %.2f", result.ObjectiveValue)
	t.Logf("Iterations: %d", result.Iterations)

	// Expected: x1 + x2 = 10, with x1 <= 6 and x2 <= 6
	// Optimal: x1 = 6, x2 = 4 or x1 = 4, x2 = 6
	sum := result.Solution[0] + result.Solution[1]
	if sum < 9.9 || sum > 10.1 {
		t.Errorf("Expected sum ~10, got %.2f", sum)
	}
}
