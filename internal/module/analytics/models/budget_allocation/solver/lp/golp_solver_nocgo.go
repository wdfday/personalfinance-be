//go:build !cgo || !golp
// +build !cgo !golp

package lp

import "errors"

// =============================================================================
// Stub implementation khi không có CGO hoặc không có golp tag
// =============================================================================

// GolpSolver stub - không khả dụng khi build không có CGO
type GolpSolver struct{}

// NewGolpSolver returns error when CGO is not available
func NewGolpSolver(numVars int) (*GolpSolver, error) {
	return nil, errors.New("golp solver requires CGO and lp_solve library")
}

func (s *GolpSolver) GetName() string {
	return "golp-lp_solve (unavailable)"
}

func (s *GolpSolver) SetObjective(coefficients []float64, maximize bool) error {
	return errors.New("golp solver not available")
}

func (s *GolpSolver) AddConstraint(coefficients []float64, op string, rhs float64) error {
	return errors.New("golp solver not available")
}

func (s *GolpSolver) SetBounds(varIndex int, lower, upper float64) error {
	return errors.New("golp solver not available")
}

func (s *GolpSolver) SetBinary(varIndex int) error {
	return errors.New("golp solver not available")
}

func (s *GolpSolver) Solve() (*LPResult, error) {
	return nil, errors.New("golp solver not available")
}

func (s *GolpSolver) Close() {}

// CreateGolpSolver returns error when CGO is not available
func CreateGolpSolver(numVars int) (LPSolver, error) {
	return nil, errors.New("golp solver requires CGO and lp_solve library. Install with: brew install lp_solve (macOS) or apt-get install lp-solve (Ubuntu)")
}
