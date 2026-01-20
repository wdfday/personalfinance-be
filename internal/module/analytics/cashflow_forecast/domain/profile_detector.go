package domain

import (
	"math"
)

// ===== INCOME PROFILE DETECTION =====
// Auto-detect whether user is Freelancer, Salaried, or Mixed based on income patterns

// IncomeProfileAnalysis contains the analysis result for income pattern detection
type IncomeProfileAnalysis struct {
	// Detected mode
	DetectedMode IncomeMode `json:"detected_mode"`
	Confidence   float64    `json:"confidence"` // 0-100, how confident we are

	// Pattern details
	Patterns IncomePatterns `json:"patterns"`

	// Recommendation
	Recommendation string `json:"recommendation"`
}

// IncomePatterns contains detailed income pattern analysis
type IncomePatterns struct {
	// Regularity: how consistent is the timing?
	// 1.0 = income arrives same day each month
	// 0.0 = completely random timing
	RegularityScore float64 `json:"regularity_score"`

	// Amount stability: how consistent is the amount?
	// Uses coefficient of variation (CV)
	AmountCV float64 `json:"amount_cv"`

	// Source diversity: how many distinct income sources?
	SourceCount    int     `json:"source_count"`
	DominantSource string  `json:"dominant_source"` // "salary", "freelance", "mixed"
	DominanceRatio float64 `json:"dominance_ratio"` // % of income from dominant source

	// Recurring vs variable split
	RecurringRatio float64 `json:"recurring_ratio"` // % of income that's recurring
	VariableRatio  float64 `json:"variable_ratio"`  // % that's variable

	// Zero-income months (feast/famine indicator)
	ZeroIncomeMonths int     `json:"zero_income_months"`
	ZeroIncomeRatio  float64 `json:"zero_income_ratio"` // % of months with 0 income

	// Gap analysis (for freelancer detection)
	LongestGapMonths int     `json:"longest_gap_months"` // Longest stretch of low income
	AverageGapMonths float64 `json:"average_gap_months"`

	// Feast/Famine ratio
	// High ratio = some months much higher than average (typical freelancer)
	FeastFamineRatio float64 `json:"feast_famine_ratio"` // Max / Median
}

// DetectIncomeProfile analyzes historical income to detect user's income profile
func DetectIncomeProfile(historicalIncome []float64, incomeSources []IncomeSource) IncomeProfileAnalysis {
	analysis := IncomeProfileAnalysis{
		DetectedMode: IncomeModeFixed, // Default
		Confidence:   50,              // Neutral
	}

	if len(historicalIncome) < 3 {
		analysis.Recommendation = "Cần ít nhất 3 tháng dữ liệu để phân tích chính xác."
		return analysis
	}

	// Analyze patterns
	patterns := analyzePatterns(historicalIncome, incomeSources)
	analysis.Patterns = patterns

	// Decision tree for mode detection
	mode, confidence, recommendation := classifyIncomeMode(patterns)
	analysis.DetectedMode = mode
	analysis.Confidence = confidence
	analysis.Recommendation = recommendation

	return analysis
}

// analyzePatterns computes all pattern metrics
func analyzePatterns(income []float64, sources []IncomeSource) IncomePatterns {
	patterns := IncomePatterns{}

	if len(income) == 0 {
		return patterns
	}

	// Basic stats
	mean := calculateAverage(income)
	stdDev := calculateStdDev(income, mean)

	if mean > 0 {
		patterns.AmountCV = stdDev / mean
	}

	// Zero-income analysis
	zeroMonths := 0
	for _, val := range income {
		if val < mean*0.1 { // Less than 10% of average = essentially zero
			zeroMonths++
		}
	}
	patterns.ZeroIncomeMonths = zeroMonths
	patterns.ZeroIncomeRatio = float64(zeroMonths) / float64(len(income))

	// Feast/Famine ratio
	sorted := make([]float64, len(income))
	copy(sorted, income)
	bubbleSortFloat(sorted)

	median := sorted[len(sorted)/2]
	maxIncome := sorted[len(sorted)-1]

	if median > 0 {
		patterns.FeastFamineRatio = maxIncome / median
	}

	// Longest gap (consecutive low-income months)
	patterns.LongestGapMonths, patterns.AverageGapMonths = calculateGaps(income, mean*0.3)

	// Source analysis
	if len(sources) > 0 {
		patterns.SourceCount = len(sources)
		patterns.RecurringRatio, patterns.VariableRatio, patterns.DominantSource, patterns.DominanceRatio =
			analyzeSourcesBreakdown(sources)
	}

	// Regularity score (based on CV - lower CV = more regular)
	// CV < 0.1 = 100% regular, CV > 0.5 = 0% regular
	patterns.RegularityScore = math.Max(0, math.Min(100, (0.5-patterns.AmountCV)*200))

	return patterns
}

// classifyIncomeMode determines the income mode based on patterns
func classifyIncomeMode(p IncomePatterns) (IncomeMode, float64, string) {
	// Decision rules based on patterns

	// FREELANCER indicators:
	freelancerScore := 0.0
	reasons := []string{}

	if p.AmountCV > 0.3 {
		freelancerScore += 25
		reasons = append(reasons, "Thu nhập biến động cao")
	}
	if p.AmountCV > 0.5 {
		freelancerScore += 15
	}

	if p.ZeroIncomeRatio > 0.1 {
		freelancerScore += 20
		reasons = append(reasons, "Có tháng không thu nhập")
	}

	if p.FeastFamineRatio > 3 {
		freelancerScore += 20
		reasons = append(reasons, "Mô hình 'no dồn đói góp'")
	}

	if p.LongestGapMonths >= 2 {
		freelancerScore += 15
		reasons = append(reasons, "Có khoảng gián đoạn thu nhập")
	}

	if p.DominantSource == "freelance" && p.DominanceRatio > 0.6 {
		freelancerScore += 20
	}

	// SALARIED indicators:
	salariedScore := 0.0

	if p.AmountCV < 0.1 {
		salariedScore += 30
	}

	if p.RegularityScore > 80 {
		salariedScore += 25
	}

	if p.ZeroIncomeRatio == 0 {
		salariedScore += 15
	}

	if p.DominantSource == "salary" && p.DominanceRatio > 0.7 {
		salariedScore += 30
	}

	if p.RecurringRatio > 0.8 {
		salariedScore += 20
	}

	// MIXED indicators:
	mixedScore := 0.0
	if p.RecurringRatio > 0.3 && p.VariableRatio > 0.3 {
		mixedScore = 50 // Significant both
	}

	// Decision
	maxScore := math.Max(freelancerScore, math.Max(salariedScore, mixedScore))
	confidence := math.Min(95, maxScore+20) // Cap at 95%

	var mode IncomeMode
	var recommendation string

	switch {
	case freelancerScore >= salariedScore && freelancerScore >= mixedScore:
		mode = IncomeModeFreelancer
		if len(reasons) > 0 {
			recommendation = "Phát hiện mô hình Freelancer: " + reasons[0] + ". Nên sử dụng Income Smoothing."
		} else {
			recommendation = "Thu nhập có đặc điểm Freelancer. Nên sử dụng Reservoir Model để làm phẳng."
		}
	case salariedScore >= freelancerScore && salariedScore >= mixedScore:
		mode = IncomeModeFixed
		recommendation = "Thu nhập ổn định như nhân viên lương. Có thể lập ngân sách trực tiếp theo thu nhập tháng."
	default:
		mode = IncomeModeMixed
		recommendation = "Thu nhập hỗn hợp: có phần cố định và phần biến động. Nên smooth phần freelance, giữ nguyên lương."
	}

	return mode, confidence, recommendation
}

// Helper functions

func calculateGaps(income []float64, threshold float64) (longestGap int, avgGap float64) {
	currentGap := 0
	gaps := []int{}

	for _, val := range income {
		if val < threshold {
			currentGap++
		} else {
			if currentGap > 0 {
				gaps = append(gaps, currentGap)
				if currentGap > longestGap {
					longestGap = currentGap
				}
				currentGap = 0
			}
		}
	}
	// Handle trailing gap
	if currentGap > 0 {
		gaps = append(gaps, currentGap)
		if currentGap > longestGap {
			longestGap = currentGap
		}
	}

	if len(gaps) > 0 {
		sum := 0
		for _, g := range gaps {
			sum += g
		}
		avgGap = float64(sum) / float64(len(gaps))
	}

	return longestGap, avgGap
}

func analyzeSourcesBreakdown(sources []IncomeSource) (recurringRatio, variableRatio float64, dominantSource string, dominanceRatio float64) {
	totalAmount := 0.0
	recurringAmount := 0.0
	variableAmount := 0.0

	sourceTypeAmounts := map[string]float64{}

	for _, src := range sources {
		totalAmount += src.Amount

		if src.IsRecurring || src.SourceType == "salary" {
			recurringAmount += src.Amount
		} else {
			variableAmount += src.Amount
		}

		// Aggregate by type
		srcType := src.SourceType
		if srcType == "" {
			if src.IsRecurring {
				srcType = "salary"
			} else {
				srcType = "freelance"
			}
		}
		sourceTypeAmounts[srcType] += src.Amount
	}

	if totalAmount > 0 {
		recurringRatio = recurringAmount / totalAmount
		variableRatio = variableAmount / totalAmount
	}

	// Find dominant source
	maxAmount := 0.0
	for srcType, amount := range sourceTypeAmounts {
		if amount > maxAmount {
			maxAmount = amount
			dominantSource = srcType
		}
	}
	if totalAmount > 0 {
		dominanceRatio = maxAmount / totalAmount
	}

	// Simplify dominant source classification
	if dominantSource != "salary" && dominantSource != "freelance" {
		if recurringRatio > variableRatio {
			dominantSource = "salary"
		} else {
			dominantSource = "freelance"
		}
	}

	return recurringRatio, variableRatio, dominantSource, dominanceRatio
}

func bubbleSortFloat(arr []float64) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

// AutoDetectAndForecast automatically detects mode and runs forecast
func AutoDetectAndForecast(input ForecastInput) (ForecastResult, IncomeProfileAnalysis) {
	analysis := DetectIncomeProfile(input.HistoricalIncome, input.IncomeSources)

	// Override mode with detected mode
	input.Mode = analysis.DetectedMode

	result := CalculateForecast(input)
	return result, analysis
}
