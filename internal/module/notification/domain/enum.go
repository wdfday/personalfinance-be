package domain

// NotificationType represents the type of notification to send
type NotificationType string

const (
	NotificationTypeWelcome           NotificationType = "welcome"
	NotificationTypeBudgetAlert       NotificationType = "budget_alert"
	NotificationTypeGoalAchieved      NotificationType = "goal_achieved"
	NotificationTypeMonthlySummary    NotificationType = "monthly_summary"
	NotificationTypeEmailVerification NotificationType = "email_verification"
	NotificationTypePasswordReset     NotificationType = "password_reset"
)

// IsValid checks if the notification type is valid
func (nt NotificationType) IsValid() bool {
	switch nt {
	case NotificationTypeWelcome,
		NotificationTypeBudgetAlert,
		NotificationTypeGoalAchieved,
		NotificationTypeMonthlySummary,
		NotificationTypeEmailVerification,
		NotificationTypePasswordReset:
		return true
	}
	return false
}

// NotificationChannel represents delivery channels for notifications
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelInApp NotificationChannel = "in_app"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

// IsValid checks if the notification channel is valid
func (nc NotificationChannel) IsValid() bool {
	switch nc {
	case ChannelEmail, ChannelInApp, ChannelSMS, ChannelPush:
		return true
	}
	return false
}
