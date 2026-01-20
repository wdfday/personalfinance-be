package domain

import "errors"

// Domain errors for budget module
var (
	ErrBudgetNotFound           = errors.New("budget not found")
	ErrBudgetAlreadyExists      = errors.New("budget already exists for this category/account")
	ErrInvalidBudgetID          = errors.New("invalid budget ID")
	ErrInvalidUserID            = errors.New("invalid user ID")
	ErrInvalidAmount            = errors.New("budget amount must be greater than 0")
	ErrInvalidName              = errors.New("budget name is required")
	ErrInvalidPeriod            = errors.New("invalid budget period")
	ErrInvalidStartDate         = errors.New("start date is required")
	ErrEndDateBeforeStartDate   = errors.New("end date must be after start date")
	ErrBudgetExpired            = errors.New("budget has expired")
	ErrBudgetNotActive          = errors.New("budget is not active")
	ErrUnauthorizedAccess       = errors.New("unauthorized access to budget")
	ErrInvalidCurrency          = errors.New("invalid currency code")
	ErrInvalidAlertThreshold    = errors.New("invalid alert threshold")
	ErrInvalidCarryOverPercent  = errors.New("carry over percent must be between 0 and 100")
	ErrInvalidAutoAdjustPercent = errors.New("auto adjust percent must be between 0 and 100")
)

// IsBudgetNotFound checks if the error is a budget not found error
func IsBudgetNotFound(err error) bool {
	return errors.Is(err, ErrBudgetNotFound)
}

// IsUnauthorizedAccess checks if the error is an unauthorized access error
func IsUnauthorizedAccess(err error) bool {
	return errors.Is(err, ErrUnauthorizedAccess)
}
