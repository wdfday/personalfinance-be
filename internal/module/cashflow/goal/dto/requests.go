package dto

import (
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
)

// CreateGoalRequest represents a request to create a new goal
type CreateGoalRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description *string             `json:"description"`
	Type        domain.GoalType     `json:"type" binding:"required"`
	Priority    domain.GoalPriority `json:"priority" binding:"required"`

	TargetAmount float64 `json:"target_amount" binding:"required,gt=0"`
	Currency     string  `json:"currency" binding:"required,len=3"`

	StartDate  time.Time  `json:"start_date" binding:"required"`
	TargetDate *time.Time `json:"target_date"`

	ContributionFrequency   *domain.ContributionFrequency `json:"contribution_frequency"`
	AutoContribute          bool                          `json:"auto_contribute"`
	AutoContributeAmount    *float64                      `json:"auto_contribute_amount"`
	AutoContributeAccountID *uuid.UUID                    `json:"auto_contribute_account_id"`

	LinkedAccountID *uuid.UUID `json:"linked_account_id"`

	EnableReminders   bool    `json:"enable_reminders"`
	ReminderFrequency *string `json:"reminder_frequency"`

	Notes *string `json:"notes"`
	Tags  *string `json:"tags"`
}

// UpdateGoalRequest represents a request to update an existing goal
type UpdateGoalRequest struct {
	Name        *string              `json:"name"`
	Description *string              `json:"description"`
	Type        *domain.GoalType     `json:"type"`
	Priority    *domain.GoalPriority `json:"priority"`

	TargetAmount *float64 `json:"target_amount" binding:"omitempty,gt=0"`
	Currency     *string  `json:"currency" binding:"omitempty,len=3"`

	StartDate  *time.Time `json:"start_date"`
	TargetDate *time.Time `json:"target_date"`

	ContributionFrequency   *domain.ContributionFrequency `json:"contribution_frequency"`
	AutoContribute          *bool                         `json:"auto_contribute"`
	AutoContributeAmount    *float64                      `json:"auto_contribute_amount"`
	AutoContributeAccountID *uuid.UUID                    `json:"auto_contribute_account_id"`

	LinkedAccountID *uuid.UUID `json:"linked_account_id"`

	EnableReminders   *bool   `json:"enable_reminders"`
	ReminderFrequency *string `json:"reminder_frequency"`

	Status *domain.GoalStatus `json:"status"`

	Notes *string `json:"notes"`
	Tags  *string `json:"tags"`
}

// AddContributionRequest represents a request to add a contribution to a goal
type AddContributionRequest struct {
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Description *string    `json:"description"`
	Date        *time.Time `json:"date"`
}
