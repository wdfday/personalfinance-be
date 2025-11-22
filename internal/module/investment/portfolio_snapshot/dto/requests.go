package dto

import "personalfinancedss/internal/module/investment/portfolio_snapshot/domain"

// CreateSnapshotRequest represents a request to create a new portfolio snapshot
type CreateSnapshotRequest struct {
	SnapshotDate string                 `json:"snapshot_date,omitempty"` // YYYY-MM-DD format, defaults to now
	SnapshotType *domain.SnapshotType   `json:"snapshot_type,omitempty"`
	Period       *domain.SnapshotPeriod `json:"period,omitempty"`
	Notes        *string                `json:"notes,omitempty"`
}

// ListSnapshotsQuery represents query parameters for listing snapshots
type ListSnapshotsQuery struct {
	SnapshotType string `form:"snapshot_type"`
	Period       string `form:"period"`
	StartDate    string `form:"start_date"` // YYYY-MM-DD format
	EndDate      string `form:"end_date"`   // YYYY-MM-DD format
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
