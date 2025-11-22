package dto

import (
	"personalfinancedss/internal/module/investment/investment_asset/domain"
)

// ToAssetResponse converts a domain asset to a response DTO
func ToAssetResponse(asset *domain.InvestmentAsset) AssetResponse {
	if asset == nil {
		return AssetResponse{}
	}

	response := AssetResponse{
		ID:        asset.ID,
		UserID:    asset.UserID,
		CreatedAt: asset.CreatedAt,
		UpdatedAt: asset.UpdatedAt,

		Symbol:     asset.Symbol,
		Name:       asset.Name,
		AssetType:  string(asset.AssetType),
		AssetClass: asset.AssetClass,
		Currency:   asset.Currency,

		Quantity:           asset.Quantity,
		AverageCostPerUnit: asset.AverageCostPerUnit,
		TotalCost:          asset.TotalCost,

		CurrentPrice:      asset.CurrentPrice,
		CurrentValue:      asset.CurrentValue,
		UnrealizedGain:    asset.UnrealizedGain,
		UnrealizedGainPct: asset.UnrealizedGainPct,

		RealizedGain:    asset.RealizedGain,
		RealizedGainPct: asset.RealizedGainPct,

		PortfolioWeight: asset.PortfolioWeight,

		TotalDividends: asset.TotalDividends,
		DividendYield:  asset.DividendYield,

		Status:          string(asset.Status),
		IsWatchlist:     asset.IsWatchlist,
		AutoUpdatePrice: asset.AutoUpdatePrice,

		Beta:            asset.Beta,
		Volatility:      asset.Volatility,
		SharpeRatio:     asset.SharpeRatio,
		MaxDrawdown:     asset.MaxDrawdown,
		OneYearReturn:   asset.OneYearReturn,
		ThreeYearReturn: asset.ThreeYearReturn,
		FiveYearReturn:  asset.FiveYearReturn,

		TotalReturn:    asset.TotalReturn(),
		TotalReturnPct: asset.TotalReturnPct(),
	}

	// Handle optional pointer fields
	if asset.Sector != nil {
		response.Sector = *asset.Sector
	}
	if asset.Industry != nil {
		response.Industry = *asset.Industry
	}
	if asset.Exchange != nil {
		response.Exchange = *asset.Exchange
	}
	if asset.Notes != nil {
		response.Notes = *asset.Notes
	}
	if asset.Tags != nil {
		response.Tags = *asset.Tags
	}
	if asset.ISIN != nil {
		response.ISIN = *asset.ISIN
	}
	if asset.CUSIP != nil {
		response.CUSIP = *asset.CUSIP
	}
	if asset.LastDividendAmount > 0 {
		response.LastDividendAmount = asset.LastDividendAmount
	}
	if asset.LastDividendDate != nil {
		response.LastDividendDate = *asset.LastDividendDate
	}
	if asset.LastPriceUpdate != nil {
		response.LastPriceUpdate = *asset.LastPriceUpdate
	}

	return response
}

// ToAssetListResponse converts a list of domain assets to a list response DTO
func ToAssetListResponse(assets []*domain.InvestmentAsset, total int64, page, pageSize int) AssetListResponse {
	assetResponses := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		assetResponses = append(assetResponses, ToAssetResponse(asset))
	}

	return AssetListResponse{
		Assets:  assetResponses,
		Total:   total,
		Page:    page,
		PerPage: pageSize,
	}
}
