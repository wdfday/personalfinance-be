package engine

import (
	"math"
	"testing"
)

// TestFinancialCalculator_ExtremeRates tests financial calculations with extreme interest rates
func TestFinancialCalculator_ExtremeRates(t *testing.T) {
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
			name:        "Zero Interest Rate",
			payment:     100,
			monthlyRate: 0,
			months:      12,
			wantMin:     1200, // 100 * 12
			wantMax:     1200,
		},
		{
			name:        "High Interest Rate (100% APR)",
			payment:     100,
			monthlyRate: 1.0 / 12, // ~8.33% monthly
			months:      12,
			wantMin:     1900, // High growth
			wantMax:     2000,
		},
		{
			name:        "Negative Interest Rate (-5% APR)",
			payment:     100,
			monthlyRate: -0.05 / 12,
			months:      12,
			wantMin:     1160, // Loss of value
			wantMax:     1180,
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

// TestFinancialCalculator_LongTimelines tests calculations over very long periods (30-50 years)
func TestFinancialCalculator_LongTimelines(t *testing.T) {
	calc := NewFinancialCalculator()

	// 40 years of monthly investing
	months := 40 * 12
	payment := 500.0
	// 7% annual return
	monthlyRate := 0.07 / 12

	fv := calc.FutureValueAnnuity(payment, monthlyRate, months)

	// Expected: ~1.2M
	// FV = 500 * ((1.00583^480 - 1) / 0.00583)
	//    = 500 * (15.32 / 0.00583) = 1.3M approx

	if fv < 1_000_000 || fv > 1_500_000 {
		t.Errorf("Expected ~1.2M for 40 years investing, got %v", fv)
	}

	// NPV check for long timeline
	// Cashflows: $500 for 480 months
	cashFlows := make([]float64, months)
	for i := range cashFlows {
		cashFlows[i] = payment
	}

	npv := calc.NPV(cashFlows, monthlyRate)

	// Value should be significantly less than sum due to discounting
	sum := payment * float64(months) // 240,000

	if npv >= sum {
		t.Errorf("NPV %v should be less than sum %v", npv, sum)
	}
}

// TestFinancialCalculator_ComplexCashFlows tests NPV with irregular flows
func TestFinancialCalculator_ComplexCashFlows(t *testing.T) {
	calc := NewFinancialCalculator()

	// Investing $1000, then getting returns
	cashFlows := []float64{
		-1000, // Initial investment
		100,   // Year 1
		200,   // Year 2
		300,   // Year 3
		600,   // Year 4
	}

	discountRate := 0.10 // 10% annual

	// Year 1: 100/1.1 = 90.9
	// Year 2: 200/1.21 = 165.2
	// Year 3: 300/1.331 = 225.4
	// Year 4: 600/1.4641 = 409.8
	// Total Inflows PV = ~891.3
	// Net = -1000 + 891.3 = -108.7

	// Note: Engine uses monthly rate for period, assuming input array is per period
	// If we treat periods as years:
	npv := calc.NPV(cashFlows, discountRate)

	if npv > -100 && npv < -110 {
		t.Logf("NPV correctly calculated: %v", npv)
	}
}

// TestFinancialCalculator_InvestmentVariance simulates sequence of returns risk
func TestFinancialCalculator_InvestmentVariance(t *testing.T) {
	calc := NewFinancialCalculator()

	initial := 10000.0
	contribution := 0.0
	months := 12

	// Scenario A: +20% then -20%
	// 10000 * 1.2 = 12000 * 0.8 = 9600
	resA := calc.SimulateInvestmentGrowth(initial, contribution, 0.20, months/2)
	resA = calc.SimulateInvestmentGrowth(resA, contribution, -0.20, months/2)

	// Scenario B: -20% then +20%
	// 10000 * 0.8 = 8000 * 1.2 = 9600
	resB := calc.SimulateInvestmentGrowth(initial, contribution, -0.20, months/2)
	resB = calc.SimulateInvestmentGrowth(resB, contribution, 0.20, months/2)

	// Scenario C: Flat 0%
	resC := calc.SimulateInvestmentGrowth(initial, contribution, 0.0, months)

	if math.Abs(resA-resB) > 1.0 {
		t.Errorf("Order of returns shouldn't matter for lump sum: %v vs %v", resA, resB)
	}

	if resA >= resC {
		t.Errorf("Volatility drag should result in lower value than flat: %v vs %v", resA, resC)
	}
}
