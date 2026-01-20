package ahp

import (
	"context"
	"errors"
	"fmt"
	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"sort"
)

// AHPModel implements mbms.Model interface
type AHPModel struct {
	name        string
	description string
}

// NewAHPModel creates a new AHP model instance
func NewAHPModel() *AHPModel {
	return &AHPModel{
		name:        "goal_prioritization",
		description: "Analytic Hierarchy Process for prioritizing financial goals",
	}
}

// Name returns model name
func (m *AHPModel) Name() string {
	return m.name
}

// Description returns model description
func (m *AHPModel) Description() string {
	return m.description
}

// Dependencies returns model dependencies (AHP has no dependencies)
func (m *AHPModel) Dependencies() []string {
	return []string{} // Không phụ thuộc model nào
}

// Validate validates input
func (m *AHPModel) Validate(ctx context.Context, input interface{}) error {
	ahpInput, ok := input.(*dto.AHPInput)
	if !ok {
		return errors.New("input must be *dto.AHPInput type")
	}

	// Validate số lượng criteria
	if len(ahpInput.Criteria) < 2 {
		return errors.New("at least 2 criteria required")
	}

	// Validate số lượng alternatives
	if len(ahpInput.Alternatives) < 2 {
		return errors.New("at least 2 alternatives required")
	}

	// Validate criteria comparisons: phải có n*(n-1)/2 comparisons
	n := len(ahpInput.Criteria)
	expectedComparisons := n * (n - 1) / 2
	if len(ahpInput.CriteriaComparisons) != expectedComparisons {
		return fmt.Errorf("criteria comparisons: expected %d, got %d",
			expectedComparisons, len(ahpInput.CriteriaComparisons))
	}

	// Validate alternative comparisons cho mỗi criterion
	numAlternatives := len(ahpInput.Alternatives)
	expectedAltComparisons := numAlternatives * (numAlternatives - 1) / 2
	for _, criterion := range ahpInput.Criteria {
		comparisons, exists := ahpInput.AlternativeComparisons[criterion.ID]
		if !exists {
			return fmt.Errorf("missing alternative comparisons for criterion: %s", criterion.ID)
		}
		if len(comparisons) != expectedAltComparisons {
			return fmt.Errorf("criterion %s: expected %d comparisons, got %d",
				criterion.ID, expectedAltComparisons, len(comparisons))
		}
	}

	// Validate comparison values (phải trong khoảng 1/9 đến 9)
	for _, comp := range ahpInput.CriteriaComparisons {
		if comp.Value < 1.0/9.0 || comp.Value > 9.0 {
			return fmt.Errorf("comparison value out of range: %f", comp.Value)
		}
	}

	return nil
}

// Execute runs the AHP algorithm
func (m *AHPModel) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	ahpInput := input.(*dto.AHPInput) // Đã validate ở trên

	// Step 1: Xây dựng và tính priority của Criteria
	criteriaMatrix := m.buildComparisonMatrixForCriteria(
		ahpInput.Criteria,
		ahpInput.CriteriaComparisons,
	)

	criteriaWeights, cr, err := m.calculatePriorities(criteriaMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate criteria weights: %w", err)
	}

	// Step 2: Tính local priorities của Alternatives cho mỗi Criterion
	localPriorities := make(map[string]map[string]float64)

	for _, criterion := range ahpInput.Criteria {
		comparisons := ahpInput.AlternativeComparisons[criterion.ID]

		altMatrix := m.buildComparisonMatrixForAlternatives(
			ahpInput.Alternatives,
			comparisons,
		)

		altPriorities, _, err := m.calculatePriorities(altMatrix)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate priorities for criterion %s: %w",
				criterion.ID, err)
		}

		localPriorities[criterion.ID] = altPriorities
	}

	// Step 3: Tính Global Priorities (tổng hợp)
	// Global Priority = Σ(Criteria Weight × Local Priority)
	globalPriorities := make(map[string]float64)

	for _, alternative := range ahpInput.Alternatives {
		globalPriority := 0.0

		for _, criterion := range ahpInput.Criteria {
			criteriaWeight := criteriaWeights[criterion.ID]
			localPriority := localPriorities[criterion.ID][alternative.ID]

			globalPriority += criteriaWeight * localPriority
		}

		globalPriorities[alternative.ID] = globalPriority
	}

	// Step 4: Tạo ranking
	ranking := m.createRanking(ahpInput.Alternatives, globalPriorities)

	// Step 5: Build output
	output := dto.ToAHPOutput(
		globalPriorities,
		criteriaWeights,
		localPriorities,
		cr,
		ranking,
	)

	return output, nil
}

// buildComparisonMatrixForCriteria xây dựng ma trận so sánh từ criteria
func (m *AHPModel) buildComparisonMatrixForCriteria(
	criteria []domain.Criteria,
	comparisons []domain.PairwiseComparison,
) *domain.ComparisonMatrix {

	ids := make([]string, len(criteria))
	for i, c := range criteria {
		ids[i] = c.ID
	}

	return m.buildMatrix(ids, comparisons)
}

// buildComparisonMatrixForAlternatives xây dựng ma trận so sánh từ alternatives
func (m *AHPModel) buildComparisonMatrixForAlternatives(
	alternatives []domain.Alternative,
	comparisons []domain.PairwiseComparison,
) *domain.ComparisonMatrix {

	ids := make([]string, len(alternatives))
	for i, a := range alternatives {
		ids[i] = a.ID
	}

	return m.buildMatrix(ids, comparisons)
}

// buildMatrix helper function to build comparison matrix
func (m *AHPModel) buildMatrix(
	ids []string,
	comparisons []domain.PairwiseComparison,
) *domain.ComparisonMatrix {

	n := len(ids)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		matrix[i][i] = 1.0 // Diagonal = 1
	}

	// Tạo map để lookup index nhanh
	idToIndex := make(map[string]int)
	for i, id := range ids {
		idToIndex[id] = i
	}

	// Fill ma trận từ comparisons
	for _, comp := range comparisons {
		i := idToIndex[comp.ElementA]
		j := idToIndex[comp.ElementB]

		matrix[i][j] = comp.Value
		matrix[j][i] = 1.0 / comp.Value // Reciprocal
	}

	return &domain.ComparisonMatrix{
		Size:     n,
		Matrix:   matrix,
		Elements: ids,
	}
}

// calculatePriorities tính priority vector và consistency ratio
func (m *AHPModel) calculatePriorities(
	cm *domain.ComparisonMatrix,
) (priorities map[string]float64, consistencyRatio float64, err error) {

	// Method: Normalized Column Average (đơn giản và đủ accurate)
	// Bước 1: Normalize ma trận
	normalized := make([][]float64, cm.Size)
	for i := range normalized {
		normalized[i] = make([]float64, cm.Size)
	}

	// Tính tổng mỗi cột
	for j := 0; j < cm.Size; j++ {
		colSum := 0.0
		for i := 0; i < cm.Size; i++ {
			colSum += cm.Matrix[i][j]
		}

		// Normalize cột
		for i := 0; i < cm.Size; i++ {
			if colSum > 0 {
				normalized[i][j] = cm.Matrix[i][j] / colSum
			}
		}
	}

	// Bước 2: Tính trung bình mỗi hàng (priority vector)
	priorityVector := make([]float64, cm.Size)
	for i := 0; i < cm.Size; i++ {
		rowSum := 0.0
		for j := 0; j < cm.Size; j++ {
			rowSum += normalized[i][j]
		}
		priorityVector[i] = rowSum / float64(cm.Size)
	}

	// Bước 3: Tính Consistency Ratio
	cr := m.calculateConsistencyRatio(cm.Matrix, priorityVector)

	// Convert vector thành map
	priorities = make(map[string]float64)
	for i, priority := range priorityVector {
		priorities[cm.Elements[i]] = priority
	}

	return priorities, cr, nil
}

// calculateConsistencyRatio tính CR để check consistency
func (m *AHPModel) calculateConsistencyRatio(
	matrix [][]float64,
	priorityVector []float64,
) float64 {
	n := len(matrix)

	// Bước 1: Tính λ_max (maximum eigenvalue approximation)
	// λ_max ≈ (1/n) × Σ[(Aw)_i / w_i]

	// Tính Aw (matrix × vector)
	aw := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			aw[i] += matrix[i][j] * priorityVector[j]
		}
	}

	// Tính λ_max
	lambdaMax := 0.0
	for i := 0; i < n; i++ {
		if priorityVector[i] > 0 {
			lambdaMax += aw[i] / priorityVector[i]
		}
	}
	lambdaMax /= float64(n)

	// Bước 2: Tính CI (Consistency Index)
	ci := (lambdaMax - float64(n)) / float64(n-1)

	// Bước 3: Lấy RI (Random Index) từ bảng chuẩn
	ri := m.getRandomIndex(n)

	// Bước 4: Tính CR
	if ri == 0 {
		return 0 // Avoid division by zero
	}

	cr := ci / ri
	return cr
}

// getRandomIndex trả về RI theo bảng chuẩn của Saaty
func (m *AHPModel) getRandomIndex(n int) float64 {
	// Random Index table (Saaty, 1980)
	riTable := map[int]float64{
		1:  0.00,
		2:  0.00,
		3:  0.58,
		4:  0.90,
		5:  1.12,
		6:  1.24,
		7:  1.32,
		8:  1.41,
		9:  1.45,
		10: 1.49,
	}

	if ri, exists := riTable[n]; exists {
		return ri
	}

	// For n > 10, use approximation
	return 1.49 + 0.01*float64(n-10)
}

// createRanking tạo ranking list sorted theo priority
func (m *AHPModel) createRanking(
	alternatives []domain.Alternative,
	priorities map[string]float64,
) []domain.RankItem {

	ranking := make([]domain.RankItem, 0, len(alternatives))

	for _, alt := range alternatives {
		ranking = append(ranking, domain.RankItem{
			AlternativeID:   alt.ID,
			AlternativeName: alt.Name,
			Priority:        priorities[alt.ID],
		})
	}

	// Sort theo priority giảm dần
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].Priority > ranking[j].Priority
	})

	// Gán rank
	for i := range ranking {
		ranking[i].Rank = i + 1
	}

	return ranking
}
