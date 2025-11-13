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

// SecurityEventType represents different types of security events
type SecurityEventType string

const (
	EventUserRegistered       SecurityEventType = "USER_REGISTERED"
	EventLoginSuccess         SecurityEventType = "LOGIN_SUCCESS"
	EventLoginFailed          SecurityEventType = "LOGIN_FAILED"
	EventLogout               SecurityEventType = "LOGOUT"
	EventAccountLocked        SecurityEventType = "ACCOUNT_LOCKED"
	EventPasswordChanged      SecurityEventType = "PASSWORD_CHANGED"
	EventPasswordResetRequest SecurityEventType = "PASSWORD_RESET_REQUESTED"
	EventPasswordReset        SecurityEventType = "PASSWORD_RESET_COMPLETED"
	EventEmailVerified        SecurityEventType = "EMAIL_VERIFIED"
	EventGoogleOAuthLogin     SecurityEventType = "GOOGLE_OAUTH_LOGIN"
	EventTokenRefreshed       SecurityEventType = "TOKEN_REFRESHED"
)

// IsValid checks if the security event type is valid
func (set SecurityEventType) IsValid() bool {
	switch set {
	case EventUserRegistered,
		EventLoginSuccess,
		EventLoginFailed,
		EventLogout,
		EventAccountLocked,
		EventPasswordChanged,
		EventPasswordResetRequest,
		EventPasswordReset,
		EventEmailVerified,
		EventGoogleOAuthLogin,
		EventTokenRefreshed:
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

// AlertRuleType represents different types of alert rules
type AlertRuleType string

const (
	AlertRuleBudgetThreshold  AlertRuleType = "budget_threshold"
	AlertRuleGoalMilestone    AlertRuleType = "goal_milestone"
	AlertRuleGoalOffTrack     AlertRuleType = "goal_off_track"
	AlertRuleScheduledReport  AlertRuleType = "scheduled_report"
	AlertRuleLargeTransaction AlertRuleType = "large_transaction" // for future use
	AlertRuleUnusualActivity  AlertRuleType = "unusual_activity"  // for future use
)

// IsValid checks if the alert rule type is valid
func (art AlertRuleType) IsValid() bool {
	switch art {
	case AlertRuleBudgetThreshold,
		AlertRuleGoalMilestone,
		AlertRuleGoalOffTrack,
		AlertRuleScheduledReport,
		AlertRuleLargeTransaction,
		AlertRuleUnusualActivity:
		return true
	}
	return false
}
