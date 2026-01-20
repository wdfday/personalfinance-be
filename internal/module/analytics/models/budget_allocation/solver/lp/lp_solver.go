package lp

// =============================================================================
// LP Solver Interface - Cho phép swap giữa các implementation
// =============================================================================

// LPSolver interface cho Linear Programming solver
type LPSolver interface {
	// SetObjective đặt hàm mục tiêu (minimize hoặc maximize)
	SetObjective(coefficients []float64, maximize bool) error

	// AddConstraint thêm ràng buộc: sum(a[i]*x[i]) <op> rhs
	// op: "<=", ">=", "="
	AddConstraint(coefficients []float64, op string, rhs float64) error

	// SetBounds đặt giới hạn cho biến: lower <= x[i] <= upper
	SetBounds(varIndex int, lower, upper float64) error

	// SetBinary marks a variable as binary (0 or 1)
	SetBinary(varIndex int) error

	// Solve giải bài toán LP
	Solve() (*LPResult, error)

	// GetName trả về tên solver
	GetName() string

	// Close releases resources
	Close()
}

// LPResult kết quả từ LP solver
type LPResult struct {
	Solution       []float64
	ObjectiveValue float64
	Status         LPStatus
	Iterations     int
	SolverName     string
}

// LPStatus trạng thái giải
type LPStatus int

const (
	LPOptimal LPStatus = iota
	LPUnbounded
	LPInfeasible
	LPMaxIterations
	LPError
)

func (s LPStatus) String() string {
	switch s {
	case LPOptimal:
		return "Optimal"
	case LPUnbounded:
		return "Unbounded"
	case LPInfeasible:
		return "Infeasible"
	case LPMaxIterations:
		return "MaxIterations"
	default:
		return "Error"
	}
}
