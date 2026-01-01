package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventType represents the category of an event
type EventType string

const (
	EventTypePersonal    EventType = "personal"
	EventTypeHoliday     EventType = "holiday"
	EventTypeBirthday    EventType = "birthday"
	EventTypeAnniversary EventType = "anniversary"
	EventTypeMeeting     EventType = "meeting"
	EventTypeReminder    EventType = "reminder"
	EventTypeOther       EventType = "other"
)

// EventSource represents where the event came from
type EventSource string

const (
	SourceUserCreated     EventSource = "user_created"
	SourceSystemGenerated EventSource = "system_generated"
)

// RecurrenceType represents simple recurring patterns
type RecurrenceType string

const (
	RecurrenceNone    RecurrenceType = "none"
	RecurrenceDaily   RecurrenceType = "daily"
	RecurrenceWeekly  RecurrenceType = "weekly"
	RecurrenceMonthly RecurrenceType = "monthly"
	RecurrenceYearly  RecurrenceType = "yearly"
)

// Event represents a calendar event (personal or system-generated)
type Event struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	// Basic Info
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`

	// Classification
	Type   EventType   `gorm:"type:varchar(50);default:'personal'" json:"type"`
	Source EventSource `gorm:"type:varchar(50);default:'user_created'" json:"source"`

	// Timeline (use TIMESTAMP for time support)
	StartDate time.Time  `gorm:"type:timestamp;not null;index" json:"start_date"`
	EndDate   *time.Time `gorm:"type:timestamp" json:"end_date,omitempty"`
	AllDay    bool       `gorm:"default:true" json:"all_day"` // Most calendar events are all-day

	// Display
	Color *string `gorm:"type:varchar(7)" json:"color,omitempty"` // Hex color like "#FF5733"
	Tags  []byte  `gorm:"type:jsonb" json:"tags,omitempty"`       // JSON array of tags

	// Recurring
	IsRecurring    bool            `gorm:"default:false" json:"is_recurring"`
	RecurrenceType *RecurrenceType `gorm:"type:varchar(20)" json:"recurrence_type,omitempty"` // "yearly", "monthly", etc.

	// Metadata
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Event
func (Event) TableName() string {
	return "calendar_events"
}

// Validate validates the event data
func (e *Event) Validate() error {
	if e.Name == "" {
		return errors.New("event name is required")
	}
	if e.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	if e.EndDate != nil && e.EndDate.Before(e.StartDate) {
		return errors.New("end date must be on or after start date")
	}
	if e.Color != nil && len(*e.Color) > 0 {
		// Simple hex color validation
		if (*e.Color)[0] != '#' || (len(*e.Color) != 7 && len(*e.Color) != 4) {
			return errors.New("color must be in hex format (#RRGGBB or #RGB)")
		}
	}
	return nil
}

// IsMultiDay checks if event spans multiple days
func (e *Event) IsMultiDay() bool {
	if e.EndDate == nil {
		return false
	}
	return e.EndDate.After(e.StartDate)
}

// Duration returns the duration in days
func (e *Event) Duration() int {
	if e.EndDate == nil {
		return 1
	}
	return int(e.EndDate.Sub(e.StartDate).Hours()/24) + 1
}

// IsHoliday checks if this is a system-generated holiday
func (e *Event) IsHoliday() bool {
	return e.Type == EventTypeHoliday && e.Source == SourceSystemGenerated
}

// IsUserEvent checks if this is a user-created event
func (e *Event) IsUserEvent() bool {
	return e.Source == SourceUserCreated
}
