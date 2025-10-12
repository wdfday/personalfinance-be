package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenBlacklist represents a blacklisted/revoked token
// When a user logs out, their refresh token is added to this table
// to prevent reuse even before expiration
type TokenBlacklist struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Token     string         `gorm:"type:varchar(500);uniqueIndex;not null" json:"token"`
	UserID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Reason    string         `gorm:"type:varchar(100)" json:"reason"` // "logout", "password_change", "security"
	ExpiresAt time.Time      `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for TokenBlacklist
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// BeforeCreate hook to generate UUID for new records
func (t *TokenBlacklist) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
