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
	Behavior    domain.GoalBehavior `json:"behavior" binding:"required"`
	Category    domain.GoalCategory `json:"category" binding:"required"`
	Priority    domain.GoalPriority `json:"priority" binding:"required"`
	Status      *domain.GoalStatus  `json:"status,omitempty"`

	TargetAmount  float64  `json:"targetAmount" binding:"required,gt=0"`
	CurrentAmount *float64 `json:"currentAmount,omitempty"`
	Currency      string   `json:"currency" binding:"required,len=3"`

	StartDate  time.Time  `json:"startDate" binding:"required"`
	TargetDate *time.Time `json:"targetDate"` // Required for 'willing' behavior

	ContributionFrequency   *domain.ContributionFrequency `json:"contributionFrequency"` // Required for 'recurring' behavior
	AutoContribute          bool                          `json:"autoContribute"`
	AutoContributeAmount    *float64                      `json:"autoContributeAmount"`
	AutoContributeAccountID *uuid.UUID                    `json:"autoContributeAccountId"`

	AccountID uuid.UUID `json:"accountId" binding:"required"` // Required, must be cash/bank/savings account

	EnableReminders   bool    `json:"enableReminders"`
	ReminderFrequency *string `json:"reminderFrequency"`

	Notes *string `json:"notes"`
	Tags  *string `json:"tags"`
}

// UpdateGoalRequest represents a request to update an existing goal
type UpdateGoalRequest struct {
	Name        *string              `json:"name"`
	Description *string              `json:"description"`
	Category    *domain.GoalCategory `json:"category"`
	Priority    *domain.GoalPriority `json:"priority"`

	TargetAmount *float64 `json:"targetAmount" binding:"omitempty,gt=0"`
	Currency     *string  `json:"currency" binding:"omitempty,len=3"`

	StartDate  *time.Time `json:"startDate"`
	TargetDate *time.Time `json:"targetDate"`

	ContributionFrequency   *domain.ContributionFrequency `json:"contributionFrequency"`
	AutoContribute          *bool                         `json:"autoContribute"`
	AutoContributeAmount    *float64                      `json:"autoContributeAmount"`
	AutoContributeAccountID *uuid.UUID                    `json:"autoContributeAccountId"`

	AccountID *uuid.UUID `json:"accountId"`

	EnableReminders   *bool   `json:"enableReminders"`
	ReminderFrequency *string `json:"reminderFrequency"`

	Status *domain.GoalStatus `json:"status"`

	Notes *string `json:"notes"`
	Tags  *string `json:"tags"`
}

// AddContributionRequest represents a request to add a contribution to a goal (deposit)
type AddContributionRequest struct {
	Amount    float64    `json:"amount" binding:"required,gt=0"`
	AccountID *uuid.UUID `json:"accountId"` // Optional: if not provided, uses goal's accountId
	Note      *string    `json:"note"`
	Source    *string    `json:"source"` // manual, auto, import (default: manual)
}

// WithdrawContributionRequest represents a request to withdraw from a goal's contributions
type WithdrawContributionRequest struct {
	Amount                  float64    `json:"amount" binding:"required,gt=0"`
	Note                    *string    `json:"note"`
	ReversingContributionID *uuid.UUID `json:"reversingContributionId"` // Optional: reference to the original contribution
}

// ApplyTo applies the update request fields to the goal domain object
func (req *UpdateGoalRequest) ApplyTo(goal *domain.Goal) {
	if req.Name != nil {
		goal.Name = *req.Name
	}
	if req.Description != nil {
		goal.Description = req.Description
	}
	if req.Category != nil {
		goal.Category = *req.Category
	}
	if req.Priority != nil {
		goal.Priority = *req.Priority
	}
	if req.TargetAmount != nil {
		goal.TargetAmount = *req.TargetAmount
	}
	if req.Currency != nil {
		goal.Currency = *req.Currency
	}
	if req.StartDate != nil {
		goal.StartDate = *req.StartDate
	}
	if req.TargetDate != nil {
		goal.TargetDate = req.TargetDate
	}
	if req.ContributionFrequency != nil {
		goal.ContributionFrequency = req.ContributionFrequency
	}
	if req.AutoContribute != nil {
		goal.AutoContribute = *req.AutoContribute
	}
	if req.AutoContributeAmount != nil {
		goal.AutoContributeAmount = req.AutoContributeAmount
	}
	if req.AutoContributeAccountID != nil {
		goal.AutoContributeAccountID = req.AutoContributeAccountID
	}
	if req.AccountID != nil {
		goal.AccountID = *req.AccountID
	}
	if req.EnableReminders != nil {
		goal.EnableReminders = *req.EnableReminders
	}
	if req.ReminderFrequency != nil {
		goal.ReminderFrequency = req.ReminderFrequency
	}
	if req.Status != nil {
		goal.Status = *req.Status
	}
	if req.Notes != nil {
		goal.Notes = req.Notes
	}
	if req.Tags != nil {
		goal.Tags = req.Tags
	}
}
