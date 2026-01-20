package dto

import (
	"time"

	"github.com/google/uuid"
)

// GoalMonthlySummary represents aggregated goal contribution data for a specific month
type GoalMonthlySummary struct {
	GoalID            uuid.UUID `json:"goal_id"`
	Name              string    `json:"name"`
	TotalContributed  float64   `json:"total_contributed"`
	ContributionCount int       `json:"contribution_count"`
}

// GoalAllTimeSummary represents aggregated goal contribution data from inception to present
type GoalAllTimeSummary struct {
	GoalID            uuid.UUID  `json:"goal_id"`
	Name              string     `json:"name"`
	TotalContributed  float64    `json:"total_contributed"` // Total deposits
	TotalWithdrawn    float64    `json:"total_withdrawn"`   // Total withdrawals
	NetContributed    float64    `json:"net_contributed"`   // = Contributed - Withdrawn
	ContributionCount int        `json:"contribution_count"`
	FirstContribution *time.Time `json:"first_contribution,omitempty"`
	LastContribution  *time.Time `json:"last_contribution,omitempty"`
}
