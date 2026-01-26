package fuzzy

import (
	"fmt"
	"math"
	"personalfinancedss/internal/module/analytics/models/budget_allocation/solver/lp"

	"github.com/google/uuid"
)

// SolveWithGolp solves the Fuzzy GP problem using Golp (lp_solve) via MILP
// Uses piecewise linear approximation of membership functions
func (s *FuzzyGPSolver) SolveWithGolp() (*FuzzyResult, error) {
	// 1. Calculate number of variables
	// Continuous vars: s.variables (allocations)
	// Binary vars: for each goal, one for each segment of membership function
	// For triangular: 3 segments (rising, peak, falling)
	// For trapezoidal: 4 segments (rising, plateau-left, plateau-right, falling)
	numContinuous := len(s.variables)
	numBinary := 0

	// Map to track binary variable indices for each segment
	type segmentVarKey struct {
		goalIdx    int
		segmentIdx int
	}
	binaryVarMap := make(map[segmentVarKey]int)
	segmentInfo := make(map[segmentVarKey]struct {
		lower, upper float64
		slope        float64
		intercept    float64
	})

	// Process each goal to create segments
	for i, goal := range s.goals {
		mf := goal.MembershipFunc
		var segments []struct {
			lower, upper float64
			slope        float64
			intercept    float64
		}

		switch mf.Type {
		case "triangular":
			// Triangular: 3 segments
			// Segment 1: rising edge (Lower to PeakLeft)
			if mf.PeakLeft > mf.Lower {
				slope1 := 1.0 / (mf.PeakLeft - mf.Lower)
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{mf.Lower, mf.PeakLeft, slope1, -mf.Lower * slope1})
			}
			// Segment 2: peak (PeakLeft to PeakRight)
			if mf.PeakRight > mf.PeakLeft {
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{mf.PeakLeft, mf.PeakRight, 0, 1.0})
			}
			// Segment 3: falling edge (PeakRight to Upper)
			if mf.Upper > mf.PeakRight {
				slope3 := -1.0 / (mf.Upper - mf.PeakRight)
				intercept3 := 1.0 - mf.PeakRight*slope3
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{mf.PeakRight, mf.Upper, slope3, intercept3})
			}

		case "trapezoidal":
			// Trapezoidal: 4 segments
			// Segment 1: rising edge
			if mf.PeakLeft > mf.Lower {
				slope1 := 1.0 / (mf.PeakLeft - mf.Lower)
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{mf.Lower, mf.PeakLeft, slope1, -mf.Lower * slope1})
			}
			// Segment 2: plateau left
			if mf.PeakRight > mf.PeakLeft {
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{mf.PeakLeft, mf.PeakRight, 0, 1.0})
			}
			// Segment 3: falling edge
			if mf.Upper > mf.PeakRight {
				slope3 := -1.0 / (mf.Upper - mf.PeakRight)
				intercept3 := 1.0 - mf.PeakRight*slope3
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{mf.PeakRight, mf.Upper, slope3, intercept3})
			}

		case "s-curve":
			// S-curve is implemented as a piecewise-linear approximation (no sigmoid/exp),
			// so it can be embedded into MILP safely.
			for _, seg := range buildSigmoidLikeSegments(mf.Lower, mf.PeakLeft, mf.PeakRight, mf.Upper) {
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{seg.Lower, seg.Upper, seg.Slope, seg.Intercept})
			}

		case "piecewise":
			// Piecewise: use segments directly from MembershipFunction.Segments
			// These segments represent cumulative increasing membership (e.g., 0 → 1.0 → 1.7 → 2.0)
			for _, seg := range mf.Segments {
				segments = append(segments, struct {
					lower, upper float64
					slope        float64
					intercept    float64
				}{seg.Lower, seg.Upper, seg.Slope, seg.Intercept})
			}
		}

		// Store segment info and create binary variables
		for j, seg := range segments {
			key := segmentVarKey{i, j}
			binaryVarMap[key] = numContinuous + numBinary
			segmentInfo[key] = seg
			numBinary++
		}
	}

	totalVars := numContinuous + numBinary

	// 2. Create Solver
	lpSolver, err := lp.CreateGolpSolver(totalVars)
	if err != nil {
		return nil, fmt.Errorf("failed to create golp solver: %w", err)
	}
	defer lpSolver.Close()

	// 3. Set Objective Function: Maximize Weighted Membership
	// Maximize sum(membership_i * weight_i)
	// Membership is approximated by piecewise linear segments
	const baseWeight = 1000.0
	objCoeffs := make([]float64, totalVars)

	// Continuous variables: small negative to prefer saving
	epsilon := 1e-6
	for i := 0; i < numContinuous; i++ {
		objCoeffs[i] = -epsilon
	}

	// Binary variables: represent membership contribution
	for i, goal := range s.goals {
		// Reduced priority scaling to avoid numerical issues
		// Priority 1: 1e6, Priority 2: 1e4, Priority 3: 1e2, Priority 4: 1
		// This ensures strict priority ordering without extreme values
		priorityPower := 6.0 - float64(goal.Priority)*2.0
		if priorityPower < 0 {
			priorityPower = 0
		}
		priorityWeight := math.Pow(10.0, priorityPower) // Use 10 instead of 1000
		goalWeight := goal.Weight * priorityWeight

		// Each segment contributes to membership
		for j := 0; ; j++ {
			key := segmentVarKey{i, j}
			binIdx, exists := binaryVarMap[key]
			if !exists {
				break
			}
			seg := segmentInfo[key]
			// Approximate membership contribution: use midpoint of segment
			midpoint := (seg.lower + seg.upper) / 2
			membershipAtMidpoint := seg.slope*midpoint + seg.intercept
			if membershipAtMidpoint < 0 {
				membershipAtMidpoint = 0
			}
			// Calculate max membership at upper bound (for piecewise, can be > 1.0)
			maxMembershipAtUpper := seg.slope*seg.upper + seg.intercept
			if membershipAtMidpoint > maxMembershipAtUpper {
				membershipAtMidpoint = maxMembershipAtUpper
			}
			// Don't cap at 1.0 - allow membership > 1.0 for flexible surplus tiers
			objCoeffs[binIdx] = membershipAtMidpoint * goalWeight
		}
	}

	if err := lpSolver.SetObjective(objCoeffs, true); err != nil {
		return nil, fmt.Errorf("failed to set objective: %w", err)
	}

	// 4. Set Variable Properties
	for i := 0; i < numContinuous; i++ {
		v := s.variables[i]
		upper := math.Inf(1)
		if v.MaxValue > 0 {
			upper = v.MaxValue
		}
		if err := lpSolver.SetBounds(i, v.MinValue, upper); err != nil {
			return nil, fmt.Errorf("failed to set bounds for variable %d: %w", i, err)
		}
	}

	// Binary variables for segments
	for _, binIdx := range binaryVarMap {
		if err := lpSolver.SetBinary(binIdx); err != nil {
			return nil, fmt.Errorf("failed to set binary for variable %d: %w", binIdx, err)
		}
	}

	// 5. Add Constraints

	// 5.0 Hard constraints: Ensure minimums are allocated first
	// Phase 1: Mandatory expenses must be exactly at minimum
	for i, v := range s.variables {
		if v.Type == "category" && v.MinValue > 0 && v.MinValue == v.MaxValue {
			// Mandatory expense: fixed amount (min == max)
			mandatoryCoeffs := make([]float64, totalVars)
			mandatoryCoeffs[i] = 1.0
			if err := lpSolver.AddConstraint(mandatoryCoeffs, "=", v.MinValue); err != nil {
				return nil, fmt.Errorf("failed to add mandatory expense constraint: %w", err)
			}
		}
	}

	// Phase 2: Debt payments must be at least minimum
	for i, v := range s.variables {
		if v.Type == "debt" && v.MinValue > 0 {
			// Debt: must pay at least minimum
			debtCoeffs := make([]float64, totalVars)
			debtCoeffs[i] = 1.0
			if err := lpSolver.AddConstraint(debtCoeffs, ">=", v.MinValue); err != nil {
				return nil, fmt.Errorf("failed to add minimum debt payment constraint: %w", err)
			}
		}
	}

	// 5.1 Budget constraint: sum of allocations = totalIncome (must allocate all)
	budgetCoeffs := make([]float64, totalVars)
	for i := 0; i < numContinuous; i++ {
		budgetCoeffs[i] = 1.0
	}
	if err := lpSolver.AddConstraint(budgetCoeffs, "=", s.totalIncome); err != nil {
		return nil, fmt.Errorf("failed to add budget constraint: %w", err)
	}

	// 5.2 For each goal: exactly one segment must be active
	for i := range s.goals {
		goalSegments := make([]int, 0)
		for j := 0; ; j++ {
			key := segmentVarKey{i, j}
			if binIdx, exists := binaryVarMap[key]; exists {
				goalSegments = append(goalSegments, binIdx)
			} else {
				break
			}
		}

		if len(goalSegments) > 0 {
			segCoeffs := make([]float64, totalVars)
			for _, binIdx := range goalSegments {
				segCoeffs[binIdx] = 1.0
			}
			if err := lpSolver.AddConstraint(segCoeffs, "=", 1.0); err != nil {
				return nil, fmt.Errorf("failed to add segment selection constraint: %w", err)
			}
		}
	}

	// 5.3 Link allocation variables to segment selection
	// For each goal and segment: if segment is selected, allocation must be in segment range
	for i, goal := range s.goals {
		varIdx := goal.VariableIdx
		for j := 0; ; j++ {
			key := segmentVarKey{i, j}
			binIdx, exists := binaryVarMap[key]
			if !exists {
				break
			}
			seg := segmentInfo[key]

			// Constraint: var >= seg.lower * binary
			lowerCoeffs := make([]float64, totalVars)
			lowerCoeffs[varIdx] = 1.0
			lowerCoeffs[binIdx] = -seg.lower
			if err := lpSolver.AddConstraint(lowerCoeffs, ">=", 0); err != nil {
				return nil, fmt.Errorf("failed to add lower bound constraint: %w", err)
			}

			// Constraint: var <= seg.upper * binary + M * (1 - binary)
			// Simplified: var <= seg.upper + M * (1 - binary)
			// Where M is large number (totalIncome)
			M := s.totalIncome
			upperCoeffs := make([]float64, totalVars)
			upperCoeffs[varIdx] = 1.0
			upperCoeffs[binIdx] = M - seg.upper
			if err := lpSolver.AddConstraint(upperCoeffs, "<=", M); err != nil {
				return nil, fmt.Errorf("failed to add upper bound constraint: %w", err)
			}
		}
	}

	// 6. Solve
	lpResult, err := lpSolver.Solve()
	if err != nil {
		return nil, fmt.Errorf("failed to solve: %w", err)
	}

	if lpResult.Status != lp.LPOptimal {
		return nil, fmt.Errorf("solver returned non-optimal status: %v", lpResult.Status)
	}

	// 7. Extract solution
	result := &FuzzyResult{
		VariableValues:  make(map[uuid.UUID]float64),
		GoalMemberships: make(map[string]float64),
		GoalValues:      make(map[string]float64),
		Iterations:      lpResult.Iterations,
	}

	// Extract allocation values
	for i, v := range s.variables {
		value := lpResult.Solution[i]
		if value > 1e-6 { // Only store non-zero values
			result.VariableValues[v.ID] = value
		}
	}

	// Calculate memberships from solution
	for i, goal := range s.goals {
		varIdx := goal.VariableIdx
		value := lpResult.Solution[varIdx]
		result.GoalValues[goal.ID] = value

		// Find which segment is active
		activeSegment := -1
		for j := 0; j < 10; j++ {
			key := segmentVarKey{i, j}
			binIdx, exists := binaryVarMap[key]
			if !exists {
				break
			}
			if lpResult.Solution[binIdx] > 0.5 { // Binary variable is 1
				activeSegment = j
				break
			}
		}

		// Calculate membership based on active segment
		if activeSegment >= 0 {
			key := segmentVarKey{i, activeSegment}
			seg := segmentInfo[key]
			membership := seg.slope*value + seg.intercept
			if membership < 0 {
				membership = 0
			}
			// Calculate max membership at upper bound (for piecewise, can be > 1.0)
			maxMembershipAtUpper := seg.slope*seg.upper + seg.intercept
			if membership > maxMembershipAtUpper {
				membership = maxMembershipAtUpper
			}
			// Don't cap at 1.0 - allow membership > 1.0 for FlexibleExpenses (piecewise)
			result.GoalMemberships[goal.ID] = membership
		} else {
			result.GoalMemberships[goal.ID] = 0
		}
	}

	// Categorize goals
	for _, goal := range s.goals {
		membership := result.GoalMemberships[goal.ID]
		if membership >= 0.8 {
			result.AchievedGoals = append(result.AchievedGoals, goal.Description)
		} else if membership >= 0.3 {
			result.PartialGoals = append(result.PartialGoals, goal.Description)
		} else {
			result.UnachievedGoals = append(result.UnachievedGoals, goal.Description)
		}
	}

	// Calculate totals
	totalWeightedMembership := 0.0
	totalWeight := 0.0
	for _, goal := range s.goals {
		membership := result.GoalMemberships[goal.ID]
		result.TotalMembership += membership
		totalWeightedMembership += membership * goal.Weight
		totalWeight += goal.Weight
	}
	if len(s.goals) > 0 {
		result.AverageMembership = result.TotalMembership / float64(len(s.goals))
	}
	if totalWeight > 0 {
		result.WeightedMembership = totalWeightedMembership / totalWeight
	}

	return result, nil
}

// Helper method to get segments from membership function
func (mf MembershipFunction) getSegments() []struct {
	lower, upper float64
	slope        float64
	intercept    float64
} {
	var segments []struct {
		lower, upper float64
		slope        float64
		intercept    float64
	}

	switch mf.Type {
	case "triangular":
		if mf.PeakLeft > mf.Lower {
			slope1 := 1.0 / (mf.PeakLeft - mf.Lower)
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{mf.Lower, mf.PeakLeft, slope1, -mf.Lower * slope1})
		}
		if mf.PeakRight > mf.PeakLeft {
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{mf.PeakLeft, mf.PeakRight, 0, 1.0})
		}
		if mf.Upper > mf.PeakRight {
			slope3 := -1.0 / (mf.Upper - mf.PeakRight)
			intercept3 := 1.0 - mf.PeakRight*slope3
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{mf.PeakRight, mf.Upper, slope3, intercept3})
		}

	case "trapezoidal":
		if mf.PeakLeft > mf.Lower {
			slope1 := 1.0 / (mf.PeakLeft - mf.Lower)
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{mf.Lower, mf.PeakLeft, slope1, -mf.Lower * slope1})
		}
		if mf.PeakRight > mf.PeakLeft {
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{mf.PeakLeft, mf.PeakRight, 0, 1.0})
		}
		if mf.Upper > mf.PeakRight {
			slope3 := -1.0 / (mf.Upper - mf.PeakRight)
			intercept3 := 1.0 - mf.PeakRight*slope3
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{mf.PeakRight, mf.Upper, slope3, intercept3})
		}

	case "s-curve":
		// S-curve is implemented as piecewise-linear (no sigmoid/exp).
		for _, seg := range buildSigmoidLikeSegments(mf.Lower, mf.PeakLeft, mf.PeakRight, mf.Upper) {
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{seg.Lower, seg.Upper, seg.Slope, seg.Intercept})
		}

	case "piecewise":
		// Piecewise: use segments directly from MembershipFunction.Segments
		// These segments represent cumulative increasing membership (e.g., 0 → 1.0 → 1.7 → 2.0)
		for _, seg := range mf.Segments {
			segments = append(segments, struct {
				lower, upper float64
				slope        float64
				intercept    float64
			}{seg.Lower, seg.Upper, seg.Slope, seg.Intercept})
		}
	}

	return segments
}
