package domain

// GoalBehavior represents how the goal accumulates funds
type GoalBehavior string

const (
	// GoalBehaviorFlexible - target balance, contribute when possible, no deadline required
	GoalBehaviorFlexible GoalBehavior = "flexible"

	// GoalBehaviorWilling - target + deadline, converts to Budget when deadline reached
	GoalBehaviorWilling GoalBehavior = "willing"

	// GoalBehaviorRecurring - periodic contributions required, may have deadline
	GoalBehaviorRecurring GoalBehavior = "recurring"
)

// IsValid checks if the goal behavior is valid
func (gb GoalBehavior) IsValid() bool {
	switch gb {
	case GoalBehaviorFlexible, GoalBehaviorWilling, GoalBehaviorRecurring:
		return true
	}
	return false
}

// GoalCategory represents the category/purpose of financial goal
type GoalCategory string

const (
	GoalCategorySavings    GoalCategory = "savings"    // General savings goal
	GoalCategoryDebt       GoalCategory = "debt"       // Debt repayment goal
	GoalCategoryInvestment GoalCategory = "investment" // Investment target goal
	GoalCategoryPurchase   GoalCategory = "purchase"   // Big purchase goal (house, car, etc.)
	GoalCategoryEmergency  GoalCategory = "emergency"  // Emergency fund goal
	GoalCategoryRetirement GoalCategory = "retirement" // Retirement savings goal
	GoalCategoryEducation  GoalCategory = "education"  // Education fund goal
	GoalCategoryTravel     GoalCategory = "travel"     // Travel goal
	GoalCategoryOther      GoalCategory = "other"      // Custom goal
)

// IsValid checks if the goal category is valid
func (gc GoalCategory) IsValid() bool {
	switch gc {
	case GoalCategorySavings, GoalCategoryDebt, GoalCategoryInvestment, GoalCategoryPurchase,
		GoalCategoryEmergency, GoalCategoryRetirement, GoalCategoryEducation, GoalCategoryTravel, GoalCategoryOther:
		return true
	}
	return false
}

// GoalStatus represents the current status of a goal
type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"    // Goal is active
	GoalStatusCompleted GoalStatus = "completed" // Goal has been achieved
	GoalStatusPaused    GoalStatus = "paused"    // Goal is temporarily paused
	GoalStatusCancelled GoalStatus = "cancelled" // Goal was cancelled
	GoalStatusOverdue   GoalStatus = "overdue"   // Goal deadline has passed
	GoalStatusArchived  GoalStatus = "archived"  // Goal is archived (soft deleted)
)

// IsValid checks if the goal status is valid
func (gs GoalStatus) IsValid() bool {
	switch gs {
	case GoalStatusActive, GoalStatusCompleted, GoalStatusPaused, GoalStatusCancelled, GoalStatusOverdue, GoalStatusArchived:
		return true
	}
	return false
}

// GoalPriority represents the priority level of a goal
type GoalPriority string

const (
	GoalPriorityLow      GoalPriority = "low"
	GoalPriorityMedium   GoalPriority = "medium"
	GoalPriorityHigh     GoalPriority = "high"
	GoalPriorityCritical GoalPriority = "critical"
)

// IsValid checks if the goal priority is valid
func (gp GoalPriority) IsValid() bool {
	switch gp {
	case GoalPriorityLow, GoalPriorityMedium, GoalPriorityHigh, GoalPriorityCritical:
		return true
	}
	return false
}

// ContributionFrequency represents how often to contribute to a goal
type ContributionFrequency string

const (
	FrequencyOneTime   ContributionFrequency = "one_time"
	FrequencyDaily     ContributionFrequency = "daily"
	FrequencyWeekly    ContributionFrequency = "weekly"
	FrequencyBiweekly  ContributionFrequency = "biweekly"
	FrequencyMonthly   ContributionFrequency = "monthly"
	FrequencyQuarterly ContributionFrequency = "quarterly"
	FrequencyYearly    ContributionFrequency = "yearly"
)

// IsValid checks if the contribution frequency is valid
func (cf ContributionFrequency) IsValid() bool {
	switch cf {
	case FrequencyOneTime, FrequencyDaily, FrequencyWeekly, FrequencyBiweekly,
		FrequencyMonthly, FrequencyQuarterly, FrequencyYearly:
		return true
	}
	return false
}

// DaysPerPeriod returns the number of days in one period
func (cf ContributionFrequency) DaysPerPeriod() int {
	switch cf {
	case FrequencyDaily:
		return 1
	case FrequencyWeekly:
		return 7
	case FrequencyBiweekly:
		return 14
	case FrequencyMonthly:
		return 30
	case FrequencyQuarterly:
		return 90
	case FrequencyYearly:
		return 365
	case FrequencyOneTime:
		return 0
	default:
		return 0
	}
}
