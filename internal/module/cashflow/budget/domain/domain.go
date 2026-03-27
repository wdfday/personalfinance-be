package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Budget represents a spending budget for a category or account
type Budget struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Budget Details
	Name        string  `gorm:"type:varchar(255);not null;column:name" json:"name"`
	Description *string `gorm:"type:text;column:description" json:"description,omitempty"`

	// Budget Amount
	Amount   float64 `gorm:"type:decimal(15,2);not null;column:amount" json:"amount"`
	Currency string  `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`

	// Period Configuration
	Period    BudgetPeriod `gorm:"type:varchar(20);not null;column:period" json:"period"`
	StartDate time.Time    `gorm:"type:date;not null;column:start_date" json:"start_date"`
	EndDate   *time.Time   `gorm:"type:date;column:end_date" json:"end_date,omitempty"` // null for recurring budgets

	// Scope - category based
	CategoryID   *uuid.UUID `gorm:"type:uuid;index;column:category_id" json:"category_id,omitempty"`     // null means all categories
	ConstraintID *uuid.UUID `gorm:"type:uuid;index;column:constraint_id" json:"constraint_id,omitempty"` // FK to budget_constraint (if created from DSS)

	// Tracking
	SpentAmount      float64      `gorm:"type:decimal(15,2);default:0;column:spent_amount" json:"spent_amount"`
	RemainingAmount  float64      `gorm:"type:decimal(15,2);default:0;column:remaining_amount" json:"remaining_amount"`
	PercentageSpent  float64      `gorm:"type:decimal(5,2);default:0;column:percentage_spent" json:"percentage_spent"`
	Status           BudgetStatus `gorm:"type:varchar(20);default:'active';column:status" json:"status"`
	LastCalculatedAt *time.Time   `gorm:"column:last_calculated_at" json:"last_calculated_at,omitempty"`

	// Alert Settings
	EnableAlerts     bool                `gorm:"default:true;column:enable_alerts" json:"enable_alerts"`
	AlertThresholds  AlertThresholdsJSON `gorm:"type:jsonb;column:alert_thresholds" json:"alert_thresholds"` // e.g. ["50","75","90","100"]
	AlertedAt        *string             `gorm:"type:jsonb;column:alerted_at" json:"alerted_at,omitempty"`   // JSON map of threshold -> timestamp
	NotificationSent bool                `gorm:"default:false;column:notification_sent" json:"notification_sent"`

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Budget
func (Budget) TableName() string {
	return "budgets"
}

// IsExpired checks if the budget has expired
func (b *Budget) IsExpired() bool {
	if b.EndDate == nil {
		return false // Recurring budgets don't expire
	}
	return time.Now().After(*b.EndDate)
}

// IsExceeded checks if the budget amount has been exceeded
func (b *Budget) IsExceeded() bool {
	return b.SpentAmount > b.Amount
}

// ShouldAlert checks if an alert should be sent for the given threshold
func (b *Budget) ShouldAlert(threshold AlertThreshold) bool {
	if !b.EnableAlerts {
		return false
	}

	for _, t := range b.AlertThresholds {
		if t == threshold {
			return true
		}
	}
	return false
}

// UpdateCalculatedFields updates spent, remaining, and percentage fields
func (b *Budget) UpdateCalculatedFields() {
	b.RemainingAmount = b.Amount - b.SpentAmount
	if b.Amount > 0 {
		b.PercentageSpent = (b.SpentAmount / b.Amount) * 100
	} else {
		b.PercentageSpent = 0
	}

	// Update status based on spending
	if b.IsExpired() {
		b.Status = BudgetStatusExpired
	} else if b.IsExceeded() {
		b.Status = BudgetStatusExceeded
	} else if b.PercentageSpent >= 80 {
		b.Status = BudgetStatusWarning
	} else {
		b.Status = BudgetStatusActive
	}

	now := time.Now()
	b.LastCalculatedAt = &now
}

// IsActive checks if the budget is active
func (b *Budget) IsActive() bool {
	return b.Status == BudgetStatusActive || b.Status == BudgetStatusWarning
}

// BelongsTo checks if the budget belongs to the given user
func (b *Budget) BelongsTo(userID uuid.UUID) bool {
	return b.UserID == userID
}
