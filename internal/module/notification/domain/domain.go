package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Domain errors
var (
	ErrInvalidNotificationType = errors.New("invalid notification type")
	ErrInvalidChannel          = errors.New("invalid notification channel")
)

// Notification represents a notification to be sent to a user
type Notification struct {
	ID      uuid.UUID           `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID  uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	Type    NotificationType    `gorm:"type:varchar(50);not null" json:"type"`
	Channel NotificationChannel `gorm:"type:varchar(20);not null;default:'email'" json:"channel"`

	Recipient string `gorm:"type:varchar(255);not null" json:"recipient"` // email address, phone number, etc.
	Subject   string `gorm:"type:varchar(500)" json:"subject,omitempty"`

	// Template and data
	TemplateName string                 `gorm:"type:varchar(100)" json:"template_name,omitempty"`
	Data         map[string]interface{} `gorm:"type:jsonb" json:"data,omitempty"`

	// Status tracking
	Status       string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, sent, failed
	SentAt       *time.Time `gorm:"index" json:"sent_at,omitempty"`
	FailedAt     *time.Time `json:"failed_at,omitempty"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	RetryCount   int        `gorm:"default:0" json:"retry_count"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// EmailConfig holds configuration for email service
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	FrontendURL  string
}

// NotificationData is a simplified structure for sending notifications
type NotificationData struct {
	Type      NotificationType
	UserEmail string
	UserName  string
	Data      map[string]interface{}
	Timestamp time.Time
}

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type    string                 `json:"type"`    // message type: notification, ping, pong, etc.
	Payload map[string]interface{} `json:"payload"` // message data
}

// WebSocketNotificationPayload represents a notification sent via WebSocket
type WebSocketNotificationPayload struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationPreference represents user preferences for notifications
type NotificationPreference struct {
	ID     uuid.UUID        `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_user_notification_type" json:"user_id"`
	Type   NotificationType `gorm:"type:varchar(50);not null;uniqueIndex:idx_user_notification_type" json:"type"`

	// Channel preferences
	Enabled           bool                  `gorm:"not null;default:true" json:"enabled"`
	PreferredChannels []NotificationChannel `gorm:"type:varchar(100)" json:"preferred_channels"` // comma-separated

	// Frequency settings
	MinInterval    int    `gorm:"default:0" json:"min_interval_minutes,omitempty"` // minimum minutes between same type
	QuietHoursFrom *int   `gorm:"type:smallint" json:"quiet_hours_from,omitempty"` // hour 0-23
	QuietHoursTo   *int   `gorm:"type:smallint" json:"quiet_hours_to,omitempty"`   // hour 0-23
	Timezone       string `gorm:"type:varchar(50);default:'UTC'" json:"timezone"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for NotificationPreference
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}
