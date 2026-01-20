package domain

// BudgetPeriod represents the period type for a budget
type BudgetPeriod string

const (
	BudgetPeriodDaily     BudgetPeriod = "daily"
	BudgetPeriodWeekly    BudgetPeriod = "weekly"
	BudgetPeriodMonthly   BudgetPeriod = "monthly"
	BudgetPeriodQuarterly BudgetPeriod = "quarterly"
	BudgetPeriodYearly    BudgetPeriod = "yearly"
	BudgetPeriodCustom    BudgetPeriod = "custom"
)

// IsValid checks if the budget period is valid
func (bp BudgetPeriod) IsValid() bool {
	switch bp {
	case BudgetPeriodDaily, BudgetPeriodWeekly, BudgetPeriodMonthly, BudgetPeriodQuarterly, BudgetPeriodYearly, BudgetPeriodCustom:
		return true
	}
	return false
}

// BudgetStatus represents the current status of a budget
type BudgetStatus string

const (
	BudgetStatusActive   BudgetStatus = "active"
	BudgetStatusExceeded BudgetStatus = "exceeded"
	BudgetStatusWarning  BudgetStatus = "warning" // 80-100% spent
	BudgetStatusPaused   BudgetStatus = "paused"
	BudgetStatusExpired  BudgetStatus = "expired"
)

// IsValid checks if the budget status is valid
func (bs BudgetStatus) IsValid() bool {
	switch bs {
	case BudgetStatusActive, BudgetStatusExceeded, BudgetStatusWarning, BudgetStatusPaused, BudgetStatusExpired:
		return true
	}
	return false
}

// AlertThreshold represents when to send alerts
type AlertThreshold string

const (
	AlertAt50  AlertThreshold = "50"  // Alert at 50% spent
	AlertAt75  AlertThreshold = "75"  // Alert at 75% spent
	AlertAt90  AlertThreshold = "90"  // Alert at 90% spent
	AlertAt100 AlertThreshold = "100" // Alert when exceeded
)

// IsValid checks if the alert threshold is valid
func (at AlertThreshold) IsValid() bool {
	switch at {
	case AlertAt50, AlertAt75, AlertAt90, AlertAt100:
		return true
	}
	return false
}

// ToFloat64 converts alert threshold to float64 percentage
func (at AlertThreshold) ToFloat64() float64 {
	switch at {
	case AlertAt50:
		return 50.0
	case AlertAt75:
		return 75.0
	case AlertAt90:
		return 90.0
	case AlertAt100:
		return 100.0
	}
	return 0.0
}
