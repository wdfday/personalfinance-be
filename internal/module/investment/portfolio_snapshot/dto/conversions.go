package dto

import "personalfinancedss/internal/module/investment/portfolio_snapshot/domain"

// ToSnapshotResponse converts a domain snapshot to a response DTO
func ToSnapshotResponse(snapshot *domain.PortfolioSnapshot) SnapshotResponse {
	if snapshot == nil {
		return SnapshotResponse{}
	}

	response := SnapshotResponse{
		ID:                  snapshot.ID,
		UserID:              snapshot.UserID,
		SnapshotDate:        snapshot.SnapshotDate,
		SnapshotType:        string(snapshot.SnapshotType),
		TotalValue:          snapshot.TotalValue,
		TotalCost:           snapshot.TotalCost,
		TotalUnrealizedGain: snapshot.TotalUnrealizedGain,
		TotalRealizedGain:   snapshot.TotalRealizedGain,
		TotalDividends:      snapshot.TotalDividends,
		TotalReturn:         snapshot.TotalReturn,
		TotalReturnPct:      snapshot.TotalReturnPct,
		DayChange:           snapshot.DayChange,
		DayChangePct:        snapshot.DayChangePct,
		TotalAssets:         snapshot.TotalAssets,
		ActiveAssets:        snapshot.ActiveAssets,
		CashInflow:          snapshot.CashInflow,
		CashOutflow:         snapshot.CashOutflow,
		NetCashFlow:         snapshot.NetCashFlow,
		Volatility:          snapshot.Volatility,
		SharpeRatio:         snapshot.SharpeRatio,
		Beta:                snapshot.Beta,
		CreatedAt:           snapshot.CreatedAt,
	}

	if snapshot.Period != nil {
		response.Period = string(*snapshot.Period)
	}
	if snapshot.AssetTypes != nil {
		response.AssetTypes = *snapshot.AssetTypes
	}
	if snapshot.SectorAllocation != nil {
		response.SectorAllocation = *snapshot.SectorAllocation
	}
	if snapshot.Notes != nil {
		response.Notes = *snapshot.Notes
	}

	return response
}

// ToSnapshotListResponse converts a list of domain snapshots to a list response DTO
func ToSnapshotListResponse(snapshots []*domain.PortfolioSnapshot, total int64, page, pageSize int) SnapshotListResponse {
	snapshotResponses := make([]SnapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotResponses = append(snapshotResponses, ToSnapshotResponse(snapshot))
	}

	return SnapshotListResponse{
		Snapshots: snapshotResponses,
		Total:     total,
		Page:      page,
		PerPage:   pageSize,
	}
}
