package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"personalfinancedss/internal/module/calendar/month/domain"
	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"

	// Analytics services
	goalDomain "personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	goalDto "personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	// Debt strategy
	debtStrategyDomain "personalfinancedss/internal/module/analytics/debt_strategy/domain"
	debtStrategyService "personalfinancedss/internal/module/analytics/debt_strategy/dto"

	// Budget allocation
	budgetAllocationService "personalfinancedss/internal/module/analytics/budget_allocation/dto"

	// Budget domain
	budgetDomain "personalfinancedss/internal/module/cashflow/budget/domain"
)

// getDebtBehavior returns the debt behavior with default fallback
// Default: "installment" (matches domain.Debt default value)
func getDebtBehavior(behavior string) string {
	if behavior != "" {
		return behavior
	}
	return "installment"
}

// buildInputSnapshotFromCachedState builds domain.InputSnapshot from Redis cached state
// so we persist the actual DSS input when saving MonthState.
func buildInputSnapshotFromCachedState(cached *DSSCachedState, capturedAt time.Time) *domain.InputSnapshot {
	input := domain.NewInputSnapshot()
	input.CapturedAt = capturedAt

	if cached.MonthlyIncome > 0 {
		input.IncomeProfiles = []domain.IncomeSnapshot{{
			Name:        "Estimated",
			Amount:      cached.MonthlyIncome,
			IsRecurring: true,
			SourceType:  "estimate",
		}}
	}

	for _, c := range cached.InputConstraints {
		cid, _ := uuid.Parse(c.ID)
		input.Constraints = append(input.Constraints, domain.ConstraintSnapshot{
			ConstraintID: &cid,
			Name:         c.Name,
			CategoryID:   c.CategoryID,
			Type:         "minimum",
			Amount:       c.MinimumAmount,
			IsRecurring:  true,
			SourceType:   "constraint",
		})
	}

	for _, g := range cached.InputGoals {
		gid, _ := uuid.Parse(g.ID)
		monthlyTarget := g.TargetAmount / 12
		if g.TargetAmount > 0 && g.CurrentAmount < g.TargetAmount {
			remaining := g.TargetAmount - g.CurrentAmount
			monthlyTarget = remaining / 12
		}
		progress := 0.0
		if g.TargetAmount > 0 {
			progress = g.CurrentAmount / g.TargetAmount * 100
		}
		deadline := g.TargetDate
		input.Goals = append(input.Goals, domain.GoalSnapshot{
			GoalID:        gid,
			Name:          g.Name,
			TargetAmount:  g.TargetAmount,
			CurrentAmount: g.CurrentAmount,
			Progress:      progress,
			MonthlyTarget: monthlyTarget,
			IsSelected:    true,
			Deadline:      &deadline,
		})
	}

	for _, d := range cached.InputDebts {
		did, _ := uuid.Parse(d.ID)
		input.Debts = append(input.Debts, domain.DebtSnapshot{
			DebtID:         did,
			Name:           d.Name,
			CurrentBalance: d.CurrentBalance,
			InterestRate:   d.InterestRate,
			MinPayment:     d.MinimumPayment,
			IsSelected:     true,
		})
	}

	input.CalculateTotals()
	return input
}

// DSS Cache invalidation rules
// Rule: mỗi lần "calculate/preview" step A => xoá kết quả từ step A trở đi (A..4)
// Apply step A => xoá kết quả sau step A (A+1..4) vì user vừa chốt step A.

func clearFromStep(state *DSSCachedState, step int) {
	if state == nil {
		return
	}

	if step <= 0 {
		state.AutoScoring = nil
	}
	if step <= 1 {
		state.GoalPrioritizationPreview = nil
		state.AcceptedGoalRanking = nil
	}
	if step <= 2 {
		// Chỉ clear preview, KHÔNG clear AcceptedDebtStrategy (user đã apply rồi)
		state.DebtStrategyPreview = nil
		// state.AcceptedDebtStrategy = "" // KHÔNG clear - user đã apply rồi
	}
	if step <= 3 {
		// Step 3: Budget Allocation
		state.BudgetAllocationPreview = nil
		state.AcceptedScenario = ""
		state.AcceptedAllocations = nil
	}
}

func clearAfterStep(state *DSSCachedState, step int) {
	clearFromStep(state, step+1)
}

// ==================== Initialize DSS Workflow ====================

// InitializeDSS creates a new DSS session with input snapshot cached in Redis
// This should be called FIRST before any preview/apply steps
func (s *monthService) InitializeDSS(ctx context.Context, req dto.InitializeDSSRequest, userID *uuid.UUID) (*dto.InitializeDSSResponse, error) {
	s.logger.Info("Initializing DSS workflow",
		zap.String("month_id", req.MonthID.String()),
		zap.Float64("income", req.MonthlyIncome),
		zap.Int("goals", len(req.Goals)),
		zap.Int("debts", len(req.Debts)),
		zap.Int("constraints", len(req.Constraints)),
	)

	// 1. Validate month ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}
	if !month.CanBeModified() {
		return nil, errors.New("cannot initialize DSS on closed month")
	}

	// 2. XÓA TOÀN BỘ DSS STATE CŨ trong Redis cho (user, month) này
	if err := s.dssCache.ClearState(ctx, req.MonthID, *userID); err != nil {
		s.logger.Warn("Failed to clear existing DSS state before initialize", zap.Error(err))
	}

	// 3. Create new DSS cached state with fresh input snapshot
	cachedState := &DSSCachedState{
		MonthID:          req.MonthID,
		UserID:           *userID,
		UpdatedAt:        time.Now(),
		MonthlyIncome:    req.MonthlyIncome,
		InputGoals:       req.Goals,
		InputDebts:       req.Debts,
		InputConstraints: req.Constraints,
	}

	// 4. Save to Redis (TTL: 3 hours)
	if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
		return nil, fmt.Errorf("failed to initialize DSS state: %w", err)
	}

	s.logger.Info("DSS workflow initialized successfully",
		zap.String("month_id", req.MonthID.String()),
	)

	// 5. Return response
	return &dto.InitializeDSSResponse{
		MonthID:         req.MonthID,
		Status:          "initialized",
		ExpiresIn:       "3h",
		Message:         "DSS workflow initialized. Input snapshot cached for 3 hours.",
		GoalCount:       len(req.Goals),
		DebtCount:       len(req.Debts),
		ConstraintCount: len(req.Constraints),
		MonthlyIncome:   req.MonthlyIncome,
	}, nil
}

// ==================== Step 0: Auto-Scoring Preview ====================

// PreviewAutoScoring runs goal auto-scoring using cached input
func (s *monthService) PreviewAutoScoring(ctx context.Context, req dto.PreviewAutoScoringRequest, userID *uuid.UUID) (*dto.PreviewAutoScoringResponse, error) {
	// 1. Validate month ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. GET CACHED STATE (requires Initialize first)
	cachedState, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSS state: %w", err)
	}

	// 3. Convert cached goals to service input format
	goals := make([]goalDto.GoalForRating, 0, len(cachedState.InputGoals))
	for _, g := range cachedState.InputGoals {
		targetDate, _ := time.Parse("2006-01-02", g.TargetDate)
		goals = append(goals, goalDto.GoalForRating{
			ID:            g.ID,
			Name:          g.Name,
			TargetAmount:  g.TargetAmount,
			CurrentAmount: g.CurrentAmount,
			TargetDate:    targetDate,
			Type:          g.Type,
			Priority:      g.Priority,
		})
	}

	// 4. Call Goal Service with cached data
	// Nếu user cung cấp goal_allocation_pct, adjust MonthlyIncome để phản ánh budget available cho goals
	adjustedIncome := cachedState.MonthlyIncome
	if req.GoalAllocationPct != nil && *req.GoalAllocationPct > 0 && *req.GoalAllocationPct <= 100 {
		adjustedIncome = cachedState.MonthlyIncome * (*req.GoalAllocationPct / 100.0)
	}

	serviceInput := &goalDto.AutoScoresRequest{
		UserID:        userID.String(),
		MonthlyIncome: adjustedIncome,
		Goals:         goals,
	}

	scores, err := s.goalPrioritization.GetAutoScores(ctx, serviceInput)
	if err != nil {
		return nil, fmt.Errorf("auto-scoring failed: %w", err)
	}

	// 5. Update cached state with results
	cachedState.AutoScoring = scores
	if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
		s.logger.Warn("Failed to save auto-scoring results", zap.Error(err))
	}

	return scores, nil
}

// ==================== Step 1: Goal Prioritization ====================

// PreviewGoalPrioritization runs AHP goal prioritization using cached input
func (s *monthService) PreviewGoalPrioritization(ctx context.Context, req dto.PreviewGoalPrioritizationRequest, userID *uuid.UUID) (*dto.PreviewGoalPrioritizationResponse, error) {
	// 1. Validate month ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. GET CACHED STATE (requires Initialize first)
	cachedState, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSS state: %w", err)
	}
	// Re-calc step 1 => clear 1..4 (giữ step0 nếu đã chạy)
	clearFromStep(cachedState, 1)

	// 3. Convert cached goals to service input format
	goals := make([]goalDto.GoalForRating, 0, len(cachedState.InputGoals))
	for _, g := range cachedState.InputGoals {
		targetDate, _ := time.Parse("2006-01-02", g.TargetDate)
		goals = append(goals, goalDto.GoalForRating{
			ID:            g.ID,
			Name:          g.Name,
			TargetAmount:  g.TargetAmount,
			CurrentAmount: g.CurrentAmount,
			TargetDate:    targetDate,
			Type:          g.Type,
			Priority:      g.Priority,
		})
	}

	// 4. Handle edge cases: 0 or 1 goal
	goalsCount := len(goals)
	if goalsCount == 0 {
		// No goals: return empty response
		response := &dto.PreviewGoalPrioritizationResponse{
			AHPOutput: &goalDto.AHPOutput{
				AlternativePriorities: make(map[string]float64),
				CriteriaWeights: map[string]float64{
					"urgency":     0.25,
					"importance":  0.25,
					"feasibility": 0.25,
					"impact":      0.25,
				},
				LocalPriorities:  make(map[string]map[string]float64),
				ConsistencyRatio: 0,
				IsConsistent:     true,
				Ranking:          []goalDomain.RankItem{},
			},
			AutoScores: nil,
		}
		cachedState.GoalPrioritizationPreview = response
		if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
			s.logger.Warn("Failed to save goal prioritization results", zap.Error(err))
		}
		return response, nil
	}

	if goalsCount == 1 {
		// Single goal: automatically rank #1 with priority 1.0 (max)
		goal := goals[0]
		goalID := goal.ID // Already string, no need to convert
		response := &dto.PreviewGoalPrioritizationResponse{
			AHPOutput: &goalDto.AHPOutput{
				AlternativePriorities: map[string]float64{
					goalID: 1.0, // Max priority
				},
				CriteriaWeights: map[string]float64{
					"urgency":     0.25,
					"importance":  0.25,
					"feasibility": 0.25,
					"impact":      0.25,
				},
				LocalPriorities: map[string]map[string]float64{
					"urgency":     {goalID: 1.0},
					"importance":  {goalID: 1.0},
					"feasibility": {goalID: 1.0},
					"impact":      {goalID: 1.0},
				},
				ConsistencyRatio: 0,
				IsConsistent:     true,
				Ranking: []goalDomain.RankItem{
					{
						AlternativeID:   goalID,
						AlternativeName: goal.Name,
						Priority:        1.0, // Max priority
						Rank:            1,
					},
				},
			},
			AutoScores: nil,
		}
		cachedState.GoalPrioritizationPreview = response
		if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
			s.logger.Warn("Failed to save goal prioritization results", zap.Error(err))
		}
		return response, nil
	}

	// 5. Use criteria_weights if provided (preferred), otherwise use criteria_ratings, otherwise default
	var criteriaRatings map[string]int
	if req.CriteriaWeights != nil && len(req.CriteriaWeights) > 0 {
		// Convert weights (0-1) to ratings (1-10) while preserving exact ratios
		// Formula: rating = weight * scale, where scale = 10 / max(weights)
		// This ensures: rating_i / sum(ratings) = weight_i / sum(weights) (exact preservation)
		maxWeight := 0.0
		for _, weight := range req.CriteriaWeights {
			if weight > maxWeight {
				maxWeight = weight
			}
		}

		criteriaRatings = make(map[string]int)
		if maxWeight > 0 {
			scale := 10.0 / maxWeight
			for criterion, weight := range req.CriteriaWeights {
				rating := int(weight * scale)
				if rating < 1 {
					rating = 1
				}
				if rating > 10 {
					rating = 10
				}
				criteriaRatings[criterion] = rating
			}
		} else {
			// All weights are 0, use equal ratings
			for criterion := range req.CriteriaWeights {
				criteriaRatings[criterion] = 5
			}
		}

		s.logger.Info("Using criteria_weights from request",
			zap.Any("original_weights", req.CriteriaWeights),
			zap.Float64("max_weight", maxWeight),
			zap.Any("converted_ratings", criteriaRatings),
		)
	} else if req.CriteriaRatings != nil && len(req.CriteriaRatings) > 0 {
		// Use provided ratings
		criteriaRatings = req.CriteriaRatings
		s.logger.Info("Using criteria_ratings from request", zap.Any("ratings", criteriaRatings))
	} else {
		// Default equal ratings
		criteriaRatings = map[string]int{
			"urgency":     5,
			"importance":  5,
			"feasibility": 5,
			"impact":      5,
		}
		s.logger.Info("Using default criteria_ratings", zap.Any("ratings", criteriaRatings))
	}

	// 6. Build AHP input (>= 2 goals)
	ahpInput := &goalDto.DirectRatingWithOverridesInput{
		DirectRatingInput: goalDto.DirectRatingInput{
			UserID:          userID.String(),
			MonthlyIncome:   cachedState.MonthlyIncome,
			Goals:           goals,
			CriteriaRatings: criteriaRatings,
		},
		GoalScoreOverrides: make(map[string]map[string]float64),
	}

	// 7. Call AHP Service
	result, err := s.goalPrioritization.ExecuteDirectRatingWithOverrides(ctx, ahpInput)
	if err != nil {
		return nil, fmt.Errorf("AHP execution failed: %w", err)
	}

	// 8. Build response
	response := &dto.PreviewGoalPrioritizationResponse{
		AHPOutput:  result,
		AutoScores: nil,
	}

	// 8. Update cached state
	cachedState.GoalPrioritizationPreview = response
	if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
		s.logger.Warn("Failed to save goal prioritization results", zap.Error(err))
	}

	return response, nil
}

// ApplyGoalPrioritization caches the user's accepted goal ranking
func (s *monthService) ApplyGoalPrioritization(ctx context.Context, req dto.ApplyGoalPrioritizationRequest, userID *uuid.UUID) error {
	// 1. Validate ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return errors.New("unauthorized: month does not belong to user")
	}
	if !month.CanBeModified() {
		return errors.New("cannot modify closed month")
	}

	// 2. Get cached state
	state, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return fmt.Errorf("failed to get DSS state: %w", err)
	}

	// 3. Update accepted ranking
	state.AcceptedGoalRanking = req.AcceptedRanking
	// Apply step1 => clear step2..4
	clearAfterStep(state, 1)

	// 4. Save back to Redis
	if err := s.dssCache.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save accepted ranking: %w", err)
	}

	s.logger.Info("Saved accepted goal prioritization",
		zap.String("month_id", req.MonthID.String()),
		zap.Int("goal_count", len(req.AcceptedRanking)),
	)

	return nil
}

// ==================== Step 2: Debt Strategy ====================

// PreviewDebtStrategy runs debt repayment simulations using cached input
func (s *monthService) PreviewDebtStrategy(ctx context.Context, req dto.PreviewDebtStrategyRequest, userID *uuid.UUID) (*dto.PreviewDebtStrategyResponse, error) {
	// 1. Validate month ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. GET CACHED STATE (requires Initialize first)
	cachedState, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSS state: %w", err)
	}
	// Re-calc step2 => clear 2..4
	clearFromStep(cachedState, 2)

	// 3. Convert cached debts to service input format
	// CHỈ đưa revolving debts vào debt strategy (có thể tối ưu extra payment)
	// Installment và interest_only chỉ trả minimum payment cố định, không cần strategy
	debts := make([]debtStrategyDomain.DebtInfo, 0, len(cachedState.InputDebts))
	totalMinPayments := 0.0
	revolvingMinPayments := 0.0
	installmentMinPayments := 0.0

	for _, d := range cachedState.InputDebts {
		totalMinPayments += d.MinimumPayment

		// Xác định behavior: nếu không có thì dùng default
		behavior := getDebtBehavior(d.Behavior)

		if behavior == "revolving" {
			// Chỉ đưa revolving debts vào strategy calculation
			debts = append(debts, debtStrategyDomain.DebtInfo{
				ID:             d.ID,
				Name:           d.Name,
				Balance:        d.CurrentBalance,
				InterestRate:   d.InterestRate / 100, // Convert to decimal
				MinimumPayment: d.MinimumPayment,
				Behavior:       behavior, // "revolving", "installment", "interest_only"
			})
			revolvingMinPayments += d.MinimumPayment
		} else if behavior == "installment" || behavior == "interest_only" {
			// Tính trực tiếp installment/interest_only minimum payments
			installmentMinPayments += d.MinimumPayment
		}
	}

	s.logger.Info("Debt classification",
		zap.Int("total_debts", len(cachedState.InputDebts)),
		zap.Int("revolving_debts", len(debts)),
		zap.Float64("total_min_payments", totalMinPayments),
		zap.Float64("revolving_min_payments", revolvingMinPayments),
		zap.Float64("installment_min_payments", installmentMinPayments),
		zap.Bool("has_revolving", len(debts) > 0),
		zap.Bool("only_installment", len(debts) == 0 && installmentMinPayments > 0),
	)

	// 4. Calculate available debt budget
	// Use debt_allocation_pct from request if provided, otherwise default to 30%
	// IMPORTANT: debt_allocation_pct là % của MonthlyIncome, KHÔNG phải % của disposable
	totalConstraintAmount := 0.0
	for _, c := range cachedState.InputConstraints {
		if !c.IsFlexible {
			totalConstraintAmount += c.MinimumAmount
		}
	}
	disposable := cachedState.MonthlyIncome - totalConstraintAmount

	// Use debt_allocation_pct from request if provided
	debtAllocationPct := 0.3 // Default 30%
	if req.DebtAllocationPct != nil && *req.DebtAllocationPct > 0 && *req.DebtAllocationPct <= 100 {
		debtAllocationPct = *req.DebtAllocationPct / 100.0
		s.logger.Info("Using debt_allocation_pct from request",
			zap.Float64("monthly_income", cachedState.MonthlyIncome),
			zap.Float64("debt_allocation_pct", *req.DebtAllocationPct),
			zap.Float64("calculated_pct", debtAllocationPct),
		)
	} else {
		s.logger.Debug("Using default debt_allocation_pct (30%)",
			zap.Any("request_debt_pct", req.DebtAllocationPct),
		)
	}

	// 1. Lấy total amount payoff từ user (từ slider - debtAllocationPct)
	// TotalDebtBudget = tổng số tiền user muốn phân bổ cho TẤT CẢ debts
	// CRITICAL: Nhân với MonthlyIncome (theo yêu cầu), KHÔNG phải disposable
	totalDebtBudget := cachedState.MonthlyIncome * debtAllocationPct

	// 2. Trừ đi installment/interest_only minimum payments (phải trả cố định)
	// Phần còn lại mới chia cho revolving debts
	revolvingDebtBudget := totalDebtBudget - installmentMinPayments

	// 3. Đảm bảo revolving debt budget ít nhất đủ cho revolving minimum payments
	if revolvingDebtBudget < revolvingMinPayments {
		revolvingDebtBudget = revolvingMinPayments
		// Nếu cần, tăng totalDebtBudget để đảm bảo đủ cho tất cả minimum payments
		if totalDebtBudget < totalMinPayments {
			totalDebtBudget = totalMinPayments
			revolvingDebtBudget = totalDebtBudget - installmentMinPayments
		}
	}

	// Log trường hợp đặc biệt: toàn installment/interest_only, không có revolving
	if len(debts) == 0 && installmentMinPayments > 0 {
		s.logger.Warn("No revolving debts - only installment/interest_only debts",
			zap.Float64("total_debt_budget", totalDebtBudget),
			zap.Float64("installment_min_payments", installmentMinPayments),
			zap.Float64("revolving_debt_budget", revolvingDebtBudget),
			zap.String("note", "All debts are installment/interest_only, no strategy calculation needed"),
		)
	}

	s.logger.Info("Calculated debt budget",
		zap.Float64("disposable_income", disposable),
		zap.Float64("debt_allocation_pct", debtAllocationPct*100),
		zap.Float64("total_debt_budget", totalDebtBudget),
		zap.Float64("revolving_debt_budget", revolvingDebtBudget),
		zap.Float64("total_min_payments", totalMinPayments),
		zap.Float64("revolving_min_payments", revolvingMinPayments),
		zap.Float64("installment_min_payments", installmentMinPayments),
		zap.Int("revolving_debts_count", len(debts)),
	)

	// 5. Build debt strategy input (chỉ cho revolving debts, với budget riêng)
	// Trường hợp đặc biệt: không có revolving debts (toàn installment/interest_only)
	if len(debts) == 0 {
		s.logger.Warn("No revolving debts to optimize - all debts are installment/interest_only",
			zap.Float64("total_debt_budget", totalDebtBudget),
			zap.Float64("installment_min_payments", installmentMinPayments),
			zap.String("action", "Will create empty strategy result with only installment/interest_only payment plans"),
		)
		// Tạo empty result với payment plans chỉ có installment/interest_only
		// (sẽ được thêm ở bước 6.5)
	}

	debtInput := &debtStrategyService.DebtStrategyInput{
		UserID:          userID.String(),
		TotalDebtBudget: revolvingDebtBudget, // Budget chỉ cho revolving debts (có thể = 0 nếu không có revolving)
		Debts:           debts,               // Có thể empty nếu không có revolving
	}

	if req.PreferredStrategy != "" {
		debtInput.PreferredStrategy = debtStrategyDomain.Strategy(req.PreferredStrategy)
	}

	// 6. Call Debt Strategy Service
	// Nếu không có revolving debts, service sẽ trả về empty result
	result, err := s.debtStrategy.ExecuteDebtStrategy(ctx, debtInput)
	if err != nil {
		return nil, fmt.Errorf("debt strategy analysis failed: %w", err)
	}

	s.logger.Info("Debt strategy result",
		zap.Int("strategy_comparison_count", len(result.StrategyComparison)),
		zap.Int("revolving_debts_input", len(debts)),
		zap.Bool("has_revolving", len(debts) > 0),
	)

	// 6.5. Nếu không có revolving debts (chỉ có installment/interest_only), tạo một strategy comparison entry mặc định
	if len(debts) == 0 && len(result.StrategyComparison) == 0 && installmentMinPayments > 0 {
		s.logger.Info("Creating default strategy comparison for installment-only debts",
			zap.Float64("installment_min_payments", installmentMinPayments),
			zap.Float64("total_debt_budget", totalDebtBudget),
		)
		// Tạo payment plans cho tất cả installment/interest_only debts
		paymentPlans := make([]debtStrategyDomain.PaymentPlan, 0)
		for _, d := range cachedState.InputDebts {
			behavior := getDebtBehavior(d.Behavior)
			if behavior == "installment" || behavior == "interest_only" {
				paymentPlans = append(paymentPlans, debtStrategyDomain.PaymentPlan{
					DebtID:         d.ID,
					DebtName:       d.Name,
					MonthlyPayment: d.MinimumPayment,
					ExtraPayment:   0,
					PayoffMonth:    999, // Placeholder
					TotalInterest:  0,
				})
			}
		}

		result.StrategyComparison = append(result.StrategyComparison, debtStrategyDomain.StrategyComparison{
			Strategy:          debtStrategyDomain.StrategyAvalanche, // Dùng strategy có sẵn
			TotalInterest:     0,
			Months:            0,
			InterestSaved:     0,
			FirstDebtCleared:  0,
			Description:       "Fixed payment plan for installment and interest-only debts",
			Pros:              []string{"Predictable payments", "No strategy optimization needed"},
			Cons:              []string{},
			PaymentPlans:      paymentPlans,
			MonthlyAllocation: installmentMinPayments, // Sẽ được recalculate ở bước 6.6
		})
		// Set recommended strategy
		result.RecommendedStrategy = debtStrategyDomain.StrategyAvalanche
		result.Reasoning = "All debts are installment or interest-only. Fixed minimum payments apply."
		result.KeyFacts = []string{
			"All debts require fixed minimum payments",
			"No optimization strategy needed",
		}
		// Set payment plans cho output chính
		result.PaymentPlans = paymentPlans
		// Set fixed payments để frontend hiển thị (không cần chart)
		result.FixedPayments = paymentPlans
	}

	// 6.6. Collect fixed payments (installment/interest_only) để populate FixedPayments field
	// Chỉ collect một lần, không cần loop
	fixedPayments := make([]debtStrategyDomain.PaymentPlan, 0)
	for _, d := range cachedState.InputDebts {
		behavior := getDebtBehavior(d.Behavior)
		if behavior == "installment" || behavior == "interest_only" {
			fixedPayments = append(fixedPayments, debtStrategyDomain.PaymentPlan{
				DebtID:         d.ID,
				DebtName:       d.Name,
				MonthlyPayment: d.MinimumPayment,
				ExtraPayment:   0,
				PayoffMonth:    999,
				TotalInterest:  0,
			})
		}
	}
	// Set fixed payments để frontend hiển thị (không cần chart)
	if len(fixedPayments) > 0 {
		result.FixedPayments = fixedPayments
	}

	// 6.7. Thêm installment/interest_only debts vào payment plans để đảm bảo monthly allocation = TotalDebtBudget
	// TotalDebtBudget (30tr) = revolving debts allocation + installment/interest_only minimum payments
	// Tạo map để track debts đã có trong payment plans
	for i := range result.StrategyComparison {
		existingDebtIDs := make(map[string]bool)
		for _, plan := range result.StrategyComparison[i].PaymentPlans {
			existingDebtIDs[plan.DebtID] = true
		}

		// Thêm installment/interest_only debts với minimum payment
		for _, d := range cachedState.InputDebts {
			behavior := getDebtBehavior(d.Behavior)
			if behavior == "installment" || behavior == "interest_only" {
				// Chỉ thêm nếu chưa có trong payment plans
				if !existingDebtIDs[d.ID] {
					fixedPlan := debtStrategyDomain.PaymentPlan{
						DebtID:         d.ID,
						DebtName:       d.Name,
						MonthlyPayment: d.MinimumPayment, // Installment/interest_only chỉ trả minimum
						ExtraPayment:   0,                // Không có extra payment
						PayoffMonth:    999,              // Placeholder - không tính payoff month cho installment
						TotalInterest:  0,                // Không tính interest cho payment plan này
					}
					result.StrategyComparison[i].PaymentPlans = append(result.StrategyComparison[i].PaymentPlans, fixedPlan)
				}
			}
		}

		// Recalculate monthly allocation = tổng của tất cả payment plans (revolving + installment/interest_only)
		totalAllocation := 0.0
		revolvingPaymentCount := 0
		installmentPaymentCount := 0
		for _, plan := range result.StrategyComparison[i].PaymentPlans {
			totalAllocation += plan.MonthlyPayment
			if plan.ExtraPayment > 0 {
				revolvingPaymentCount++
			} else {
				installmentPaymentCount++
			}
		}

		// Đảm bảo monthly allocation đúng
		// Nếu chỉ có installment/interest_only (không có revolving), monthly allocation = tổng minimum payments
		// Nếu có revolving debts, monthly allocation = TotalDebtBudget (budget thực tế user đã set)
		if len(debts) == 0 {
			// Chỉ có installment/interest_only: monthly allocation = tổng minimum payments
			result.StrategyComparison[i].MonthlyAllocation = installmentMinPayments
		} else {
			// Có revolving debts: monthly allocation = TotalDebtBudget (bao gồm cả installment/interest_only minimum payments)
			if totalAllocation < totalDebtBudget {
				// Có thể có rounding errors, set = TotalDebtBudget để đảm bảo phân bổ hết
				result.StrategyComparison[i].MonthlyAllocation = totalDebtBudget
			} else {
				result.StrategyComparison[i].MonthlyAllocation = totalAllocation
			}
		}

		// Log chi tiết cho trường hợp đặc biệt
		if len(debts) == 0 {
			s.logger.Info("Strategy comparison for installment-only scenario",
				zap.String("strategy", string(result.StrategyComparison[i].Strategy)),
				zap.Float64("total_allocation", totalAllocation),
				zap.Float64("total_debt_budget", totalDebtBudget),
				zap.Float64("installment_min_payments", installmentMinPayments),
				zap.Int("payment_plans_count", len(result.StrategyComparison[i].PaymentPlans)),
				zap.Int("installment_payment_count", installmentPaymentCount),
				zap.Int("revolving_payment_count", revolvingPaymentCount),
			)
		}
	}

	// 7. Tính fuzzy-based extra weights cho từng chiến lược dựa trên PaymentPlans
	strategyWeights := make(map[string]map[uuid.UUID]float64)
	for _, sc := range result.StrategyComparison {
		weights := make(map[uuid.UUID]float64)
		totalExtra := 0.0

		for _, plan := range sc.PaymentPlans {
			if plan.ExtraPayment <= 0 {
				continue
			}
			debtID, err := uuid.Parse(plan.DebtID)
			if err != nil {
				continue
			}
			weights[debtID] += plan.ExtraPayment
			totalExtra += plan.ExtraPayment
		}

		if totalExtra <= 0 {
			continue
		}

		for id, v := range weights {
			weights[id] = v / totalExtra
		}

		strategyKey := string(sc.Strategy)
		strategyWeights[strategyKey] = weights
	}

	// Lưu toàn bộ map chiến lược -> weights để Step 4 có thể dùng theo AcceptedDebtStrategy
	if len(strategyWeights) > 0 {
		cachedState.DebtAllocationWeights = strategyWeights
	} else {
		cachedState.DebtAllocationWeights = nil
	}

	// 8. Update cached state
	cachedState.DebtStrategyPreview = result
	if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
		s.logger.Warn("Failed to save debt strategy results", zap.Error(err))
	}

	return result, nil
}

// ApplyDebtStrategy caches the user's selected debt strategy
func (s *monthService) ApplyDebtStrategy(ctx context.Context, req dto.ApplyDebtStrategyRequest, userID *uuid.UUID) error {
	// 1. Validate ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return errors.New("unauthorized: month does not belong to user")
	}
	if !month.CanBeModified() {
		return errors.New("cannot modify closed month")
	}

	// 2. Get cached state
	state, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return fmt.Errorf("failed to get DSS state: %w", err)
	}

	// 3. Update accepted strategy
	state.AcceptedDebtStrategy = req.SelectedStrategy
	// Apply step2 => clear step3..4
	clearAfterStep(state, 2)

	// 4. Save back to Redis
	if err := s.dssCache.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save accepted strategy: %w", err)
	}

	s.logger.Info("Saved accepted debt strategy",
		zap.String("month_id", req.MonthID.String()),
		zap.String("strategy", req.SelectedStrategy),
	)

	return nil
}

// ==================== Step 4: Budget Allocation ====================

// PreviewBudgetAllocation runs Goal Programming allocation using cached data
func (s *monthService) PreviewBudgetAllocation(ctx context.Context, req dto.PreviewBudgetAllocationRequest, userID *uuid.UUID) (*dto.PreviewBudgetAllocationResponse, error) {
	// 1. Validate month ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. GET CACHED STATE (requires Initialize first)
	cachedState, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSS state: %w", err)
	}
	// Re-calc step 3 (Budget Allocation) => clear step 3
	clearFromStep(cachedState, 3)

	// 3. Nếu user chưa chọn tỉ lệ goal/debt (0-0) thì set default theo ngữ cảnh
	goalPct := req.GoalAllocationPct
	debtPct := 0.0
	if req.DebtAllocationPct != nil {
		debtPct = *req.DebtAllocationPct
	}
	if goalPct == 0 && debtPct == 0 {
		goalsCount := len(cachedState.InputGoals)
		debtsCount := len(cachedState.InputDebts)

		switch {
		case goalsCount > 0 && debtsCount == 0:
			// Không có nợ, dồn toàn bộ phần “tối ưu” cho goals
			goalPct = 100
			debtPct = 0
		case goalsCount == 0 && debtsCount > 0:
			// Không có goals, toàn bộ cho trả nợ
			goalPct = 0
			debtPct = 100
		case goalsCount == 0 && debtsCount == 0:
			// Không có cả goals và debts: toàn bộ cho flexible expenses/surplus
			goalPct = 0
			debtPct = 0
		default:
			// Có cả 2 mà chưa chọn gì: chia đôi
			goalPct = 50
			debtPct = 50
		}
	}

	// 3.1. Lấy PaymentPlans từ debt strategy preview để ghi đè payment nếu lớn hơn minimum
	// Tạo map debtID -> PaymentPlan từ strategy đã được accept
	debtPaymentPlans := make(map[string]debtStrategyDomain.PaymentPlan)
	if cachedState.DebtStrategyPreview != nil && cachedState.AcceptedDebtStrategy != "" {
		// Parse DebtStrategyPreview (có thể là JSON đã unmarshal hoặc struct)
		if previewMap, ok := cachedState.DebtStrategyPreview.(map[string]interface{}); ok {
			// Tìm strategy comparison cho accepted strategy
			if strategyComparison, ok := previewMap["strategy_comparison"].([]interface{}); ok {
				for _, scInterface := range strategyComparison {
					if sc, ok := scInterface.(map[string]interface{}); ok {
						if strategy, ok := sc["strategy"].(string); ok && strategy == cachedState.AcceptedDebtStrategy {
							// Tìm thấy strategy đã accept, lấy payment plans
							if paymentPlans, ok := sc["payment_plans"].([]interface{}); ok {
								for _, planInterface := range paymentPlans {
									if planMap, ok := planInterface.(map[string]interface{}); ok {
										if debtID, ok := planMap["debt_id"].(string); ok {
											if monthlyPayment, ok := planMap["monthly_payment"].(float64); ok {
												plan := debtStrategyDomain.PaymentPlan{
													DebtID:         debtID,
													MonthlyPayment: monthlyPayment,
												}
												if extraPayment, ok := planMap["extra_payment"].(float64); ok {
													plan.ExtraPayment = extraPayment
												}
												if payoffMonth, ok := planMap["payoff_month"].(float64); ok {
													plan.PayoffMonth = int(payoffMonth)
												}
												debtPaymentPlans[debtID] = plan
											}
										}
									}
								}
							}
							break
						}
					}
				}
			}
			// Nếu không tìm thấy trong strategy_comparison, thử lấy từ payment_plans trực tiếp
			if len(debtPaymentPlans) == 0 {
				if paymentPlans, ok := previewMap["payment_plans"].([]interface{}); ok {
					for _, planInterface := range paymentPlans {
						if planMap, ok := planInterface.(map[string]interface{}); ok {
							if debtID, ok := planMap["debt_id"].(string); ok {
								if monthlyPayment, ok := planMap["monthly_payment"].(float64); ok {
									plan := debtStrategyDomain.PaymentPlan{
										DebtID:         debtID,
										MonthlyPayment: monthlyPayment,
									}
									if extraPayment, ok := planMap["extra_payment"].(float64); ok {
										plan.ExtraPayment = extraPayment
									}
									if payoffMonth, ok := planMap["payoff_month"].(float64); ok {
										plan.PayoffMonth = int(payoffMonth)
									}
									debtPaymentPlans[debtID] = plan
								}
							}
						}
					}
				}
			}
		}
	}

	// 4. Build Allocation Input
	allocationInput := &budgetAllocationService.BudgetAllocationModelInput{
		UserID:            *userID,
		Year:              time.Now().Year(),
		Month:             int(time.Now().Month()),
		TotalIncome:       cachedState.MonthlyIncome,
		UseAllScenarios:   true,
		MandatoryExpenses: make([]budgetAllocationService.MandatoryExpense, 0),
		FlexibleExpenses:  make([]budgetAllocationService.FlexibleExpense, 0),
		Debts:             make([]budgetAllocationService.DebtInput, 0),
		Goals:             make([]budgetAllocationService.GoalInput, 0),
	}

	// 5. Map constraints từ cache
	// Heuristic:
	// - Mandatory (IsFlexible=false): dùng MinimumAmount làm fixed amount
	// - Flexible (IsFlexible=true): MinAmount = MinimumAmount, MaxAmount = MaximumAmount (target range = Max - Min)
	// Tạo map để track categories đã có từ constraints
	constraintCategoryMap := make(map[uuid.UUID]bool)
	for _, c := range cachedState.InputConstraints {
		constraintCategoryMap[c.CategoryID] = true
		if !c.IsFlexible {
			// Mandatory: fixed = MinimumAmount
			allocationInput.MandatoryExpenses = append(allocationInput.MandatoryExpenses, budgetAllocationService.MandatoryExpense{
				CategoryID: c.CategoryID,
				Name:       c.Name,
				Amount:     c.MinimumAmount, // Fixed heuristic amount
				Priority:   c.Priority,
			})
		} else {

			allocationInput.FlexibleExpenses = append(allocationInput.FlexibleExpenses, budgetAllocationService.FlexibleExpense{
				CategoryID: c.CategoryID,
				Name:       c.Name,
				MinAmount:  c.MinimumAmount, // Heuristic minimum
				MaxAmount:  c.MaximumAmount, // Target range = MaxAmount - MinAmount
				Priority:   c.Priority,
			})
		}
	}

	// 5.1. Load active budgets cho tháng hiện tại và thêm vào flexible expenses
	// Budgets hiện tại được coi như flexible expenses với min = max = amount
	startDate, endDate := CalculateMonthBoundaries(month.Month)
	activeBudgets, err := s.budgetService.GetActiveBudgets(ctx, *userID)
	if err != nil {
		s.logger.Warn("Failed to load active budgets for allocation", zap.Error(err))
	} else {
		// Filter budgets overlap với tháng hiện tại và có CategoryID
		for _, budget := range activeBudgets {
			if budget.CategoryID == nil {
				continue // Skip budgets không có category
			}

			// Check overlap: budget.StartDate <= endDate && (budget.EndDate == nil || budget.EndDate >= startDate)
			budgetEnd := endDate
			if budget.EndDate != nil {
				budgetEnd = *budget.EndDate
			}
			if budget.StartDate.After(endDate) || budgetEnd.Before(startDate) {
				continue // Không overlap với tháng hiện tại
			}

			// Nếu category đã có trong constraints, merge hoặc skip (tùy logic)
			// Ở đây ta ưu tiên budget nếu chưa có constraint cho category này
			if !constraintCategoryMap[*budget.CategoryID] {
				// Thêm budget như flexible expense
				// Heuristic: min = budget.Amount, target range = (max - min) = budget.Amount * 0.2 (20% buffer)
				allocationInput.FlexibleExpenses = append(allocationInput.FlexibleExpenses, budgetAllocationService.FlexibleExpense{
					CategoryID: *budget.CategoryID,
					Name:       budget.Name,
					MinAmount:  budget.Amount,       // Heuristic minimum
					MaxAmount:  budget.Amount * 1.2, // Target range = MaxAmount - MinAmount = 20% of budget
					Priority:   len(allocationInput.FlexibleExpenses) + 1,
				})
				s.logger.Info("Added active budget to allocation input",
					zap.String("budget_id", budget.ID.String()),
					zap.String("category_id", budget.CategoryID.String()),
					zap.String("name", budget.Name),
					zap.Float64("min_amount", budget.Amount),
					zap.Float64("max_amount", budget.Amount*1.2),
					zap.Float64("target_range", budget.Amount*0.2),
				)
			} else {
				// Category đã có constraint, merge: update min nếu budget lớn hơn
				for i := range allocationInput.FlexibleExpenses {
					if allocationInput.FlexibleExpenses[i].CategoryID == *budget.CategoryID {
						// Update min amount nếu budget lớn hơn, và đảm bảo max >= min
						if budget.Amount > allocationInput.FlexibleExpenses[i].MinAmount {
							oldMin := allocationInput.FlexibleExpenses[i].MinAmount
							allocationInput.FlexibleExpenses[i].MinAmount = budget.Amount
							// Đảm bảo max >= min (nếu max < min mới thì tăng max)
							if allocationInput.FlexibleExpenses[i].MaxAmount < budget.Amount {
								allocationInput.FlexibleExpenses[i].MaxAmount = budget.Amount * 1.2
							}
							s.logger.Info("Updated flexible expense from active budget",
								zap.String("category_id", budget.CategoryID.String()),
								zap.Float64("old_min", oldMin),
								zap.Float64("new_min", budget.Amount),
								zap.Float64("max", allocationInput.FlexibleExpenses[i].MaxAmount),
								zap.Float64("target_range", allocationInput.FlexibleExpenses[i].MaxAmount-budget.Amount),
							)
						}
						break
					}
				}
			}
		}
	}

	// 4. Tính toán allocation percentage cho goals

	// 5. Map goals from cache với allocation percentage
	goalAllocationPct := goalPct / 100.0
	for _, g := range cachedState.InputGoals {
		remaining := g.TargetAmount - g.CurrentAmount
		if remaining > 0 {
			// Calculate suggested contribution based on months remaining
			var suggestedContribution float64

			// If target date is provided, calculate based on months remaining
			if g.TargetDate != "" {
				targetDate, err := time.Parse("2006-01-02", g.TargetDate)
				if err == nil {
					daysRemaining := int(targetDate.Sub(time.Now()).Hours() / 24)
					if daysRemaining > 0 {
						monthsRemaining := float64(daysRemaining) / 30.0
						if monthsRemaining > 0 {
							suggestedContribution = remaining / monthsRemaining
						} else {
							// Overdue: suggest all remaining
							suggestedContribution = remaining
						}
					} else {
						// Overdue: suggest all remaining
						suggestedContribution = remaining
					}
				} else {
					// Invalid date: default to 12 months
					suggestedContribution = remaining / 12.0
				}
			} else {
				// No target date: default to 12 months
				suggestedContribution = remaining / 12.0
			}

			// Apply goal allocation percentage (adjust suggested contribution)
			suggestedContribution = suggestedContribution * goalAllocationPct

			allocationInput.Goals = append(allocationInput.Goals, budgetAllocationService.GoalInput{
				GoalID:                uuid.MustParse(g.ID),
				Name:                  g.Name,
				Type:                  g.Type,
				Priority:              g.Priority,
				RemainingAmount:       remaining,
				SuggestedContribution: suggestedContribution,
			})
		}
	}

	// 6. Map debts from cache - lấy payment từ PaymentPlan nếu có, nếu không thì dùng minimum payment
	// Nếu PaymentPlan.MonthlyPayment > MinimumPayment thì ghi đè lên
	for _, d := range cachedState.InputDebts {
		adjustedPayment := d.MinimumPayment

		// Lấy payment từ PaymentPlan nếu có
		if plan, ok := debtPaymentPlans[d.ID]; ok {
			// Nếu MonthlyPayment từ plan lớn hơn MinimumPayment thì ghi đè
			if plan.MonthlyPayment > d.MinimumPayment {
				adjustedPayment = plan.MonthlyPayment
			}
		}

		allocationInput.Debts = append(allocationInput.Debts, budgetAllocationService.DebtInput{
			DebtID:         uuid.MustParse(d.ID),
			Name:           d.Name,
			Balance:        d.CurrentBalance,
			InterestRate:   d.InterestRate / 100,
			MinimumPayment: adjustedPayment, // This is from debt strategy output, will be forced as FixedPayment
		})
	}

	// 7.5. Map custom scenario parameters from request (if provided)
	if len(req.ScenarioOverrides) > 0 {
		customParams := make([]budgetAllocationService.CustomScenarioParams, 0)
		for _, override := range req.ScenarioOverrides {
			// Validate percentages sum to <= 1.0
			totalPercent := 0.0
			if override.EmergencyFundPercent != nil {
				totalPercent += *override.EmergencyFundPercent
			}
			if override.GoalsPercent != nil {
				totalPercent += *override.GoalsPercent
			}
			if override.FlexiblePercent != nil {
				totalPercent += *override.FlexiblePercent
			}
			if totalPercent > 1.0 {
				s.logger.Warn("Custom scenario parameters exceed 100%",
					zap.String("scenario_type", override.ScenarioType),
					zap.Float64("total_percent", totalPercent),
				)
				// Normalize to 100%
				scale := 1.0 / totalPercent
				if override.EmergencyFundPercent != nil {
					*override.EmergencyFundPercent *= scale
				}
				if override.GoalsPercent != nil {
					*override.GoalsPercent *= scale
				}
				if override.FlexiblePercent != nil {
					*override.FlexiblePercent *= scale
				}
			}

			customParam := budgetAllocationService.CustomScenarioParams{
				ScenarioType:           override.ScenarioType,
				GoalContributionFactor: 1.0, // Default
				FlexibleSpendingLevel:  0.5, // Default
				EmergencyFundPercent:   0.4, // Default
				GoalsPercent:           0.2, // Default
				FlexiblePercent:        0.1, // Default
			}

			if override.GoalContributionFactor != nil {
				customParam.GoalContributionFactor = *override.GoalContributionFactor
			}
			if override.FlexibleSpendingLevel != nil {
				customParam.FlexibleSpendingLevel = *override.FlexibleSpendingLevel
			}
			if override.EmergencyFundPercent != nil {
				customParam.EmergencyFundPercent = *override.EmergencyFundPercent
			}
			if override.GoalsPercent != nil {
				customParam.GoalsPercent = *override.GoalsPercent
			}
			if override.FlexiblePercent != nil {
				customParam.FlexiblePercent = *override.FlexiblePercent
			}

			customParams = append(customParams, customParam)
		}
		allocationInput.CustomScenarioParams = customParams
	}

	// 7.9. Truyền debt strategy preview để lấy PayoffMonth cho mỗi debt
	if cachedState.DebtStrategyPreview != nil {
		allocationInput.DebtStrategyPreview = cachedState.DebtStrategyPreview
	}

	// 8. Call Budget Allocation Service
	result, err := s.budgetAllocation.ExecuteBudgetAllocation(ctx, allocationInput)
	if err != nil {
		return nil, fmt.Errorf("budget allocation failed: %w", err)
	}

	// 9. Update cached state
	cachedState.BudgetAllocationPreview = result
	if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
		s.logger.Warn("Failed to save budget allocation results", zap.Error(err))
	}

	return result, nil
}

// ==================== Workflow Management ====================

// GetDSSWorkflowStatus returns the current workflow state
func (s *monthService) GetDSSWorkflowStatus(ctx context.Context, monthID uuid.UUID, userID *uuid.UUID) (*dto.DSSWorkflowStatusResponse, error) {
	// 1. Load Month
	month, err := s.repo.GetMonthByID(ctx, monthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}

	// 2. Validate ownership
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 3. Get current state
	state := month.CurrentState()

	// 4. Build response
	response := &dto.DSSWorkflowStatusResponse{
		MonthID:        monthID,
		CurrentStep:    0,
		CompletedSteps: []int{},
		CanProceed:     true,
		IsComplete:     false,
	}

	if state.DSSWorkflow != nil {
		response.CurrentStep = state.DSSWorkflow.CurrentStep
		response.CompletedSteps = state.DSSWorkflow.CompletedSteps
		response.IsComplete = state.DSSWorkflow.IsComplete()
		response.CanProceed = !response.IsComplete
		response.LastUpdated = &state.DSSWorkflow.LastUpdated

		response.Step1Applied = state.DSSWorkflow.AppliedAtStep1 != nil
		response.Step2Applied = state.DSSWorkflow.AppliedAtStep2 != nil
		response.Step3Applied = false                                   // Step 3 tradeoff removed
		response.Step4Applied = state.DSSWorkflow.AppliedAtStep3 != nil // Step 3 giờ là Budget Allocation
	}

	return response, nil
}

// ResetDSSWorkflow clears all DSS workflow results
func (s *monthService) ResetDSSWorkflow(ctx context.Context, req dto.ResetDSSWorkflowRequest, userID *uuid.UUID) error {
	// 1. Load Month
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return fmt.Errorf("failed to load month: %w", err)
	}

	// 2. Validate ownership
	if userID != nil && month.UserID != *userID {
		return errors.New("unauthorized: month does not belong to user")
	}

	// 3. Validate month is open
	if !month.CanBeModified() {
		return errors.New("cannot reset workflow on closed month")
	}

	// 4. Clear Redis cache
	if userID != nil {
		if err := s.dssCache.ClearState(ctx, req.MonthID, *userID); err != nil {
			s.logger.Warn("Failed to clear DSS cache", zap.Error(err))
		}
	}

	// 5. Reset DB workflow if exists
	state := month.CurrentState()
	if state.DSSWorkflow != nil {
		state.DSSWorkflow.Reset()
	} else {
		state.DSSWorkflow = &domain.DSSWorkflowResults{
			StartedAt:      time.Now(),
			LastUpdated:    time.Now(),
			CurrentStep:    0,
			CompletedSteps: []int{},
		}
	}

	state.IsApplied = false

	month.IncrementVersion()
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to save month: %w", err)
	}

	s.logger.Info("Reset DSS workflow",
		zap.String("month_id", req.MonthID.String()),
	)

	return nil
}

// ==================== DSS Finalization ====================

// FinalizeDSS applies all DSS results at once from frontend
// IMPORTANT: Backend ONLY saves what frontend sends. All allocations (categories, goals, debts) come from FE.
// Frontend/user is responsible for computing all allocations.
func (s *monthService) FinalizeDSS(ctx context.Context, req dto.FinalizeDSSRequest, monthID uuid.UUID, userID *uuid.UUID) (*dto.FinalizeDSSResponse, error) {
	s.logger.Info("Finalizing DSS workflow with all allocations from frontend",
		zap.String("month_id", monthID.String()),
		zap.Int("budget_allocations", len(req.BudgetAllocations)),
		zap.Int("goal_fundings", len(req.GoalFundings)),
		zap.Int("debt_payments", len(req.DebtPayments)),
	)

	// 1. Validate ownership
	month, err := s.repo.GetMonthByID(ctx, monthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}
	if !month.CanBeModified() {
		return nil, errors.New("cannot modify closed month")
	}

	// 2. Validate allocations provided
	if len(req.BudgetAllocations) == 0 {
		return nil, errors.New("no budget allocations provided")
	}

	// 2b. Validate allocation amounts (must be positive)
	totalBudgetAllocated := 0.0
	for catID, amount := range req.BudgetAllocations {
		if amount < 0 {
			return nil, fmt.Errorf("invalid allocation amount for category %s: %f (must be >= 0)", catID.String(), amount)
		}
		totalBudgetAllocated += amount
	}

	// Validate goal fundings
	totalGoalFundings := 0.0
	for _, gf := range req.GoalFundings {
		amount := gf.SuggestedAmount
		if gf.UserAdjustedAmount != nil && *gf.UserAdjustedAmount > 0 {
			amount = *gf.UserAdjustedAmount
		}
		if amount < 0 {
			return nil, fmt.Errorf("invalid goal funding amount for goal %s: %f (must be >= 0)", gf.GoalID.String(), amount)
		}
		totalGoalFundings += amount
	}

	// Validate debt payments
	totalDebtPayments := 0.0
	for _, dp := range req.DebtPayments {
		amount := dp.SuggestedPayment
		if dp.UserAdjustedPayment != nil && *dp.UserAdjustedPayment > 0 {
			amount = *dp.UserAdjustedPayment
		}
		if amount < 0 {
			return nil, fmt.Errorf("invalid debt payment amount for debt %s: %f (must be >= 0)", dp.DebtID.String(), amount)
		}
		if amount < dp.MinimumPayment {
			s.logger.Warn("Debt payment is less than minimum payment",
				zap.String("debt_id", dp.DebtID.String()),
				zap.Float64("amount", amount),
				zap.Float64("minimum_payment", dp.MinimumPayment),
			)
			// Don't return error - allow user to pay less (they might adjust later)
		}
		totalDebtPayments += amount
	}

	// Calculate total allocations (categories + goals + debts)
	totalAllocated := totalBudgetAllocated + totalGoalFundings + totalDebtPayments

	// 3. Get cached state from Redis (for metadata like monthly income)
	cachedState, err := s.dssCache.GetOrInitState(ctx, monthID, *userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DSS state: %w", err)
	}

	// 3b. Validate total allocation doesn't exceed income (warning only)
	if totalAllocated > cachedState.MonthlyIncome {
		s.logger.Warn("Total allocations exceed monthly income",
			zap.Float64("total_allocated", totalAllocated),
			zap.Float64("budget_allocated", totalBudgetAllocated),
			zap.Float64("goal_fundings", totalGoalFundings),
			zap.Float64("debt_payments", totalDebtPayments),
			zap.Float64("monthly_income", cachedState.MonthlyIncome),
			zap.Float64("deficit", totalAllocated-cachedState.MonthlyIncome),
		)
		// Don't return error - allow user to allocate more than income (they might have other income sources)
	}

	// 4. Create budgets from category allocations (exactly as FE sends)
	startDate, endDate := CalculateMonthBoundaries(month.Month)
	createdCount := 0
	skippedCount := 0

	for catID, amount := range req.BudgetAllocations {
		if amount <= 0 {
			skippedCount++
			continue
		}

		// Find category name and constraint ID from cached state
		categoryName := "Unknown"
		var constraintID *uuid.UUID
		for _, c := range cachedState.InputConstraints {
			if c.CategoryID == catID {
				categoryName = c.Name
				if c.ID != "" {
					if parsedID, err := uuid.Parse(c.ID); err == nil {
						constraintID = &parsedID
					}
				}
				break
			}
		}

		budget := &budgetDomain.Budget{
			ID:           uuid.New(),
			UserID:       month.UserID,
			Name:         fmt.Sprintf("%s - %s", categoryName, month.Month),
			Amount:       amount,
			Currency:     "VND",
			Period:       budgetDomain.BudgetPeriodMonthly,
			StartDate:    startDate,
			EndDate:      &endDate,
			CategoryID:   &catID,
			ConstraintID: constraintID,
			EnableAlerts: true,
			AlertThresholds: budgetDomain.AlertThresholdsJSON{
				budgetDomain.AlertAt75,
				budgetDomain.AlertAt90,
				budgetDomain.AlertAt100,
			},
		}

		if err := s.budgetService.CreateBudgetFromDomain(ctx, budget); err != nil {
			s.logger.Warn("Failed to create budget for category, skipping",
				zap.String("category_id", catID.String()),
				zap.String("category_name", categoryName),
				zap.Error(err),
			)
			continue
		}

		createdCount++
		s.logger.Info("Created budget from finalize request",
			zap.String("budget_id", budget.ID.String()),
			zap.String("category_id", catID.String()),
			zap.String("category_name", categoryName),
			zap.Float64("amount", amount),
		)
	}

	s.logger.Info("Budget creation summary",
		zap.Int("created", createdCount),
		zap.Int("skipped", skippedCount),
		zap.Int("total_categories", len(req.BudgetAllocations)),
	)

	// 5. Get current MonthState
	state := month.CurrentState()
	if state == nil {
		initial := domain.NewMonthState()
		month.States = []domain.MonthState{*initial}
		state = &month.States[0]
	}

	// 5b. Fix initial state metadata
	if state.Version == 0 && state.CreatedAt.IsZero() {
		state.Version = 1
		state.CreatedAt = time.Now()
	}

	// 6. Create NEW state with all DSS results
	newState := domain.MonthState{
		Version:        0, // set below
		CreatedAt:      time.Now(),
		IsApplied:      true,
		ToBeBudgeted:   0,
		ActualIncome:   cachedState.MonthlyIncome,
		CategoryStates: []domain.CategoryState{},
		Input:          buildInputSnapshotFromCachedState(cachedState, time.Now()),
		DSSWorkflow:    nil,
	}

	// 7. Build DSSWorkflowResults from FinalizeDSSRequest (all from FE)
	workflow := &domain.DSSWorkflowResults{
		StartedAt:      time.Now(),
		LastUpdated:    time.Now(),
		CurrentStep:    3,
		CompletedSteps: []int{},
	}

	// Step 1: Goal Prioritization (from FE)
	if len(req.GoalPriorities) > 0 {
		rankings := make([]domain.GoalRanking, len(req.GoalPriorities))
		for i, gp := range req.GoalPriorities {
			rankings[i] = domain.GoalRanking{
				GoalID: gp.GoalID,
				Rank:   i + 1,
				Score:  gp.Priority,
			}
		}
		workflow.GoalPrioritization = &domain.GoalPriorityResult{
			Rankings: rankings,
			Method:   req.GoalPriorities[0].Method,
		}
		now := time.Now()
		workflow.AppliedAtStep1 = &now
		workflow.MarkStepCompleted(1)
	}

	// Step 2: Debt Strategy (from FE)
	if req.DebtStrategy != nil && *req.DebtStrategy != "" {
		workflow.DebtStrategy = &domain.DebtStrategyResult{Strategy: *req.DebtStrategy}
		now := time.Now()
		workflow.AppliedAtStep2 = &now
		workflow.MarkStepCompleted(2)
	}

	// Step 3: Budget Allocation (from FE - categories, goals, debts)
	if len(req.BudgetAllocations) > 0 {
		// Build goal fundings from FE request
		goalFundings := make([]domain.GoalFundingItem, 0)
		for _, gf := range req.GoalFundings {
			// Use user adjusted amount if provided, otherwise use suggested
			amount := gf.SuggestedAmount
			if gf.UserAdjustedAmount != nil && *gf.UserAdjustedAmount > 0 {
				amount = *gf.UserAdjustedAmount
			}
			if amount > 0 {
				// Find goal name from cached state
				goalName := "Unknown"
				for _, g := range cachedState.InputGoals {
					if g.ID == gf.GoalID.String() {
						goalName = g.Name
						break
					}
				}
				goalFundings = append(goalFundings, domain.GoalFundingItem{
					GoalID:     gf.GoalID,
					GoalName:   goalName,
					Amount:     amount,
					Percentage: 0, // Can be calculated if needed
				})
			}
		}

		// Build debt payments from FE request
		debtPayments := make([]domain.DebtPaymentItem, 0)
		for _, dp := range req.DebtPayments {
			// Use user adjusted payment if provided, otherwise use suggested
			amount := dp.SuggestedPayment
			if dp.UserAdjustedPayment != nil && *dp.UserAdjustedPayment > 0 {
				amount = *dp.UserAdjustedPayment
			}
			if amount > 0 {
				// Find debt name from cached state
				debtName := "Unknown"
				for _, d := range cachedState.InputDebts {
					if d.ID == dp.DebtID.String() {
						debtName = d.Name
						break
					}
				}
				debtPayments = append(debtPayments, domain.DebtPaymentItem{
					DebtID:    dp.DebtID,
					DebtName:  debtName,
					Amount:    amount,
					IsMinimum: amount <= dp.MinimumPayment,
				})
			}
		}

		workflow.BudgetAllocation = &domain.BudgetAllocationResult{
			Recommendations: req.BudgetAllocations, // Store exactly what FE sent (categories)
			GoalFundings:    goalFundings,          // Store exactly what FE sent (goals)
			DebtPayments:    debtPayments,          // Store exactly what FE sent (debts)
			Method:          "finalize",
		}
		now := time.Now()
		workflow.AppliedAtStep3 = &now
		workflow.MarkStepCompleted(3)
	}
	newState.DSSWorkflow = workflow

	// 8. Build CategoryStates from budget allocations (exactly as FE sent)
	// 8a. Categories (CONSTRAINT type)
	categoryNames := make(map[uuid.UUID]string)
	for _, c := range cachedState.InputConstraints {
		categoryNames[c.CategoryID] = c.Name
	}
	totalBudgeted := 0.0
	for catID, amount := range req.BudgetAllocations {
		if amount <= 0 {
			continue
		}
		totalBudgeted += amount
		name := categoryNames[catID]
		if name == "" {
			name = "Unknown"
		}
		// New month: Rollover = 0, Activity = 0, Available = Assigned
		newState.CategoryStates = append(newState.CategoryStates, domain.CategoryState{
			CategoryID: catID,
			Type:       domain.ItemTypeConstraint,
			Name:       name,
			Rollover:   0, // New month, no rollover
			Assigned:   amount,
			Activity:   0,      // No activity yet
			Available:  amount, // Available = Rollover(0) + Assigned + Activity(0)
		})
	}

	// 8b. Goals (GOAL type)
	goalMap := make(map[uuid.UUID]dto.InitGoalInput)
	for _, g := range cachedState.InputGoals {
		if goalID, err := uuid.Parse(g.ID); err == nil {
			goalMap[goalID] = g
		}
	}
	for _, gf := range req.GoalFundings {
		amount := gf.SuggestedAmount
		if gf.UserAdjustedAmount != nil && *gf.UserAdjustedAmount > 0 {
			amount = *gf.UserAdjustedAmount
		}
		if amount > 0 {
			goal, exists := goalMap[gf.GoalID]
			if !exists {
				continue
			}
			goalTarget := goal.TargetAmount
			newState.CategoryStates = append(newState.CategoryStates, domain.CategoryState{
				CategoryID: gf.GoalID,
				Type:       domain.ItemTypeGoal,
				Name:       goal.Name,
				Rollover:   0, // New month, no rollover
				Assigned:   amount,
				Activity:   0,      // No activity yet
				Available:  amount, // Available = Rollover(0) + Assigned + Activity(0)
				GoalTarget: &goalTarget,
			})
		}
	}

	// 8c. Debts (DEBT type)
	debtMap := make(map[uuid.UUID]dto.InitDebtInput)
	for _, d := range cachedState.InputDebts {
		if debtID, err := uuid.Parse(d.ID); err == nil {
			debtMap[debtID] = d
		}
	}
	for _, dp := range req.DebtPayments {
		amount := dp.SuggestedPayment
		if dp.UserAdjustedPayment != nil && *dp.UserAdjustedPayment > 0 {
			amount = *dp.UserAdjustedPayment
		}
		if amount > 0 {
			debt, exists := debtMap[dp.DebtID]
			if !exists {
				continue
			}
			minPayment := debt.MinimumPayment
			newState.CategoryStates = append(newState.CategoryStates, domain.CategoryState{
				CategoryID:     dp.DebtID,
				Type:           domain.ItemTypeDebt,
				Name:           debt.Name,
				Rollover:       0, // New month, no rollover
				Assigned:       amount,
				Activity:       0,      // No activity yet
				Available:      amount, // Available = Rollover(0) + Assigned + Activity(0)
				DebtMinPayment: &minPayment,
			})
		}
	}

	// Calculate ToBeBudgeted including goals and debts
	totalAllocatedForTBB := totalBudgeted + totalGoalFundings + totalDebtPayments
	newState.ToBeBudgeted = cachedState.MonthlyIncome - totalAllocatedForTBB
	if newState.ToBeBudgeted < 0 {
		newState.ToBeBudgeted = 0
	}

	// 9. Append new state with correct version
	newVersion := len(month.States) + 1
	newState.Version = newVersion
	month.States = append(month.States, newState)
	month.IncrementVersion()

	// 10. Save to DB
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return nil, fmt.Errorf("failed to save month with DSS workflow: %w", err)
	}

	s.logger.Info("Successfully finalized DSS workflow",
		zap.String("month_id", monthID.String()),
		zap.Int("new_state_version", newVersion),
		zap.Float64("tbb_remaining", newState.ToBeBudgeted),
	)

	// 11. CLEAR REDIS CACHE after successful DB save
	if err := s.dssCache.ClearState(ctx, monthID, *userID); err != nil {
		s.logger.Warn("Failed to clear DSS cache after finalizing", zap.Error(err))
	}

	// 12. Build response
	lastUpdated := workflow.LastUpdated
	statusResp := &dto.DSSWorkflowStatusResponse{
		MonthID:        monthID,
		CurrentStep:    workflow.CurrentStep,
		CompletedSteps: workflow.CompletedSteps,
		IsComplete:     workflow.IsComplete(),
		CanProceed:     false, // Workflow is complete after finalization
		LastUpdated:    &lastUpdated,
		Step1Applied:   workflow.AppliedAtStep1 != nil,
		Step2Applied:   workflow.AppliedAtStep2 != nil,
		Step3Applied:   false,                          // Step 3 tradeoff removed
		Step4Applied:   workflow.AppliedAtStep3 != nil, // Step 3 giờ là Budget Allocation
	}

	response := &dto.FinalizeDSSResponse{
		MonthID:      monthID,
		StateVersion: newVersion,
		ToBeBudgeted: newState.ToBeBudgeted,
		Status:       "dss_applied",
		DSSWorkflow:  statusResp,
		Message:      "DSS workflow finalized successfully. All allocations saved to database.",
	}

	return response, nil
}
