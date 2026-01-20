package engine

import (
	"math"
	"testing"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
)

func TestFutureValueAnnuity(t *testing.T) {
	calc := NewFinancialCalculator()

	tests := []struct {
		name        string
		payment     float64
		monthlyRate float64
		months      int
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "zero rate",
			payment:     1000,
			monthlyRate: 0,
			months:      12,
			wantMin:     12000,
			wantMax:     12000,
		},
		{
			name:        "7% annual rate for 5 years",
			payment:     500,
			monthlyRate: 0.07 / 12,
			months:      60,
			wantMin:     35000, // Should be around $35,796
			wantMax:     37000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.FutureValueAnnuity(tt.payment, tt.monthlyRate, tt.months)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("FutureValueAnnuity() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSimulateDebtPayoff(t *testing.T) {
	calc := NewFinancialCalculator()

	debts := []domain.DebtInfo{
		{
			ID:             "cc1",
			Name:           "Credit Card",
			Balance:        5000,
			InterestRate:   0.18, // 18% APR
			MinimumPayment: 100,
		},
		{
			ID:             "car",
			Name:           "Car Loan",
			Balance:        10000,
			InterestRate:   0.06, // 6% APR
			MinimumPayment: 200,
		},
	}

	tests := []struct {
		name              string
		extraPayment      float64
		maxMonths         int
		wantMonthsMax     int
		wantInterestSaved bool
	}{
		{
			name:              "no extra payment",
			extraPayment:      0,
			maxMonths:         120,
			wantMonthsMax:     120,
			wantInterestSaved: false,
		},
		{
			name:              "with extra payment",
			extraPayment:      500,
			maxMonths:         120,
			wantMonthsMax:     30, // Should pay off faster
			wantInterestSaved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			months, _, interestSaved := calc.SimulateDebtPayoff(debts, tt.extraPayment, tt.maxMonths)

			if months > tt.wantMonthsMax {
				t.Errorf("SimulateDebtPayoff() months = %v, want <= %v", months, tt.wantMonthsMax)
			}

			if tt.wantInterestSaved && interestSaved <= 0 {
				t.Errorf("SimulateDebtPayoff() interestSaved = %v, want > 0", interestSaved)
			}
		})
	}
}

func TestCalculateGoalMonths(t *testing.T) {
	calc := NewFinancialCalculator()

	tests := []struct {
		name         string
		current      float64
		target       float64
		monthly      float64
		annualReturn float64
		wantMin      int
		wantMax      int
	}{
		{
			name:         "already achieved",
			current:      10000,
			target:       10000,
			monthly:      500,
			annualReturn: 0.07,
			wantMin:      0,
			wantMax:      0,
		},
		{
			name:         "simple savings no return",
			current:      0,
			target:       6000,
			monthly:      500,
			annualReturn: 0,
			wantMin:      12,
			wantMax:      12,
		},
		{
			name:         "with investment return",
			current:      0,
			target:       6000,
			monthly:      500,
			annualReturn: 0.07,
			wantMin:      10, // Should be faster with returns
			wantMax:      12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.CalculateGoalMonths(tt.current, tt.target, tt.monthly, tt.annualReturn)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateGoalMonths() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestNPV(t *testing.T) {
	calc := NewFinancialCalculator()

	// Simple test: $100 per month for 12 months at 5% annual discount rate
	cashFlows := make([]float64, 12)
	for i := range cashFlows {
		cashFlows[i] = 100
	}

	npv := calc.NPV(cashFlows, 0.05/12)

	// NPV should be less than simple sum (1200) due to discounting
	if npv >= 1200 || npv < 1100 {
		t.Errorf("NPV() = %v, want between 1100 and 1200", npv)
	}
}

func TestGenerateNetWorthTimeline(t *testing.T) {
	calc := NewFinancialCalculator()

	debts := []domain.DebtInfo{
		{Balance: 5000, InterestRate: 0.10, MinimumPayment: 100},
	}

	points := calc.GenerateNetWorthTimeline(
		debts,
		1000, // initial savings
		200,  // monthly debt payment
		300,  // monthly savings
		0.07, // 7% return
		24,   // 24 months
		6,    // every 6 months
	)

	if len(points) < 4 {
		t.Errorf("Expected at least 4 points, got %d", len(points))
	}

	// Net worth should increase over time
	if len(points) >= 2 {
		if points[len(points)-1].NetWorth <= points[0].NetWorth {
			t.Errorf("Net worth should increase over time")
		}
	}

	// Debt should decrease over time
	if len(points) >= 2 {
		if points[len(points)-1].DebtTotal >= points[0].DebtTotal {
			t.Errorf("Debt should decrease over time")
		}
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
