package dto

import (
	"personalfinancedss/internal/module/cashflow/income_profile/domain"

	"github.com/google/uuid"
)

// FromCreateIncomeProfileRequest converts create request to domain model
func FromCreateIncomeProfileRequest(req CreateIncomeProfileRequest, userID uuid.UUID) (*domain.IncomeProfile, error) {
	ip := &domain.IncomeProfile{
		ID:     uuid.New(),
		UserID: userID,
		Year:   req.Year,
		Month:  req.Month,
	}

	// Set income components (default to 0 if not provided)
	if req.BaseSalary != nil {
		ip.BaseSalary = *req.BaseSalary
	}
	if req.Bonus != nil {
		ip.Bonus = *req.Bonus
	}
	if req.FreelanceIncome != nil {
		ip.FreelanceIncome = *req.FreelanceIncome
	}
	if req.OtherIncome != nil {
		ip.OtherIncome = *req.OtherIncome
	}

	// Set status
	if req.IsActual != nil {
		ip.IsActual = *req.IsActual
	}
	if req.Notes != nil {
		ip.Notes = *req.Notes
	}

	// Validate domain model
	if err := ip.Validate(); err != nil {
		return nil, err
	}

	return ip, nil
}

// FromUpdateIncomeProfileRequest applies update request to existing domain model
func FromUpdateIncomeProfileRequest(req UpdateIncomeProfileRequest, existing *domain.IncomeProfile) error {
	// Update income components if provided
	if req.BaseSalary != nil {
		existing.BaseSalary = *req.BaseSalary
	}
	if req.Bonus != nil {
		if err := existing.AddBonus(*req.Bonus); err != nil {
			return err
		}
	}
	if req.FreelanceIncome != nil {
		if err := existing.AddFreelanceIncome(*req.FreelanceIncome); err != nil {
			return err
		}
	}
	if req.OtherIncome != nil {
		if err := existing.AddOtherIncome(*req.OtherIncome); err != nil {
			return err
		}
	}

	// Update status if provided
	if req.IsActual != nil {
		if *req.IsActual {
			existing.MarkAsActual()
		} else {
			existing.MarkAsProjected()
		}
	}
	if req.Notes != nil {
		existing.UpdateNotes(*req.Notes)
	}

	// Validate updated model
	return existing.Validate()
}

// ToIncomeProfileResponse converts domain model to response
func ToIncomeProfileResponse(ip *domain.IncomeProfile, includeBreakdown bool) IncomeProfileResponse {
	response := IncomeProfileResponse{
		ID:              ip.ID.String(),
		UserID:          ip.UserID.String(),
		Year:            ip.Year,
		Month:           ip.Month,
		BaseSalary:      ip.BaseSalary,
		Bonus:           ip.Bonus,
		FreelanceIncome: ip.FreelanceIncome,
		OtherIncome:     ip.OtherIncome,
		TotalIncome:     ip.TotalIncome(),
		IsActual:        ip.IsActual,
		Notes:           ip.Notes,
		CreatedAt:       ip.CreatedAt,
		UpdatedAt:       ip.UpdatedAt,
	}

	// Include breakdown if requested
	if includeBreakdown {
		response.IncomeBreakdown = ip.GetIncomeBreakdown()
	}

	return response
}

// ToIncomeProfileListResponse converts list of domain models to list response
func ToIncomeProfileListResponse(profiles []*domain.IncomeProfile, includeBreakdown bool) IncomeProfileListResponse {
	responses := make([]IncomeProfileResponse, 0, len(profiles))
	for _, ip := range profiles {
		responses = append(responses, ToIncomeProfileResponse(ip, includeBreakdown))
	}

	return IncomeProfileListResponse{
		IncomeProfiles: responses,
		Count:          len(responses),
	}
}
