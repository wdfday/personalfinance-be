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

// ApplyTo applies the update request fields to the budget domain object
func (req *UpdateBudgetRequest) ApplyTo(budget *domain.Budget) {
	if req.Name != nil {
		budget.Name = *req.Name
	}
	if req.Description != nil {
		budget.Description = req.Description
	}
	if req.Amount != nil {
		budget.Amount = *req.Amount
	}
	if req.Currency != nil {
		budget.Currency = *req.Currency
	}
	if req.Period != nil {
		budget.Period = *req.Period
	}
	if req.StartDate != nil {
		budget.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		budget.EndDate = req.EndDate
	}
	if req.CategoryID != nil {
		budget.CategoryID = req.CategoryID
	}
	if req.AccountID != nil {
		budget.AccountID = req.AccountID
	}
	if req.EnableAlerts != nil {
		budget.EnableAlerts = *req.EnableAlerts
	}
	if len(req.AlertThresholds) > 0 {
		budget.AlertThresholds = req.AlertThresholds
	}
	if req.AllowRollover != nil {
		budget.AllowRollover = *req.AllowRollover
	}
	if req.CarryOverPercent != nil {
		budget.CarryOverPercent = req.CarryOverPercent
	}
	if req.AutoAdjust != nil {
		budget.AutoAdjust = *req.AutoAdjust
	}
	if req.AutoAdjustPercentage != nil {
		budget.AutoAdjustPercentage = req.AutoAdjustPercentage
	}
	if req.AutoAdjustBasedOn != nil {
		budget.AutoAdjustBasedOn = req.AutoAdjustBasedOn
	}
	if req.Status != nil {
		budget.Status = *req.Status
	}
}
