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

	// Debt tradeoff
	tradeoffDomain "personalfinancedss/internal/module/analytics/debt_tradeoff/domain"
	debtTradeoffDTO "personalfinancedss/internal/module/analytics/debt_tradeoff/dto"

	// Budget allocation
	budgetAllocationService "personalfinancedss/internal/module/analytics/budget_allocation/dto"
)

// ==================== DSS Cache invalidation rules ====================
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
		state.DebtStrategyPreview = nil
		state.AcceptedDebtStrategy = ""
	}
	if step <= 3 {
		state.TradeoffPreview = nil
		state.AcceptedGoalAllocationPct = 0
		state.AcceptedDebtAllocationPct = 0
	}
	if step <= 4 {
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
	serviceInput := &goalDto.AutoScoresRequest{
		UserID:        userID.String(),
		MonthlyIncome: cachedState.MonthlyIncome,
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

	// 5. Use custom criteria ratings if provided, otherwise default
	criteriaRatings := req.CriteriaRatings
	if criteriaRatings == nil || len(criteriaRatings) == 0 {
		criteriaRatings = map[string]int{
			"urgency":     5,
			"importance":  5,
			"feasibility": 5,
			"impact":      5,
		}
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
	debts := make([]debtStrategyDomain.DebtInfo, 0, len(cachedState.InputDebts))
	totalMinPayments := 0.0
	for _, d := range cachedState.InputDebts {
		debts = append(debts, debtStrategyDomain.DebtInfo{
			ID:             d.ID,
			Name:           d.Name,
			Balance:        d.CurrentBalance,
			InterestRate:   d.InterestRate / 100, // Convert to decimal
			MinimumPayment: d.MinimumPayment,
		})
		totalMinPayments += d.MinimumPayment
	}

	// 4. Calculate available debt budget (after fixed expenses)
	// Use 30% of disposable income as default debt budget
	totalConstraintAmount := 0.0
	for _, c := range cachedState.InputConstraints {
		if !c.IsFlexible {
			totalConstraintAmount += c.MinimumAmount
		}
	}
	disposable := cachedState.MonthlyIncome - totalConstraintAmount
	debtBudget := disposable * 0.3 // 30% for debt repayment
	if debtBudget < totalMinPayments {
		debtBudget = totalMinPayments // At least minimum payments
	}

	// 5. Build debt strategy input
	debtInput := &debtStrategyService.DebtStrategyInput{
		UserID:          userID.String(),
		TotalDebtBudget: debtBudget,
		Debts:           debts,
	}

	if req.PreferredStrategy != "" {
		debtInput.PreferredStrategy = debtStrategyDomain.Strategy(req.PreferredStrategy)
	}

	// 6. Call Debt Strategy Service
	result, err := s.debtStrategy.ExecuteDebtStrategy(ctx, debtInput)
	if err != nil {
		return nil, fmt.Errorf("debt strategy analysis failed: %w", err)
	}

	// 7. Update cached state
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

// ==================== Step 3: Goal-Debt Trade-off ====================

// PreviewGoalDebtTradeoff runs Monte Carlo trade-off analysis using cached data
func (s *monthService) PreviewGoalDebtTradeoff(ctx context.Context, req dto.PreviewGoalDebtTradeoffRequest, userID *uuid.UUID) (*dto.PreviewGoalDebtTradeoffResponse, error) {
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
	// Re-calc step3 => clear 3..4
	clearFromStep(cachedState, 3)

	// 3. Convert cached goals to tradeoff format
	goals := make([]tradeoffDomain.GoalInfo, 0, len(cachedState.InputGoals))
	for _, g := range cachedState.InputGoals {
		remaining := g.TargetAmount - g.CurrentAmount
		if remaining > 0 {
			deadline, _ := time.Parse("2006-01-02", g.TargetDate)
			// Map priority string to numeric
			priorityNum := 3.0 // default medium
			switch g.Priority {
			case "critical":
				priorityNum = 1.0
			case "high":
				priorityNum = 2.0
			case "medium":
				priorityNum = 3.0
			case "low":
				priorityNum = 4.0
			}
			goals = append(goals, tradeoffDomain.GoalInfo{
				ID:            g.ID,
				Name:          g.Name,
				TargetAmount:  g.TargetAmount,
				CurrentAmount: g.CurrentAmount,
				Priority:      priorityNum,
				Deadline:      deadline,
			})
		}
	}

	// 4. Convert cached debts to tradeoff format
	debts := make([]tradeoffDomain.DebtInfo, 0, len(cachedState.InputDebts))
	totalMinPayments := 0.0
	for _, d := range cachedState.InputDebts {
		debts = append(debts, tradeoffDomain.DebtInfo{
			ID:             d.ID,
			Name:           d.Name,
			Type:           "loan",
			Balance:        d.CurrentBalance,
			InterestRate:   d.InterestRate / 100,
			MinimumPayment: d.MinimumPayment,
		})
		totalMinPayments += d.MinimumPayment
	}

	// 5. Calculate essential expenses from constraints
	essentialExpenses := 0.0
	for _, c := range cachedState.InputConstraints {
		if !c.IsFlexible {
			essentialExpenses += c.MinimumAmount
		}
	}

	// 6. Build tradeoff input
	tradeoffInput := &debtTradeoffDTO.TradeoffInput{
		UserID:            userID.String(),
		MonthlyIncome:     cachedState.MonthlyIncome,
		EssentialExpenses: essentialExpenses,
		Debts:             debts,
		TotalMinPayments:  totalMinPayments,
		Goals:             goals,
		InvestmentProfile: tradeoffDomain.InvestmentProfile{
			RiskTolerance:  req.Preferences.RiskTolerance,
			ExpectedReturn: 0.08,
			TimeHorizon:    10,
		},
		EmergencyFund: tradeoffDomain.EmergencyFundStatus{
			CurrentAmount:   0,
			TargetAmount:    cachedState.MonthlyIncome * 6,
			MonthlyExpenses: essentialExpenses,
			TargetMonths:    6,
		},
		Preferences: req.Preferences,
	}

	// 7. Call Tradeoff Service
	result, err := s.debtTradeoff.ExecuteTradeoff(ctx, tradeoffInput)
	if err != nil {
		return nil, fmt.Errorf("tradeoff analysis failed: %w", err)
	}

	// 8. Update cached state
	cachedState.TradeoffPreview = result
	if err := s.dssCache.SaveState(ctx, cachedState); err != nil {
		s.logger.Warn("Failed to save tradeoff results", zap.Error(err))
	}

	return result, nil
}

// ApplyGoalDebtTradeoff caches the user's allocation decision
func (s *monthService) ApplyGoalDebtTradeoff(ctx context.Context, req dto.ApplyGoalDebtTradeoffRequest, userID *uuid.UUID) error {
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

	// 2. Validate allocations sum to 100%
	total := req.GoalAllocationPercent + req.DebtAllocationPercent
	if total < 99.9 || total > 100.1 {
		return fmt.Errorf("allocations must sum to 100%%, got %.2f%%", total)
	}

	// 3. Get cached state
	state, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return fmt.Errorf("failed to get DSS state: %w", err)
	}

	// 4. Update accepted allocation
	state.AcceptedGoalAllocationPct = req.GoalAllocationPercent
	state.AcceptedDebtAllocationPct = req.DebtAllocationPercent
	// Apply step3 => clear step4
	clearAfterStep(state, 3)

	// 5. Save back to Redis
	if err := s.dssCache.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save accepted tradeoff: %w", err)
	}

	s.logger.Info("Saved accepted goal-debt tradeoff",
		zap.String("month_id", req.MonthID.String()),
		zap.Float64("goal_percent", req.GoalAllocationPercent),
		zap.Float64("debt_percent", req.DebtAllocationPercent),
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
	// Re-calc step4 => clear step4
	clearFromStep(cachedState, 4)

	// 3. Nếu user chưa chọn tỉ lệ goal/debt (0-0) thì set default theo ngữ cảnh
	goalPct := req.GoalAllocationPct
	debtPct := req.DebtAllocationPct
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
		default:
			// Có cả 2 mà chưa chọn gì: chia đôi
			goalPct = 50
			debtPct = 50
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
	for i, c := range cachedState.InputConstraints {
		if !c.IsFlexible {
			allocationInput.MandatoryExpenses = append(allocationInput.MandatoryExpenses, budgetAllocationService.MandatoryExpense{
				CategoryID: c.CategoryID,
				Name:       c.Name,
				Amount:     c.MinimumAmount,
				Priority:   c.Priority,
			})
		} else {
			allocationInput.FlexibleExpenses = append(allocationInput.FlexibleExpenses, budgetAllocationService.FlexibleExpense{
				CategoryID: c.CategoryID,
				Name:       c.Name,
				MinAmount:  c.MinimumAmount,
				MaxAmount:  c.MaximumAmount,
				Priority:   i + 1, // Use index as priority if not set
			})
		}
	}

	// 6. Map goals from cache với allocation percentage
	goalAllocationPct := goalPct / 100.0
	for _, g := range cachedState.InputGoals {
		remaining := g.TargetAmount - g.CurrentAmount
		if remaining > 0 {
			suggestedContribution := remaining * 0.1 * goalAllocationPct // 10% of remaining adjusted by allocation

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

	// 7. Map debts from cache với allocation percentage
	debtAllocationPct := debtPct / 100.0
	for _, d := range cachedState.InputDebts {
		adjustedPayment := d.MinimumPayment * (1 + debtAllocationPct) // Min + extra based on allocation

		allocationInput.Debts = append(allocationInput.Debts, budgetAllocationService.DebtInput{
			DebtID:         uuid.MustParse(d.ID),
			Name:           d.Name,
			Balance:        d.CurrentBalance,
			InterestRate:   d.InterestRate / 100,
			MinimumPayment: adjustedPayment,
		})
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

// ApplyBudgetAllocation caches the user's selected scenario allocations
func (s *monthService) ApplyBudgetAllocation(ctx context.Context, req dto.ApplyBudgetAllocationRequest, userID *uuid.UUID) error {
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

	// 2. Validate allocations provided
	if len(req.Allocations) == 0 {
		return errors.New("no allocations provided")
	}

	// 3. Get cached state
	state, err := s.dssCache.GetOrInitState(ctx, req.MonthID, *userID)
	if err != nil {
		return fmt.Errorf("failed to get DSS state: %w", err)
	}

	// 4. Update accepted allocations
	state.AcceptedScenario = req.SelectedScenario
	state.AcceptedAllocations = req.Allocations

	// 5. Save back to Redis
	if err := s.dssCache.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save accepted allocations: %w", err)
	}

	s.logger.Info("Saved accepted budget allocations",
		zap.String("month_id", req.MonthID.String()),
		zap.String("scenario", req.SelectedScenario),
		zap.Int("categories", len(req.Allocations)),
	)

	return nil
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
		response.Step3Applied = state.DSSWorkflow.AppliedAtStep3 != nil
		response.Step4Applied = state.DSSWorkflow.AppliedAtStep4 != nil
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

// FinalizeDSS applies all cached DSS results to MonthState and saves to DB
func (s *monthService) FinalizeDSS(ctx context.Context, req dto.FinalizeDSSRequest, monthID uuid.UUID, userID *uuid.UUID) (*dto.FinalizeDSSResponse, error) {
	s.logger.Info("Finalizing DSS workflow",
		zap.String("month_id", monthID.String()),
	)

	// 1. GET CACHED STATE FROM REDIS
	cachedState, err := s.dssCache.GetState(ctx, monthID, *userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached DSS state: %w", err)
	}
	if cachedState == nil {
		return nil, errors.New("no cached DSS state found - run Initialize and preview/apply steps first")
	}

	s.logger.Info("Retrieved cached DSS state",
		zap.Bool("has_goal_ranking", len(cachedState.AcceptedGoalRanking) > 0),
		zap.Bool("has_debt_strategy", cachedState.AcceptedDebtStrategy != ""),
		zap.Bool("has_tradeoff", cachedState.AcceptedGoalAllocationPct > 0),
		zap.Bool("has_allocations", len(cachedState.AcceptedAllocations) > 0),
	)

	// 2. Load Month & Validate
	month, err := s.repo.GetMonthByID(ctx, monthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}
	if !month.CanBeModified() {
		return nil, errors.New("cannot finalize DSS on closed month")
	}

	// 3. Get Current MonthState
	state := month.CurrentState()
	if state == nil {
		state = domain.NewMonthState()
		month.States = []domain.MonthState{*state}
	}

	// 4. BUILD DSSWorkflowResults FROM CACHED STATE
	workflow := &domain.DSSWorkflowResults{
		StartedAt:      time.Now(),
		LastUpdated:    time.Now(),
		CurrentStep:    5,
		CompletedSteps: []int{},
	}

	// Step 1: Goal Prioritization
	if len(cachedState.AcceptedGoalRanking) > 0 {
		rankings := make([]domain.GoalRanking, len(cachedState.AcceptedGoalRanking))
		for i, goalID := range cachedState.AcceptedGoalRanking {
			rankings[i] = domain.GoalRanking{
				GoalID: goalID,
				Rank:   i + 1,
				Score:  float64(len(cachedState.AcceptedGoalRanking) - i),
			}
		}
		workflow.GoalPrioritization = &domain.GoalPriorityResult{
			Rankings: rankings,
			Method:   "user_approved",
		}
		workflow.MarkStepCompleted(1)
	}

	// Step 2: Debt Strategy
	if cachedState.AcceptedDebtStrategy != "" {
		workflow.DebtStrategy = &domain.DebtStrategyResult{
			Strategy: cachedState.AcceptedDebtStrategy,
		}
		workflow.MarkStepCompleted(2)
	}

	// Step 3: Goal-Debt Tradeoff
	if cachedState.AcceptedGoalAllocationPct > 0 || cachedState.AcceptedDebtAllocationPct > 0 {
		workflow.GoalDebtTradeoff = &domain.GoalDebtTradeoffResult{
			GoalAllocationPercent: cachedState.AcceptedGoalAllocationPct,
			DebtAllocationPercent: cachedState.AcceptedDebtAllocationPct,
		}
		workflow.MarkStepCompleted(3)
	}

	// Step 4: Budget Allocation
	if len(cachedState.AcceptedAllocations) > 0 {
		workflow.BudgetAllocation = &domain.BudgetAllocationResult{
			Recommendations: cachedState.AcceptedAllocations,
			Method:          cachedState.AcceptedScenario,
		}
		workflow.MarkStepCompleted(4)
	}

	state.DSSWorkflow = workflow

	// 5. Update CategoryStates from cached allocations
	if len(cachedState.AcceptedAllocations) > 0 {
		// Build category name lookup
		categoryNames := make(map[uuid.UUID]string)
		for _, c := range cachedState.InputConstraints {
			categoryNames[c.CategoryID] = c.Name
		}

		for catID, amount := range cachedState.AcceptedAllocations {
			// Find existing category state
			var catState *domain.CategoryState
			for i := range state.CategoryStates {
				if state.CategoryStates[i].CategoryID == catID {
					catState = &state.CategoryStates[i]
					break
				}
			}

			if catState != nil {
				catState.Assigned = amount
				catState.Available = catState.Rollover + amount + catState.Activity
			} else {
				name := categoryNames[catID]
				if name == "" {
					name = "Unknown"
				}
				state.CategoryStates = append(state.CategoryStates, domain.CategoryState{
					CategoryID: catID,
					Name:       name,
					Assigned:   amount,
					Activity:   0,
					Available:  amount,
				})
			}
		}
	}

	// 6. Calculate remaining TBB
	totalBudgeted := 0.0
	for _, amount := range cachedState.AcceptedAllocations {
		totalBudgeted += amount
	}
	state.ToBeBudgeted = cachedState.MonthlyIncome - totalBudgeted
	state.ActualIncome = cachedState.MonthlyIncome

	// 7. Mark state as applied
	state.IsApplied = true

	// 8. Create NEW MonthState version
	newVersion := len(month.States) + 1
	newState := *state
	newState.Version = newVersion
	newState.CreatedAt = time.Now()

	month.States = append(month.States, newState)
	month.IncrementVersion()

	// 9. Save to DB
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return nil, fmt.Errorf("failed to save month with finalized DSS: %w", err)
	}

	s.logger.Info("Successfully finalized DSS workflow",
		zap.String("month_id", monthID.String()),
		zap.Int("new_state_version", newVersion),
		zap.Float64("tbb_remaining", state.ToBeBudgeted),
	)

	// 10. CLEAR REDIS CACHE after successful DB save
	if err := s.dssCache.ClearState(ctx, monthID, *userID); err != nil {
		s.logger.Warn("Failed to clear DSS cache after finalization", zap.Error(err))
	}

	// 11. Build Response
	response := &dto.FinalizeDSSResponse{
		MonthID:      monthID,
		StateVersion: newVersion,
		ToBeBudgeted: state.ToBeBudgeted,
		Status:       "dss_applied",
		DSSWorkflow: &dto.DSSWorkflowStatusResponse{
			MonthID:        monthID,
			CurrentStep:    workflow.CurrentStep,
			CompletedSteps: workflow.CompletedSteps,
			IsComplete:     workflow.IsComplete(),
			CanProceed:     false,
			LastUpdated:    &workflow.LastUpdated,
			Step1Applied:   workflow.GoalPrioritization != nil,
			Step2Applied:   workflow.DebtStrategy != nil,
			Step3Applied:   workflow.GoalDebtTradeoff != nil,
			Step4Applied:   workflow.BudgetAllocation != nil,
		},
		Message: fmt.Sprintf("DSS workflow finalized successfully. Created state version %d.", newVersion),
	}

	return response, nil
}
