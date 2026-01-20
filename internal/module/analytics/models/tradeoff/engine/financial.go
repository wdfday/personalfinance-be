package engine

import (
	"math"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
)

// FinancialCalculator provides financial calculation utilities
type FinancialCalculator struct{}

// NewFinancialCalculator creates a new calculator
func NewFinancialCalculator() *FinancialCalculator {
	return &FinancialCalculator{}
}

// FutureValueAnnuity calculates future value of regular payments
// FV = PMT × ((1 + r)^n - 1) / r
func (c *FinancialCalculator) FutureValueAnnuity(payment, monthlyRate float64, months int) float64 {
	if monthlyRate == 0 {
		return payment * float64(months)
	}
	return payment * ((math.Pow(1+monthlyRate, float64(months)) - 1) / monthlyRate)
}

// PresentValue calculates present value of a future amount
// PV = FV / (1 + r)^n
func (c *FinancialCalculator) PresentValue(futureValue, monthlyRate float64, months int) float64 {
	if monthlyRate == 0 {
		return futureValue
	}
	return futureValue / math.Pow(1+monthlyRate, float64(months))
}

// NPV calculates Net Present Value of cash flows
// NPV = Σ(CF_t / (1 + r)^t)
func (c *FinancialCalculator) NPV(cashFlows []float64, monthlyDiscountRate float64) float64 {
	npv := 0.0
	for t, cf := range cashFlows {
		npv += cf / math.Pow(1+monthlyDiscountRate, float64(t+1))
	}
	return npv
}

// SimulateDebtPayoff simulates debt payoff and returns months to debt-free and total interest
func (c *FinancialCalculator) SimulateDebtPayoff(
	debts []domain.DebtInfo,
	extraPayment float64,
	maxMonths int,
) (monthsToDebtFree int, totalInterest float64, interestSaved float64) {
	// Copy debts to avoid modifying original
	currentDebts := make([]debtState, len(debts))
	for i, d := range debts {
		currentDebts[i] = debtState{
			balance:        d.Balance,
			interestRate:   d.InterestRate,
			minimumPayment: d.MinimumPayment,
		}
	}

	// Calculate baseline interest (minimum payments only)
	baselineInterest := c.calculateBaselineInterest(debts, maxMonths)

	totalInterest = 0.0
	month := 0

	for month < maxMonths {
		// Check if all debts paid
		allPaid := true
		for _, d := range currentDebts {
			if d.balance > 0.01 {
				allPaid = false
				break
			}
		}
		if allPaid {
			break
		}

		month++
		remainingExtra := extraPayment

		// Apply interest and minimum payments
		for i := range currentDebts {
			if currentDebts[i].balance <= 0 {
				continue
			}

			// Apply monthly interest
			monthlyRate := currentDebts[i].interestRate / 12
			interest := currentDebts[i].balance * monthlyRate
			currentDebts[i].balance += interest
			totalInterest += interest

			// Apply minimum payment
			payment := math.Min(currentDebts[i].minimumPayment, currentDebts[i].balance)
			currentDebts[i].balance -= payment
		}

		// Apply extra payment to highest interest debt (Avalanche method)
		for remainingExtra > 0 {
			highestIdx := -1
			highestRate := 0.0

			for i, d := range currentDebts {
				if d.balance > 0.01 && d.interestRate > highestRate {
					highestRate = d.interestRate
					highestIdx = i
				}
			}

			if highestIdx == -1 {
				break
			}

			payment := math.Min(remainingExtra, currentDebts[highestIdx].balance)
			currentDebts[highestIdx].balance -= payment
			remainingExtra -= payment
		}
	}

	monthsToDebtFree = month
	interestSaved = baselineInterest - totalInterest
	if interestSaved < 0 {
		interestSaved = 0
	}

	return monthsToDebtFree, totalInterest, interestSaved
}

type debtState struct {
	balance        float64
	interestRate   float64
	minimumPayment float64
}

// calculateBaselineInterest calculates interest with minimum payments only
func (c *FinancialCalculator) calculateBaselineInterest(debts []domain.DebtInfo, maxMonths int) float64 {
	currentDebts := make([]debtState, len(debts))
	for i, d := range debts {
		currentDebts[i] = debtState{
			balance:        d.Balance,
			interestRate:   d.InterestRate,
			minimumPayment: d.MinimumPayment,
		}
	}

	totalInterest := 0.0

	for month := 0; month < maxMonths; month++ {
		allPaid := true
		for _, d := range currentDebts {
			if d.balance > 0.01 {
				allPaid = false
				break
			}
		}
		if allPaid {
			break
		}

		for i := range currentDebts {
			if currentDebts[i].balance <= 0 {
				continue
			}

			monthlyRate := currentDebts[i].interestRate / 12
			interest := currentDebts[i].balance * monthlyRate
			currentDebts[i].balance += interest
			totalInterest += interest

			payment := math.Min(currentDebts[i].minimumPayment, currentDebts[i].balance)
			currentDebts[i].balance -= payment
		}
	}

	return totalInterest
}

// SimulateInvestmentGrowth simulates investment growth over time
func (c *FinancialCalculator) SimulateInvestmentGrowth(
	initialAmount float64,
	monthlyContribution float64,
	annualReturn float64,
	months int,
) float64 {
	monthlyRate := annualReturn / 12
	balance := initialAmount

	for m := 0; m < months; m++ {
		balance += monthlyContribution
		balance *= (1 + monthlyRate)
	}

	return balance
}

// CalculateGoalMonths calculates months needed to reach a goal
func (c *FinancialCalculator) CalculateGoalMonths(
	currentAmount float64,
	targetAmount float64,
	monthlyContribution float64,
	annualReturn float64,
) int {
	if monthlyContribution <= 0 {
		return 9999
	}

	gap := targetAmount - currentAmount
	if gap <= 0 {
		return 0
	}

	monthlyRate := annualReturn / 12
	if monthlyRate == 0 {
		return int(math.Ceil(gap / monthlyContribution))
	}

	// Using formula: n = ln((FV*r/PMT) + 1) / ln(1+r)
	// Simplified iterative approach for accuracy
	balance := currentAmount
	months := 0
	maxMonths := 600 // 50 years cap

	for balance < targetAmount && months < maxMonths {
		balance += monthlyContribution
		balance *= (1 + monthlyRate)
		months++
	}

	return months
}

// GenerateNetWorthTimeline generates net worth projection over time
func (c *FinancialCalculator) GenerateNetWorthTimeline(
	debts []domain.DebtInfo,
	initialSavings float64,
	monthlyDebtPayment float64,
	monthlySavings float64,
	annualReturn float64,
	months int,
	intervalMonths int,
) []domain.NetWorthPoint {
	points := make([]domain.NetWorthPoint, 0)

	// Copy debts
	currentDebts := make([]debtState, len(debts))
	for i, d := range debts {
		currentDebts[i] = debtState{
			balance:        d.Balance,
			interestRate:   d.InterestRate,
			minimumPayment: d.MinimumPayment,
		}
	}

	savings := initialSavings
	monthlyRate := annualReturn / 12

	for m := 0; m <= months; m++ {
		if m%intervalMonths == 0 {
			totalDebt := 0.0
			for _, d := range currentDebts {
				if d.balance > 0 {
					totalDebt += d.balance
				}
			}

			points = append(points, domain.NetWorthPoint{
				Month:     m,
				NetWorth:  savings - totalDebt,
				DebtTotal: totalDebt,
				Assets:    savings,
				Savings:   savings,
			})
		}

		if m == months {
			break
		}

		// Update debts
		remainingExtra := monthlyDebtPayment
		for i := range currentDebts {
			if currentDebts[i].balance <= 0 {
				continue
			}

			// Interest
			interest := currentDebts[i].balance * (currentDebts[i].interestRate / 12)
			currentDebts[i].balance += interest

			// Minimum payment
			payment := math.Min(currentDebts[i].minimumPayment, currentDebts[i].balance)
			currentDebts[i].balance -= payment
		}

		// Extra payment to highest interest
		for remainingExtra > 0 {
			highestIdx := -1
			highestRate := 0.0
			for i, d := range currentDebts {
				if d.balance > 0.01 && d.interestRate > highestRate {
					highestRate = d.interestRate
					highestIdx = i
				}
			}
			if highestIdx == -1 {
				break
			}
			payment := math.Min(remainingExtra, currentDebts[highestIdx].balance)
			currentDebts[highestIdx].balance -= payment
			remainingExtra -= payment
		}

		// Update savings
		savings += monthlySavings
		savings *= (1 + monthlyRate)
	}

	return points
}
