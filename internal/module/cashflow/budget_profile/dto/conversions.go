package dto

import (
	"encoding/json"
	"personalfinancedss/internal/module/cashflow/budget_profile/domain"
	"time"

	"github.com/google/uuid"
)

// FromCreateBudgetConstraintRequest converts create request to domain model
func FromCreateBudgetConstraintRequest(req CreateBudgetConstraintRequest, userID uuid.UUID) (*domain.BudgetConstraint, error) {
	// Parse category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, domain.ErrInvalidCategoryID
	}

	// Create new budget constraint
	bc, err := domain.NewBudgetConstraint(userID, categoryID, req.MinimumAmount, req.StartDate)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	if req.EndDate != nil {
		bc.EndDate = req.EndDate
	}

	// Set flexibility
	if req.IsFlexible != nil && *req.IsFlexible {
		maxAmount := float64(0)
		if req.MaximumAmount != nil {
			maxAmount = *req.MaximumAmount
		}
		if err := bc.SetFlexible(maxAmount); err != nil {
			return nil, err
		}
	}

	// Set priority
	if req.Priority != nil {
		if err := bc.SetPriority(*req.Priority); err != nil {
			return nil, err
		}
	}

	// Set description
	if req.Description != nil {
		bc.Description = *req.Description
	}

	// Set recurring
	if req.IsRecurring != nil {
		bc.IsRecurring = *req.IsRecurring
	}

	// Validate domain model
	if err := bc.Validate(); err != nil {
		return nil, err
	}

	return bc, nil
}

// ApplyUpdateBudgetConstraintRequest applies update request to create new version
func ApplyUpdateBudgetConstraintRequest(req UpdateBudgetConstraintRequest, existing *domain.BudgetConstraint) (*domain.BudgetConstraint, error) {
	// Create new version from existing
	newVersion := existing.CreateNewVersion()

	// Apply updates
	if req.MinimumAmount != nil {
		if err := newVersion.UpdateMinimum(*req.MinimumAmount); err != nil {
			return nil, err
		}
	}

	// Update flexibility if provided
	if req.IsFlexible != nil {
		if *req.IsFlexible {
			maxAmount := newVersion.MaximumAmount
			if req.MaximumAmount != nil {
				maxAmount = *req.MaximumAmount
			}
			if err := newVersion.SetFlexible(maxAmount); err != nil {
				return nil, err
			}
		} else {
			newVersion.SetFixed()
		}
	} else if req.MaximumAmount != nil {
		// Update maximum amount only if flexible
		if newVersion.IsFlexible {
			if err := newVersion.SetFlexible(*req.MaximumAmount); err != nil {
				return nil, err
			}
		}
	}

	// Update priority if provided
	if req.Priority != nil {
		if err := newVersion.SetPriority(*req.Priority); err != nil {
			return nil, err
		}
	}

	// Update end date
	if req.EndDate != nil {
		newVersion.EndDate = req.EndDate
	}

	// Update description
	if req.Description != nil {
		newVersion.Description = *req.Description
	}

	// Validate updated model
	if err := newVersion.Validate(); err != nil {
		return nil, err
	}

	return newVersion, nil
}

// ToBudgetConstraintResponse converts domain model to response
func ToBudgetConstraintResponse(bc *domain.BudgetConstraint, includeDetails bool) BudgetConstraintResponse {
	// Calculate duration
	duration := 0
	if !bc.StartDate.IsZero() {
		end := time.Now()
		if bc.EndDate != nil {
			end = *bc.EndDate
		}
		duration = int(end.Sub(bc.StartDate).Hours() / 24)
	}

	response := BudgetConstraintResponse{
		ID:            bc.ID.String(),
		UserID:        bc.UserID.String(),
		CategoryID:    bc.CategoryID.String(),
		StartDate:     &bc.StartDate,
		EndDate:       bc.EndDate,
		Duration:      duration,
		MinimumAmount: bc.MinimumAmount,
		IsFlexible:    bc.IsFlexible,
		MaximumAmount: bc.MaximumAmount,
		Priority:      bc.Priority,
		Status:        string(bc.Status),
		IsRecurring:   bc.IsRecurring,
		IsActive:      bc.IsActive(),
		IsArchived:    bc.IsArchived(),
		Description:   bc.Description,
		CreatedAt:     bc.CreatedAt,
		UpdatedAt:     bc.UpdatedAt,
		ArchivedAt:    bc.ArchivedAt,
	}

	// Include details if requested
	if includeDetails {
		response.FlexibilityRange = bc.GetFlexibilityRange()
		response.DisplayString = bc.String()
	}

	// Parse tags
	if len(bc.Tags) > 0 {
		var tags []string
		if err := json.Unmarshal(bc.Tags, &tags); err == nil {
			response.Tags = tags
		}
	}

	// Previous version ID
	if bc.PreviousVersionID != nil {
		prevID := bc.PreviousVersionID.String()
		response.PreviousVersionID = &prevID
	}

	return response
}

// ToBudgetConstraintWithHistoryResponse converts domain model with version history
func ToBudgetConstraintWithHistoryResponse(current *domain.BudgetConstraint, history domain.BudgetConstraints, includeDetails bool) BudgetConstraintWithHistoryResponse {
	response := BudgetConstraintWithHistoryResponse{
		Current: ToBudgetConstraintResponse(current, includeDetails),
	}

	if len(history) > 0 {
		response.VersionHistory = make([]BudgetConstraintResponse, 0, len(history))
		for _, h := range history {
			response.VersionHistory = append(response.VersionHistory, ToBudgetConstraintResponse(h, includeDetails))
		}
	}

	return response
}

// ToBudgetConstraintListResponse converts list of domain models to list response
func ToBudgetConstraintListResponse(constraints domain.BudgetConstraints, includeDetails, includeSummary bool) BudgetConstraintListResponse {
	responses := make([]BudgetConstraintResponse, 0, len(constraints))
	for _, bc := range constraints {
		responses = append(responses, ToBudgetConstraintResponse(bc, includeDetails))
	}

	result := BudgetConstraintListResponse{
		BudgetConstraints: responses,
		Count:             len(responses),
	}

	// Add summary if requested
	if includeSummary {
		result.Summary = calculateSummary(constraints)
	}

	return result
}

// calculateSummary calculates summary statistics for budget constraints
func calculateSummary(constraints domain.BudgetConstraints) *BudgetConstraintSummaryResponse {
	summary := &BudgetConstraintSummaryResponse{
		TotalMandatoryExpenses: constraints.TotalMandatoryExpenses(),
		TotalFlexible:          len(constraints.GetFlexible()),
		TotalFixed:             len(constraints.GetFixed()),
		Count:                  len(constraints),
	}

	// Count active constraints
	activeCount := 0
	for _, bc := range constraints {
		if bc.IsActive() {
			activeCount++
		}
	}
	summary.ActiveCount = activeCount

	return summary
}
