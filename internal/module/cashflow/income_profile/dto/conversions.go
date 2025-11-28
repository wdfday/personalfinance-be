package dto

import (
	"encoding/json"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"

	"github.com/google/uuid"
)

// FromCreateIncomeProfileRequest converts create request to domain model
func FromCreateIncomeProfileRequest(req CreateIncomeProfileRequest, userID uuid.UUID) (*domain.IncomeProfile, error) {
	// Create new income profile
	ip, err := domain.NewIncomeProfile(
		userID,
		req.Source,
		req.Amount,
		req.Frequency,
		req.StartDate,
	)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	if req.Currency != "" {
		ip.Currency = req.Currency
	}
	if req.EndDate != nil {
		ip.EndDate = req.EndDate
	}
	if req.IsRecurring != nil {
		ip.IsRecurring = *req.IsRecurring
	}

	// Set income component breakdown if provided
	if req.BaseSalary != nil || req.Bonus != nil || req.Commission != nil || req.Allowance != nil || req.OtherIncome != nil {
		baseSalary := float64(0)
		bonus := float64(0)
		commission := float64(0)
		allowance := float64(0)
		otherIncome := float64(0)

		if req.BaseSalary != nil {
			baseSalary = *req.BaseSalary
		}
		if req.Bonus != nil {
			bonus = *req.Bonus
		}
		if req.Commission != nil {
			commission = *req.Commission
		}
		if req.Allowance != nil {
			allowance = *req.Allowance
		}
		if req.OtherIncome != nil {
			otherIncome = *req.OtherIncome
		}

		if err := ip.UpdateComponents(baseSalary, bonus, commission, allowance, otherIncome); err != nil {
			return nil, err
		}
	}

	if req.Description != nil {
		ip.UpdateDescription(*req.Description)
	}

	return ip, nil
}

// ApplyUpdateIncomeProfileRequest applies update request to create new version
func ApplyUpdateIncomeProfileRequest(req UpdateIncomeProfileRequest, existing *domain.IncomeProfile) (*domain.IncomeProfile, error) {
	// Create new version from existing
	newVersion := existing.CreateNewVersion()

	// Apply updates
	if req.Source != nil {
		newVersion.Source = *req.Source
	}
	if req.Amount != nil {
		newVersion.Amount = *req.Amount
	}
	if req.Currency != nil {
		newVersion.Currency = *req.Currency
	}
	if req.Frequency != nil {
		newVersion.Frequency = *req.Frequency
		newVersion.IsRecurring = *req.Frequency != "one-time"
	}
	if req.EndDate != nil {
		newVersion.EndDate = req.EndDate
	}
	if req.IsRecurring != nil {
		newVersion.IsRecurring = *req.IsRecurring
	}

	// Update income components if provided
	if req.BaseSalary != nil || req.Bonus != nil || req.Commission != nil || req.Allowance != nil || req.OtherIncome != nil {
		baseSalary := newVersion.BaseSalary
		bonus := newVersion.Bonus
		commission := newVersion.Commission
		allowance := newVersion.Allowance
		otherIncome := newVersion.OtherIncome

		if req.BaseSalary != nil {
			baseSalary = *req.BaseSalary
		}
		if req.Bonus != nil {
			bonus = *req.Bonus
		}
		if req.Commission != nil {
			commission = *req.Commission
		}
		if req.Allowance != nil {
			allowance = *req.Allowance
		}
		if req.OtherIncome != nil {
			otherIncome = *req.OtherIncome
		}

		if err := newVersion.UpdateComponents(baseSalary, bonus, commission, allowance, otherIncome); err != nil {
			return nil, err
		}
	}

	if req.Description != nil {
		newVersion.UpdateDescription(*req.Description)
	}

	// Validate new version
	if err := newVersion.Validate(); err != nil {
		return nil, err
	}

	return newVersion, nil
}

// ToIncomeProfileResponse converts domain model to response
func ToIncomeProfileResponse(ip *domain.IncomeProfile, includeBreakdown bool) IncomeProfileResponse {
	response := IncomeProfileResponse{
		ID:          ip.ID.String(),
		UserID:      ip.UserID.String(),
		StartDate:   &ip.StartDate,
		EndDate:     ip.EndDate,
		Duration:    ip.GetDuration(),
		Source:      ip.Source,
		Amount:      ip.Amount,
		Currency:    ip.Currency,
		Frequency:   ip.Frequency,
		BaseSalary:  ip.BaseSalary,
		Bonus:       ip.Bonus,
		Commission:  ip.Commission,
		Allowance:   ip.Allowance,
		OtherIncome: ip.OtherIncome,
		TotalIncome: ip.TotalIncome(),
		Status:      string(ip.Status),
		IsRecurring: ip.IsRecurring,
		IsVerified:  ip.IsVerified,
		IsActive:    ip.IsActive(),
		IsArchived:  ip.IsArchived(),
		DSSScore:    ip.GetDSSScore(),
		Description: ip.Description,
		CreatedAt:   ip.CreatedAt,
		UpdatedAt:   ip.UpdatedAt,
		ArchivedAt:  ip.ArchivedAt,
	}

	// Include breakdown if requested
	if includeBreakdown && ip.HasMultipleComponents() {
		response.IncomeBreakdown = ip.GetIncomeBreakdown()
	}

	// Parse DSS metadata
	if len(ip.DSSMetadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(ip.DSSMetadata, &metadata); err == nil {
			response.DSSMetadata = parseDSSMetadata(metadata)
		}
	}

	// Parse tags
	if len(ip.Tags) > 0 {
		var tags []string
		if err := json.Unmarshal(ip.Tags, &tags); err == nil {
			response.Tags = tags
		}
	}

	// Previous version ID
	if ip.PreviousVersionID != nil {
		prevID := ip.PreviousVersionID.String()
		response.PreviousVersionID = &prevID
	}

	return response
}

// ToIncomeProfileWithHistoryResponse converts domain model with version history
func ToIncomeProfileWithHistoryResponse(current *domain.IncomeProfile, history []*domain.IncomeProfile, includeBreakdown bool) IncomeProfileWithHistoryResponse {
	response := IncomeProfileWithHistoryResponse{
		Current: ToIncomeProfileResponse(current, includeBreakdown),
	}

	if len(history) > 0 {
		response.VersionHistory = make([]IncomeProfileResponse, 0, len(history))
		for _, h := range history {
			response.VersionHistory = append(response.VersionHistory, ToIncomeProfileResponse(h, includeBreakdown))
		}
	}

	return response
}

// ToIncomeProfileListResponse converts list of domain models to list response
func ToIncomeProfileListResponse(profiles []*domain.IncomeProfile, includeBreakdown, includeSummary bool) IncomeProfileListResponse {
	responses := make([]IncomeProfileResponse, 0, len(profiles))
	for _, ip := range profiles {
		responses = append(responses, ToIncomeProfileResponse(ip, includeBreakdown))
	}

	result := IncomeProfileListResponse{
		IncomeProfiles: responses,
		Count:          len(responses),
	}

	// Add summary if requested
	if includeSummary {
		result.Summary = calculateSummary(profiles)
	}

	return result
}

// parseDSSMetadata parses DSS metadata map to response struct
func parseDSSMetadata(metadata map[string]interface{}) *DSSMetadataResponse {
	dss := &DSSMetadataResponse{}

	if val, ok := metadata["stability_score"].(float64); ok {
		dss.StabilityScore = val
	}
	if val, ok := metadata["risk_level"].(string); ok {
		dss.RiskLevel = val
	}
	if val, ok := metadata["confidence"].(float64); ok {
		dss.Confidence = val
	}
	if val, ok := metadata["variance"].(float64); ok {
		dss.Variance = val
	}
	if val, ok := metadata["trend"].(string); ok {
		dss.Trend = val
	}
	if val, ok := metadata["recommended_savings_rate"].(float64); ok {
		dss.RecommendedSavingsRate = val
	}
	if val, ok := metadata["last_analyzed"].(string); ok {
		dss.LastAnalyzed = val
	}
	if val, ok := metadata["analysis_version"].(string); ok {
		dss.AnalysisVersion = val
	}

	return dss
}

// calculateSummary calculates summary statistics for income profiles
func calculateSummary(profiles []*domain.IncomeProfile) *IncomeSummaryResponse {
	summary := &IncomeSummaryResponse{}

	totalMonthly := float64(0)
	totalYearly := float64(0)
	activeCount := 0
	recurringCount := 0
	totalStability := float64(0)
	stabilityCount := 0

	for _, ip := range profiles {
		if ip.IsActive() {
			activeCount++

			// Calculate income contribution based on frequency
			monthlyEquivalent := calculateMonthlyEquivalent(ip.Amount, ip.Frequency)
			totalMonthly += monthlyEquivalent
			totalYearly += monthlyEquivalent * 12
		}

		if ip.IsRecurring {
			recurringCount++
		}

		// Calculate average stability from DSS metadata
		score := ip.GetDSSScore()
		if score > 0 {
			totalStability += score
			stabilityCount++
		}
	}

	summary.TotalMonthlyIncome = totalMonthly
	summary.TotalYearlyIncome = totalYearly
	summary.ActiveIncomeCount = activeCount
	summary.RecurringIncomeCount = recurringCount

	if stabilityCount > 0 {
		summary.AverageStability = totalStability / float64(stabilityCount)
	}

	return summary
}

// calculateMonthlyEquivalent converts income to monthly equivalent based on frequency
func calculateMonthlyEquivalent(amount float64, frequency string) float64 {
	switch frequency {
	case "weekly":
		return amount * 52 / 12 // 52 weeks per year / 12 months
	case "bi-weekly":
		return amount * 26 / 12 // 26 bi-weeks per year / 12 months
	case "monthly":
		return amount
	case "quarterly":
		return amount / 3
	case "yearly":
		return amount / 12
	case "one-time":
		return 0 // One-time income doesn't contribute to recurring monthly
	default:
		return amount
	}
}

// FromUpdateDSSMetadataRequest converts request to metadata map
func FromUpdateDSSMetadataRequest(req UpdateDSSMetadataRequest) map[string]interface{} {
	metadata := make(map[string]interface{})

	if req.StabilityScore != nil {
		metadata["stability_score"] = *req.StabilityScore
	}
	if req.RiskLevel != nil {
		metadata["risk_level"] = *req.RiskLevel
	}
	if req.Confidence != nil {
		metadata["confidence"] = *req.Confidence
	}
	if req.Variance != nil {
		metadata["variance"] = *req.Variance
	}
	if req.Trend != nil {
		metadata["trend"] = *req.Trend
	}
	if req.RecommendedSavingsRate != nil {
		metadata["recommended_savings_rate"] = *req.RecommendedSavingsRate
	}

	// Add version info
	metadata["analysis_version"] = "v1.0"

	return metadata
}

// Helper to convert JSON to map for debugging
func JSONToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	return result, err
}
