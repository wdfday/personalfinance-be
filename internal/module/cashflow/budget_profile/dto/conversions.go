package dto

import (
	"personalfinancedss/internal/module/cashflow/budget_profile/domain"

	"github.com/google/uuid"
)

// FromCreateBudgetConstraintRequest converts create request to domain model
func FromCreateBudgetConstraintRequest(req CreateBudgetConstraintRequest, userID uuid.UUID) (*domain.BudgetConstraint, error) {
	// Parse category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, domain.ErrInvalidCategoryID
	}

	bc := &domain.BudgetConstraint{
		ID:            uuid.New(),
		UserID:        userID,
		CategoryID:    categoryID,
		MinimumAmount: req.MinimumAmount,
	}

	// Set flexibility
	if req.IsFlexible != nil && *req.IsFlexible {
		maxAmount := float64(0)
		if req.MaximumAmount != nil {
			maxAmount = *req.MaximumAmount
		}
		bc.IsFlexible = true
		bc.MaximumAmount = maxAmount
	}

	// Set priority
	if req.Priority != nil {
		bc.Priority = *req.Priority
	} else {
		bc.Priority = 99 // Default priority
	}

	// Validate domain model
	if err := bc.Validate(); err != nil {
		return nil, err
	}

	return bc, nil
}

// FromUpdateBudgetConstraintRequest applies update request to existing domain model
func FromUpdateBudgetConstraintRequest(req UpdateBudgetConstraintRequest, existing *domain.BudgetConstraint) error {
	// Update minimum amount if provided
	if req.MinimumAmount != nil {
		if err := existing.UpdateMinimum(*req.MinimumAmount); err != nil {
			return err
		}
	}

	// Update flexibility if provided
	if req.IsFlexible != nil {
		if *req.IsFlexible {
			maxAmount := existing.MaximumAmount
			if req.MaximumAmount != nil {
				maxAmount = *req.MaximumAmount
			}
			if err := existing.SetFlexible(maxAmount); err != nil {
				return err
			}
		} else {
			existing.SetFixed()
		}
	} else if req.MaximumAmount != nil {
		// Update maximum amount only if flexible
		if existing.IsFlexible {
			if err := existing.SetFlexible(*req.MaximumAmount); err != nil {
				return err
			}
		}
	}

	// Update priority if provided
	if req.Priority != nil {
		if err := existing.SetPriority(*req.Priority); err != nil {
			return err
		}
	}

	// Validate updated model
	return existing.Validate()
}

// ToBudgetConstraintResponse converts domain model to response
func ToBudgetConstraintResponse(bc *domain.BudgetConstraint, includeDetails bool) BudgetConstraintResponse {
	response := BudgetConstraintResponse{
		ID:            bc.ID.String(),
		UserID:        bc.UserID.String(),
		CategoryID:    bc.CategoryID.String(),
		MinimumAmount: bc.MinimumAmount,
		IsFlexible:    bc.IsFlexible,
		MaximumAmount: bc.MaximumAmount,
		Priority:      bc.Priority,
		CreatedAt:     bc.CreatedAt,
		UpdatedAt:     bc.UpdatedAt,
	}

	// Include details if requested
	if includeDetails {
		response.FlexibilityRange = bc.GetFlexibilityRange()
		response.DisplayString = bc.String()
	}

	return response
}

// ToBudgetConstraintListResponse converts list of domain models to list response
func ToBudgetConstraintListResponse(constraints domain.BudgetConstraints, includeDetails bool) BudgetConstraintListResponse {
	responses := make([]BudgetConstraintResponse, 0, len(constraints))
	for _, bc := range constraints {
		responses = append(responses, ToBudgetConstraintResponse(bc, includeDetails))
	}

	return BudgetConstraintListResponse{
		BudgetConstraints: responses,
		Count:             len(responses),
	}
}

// ToBudgetConstraintSummaryResponse converts budget constraints to summary response
func ToBudgetConstraintSummaryResponse(constraints domain.BudgetConstraints) BudgetConstraintSummaryResponse {
	return BudgetConstraintSummaryResponse{
		TotalMandatoryExpenses: constraints.TotalMandatoryExpenses(),
		TotalFlexible:          len(constraints.GetFlexible()),
		TotalFixed:             len(constraints.GetFixed()),
		Count:                  len(constraints),
	}
}
