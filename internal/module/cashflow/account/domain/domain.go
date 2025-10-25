package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Account represents a user's financial account.
type Account struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;column:user_id" json:"user_id"`

	AccountName     string      `gorm:"type:varchar(255);not null;column:account_name" json:"account_name"`
	AccountType     AccountType `gorm:"type:varchar(50);not null;column:account_type" json:"account_type"`
	InstitutionName *string     `gorm:"type:varchar(255);column:institution_name" json:"institution_name,omitempty"`

	CurrentBalance   float64  `gorm:"type:decimal(15,2);not null;default:0;column:current_balance" json:"current_balance"`
	AvailableBalance *float64 `gorm:"type:decimal(15,2);column:available_balance" json:"available_balance,omitempty"`
	Currency         Currency `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`

	AccountNumberMasked    *string `gorm:"type:varchar(50);column:account_number_masked" json:"account_number_masked,omitempty"`
	AccountNumberEncrypted *string `gorm:"type:text;column:account_number_encrypted" json:"-"`

	IsActive          bool `gorm:"default:true;column:is_active" json:"is_active"`
	IsPrimary         bool `gorm:"default:false;column:is_primary" json:"is_primary"`
	IncludeInNetWorth bool `gorm:"default:true;column:include_in_net_worth" json:"include_in_net_worth"`

	// IsAutoSync, true for investment - crypto - which has api, false for others
	IsAutoSync       bool        `gorm:"default:false;column:is_auto_sync" json:"is_auto_sync"`
	LastSyncedAt     *time.Time  `gorm:"column:last_synced_at" json:"last_synced_at,omitempty"`
	SyncStatus       *SyncStatus `gorm:"type:varchar(20);column:sync_status" json:"sync_status,omitempty"`
	SyncErrorMessage *string     `gorm:"type:text;column:sync_error_message" json:"sync_error_message,omitempty"`

	// Broker Integration (for investment/crypto_wallet accounts)
	BrokerIntegration datatypes.JSON `gorm:"type:jsonb;column:broker_integration" json:"broker_integration,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName matches the database table.
func (Account) TableName() string {
	return "accounts"
}
