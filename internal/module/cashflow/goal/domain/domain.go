package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Goal represents a financial goal
type Goal struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Goal Details
	Name        string       `gorm:"type:varchar(255);not null;column:name" json:"name"`
	Description *string      `gorm:"type:text;column:description" json:"description,omitempty"`
	Type        GoalType     `gorm:"type:varchar(50);not null;column:type" json:"type"`
	Priority    GoalPriority `gorm:"type:varchar(20);default:'medium';column:priority" json:"priority"`

	// Financial Details
	TargetAmount  float64 `gorm:"type:decimal(15,2);not null;column:target_amount" json:"target_amount"`
	CurrentAmount float64 `gorm:"type:decimal(15,2);default:0;column:current_amount" json:"current_amount"`
	Currency      string  `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`

	// Timeline
	StartDate   time.Time  `gorm:"type:date;not null;column:start_date" json:"start_date"`
	TargetDate  *time.Time `gorm:"type:date;column:target_date" json:"target_date,omitempty"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`

	// Progress Tracking
	PercentageComplete float64    `gorm:"type:decimal(5,2);default:0;column:percentage_complete" json:"percentage_complete"`
	RemainingAmount    float64    `gorm:"type:decimal(15,2);default:0;column:remaining_amount" json:"remaining_amount"`
	Status             GoalStatus `gorm:"type:varchar(20);default:'active';column:status" json:"status"`

	// Contribution Settings
	SuggestedContribution   *float64               `gorm:"type:decimal(15,2);column:suggested_contribution" json:"suggested_contribution,omitempty"`
	ContributionFrequency   *ContributionFrequency `gorm:"type:varchar(20);column:contribution_frequency" json:"contribution_frequency,omitempty"`
	AutoContribute          bool                   `gorm:"default:false;column:auto_contribute" json:"auto_contribute"`
	AutoContributeAmount    *float64               `gorm:"type:decimal(15,2);column:auto_contribute_amount" json:"auto_contribute_amount,omitempty"`
	AutoContributeAccountID *uuid.UUID             `gorm:"type:uuid;column:auto_contribute_account_id" json:"auto_contribute_account_id,omitempty"`

	// Linked Resources
	LinkedAccountID *uuid.UUID `gorm:"type:uuid;index;column:linked_account_id" json:"linked_account_id,omitempty"` // Account where funds are saved

	// Notifications
	EnableReminders    bool       `gorm:"default:true;column:enable_reminders" json:"enable_reminders"`
	ReminderFrequency  *string    `gorm:"type:varchar(20);column:reminder_frequency" json:"reminder_frequency,omitempty"` // "weekly", "monthly"
	LastReminderSentAt *time.Time `gorm:"column:last_reminder_sent_at" json:"last_reminder_sent_at,omitempty"`

	// Milestones
	Milestones *string `gorm:"type:jsonb;column:milestones" json:"milestones,omitempty"` // JSON array of milestone objects

	// Metadata
	Notes *string `gorm:"type:text;column:notes" json:"notes,omitempty"`
	Tags  *string `gorm:"type:jsonb;column:tags" json:"tags,omitempty"` // JSON array of tags

	// DSS Input/Output Metadata - Both constraint (INPUT) and allocation (OUTPUT)
	DSSMetadata datatypes.JSON `gorm:"type:jsonb;column:dss_metadata" json:"dss_metadata,omitempty"`
	// Structure:
	// {
	//   // INPUT - User constraints
	//   "minimum_contribution": 2000000,      // Min monthly contribution (INPUT)
	//   "target_contribution": 3000000,       // Target monthly contribution (INPUT)
	//   "maximum_contribution": 5000000,      // Max affordable contribution (INPUT)
	//   "constraint_type": "soft",            // hard, soft, aspirational (INPUT)
	//   "is_flexible": true,                  // Can adjust contribution (INPUT)
	//   "gp_priority": 2,                     // P1, P2, P3... (from AHP) (INPUT)
	//   "gp_weight": 0.82,                    // Weight from AHP score (INPUT)
	//
	//   // OUTPUT - DSS allocation
	//   "allocated_contribution": 2800000,    // DSS allocated contribution (OUTPUT)
	//   "satisfaction_level": 0.75,           // 0-1, achievement level (OUTPUT)
	//   "negative_deviation": 200000,         // d⁻ (under target) (OUTPUT)
	//   "positive_deviation": 0,              // d⁺ (over target) (OUTPUT)
	//   "achievement_rate": 0.93,             // Allocated/Target ratio (OUTPUT)
	//   "allocation_reason": "balanced_with_other_priorities",
	//
	//   // ANALYSIS
	//   "progress_rate": 0.65,                // Overall progress
	//   "required_monthly_rate": 3500000,     // Required to meet deadline
	//   "on_track": false,                    // Is on track to complete
	//   "months_remaining": 12,               // Months until deadline
	//   "required_acceleration": 1.2,         // Multiplier needed to catch up
	//   "feasibility_score": 0.85,            // How feasible to complete
	//   "completion_probability": 0.75,       // Probability of completion
	//   "estimated_completion_date": "2025-06-30",
	//   "last_optimized": "2024-01-15T10:00:00Z",
	//   "last_analyzed": "2024-01-15T10:00:00Z",
	//   "optimization_version": "v1.0"
	// }

	// AHP Metadata for Prioritization
	AHPMetadata datatypes.JSON `gorm:"type:jsonb;column:ahp_metadata" json:"ahp_metadata,omitempty"`
	// Structure:
	// {
	//   "criteria_scores": {
	//     "urgency": 0.8,                     // How urgent (deadline proximity)
	//     "importance": 0.9,                  // How important (user priority)
	//     "feasibility": 0.7,                 // How feasible (income vs target)
	//     "impact": 0.85                      // Financial impact
	//   },
	//   "criteria_weights": {
	//     "urgency": 0.3,
	//     "importance": 0.4,
	//     "feasibility": 0.2,
	//     "impact": 0.1
	//   },
	//   "overall_score": 0.82,                // Weighted sum
	//   "rank": 2,                            // Rank among all goals
	//   "consistency_ratio": 0.08,            // AHP consistency check
	//   "last_evaluated": "2024-01-15T10:00:00Z"
	// }

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Goal
func (Goal) TableName() string {
	return "goals"
}

// IsCompleted checks if the goal has been completed
func (g *Goal) IsCompleted() bool {
	return g.CurrentAmount >= g.TargetAmount || g.Status == GoalStatusCompleted
}

// IsOverdue checks if the goal is past its target date and not completed
func (g *Goal) IsOverdue() bool {
	if g.TargetDate == nil || g.IsCompleted() {
		return false
	}
	return time.Now().After(*g.TargetDate)
}

// DaysRemaining calculates days remaining until target date
func (g *Goal) DaysRemaining() int {
	if g.TargetDate == nil {
		return 0
	}
	duration := time.Until(*g.TargetDate)
	return int(duration.Hours() / 24)
}

// UpdateCalculatedFields updates progress, remaining amount, and status
func (g *Goal) UpdateCalculatedFields() {
	g.RemainingAmount = g.TargetAmount - g.CurrentAmount
	if g.RemainingAmount < 0 {
		g.RemainingAmount = 0
	}

	if g.TargetAmount > 0 {
		g.PercentageComplete = (g.CurrentAmount / g.TargetAmount) * 100
		if g.PercentageComplete > 100 {
			g.PercentageComplete = 100
		}
	} else {
		g.PercentageComplete = 0
	}

	// Update status
	if g.IsCompleted() && g.Status != GoalStatusCompleted {
		g.Status = GoalStatusCompleted
		now := time.Now()
		g.CompletedAt = &now
	} else if g.IsOverdue() && g.Status == GoalStatusActive {
		g.Status = GoalStatusOverdue
	}
}

// CalculateSuggestedContribution calculates suggested contribution amount
func (g *Goal) CalculateSuggestedContribution(frequency ContributionFrequency) float64 {
	if g.TargetDate == nil || g.RemainingAmount <= 0 {
		return 0
	}

	daysRemaining := g.DaysRemaining()
	if daysRemaining <= 0 {
		return g.RemainingAmount // Pay all at once if overdue
	}

	var periodsRemaining float64
	switch frequency {
	case FrequencyDaily:
		periodsRemaining = float64(daysRemaining)
	case FrequencyWeekly:
		periodsRemaining = float64(daysRemaining) / 7
	case FrequencyBiweekly:
		periodsRemaining = float64(daysRemaining) / 14
	case FrequencyMonthly:
		periodsRemaining = float64(daysRemaining) / 30
	case FrequencyQuarterly:
		periodsRemaining = float64(daysRemaining) / 90
	case FrequencyYearly:
		periodsRemaining = float64(daysRemaining) / 365
	default:
		return 0
	}

	if periodsRemaining <= 0 {
		return g.RemainingAmount
	}

	return g.RemainingAmount / periodsRemaining
}

// AddContribution adds a contribution to the goal
func (g *Goal) AddContribution(amount float64) {
	g.CurrentAmount += amount
	g.UpdateCalculatedFields()
}

// Milestone represents a milestone in goal progress
type Milestone struct {
	Percentage  float64    `json:"percentage"`
	Amount      float64    `json:"amount"`
	Description string     `json:"description"`
	Achieved    bool       `json:"achieved"`
	AchievedAt  *time.Time `json:"achieved_at,omitempty"`
}
