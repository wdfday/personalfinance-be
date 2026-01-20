package lp

import (
	"errors"
	"math"
)

// =============================================================================
// Pure Go Simplex Solver Implementation
// Two-Phase Simplex Method
// =============================================================================

// PureGoSimplexSolver implements the Simplex algorithm (Pure Go, no CGO)
type PureGoSimplexSolver struct {
	c           []float64   // Objective coefficients
	A           [][]float64 // Constraint matrix
	b           []float64   // Right-hand side
	ops         []string    // Constraint operators
	lowerBounds []float64   // Lower bounds for variables
	upperBounds []float64   // Upper bounds for variables
	numVars     int
	maximize    bool
	maxIter     int
	tolerance   float64
}

// NewPureGoSimplexSolver creates a new Pure Go Simplex solver
func NewPureGoSimplexSolver(numVars int) *PureGoSimplexSolver {
	lower := make([]float64, numVars)
	upper := make([]float64, numVars)
	for i := range upper {
		upper[i] = math.Inf(1) // No upper bound by default
	}

	return &PureGoSimplexSolver{
		c:           make([]float64, numVars),
		A:           make([][]float64, 0),
		b:           make([]float64, 0),
		ops:         make([]string, 0),
		lowerBounds: lower,
		upperBounds: upper,
		numVars:     numVars,
		maximize:    false,
		maxIter:     1000,
		tolerance:   1e-9,
	}
}

func (s *PureGoSimplexSolver) GetName() string {
	return "PureGo-Simplex"
}

func (s *PureGoSimplexSolver) Close() {
	// Nothing to close
}

func (s *PureGoSimplexSolver) SetObjective(coefficients []float64, maximize bool) error {
	if len(coefficients) != s.numVars {
		return errors.New("coefficient count must match number of variables")
	}
	s.c = make([]float64, len(coefficients))
	copy(s.c, coefficients)
	s.maximize = maximize
	return nil
}

func (s *PureGoSimplexSolver) AddConstraint(coefficients []float64, op string, rhs float64) error {
	if len(coefficients) != s.numVars {
		return errors.New("coefficient count must match number of variables")
	}
	if op != "<=" && op != ">=" && op != "=" {
		return errors.New("operator must be <=, >=, or =")
	}

	row := make([]float64, len(coefficients))
	copy(row, coefficients)
	s.A = append(s.A, row)
	s.b = append(s.b, rhs)
	s.ops = append(s.ops, op)
	return nil
}

func (s *PureGoSimplexSolver) SetBounds(varIndex int, lower, upper float64) error {
	if varIndex < 0 || varIndex >= s.numVars {
		return errors.New("variable index out of range")
	}
	s.lowerBounds[varIndex] = lower
	s.upperBounds[varIndex] = upper
	return nil
}

func (s *PureGoSimplexSolver) SetBinary(varIndex int) error {
	return errors.New("binary variables not supported in PureGoSimplexSolver")
}

// Solve implements the Two-Phase Simplex Method
func (s *PureGoSimplexSolver) Solve() (*LPResult, error) {
	result := &LPResult{
		Solution:   make([]float64, s.numVars),
		SolverName: s.GetName(),
	}

	if s.numVars == 0 {
		result.Status = LPOptimal
		return result, nil
	}

	// Convert to standard form and solve
	solution, objValue, status, iterations := s.solveStandardForm()

	result.Solution = solution
	result.ObjectiveValue = objValue
	result.Status = status
	result.Iterations = iterations

	return result, nil
}

// solveStandardForm converts to standard form and applies simplex
func (s *PureGoSimplexSolver) solveStandardForm() ([]float64, float64, LPStatus, int) {
	numConstraints := len(s.A)
	if numConstraints == 0 {
		// No constraints - use bounds only
		solution := make([]float64, s.numVars)
		objValue := 0.0
		for i := 0; i < s.numVars; i++ {
			if s.maximize {
				if s.c[i] > 0 && !math.IsInf(s.upperBounds[i], 1) {
					solution[i] = s.upperBounds[i]
				} else {
					solution[i] = s.lowerBounds[i]
				}
			} else {
				if s.c[i] < 0 && !math.IsInf(s.upperBounds[i], 1) {
					solution[i] = s.upperBounds[i]
				} else {
					solution[i] = s.lowerBounds[i]
				}
			}
			objValue += s.c[i] * solution[i]
		}
		return solution, objValue, LPOptimal, 0
	}

	// Count slack/surplus/artificial variables needed
	numSlack := 0
	numArtificial := 0
	for _, op := range s.ops {
		switch op {
		case "<=":
			numSlack++
		case ">=":
			numSlack++
			numArtificial++
		case "=":
			numArtificial++
		}
	}

	totalVars := s.numVars + numSlack + numArtificial

	// Build tableau
	tableau := make([][]float64, numConstraints+1)
	for i := range tableau {
		tableau[i] = make([]float64, totalVars+1)
	}

	slackIdx := s.numVars
	artificialIdx := s.numVars + numSlack
	basicVars := make([]int, numConstraints)

	// Fill constraint rows
	for i := 0; i < numConstraints; i++ {
		for j := 0; j < s.numVars; j++ {
			tableau[i][j] = s.A[i][j]
		}

		rhs := s.b[i]

		switch s.ops[i] {
		case "<=":
			tableau[i][slackIdx] = 1
			basicVars[i] = slackIdx
			slackIdx++
		case ">=":
			tableau[i][slackIdx] = -1
			slackIdx++
			tableau[i][artificialIdx] = 1
			basicVars[i] = artificialIdx
			artificialIdx++
		case "=":
			tableau[i][artificialIdx] = 1
			basicVars[i] = artificialIdx
			artificialIdx++
		}

		if rhs < 0 {
			for j := 0; j <= totalVars; j++ {
				tableau[i][j] = -tableau[i][j]
			}
			rhs = -rhs
		}
		tableau[i][totalVars] = rhs
	}

	iterations := 0

	// Phase 1: Minimize sum of artificial variables
	if numArtificial > 0 {
		for j := 0; j < totalVars; j++ {
			tableau[numConstraints][j] = 0
		}
		for i := s.numVars + numSlack; i < totalVars; i++ {
			tableau[numConstraints][i] = 1
		}

		for i := 0; i < numConstraints; i++ {
			if basicVars[i] >= s.numVars+numSlack {
				for j := 0; j <= totalVars; j++ {
					tableau[numConstraints][j] -= tableau[i][j]
				}
			}
		}

		var status LPStatus
		status, iterations = s.simplexIterate(tableau, basicVars, totalVars, numConstraints)
		if status != LPOptimal {
			return make([]float64, s.numVars), 0, status, iterations
		}

		if math.Abs(tableau[numConstraints][totalVars]) > s.tolerance {
			return make([]float64, s.numVars), 0, LPInfeasible, iterations
		}
	}

	// Phase 2: Optimize original objective
	for j := 0; j < s.numVars; j++ {
		if s.maximize {
			tableau[numConstraints][j] = -s.c[j]
		} else {
			tableau[numConstraints][j] = s.c[j]
		}
	}
	for j := s.numVars; j <= totalVars; j++ {
		tableau[numConstraints][j] = 0
	}

	for i := 0; i < numConstraints; i++ {
		if basicVars[i] < s.numVars {
			coef := tableau[numConstraints][basicVars[i]]
			for j := 0; j <= totalVars; j++ {
				tableau[numConstraints][j] -= coef * tableau[i][j]
			}
		}
	}

	status, iter2 := s.simplexIterate(tableau, basicVars, totalVars, numConstraints)
	iterations += iter2

	// Extract solution
	solution := make([]float64, s.numVars)
	for i := 0; i < numConstraints; i++ {
		if basicVars[i] < s.numVars {
			solution[basicVars[i]] = tableau[i][totalVars]
		}
	}

	// Apply bounds
	for i := 0; i < s.numVars; i++ {
		solution[i] = math.Max(s.lowerBounds[i], solution[i])
		if !math.IsInf(s.upperBounds[i], 1) {
			solution[i] = math.Min(s.upperBounds[i], solution[i])
		}
	}

	objValue := -tableau[numConstraints][totalVars]
	if s.maximize {
		objValue = -objValue
	}

	return solution, objValue, status, iterations
}

// simplexIterate performs simplex iterations
func (s *PureGoSimplexSolver) simplexIterate(tableau [][]float64, basicVars []int, totalVars, numConstraints int) (LPStatus, int) {
	iterations := 0

	for iterations < s.maxIter {
		iterations++

		// Find entering variable (most negative coefficient)
		enterCol := -1
		minCoef := -s.tolerance
		for j := 0; j < totalVars; j++ {
			if tableau[numConstraints][j] < minCoef {
				minCoef = tableau[numConstraints][j]
				enterCol = j
			}
		}

		if enterCol == -1 {
			return LPOptimal, iterations
		}

		// Find leaving variable (minimum ratio test)
		leaveRow := -1
		minRatio := math.Inf(1)
		for i := 0; i < numConstraints; i++ {
			if tableau[i][enterCol] > s.tolerance {
				ratio := tableau[i][totalVars] / tableau[i][enterCol]
				if ratio < minRatio {
					minRatio = ratio
					leaveRow = i
				}
			}
		}

		if leaveRow == -1 {
			return LPUnbounded, iterations
		}

		// Pivot
		pivot := tableau[leaveRow][enterCol]
		for j := 0; j <= totalVars; j++ {
			tableau[leaveRow][j] /= pivot
		}

		for i := 0; i <= numConstraints; i++ {
			if i != leaveRow {
				factor := tableau[i][enterCol]
				for j := 0; j <= totalVars; j++ {
					tableau[i][j] -= factor * tableau[leaveRow][j]
				}
			}
		}

		basicVars[leaveRow] = enterCol
	}

	return LPMaxIterations, iterations
}
