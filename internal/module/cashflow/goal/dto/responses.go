package dto

import (
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/service"
	"time"

	"github.com/google/uuid"
)

// GoalResponse represents a goal in API responses
type GoalResponse struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`

	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	Type        domain.GoalType     `json:"type"`
	Priority    domain.GoalPriority `json:"priority"`

	TargetAmount  float64 `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	Currency      string  `json:"currency"`

	StartDate   time.Time  `json:"start_date"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	PercentageComplete float64           `json:"percentage_complete"`
	RemainingAmount    float64           `json:"remaining_amount"`
	Status             domain.GoalStatus `json:"status"`
	DaysRemaining      int               `json:"days_remaining"`

	SuggestedContribution   *float64                      `json:"suggested_contribution,omitempty"`
	ContributionFrequency   *domain.ContributionFrequency `json:"contribution_frequency,omitempty"`
	AutoContribute          bool                          `json:"auto_contribute"`
	AutoContributeAmount    *float64                      `json:"auto_contribute_amount,omitempty"`
	AutoContributeAccountID *uuid.UUID                    `json:"auto_contribute_account_id,omitempty"`

	LinkedAccountID *uuid.UUID `json:"linked_account_id,omitempty"`

	EnableReminders    bool       `json:"enable_reminders"`
	ReminderFrequency  *string    `json:"reminder_frequency,omitempty"`
	LastReminderSentAt *time.Time `json:"last_reminder_sent_at,omitempty"`

	Milestones *string `json:"milestones,omitempty"`
	Notes      *string `json:"notes,omitempty"`
	Tags       *string `json:"tags,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// GoalSummaryResponse represents a goal summary in API responses
type GoalSummaryResponse struct {
	TotalGoals         int                             `json:"total_goals"`
	ActiveGoals        int                             `json:"active_goals"`
	CompletedGoals     int                             `json:"completed_goals"`
	OverdueGoals       int                             `json:"overdue_goals"`
	TotalTargetAmount  float64                         `json:"total_target_amount"`
	TotalCurrentAmount float64                         `json:"total_current_amount"`
	TotalRemaining     float64                         `json:"total_remaining"`
	AverageProgress    float64                         `json:"average_progress"`
	GoalsByType        map[string]*service.GoalTypeSum `json:"goals_by_type"`
	GoalsByPriority    map[string]int                  `json:"goals_by_priority"`
}

// ToGoalResponse converts a domain goal to response DTO
func ToGoalResponse(goal *domain.Goal) *GoalResponse {
	if goal == nil {
		return nil
	}

	return &GoalResponse{
		ID:                      goal.ID,
		UserID:                  goal.UserID,
		Name:                    goal.Name,
		Description:             goal.Description,
		Type:                    goal.Type,
		Priority:                goal.Priority,
		TargetAmount:            goal.TargetAmount,
		CurrentAmount:           goal.CurrentAmount,
		Currency:                goal.Currency,
		StartDate:               goal.StartDate,
		TargetDate:              goal.TargetDate,
		CompletedAt:             goal.CompletedAt,
		PercentageComplete:      goal.PercentageComplete,
		RemainingAmount:         goal.RemainingAmount,
		Status:                  goal.Status,
		DaysRemaining:           goal.DaysRemaining(),
		SuggestedContribution:   goal.SuggestedContribution,
		ContributionFrequency:   goal.ContributionFrequency,
		AutoContribute:          goal.AutoContribute,
		AutoContributeAmount:    goal.AutoContributeAmount,
		AutoContributeAccountID: goal.AutoContributeAccountID,
		LinkedAccountID:         goal.LinkedAccountID,
		EnableReminders:         goal.EnableReminders,
		ReminderFrequency:       goal.ReminderFrequency,
		LastReminderSentAt:      goal.LastReminderSentAt,
		Milestones:              goal.Milestones,
		Notes:                   goal.Notes,
		Tags:                    goal.Tags,
		CreatedAt:               goal.CreatedAt,
		UpdatedAt:               goal.UpdatedAt,
	}
}

// ToGoalResponseList converts a list of domain goals to response DTOs
func ToGoalResponseList(goals []domain.Goal) []*GoalResponse {
	responses := make([]*GoalResponse, len(goals))
	for i := range goals {
		responses[i] = ToGoalResponse(&goals[i])
	}
	return responses
}

// ToGoalSummaryResponse converts a service goal summary to response DTO
func ToGoalSummaryResponse(summary *service.GoalSummary) *GoalSummaryResponse {
	if summary == nil {
		return nil
	}

	return &GoalSummaryResponse{
		TotalGoals:         summary.TotalGoals,
		ActiveGoals:        summary.ActiveGoals,
		CompletedGoals:     summary.CompletedGoals,
		OverdueGoals:       summary.OverdueGoals,
		TotalTargetAmount:  summary.TotalTargetAmount,
		TotalCurrentAmount: summary.TotalCurrentAmount,
		TotalRemaining:     summary.TotalRemaining,
		AverageProgress:    summary.AverageProgress,
		GoalsByType:        summary.GoalsByType,
		GoalsByPriority:    summary.GoalsByPriority,
	}
}
