package domain

// AssetType represents the type of investment asset
type AssetType string

const (
	AssetTypeStock         AssetType = "stock"          // Individual stocks
	AssetTypeETF           AssetType = "etf"            // Exchange-traded funds
	AssetTypeMutualFund    AssetType = "mutual_fund"    // Mutual funds
	AssetTypeBond          AssetType = "bond"           // Bonds
	AssetTypeCrypto        AssetType = "crypto"         // Cryptocurrency
	AssetTypeCommodity     AssetType = "commodity"      // Commodities (gold, silver, oil, etc.)
	AssetTypeRealEstate    AssetType = "real_estate"    // REITs or real estate
	AssetTypeCash          AssetType = "cash"           // Cash or cash equivalents
	AssetTypeOption        AssetType = "option"         // Options contracts
	AssetTypeFuture        AssetType = "future"         // Futures contracts
	AssetTypeForex         AssetType = "forex"          // Foreign exchange
	AssetTypePrivateEquity AssetType = "private_equity" // Private equity
	AssetTypeOther         AssetType = "other"          // Other asset types
)

// IsValid checks if the asset type is valid
func (at AssetType) IsValid() bool {
	switch at {
	case AssetTypeStock, AssetTypeETF, AssetTypeMutualFund, AssetTypeBond,
		AssetTypeCrypto, AssetTypeCommodity, AssetTypeRealEstate, AssetTypeCash,
		AssetTypeOption, AssetTypeFuture, AssetTypeForex, AssetTypePrivateEquity,
		AssetTypeOther:
		return true
	}
	return false
}

// String returns the string representation
func (at AssetType) String() string {
	return string(at)
}

// AssetStatus represents the status of an investment asset
type AssetStatus string

const (
	AssetStatusActive    AssetStatus = "active"    // Currently held
	AssetStatusSold      AssetStatus = "sold"      // Completely sold
	AssetStatusWatchlist AssetStatus = "watchlist" // On watchlist (not owned)
	AssetStatusInactive  AssetStatus = "inactive"  // Temporarily inactive
)

// IsValid checks if the asset status is valid
func (as AssetStatus) IsValid() bool {
	switch as {
	case AssetStatusActive, AssetStatusSold, AssetStatusWatchlist, AssetStatusInactive:
		return true
	}
	return false
}

// String returns the string representation
func (as AssetStatus) String() string {
	return string(as)
}
