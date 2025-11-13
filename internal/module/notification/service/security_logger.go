package service

import (
	"context"
	"log"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/repository"
	"time"

	"github.com/google/uuid"
)

type securityLogger struct {
	repo repository.SecurityEventRepository
}

// NewSecurityLogger creates a new security logger
func NewSecurityLogger(repo repository.SecurityEventRepository) SecurityLogger {
	return &securityLogger{
		repo: repo,
	}
}

func (l *securityLogger) LogEvent(ctx context.Context, event domain.SecurityEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Persist to database
	if err := l.repo.Create(ctx, &event); err != nil {
		log.Printf("[ERROR] Failed to persist security event: %v", err)
	}

	// Format log message for stdout
	logMsg := formatSecurityEvent(event)

	// Log to stdout
	if event.Success {
		log.Printf("[SECURITY] %s", logMsg)
	} else {
		log.Printf("[SECURITY_ALERT] %s", logMsg)
	}
}

func (l *securityLogger) LogRegistration(ctx context.Context, userID, email, ipAddress string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventUserRegistered,
		UserID:    &uid,
		Email:     email,
		IPAddress: ipAddress,
		Success:   true,
		Message:   "New user registered",
	})
}

func (l *securityLogger) LogLoginSuccess(ctx context.Context, userID, email, ipAddress string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventLoginSuccess,
		UserID:    &uid,
		Email:     email,
		IPAddress: ipAddress,
		Success:   true,
		Message:   "User logged in successfully",
	})
}

func (l *securityLogger) LogLogout(ctx context.Context, userID, email, ipAddress string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventLogout,
		UserID:    &uid,
		Email:     email,
		IPAddress: ipAddress,
		Success:   true,
		Message:   "User logged out",
	})
}

func (l *securityLogger) LogLoginFailed(ctx context.Context, email, ipAddress, reason string) {
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventLoginFailed,
		Email:     email,
		IPAddress: ipAddress,
		Success:   false,
		Message:   "Login failed: " + reason,
	})
}

func (l *securityLogger) LogAccountLocked(ctx context.Context, userID, email, ipAddress string, lockedUntil time.Time) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventAccountLocked,
		UserID:    &uid,
		Email:     email,
		IPAddress: ipAddress,
		Success:   false,
		Message:   "Account locked due to too many failed login attempts",
		Metadata: map[string]interface{}{
			"locked_until": lockedUntil,
		},
	})
}

func (l *securityLogger) LogPasswordChanged(ctx context.Context, userID, email string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:    domain.EventPasswordChanged,
		UserID:  &uid,
		Email:   email,
		Success: true,
		Message: "User changed password",
	})
}

func (l *securityLogger) LogPasswordResetRequest(ctx context.Context, email, ipAddress string) {
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventPasswordResetRequest,
		Email:     email,
		IPAddress: ipAddress,
		Success:   true,
		Message:   "Password reset requested",
	})
}

func (l *securityLogger) LogPasswordReset(ctx context.Context, userID, email string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:    domain.EventPasswordReset,
		UserID:  &uid,
		Email:   email,
		Success: true,
		Message: "Password reset completed",
	})
}

func (l *securityLogger) LogEmailVerified(ctx context.Context, userID, email string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:    domain.EventEmailVerified,
		UserID:  &uid,
		Email:   email,
		Success: true,
		Message: "Email verified",
	})
}

func (l *securityLogger) LogGoogleOAuthLogin(ctx context.Context, userID, email, ipAddress string, isNewUser bool) {
	message := "Google OAuth login"
	if isNewUser {
		message = "New user registered via Google OAuth"
	}

	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:      domain.EventGoogleOAuthLogin,
		UserID:    &uid,
		Email:     email,
		IPAddress: ipAddress,
		Success:   true,
		Message:   message,
		Metadata: map[string]interface{}{
			"is_new_user": isNewUser,
		},
	})
}

func (l *securityLogger) LogTokenRefreshed(ctx context.Context, userID, email string) {
	uid, _ := uuid.Parse(userID)
	l.LogEvent(ctx, domain.SecurityEvent{
		Type:    domain.EventTokenRefreshed,
		UserID:  &uid,
		Email:   email,
		Success: true,
		Message: "Access token refreshed",
	})
}

// Helper function to format security event for logging
func formatSecurityEvent(event domain.SecurityEvent) string {
	msg := event.Timestamp.Format("2006-01-02 15:04:05")
	msg += " | " + string(event.Type)

	if event.UserID != nil {
		msg += " | UserID: " + event.UserID.String()
	}
	if event.Email != "" {
		msg += " | Email: " + event.Email
	}
	if event.IPAddress != "" {
		msg += " | IP: " + event.IPAddress
	}

	msg += " | " + event.Message

	if len(event.Metadata) > 0 {
		msg += " | Metadata: "
		for k, v := range event.Metadata {
			msg += k + "=" + toString(v) + " "
		}
	}

	return msg
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return ""
	}
}
