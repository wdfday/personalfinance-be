package dto

import (
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"time"

	"github.com/google/uuid"
)

// CreateBudgetRequest represents a request to create a new budget
type CreateBudgetRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`

	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required,len=3"`

	Period    domain.BudgetPeriod `json:"period" binding:"required"`
	StartDate time.Time           `json:"start_date" binding:"required"`
	EndDate   *time.Time          `json:"end_date"`

	CategoryID *uuid.UUID `json:"category_id"`
	AccountID  *uuid.UUID `json:"account_id"`

	EnableAlerts    bool                    `json:"enable_alerts"`
	AlertThresholds []domain.AlertThreshold `json:"alert_thresholds"`

	AllowRollover    bool `json:"allow_rollover"`
	CarryOverPercent *int `json:"carry_over_percent" binding:"omitempty,min=0,max=100"`

	AutoAdjust           bool    `json:"auto_adjust"`
	AutoAdjustPercentage *int    `json:"auto_adjust_percentage" binding:"omitempty,min=0,max=100"`
	AutoAdjustBasedOn    *string `json:"auto_adjust_based_on"`
}

// UpdateBudgetRequest represents a request to update an existing budget
type UpdateBudgetRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`

	Amount   *float64 `json:"amount" binding:"omitempty,gt=0"`
	Currency *string  `json:"currency" binding:"omitempty,len=3"`

	Period    *domain.BudgetPeriod `json:"period"`
	StartDate *time.Time           `json:"start_date"`
	EndDate   *time.Time           `json:"end_date"`

	CategoryID *uuid.UUID `json:"category_id"`
	AccountID  *uuid.UUID `json:"account_id"`

	EnableAlerts    *bool                   `json:"enable_alerts"`
	AlertThresholds []domain.AlertThreshold `json:"alert_thresholds"`

	AllowRollover    *bool `json:"allow_rollover"`
	CarryOverPercent *int  `json:"carry_over_percent" binding:"omitempty,min=0,max=100"`

	AutoAdjust           *bool   `json:"auto_adjust"`
	AutoAdjustPercentage *int    `json:"auto_adjust_percentage" binding:"omitempty,min=0,max=100"`
	AutoAdjustBasedOn    *string `json:"auto_adjust_based_on"`

	Status *domain.BudgetStatus `json:"status"`
}

// BudgetFilterRequest represents filters for listing budgets
type BudgetFilterRequest struct {
	Status     *domain.BudgetStatus `form:"status"`
	Period     *domain.BudgetPeriod `form:"period"`
	CategoryID *uuid.UUID           `form:"category_id"`
	AccountID  *uuid.UUID           `form:"account_id"`
	StartDate  *time.Time           `form:"start_date"`
	EndDate    *time.Time           `form:"end_date"`
}
