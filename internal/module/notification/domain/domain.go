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

// SecurityEvent represents a security-related event for audit logging
type SecurityEvent struct {
	ID        uuid.UUID         `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	Type      SecurityEventType `gorm:"type:varchar(50);not null;index" json:"type"`
	UserID    *uuid.UUID        `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Email     string            `gorm:"type:varchar(255);index" json:"email,omitempty"`
	IPAddress string            `gorm:"type:varchar(50);index" json:"ip_address,omitempty"`
	UserAgent string            `gorm:"type:text" json:"user_agent,omitempty"`

	Success  bool                   `gorm:"not null;index" json:"success"`
	Message  string                 `gorm:"type:text" json:"message"`
	Metadata map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`

	Timestamp time.Time      `gorm:"autoCreateTime;index" json:"timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for SecurityEvent
func (SecurityEvent) TableName() string {
	return "security_events"
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

// AlertRule represents a rule that triggers notifications based on conditions
type AlertRule struct {
	ID     uuid.UUID     `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	UserID uuid.UUID     `gorm:"type:uuid;not null;index" json:"user_id"`
	Name   string        `gorm:"type:varchar(255);not null" json:"name"`
	Type   AlertRuleType `gorm:"type:varchar(50);not null;index" json:"type"`

	// Alert configuration
	Enabled     bool   `gorm:"not null;default:true" json:"enabled"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	// Condition configuration (stored as JSONB for flexibility)
	// For budget alerts: {"budget_id": "uuid", "threshold_percentage": 80}
	// For goal alerts: {"goal_id": "uuid", "milestone_percentage": 50}
	// For scheduled reports: {"frequency": "daily", "time": "09:00", "timezone": "UTC"}
	Conditions map[string]interface{} `gorm:"type:jsonb;not null" json:"conditions"`

	// Notification channels to use (can override user preferences)
	Channels []NotificationChannel `gorm:"type:varchar(100)" json:"channels,omitempty"` // comma-separated: "email,in_app"

	// Scheduling (for recurring alerts)
	Schedule        *string    `gorm:"type:varchar(100)" json:"schedule,omitempty"` // cron expression
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	NextTriggerAt   *time.Time `json:"next_trigger_at,omitempty"`

	// Metadata
	Metadata map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for AlertRule
func (AlertRule) TableName() string {
	return "alert_rules"
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

// NotificationAnalytics represents tracking data for sent notifications
type NotificationAnalytics struct {
	ID             uuid.UUID           `gorm:"type:uuid;default:uuidv7();primaryKey" json:"id"`
	NotificationID uuid.UUID           `gorm:"type:uuid;not null;uniqueIndex" json:"notification_id"`
	UserID         uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	Type           NotificationType    `gorm:"type:varchar(50);not null;index" json:"type"`
	Channel        NotificationChannel `gorm:"type:varchar(20);not null" json:"channel"`

	// Lifecycle tracking
	QueuedAt    time.Time  `gorm:"not null;index" json:"queued_at"`
	SentAt      *time.Time `gorm:"index" json:"sent_at,omitempty"`
	DeliveredAt *time.Time `gorm:"index" json:"delivered_at,omitempty"`
	ReadAt      *time.Time `gorm:"index" json:"read_at,omitempty"`
	ClickedAt   *time.Time `gorm:"index" json:"clicked_at,omitempty"`
	FailedAt    *time.Time `gorm:"index" json:"failed_at,omitempty"`

	// Status and errors
	Status        string  `gorm:"type:varchar(20);not null;index" json:"status"` // queued, sent, delivered, read, clicked, failed
	FailureReason *string `gorm:"type:text" json:"failure_reason,omitempty"`

	// Engagement metrics
	OpenCount  int `gorm:"default:0" json:"open_count"`
	ClickCount int `gorm:"default:0" json:"click_count"`

	// Metadata for tracking links, etc.
	Metadata map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for NotificationAnalytics
func (NotificationAnalytics) TableName() string {
	return "notification_analytics"
}
