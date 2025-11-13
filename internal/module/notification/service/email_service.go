package service

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"personalfinancedss/internal/module/notification/domain"
	"time"

	"go.uber.org/zap"
)

type emailService struct {
	config domain.EmailConfig
	logger *zap.Logger
}

// NewEmailService creates a new email service with configuration
func NewEmailService(config domain.EmailConfig, logger *zap.Logger) EmailService {
	return &emailService{
		config: config,
		logger: logger,
	}
}

func (es *emailService) SendEmailFromTemplate(to, subject, templatePath string, data map[string]interface{}) error {
	// Development mode - just log
	if es.config.SMTPUsername == "" || es.config.SMTPPassword == "" {
		es.logger.Info(
			"Email skipped (dev mode)",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.String("template", templatePath),
			zap.Any("data", data),
		)
		return nil
	}

	// Read template file
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read email template: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New("email").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Create email message
	from := fmt.Sprintf("%s <%s>", es.config.FromName, es.config.FromEmail)
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body.String()

	// Send email
	auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", es.config.SMTPHost, es.config.SMTPPort)

	return smtp.SendMail(addr, auth, es.config.FromEmail, []string{to}, []byte(message))
}

func (es *emailService) SendVerificationEmail(to, name, token string) error {
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", es.config.FrontendURL, token)

	data := map[string]interface{}{
		"Name":             name,
		"VerificationLink": verificationLink,
	}

	return es.SendEmailFromTemplate(
		to,
		"Verify Your Email Address - Personal Finance DSS",
		"resource/email_templates/verify_email.html",
		data,
	)
}

func (es *emailService) SendPasswordResetEmail(to, name, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", es.config.FrontendURL, token)

	data := map[string]interface{}{
		"Name":      name,
		"ResetLink": resetLink,
	}

	return es.SendEmailFromTemplate(
		to,
		"Reset Your Password - Personal Finance DSS",
		"resource/email_templates/reset_password.html",
		data,
	)
}

func (es *emailService) SendWelcomeEmail(to, name, email string) error {
	loginURL := fmt.Sprintf("%s/login", es.config.FrontendURL)

	data := map[string]interface{}{
		"Name":             name,
		"Email":            email,
		"RegistrationDate": time.Now().Format("January 2, 2006"),
		"LoginURL":         loginURL,
	}

	return es.SendEmailFromTemplate(
		to,
		"Welcome to Personal Finance DSS!",
		"resource/email_templates/welcome.html",
		data,
	)
}

func (es *emailService) SendBudgetAlert(to, name, category string, budgetAmount, spentAmount, threshold float64) error {
	percentage := (spentAmount / budgetAmount) * 100
	remaining := budgetAmount - spentAmount
	exceeded := percentage >= 100
	dashboardURL := fmt.Sprintf("%s/dashboard", es.config.FrontendURL)

	data := map[string]interface{}{
		"Name":            name,
		"Category":        category,
		"BudgetAmount":    fmt.Sprintf("%.2f", budgetAmount),
		"SpentAmount":     fmt.Sprintf("%.2f", spentAmount),
		"RemainingAmount": fmt.Sprintf("%.2f", remaining),
		"Percentage":      fmt.Sprintf("%.1f", percentage),
		"Threshold":       fmt.Sprintf("%.0f", threshold),
		"Exceeded":        exceeded,
		"DashboardURL":    dashboardURL,
	}

	return es.SendEmailFromTemplate(
		to,
		fmt.Sprintf("Budget Alert: %s", category),
		"resource/email_templates/budget_alert.html",
		data,
	)
}

func (es *emailService) SendGoalAchievedEmail(to, name, goalName string, targetAmount, currentAmount float64, targetDate, achievedDate time.Time) error {
	goalsURL := fmt.Sprintf("%s/goals", es.config.FrontendURL)

	data := map[string]interface{}{
		"Name":          name,
		"GoalName":      goalName,
		"TargetAmount":  fmt.Sprintf("%.2f", targetAmount),
		"CurrentAmount": fmt.Sprintf("%.2f", currentAmount),
		"TargetDate":    targetDate.Format("January 2, 2006"),
		"AchievedDate":  achievedDate.Format("January 2, 2006"),
		"GoalsURL":      goalsURL,
	}

	return es.SendEmailFromTemplate(
		to,
		fmt.Sprintf("🎉 Goal Achieved: %s", goalName),
		"resource/email_templates/goal_achieved.html",
		data,
	)
}

func (es *emailService) SendMonthlySummary(to, name, month string, year int, summary map[string]interface{}) error {
	dashboardURL := fmt.Sprintf("%s/dashboard", es.config.FrontendURL)

	data := map[string]interface{}{
		"Name":          name,
		"Month":         month,
		"Year":          year,
		"TotalIncome":   fmt.Sprintf("%.2f", summary["total_income"]),
		"TotalExpenses": fmt.Sprintf("%.2f", summary["total_expenses"]),
		"NetSavings":    fmt.Sprintf("%.2f", summary["net_savings"]),
		"SavingsRate":   fmt.Sprintf("%.1f", summary["savings_rate"]),
		"HealthScore":   summary["health_score"],
		"Grade":         summary["grade"],
		"TopCategories": summary["top_categories"],
		"BudgetAlerts":  summary["budget_alerts"],
		"DashboardURL":  dashboardURL,
	}

	return es.SendEmailFromTemplate(
		to,
		"Monthly Financial Summary",
		"resource/email_templates/monthly_summary.html",
		data,
	)
}

func (es *emailService) SendCustomEmail(to, subject, templateFileName string, data map[string]interface{}) error {
	templatePath := filepath.Join("resource", "email_templates", templateFileName)
	return es.SendEmailFromTemplate(to, subject, templatePath, data)
}
