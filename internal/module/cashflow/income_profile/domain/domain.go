// internal/domain/income_profile.go
package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ================================================================
// INCOME PROFILE DOMAIN
// ================================================================

// IncomeProfile represents monthly income data as input for budget allocation
type IncomeProfile struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`
	Year      int       `gorm:"not null;index;column:year" json:"year"`
	Month     int       `gorm:"not null;index;column:month" json:"month"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`

	// Income components
	BaseSalary      float64 `gorm:"type:decimal(15,2);default:0;column:base_salary" json:"base_salary"`
	Bonus           float64 `gorm:"type:decimal(15,2);default:0;column:bonus" json:"bonus"`
	FreelanceIncome float64 `gorm:"type:decimal(15,2);default:0;column:freelance_income" json:"freelance_income"`
	OtherIncome     float64 `gorm:"type:decimal(15,2);default:0;column:other_income" json:"other_income"`

	// Status
	IsActual bool   `gorm:"default:false;column:is_actual" json:"is_actual"` // true = actual occurred, false = projected
	Notes    string `gorm:"type:text;column:notes" json:"notes"`
}

// TableName specifies the database table name
func (IncomeProfile) TableName() string {
	return "income_profiles"
}

// NewIncomeProfile creates a new income profile
func NewIncomeProfile(userID uuid.UUID, year, month int, baseSalary float64) (*IncomeProfile, error) {
	ip := &IncomeProfile{
		ID:         uuid.New(),
		UserID:     userID,
		Year:       year,
		Month:      month,
		BaseSalary: baseSalary,
		IsActual:   false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := ip.Validate(); err != nil {
		return nil, err
	}

	return ip, nil
}

// TotalIncome calculates total monthly income
func (ip *IncomeProfile) TotalIncome() float64 {
	return ip.BaseSalary + ip.Bonus + ip.FreelanceIncome + ip.OtherIncome
}

// AddBonus adds bonus income
func (ip *IncomeProfile) AddBonus(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	ip.Bonus = amount
	ip.UpdatedAt = time.Now()
	return nil
}

// AddFreelanceIncome adds freelance income
func (ip *IncomeProfile) AddFreelanceIncome(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	ip.FreelanceIncome = amount
	ip.UpdatedAt = time.Now()
	return nil
}

// AddOtherIncome adds other income
func (ip *IncomeProfile) AddOtherIncome(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	ip.OtherIncome = amount
	ip.UpdatedAt = time.Now()
	return nil
}

// MarkAsActual marks this income profile as actual (occurred)
func (ip *IncomeProfile) MarkAsActual() {
	ip.IsActual = true
	ip.UpdatedAt = time.Now()
}

// MarkAsProjected marks this income profile as projected (not occurred yet)
func (ip *IncomeProfile) MarkAsProjected() {
	ip.IsActual = false
	ip.UpdatedAt = time.Now()
}

// UpdateNotes updates the notes field
func (ip *IncomeProfile) UpdateNotes(notes string) {
	ip.Notes = notes
	ip.UpdatedAt = time.Now()
}

// Validate performs domain validation
func (ip *IncomeProfile) Validate() error {
	if ip.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if ip.Year < 2000 || ip.Year > 2100 {
		return ErrInvalidYear
	}

	if ip.Month < 1 || ip.Month > 12 {
		return ErrInvalidMonth
	}

	if ip.BaseSalary < 0 {
		return ErrNegativeBaseSalary
	}

	if ip.Bonus < 0 {
		return ErrNegativeAmount
	}

	if ip.FreelanceIncome < 0 {
		return ErrNegativeAmount
	}

	if ip.OtherIncome < 0 {
		return ErrNegativeAmount
	}

	return nil
}

// IsInFuture checks if this income profile is for a future month
func (ip *IncomeProfile) IsInFuture() bool {
	now := time.Now()
	profileDate := time.Date(ip.Year, time.Month(ip.Month), 1, 0, 0, 0, 0, time.UTC)
	return profileDate.After(now)
}

// IsInPast checks if this income profile is for a past month
func (ip *IncomeProfile) IsInPast() bool {
	now := time.Now()
	profileDate := time.Date(ip.Year, time.Month(ip.Month), 1, 0, 0, 0, 0, time.UTC)
	return profileDate.Before(now)
}

// IsCurrentMonth checks if this income profile is for the current month
func (ip *IncomeProfile) IsCurrentMonth() bool {
	now := time.Now()
	return ip.Year == now.Year() && ip.Month == int(now.Month())
}

// GetPeriodKey returns a unique key for this period (YYYY-MM)
func (ip *IncomeProfile) GetPeriodKey() string {
	return formatPeriodKey(ip.Year, ip.Month)
}

// HasMultipleIncomeSources checks if user has income from multiple sources
func (ip *IncomeProfile) HasMultipleIncomeSources() bool {
	sources := 0
	if ip.BaseSalary > 0 {
		sources++
	}
	if ip.Bonus > 0 {
		sources++
	}
	if ip.FreelanceIncome > 0 {
		sources++
	}
	if ip.OtherIncome > 0 {
		sources++
	}
	return sources > 1
}

// GetIncomeBreakdown returns breakdown of income sources with percentages
func (ip *IncomeProfile) GetIncomeBreakdown() map[string]float64 {
	total := ip.TotalIncome()
	if total == 0 {
		return map[string]float64{}
	}

	breakdown := make(map[string]float64)
	if ip.BaseSalary > 0 {
		breakdown["base_salary"] = (ip.BaseSalary / total) * 100
	}
	if ip.Bonus > 0 {
		breakdown["bonus"] = (ip.Bonus / total) * 100
	}
	if ip.FreelanceIncome > 0 {
		breakdown["freelance"] = (ip.FreelanceIncome / total) * 100
	}
	if ip.OtherIncome > 0 {
		breakdown["other"] = (ip.OtherIncome / total) * 100
	}

	return breakdown
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

func formatPeriodKey(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

// ================================================================
// REPOSITORY INTERFACE
// ================================================================

// IncomeProfileRepository defines the interface for income profile persistence
type IncomeProfileRepository interface {
	// Create creates a new income profile
	Create(ip *IncomeProfile) error

	// GetByID retrieves an income profile by ID
	GetByID(id uuid.UUID) (*IncomeProfile, error)

	// GetByUserAndPeriod retrieves an income profile by user and period
	GetByUserAndPeriod(userID uuid.UUID, year, month int) (*IncomeProfile, error)

	// GetByUser retrieves all income profiles for a user
	GetByUser(userID uuid.UUID) ([]*IncomeProfile, error)

	// GetByUserAndYear retrieves all income profiles for a user in a year
	GetByUserAndYear(userID uuid.UUID, year int) ([]*IncomeProfile, error)

	// Update updates an existing income profile
	Update(ip *IncomeProfile) error

	// Delete deletes an income profile
	Delete(id uuid.UUID) error

	// Exists checks if an income profile exists for user and period
	Exists(userID uuid.UUID, year, month int) (bool, error)
}

// ERRORS

var (
	ErrIncomeProfileNotFound = errors.New("income profile not found")
	ErrIncomeProfileExists   = errors.New("income profile already exists for this period")
	ErrInvalidUserID         = errors.New("invalid user ID")
	ErrInvalidYear           = errors.New("year must be between 2000 and 2100")
	ErrInvalidMonth          = errors.New("month must be between 1 and 12")
	ErrNegativeBaseSalary    = errors.New("base salary cannot be negative")
	ErrNegativeAmount        = errors.New("amount cannot be negative")
)
