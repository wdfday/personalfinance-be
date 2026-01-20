package domain

import (
	"time"

	"github.com/google/uuid"
)

// ContributionType represents the type of contribution transaction
type ContributionType string

const (
	// ContributionTypeDeposit - money contributed to goal (increases goal amount, decreases available balance)
	ContributionTypeDeposit ContributionType = "deposit"
	// ContributionTypeWithdrawal - money withdrawn from goal (decreases goal amount, increases available balance)
	ContributionTypeWithdrawal ContributionType = "withdrawal"
)

// IsValid checks if the contribution type is valid
func (ct ContributionType) IsValid() bool {
	switch ct {
	case ContributionTypeDeposit, ContributionTypeWithdrawal:
		return true
	}
	return false
}

// GoalContribution represents a contribution or withdrawal transaction for a goal
// These are virtual transactions that affect AvailableBalance but not CurrentBalance
type GoalContribution struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	GoalID    uuid.UUID `gorm:"type:uuid;not null;index;column:goal_id" json:"goal_id"`
	AccountID uuid.UUID `gorm:"type:uuid;not null;index;column:account_id" json:"account_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Transaction details
	Type     ContributionType `gorm:"type:varchar(20);not null;column:type" json:"type"`
	Amount   float64          `gorm:"type:decimal(15,2);not null;column:amount" json:"amount"` // Always positive
	Currency string           `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`

	// Optional metadata
	Note   *string `gorm:"type:text;column:note" json:"note,omitempty"`
	Source string  `gorm:"type:varchar(20);default:'manual';column:source" json:"source"` // manual, auto, import

	// For withdrawals - optional reference to the contribution being reversed
	ReversingContributionID *uuid.UUID `gorm:"type:uuid;column:reversing_contribution_id" json:"reversing_contribution_id,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" json:"created_at"`
}

// TableName returns the table name for GORM
func (GoalContribution) TableName() string {
	return "goal_contributions"
}

// NetAmount returns the effective amount (positive for deposit, negative for withdrawal)
func (gc *GoalContribution) NetAmount() float64 {
	if gc.Type == ContributionTypeWithdrawal {
		return -gc.Amount
	}
	return gc.Amount
}

// NewDeposit creates a new deposit contribution
func NewDeposit(goalID, accountID, userID uuid.UUID, amount float64, note *string) *GoalContribution {
	return &GoalContribution{
		GoalID:    goalID,
		AccountID: accountID,
		UserID:    userID,
		Type:      ContributionTypeDeposit,
		Amount:    amount,
		Currency:  "VND",
		Note:      note,
		Source:    "manual",
	}
}

// NewWithdrawal creates a new withdrawal contribution
func NewWithdrawal(goalID, accountID, userID uuid.UUID, amount float64, note *string, reversingID *uuid.UUID) *GoalContribution {
	return &GoalContribution{
		GoalID:                  goalID,
		AccountID:               accountID,
		UserID:                  userID,
		Type:                    ContributionTypeWithdrawal,
		Amount:                  amount,
		Currency:                "VND",
		Note:                    note,
		Source:                  "manual",
		ReversingContributionID: reversingID,
	}
}
