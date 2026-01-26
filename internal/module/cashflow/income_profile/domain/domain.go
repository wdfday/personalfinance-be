// internal/domain/income_profile.go
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ================================================================
// INCOME PROFILE DOMAIN
// ================================================================

// IncomeProfile represents income data with period tracking and DSS analysis
type IncomeProfile struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Linked Resources
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index;column:category_id" json:"category_id"`

	// Period tracking
	StartDate time.Time  `gorm:"not null;column:start_date" json:"start_date"`
	EndDate   *time.Time `gorm:"column:end_date" json:"end_date,omitempty"`

	// Income details
	Source    string  `gorm:"type:varchar(100);not null;column:source" json:"source"` // e.g., "Salary - Company X", "Freelance", "Investment"
	Amount    float64 `gorm:"type:decimal(15,2);not null;column:amount" json:"amount"`
	Currency  string  `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`
	Frequency string  `gorm:"type:varchar(20);not null;column:frequency" json:"frequency"` // monthly, weekly, one-time, quarterly

	// Status and lifecycle
	Status      IncomeStatus `gorm:"type:varchar(20);default:'active';column:status" json:"status"`
	IsRecurring bool         `gorm:"column:is_recurring" json:"is_recurring"`

	// DSS Analysis Metadata (JSONB)
	DSSMetadata datatypes.JSON `gorm:"type:jsonb;column:dss_metadata" json:"dss_metadata,omitempty"`
	// Structure:
	// {
	//   "stability_score": 0.95,           // 0-1, higher = more stable
	//   "risk_level": "low",               // low, medium, high
	//   "confidence": 0.85,                // confidence in predictions
	//   "variance": 0.05,                  // income variance over time
	//   "trend": "stable",                 // increasing, stable, decreasing
	//   "recommended_savings_rate": 0.3,   // suggested % to save
	//   "last_analyzed": "2024-01-15T10:00:00Z",
	//   "analysis_version": "v1.0"
	// }

	// Additional metadata
	Description string         `gorm:"type:text;column:description" json:"description,omitempty"`
	Tags        datatypes.JSON `gorm:"type:jsonb;column:tags" json:"tags,omitempty"`
	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`

	// Archived tracking
	ArchivedAt *time.Time `gorm:"column:archived_at" json:"archived_at,omitempty"`
	ArchivedBy *uuid.UUID `gorm:"type:uuid;column:archived_by" json:"archived_by,omitempty"`

	// Versioning support
	PreviousVersionID *uuid.UUID `gorm:"type:uuid;column:previous_version_id;index" json:"previous_version_id,omitempty"`
}

// IncomeStatus represents the lifecycle status of income
type IncomeStatus string

const (
	IncomeStatusActive   IncomeStatus = "active"   // Currently receiving
	IncomeStatusEnded    IncomeStatus = "ended"    // Naturally ended
	IncomeStatusArchived IncomeStatus = "archived" // Manually archived by user
	IncomeStatusPaused   IncomeStatus = "paused"   // Temporarily paused
)

// TableName specifies the database table name
func (IncomeProfile) TableName() string {
	return "income_profiles"
}

// NewIncomeProfile creates a new income profile
func NewIncomeProfile(
	userID uuid.UUID,
	source string,
	amount float64,
	frequency string,
	startDate time.Time,
) (*IncomeProfile, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	ip := &IncomeProfile{
		ID:          id,
		UserID:      userID,
		Source:      source,
		Amount:      amount,
		Currency:    "VND",
		Frequency:   frequency,
		StartDate:   startDate,
		Status:      IncomeStatusActive,
		IsRecurring: frequency != "one-time",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := ip.Validate(); err != nil {
		return nil, err
	}

	return ip, nil
}

// UpdateDescription updates the description field
func (ip *IncomeProfile) UpdateDescription(description string) {
	ip.Description = description
	ip.UpdatedAt = time.Now()
}

// UpdateStatus updates the status of the income profile
func (ip *IncomeProfile) UpdateStatus(status IncomeStatus) error {
	if !isValidStatus(status) {
		return ErrInvalidStatus
	}
	ip.Status = status
	ip.UpdatedAt = time.Now()
	return nil
}

// Archive archives this income profile
func (ip *IncomeProfile) Archive(archivedBy uuid.UUID) {
	now := time.Now()
	ip.Status = IncomeStatusArchived
	ip.ArchivedAt = &now
	ip.ArchivedBy = &archivedBy
	ip.UpdatedAt = now
}

// CreateNewVersion creates a new version from current profile
func (ip *IncomeProfile) CreateNewVersion() *IncomeProfile {
	newVersion := &IncomeProfile{
		ID:                uuid.New(),
		UserID:            ip.UserID,
		StartDate:         time.Now(),
		EndDate:           ip.EndDate,
		Source:            ip.Source,
		Amount:            ip.Amount,
		Currency:          ip.Currency,
		Frequency:         ip.Frequency,
		Status:            IncomeStatusActive,
		IsRecurring:       ip.IsRecurring,
		DSSMetadata:       ip.DSSMetadata,
		Description:       ip.Description,
		Tags:              ip.Tags,
		PreviousVersionID: &ip.ID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	return newVersion
}

// CheckAndMarkAsEnded checks if income has ended and marks it as ended (does not archive)
func (ip *IncomeProfile) CheckAndMarkAsEnded() bool {
	if ip.EndDate != nil && time.Now().After(*ip.EndDate) {
		if ip.Status == IncomeStatusActive {
			ip.MarkAsEnded()
			return true
		}
	}
	return false
}

// UpdateDSSMetadata updates DSS analysis metadata
func (ip *IncomeProfile) UpdateDSSMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		return ErrInvalidMetadata
	}

	// Add last_analyzed timestamp
	metadata["last_analyzed"] = time.Now().Format(time.RFC3339)

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	ip.DSSMetadata = jsonBytes
	ip.UpdatedAt = time.Now()
	return nil
}

// Validate performs domain validation
func (ip *IncomeProfile) Validate() error {
	if ip.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if ip.Source == "" {
		return ErrInvalidSource
	}

	if ip.Amount < 0 {
		return ErrNegativeAmount
	}

	if !isValidFrequency(ip.Frequency) {
		return ErrInvalidFrequency
	}

	if ip.StartDate.IsZero() {
		return ErrInvalidStartDate
	}

	if ip.EndDate != nil && ip.EndDate.Before(ip.StartDate) {
		return ErrEndDateBeforeStartDate
	}

	if !isValidStatus(ip.Status) {
		return ErrInvalidStatus
	}

	return nil
}

// IsActive checks if this income profile is currently active
func (ip *IncomeProfile) IsActive() bool {
	now := time.Now()
	if ip.Status != IncomeStatusActive {
		return false
	}
	if ip.StartDate.After(now) {
		return false
	}
	if ip.EndDate != nil && ip.EndDate.Before(now) {
		return false
	}
	return true
}

// MarkAsEnded marks this income profile as ended
func (ip *IncomeProfile) MarkAsEnded() {
	ip.Status = IncomeStatusEnded
	now := time.Now()
	if ip.EndDate == nil {
		ip.EndDate = &now
	}
	ip.UpdatedAt = now
}

// IsEnded checks if this income profile has ended
func (ip *IncomeProfile) IsEnded() bool {
	if ip.Status == IncomeStatusEnded || ip.Status == IncomeStatusArchived {
		return true
	}
	if ip.EndDate != nil && ip.EndDate.Before(time.Now()) {
		return true
	}
	return false
}

// IsArchived checks if this income profile is archived
func (ip *IncomeProfile) IsArchived() bool {
	return ip.ArchivedAt != nil || ip.Status == IncomeStatusArchived
}

// GetDuration returns the duration of this income in days
func (ip *IncomeProfile) GetDuration() int {
	end := time.Now()
	if ip.EndDate != nil {
		end = *ip.EndDate
	}
	return int(end.Sub(ip.StartDate).Hours() / 24)
}

// GetDSSScore returns the stability score from DSS metadata
func (ip *IncomeProfile) GetDSSScore() float64 {
	if len(ip.DSSMetadata) == 0 {
		return 0.0
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(ip.DSSMetadata, &metadata); err != nil {
		return 0.0
	}
	if score, ok := metadata["stability_score"].(float64); ok {
		return score
	}
	return 0.0
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

func isValidFrequency(frequency string) bool {
	validFrequencies := map[string]bool{
		"monthly":   true,
		"weekly":    true,
		"bi-weekly": true,
		"quarterly": true,
		"yearly":    true,
		"one-time":  true,
	}
	return validFrequencies[frequency]
}

func isValidStatus(status IncomeStatus) bool {
	validStatuses := map[IncomeStatus]bool{
		IncomeStatusActive:   true,
		IncomeStatusEnded:    true,
		IncomeStatusArchived: true,
		IncomeStatusPaused:   true,
	}
	return validStatuses[status]
}

// ================================================================
// REPOSITORY INTERFACE
// ================================================================

// ================================================================
// ERRORS
// ================================================================

var (
	ErrIncomeProfileNotFound  = errors.New("income profile not found")
	ErrIncomeProfileExists    = errors.New("income profile already exists")
	ErrInvalidUserID          = errors.New("invalid user ID")
	ErrInvalidSource          = errors.New("income source cannot be empty")
	ErrInvalidFrequency       = errors.New("invalid frequency")
	ErrInvalidStartDate       = errors.New("start date is required")
	ErrEndDateBeforeStartDate = errors.New("end date cannot be before start date")
	ErrNegativeAmount         = errors.New("amount cannot be negative")
	ErrInvalidStatus          = errors.New("invalid income status")
	ErrInvalidMetadata        = errors.New("invalid DSS metadata")
	ErrCannotArchiveArchived  = errors.New("cannot archive already archived profile")
	ErrCannotUpdateArchived   = errors.New("cannot update archived profile")
)
