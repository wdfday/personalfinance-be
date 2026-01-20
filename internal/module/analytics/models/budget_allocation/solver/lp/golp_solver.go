//go:build cgo && golp
// +build cgo,golp

package lp

/*
#cgo CFLAGS: -I/opt/homebrew/include
#cgo LDFLAGS: -L/opt/homebrew/lib -llpsolve55
*/
import "C"

import (
	"errors"
	"math"

	"github.com/draffensperger/golp"
)

// =============================================================================
// golp Solver Implementation (CGO + lp_solve)
// =============================================================================

// GolpSolver wraps the golp library (lp_solve)
type GolpSolver struct {
	numVars     int
	maximize    bool
	constraints []golpConstraint
	objective   []float64
	lowerBounds []float64
	upperBounds []float64
	binaryVars  []int
}

type golpConstraint struct {
	coefficients []float64
	op           string
	rhs          float64
}

// NewGolpSolver creates a new golp-based LP solver
func NewGolpSolver(numVars int) (*GolpSolver, error) {
	// We don't create s.lp here anymore, we create it in Solve
	lower := make([]float64, numVars)
	upper := make([]float64, numVars)
	for i := range upper {
		upper[i] = math.Inf(1) // No upper bound
	}

	return &GolpSolver{
		numVars:     numVars,
		maximize:    false,
		constraints: make([]golpConstraint, 0),
		objective:   make([]float64, numVars),
		lowerBounds: lower,
		upperBounds: upper,
		binaryVars:  make([]int, 0),
	}, nil
}

func (s *GolpSolver) GetName() string {
	return "golp-lp_solve"
}

func (s *GolpSolver) SetObjective(coefficients []float64, maximize bool) error {
	if len(coefficients) != s.numVars {
		return errors.New("coefficient count must match number of variables")
	}
	s.objective = make([]float64, len(coefficients))
	copy(s.objective, coefficients)
	s.maximize = maximize
	return nil
}

func (s *GolpSolver) AddConstraint(coefficients []float64, op string, rhs float64) error {
	if len(coefficients) != s.numVars {
		return errors.New("coefficient count must match number of variables")
	}
	if op != "<=" && op != ">=" && op != "=" {
		return errors.New("operator must be <=, >=, or =")
	}

	s.constraints = append(s.constraints, golpConstraint{
		coefficients: coefficients,
		op:           op,
		rhs:          rhs,
	})
	return nil
}

func (s *GolpSolver) SetBounds(varIndex int, lower, upper float64) error {
	if varIndex < 0 || varIndex >= s.numVars {
		return errors.New("variable index out of range")
	}
	s.lowerBounds[varIndex] = lower
	s.upperBounds[varIndex] = upper
	return nil
}

func (s *GolpSolver) SetBinary(varIndex int) error {
	if varIndex < 0 || varIndex >= s.numVars {
		return errors.New("variable index out of range")
	}
	s.binaryVars = append(s.binaryVars, varIndex)
	return nil
}

func (s *GolpSolver) Solve() (*LPResult, error) {
	result := &LPResult{
		Solution:   make([]float64, s.numVars),
		SolverName: s.GetName(),
	}

	// Create new LP for solving
	lp := golp.NewLP(0, s.numVars)
	if lp == nil {
		return nil, errors.New("failed to create LP model")
	}

	// Set objective (golp uses 0-indexed columns)
	lp.SetObjFn(s.objective)

	// Set maximize/minimize
	if s.maximize {
		lp.SetMaximize()
	}
	// Note: golp defaults to minimize

	// Add constraints
	for _, con := range s.constraints {
		var conType golp.ConstraintType
		switch con.op {
		case "<=":
			conType = golp.LE
		case ">=":
			conType = golp.GE
		case "=":
			conType = golp.EQ
		}
		err := lp.AddConstraint(con.coefficients, conType, con.rhs)
		if err != nil {
			return nil, err
		}
	}

	// Set bounds
	for i := 0; i < s.numVars; i++ {
		upper := s.upperBounds[i]
		if math.IsInf(upper, 1) {
			upper = 1e30 // Large number instead of infinity
		}
		lp.SetBounds(i, s.lowerBounds[i], upper)
	}

	// Set binary variables
	for _, varIndex := range s.binaryVars {
		lp.SetBinary(varIndex, true)
	}

	// Set verbose level to quiet
	lp.SetVerboseLevel(golp.NEUTRAL)

	// Solve
	solveResult := lp.Solve()

	switch solveResult {
	case golp.OPTIMAL:
		result.Status = LPOptimal
	case golp.INFEASIBLE:
		result.Status = LPInfeasible
	case golp.UNBOUNDED:
		result.Status = LPUnbounded
	default:
		result.Status = LPError
	}

	if result.Status == LPOptimal {
		result.ObjectiveValue = lp.Objective()
		vars := lp.Variables()
		for i := 0; i < s.numVars && i < len(vars); i++ {
			result.Solution[i] = vars[i]
		}
	}

	return result, nil
}

func (s *GolpSolver) Close() {
	// golp doesn't have explicit Delete in new API
}

// CreateGolpSolver creates a golp solver (only available with CGO + golp tag)
func CreateGolpSolver(numVars int) (LPSolver, error) {
	return NewGolpSolver(numVars)
}
