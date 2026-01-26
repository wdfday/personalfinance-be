package dto

import "personalfinancedss/internal/module/notification/domain"

// SendNotificationRequest is the request for sending a notification
type SendNotificationRequest struct {
	Type     string                 `json:"type" binding:"required"`    // notification type
	UserID   string                 `json:"user_id" binding:"required"` // target user ID
	Channels []string               `json:"channels"`                   // delivery channels: email, in_app (default: both)
	Subject  string                 `json:"subject"`                    // for email notifications
	Data     map[string]interface{} `json:"data"`                       // template data
}

// MarkAsReadRequest is the request for marking notification as read
type MarkAsReadRequest struct {
	NotificationID string `json:"notification_id" binding:"required"`
}

// ListNotificationsRequest is the request for listing notifications
type ListNotificationsRequest struct {
	Limit  int  `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int  `form:"offset" binding:"omitempty,min=0"`
	Unread bool `form:"unread"` // filter only unread notifications
}

// Helper to convert request type to domain type
func (r *SendNotificationRequest) ToNotificationType() (domain.NotificationType, error) {
	notifType := domain.NotificationType(r.Type)
	if !notifType.IsValid() {
		return "", domain.ErrInvalidNotificationType
	}
	return notifType, nil
}

// Helper to get channels with defaults
func (r *SendNotificationRequest) GetChannels() []domain.NotificationChannel {
	// Default to both email and in-app
	if len(r.Channels) == 0 {
		return []domain.NotificationChannel{
			domain.ChannelEmail,
			domain.ChannelInApp,
		}
	}

	channels := make([]domain.NotificationChannel, 0, len(r.Channels))
	for _, ch := range r.Channels {
		channel := domain.NotificationChannel(ch)
		if channel.IsValid() {
			channels = append(channels, channel)
		}
	}
	return channels
}

// CreateNotificationPreferenceRequest is the request for creating a notification preference
type CreateNotificationPreferenceRequest struct {
	Type              string   `json:"type" binding:"required"`                           // notification type
	Enabled           bool     `json:"enabled"`                                           // default: true
	PreferredChannels []string `json:"preferred_channels"`                                // email, in_app, sms, push
	MinInterval       int      `json:"min_interval_minutes"`                              // minimum minutes between notifications
	QuietHoursFrom    *int     `json:"quiet_hours_from" binding:"omitempty,min=0,max=23"` // hour 0-23
	QuietHoursTo      *int     `json:"quiet_hours_to" binding:"omitempty,min=0,max=23"`   // hour 0-23
	Timezone          string   `json:"timezone"`                                          // default: UTC
}

// UpdateNotificationPreferenceRequest is the request for updating a notification preference
type UpdateNotificationPreferenceRequest struct {
	Enabled           *bool    `json:"enabled"`
	PreferredChannels []string `json:"preferred_channels"`
	MinInterval       *int     `json:"min_interval_minutes"`
	QuietHoursFrom    *int     `json:"quiet_hours_from" binding:"omitempty,min=0,max=23"`
	QuietHoursTo      *int     `json:"quiet_hours_to" binding:"omitempty,min=0,max=23"`
	Timezone          *string  `json:"timezone"`
}
