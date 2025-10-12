package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
)

// VerificationToken represents a token for email verification or password reset
type VerificationToken struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Token  string    `gorm:"type:text;uniqueIndex;not null"`
	Type   string    `gorm:"type:text;not null"` // email_verification, password_reset

	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"index"`

	// Metadata
	IPAddress *string `gorm:"type:text"`
	UserAgent *string `gorm:"type:text"`

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName overrides the table name
func (VerificationToken) TableName() string {
	return "verification_tokens"
}

// IsExpired checks if the token has expired
func (t *VerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *VerificationToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired and not used)
func (t *VerificationToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}
