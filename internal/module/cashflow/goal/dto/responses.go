package dto

import (
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"time"

	"github.com/google/uuid"
)

// GoalResponse represents a goal in API responses
type GoalResponse struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"userId"`

	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	Behavior    domain.GoalBehavior `json:"behavior"`
	Category    domain.GoalCategory `json:"category"`
	Priority    domain.GoalPriority `json:"priority"`

	TargetAmount  float64 `json:"targetAmount"`
	CurrentAmount float64 `json:"currentAmount"`
	Currency      string  `json:"currency"`

	StartDate   time.Time  `json:"startDate"`
	TargetDate  *time.Time `json:"targetDate,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	PercentageComplete float64           `json:"percentageComplete"`
	RemainingAmount    float64           `json:"remainingAmount"`
	Status             domain.GoalStatus `json:"status"`
	DaysRemaining      int               `json:"daysRemaining"`

	SuggestedContribution   *float64                      `json:"suggestedContribution,omitempty"`
	ContributionFrequency   *domain.ContributionFrequency `json:"contributionFrequency,omitempty"`
	AutoContribute          bool                          `json:"autoContribute"`
	AutoContributeAmount    *float64                      `json:"autoContributeAmount,omitempty"`
	AutoContributeAccountID *uuid.UUID                    `json:"autoContributeAccountId,omitempty"`

	AccountID         uuid.UUID  `json:"accountId"`
	ConvertedBudgetID *uuid.UUID `json:"convertedBudgetId,omitempty"`

	EnableReminders    bool       `json:"enableReminders"`
	ReminderFrequency  *string    `json:"reminderFrequency,omitempty"`
	LastReminderSentAt *time.Time `json:"lastReminderSentAt,omitempty"`

	Milestones *string `json:"milestones,omitempty"`
	Notes      *string `json:"notes,omitempty"`
	Tags       *string `json:"tags,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
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
		Behavior:                goal.Behavior,
		Category:                goal.Category,
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
		AccountID:               goal.AccountID,
		ConvertedBudgetID:       goal.ConvertedBudgetID,
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

// GoalSummaryResponse represents a goal summary in API responses
type GoalSummaryResponse struct {
	TotalGoals         int                                `json:"totalGoals"`
	ActiveGoals        int                                `json:"activeGoals"`
	CompletedGoals     int                                `json:"completedGoals"`
	OverdueGoals       int                                `json:"overdueGoals"`
	TotalTargetAmount  float64                            `json:"totalTargetAmount"`
	TotalCurrentAmount float64                            `json:"totalCurrentAmount"`
	TotalRemaining     float64                            `json:"totalRemaining"`
	AverageProgress    float64                            `json:"averageProgress"`
	GoalsByCategory    map[string]*domain.GoalCategorySum `json:"goalsByCategory"`
	GoalsByPriority    map[string]int                     `json:"goalsByPriority"`
}

// ToGoalSummaryResponse converts a service goal summary to response DTO
func ToGoalSummaryResponse(summary *domain.GoalSummary) *GoalSummaryResponse {
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
		GoalsByCategory:    summary.GoalsByCategory,
		GoalsByPriority:    summary.GoalsByPriority,
	}
}

// ContributionResponse represents a goal contribution in API responses
type ContributionResponse struct {
	ID        uuid.UUID `json:"id"`
	GoalID    uuid.UUID `json:"goalId"`
	AccountID uuid.UUID `json:"accountId"`
	UserID    uuid.UUID `json:"userId"`

	Type     domain.ContributionType `json:"type"`
	Amount   float64                 `json:"amount"`
	Currency string                  `json:"currency"`

	Note   *string `json:"note,omitempty"`
	Source string  `json:"source"`

	ReversingContributionID *uuid.UUID `json:"reversingContributionId,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

// ContributionListResponse represents a list of contributions with summary
type ContributionListResponse struct {
	Contributions    []*ContributionResponse `json:"contributions"`
	TotalDeposits    float64                 `json:"totalDeposits"`
	TotalWithdrawals float64                 `json:"totalWithdrawals"`
	NetAmount        float64                 `json:"netAmount"`
}

// ToContributionResponse converts a domain contribution to response DTO
func ToContributionResponse(contribution *domain.GoalContribution) *ContributionResponse {
	if contribution == nil {
		return nil
	}

	return &ContributionResponse{
		ID:                      contribution.ID,
		GoalID:                  contribution.GoalID,
		AccountID:               contribution.AccountID,
		UserID:                  contribution.UserID,
		Type:                    contribution.Type,
		Amount:                  contribution.Amount,
		Currency:                contribution.Currency,
		Note:                    contribution.Note,
		Source:                  contribution.Source,
		ReversingContributionID: contribution.ReversingContributionID,
		CreatedAt:               contribution.CreatedAt,
	}
}

// ToContributionResponseList converts a list of domain contributions to response DTOs
func ToContributionResponseList(contributions []domain.GoalContribution) *ContributionListResponse {
	responses := make([]*ContributionResponse, len(contributions))
	var totalDeposits, totalWithdrawals float64

	for i := range contributions {
		responses[i] = ToContributionResponse(&contributions[i])
		if contributions[i].Type == domain.ContributionTypeDeposit {
			totalDeposits += contributions[i].Amount
		} else {
			totalWithdrawals += contributions[i].Amount
		}
	}

	return &ContributionListResponse{
		Contributions:    responses,
		TotalDeposits:    totalDeposits,
		TotalWithdrawals: totalWithdrawals,
		NetAmount:        totalDeposits - totalWithdrawals,
	}
}
