package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Debt represents a debt obligation
type Debt struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_id"`

	// Debt Details
	Name        string     `gorm:"type:varchar(255);not null;column:name" json:"name"`
	Description *string    `gorm:"type:text;column:description" json:"description,omitempty"`
	Type        DebtType   `gorm:"type:varchar(50);not null;column:type" json:"type"`
	Status      DebtStatus `gorm:"type:varchar(20);default:'active';column:status" json:"status"`

	// Financial Details
	PrincipalAmount float64 `gorm:"type:decimal(15,2);not null;column:principal_amount" json:"principal_amount"` // Original debt amount
	CurrentBalance  float64 `gorm:"type:decimal(15,2);not null;column:current_balance" json:"current_balance"`   // Current remaining balance
	InterestRate    float64 `gorm:"type:decimal(5,2);default:0;column:interest_rate" json:"interest_rate"`       // Annual interest rate (%)
	MinimumPayment  float64 `gorm:"type:decimal(15,2);default:0;column:minimum_payment" json:"minimum_payment"`  // Minimum monthly payment
	PaymentAmount   float64 `gorm:"type:decimal(15,2);default:0;column:payment_amount" json:"payment_amount"`    // Actual payment amount
	Currency        string  `gorm:"type:varchar(3);default:'VND';column:currency" json:"currency"`

	// Payment Information
	PaymentFrequency  *PaymentFrequency `gorm:"type:varchar(20);column:payment_frequency" json:"payment_frequency,omitempty"`
	NextPaymentDate   *time.Time        `gorm:"type:date;column:next_payment_date" json:"next_payment_date,omitempty"`
	LastPaymentDate   *time.Time        `gorm:"type:date;column:last_payment_date" json:"last_payment_date,omitempty"`
	LastPaymentAmount *float64          `gorm:"type:decimal(15,2);column:last_payment_amount" json:"last_payment_amount,omitempty"`

	// Timeline
	StartDate   time.Time  `gorm:"type:date;not null;column:start_date" json:"start_date"`
	DueDate     *time.Time `gorm:"type:date;column:due_date" json:"due_date,omitempty"`
	PaidOffDate *time.Time `gorm:"type:date;column:paid_off_date" json:"paid_off_date,omitempty"`

	// Progress Tracking
	TotalPaid         float64 `gorm:"type:decimal(15,2);default:0;column:total_paid" json:"total_paid"`                   // Total amount paid
	RemainingAmount   float64 `gorm:"type:decimal(15,2);default:0;column:remaining_amount" json:"remaining_amount"`       // Remaining to pay
	PercentagePaid    float64 `gorm:"type:decimal(5,2);default:0;column:percentage_paid" json:"percentage_paid"`          // Percentage paid off
	TotalInterestPaid float64 `gorm:"type:decimal(15,2);default:0;column:total_interest_paid" json:"total_interest_paid"` // Total interest paid

	// Linked Resources
	CreditorName    *string    `gorm:"type:varchar(255);column:creditor_name" json:"creditor_name,omitempty"`       // Name of creditor
	AccountNumber   *string    `gorm:"type:varchar(100);column:account_number" json:"account_number,omitempty"`     // Account number with creditor
	LinkedAccountID *uuid.UUID `gorm:"type:uuid;index;column:linked_account_id" json:"linked_account_id,omitempty"` // Account used for payments

	// Notifications
	EnableReminders    bool       `gorm:"default:true;column:enable_reminders" json:"enable_reminders"`
	ReminderFrequency  *string    `gorm:"type:varchar(20);column:reminder_frequency" json:"reminder_frequency,omitempty"` // "weekly", "monthly"
	LastReminderSentAt *time.Time `gorm:"column:last_reminder_sent_at" json:"last_reminder_sent_at,omitempty"`

	// Metadata
	Notes *string `gorm:"type:text;column:notes" json:"notes,omitempty"`
	Tags  *string `gorm:"type:jsonb;column:tags" json:"tags,omitempty"` // JSON array of tags

	// DSS Input/Output Metadata - Both constraint (INPUT) and allocation (OUTPUT)
	DSSMetadata datatypes.JSON `gorm:"type:jsonb;column:dss_metadata" json:"dss_metadata,omitempty"`
	// Structure:
	// {
	//   // INPUT - User constraints
	//   "minimum_payment": 2000000,           // Minimum required payment (INPUT)
	//   "target_payment": 3000000,            // Target payment (INPUT)
	//   "maximum_payment": 5000000,           // Max affordable payment (INPUT)
	//   "constraint_type": "hard",            // hard (minimum), soft (target) (INPUT)
	//   "is_flexible": true,                  // Can adjust payment (INPUT)
	//   "gp_priority": 1,                     // P1, P2, P3... (high priority) (INPUT)
	//   "gp_weight": 0.95,                    // Weight (interest rate based) (INPUT)
	//
	//   // OUTPUT - DSS allocation
	//   "allocated_payment": 3200000,         // DSS allocated payment (OUTPUT)
	//   "satisfaction_level": 0.80,           // 0-1, payment achievement level (OUTPUT)
	//   "negative_deviation": 0,              // d⁻ (under target) (OUTPUT)
	//   "positive_deviation": 200000,         // d⁺ (over target - good!) (OUTPUT)
	//   "achievement_rate": 1.07,             // Allocated/Target ratio (OUTPUT)
	//   "allocation_reason": "high_priority_debt_payoff",
	//
	//   // ANALYSIS
	//   "monthly_interest_cost": 500000,      // Interest per month
	//   "monthly_principal_payment": 2700000, // Principal per month (allocated - interest)
	//   "estimated_payoff_months": 22,        // Months to payoff at allocated rate
	//   "estimated_payoff_date": "2025-11-30",
	//   "interest_saved": 2000000,            // Interest saved vs minimum
	//   "months_accelerated": 6,              // Months saved vs minimum
	//   "optimization_score": 0.88,           // How well optimized (0-1)
	//   "last_optimized": "2024-01-15T10:00:00Z",
	//   "last_analyzed": "2024-01-15T10:00:00Z",
	//   "optimization_version": "v1.0"
	// }

	// Strategy Metadata for Debt Payoff Strategy
	StrategyMetadata datatypes.JSON `gorm:"type:jsonb;column:strategy_metadata" json:"strategy_metadata,omitempty"`
	// Structure:
	// {
	//   "payoff_strategy": "avalanche",       // avalanche, snowball, custom
	//   "avalanche_rank": 2,                  // Rank in avalanche (by interest rate)
	//   "snowball_rank": 3,                   // Rank in snowball (by balance)
	//   "avalanche_total_interest": 3000000,  // Total interest if avalanche
	//   "snowball_total_interest": 3500000,   // Total interest if snowball
	//   "avalanche_payoff_months": 18,        // Months to payoff (avalanche)
	//   "snowball_payoff_months": 20,         // Months to payoff (snowball)
	//   "recommended_strategy": "avalanche",  // Recommended strategy
	//   "savings_vs_minimum": 2000000,        // Savings vs minimum payments
	//   "psychological_benefit_score": 0.7,   // Snowball motivation score
	//   "financial_benefit_score": 0.9,       // Avalanche savings score
	//   "last_evaluated": "2024-01-15T10:00:00Z"
	// }

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Debt
func (Debt) TableName() string {
	return "debts"
}

// IsPaidOff checks if the debt has been paid off
func (d *Debt) IsPaidOff() bool {
	return d.CurrentBalance <= 0 || d.Status == DebtStatusPaidOff
}

// IsOverdue checks if the debt payment is overdue
func (d *Debt) IsOverdue() bool {
	if d.NextPaymentDate == nil || d.IsPaidOff() {
		return false
	}
	return time.Now().After(*d.NextPaymentDate)
}

// DaysUntilNextPayment calculates days until next payment
func (d *Debt) DaysUntilNextPayment() int {
	if d.NextPaymentDate == nil {
		return 0
	}
	duration := time.Until(*d.NextPaymentDate)
	return int(duration.Hours() / 24)
}

// UpdateCalculatedFields updates progress, remaining amount, and status
func (d *Debt) UpdateCalculatedFields() {
	d.RemainingAmount = d.CurrentBalance
	if d.RemainingAmount < 0 {
		d.RemainingAmount = 0
	}

	if d.PrincipalAmount > 0 {
		d.PercentagePaid = ((d.PrincipalAmount - d.CurrentBalance) / d.PrincipalAmount) * 100
		if d.PercentagePaid > 100 {
			d.PercentagePaid = 100
		}
		if d.PercentagePaid < 0 {
			d.PercentagePaid = 0
		}
	} else {
		d.PercentagePaid = 0
	}

	// Update status
	if d.IsPaidOff() && d.Status != DebtStatusPaidOff {
		d.Status = DebtStatusPaidOff
		now := time.Now()
		d.PaidOffDate = &now
	} else if d.IsOverdue() && d.Status == DebtStatusActive {
		d.Status = DebtStatusDefaulted
	}
}

// AddPayment adds a payment to the debt
func (d *Debt) AddPayment(amount float64) {
	if amount <= 0 {
		return
	}

	// Calculate interest portion if applicable
	interestPortion := 0.0
	if d.InterestRate > 0 && d.CurrentBalance > 0 {
		// Simple interest calculation for the period
		// This is a simplified calculation - actual interest depends on payment frequency
		monthlyInterestRate := d.InterestRate / 12 / 100
		interestPortion = d.CurrentBalance * monthlyInterestRate
		if interestPortion > amount {
			interestPortion = amount
		}
	}

	principalPortion := amount - interestPortion
	d.CurrentBalance -= principalPortion
	if d.CurrentBalance < 0 {
		d.CurrentBalance = 0
	}

	d.TotalPaid += amount
	d.TotalInterestPaid += interestPortion

	now := time.Now()
	d.LastPaymentDate = &now
	d.LastPaymentAmount = &amount

	d.UpdateCalculatedFields()
}

// CalculateNextPaymentDate calculates the next payment date based on frequency
func (d *Debt) CalculateNextPaymentDate() *time.Time {
	if d.PaymentFrequency == nil {
		return nil
	}

	baseDate := time.Now()
	if d.LastPaymentDate != nil {
		baseDate = *d.LastPaymentDate
	} else if d.NextPaymentDate != nil {
		baseDate = *d.NextPaymentDate
	}

	var nextDate time.Time
	switch *d.PaymentFrequency {
	case FrequencyDaily:
		nextDate = baseDate.AddDate(0, 0, 1)
	case FrequencyWeekly:
		nextDate = baseDate.AddDate(0, 0, 7)
	case FrequencyBiweekly:
		nextDate = baseDate.AddDate(0, 0, 14)
	case FrequencyMonthly:
		nextDate = baseDate.AddDate(0, 1, 0)
	case FrequencyQuarterly:
		nextDate = baseDate.AddDate(0, 3, 0)
	case FrequencyYearly:
		nextDate = baseDate.AddDate(1, 0, 0)
	default:
		return nil
	}

	return &nextDate
}
