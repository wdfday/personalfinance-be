package domain

// DebtType represents the type of debt
type DebtType string

const (
	DebtTypeCreditCard   DebtType = "credit_card"   // Credit card debt
	DebtTypePersonalLoan DebtType = "personal_loan" // Personal loan
	DebtTypeMortgage     DebtType = "mortgage"      // Mortgage
	DebtTypeAutoLoan     DebtType = "auto_loan"     // Auto loan
	DebtTypeStudentLoan  DebtType = "student_loan"  // Student loan
	DebtTypeMedical      DebtType = "medical"       // Medical debt
	DebtTypeOther        DebtType = "other"         // Other debt
)

// IsValid checks if the debt type is valid
func (dt DebtType) IsValid() bool {
	switch dt {
	case DebtTypeCreditCard, DebtTypePersonalLoan, DebtTypeMortgage,
		DebtTypeAutoLoan, DebtTypeStudentLoan, DebtTypeMedical, DebtTypeOther:
		return true
	}
	return false
}

// DebtStatus represents the current status of a debt
type DebtStatus string

const (
	DebtStatusActive    DebtStatus = "active"    // Debt is active
	DebtStatusPaidOff   DebtStatus = "paid_off"  // Debt has been paid off
	DebtStatusSettled   DebtStatus = "settled"   // Debt was settled
	DebtStatusDefaulted DebtStatus = "defaulted" // Debt is in default
	DebtStatusInactive  DebtStatus = "inactive"  // Debt is inactive
)

// IsValid checks if the debt status is valid
func (ds DebtStatus) IsValid() bool {
	switch ds {
	case DebtStatusActive, DebtStatusPaidOff, DebtStatusSettled,
		DebtStatusDefaulted, DebtStatusInactive:
		return true
	}
	return false
}

// PaymentFrequency represents how often payments are made
type PaymentFrequency string

const (
	FrequencyOneTime   PaymentFrequency = "one_time"
	FrequencyDaily     PaymentFrequency = "daily"
	FrequencyWeekly    PaymentFrequency = "weekly"
	FrequencyBiweekly  PaymentFrequency = "biweekly"
	FrequencyMonthly   PaymentFrequency = "monthly"
	FrequencyQuarterly PaymentFrequency = "quarterly"
	FrequencyYearly    PaymentFrequency = "yearly"
)

// IsValid checks if the payment frequency is valid
func (pf PaymentFrequency) IsValid() bool {
	switch pf {
	case FrequencyOneTime, FrequencyDaily, FrequencyWeekly, FrequencyBiweekly,
		FrequencyMonthly, FrequencyQuarterly, FrequencyYearly:
		return true
	}
	return false
}
