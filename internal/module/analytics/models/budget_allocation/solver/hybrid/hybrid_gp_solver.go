package hybrid

import (
	"fmt"
	"personalfinancedss/internal/module/analytics/budget_allocation/domain"
	"sort"

	"github.com/google/uuid"
)

// =============================================================================
// Hybrid GP Solver: 3-Phase Approach
// Phase 1: Heuristic (ensure minimums)
// Phase 2: Proportional (allocate surplus by ratios)
// Phase 3: Meta GP (optimize within buckets by satisfaction levels)
// =============================================================================

// HybridGPSolver implements the 3-phase hybrid approach
type HybridGPSolver struct {
	model  *domain.ConstraintModel
	params domain.ScenarioParameters
}

// HybridResult contains the complete result from hybrid solver
type HybridResult struct {
	// Phase 1 results
	MandatoryAllocations map[uuid.UUID]float64
	DebtMinAllocations   map[uuid.UUID]float64
	Phase1Total          float64
	Surplus              float64
	IsFeasible           bool
	DeficitAmount        float64

	// Phase 2 results
	Buckets map[string]*BudgetBucket

	// Phase 3 results
	FinalAllocations  map[uuid.UUID]AllocationDetail
	TotalAllocated    float64
	TotalReward       float64
	MaxPossibleReward float64
}

// TargetLevel represents a satisfaction level for a goal
type TargetLevel struct {
	Level       string  // "minimum", "satisfactory", "ideal"
	Value       float64 // Amount needed to achieve this level
	Reward      float64 // Reward points for achieving this level
	Description string  // Optional description
}

// BudgetBucket represents a category of spending with allocated budget
type BudgetBucket struct {
	Name        string
	Budget      float64
	Allocated   float64
	Remaining   float64
	Items       []BucketItem
	TotalReward float64
}

// BucketItem represents an item within a bucket
type BucketItem struct {
	ID            uuid.UUID
	Name          string
	Type          string // "goal", "debt_extra", "flexible"
	TargetLevels  []TargetLevel
	AchievedLevel string
	Allocated     float64
	Reward        float64
	Priority      int     // For sorting within bucket
	Weight        float64 // For reward calculation
}

// AllocationDetail contains detailed allocation info
type AllocationDetail struct {
	ID            uuid.UUID
	Name          string
	Type          string
	Allocated     float64
	Target        float64
	AchievedLevel string
	Reward        float64
	Bucket        string
}

// NewHybridGPSolver creates a new hybrid solver
func NewHybridGPSolver(model *domain.ConstraintModel, params domain.ScenarioParameters) *HybridGPSolver {
	return &HybridGPSolver{
		model:  model,
		params: params,
	}
}

// Solve executes the 3-phase hybrid algorithm
func (h *HybridGPSolver) Solve() (*HybridResult, error) {
	result := &HybridResult{
		MandatoryAllocations: make(map[uuid.UUID]float64),
		DebtMinAllocations:   make(map[uuid.UUID]float64),
		Buckets:              make(map[string]*BudgetBucket),
		FinalAllocations:     make(map[uuid.UUID]AllocationDetail),
		IsFeasible:           true,
	}

	// Phase 1: Ensure minimums
	h.phase1EnsureMinimums(result)

	if !result.IsFeasible {
		return result, fmt.Errorf("budget deficit: %.2f", result.DeficitAmount)
	}

	// Phase 2: Allocate surplus proportionally
	h.phase2AllocateSurplus(result)

	// Phase 3: Optimize within buckets using Meta GP logic
	h.phase3OptimizeBuckets(result)

	// Calculate totals
	h.calculateTotals(result)

	return result, nil
}

// =============================================================================
// Phase 1: Heuristic - Ensure Minimums
// =============================================================================

func (h *HybridGPSolver) phase1EnsureMinimums(result *HybridResult) {
	totalMinimum := 0.0

	// Allocate mandatory expenses
	for id, constraint := range h.model.MandatoryExpenses {
		result.MandatoryAllocations[id] = constraint.Minimum
		totalMinimum += constraint.Minimum

		result.FinalAllocations[id] = AllocationDetail{
			ID:            id,
			Name:          fmt.Sprintf("Mandatory_%s", id.String()[:8]),
			Type:          "mandatory",
			Allocated:     constraint.Minimum,
			Target:        constraint.Minimum,
			AchievedLevel: "required",
			Reward:        100, // Full reward for mandatory
			Bucket:        "mandatory",
		}
	}

	// Allocate minimum debt payments
	for id, constraint := range h.model.DebtPayments {
		result.DebtMinAllocations[id] = constraint.MinimumPayment
		totalMinimum += constraint.MinimumPayment

		result.FinalAllocations[id] = AllocationDetail{
			ID:            id,
			Name:          constraint.DebtName,
			Type:          "debt_min",
			Allocated:     constraint.MinimumPayment,
			Target:        constraint.MinimumPayment,
			AchievedLevel: "required",
			Reward:        100,
			Bucket:        "debt_minimum",
		}
	}

	result.Phase1Total = totalMinimum

	// Check feasibility
	if totalMinimum > h.model.TotalIncome {
		result.IsFeasible = false
		result.DeficitAmount = totalMinimum - h.model.TotalIncome
		result.Surplus = 0
	} else {
		result.Surplus = h.model.TotalIncome - totalMinimum
	}
}

// =============================================================================
// Phase 2: Proportional - Allocate Surplus by Ratios
// =============================================================================

func (h *HybridGPSolver) phase2AllocateSurplus(result *HybridResult) {
	if result.Surplus <= 0 {
		return
	}

	sa := h.params.SurplusAllocation

	// Create buckets with allocated budgets
	result.Buckets["emergency"] = &BudgetBucket{
		Name:   "Quỹ Khẩn Cấp",
		Budget: result.Surplus * sa.EmergencyFundPercent,
		Items:  make([]BucketItem, 0),
	}

	result.Buckets["debt_extra"] = &BudgetBucket{
		Name:   "Trả Nợ Thêm",
		Budget: result.Surplus * sa.DebtExtraPercent,
		Items:  make([]BucketItem, 0),
	}

	result.Buckets["goals"] = &BudgetBucket{
		Name:   "Mục Tiêu",
		Budget: result.Surplus * sa.GoalsPercent,
		Items:  make([]BucketItem, 0),
	}

	result.Buckets["flexible"] = &BudgetBucket{
		Name:   "Chi Linh Hoạt",
		Budget: result.Surplus * sa.FlexiblePercent,
		Items:  make([]BucketItem, 0),
	}

	// Populate bucket items
	h.populateBucketItems(result)
}

func (h *HybridGPSolver) populateBucketItems(result *HybridResult) {
	// Emergency fund goals
	for id, constraint := range h.model.GoalTargets {
		if constraint.GoalType == "emergency" {
			target := constraint.SuggestedContribution * h.params.GoalContributionFactor
			item := BucketItem{
				ID:           id,
				Name:         constraint.GoalName,
				Type:         "goal",
				TargetLevels: h.createGoalTargetLevels(target, "emergency"),
				Priority:     1, // Emergency is highest priority
				Weight:       1.0,
			}
			result.Buckets["emergency"].Items = append(result.Buckets["emergency"].Items, item)
		}
	}

	// Extra debt payments (sorted by interest rate - higher first)
	for id, constraint := range h.model.DebtPayments {
		// Extra target = additional payment beyond minimum
		extraTarget := constraint.MinimumPayment * h.params.SurplusAllocation.DebtExtraPercent * 10
		if extraTarget > constraint.CurrentBalance-constraint.MinimumPayment {
			extraTarget = constraint.CurrentBalance - constraint.MinimumPayment
		}

		if extraTarget > 0 {
			item := BucketItem{
				ID:           id,
				Name:         constraint.DebtName + " (extra)",
				Type:         "debt_extra",
				TargetLevels: h.createDebtTargetLevels(extraTarget, constraint.InterestRate),
				Priority:     constraint.Priority,
				Weight:       constraint.InterestRate / 10, // Higher interest = higher weight
			}
			result.Buckets["debt_extra"].Items = append(result.Buckets["debt_extra"].Items, item)
		}
	}

	// Non-emergency goals
	for id, constraint := range h.model.GoalTargets {
		if constraint.GoalType != "emergency" {
			target := constraint.SuggestedContribution * h.params.GoalContributionFactor
			item := BucketItem{
				ID:           id,
				Name:         constraint.GoalName,
				Type:         "goal",
				TargetLevels: h.createGoalTargetLevels(target, constraint.Priority),
				Priority:     constraint.PriorityWeight,
				Weight:       float64(100 - constraint.PriorityWeight),
			}
			result.Buckets["goals"].Items = append(result.Buckets["goals"].Items, item)
		}
	}

	// Flexible expenses
	for id, constraint := range h.model.FlexibleExpenses {
		// Target is the range above minimum
		extraTarget := (constraint.Maximum - constraint.Minimum) * h.params.FlexibleSpendingLevel
		if extraTarget > 0 {
			item := BucketItem{
				ID:           id,
				Name:         fmt.Sprintf("Flexible_%s", id.String()[:8]),
				Type:         "flexible",
				TargetLevels: h.createFlexibleTargetLevels(constraint.Minimum, extraTarget),
				Priority:     constraint.Priority,
				Weight:       1.0,
			}
			result.Buckets["flexible"].Items = append(result.Buckets["flexible"].Items, item)
		}
	}
}

// =============================================================================
// Helper Functions: Create Target Levels
// =============================================================================

// createGoalTargetLevels creates 3 satisfaction levels for a goal
func (h *HybridGPSolver) createGoalTargetLevels(target float64, priority string) []TargetLevel {
	// Adjust percentages based on priority
	var minPct, satPct float64
	var minReward, satReward, idealReward float64

	switch priority {
	case "emergency", "critical":
		minPct, satPct = 0.5, 0.8
		minReward, satReward, idealReward = 40, 75, 100
	case "high":
		minPct, satPct = 0.3, 0.7
		minReward, satReward, idealReward = 30, 65, 100
	case "medium":
		minPct, satPct = 0.3, 0.6
		minReward, satReward, idealReward = 25, 55, 100
	default: // low
		minPct, satPct = 0.2, 0.5
		minReward, satReward, idealReward = 20, 50, 100
	}

	return []TargetLevel{
		{Level: "minimum", Value: target * minPct, Reward: minReward},
		{Level: "satisfactory", Value: target * satPct, Reward: satReward},
		{Level: "ideal", Value: target, Reward: idealReward},
	}
}

// createDebtTargetLevels creates levels for extra debt payments
// Higher interest debts get higher rewards for paying extra
func (h *HybridGPSolver) createDebtTargetLevels(extraTarget float64, interestRate float64) []TargetLevel {
	// Reward scales with interest rate (paying high-interest debt is more valuable)
	baseReward := 50.0
	if interestRate >= 20 {
		baseReward = 80
	} else if interestRate >= 10 {
		baseReward = 65
	}

	return []TargetLevel{
		{Level: "minimum", Value: extraTarget * 0.3, Reward: baseReward * 0.4},
		{Level: "satisfactory", Value: extraTarget * 0.7, Reward: baseReward * 0.75},
		{Level: "ideal", Value: extraTarget, Reward: baseReward},
	}
}

// createFlexibleTargetLevels creates levels for flexible spending
func (h *HybridGPSolver) createFlexibleTargetLevels(minimum float64, extraRange float64) []TargetLevel {
	return []TargetLevel{
		{Level: "minimum", Value: 0, Reward: 0}, // Already allocated in Phase 1
		{Level: "satisfactory", Value: extraRange * 0.5, Reward: 40},
		{Level: "ideal", Value: extraRange, Reward: 70},
	}
}

// =============================================================================
// Phase 3: Meta GP - Optimize Within Buckets
// =============================================================================

func (h *HybridGPSolver) phase3OptimizeBuckets(result *HybridResult) {
	for bucketName, bucket := range result.Buckets {
		if bucket.Budget <= 0 || len(bucket.Items) == 0 {
			continue
		}

		h.optimizeBucket(bucket, bucketName, result)
	}
}

// optimizeBucket applies Meta GP logic within a single bucket
func (h *HybridGPSolver) optimizeBucket(bucket *BudgetBucket, bucketName string, result *HybridResult) {
	// Sort items by priority (lower = higher priority)
	sort.Slice(bucket.Items, func(i, j int) bool {
		return bucket.Items[i].Priority < bucket.Items[j].Priority
	})

	remainingBudget := bucket.Budget

	// First pass: Try to achieve highest possible level for each item
	for i := range bucket.Items {
		item := &bucket.Items[i]
		if len(item.TargetLevels) == 0 {
			continue
		}

		// Try from highest level to lowest
		for levelIdx := len(item.TargetLevels) - 1; levelIdx >= 0; levelIdx-- {
			level := item.TargetLevels[levelIdx]
			if level.Value <= remainingBudget {
				item.Allocated = level.Value
				item.AchievedLevel = level.Level
				item.Reward = level.Reward * item.Weight
				remainingBudget -= level.Value
				break
			}
		}

		// If no level achieved, try partial allocation
		if item.Allocated == 0 && remainingBudget > 0 {
			// Allocate what we can
			minLevel := item.TargetLevels[0]
			if remainingBudget >= minLevel.Value*0.5 {
				item.Allocated = remainingBudget
				item.AchievedLevel = "partial"
				item.Reward = minLevel.Reward * 0.5 * item.Weight
				remainingBudget = 0
			}
		}

		// Update final allocations
		if item.Allocated > 0 {
			result.FinalAllocations[item.ID] = AllocationDetail{
				ID:            item.ID,
				Name:          item.Name,
				Type:          item.Type,
				Allocated:     item.Allocated,
				Target:        item.TargetLevels[len(item.TargetLevels)-1].Value,
				AchievedLevel: item.AchievedLevel,
				Reward:        item.Reward,
				Bucket:        bucketName,
			}
		}

		bucket.TotalReward += item.Reward
	}

	// Second pass: If budget remains, try to upgrade items to higher levels
	if remainingBudget > 0.01 {
		h.upgradeItemsInBucket(bucket, &remainingBudget, bucketName, result)
	}

	bucket.Allocated = bucket.Budget - remainingBudget
	bucket.Remaining = remainingBudget
}

// upgradeItemsInBucket tries to upgrade items to higher satisfaction levels
func (h *HybridGPSolver) upgradeItemsInBucket(bucket *BudgetBucket, remainingBudget *float64, bucketName string, result *HybridResult) {
	type upgradeOption struct {
		itemIdx          int
		currentLevelIdx  int
		nextLevelIdx     int
		cost             float64
		additionalReward float64
		rewardPerCost    float64
	}

	for *remainingBudget > 0.01 {
		var options []upgradeOption

		for i := range bucket.Items {
			item := &bucket.Items[i]
			if len(item.TargetLevels) == 0 {
				continue
			}

			// Find current level index
			currentIdx := -1
			for j, level := range item.TargetLevels {
				if item.AchievedLevel == level.Level {
					currentIdx = j
					break
				}
			}

			// Check if upgrade is possible
			nextIdx := currentIdx + 1
			if nextIdx < len(item.TargetLevels) {
				nextLevel := item.TargetLevels[nextIdx]
				cost := nextLevel.Value - item.Allocated

				if cost > 0 && cost <= *remainingBudget {
					currentReward := 0.0
					if currentIdx >= 0 {
						currentReward = item.TargetLevels[currentIdx].Reward
					}
					additionalReward := (nextLevel.Reward - currentReward) * item.Weight

					options = append(options, upgradeOption{
						itemIdx:          i,
						currentLevelIdx:  currentIdx,
						nextLevelIdx:     nextIdx,
						cost:             cost,
						additionalReward: additionalReward,
						rewardPerCost:    additionalReward / cost,
					})
				}
			}
		}

		if len(options) == 0 {
			break
		}

		// Sort by reward per cost (best value first)
		sort.Slice(options, func(i, j int) bool {
			return options[i].rewardPerCost > options[j].rewardPerCost
		})

		// Apply best upgrade
		best := options[0]
		item := &bucket.Items[best.itemIdx]
		nextLevel := item.TargetLevels[best.nextLevelIdx]

		item.Allocated = nextLevel.Value
		item.AchievedLevel = nextLevel.Level
		item.Reward = nextLevel.Reward * item.Weight
		*remainingBudget -= best.cost
		bucket.TotalReward += best.additionalReward

		// Update final allocation
		result.FinalAllocations[item.ID] = AllocationDetail{
			ID:            item.ID,
			Name:          item.Name,
			Type:          item.Type,
			Allocated:     item.Allocated,
			Target:        item.TargetLevels[len(item.TargetLevels)-1].Value,
			AchievedLevel: item.AchievedLevel,
			Reward:        item.Reward,
			Bucket:        bucketName,
		}
	}
}

// =============================================================================
// Calculate Totals
// =============================================================================

func (h *HybridGPSolver) calculateTotals(result *HybridResult) {
	totalAllocated := result.Phase1Total
	totalReward := 0.0
	maxReward := 0.0

	// Add Phase 1 rewards (mandatory = 100 points each)
	for range result.MandatoryAllocations {
		totalReward += 100
		maxReward += 100
	}
	for range result.DebtMinAllocations {
		totalReward += 100
		maxReward += 100
	}

	// Add Phase 3 allocations and rewards
	for _, bucket := range result.Buckets {
		totalAllocated += bucket.Allocated
		totalReward += bucket.TotalReward

		// Calculate max possible reward for this bucket
		for _, item := range bucket.Items {
			if len(item.TargetLevels) > 0 {
				maxReward += item.TargetLevels[len(item.TargetLevels)-1].Reward * item.Weight
			}
		}
	}

	result.TotalAllocated = totalAllocated
	result.TotalReward = totalReward
	result.MaxPossibleReward = maxReward
}

// Note: Integration with GoalProgrammingSolver is handled in the parent solver package
// via the SolveHybrid and mapHybridResult methods in goal_programming.go
