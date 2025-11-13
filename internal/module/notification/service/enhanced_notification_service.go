package service

import (
	"context"
	"fmt"
	"log"
	"personalfinancedss/internal/module/notification/domain"
	"personalfinancedss/internal/module/notification/repository"
	"time"

	"github.com/google/uuid"
)

// EnhancedNotificationService handles both in-app and email notifications
type EnhancedNotificationService struct {
	emailService EmailService
	notifRepo    repository.NotificationRepository
	wsHub        *WebSocketHub // Optional: nil if WebSocket not enabled
}

// NewEnhancedNotificationService creates an enhanced notification service
func NewEnhancedNotificationService(
	emailService EmailService,
	notifRepo repository.NotificationRepository,
	wsHub *WebSocketHub,
) NotificationService {
	return &EnhancedNotificationService{
		emailService: emailService,
		notifRepo:    notifRepo,
		wsHub:        wsHub,
	}
}

// SendNotificationMultiChannel sends notification through specified channels
func (s *EnhancedNotificationService) SendNotificationMultiChannel(
	ctx context.Context,
	userID uuid.UUID,
	userEmail string,
	userName string,
	notifType domain.NotificationType,
	subject string,
	data map[string]interface{},
	channels []domain.NotificationChannel,
) error {
	var lastErr error

	// Send through each channel
	for _, channel := range channels {
		switch channel {
		case domain.ChannelInApp:
			if err := s.sendInAppNotification(ctx, userID, notifType, subject, data); err != nil {
				log.Printf("Failed to send in-app notification: %v", err)
				lastErr = err
			}

		case domain.ChannelEmail:
			// Convert to NotificationData and send email
			notifData := domain.NotificationData{
				Type:      notifType,
				UserEmail: userEmail,
				UserName:  userName,
				Data:      data,
				Timestamp: time.Now(),
			}
			if err := s.SendNotification(notifData); err != nil {
				log.Printf("Failed to send email notification: %v", err)
				lastErr = err
			}
		}
	}

	return lastErr
}

// sendInAppNotification creates in-app notification record
func (s *EnhancedNotificationService) sendInAppNotification(
	ctx context.Context,
	userID uuid.UUID,
	notifType domain.NotificationType,
	subject string,
	data map[string]interface{},
) error {
	notification := &domain.Notification{
		UserID:       userID,
		Type:         notifType,
		Channel:      domain.ChannelInApp,
		Recipient:    userID.String(), // For in-app, recipient is user ID
		Subject:      subject,
		Data:         data,
		Status:       "pending", // pending = unread
		TemplateName: string(notifType),
	}

	// Save to database
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		return err
	}

	// Send via WebSocket if user is connected
	if s.wsHub != nil && s.wsHub.IsUserConnected(userID) {
		payload := domain.WebSocketNotificationPayload{
			ID:        notification.ID.String(),
			Type:      string(notification.Type),
			Subject:   notification.Subject,
			Data:      notification.Data,
			CreatedAt: notification.CreatedAt,
		}
		s.wsHub.SendToUser(userID, "notification", payload)
		log.Printf("Sent notification via WebSocket to user %s", userID)
	}

	return nil
}

// SendNotification sends email notification (backward compatibility)
func (s *EnhancedNotificationService) SendNotification(notification domain.NotificationData) error {
	switch notification.Type {
	case domain.NotificationTypeWelcome:
		return s.sendWelcomeNotification(notification)
	case domain.NotificationTypeBudgetAlert:
		return s.sendBudgetAlertNotification(notification)
	case domain.NotificationTypeGoalAchieved:
		return s.sendGoalAchievedNotification(notification)
	case domain.NotificationTypeMonthlySummary:
		return s.sendMonthlySummaryNotification(notification)
	case domain.NotificationTypeEmailVerification:
		return s.sendEmailVerification(notification)
	case domain.NotificationTypePasswordReset:
		return s.sendPasswordReset(notification)
	default:
		log.Printf("Unknown notification type: %s", notification.Type)
		return nil
	}
}

func (s *EnhancedNotificationService) sendWelcomeNotification(notification domain.NotificationData) error {
	username, ok := notification.Data["username"].(string)
	if !ok {
		username = "User"
	}

	return s.emailService.SendWelcomeEmail(
		notification.UserEmail,
		notification.UserName,
		username,
	)
}

func (s *EnhancedNotificationService) sendBudgetAlertNotification(notification domain.NotificationData) error {
	category, ok := notification.Data["category"].(string)
	if !ok {
		return nil
	}

	budgetAmount, ok := notification.Data["budget_amount"].(float64)
	if !ok {
		return nil
	}

	spentAmount, ok := notification.Data["spent_amount"].(float64)
	if !ok {
		return nil
	}

	threshold, ok := notification.Data["threshold"].(float64)
	if !ok {
		threshold = 80.0
	}

	return s.emailService.SendBudgetAlert(
		notification.UserEmail,
		notification.UserName,
		category,
		budgetAmount,
		spentAmount,
		threshold,
	)
}

func (s *EnhancedNotificationService) sendGoalAchievedNotification(notification domain.NotificationData) error {
	goalName, ok := notification.Data["goal_name"].(string)
	if !ok {
		return nil
	}

	targetAmount, ok := notification.Data["target_amount"].(float64)
	if !ok {
		return nil
	}

	currentAmount, ok := notification.Data["current_amount"].(float64)
	if !ok {
		return nil
	}

	targetDate, ok := notification.Data["target_date"].(time.Time)
	if !ok {
		targetDate = notification.Timestamp
	}

	achievedDate, ok := notification.Data["achieved_date"].(time.Time)
	if !ok {
		achievedDate = notification.Timestamp
	}

	return s.emailService.SendGoalAchievedEmail(
		notification.UserEmail,
		notification.UserName,
		goalName,
		targetAmount,
		currentAmount,
		targetDate,
		achievedDate,
	)
}

func (s *EnhancedNotificationService) sendMonthlySummaryNotification(notification domain.NotificationData) error {
	month, ok := notification.Data["month"].(string)
	if !ok {
		month = notification.Timestamp.Format("January")
	}

	year, ok := notification.Data["year"].(int)
	if !ok {
		year = notification.Timestamp.Year()
	}

	return s.emailService.SendMonthlySummary(
		notification.UserEmail,
		notification.UserName,
		month,
		year,
		notification.Data,
	)
}

func (s *EnhancedNotificationService) sendEmailVerification(notification domain.NotificationData) error {
	token, ok := notification.Data["token"].(string)
	if !ok {
		return fmt.Errorf("missing token in notification data")
	}

	return s.emailService.SendVerificationEmail(
		notification.UserEmail,
		notification.UserName,
		token,
	)
}

func (s *EnhancedNotificationService) sendPasswordReset(notification domain.NotificationData) error {
	token, ok := notification.Data["token"].(string)
	if !ok {
		return fmt.Errorf("missing token in notification data")
	}

	return s.emailService.SendPasswordResetEmail(
		notification.UserEmail,
		notification.UserName,
		token,
	)
}

// Convenience methods
func (s *EnhancedNotificationService) NotifyUserRegistration(userEmail, userName, username string) {
	notification := domain.NotificationData{
		Type:      domain.NotificationTypeWelcome,
		UserEmail: userEmail,
		UserName:  userName,
		Data: map[string]interface{}{
			"username": username,
		},
		Timestamp: time.Now(),
	}

	if err := s.SendNotification(notification); err != nil {
		log.Printf("Failed to send welcome notification: %v", err)
	}
}

func (s *EnhancedNotificationService) NotifyBudgetAlert(userEmail, userName, category string, budgetAmount, spentAmount, threshold float64) {
	notification := domain.NotificationData{
		Type:      domain.NotificationTypeBudgetAlert,
		UserEmail: userEmail,
		UserName:  userName,
		Data: map[string]interface{}{
			"category":      category,
			"budget_amount": budgetAmount,
			"spent_amount":  spentAmount,
			"threshold":     threshold,
		},
		Timestamp: time.Now(),
	}

	if err := s.SendNotification(notification); err != nil {
		log.Printf("Failed to send budget alert notification: %v", err)
	}
}

func (s *EnhancedNotificationService) NotifyGoalAchieved(userEmail, userName, goalName string, targetAmount, currentAmount float64, targetDate, achievedDate time.Time) {
	notification := domain.NotificationData{
		Type:      domain.NotificationTypeGoalAchieved,
		UserEmail: userEmail,
		UserName:  userName,
		Data: map[string]interface{}{
			"goal_name":      goalName,
			"target_amount":  targetAmount,
			"current_amount": currentAmount,
			"target_date":    targetDate,
			"achieved_date":  achievedDate,
		},
		Timestamp: time.Now(),
	}

	if err := s.SendNotification(notification); err != nil {
		log.Printf("Failed to send goal achieved notification: %v", err)
	}
}

func (s *EnhancedNotificationService) NotifyMonthlySummary(userEmail, userName, month string, year int, summary map[string]interface{}) {
	if summary == nil {
		summary = make(map[string]interface{})
	}
	summary["month"] = month
	summary["year"] = year

	notification := domain.NotificationData{
		Type:      domain.NotificationTypeMonthlySummary,
		UserEmail: userEmail,
		UserName:  userName,
		Data:      summary,
		Timestamp: time.Now(),
	}

	if err := s.SendNotification(notification); err != nil {
		log.Printf("Failed to send monthly summary notification: %v", err)
	}
}
