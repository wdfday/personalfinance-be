package service

import (
	"fmt"
	"math"
	"time"
)

// AutoScorer automatically calculates AHP criteria scores from Goal attributes
type AutoScorer struct{}

// NewAutoScorer creates a new AutoScorer instance
func NewAutoScorer() *AutoScorer {
	return &AutoScorer{}
}

// CalculateUrgency calculates urgency score based on deadline proximity
// Returns a score between 0.0 (not urgent) and 1.0 (very urgent)
//
// Improved formula using exponential decay for better differentiation:
//   - Uses exponential function to emphasize short-term urgency
//   - Better distinguishes between 1-6 month deadlines
//   - Normalized by 1095 days (3 years) for practical goal distribution
//
// Formula: urgency = exp(-days_remaining / decay_factor)
//
//	where decay_factor = 365 (1 year) for exponential decay
//
// Examples:
//   - 7 days:    exp(-7/365)   = 0.981 (extremely urgent)
//   - 30 days:   exp(-30/365)  = 0.921 (very urgent)
//   - 90 days:   exp(-90/365)  = 0.784 (urgent)
//   - 180 days:  exp(-180/365) = 0.614 (moderately urgent)
//   - 365 days:  exp(-365/365) = 0.368 (moderate)
//   - 730 days:  exp(-730/365) = 0.135 (low urgency)
//   - Overdue:   1.0 (maximum urgency)
func (as *AutoScorer) CalculateUrgency(goal *GoalLike) float64 {
	// No deadline = low urgency
	if goal.TargetDate == nil {
		return 0.1
	}

	now := time.Now()
	daysRemaining := goal.TargetDate.Sub(now).Hours() / 24

	// Overdue = maximum urgency
	if daysRemaining <= 0 {
		return 1.0
	}

	// Use exponential decay for better differentiation, especially in short-term
	// Decay factor of 365 days (1 year) provides good distribution
	decayFactor := 365.0
	urgency := math.Exp(-daysRemaining / decayFactor)

	// Normalize to ensure very long-term goals (3+ years) get low urgency
	// Apply additional scaling for goals beyond 3 years
	if daysRemaining > 1095 { // 3 years
		// Further reduce urgency for very long-term goals
		additionalScale := 1.0 - math.Min(1.0, (daysRemaining-1095)/1095.0)
		urgency *= math.Max(0.05, additionalScale) // Minimum 5% for very long-term
	}

	return urgency
}

// CalculateFeasibility calculates how feasible it is to achieve the goal
// Returns a score between 0.0 (not feasible) and 1.0 (very feasible)
//
// Improved formula considers:
//  1. Ratio of required monthly contribution to income
//  2. Progress bonus: goals near completion get higher feasibility
//  3. Remaining amount relative to income (smaller = more feasible)
//
// Formula:
//
//	base_feasibility = min(1, monthly_income / required_monthly)
//	progress_bonus = (current_amount / target_amount) * 0.2 (max 20% bonus)
//	size_factor = 1 - min(0.3, remaining_amount / (monthly_income * 12)) (max 30% penalty)
//	feasibility = base_feasibility * (1 + progress_bonus - size_factor)
//
// Examples:
//   - Income 50M, need 20M/month, 80% complete: ~1.0 (very feasible)
//   - Income 50M, need 50M/month: ~1.0 (just feasible)
//   - Income 50M, need 80M/month: ~0.625 (challenging)
//   - Income 50M, need 100M/month, large amount: ~0.4 (difficult)
func (as *AutoScorer) CalculateFeasibility(goal *GoalLike, monthlyIncome float64) float64 {
	// Invalid income
	if monthlyIncome <= 0 {
		return 0.0
	}

	// Already completed
	if goal.IsCompleted() {
		return 1.0
	}

	// No deadline = can't calculate precisely
	if goal.TargetDate == nil {
		// Estimate based on remaining amount vs income
		if goal.TargetAmount > 0 {
			progressRatio := goal.CurrentAmount / goal.TargetAmount
			// If already 50%+ complete, assume feasible
			if progressRatio >= 0.5 {
				return 0.7
			}
		}
		return 0.5 // Neutral
	}

	// Calculate required monthly contribution
	daysRemaining := goal.DaysRemaining()
	if daysRemaining <= 0 {
		return 0.0 // Overdue = not feasible
	}

	monthsRemaining := float64(daysRemaining) / 30.0
	if monthsRemaining <= 0 {
		return 0.0
	}

	remainingAmount := goal.RemainingAmount
	if remainingAmount <= 0 {
		return 1.0 // Nothing left to save
	}

	requiredMonthly := remainingAmount / monthsRemaining

	// Base feasibility: ratio of income to required contribution
	baseFeasibility := math.Min(1.0, monthlyIncome/requiredMonthly)

	// Progress bonus: goals near completion are more feasible
	// Bonus up to 20% for goals that are 80%+ complete
	progressRatio := 0.0
	if goal.TargetAmount > 0 {
		progressRatio = goal.CurrentAmount / goal.TargetAmount
	}
	progressBonus := 0.0
	if progressRatio >= 0.8 {
		progressBonus = 0.2 // 20% bonus for 80%+ complete
	} else if progressRatio >= 0.5 {
		progressBonus = 0.1 * (progressRatio - 0.5) / 0.3 // Linear from 0% to 10% for 50-80%
	}

	// Size factor: very large remaining amounts relative to income reduce feasibility
	// Penalty up to 30% for goals requiring >1 year of income
	annualIncome := monthlyIncome * 12
	sizePenalty := 0.0
	if annualIncome > 0 {
		yearsOfIncome := remainingAmount / annualIncome
		if yearsOfIncome > 1.0 {
			// Penalty increases for goals requiring more than 1 year of income
			sizePenalty = math.Min(0.3, 0.3*(yearsOfIncome-1.0)/2.0) // Max 30% penalty at 3+ years
		}
	}

	// Combine factors
	feasibility := baseFeasibility * (1.0 + progressBonus - sizePenalty)

	// Ensure result is in valid range [0, 1]
	feasibility = math.Max(0.0, math.Min(1.0, feasibility))

	return feasibility
}

// CalculateImportance calculates importance based on goal type, progress, and scale
// Returns a score between 0.0 (low importance) and 1.0 (critical importance)
//
// Improved formula considers:
//  1. Base importance from category (financial planning best practices)
//  2. User-set priority adjustment
//  3. Progress factor: goals near completion are more important to finish
//  4. Scale factor: larger goals relative to income are more important
//
// Mapping based on financial planning best practices:
//   - Emergency Fund:  1.00 (foundation of financial security)
//   - Debt Repayment:  0.95 (high interest debt is critical)
//   - Retirement:      0.90 (long-term security)
//   - Education:       0.80 (investment in future)
//   - Purchase (home): 0.70 (major life goal)
//   - Investment:      0.65 (wealth building)
//   - Savings:         0.60 (general savings)
//   - Other:           0.50 (user-defined)
func (as *AutoScorer) CalculateImportance(goal *GoalLike) float64 {
	importanceMap := map[GoalCategory]float64{
		GoalCategoryEmergency:  1.00,
		GoalCategoryDebt:       0.95,
		GoalCategoryRetirement: 0.90,
		GoalCategoryEducation:  0.80,
		GoalCategoryPurchase:   0.70,
		GoalCategoryInvestment: 0.65,
		GoalCategorySavings:    0.60,
		GoalCategoryTravel:     0.55,
		GoalCategoryOther:      0.50,
	}

	importance, exists := importanceMap[goal.Category]
	if !exists {
		importance = 0.50 // Default
	}

	// Adjust by user-set priority if available
	adjustmentMap := map[GoalPriority]float64{
		GoalPriorityCritical: 1.10, // +10%
		GoalPriorityHigh:     1.05, // +5%
		GoalPriorityMedium:   1.00, // No change
		GoalPriorityLow:      0.90, // -10%
	}

	adjustment := adjustmentMap[goal.Priority]
	if adjustment == 0 {
		adjustment = 1.0
	}

	adjusted := importance * adjustment

	// Progress factor: goals near completion (80%+) get importance boost
	// This encourages finishing goals that are almost done
	progressFactor := 0.0
	if goal.TargetAmount > 0 {
		progressRatio := goal.CurrentAmount / goal.TargetAmount
		if progressRatio >= 0.8 {
			// Boost importance by up to 10% for goals 80%+ complete
			progressFactor = 0.1 * math.Min(1.0, (progressRatio-0.8)/0.2)
		}
	}

	// Scale factor: larger goals (relative to typical income) are more important
	// This is a subtle adjustment - goals requiring significant resources matter more
	// Note: We don't have monthlyIncome here, so we use a relative scale based on target amount
	// Goals > 100M get slight boost, goals > 500M get moderate boost
	scaleFactor := 0.0
	if goal.TargetAmount >= 500_000_000 { // 500M+
		scaleFactor = 0.05 // 5% boost for very large goals
	} else if goal.TargetAmount >= 100_000_000 { // 100M+
		scaleFactor = 0.02 // 2% boost for large goals
	}

	// Combine all factors
	finalImportance := adjusted * (1.0 + progressFactor + scaleFactor)

	// Cap at 1.0
	if finalImportance > 1.0 {
		finalImportance = 1.0
	}

	return finalImportance
}

// CalculateImpact calculates financial impact based on target amount
// Returns a score between 0.0 (low impact) and 1.0 (high impact)
//
// Formula: impact = min(1, target_amount / (monthly_income × 12))
//
// This measures how significant the goal is relative to annual income:
//   - Target = 100M, Income = 50M/month (600M/year): min(1, 100/600) = 0.167 (low impact)
//   - Target = 600M, Income = 50M/month:              min(1, 600/600) = 1.0   (high impact)
//   - Target = 1200M, Income = 50M/month:             min(1, 1200/600) = 1.0  (very high impact)
func (as *AutoScorer) CalculateImpact(goal *GoalLike, monthlyIncome float64) float64 {
	// Invalid income
	if monthlyIncome <= 0 {
		return 0.5 // Neutral
	}

	annualIncome := monthlyIncome * 12

	// Calculate impact as ratio of target to annual income
	impact := math.Min(1.0, goal.TargetAmount/annualIncome)

	return impact
}

// CalculateAllCriteria calculates all criteria scores for a goal
// Returns a map of criterion name to score (0.0 - 1.0)
// NOTE: Impact is temporarily disabled
func (as *AutoScorer) CalculateAllCriteria(goal *GoalLike, monthlyIncome float64) map[string]float64 {
	return map[string]float64{
		"urgency":     as.CalculateUrgency(goal),
		"feasibility": as.CalculateFeasibility(goal, monthlyIncome),
		"importance":  as.CalculateImportance(goal),
		// "impact":      as.CalculateImpact(goal, monthlyIncome), // Temporarily disabled
	}
}

// ScoreWithReason holds a score value and its explanation
// This is defined here instead of importing from dto to avoid import cycle
type ScoreWithReason struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// CalculateAllCriteriaWithReasons calculates all criteria scores with explanations
// This is used for the "auto-score with user review" flow
// NOTE: Impact is temporarily disabled
func (as *AutoScorer) CalculateAllCriteriaWithReasons(goal *GoalLike, monthlyIncome float64) map[string]ScoreWithReason {
	return map[string]ScoreWithReason{
		"urgency":     as.CalculateUrgencyWithReason(goal),
		"feasibility": as.CalculateFeasibilityWithReason(goal, monthlyIncome),
		"importance":  as.CalculateImportanceWithReason(goal),
		// "impact":      as.CalculateImpactWithReason(goal, monthlyIncome), // Temporarily disabled
	}
}

// CalculateUrgencyWithReason calculates urgency score with explanation
func (as *AutoScorer) CalculateUrgencyWithReason(goal *GoalLike) ScoreWithReason {
	score := as.CalculateUrgency(goal)
	var reason string

	if goal.TargetDate == nil {
		reason = "Không có deadline - độ khẩn cấp thấp"
	} else {
		days := goal.DaysRemaining()
		if days <= 0 {
			reason = "Đã quá hạn - cần ưu tiên ngay!"
		} else if days <= 30 {
			reason = "Còn " + formatDays(days) + " - rất khẩn cấp"
		} else if days <= 90 {
			reason = "Còn " + formatDays(days) + " - khẩn cấp"
		} else if days <= 180 {
			reason = "Còn " + formatDays(days) + " - cần chú ý"
		} else {
			reason = "Còn " + formatDays(days) + " - có thời gian chuẩn bị"
		}
	}

	return ScoreWithReason{Score: score, Reason: reason}
}

// CalculateFeasibilityWithReason calculates feasibility score with explanation
func (as *AutoScorer) CalculateFeasibilityWithReason(goal *GoalLike, monthlyIncome float64) ScoreWithReason {
	score := as.CalculateFeasibility(goal, monthlyIncome)
	var reason string

	if monthlyIncome <= 0 {
		reason = "Không có thông tin thu nhập"
	} else if goal.IsCompleted() {
		reason = "Mục tiêu đã hoàn thành"
	} else if goal.TargetDate == nil {
		// Estimate based on progress
		if goal.TargetAmount > 0 {
			progressRatio := goal.CurrentAmount / goal.TargetAmount
			if progressRatio >= 0.5 {
				reason = fmt.Sprintf("Đã hoàn thành %.0f%% - khả thi", progressRatio*100)
			} else {
				reason = "Không có deadline - không thể đánh giá chính xác"
			}
		} else {
			reason = "Không có deadline - không thể đánh giá"
		}
	} else {
		days := goal.DaysRemaining()
		if days <= 0 {
			reason = "Đã quá hạn - không khả thi"
		} else {
			months := float64(days) / 30.0
			requiredMonthly := goal.RemainingAmount / months
			percentOfIncome := (requiredMonthly / monthlyIncome) * 100

			// Add progress info if significant
			progressInfo := ""
			if goal.TargetAmount > 0 {
				progressRatio := goal.CurrentAmount / goal.TargetAmount
				if progressRatio >= 0.8 {
					progressInfo = fmt.Sprintf(", đã hoàn thành %.0f%%", progressRatio*100)
				} else if progressRatio >= 0.5 {
					progressInfo = fmt.Sprintf(", đã hoàn thành %.0f%%", progressRatio*100)
				}
			}

			if percentOfIncome <= 20 {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập)" + progressInfo + " - rất khả thi"
			} else if percentOfIncome <= 50 {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập)" + progressInfo + " - khả thi"
			} else if percentOfIncome <= 80 {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập)" + progressInfo + " - thử thách"
			} else {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập)" + progressInfo + " - khó đạt được"
			}
		}
	}

	return ScoreWithReason{Score: score, Reason: reason}
}

// CalculateImportanceWithReason calculates importance score with explanation
func (as *AutoScorer) CalculateImportanceWithReason(goal *GoalLike) ScoreWithReason {
	score := as.CalculateImportance(goal)
	var reason string

	categoryReasons := map[GoalCategory]string{
		GoalCategoryEmergency:  "Quỹ khẩn cấp - nền tảng an toàn tài chính",
		GoalCategoryDebt:       "Trả nợ - giảm gánh nặng lãi suất",
		GoalCategoryRetirement: "Hưu trí - đảm bảo tương lai dài hạn",
		GoalCategoryEducation:  "Giáo dục - đầu tư vào tương lai",
		GoalCategoryPurchase:   "Mua sắm lớn - mục tiêu cuộc sống",
		GoalCategoryInvestment: "Đầu tư - xây dựng tài sản",
		GoalCategorySavings:    "Tiết kiệm - dự phòng tài chính",
		GoalCategoryTravel:     "Du lịch - trải nghiệm cuộc sống",
		GoalCategoryOther:      "Mục tiêu cá nhân",
	}

	reason = categoryReasons[goal.Category]
	if reason == "" {
		reason = "Mục tiêu cá nhân"
	}

	// Add priority modifier info
	switch goal.Priority {
	case GoalPriorityCritical:
		reason += " (ưu tiên: quan trọng nhất)"
	case GoalPriorityHigh:
		reason += " (ưu tiên: cao)"
	case GoalPriorityMedium:
		reason += " (ưu tiên: trung bình)"
	case GoalPriorityLow:
		reason += " (ưu tiên: thấp)"
	}

	// Add progress info if significant
	if goal.TargetAmount > 0 {
		progressRatio := goal.CurrentAmount / goal.TargetAmount
		if progressRatio >= 0.8 {
			reason += fmt.Sprintf(" (đã hoàn thành %.0f%% - gần xong)", progressRatio*100)
		} else if progressRatio >= 0.5 {
			reason += fmt.Sprintf(" (đã hoàn thành %.0f%%)", progressRatio*100)
		}
	}

	// Add scale info for large goals
	if goal.TargetAmount >= 500_000_000 {
		reason += " (mục tiêu quy mô lớn)"
	}

	return ScoreWithReason{Score: score, Reason: reason}
}

// CalculateImpactWithReason calculates impact score with explanation
func (as *AutoScorer) CalculateImpactWithReason(goal *GoalLike, monthlyIncome float64) ScoreWithReason {
	score := as.CalculateImpact(goal, monthlyIncome)
	var reason string

	if monthlyIncome <= 0 {
		reason = "Không có thông tin thu nhập"
	} else {
		annualIncome := monthlyIncome * 12
		monthsOfIncome := goal.TargetAmount / monthlyIncome

		if monthsOfIncome <= 1 {
			reason = formatMoney(goal.TargetAmount) + " = " + formatFloat(monthsOfIncome) + " tháng thu nhập - ảnh hưởng nhỏ"
		} else if monthsOfIncome <= 6 {
			reason = formatMoney(goal.TargetAmount) + " = " + formatFloat(monthsOfIncome) + " tháng thu nhập - ảnh hưởng vừa"
		} else if goal.TargetAmount <= annualIncome {
			reason = formatMoney(goal.TargetAmount) + " = " + formatFloat(monthsOfIncome) + " tháng thu nhập - ảnh hưởng lớn"
		} else {
			years := goal.TargetAmount / annualIncome
			reason = formatMoney(goal.TargetAmount) + " = " + formatFloat(years) + " năm thu nhập - ảnh hưởng rất lớn"
		}
	}

	return ScoreWithReason{Score: score, Reason: reason}
}

// Helper functions for formatting
func formatDays(days int) string {
	if days == 1 {
		return "1 ngày"
	}
	return formatInt(days) + " ngày"
}

func formatMoney(amount float64) string {
	if amount >= 1_000_000_000 {
		return formatFloat(amount/1_000_000_000) + " tỷ"
	} else if amount >= 1_000_000 {
		return formatFloat(amount/1_000_000) + " triệu"
	}
	return formatFloat(amount/1_000) + "K"
}

func formatPercent(percent float64) string {
	return formatFloat(percent) + "%"
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return formatInt(int(f))
	}
	return fmt.Sprintf("%.1f", f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
