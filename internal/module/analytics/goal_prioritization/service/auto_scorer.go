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
// Formula: urgency = 1 - min(1, days_remaining / 1825)
//
// # Normalized by 5 years (1825 days) to better distribute scores for long-term goals
//
// Examples:
//   - 1 month (30 days):   1 - 30/1825   = 0.984 (very urgent)
//   - 6 months (180 days): 1 - 180/1825  = 0.901 (very urgent)
//   - 1 year (365 days):   1 - 365/1825  = 0.800 (urgent)
//   - 2 years (730 days):  1 - 730/1825  = 0.600 (moderately urgent)
//   - 3 years (1095 days): 1 - 1095/1825 = 0.400 (moderate)
//   - 5 years (1825 days): 1 - 1825/1825 = 0.0   (not urgent)
//   - Overdue:             1.0            (maximum urgency)
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

	// Calculate urgency: more days = less urgent
	// Normalize by 1825 days (5 years) for better long-term goal distribution
	urgency := 1.0 - math.Min(1.0, daysRemaining/1825.0)

	return urgency
}

// CalculateFeasibility calculates how feasible it is to achieve the goal
// Returns a score between 0.0 (not feasible) and 1.0 (very feasible)
//
// Formula: feasibility = min(1, monthly_income / required_monthly_contribution)
//
// Examples:
//   - Income 50M, need 20M/month: min(1, 50/20) = 1.0   (very feasible)
//   - Income 50M, need 50M/month: min(1, 50/50) = 1.0   (just feasible)
//   - Income 50M, need 80M/month: min(1, 50/80) = 0.625 (challenging)
//   - No deadline:                0.5 (neutral)
func (as *AutoScorer) CalculateFeasibility(goal *GoalLike, monthlyIncome float64) float64 {
	// Invalid income
	if monthlyIncome <= 0 {
		return 0.0
	}

	// Already completed
	if goal.IsCompleted() {
		return 1.0
	}

	// No deadline = can't calculate
	if goal.TargetDate == nil {
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

	// Calculate feasibility ratio
	feasibility := math.Min(1.0, monthlyIncome/requiredMonthly)

	return feasibility
}

// CalculateImportance calculates importance based on goal type
// Returns a score between 0.0 (low importance) and 1.0 (critical importance)
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
		return 0.50 // Default
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

	// Cap at 1.0
	if adjusted > 1.0 {
		adjusted = 1.0
	}

	return adjusted
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
func (as *AutoScorer) CalculateAllCriteria(goal *GoalLike, monthlyIncome float64) map[string]float64 {
	return map[string]float64{
		"urgency":     as.CalculateUrgency(goal),
		"feasibility": as.CalculateFeasibility(goal, monthlyIncome),
		"importance":  as.CalculateImportance(goal),
		"impact":      as.CalculateImpact(goal, monthlyIncome),
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
func (as *AutoScorer) CalculateAllCriteriaWithReasons(goal *GoalLike, monthlyIncome float64) map[string]ScoreWithReason {
	return map[string]ScoreWithReason{
		"urgency":     as.CalculateUrgencyWithReason(goal),
		"feasibility": as.CalculateFeasibilityWithReason(goal, monthlyIncome),
		"importance":  as.CalculateImportanceWithReason(goal),
		"impact":      as.CalculateImpactWithReason(goal, monthlyIncome),
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
		reason = "Không có deadline - không thể đánh giá"
	} else {
		days := goal.DaysRemaining()
		if days <= 0 {
			reason = "Đã quá hạn - không khả thi"
		} else {
			months := float64(days) / 30.0
			requiredMonthly := goal.RemainingAmount / months
			percentOfIncome := (requiredMonthly / monthlyIncome) * 100

			if percentOfIncome <= 20 {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập) - rất khả thi"
			} else if percentOfIncome <= 50 {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập) - khả thi"
			} else if percentOfIncome <= 80 {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập) - thử thách"
			} else {
				reason = formatMoney(requiredMonthly) + "/tháng (" + formatPercent(percentOfIncome) + " thu nhập) - khó đạt được"
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
