package service

import (
	"log"
	"personalfinancedss/internal/module/notification/domain"
	"time"
)

type notificationService struct {
	emailService EmailService
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailService EmailService) NotificationService {
	return &notificationService{
		emailService: emailService,
	}
}

func (ns *notificationService) SendNotification(notification domain.NotificationData) error {
	switch notification.Type {
	case domain.NotificationTypeWelcome:
		return ns.sendWelcomeNotification(notification)
	case domain.NotificationTypeBudgetAlert:
		return ns.sendBudgetAlertNotification(notification)
	case domain.NotificationTypeGoalAchieved:
		return ns.sendGoalAchievedNotification(notification)
	case domain.NotificationTypeMonthlySummary:
		return ns.sendMonthlySummaryNotification(notification)
	default:
		log.Printf("Unknown notification type: %s", notification.Type)
		return nil
	}
}

func (ns *notificationService) sendWelcomeNotification(notification domain.NotificationData) error {
	username, ok := notification.Data["username"].(string)
	if !ok {
		username = "User"
	}

	return ns.emailService.SendWelcomeEmail(
		notification.UserEmail,
		notification.UserName,
		username,
	)
}

func (ns *notificationService) sendBudgetAlertNotification(notification domain.NotificationData) error {
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

	return ns.emailService.SendBudgetAlert(
		notification.UserEmail,
		notification.UserName,
		category,
		budgetAmount,
		spentAmount,
		threshold,
	)
}

func (ns *notificationService) sendGoalAchievedNotification(notification domain.NotificationData) error {
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

	// Get target date and achieved date, use notification timestamp as fallback
	targetDate, ok := notification.Data["target_date"].(time.Time)
	if !ok {
		targetDate = notification.Timestamp
	}

	achievedDate, ok := notification.Data["achieved_date"].(time.Time)
	if !ok {
		achievedDate = notification.Timestamp
	}

	return ns.emailService.SendGoalAchievedEmail(
		notification.UserEmail,
		notification.UserName,
		goalName,
		targetAmount,
		currentAmount,
		targetDate,
		achievedDate,
	)
}

func (ns *notificationService) sendMonthlySummaryNotification(notification domain.NotificationData) error {
	// Extract month and year from data
	month, ok := notification.Data["month"].(string)
	if !ok {
		month = notification.Timestamp.Format("January")
	}

	year, ok := notification.Data["year"].(int)
	if !ok {
		year = notification.Timestamp.Year()
	}

	return ns.emailService.SendMonthlySummary(
		notification.UserEmail,
		notification.UserName,
		month,
		year,
		notification.Data,
	)
}

// Convenience methods for cashflow notifications
func (ns *notificationService) NotifyUserRegistration(userEmail, userName, username string) {
	notification := domain.NotificationData{
		Type:      domain.NotificationTypeWelcome,
		UserEmail: userEmail,
		UserName:  userName,
		Data: map[string]interface{}{
			"username": username,
		},
		Timestamp: time.Now(),
	}

	if err := ns.SendNotification(notification); err != nil {
		log.Printf("Failed to send welcome notification: %v", err)
	}
}

func (ns *notificationService) NotifyBudgetAlert(userEmail, userName, category string, budgetAmount, spentAmount, threshold float64) {
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

	if err := ns.SendNotification(notification); err != nil {
		log.Printf("Failed to send budget alert notification: %v", err)
	}
}

func (ns *notificationService) NotifyGoalAchieved(userEmail, userName, goalName string, targetAmount, currentAmount float64, targetDate, achievedDate time.Time) {
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

	if err := ns.SendNotification(notification); err != nil {
		log.Printf("Failed to send goal achieved notification: %v", err)
	}
}

func (ns *notificationService) NotifyMonthlySummary(userEmail, userName, month string, year int, summary map[string]interface{}) {
	// Merge month and year into summary data
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

	if err := ns.SendNotification(notification); err != nil {
		log.Printf("Failed to send monthly summary notification: %v", err)
	}
}
