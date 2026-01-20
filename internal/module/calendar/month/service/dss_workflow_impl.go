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

// ==================== Step 0: Auto-Scoring Preview ====================

// PreviewAutoScoring runs goal auto-scoring and returns scored goals
// This is a PREVIEW-ONLY step - it does not modify DSSWorkflow state
func (s *monthService) PreviewAutoScoring(ctx context.Context, req dto.PreviewAutoScoringRequest, userID *uuid.UUID) (*dto.PreviewAutoScoringResponse, error) {
	// 1. Load Month & Validate Ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. Map Frontend Goals to Service Input
	// Note: req.Goals is ALREADY []goalDto.GoalForRating from frontend - use directly
	serviceInput := &goalDto.AutoScoresRequest{
		UserID:        userID.String(),
		MonthlyIncome: req.MonthlyIncome, // From frontend request
		Goals:         req.Goals,
	}

	// 3. Call Goal Service
	scores, err := s.goalPrioritization.GetAutoScores(ctx, serviceInput)
	if err != nil {
		return nil, fmt.Errorf("auto-scoring failed: %w", err)
	}

	// 4. Cache preview result in Redis
	if err := s.dssCache.SetPreview(ctx, req.MonthID, *userID, "auto_scoring", scores); err != nil {
		s.logger.Warn("Failed to cache auto-scoring preview", zap.Error(err))
		// Don't fail - caching is optional
	}

	// 5. Return analytics response directly (PreviewAutoScoringResponse is type alias)
	return scores, nil
}

// ==================== Step 1: Goal Prioritization ====================

// PreviewGoalPrioritization runs AHP goal prioritization and returns preview
func (s *monthService) PreviewGoalPrioritization(ctx context.Context, req dto.PreviewGoalPrioritizationRequest, userID *uuid.UUID) (*dto.PreviewGoalPrioritizationResponse, error) {
	// 1. Load Month & Validate Ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. Prepare AHP Input - req.Goals is already []goalDto.GoalForRating
	// Use custom criteria ratings if provided, otherwise default to 5 (equal importance)
	criteriaRatings := req.CriteriaRatings
	if criteriaRatings == nil || len(criteriaRatings) == 0 {
		criteriaRatings = map[string]int{
			"urgency":     5,
			"importance":  5,
			"feasibility": 5,
			"impact":      5,
		}
	}

	ahpInput := &goalDto.DirectRatingWithOverridesInput{
		DirectRatingInput: goalDto.DirectRatingInput{
			UserID:          userID.String(),
			MonthlyIncome:   0,
			Goals:           req.Goals,
			CriteriaRatings: criteriaRatings,
		},
		GoalScoreOverrides: make(map[string]map[string]float64),
	}

	// Populate manual score overrides if provided
	for goalID, manual := range req.ManualScores {
		// Manual scores are 0-10, normalize to 0.0-1.0
		ahpInput.GoalScoreOverrides[goalID] = map[string]float64{
			"urgency":     manual.Urgency / 10.0,
			"importance":  manual.Importance / 10.0,
			"feasibility": manual.Effort / 10.0, // Mapping Effort -> Feasibility
			"impact":      manual.ROI / 10.0,    // Mapping ROI -> Impact
		}
	}

	// 3. Call AHP Service
	result, err := s.goalPrioritization.ExecuteDirectRatingWithOverrides(ctx, ahpInput)
	if err != nil {
		return nil, fmt.Errorf("AHP execution failed: %w", err)
	}

	// 4. Prepare response
	response := &dto.PreviewGoalPrioritizationResponse{
		AHPOutput:  result,
		AutoScores: nil, // Auto scores from step 0 if needed
	}

	// 5. Cache preview result in Redis
	if err := s.dssCache.SetPreview(ctx, req.MonthID, *userID, "goal_prioritization", response); err != nil {
		s.logger.Warn("Failed to cache goal prioritization preview", zap.Error(err))
		// Don't fail - caching is optional
	}

	// 6. Return response
	return response, nil
}

// ApplyGoalPrioritization saves the user's accepted goal ranking to MonthState
func (s *monthService) ApplyGoalPrioritization(ctx context.Context, req dto.ApplyGoalPrioritizationRequest, userID *uuid.UUID) error {
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
		return errors.New("cannot modify closed month")
	}

	// 4. Get current state
	state := month.CurrentState()

	// 5. Initialize DSSWorkflow if nil
	if state.DSSWorkflow == nil {
		state.DSSWorkflow = &domain.DSSWorkflowResults{
			StartedAt:      time.Now(),
			LastUpdated:    time.Now(),
			CurrentStep:    0,
			CompletedSteps: []int{},
		}
	}

	// 6. Validate workflow step (must be step 0 or 1)
	if state.DSSWorkflow.CurrentStep > 1 {
		return errors.New("cannot apply goal prioritization after completing step 1")
	}

	// 7. Save accepted ranking
	now := time.Now()
	rankings := make([]domain.GoalRanking, len(req.AcceptedRanking))
	for i, goalID := range req.AcceptedRanking {
		rankings[i] = domain.GoalRanking{
			GoalID: goalID,
			Rank:   i + 1,
			Score:  float64(len(req.AcceptedRanking) - i), // Higher rank = higher score
		}
	}

	state.DSSWorkflow.GoalPrioritization = &domain.GoalPriorityResult{
		Method:   "user_approved",
		Rankings: rankings,
	}
	state.DSSWorkflow.AppliedAtStep1 = &now
	state.DSSWorkflow.MarkStepCompleted(1)

	// 8. Save to DB
	month.IncrementVersion()
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to save month: %w", err)
	}

	s.logger.Info("Applied goal prioritization",
		zap.String("month_id", req.MonthID.String()),
		zap.Int("goal_count", len(req.AcceptedRanking)),
	)

	return nil
}

// ==================== Step 2: Debt Strategy ====================

// PreviewDebtStrategy runs debt repayment simulations and returns scenarios
func (s *monthService) PreviewDebtStrategy(ctx context.Context, req dto.PreviewDebtStrategyRequest, userID *uuid.UUID) (*dto.PreviewDebtStrategyResponse, error) {
	// 1. Load Month & Validate Ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. Prepare Debt Strategy Input
	// req.Debts is already []domain.DebtInfo from analytics module
	debtInput := &debtStrategyService.DebtStrategyInput{
		UserID:          userID.String(),
		TotalDebtBudget: req.TotalDebtBudget,
		Debts:           req.Debts, // Already correct type
	}

	// Handle preferred strategy if provided
	if req.PreferredStrategy != "" {
		debtInput.PreferredStrategy = debtStrategyDomain.Strategy(req.PreferredStrategy)
	}

	// 3. Call Debt Strategy Service
	result, err := s.debtStrategy.ExecuteDebtStrategy(ctx, debtInput)
	if err != nil {
		return nil, fmt.Errorf("debt strategy analysis failed: %w", err)
	}

	// 4. Cache preview result in Redis
	if err := s.dssCache.SetPreview(ctx, req.MonthID, *userID, "debt_strategy", result); err != nil {
		s.logger.Warn("Failed to cache debt strategy preview", zap.Error(err))
		// Don't fail - caching is optional
	}

	// 5. Return analytics response directly (PreviewDebtStrategyResponse is type alias)
	return result, nil
}

// ApplyDebtStrategy saves the user's selected debt strategy to MonthState
func (s *monthService) ApplyDebtStrategy(ctx context.Context, req dto.ApplyDebtStrategyRequest, userID *uuid.UUID) error {
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
		return errors.New("cannot modify closed month")
	}

	// 4. Get current state and validate workflow
	state := month.CurrentState()
	if state.DSSWorkflow == nil || !state.DSSWorkflow.CanProceedToStep(2) {
		return errors.New("must complete step 1 before applying debt strategy")
	}

	// 5. Save selected strategy
	now := time.Now()
	state.DSSWorkflow.DebtStrategy = &domain.DebtStrategyResult{
		Strategy:    req.SelectedStrategy,
		PaymentPlan: []domain.DebtPaymentPlan{},
	}
	state.DSSWorkflow.AppliedAtStep2 = &now
	state.DSSWorkflow.MarkStepCompleted(2)

	// 6. Save to DB
	month.IncrementVersion()
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to save month: %w", err)
	}

	s.logger.Info("Applied debt strategy",
		zap.String("month_id", req.MonthID.String()),
		zap.String("strategy", req.SelectedStrategy),
	)

	return nil
}

// ==================== Step 3: Goal-Debt Trade-off ====================

// PreviewGoalDebtTradeoff runs Monte Carlo trade-off analysis
func (s *monthService) PreviewGoalDebtTradeoff(ctx context.Context, req dto.PreviewGoalDebtTradeoffRequest, userID *uuid.UUID) (*dto.PreviewGoalDebtTradeoffResponse, error) {
	// 1. Load Month & Validate Ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. Validate Step 1 & 2 have been completed
	state := month.CurrentState()
	if state.DSSWorkflow == nil || !state.DSSWorkflow.CanProceedToStep(3) {
		return nil, errors.New("must complete steps 1-2 before previewing trade-off")
	}

	// 3. Extract Goals from Step 1 (GoalPrioritization)
	var goals []tradeoffDomain.GoalInfo
	if state.DSSWorkflow.GoalPrioritization != nil {
		for _, ranking := range state.DSSWorkflow.GoalPrioritization.Rankings {
			// Get actual goal data from original input snapshot
			for _, goalSnap := range state.Goals {
				if goalSnap.GoalID == ranking.GoalID {
					remaining := goalSnap.TargetAmount - goalSnap.CurrentAmount
					if remaining > 0 {
						// Parse deadline
						var deadline time.Time
						if goalSnap.TargetDate != nil {
							if parsed, err := time.Parse("2006-01-02", *goalSnap.TargetDate); err == nil {
								deadline = parsed
							}
						}

						goals = append(goals, tradeoffDomain.GoalInfo{
							ID:            ranking.GoalID.String(),
							Name:          ranking.GoalName,
							TargetAmount:  goalSnap.TargetAmount,
							CurrentAmount: goalSnap.CurrentAmount,
							Priority:      float64(ranking.Rank),
							Deadline:      deadline,
						})
					}
					break
				}
			}
		}
	}

	// 4. Extract Debts from Step 2 (DebtStrategy)
	var debts []tradeoffDomain.DebtInfo
	var totalMinPayments float64
	if state.DSSWorkflow.DebtStrategy != nil {
		for _, debtPlan := range state.DSSWorkflow.DebtStrategy.PaymentPlan {
			// Get actual debt data from original input snapshot
			for _, debtSnap := range state.Debts {
				if debtSnap.DebtID == debtPlan.DebtID {
					debts = append(debts, tradeoffDomain.DebtInfo{
						ID:             debtPlan.DebtID.String(),
						Name:           debtPlan.DebtName,
						Type:           "loan", // Default type
						Balance:        debtSnap.CurrentBalance,
						InterestRate:   debtSnap.InterestRate / 100, // Convert to decimal
						MinimumPayment: debtPlan.MinPayment,
					})
					totalMinPayments += debtPlan.MinPayment
					break
				}
			}
		}
	}

	// 5. Calculate essential expenses from constraints (legacy field support)
	essentialExpenses := 0.0
	// Note: state.Constraints is deprecated, should use state.Input.Constraints in future
	// For now, we'll calculate from CategoryStates if available
	for _, catState := range state.CategoryStates {
		// Sum assigned amounts as proxy for essential expenses
		essentialExpenses += catState.Assigned
	}

	// 6. Get monthly income from state
	monthlyIncome := state.Income
	if monthlyIncome == 0 {
		monthlyIncome = state.ActualIncome
	}

	// 7. Build TradeoffInput - wrap data from Step 1 & 2 + preferences from request
	tradeoffInput := &debtTradeoffDTO.TradeoffInput{
		UserID:            userID.String(),
		MonthlyIncome:     monthlyIncome,
		EssentialExpenses: essentialExpenses,
		Debts:             debts,
		TotalMinPayments:  totalMinPayments,
		Goals:             goals,
		InvestmentProfile: tradeoffDomain.InvestmentProfile{
			RiskTolerance:      req.Preferences.RiskTolerance,
			ExpectedReturn:     0.08, // TODO: Get from user profile or FE
			TimeHorizon:        10,   // TODO: Get from user profile or FE
			CurrentInvestments: 0,    // TODO: Get from investment module
		},
		EmergencyFund: tradeoffDomain.EmergencyFundStatus{
			CurrentAmount:   0,                 // TODO: Calculate from accounts
			TargetAmount:    monthlyIncome * 6, // TODO: Get from user profile
			MonthlyExpenses: essentialExpenses,
			TargetMonths:    6, // TODO: Get from user profile
		},
		Preferences: req.Preferences, // Use preferences from FE request
	}

	// 8. Call Debt Tradeoff Service
	result, err := s.debtTradeoff.ExecuteTradeoff(ctx, tradeoffInput)
	if err != nil {
		return nil, fmt.Errorf("tradeoff analysis failed: %w", err)
	}

	// 9. Cache preview result in Redis
	if err := s.dssCache.SetPreview(ctx, req.MonthID, *userID, "tradeoff", result); err != nil {
		s.logger.Warn("Failed to cache tradeoff preview", zap.Error(err))
		// Don't fail - caching is optional
	}

	// 10. Return analytics response directly
	return result, nil
}

// ApplyGoalDebtTradeoff saves the user's allocation decision to MonthState
func (s *monthService) ApplyGoalDebtTradeoff(ctx context.Context, req dto.ApplyGoalDebtTradeoffRequest, userID *uuid.UUID) error {
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
		return errors.New("cannot modify closed month")
	}

	// 4. Validate workflow
	state := month.CurrentState()
	if state.DSSWorkflow == nil || !state.DSSWorkflow.CanProceedToStep(3) {
		return errors.New("must complete steps 1-2 before applying trade-off")
	}

	// 5. Validate allocations sum to 100%
	total := req.GoalAllocationPercent + req.DebtAllocationPercent
	if total < 99.9 || total > 100.1 { // Allow small floating point errors
		return fmt.Errorf("allocations must sum to 100%%, got %.2f%%", total)
	}

	// 6. Save trade-off decision
	now := time.Now()
	state.DSSWorkflow.GoalDebtTradeoff = &domain.GoalDebtTradeoffResult{
		GoalAllocationPercent: req.GoalAllocationPercent,
		DebtAllocationPercent: req.DebtAllocationPercent,
		// Analysis will be populated in future enhancement
	}
	state.DSSWorkflow.AppliedAtStep3 = &now
	state.DSSWorkflow.MarkStepCompleted(3)

	// 7. Save to DB
	month.IncrementVersion()
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to save month: %w", err)
	}

	s.logger.Info("Applied goal-debt tradeoff",
		zap.String("month_id", req.MonthID.String()),
		zap.Float64("goal_percent", req.GoalAllocationPercent),
		zap.Float64("debt_percent", req.DebtAllocationPercent),
	)

	return nil
}

// ==================== Step 4: Budget Allocation ====================

// PreviewBudgetAllocation runs Goal Programming allocation with results from steps 1-3
func (s *monthService) PreviewBudgetAllocation(ctx context.Context, req dto.PreviewBudgetAllocationRequest, userID *uuid.UUID) (*dto.PreviewBudgetAllocationResponse, error) {
	// 1. Load Month & Validate Ownership
	month, err := s.repo.GetMonthByID(ctx, req.MonthID)
	if err != nil {
		return nil, fmt.Errorf("failed to load month: %w", err)
	}
	if userID != nil && month.UserID != *userID {
		return nil, errors.New("unauthorized: month does not belong to user")
	}

	// 2. Skip validation for now - Step 4 can run independently
	// Steps 1-3 are optional (skipped if no goals/debts)
	state := month.CurrentState()
	if state == nil {
		return nil, errors.New("month state not initialized")
	}

	// 3. Build Allocation Input - integrate all previous steps
	allocationInput := &budgetAllocationService.BudgetAllocationModelInput{
		UserID:            *userID,
		Year:              time.Now().Year(),
		Month:             int(time.Now().Month()),
		TotalIncome:       req.TotalIncome,
		UseAllScenarios:   true,
		MandatoryExpenses: make([]budgetAllocationService.MandatoryExpense, 0),
		FlexibleExpenses:  make([]budgetAllocationService.FlexibleExpense, 0),
		Debts:             make([]budgetAllocationService.DebtInput, 0),
		Goals:             make([]budgetAllocationService.GoalInput, 0),
	}

	// 4. Map constraints from request (from FE)
	// Build category name lookup from state.Constraints
	categoryNames := make(map[uuid.UUID]string)
	for _, constraint := range state.Constraints {
		categoryNames[constraint.CategoryID] = constraint.CategoryName
	}

	for _, c := range req.Constraints {
		// Get category name from state.Constraints or use from request
		categoryName := c.CategoryName
		if categoryName == "" || categoryName == "Unknown Category" {
			if name, exists := categoryNames[c.CategoryID]; exists && name != "" {
				categoryName = name
			} else {
				categoryName = "Unknown Category" // Fallback
			}
		}

		if c.Flexibility == "fixed" {
			allocationInput.MandatoryExpenses = append(allocationInput.MandatoryExpenses, budgetAllocationService.MandatoryExpense{
				CategoryID: c.CategoryID,
				Name:       categoryName,
				Amount:     c.MinAmount,
				Priority:   c.Priority,
			})
		} else {
			allocationInput.FlexibleExpenses = append(allocationInput.FlexibleExpenses, budgetAllocationService.FlexibleExpense{
				CategoryID: c.CategoryID,
				Name:       categoryName,
				MinAmount:  c.MinAmount,
				MaxAmount:  c.MaxAmount,
				Priority:   c.Priority,
			})
		}
	}

	// 5. Integrate Goals from Step 1 (with allocation % from Step 3)
	goalAllocationPct := req.GoalAllocationPct / 100.0 // Convert to decimal
	if state.DSSWorkflow != nil && state.DSSWorkflow.GoalPrioritization != nil {
		// Use prioritized goals from Step 1
		for _, ranking := range state.DSSWorkflow.GoalPrioritization.Rankings {
			for _, goalSnap := range state.Goals {
				if goalSnap.GoalID == ranking.GoalID {
					remaining := goalSnap.TargetAmount - goalSnap.CurrentAmount
					if remaining > 0 {
						adjustedAmount := ranking.SuggestedAmount * goalAllocationPct
						priorityStr := "medium"
						switch {
						case ranking.Rank == 1:
							priorityStr = "critical"
						case ranking.Rank <= 3:
							priorityStr = "high"
						case ranking.Rank <= 6:
							priorityStr = "medium"
						default:
							priorityStr = "low"
						}
						allocationInput.Goals = append(allocationInput.Goals, budgetAllocationService.GoalInput{
							GoalID:                ranking.GoalID,
							Name:                  ranking.GoalName,
							Type:                  "savings",
							Priority:              priorityStr,
							RemainingAmount:       remaining,
							SuggestedContribution: adjustedAmount,
						})
					}
					break
				}
			}
		}
	} else if len(state.Goals) > 0 {
		// Fallback: Use goals directly from state (if Steps 1-3 were skipped)
		for _, goalSnap := range state.Goals {
			remaining := goalSnap.TargetAmount - goalSnap.CurrentAmount
			if remaining > 0 {
				// Use simple heuristic for contribution: remaining / months until deadline
				// Or a fixed percentage of remaining (e.g., 10%)
				suggestedContribution := remaining * 0.1 // 10% of remaining per month

				// Map int priority to string priority
				priorityStr := "medium"
				switch goalSnap.Priority {
				case 1:
					priorityStr = "critical"
				case 2:
					priorityStr = "high"
				case 3:
					priorityStr = "medium"
				case 4:
					priorityStr = "low"
				}

				allocationInput.Goals = append(allocationInput.Goals, budgetAllocationService.GoalInput{
					GoalID:                goalSnap.GoalID,
					Name:                  goalSnap.Name,
					Type:                  "savings",
					Priority:              priorityStr,
					RemainingAmount:       remaining,
					SuggestedContribution: suggestedContribution,
				})
			}
		}
	}

	// 6. Integrate Debts from Step 2 (with allocation % from Step 3)
	debtAllocationPct := req.DebtAllocationPct / 100.0 // Convert to decimal
	if state.DSSWorkflow != nil && state.DSSWorkflow.DebtStrategy != nil {
		for _, debtPlan := range state.DSSWorkflow.DebtStrategy.PaymentPlan {
			// Find debt details from input snapshot
			for _, debtSnap := range state.Debts {
				if debtSnap.DebtID == debtPlan.DebtID {
					// Apply allocation percentage from Step 3
					// Note: MinimumPayment will be used by budget allocation model
					// The model will allocate based on strategy from Step 2
					allocationInput.Debts = append(allocationInput.Debts, budgetAllocationService.DebtInput{
						DebtID:         debtPlan.DebtID,
						Name:           debtPlan.DebtName,
						Balance:        debtSnap.CurrentBalance,
						InterestRate:   debtSnap.InterestRate / 100,               // Convert to decimal
						MinimumPayment: debtPlan.TotalPayment * debtAllocationPct, // Adjusted payment
					})
					break
				}
			}
		}
	}

	// 7. Call Budget Allocation Service
	result, err := s.budgetAllocation.ExecuteBudgetAllocation(ctx, allocationInput)
	if err != nil {
		return nil, fmt.Errorf("budget allocation failed: %w", err)
	}

	// 8. Return analytics response directly (PreviewBudgetAllocationResponse is type alias)
	return result, nil
}

// ApplyBudgetAllocation applies the selected allocation scenario to MonthState categories
// Note: Frontend should pass allocations from the selected scenario
func (s *monthService) ApplyBudgetAllocation(ctx context.Context, req dto.ApplyBudgetAllocationRequest, userID *uuid.UUID) error {
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
		return errors.New("cannot modify closed month")
	}

	// 4. Validate workflow
	state := month.CurrentState()
	if state.DSSWorkflow == nil || !state.DSSWorkflow.CanProceedToStep(4) {
		return errors.New("must complete steps 1-3 before applying budget allocation")
	}

	// 5. Validate allocations provided
	if len(req.Allocations) == 0 {
		return errors.New("no allocations provided - frontend must send selected scenario allocations")
	}

	// 6. Update CategoryStates with allocations
	// Initialize CategoryStates if nil
	if state.CategoryStates == nil {
		state.CategoryStates = make(map[uuid.UUID]*domain.CategoryState)
	}

	for categoryID, amount := range req.Allocations {
		// Get existing category state or create new
		catState, exists := state.CategoryStates[categoryID]
		if !exists {
			catState = &domain.CategoryState{
				Rollover:  0,
				Assigned:  0,
				Activity:  0,
				Available: 0,
			}
			state.CategoryStates[categoryID] = catState
		}

		// Update Assigned from DSS allocation
		catState.Assigned = amount
		// Recalculate Available = Rollover + Assigned + Activity
		catState.Available = catState.Rollover + catState.Assigned + catState.Activity
	}

	// 7. Update TBB (subtract allocated amount)
	totalAllocated := 0.0
	for _, amount := range req.Allocations {
		totalAllocated += amount
	}
	state.ToBeBudgeted -= totalAllocated

	// 8. Save budget allocation result to DSSWorkflow
	now := time.Now()
	state.DSSWorkflow.BudgetAllocation = &domain.BudgetAllocationResult{
		Method:          req.SelectedScenario,
		Recommendations: req.Allocations, // Save allocations map
		OptimalityScore: 100.0,           // TODO: Get from analytics result
	}
	state.DSSWorkflow.AppliedAtStep4 = &now
	state.DSSWorkflow.MarkStepCompleted(4)

	// 9. Mark state as applied (DSS workflow complete)
	state.IsApplied = true

	// 10. Save to DB
	month.IncrementVersion()
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to save month: %w", err)
	}

	s.logger.Info("Applied budget allocation - DSS workflow complete",
		zap.String("month_id", req.MonthID.String()),
		zap.String("scenario", req.SelectedScenario),
		zap.Int("categories_updated", len(req.Allocations)),
		zap.Float64("total_allocated", totalAllocated),
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
		CanProceed:     true, // Can always start step 1
		IsComplete:     false,
	}

	if state.DSSWorkflow != nil {
		response.CurrentStep = state.DSSWorkflow.CurrentStep
		response.CompletedSteps = state.DSSWorkflow.CompletedSteps
		response.IsComplete = state.DSSWorkflow.IsComplete()
		response.CanProceed = !response.IsComplete
		response.LastUpdated = &state.DSSWorkflow.LastUpdated

		// Step status
		response.Step1Applied = state.DSSWorkflow.AppliedAtStep1 != nil
		response.Step2Applied = state.DSSWorkflow.AppliedAtStep2 != nil
		response.Step3Applied = state.DSSWorkflow.AppliedAtStep3 != nil
		response.Step4Applied = state.DSSWorkflow.AppliedAtStep4 != nil
	}

	return response, nil
}

// ResetDSSWorkflow clears all DSS workflow results and starts over
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

	// 4. Reset workflow
	state := month.CurrentState()
	if state.DSSWorkflow != nil {
		state.DSSWorkflow.Reset()
	} else {
		// Initialize empty workflow
		state.DSSWorkflow = &domain.DSSWorkflowResults{
			StartedAt:      time.Now(),
			LastUpdated:    time.Now(),
			CurrentStep:    0,
			CompletedSteps: []int{},
		}
	}

	// 5. Mark state as not applied
	state.IsApplied = false

	// 6. Save to DB
	month.IncrementVersion()
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return fmt.Errorf("failed to save month: %w", err)
	}

	s.logger.Info("Reset DSS workflow",
		zap.String("month_id", req.MonthID.String()),
	)

	return nil
}

// ==================== DSS Finalization (Apply All at Once) ====================

// FinalizeDSS applies all DSS workflow results at once and creates a new MonthState version
// This implements Approach 2: Preview all → Review → Finalize once
func (s *monthService) FinalizeDSS(ctx context.Context, req dto.FinalizeDSSRequest, monthID uuid.UUID, userID *uuid.UUID) (*dto.FinalizeDSSResponse, error) {
	s.logger.Info("Finalizing DSS workflow",
		zap.String("month_id", monthID.String()),
		zap.Int("goal_count", len(req.GoalPriorities)),
		zap.Bool("has_debt_strategy", req.DebtStrategy != nil),
		zap.Bool("has_tradeoff", req.TradeoffChoice != nil),
	)

	// 1. Load Month & Validate
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

	// 2. Get or Initialize Current State
	state := month.CurrentState()
	if state == nil {
		// Create first state if doesn't exist
		state = &domain.MonthState{
			Version:        1,
			ToBeBudgeted:   0,
			ActualIncome:   0,
			CategoryStates: make(map[uuid.UUID]*domain.CategoryState),
			CreatedAt:      time.Now(),
			IsApplied:      false,
		}
		month.States = []domain.MonthState{*state}
	}

	// 3. Initialize or Get DSSWorkflow
	if state.DSSWorkflow == nil {
		state.DSSWorkflow = &domain.DSSWorkflowResults{
			StartedAt:      time.Now(),
			CurrentStep:    0,
			CompletedSteps: []int{},
		}
	}

	// 4. Build Complete DSSWorkflowResults from Request
	workflow := state.DSSWorkflow
	workflow.LastUpdated = time.Now()

	// Step 1: Goal Prioritization
	if len(req.GoalPriorities) > 0 {
		rankings := make([]domain.GoalRanking, len(req.GoalPriorities))
		for i, gp := range req.GoalPriorities {
			rankings[i] = domain.GoalRanking{
				GoalID: gp.GoalID,
				Score:  gp.Priority,
				Rank:   i + 1,
			}
		}
		workflow.GoalPrioritization = &domain.GoalPriorityResult{
			Rankings: rankings,
			Method:   "user_selected",
		}
		workflow.MarkStepCompleted(1)
	}

	// Step 2: Debt Strategy
	if req.DebtStrategy != nil {
		workflow.DebtStrategy = &domain.DebtStrategyResult{
			Strategy: *req.DebtStrategy,
		}
		workflow.MarkStepCompleted(2)
	}

	// Step 3: Goal-Debt Tradeoff
	if req.TradeoffChoice != nil {
		workflow.GoalDebtTradeoff = &domain.GoalDebtTradeoffResult{
			SelectedScenario:      req.TradeoffChoice.ScenarioType,
			GoalAllocationPercent: req.TradeoffChoice.GoalAllocationPct,
			DebtAllocationPercent: req.TradeoffChoice.DebtAllocationPct,
		}
		workflow.MarkStepCompleted(3)
	}

	// Step 4: Budget Allocation
	if len(req.BudgetAllocations) > 0 {
		workflow.BudgetAllocation = &domain.BudgetAllocationResult{
			Recommendations: req.BudgetAllocations,
			Method:          "user_finalized",
		}
		workflow.MarkStepCompleted(4)
	}

	// Mark workflow as complete
	workflow.CurrentStep = 5 // All steps done

	// 5. Update CategoryStates with allocations
	if len(req.BudgetAllocations) > 0 {
		// Ensure CategoryStates map is initialized
		if state.CategoryStates == nil {
			state.CategoryStates = make(map[uuid.UUID]*domain.CategoryState)
		}

		for catID, amount := range req.BudgetAllocations {
			if cs, exists := state.CategoryStates[catID]; exists {
				cs.Assigned = amount
				cs.Available = cs.Available + amount
			} else {
				// Create new category state
				state.CategoryStates[catID] = &domain.CategoryState{
					CategoryID: catID,
					Assigned:   amount,
					Activity:   0,
					Available:  amount,
				}
			}
		}
	}

	// 6. Calculate remaining TBB
	totalBudgeted := 0.0
	for _, amount := range req.BudgetAllocations {
		totalBudgeted += amount
	}
	state.ToBeBudgeted = state.ActualIncome - totalBudgeted

	// 7. Mark state as applied
	state.IsApplied = true

	// 8. Create NEW MonthState version (Approach 2: complete snapshot)
	newVersion := len(month.States) + 1
	newState := *state // Copy current state
	newState.Version = newVersion
	newState.CreatedAt = time.Now()

	// Append new state version
	month.States = append(month.States, newState)
	month.IncrementVersion() // Increment Month.Version for optimistic locking

	// 9. Save to DB
	if err := s.repo.UpdateMonth(ctx, month); err != nil {
		return nil, fmt.Errorf("failed to save month with finalized DSS: %w", err)
	}

	s.logger.Info("Successfully finalized DSS workflow",
		zap.String("month_id", monthID.String()),
		zap.Int("new_state_version", newVersion),
		zap.Float64("tbb_remaining", state.ToBeBudgeted),
	)

	// Clear Redis cache after successful DB save
	if err := s.dssCache.ClearPreviews(ctx, monthID, *userID); err != nil {
		s.logger.Warn("Failed to clear DSS cache after finalization", zap.Error(err))
		// Don't fail - cache clearing is not critical
	}

	// 10. Build Response
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
			CanProceed:     false, // All done
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
