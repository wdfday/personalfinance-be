package domain

// GoalType represents the type of financial goal
type GoalType string

const (
	GoalTypeSavings    GoalType = "savings"    // General savings goal
	GoalTypeDebt       GoalType = "debt"       // Debt repayment goal
	GoalTypeInvestment GoalType = "investment" // Investment target goal
	GoalTypePurchase   GoalType = "purchase"   // Big purchase goal (house, car, etc.)
	GoalTypeEmergency  GoalType = "emergency"  // Emergency fund goal
	GoalTypeRetirement GoalType = "retirement" // Retirement savings goal
	GoalTypeEducation  GoalType = "education"  // Education fund goal
	GoalTypeOther      GoalType = "other"      // Custom goal
)

// IsValid checks if the goal type is valid
func (gt GoalType) IsValid() bool {
	switch gt {
	case GoalTypeSavings, GoalTypeDebt, GoalTypeInvestment, GoalTypePurchase,
		GoalTypeEmergency, GoalTypeRetirement, GoalTypeEducation, GoalTypeOther:
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
)

// IsValid checks if the goal status is valid
func (gs GoalStatus) IsValid() bool {
	switch gs {
	case GoalStatusActive, GoalStatusCompleted, GoalStatusPaused, GoalStatusCancelled, GoalStatusOverdue:
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
