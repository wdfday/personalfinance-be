// internal/domain/budget_constraint.go
package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Domain errors
var (
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidCategoryID   = errors.New("invalid category ID")
	ErrMaximumBelowMinimum = errors.New("maximum amount must be greater than or equal to minimum amount")
	ErrInvalidPriority     = errors.New("priority must be between 1 and 10")
	ErrInvalidDateRange    = errors.New("end date must be after start date")
)

// ================================================================
// BUDGET CONSTRAINT DOMAIN
// ================================================================

// ConstraintStatus represents the lifecycle status of a budget constraint
type ConstraintStatus string

const (
	ConstraintStatusActive   ConstraintStatus = "active"   // Currently in effect
	ConstraintStatusPending  ConstraintStatus = "pending"  // Scheduled to start
	ConstraintStatusEnded    ConstraintStatus = "ended"    // Naturally ended
	ConstraintStatusArchived ConstraintStatus = "archived" // Manually archived
	ConstraintStatusPaused   ConstraintStatus = "paused"   // Temporarily suspended
)

// BudgetConstraint represents minimum required spending per category with versioning
// Example: Rent = 2,000,000 VND/month (fixed), Food >= 4,000,000 (flexible)
type BudgetConstraint struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index;column:category_id" json:"category_id"`

	// Period tracking
	Period    string     `gorm:"default:'monthly';column:period" json:"period"`
	StartDate time.Time  `gorm:"not null;column:start_date" json:"start_date"`
	EndDate   *time.Time `gorm:"column:end_date" json:"end_date,omitempty"`

	// Minimum required amount (lower bound)
	MinimumAmount float64 `gorm:"type:decimal(15,2);not null;column:minimum_amount" json:"minimum_amount"`

	// Flexibility
	IsFlexible    bool    `gorm:"default:false;column:is_flexible" json:"is_flexible"`                      // Can DSS allocate more than minimum?
	MaximumAmount float64 `gorm:"type:decimal(15,2);default:0;column:maximum_amount" json:"maximum_amount"` // Upper bound if flexible (0 = no limit)

	// Priority for allocation (1 = highest priority)
	Priority int `gorm:"default:99;column:priority" json:"priority"`

	// Status and lifecycle
	Status      ConstraintStatus `gorm:"type:varchar(20);default:'active';column:status" json:"status"`
	IsRecurring bool             `gorm:"default:true;column:is_recurring" json:"is_recurring"`

	// DSS metadata for analysis
	DSSMetadata []byte `gorm:"type:jsonb;column:dss_metadata" json:"dss_metadata,omitempty"`

	// Additional metadata
	Description string `gorm:"type:text;column:description" json:"description,omitempty"`
	Tags        []byte `gorm:"type:jsonb;column:tags" json:"tags,omitempty"`

	// Timestamps
	CreatedAt time.Time  `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	// Archived tracking
	ArchivedAt *time.Time `gorm:"column:archived_at" json:"archived_at,omitempty"`
	ArchivedBy *uuid.UUID `gorm:"type:uuid;column:archived_by" json:"archived_by,omitempty"`

	// Versioning support
	PreviousVersionID *uuid.UUID `gorm:"type:uuid;column:previous_version_id;index" json:"previous_version_id,omitempty"`
}

// TableName specifies the database table name
func (BudgetConstraint) TableName() string {
	return "budget_constraints"
}

// NewBudgetConstraint creates a new budget constraint
func NewBudgetConstraint(userID, categoryID uuid.UUID, minimumAmount float64, startDate time.Time) (*BudgetConstraint, error) {
	bc := &BudgetConstraint{
		ID:            uuid.New(),
		UserID:        userID,
		CategoryID:    categoryID,
		Period:        "monthly",
		StartDate:     startDate,
		MinimumAmount: minimumAmount,
		IsFlexible:    false,
		MaximumAmount: 0,
		Priority:      99, // Default low priority
		Status:        ConstraintStatusActive,
		IsRecurring:   true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := bc.Validate(); err != nil {
		return nil, err
	}

	return bc, nil
}

// NewFlexibleBudgetConstraint creates a flexible budget constraint with range
func NewFlexibleBudgetConstraint(userID, categoryID uuid.UUID, min, max float64, startDate time.Time) (*BudgetConstraint, error) {
	bc := &BudgetConstraint{
		ID:            uuid.New(),
		UserID:        userID,
		CategoryID:    categoryID,
		Period:        "monthly",
		StartDate:     startDate,
		MinimumAmount: min,
		IsFlexible:    true,
		MaximumAmount: max,
		Priority:      99,
		Status:        ConstraintStatusActive,
		IsRecurring:   true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := bc.Validate(); err != nil {
		return nil, err
	}

	return bc, nil
}

// GetRange returns [min, max] allocation range for this constraint
func (bc *BudgetConstraint) GetRange() (min, max float64) {
	min = bc.MinimumAmount
	max = bc.MinimumAmount // Default: fixed amount

	if bc.IsFlexible {
		if bc.MaximumAmount > 0 {
			max = bc.MaximumAmount
		} else {
			max = 0 // No upper limit
		}
	}

	return min, max
}

// IsFixed returns true if this is a fixed amount (not flexible)
func (bc *BudgetConstraint) IsFixed() bool {
	return !bc.IsFlexible
}

// HasUpperLimit returns true if there's a maximum amount specified
func (bc *BudgetConstraint) HasUpperLimit() bool {
	return bc.IsFlexible && bc.MaximumAmount > 0
}

// GetFlexibilityRange returns the range of flexibility (max - min)
func (bc *BudgetConstraint) GetFlexibilityRange() float64 {
	if !bc.IsFlexible {
		return 0
	}
	if bc.MaximumAmount > 0 {
		return bc.MaximumAmount - bc.MinimumAmount
	}
	return 0 // Unlimited flexibility
}

// CanAllocate checks if a proposed amount is within valid range
func (bc *BudgetConstraint) CanAllocate(amount float64) bool {
	if amount < bc.MinimumAmount {
		return false // Below minimum
	}

	if !bc.IsFlexible {
		return amount == bc.MinimumAmount // Must be exact for fixed
	}

	if bc.MaximumAmount > 0 && amount > bc.MaximumAmount {
		return false // Above maximum
	}

	return true
}

// UpdateMinimum updates the minimum required amount
func (bc *BudgetConstraint) UpdateMinimum(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}

	// If flexible with max, ensure min <= max
	if bc.IsFlexible && bc.MaximumAmount > 0 && amount > bc.MaximumAmount {
		return ErrMinimumExceedsMaximum
	}

	bc.MinimumAmount = amount
	bc.UpdatedAt = time.Now()
	return nil
}

// SetFlexible makes this constraint flexible with optional max
func (bc *BudgetConstraint) SetFlexible(maxAmount float64) error {
	if maxAmount > 0 && maxAmount < bc.MinimumAmount {
		return ErrMaximumBelowMinimum
	}

	bc.IsFlexible = true
	bc.MaximumAmount = maxAmount
	bc.UpdatedAt = time.Now()
	return nil
}

// SetFixed makes this constraint fixed (non-flexible)
func (bc *BudgetConstraint) SetFixed() {
	bc.IsFlexible = false
	bc.MaximumAmount = 0
	bc.UpdatedAt = time.Now()
}

// SetPriority sets the allocation priority (1 = highest)
func (bc *BudgetConstraint) SetPriority(priority int) error {
	if priority < 1 {
		return ErrInvalidPriority
	}
	bc.Priority = priority
	bc.UpdatedAt = time.Now()
	return nil
}

// Archive archives this budget constraint
func (bc *BudgetConstraint) Archive(archivedBy uuid.UUID) {
	now := time.Now()
	bc.Status = ConstraintStatusArchived
	bc.ArchivedAt = &now
	bc.ArchivedBy = &archivedBy
	bc.UpdatedAt = now
}

// CreateNewVersion creates a new version from current constraint
func (bc *BudgetConstraint) CreateNewVersion() *BudgetConstraint {
	newVersion := &BudgetConstraint{
		ID:                uuid.New(),
		UserID:            bc.UserID,
		CategoryID:        bc.CategoryID,
		StartDate:         time.Now(),
		EndDate:           bc.EndDate,
		MinimumAmount:     bc.MinimumAmount,
		IsFlexible:        bc.IsFlexible,
		MaximumAmount:     bc.MaximumAmount,
		Priority:          bc.Priority,
		Status:            ConstraintStatusActive,
		IsRecurring:       bc.IsRecurring,
		DSSMetadata:       bc.DSSMetadata,
		Description:       bc.Description,
		Tags:              bc.Tags,
		PreviousVersionID: &bc.ID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	return newVersion
}

// CheckAndArchiveIfEnded checks if constraint has ended and auto-archives
func (bc *BudgetConstraint) CheckAndArchiveIfEnded() bool {
	if bc.EndDate != nil && time.Now().After(*bc.EndDate) {
		if bc.Status == ConstraintStatusActive {
			bc.Status = ConstraintStatusEnded
			now := time.Now()
			bc.ArchivedAt = &now
			bc.UpdatedAt = now
			return true
		}
	}
	return false
}

// IsActive checks if this constraint is currently active
func (bc *BudgetConstraint) IsActive() bool {
	now := time.Now()
	if bc.Status != ConstraintStatusActive {
		return false
	}
	if bc.StartDate.After(now) {
		return false
	}
	if bc.EndDate != nil && bc.EndDate.Before(now) {
		return false
	}
	return true
}

// IsArchived checks if this constraint is archived
func (bc *BudgetConstraint) IsArchived() bool {
	return bc.ArchivedAt != nil || bc.Status == ConstraintStatusArchived
}

// Validate performs domain validation
func (bc *BudgetConstraint) Validate() error {
	if bc.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if bc.CategoryID == uuid.Nil {
		return ErrInvalidCategoryID
	}

	if bc.MinimumAmount < 0 {
		return ErrNegativeAmount
	}

	if bc.IsFlexible {
		if bc.MaximumAmount > 0 && bc.MaximumAmount < bc.MinimumAmount {
			return ErrMaximumBelowMinimum
		}
	}

	if bc.Priority < 0 {
		return ErrInvalidPriority
	}

	if bc.StartDate.IsZero() {
		return ErrInvalidStartDate
	}

	if bc.EndDate != nil && bc.EndDate.Before(bc.StartDate) {
		return ErrEndDateBeforeStartDate
	}

	return nil
}

// String returns a human-readable representation
func (bc *BudgetConstraint) String() string {
	if bc.IsFixed() {
		return fmt.Sprintf("Fixed: %.2f", bc.MinimumAmount)
	}

	if bc.HasUpperLimit() {
		return fmt.Sprintf("Flexible: [%.2f, %.2f]", bc.MinimumAmount, bc.MaximumAmount)
	}

	return fmt.Sprintf("Flexible: >= %.2f (no limit)", bc.MinimumAmount)
}

// ================================================================
// BUDGET CONSTRAINT COLLECTION
// ================================================================

// BudgetConstraints represents a collection of budget constraints for a user
type BudgetConstraints []*BudgetConstraint

// TotalMandatoryExpenses calculates sum of all minimum amounts
func (bcs BudgetConstraints) TotalMandatoryExpenses() float64 {
	var total float64
	for _, bc := range bcs {
		total += bc.MinimumAmount
	}
	return total
}

// GetByCategory finds constraint by category ID
func (bcs BudgetConstraints) GetByCategory(categoryID uuid.UUID) *BudgetConstraint {
	for _, bc := range bcs {
		if bc.CategoryID == categoryID {
			return bc
		}
	}
	return nil
}

// GetFlexible returns only flexible constraints
func (bcs BudgetConstraints) GetFlexible() BudgetConstraints {
	var flexible BudgetConstraints
	for _, bc := range bcs {
		if bc.IsFlexible {
			flexible = append(flexible, bc)
		}
	}
	return flexible
}

// GetFixed returns only fixed constraints
func (bcs BudgetConstraints) GetFixed() BudgetConstraints {
	var fixed BudgetConstraints
	for _, bc := range bcs {
		if bc.IsFixed() {
			fixed = append(fixed, bc)
		}
	}
	return fixed
}

// SortByPriority returns constraints sorted by priority (ascending)
func (bcs BudgetConstraints) SortByPriority() BudgetConstraints {
	sorted := make(BudgetConstraints, len(bcs))
	copy(sorted, bcs)

	// Simple bubble sort (good enough for small lists)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority > sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// Validate validates all constraints in collection
func (bcs BudgetConstraints) Validate() error {
	for i, bc := range bcs {
		if err := bc.Validate(); err != nil {
			return fmt.Errorf("constraint %d: %w", i, err)
		}
	}
	return nil
}

// HasDuplicateCategories checks if there are multiple constraints for same category
func (bcs BudgetConstraints) HasDuplicateCategories() bool {
	seen := make(map[uuid.UUID]bool)
	for _, bc := range bcs {
		if seen[bc.CategoryID] {
			return true
		}
		seen[bc.CategoryID] = true
	}
	return false
}

// ================================================================
// REPOSITORY INTERFACE
// ================================================================

// BudgetConstraintRepository defines the interface for budget constraint persistence with versioning
type BudgetConstraintRepository interface {
	// Create creates a new budget constraint
	Create(bc *BudgetConstraint) error

	// GetByID retrieves a budget constraint by ID
	GetByID(id uuid.UUID) (*BudgetConstraint, error)

	// GetByUser retrieves all active budget constraints for a user (not archived)
	GetByUser(userID uuid.UUID) (BudgetConstraints, error)

	// GetActiveByUser retrieves all currently active budget constraints
	GetActiveByUser(userID uuid.UUID) (BudgetConstraints, error)

	// GetArchivedByUser retrieves all archived budget constraints
	GetArchivedByUser(userID uuid.UUID) (BudgetConstraints, error)

	// GetByUserAndCategory retrieves active constraint by user and category
	GetByUserAndCategory(userID, categoryID uuid.UUID) (*BudgetConstraint, error)

	// GetByStatus retrieves constraints by user and status
	GetByStatus(userID uuid.UUID, status ConstraintStatus) (BudgetConstraints, error)

	// GetVersionHistory retrieves all versions of a constraint
	GetVersionHistory(constraintID uuid.UUID) (BudgetConstraints, error)

	// GetLatestVersion retrieves the latest version of a constraint chain
	GetLatestVersion(constraintID uuid.UUID) (*BudgetConstraint, error)

	// Update updates an existing budget constraint
	Update(bc *BudgetConstraint) error

	// Delete soft deletes a budget constraint
	Delete(id uuid.UUID) error

	// Archive archives a budget constraint
	Archive(id uuid.UUID, archivedBy uuid.UUID) error

	// Exists checks if a budget constraint exists for user and category
	Exists(userID, categoryID uuid.UUID) (bool, error)

	// GetTotalMandatory calculates total mandatory expenses for user
	GetTotalMandatory(userID uuid.UUID) (float64, error)
}

// ================================================================
// ERRORS
// ================================================================

var (
	ErrBudgetConstraintNotFound = errors.New("budget constraint not found")
	ErrBudgetConstraintExists   = errors.New("budget constraint already exists for this category")
	ErrNegativeAmount           = errors.New("amount cannot be negative")
	ErrMinimumExceedsMaximum    = errors.New("minimum amount cannot exceed maximum")
	ErrInvalidStartDate         = errors.New("start date is required")
	ErrEndDateBeforeStartDate   = errors.New("end date cannot be before start date")
	ErrCannotUpdateArchived     = errors.New("cannot update archived constraint")
)
