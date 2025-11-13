package dto

import (
	"personalfinancedss/internal/module/notification/domain"
	"time"
)

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Subject      string                 `json:"subject,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Status       string                 `json:"status"` // pending, read, sent, failed
	SentAt       *time.Time             `json:"sent_at,omitempty"`
	FailedAt     *time.Time             `json:"failed_at,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// NotificationListResponse represents a list of notifications
type NotificationListResponse struct {
	Items      []NotificationResponse `json:"items"`
	Total      int64                  `json:"total"`
	Unread     int64                  `json:"unread"`
	Page       int                    `json:"page"`
	PerPage    int                    `json:"per_page"`
	TotalPages int                    `json:"total_pages"`
}

// NotificationStatsResponse represents notification statistics
type NotificationStatsResponse struct {
	TotalNotifications int64 `json:"total_notifications"`
	UnreadCount        int64 `json:"unread_count"`
	ReadCount          int64 `json:"read_count"`
}

// ToNotificationResponse converts domain.Notification to NotificationResponse
func ToNotificationResponse(n domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:           n.ID.String(),
		Type:         string(n.Type),
		Subject:      n.Subject,
		Data:         n.Data,
		Status:       n.Status,
		SentAt:       n.SentAt,
		FailedAt:     n.FailedAt,
		ErrorMessage: n.ErrorMessage,
		CreatedAt:    n.CreatedAt,
	}
}

// ToNotificationListResponse converts a list of domain.Notification to NotificationListResponse
func ToNotificationListResponse(notifications []domain.Notification, total, unread int64, page, perPage int) NotificationListResponse {
	items := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		items[i] = ToNotificationResponse(n)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return NotificationListResponse{
		Items:      items,
		Total:      total,
		Unread:     unread,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// NotificationPreferenceResponse represents a notification preference in API responses
type NotificationPreferenceResponse struct {
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	Enabled           bool      `json:"enabled"`
	PreferredChannels []string  `json:"preferred_channels"`
	MinInterval       int       `json:"min_interval_minutes"`
	QuietHoursFrom    *int      `json:"quiet_hours_from,omitempty"`
	QuietHoursTo      *int      `json:"quiet_hours_to,omitempty"`
	Timezone          string    `json:"timezone"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AlertRuleResponse represents an alert rule in API responses
type AlertRuleResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Enabled         bool                   `json:"enabled"`
	Description     string                 `json:"description"`
	Conditions      map[string]interface{} `json:"conditions"`
	Channels        []string               `json:"channels,omitempty"`
	Schedule        *string                `json:"schedule,omitempty"`
	LastTriggeredAt *time.Time             `json:"last_triggered_at,omitempty"`
	NextTriggerAt   *time.Time             `json:"next_trigger_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// NotificationAnalyticsResponse represents notification analytics in API responses
type NotificationAnalyticsResponse struct {
	ID             string                 `json:"id"`
	NotificationID string                 `json:"notification_id"`
	Type           string                 `json:"type"`
	Channel        string                 `json:"channel"`
	QueuedAt       time.Time              `json:"queued_at"`
	SentAt         *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time             `json:"delivered_at,omitempty"`
	ReadAt         *time.Time             `json:"read_at,omitempty"`
	ClickedAt      *time.Time             `json:"clicked_at,omitempty"`
	FailedAt       *time.Time             `json:"failed_at,omitempty"`
	Status         string                 `json:"status"`
	FailureReason  *string                `json:"failure_reason,omitempty"`
	OpenCount      int                    `json:"open_count"`
	ClickCount     int                    `json:"click_count"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ToNotificationPreferenceResponse converts domain.NotificationPreference to NotificationPreferenceResponse
func ToNotificationPreferenceResponse(p domain.NotificationPreference) NotificationPreferenceResponse {
	channels := make([]string, len(p.PreferredChannels))
	for i, ch := range p.PreferredChannels {
		channels[i] = string(ch)
	}

	return NotificationPreferenceResponse{
		ID:                p.ID.String(),
		Type:              string(p.Type),
		Enabled:           p.Enabled,
		PreferredChannels: channels,
		MinInterval:       p.MinInterval,
		QuietHoursFrom:    p.QuietHoursFrom,
		QuietHoursTo:      p.QuietHoursTo,
		Timezone:          p.Timezone,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// ToAlertRuleResponse converts domain.AlertRule to AlertRuleResponse
func ToAlertRuleResponse(r domain.AlertRule) AlertRuleResponse {
	channels := make([]string, len(r.Channels))
	for i, ch := range r.Channels {
		channels[i] = string(ch)
	}

	return AlertRuleResponse{
		ID:              r.ID.String(),
		Name:            r.Name,
		Type:            string(r.Type),
		Enabled:         r.Enabled,
		Description:     r.Description,
		Conditions:      r.Conditions,
		Channels:        channels,
		Schedule:        r.Schedule,
		LastTriggeredAt: r.LastTriggeredAt,
		NextTriggerAt:   r.NextTriggerAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// ToNotificationAnalyticsResponse converts domain.NotificationAnalytics to NotificationAnalyticsResponse
func ToNotificationAnalyticsResponse(a domain.NotificationAnalytics) NotificationAnalyticsResponse {
	return NotificationAnalyticsResponse{
		ID:             a.ID.String(),
		NotificationID: a.NotificationID.String(),
		Type:           string(a.Type),
		Channel:        string(a.Channel),
		QueuedAt:       a.QueuedAt,
		SentAt:         a.SentAt,
		DeliveredAt:    a.DeliveredAt,
		ReadAt:         a.ReadAt,
		ClickedAt:      a.ClickedAt,
		FailedAt:       a.FailedAt,
		Status:         a.Status,
		FailureReason:  a.FailureReason,
		OpenCount:      a.OpenCount,
		ClickCount:     a.ClickCount,
		Metadata:       a.Metadata,
	}
}
