package engine

import (
	"math"
	"personalfinancedss/internal/module/analytics/debt_strategy/domain"
	"sort"
)

// DebtSimulator handles debt payoff simulation
type DebtSimulator struct {
	maxMonths int
}

// NewDebtSimulator creates a new simulator
func NewDebtSimulator() *DebtSimulator {
	return &DebtSimulator{maxMonths: 360} // 30 years max
}

// SimulateStrategy runs simulation for a specific strategy
func (s *DebtSimulator) SimulateStrategy(
	strategy domain.Strategy,
	debts []domain.DebtInfo,
	extraPayment float64,
	hybridWeights *domain.HybridWeights,
) *domain.SimulationResult {
	// Copy debts to avoid modifying original
	currentDebts := make([]domain.DebtInfo, len(debts))
	copy(currentDebts, debts)

	// Initialize timelines
	timelines := make(map[string]*domain.DebtTimeline)
	for _, debt := range currentDebts {
		timelines[debt.ID] = &domain.DebtTimeline{
			DebtID:    debt.ID,
			DebtName:  debt.Name,
			Snapshots: make([]domain.MonthlySnapshot, 0),
		}
	}

	month := 0
	totalInterest := 0.0
	firstCleared := 0

	for {
		// Check if all debts paid
		allPaid := true
		for _, debt := range currentDebts {
			if debt.Balance > 0.01 {
				allPaid = false
				break
			}
		}

		if allPaid || month >= s.maxMonths {
			break
		}

		month++

		// Step 1: Record start balance and apply interest
		for i := range currentDebts {
			if currentDebts[i].Balance > 0 {
				startBalance := currentDebts[i].Balance
				monthlyRate := currentDebts[i].InterestRate / 12
				interest := currentDebts[i].Balance * monthlyRate
				currentDebts[i].Balance += interest
				totalInterest += interest

				timelines[currentDebts[i].ID].Snapshots = append(
					timelines[currentDebts[i].ID].Snapshots,
					domain.MonthlySnapshot{
						Month:        month,
						StartBalance: startBalance,
						Interest:     interest,
						Payment:      0,
						EndBalance:   0,
					},
				)
			}
		}

		// Step 2: Pay minimum on all debts
		for i := range currentDebts {
			if currentDebts[i].Balance > 0 {
				payment := math.Min(currentDebts[i].MinimumPayment, currentDebts[i].Balance)
				currentDebts[i].Balance -= payment

				idx := len(timelines[currentDebts[i].ID].Snapshots) - 1
				if idx >= 0 {
					timelines[currentDebts[i].ID].Snapshots[idx].Payment += payment
				}
			}
		}

		// Step 3: Allocate extra payment based on strategy
		remainingExtra := extraPayment
		sortedIndices := s.sortDebtsByStrategy(currentDebts, strategy, hybridWeights)

		for _, idx := range sortedIndices {
			if remainingExtra <= 0 {
				break
			}
			if currentDebts[idx].Balance > 0 {
				payment := math.Min(remainingExtra, currentDebts[idx].Balance)
				currentDebts[idx].Balance -= payment
				remainingExtra -= payment

				tlIdx := len(timelines[currentDebts[idx].ID].Snapshots) - 1
				if tlIdx >= 0 {
					timelines[currentDebts[idx].ID].Snapshots[tlIdx].Payment += payment
				}
			}
		}

		// Step 4: Update end balances and check for cleared debts
		for i := range currentDebts {
			tlIdx := len(timelines[currentDebts[i].ID].Snapshots) - 1
			if tlIdx >= 0 {
				timelines[currentDebts[i].ID].Snapshots[tlIdx].EndBalance = currentDebts[i].Balance

				// Mark payoff month
				if currentDebts[i].Balance < 0.01 && timelines[currentDebts[i].ID].PayoffMonth == 0 {
					timelines[currentDebts[i].ID].PayoffMonth = month
					if firstCleared == 0 {
						firstCleared = month
					}
				}
			}
		}
	}

	// Calculate totals for each debt
	for _, timeline := range timelines {
		for _, snapshot := range timeline.Snapshots {
			timeline.TotalInterest += snapshot.Interest
			timeline.TotalPrincipal += snapshot.Payment - snapshot.Interest
		}
	}

	return &domain.SimulationResult{
		Strategy:      strategy,
		Months:        month,
		TotalInterest: totalInterest,
		DebtTimelines: timelines,
		FirstCleared:  firstCleared,
	}
}

// SimulateWithLumpSum simulates paying a lump sum to a specific debt
func (s *DebtSimulator) SimulateWithLumpSum(
	strategy domain.Strategy,
	debts []domain.DebtInfo,
	extraPayment float64,
	lumpSum float64,
	targetDebtID string,
	hybridWeights *domain.HybridWeights,
) *domain.SimulationResult {
	// Apply lump sum to target debt first
	modifiedDebts := make([]domain.DebtInfo, len(debts))
	copy(modifiedDebts, debts)

	for i := range modifiedDebts {
		if modifiedDebts[i].ID == targetDebtID {
			modifiedDebts[i].Balance = math.Max(0, modifiedDebts[i].Balance-lumpSum)
			break
		}
	}

	return s.SimulateStrategy(strategy, modifiedDebts, extraPayment, hybridWeights)
}

// FindBestDebtForLumpSum finds which debt benefits most from lump sum
func (s *DebtSimulator) FindBestDebtForLumpSum(
	strategy domain.Strategy,
	debts []domain.DebtInfo,
	extraPayment float64,
	lumpSum float64,
	hybridWeights *domain.HybridWeights,
) (string, float64) {
	baseline := s.SimulateStrategy(strategy, debts, extraPayment, hybridWeights)

	bestDebtID := ""
	bestSavings := 0.0

	for _, debt := range debts {
		if debt.Balance <= 0 {
			continue
		}
		result := s.SimulateWithLumpSum(strategy, debts, extraPayment, lumpSum, debt.ID, hybridWeights)
		savings := baseline.TotalInterest - result.TotalInterest
		if savings > bestSavings {
			bestSavings = savings
			bestDebtID = debt.ID
		}
	}

	return bestDebtID, bestSavings
}

// sortDebtsByStrategy returns sorted indices based on strategy
func (s *DebtSimulator) sortDebtsByStrategy(
	debts []domain.DebtInfo,
	strategy domain.Strategy,
	hybridWeights *domain.HybridWeights,
) []int {
	indices := make([]int, len(debts))
	for i := range indices {
		indices[i] = i
	}

	switch strategy {
	case domain.StrategyAvalanche:
		// Sort by interest rate descending (highest first)
		sort.Slice(indices, func(i, j int) bool {
			return debts[indices[i]].InterestRate > debts[indices[j]].InterestRate
		})

	case domain.StrategySnowball:
		// Sort by balance ascending (smallest first)
		sort.Slice(indices, func(i, j int) bool {
			return debts[indices[i]].Balance < debts[indices[j]].Balance
		})

	case domain.StrategyCashFlow:
		// Sort by min_payment/balance ratio descending (highest ratio first)
		sort.Slice(indices, func(i, j int) bool {
			ratioI := 0.0
			ratioJ := 0.0
			if debts[indices[i]].Balance > 0 {
				ratioI = debts[indices[i]].MinimumPayment / debts[indices[i]].Balance
			}
			if debts[indices[j]].Balance > 0 {
				ratioJ = debts[indices[j]].MinimumPayment / debts[indices[j]].Balance
			}
			return ratioI > ratioJ
		})

	case domain.StrategyStress:
		// Sort by stress score descending (highest stress first)
		sort.Slice(indices, func(i, j int) bool {
			return debts[indices[i]].StressScore > debts[indices[j]].StressScore
		})

	case domain.StrategyHybrid:
		// Weighted scoring
		weights := domain.DefaultHybridWeights()
		if hybridWeights != nil {
			weights = *hybridWeights
		}
		scores := s.calculateHybridScores(debts, weights)
		sort.Slice(indices, func(i, j int) bool {
			return scores[indices[i]] > scores[indices[j]]
		})

	default:
		// Default to avalanche
		sort.Slice(indices, func(i, j int) bool {
			return debts[indices[i]].InterestRate > debts[indices[j]].InterestRate
		})
	}

	return indices
}

// calculateHybridScores calculates weighted scores for hybrid strategy
func (s *DebtSimulator) calculateHybridScores(debts []domain.DebtInfo, weights domain.HybridWeights) []float64 {
	scores := make([]float64, len(debts))

	// Find max values for normalization
	maxRate := 0.0
	maxBalance := 0.0
	maxStress := 0
	maxCashFlow := 0.0

	for _, d := range debts {
		if d.InterestRate > maxRate {
			maxRate = d.InterestRate
		}
		if d.Balance > maxBalance {
			maxBalance = d.Balance
		}
		if d.StressScore > maxStress {
			maxStress = d.StressScore
		}
		if d.Balance > 0 {
			cf := d.MinimumPayment / d.Balance
			if cf > maxCashFlow {
				maxCashFlow = cf
			}
		}
	}

	// Calculate normalized scores
	for i, d := range debts {
		// Interest rate component (higher = higher score)
		rateScore := 0.0
		if maxRate > 0 {
			rateScore = d.InterestRate / maxRate
		}

		// Balance component (lower balance = higher score, so invert)
		balanceScore := 0.0
		if maxBalance > 0 && d.Balance > 0 {
			balanceScore = 1.0 - (d.Balance / maxBalance)
		}

		// Stress component (higher stress = higher score)
		stressScore := 0.0
		if maxStress > 0 {
			stressScore = float64(d.StressScore) / float64(maxStress)
		}

		// Cash flow component (higher ratio = higher score)
		cashFlowScore := 0.0
		if maxCashFlow > 0 && d.Balance > 0 {
			cashFlowScore = (d.MinimumPayment / d.Balance) / maxCashFlow
		}

		// Weighted sum
		scores[i] = weights.InterestRateWeight*rateScore +
			weights.BalanceWeight*balanceScore +
			weights.StressWeight*stressScore +
			weights.CashFlowWeight*cashFlowScore
	}

	return scores
}
