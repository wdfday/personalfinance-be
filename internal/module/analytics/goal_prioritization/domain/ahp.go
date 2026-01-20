package domain

import "time"

// Criteria định nghĩa một tiêu chí đánh giá
type Criteria struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"` // Tính toán được từ AHP
}

// Alternative là một lựa chọn (trong case này là Goal)
type Alternative struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	TargetAmount    float64            `json:"target_amount"`
	CurrentAmount   float64            `json:"current_amount"`
	Deadline        time.Time          `json:"deadline"`
	Description     string             `json:"description"`
	Priority        float64            `json:"priority"`                   // Output từ AHP
	LocalPriorities map[string]float64 `json:"local_priorities,omitempty"` // Priority theo từng criterion
}

// PairwiseComparison đại diện một so sánh cặp
type PairwiseComparison struct {
	ElementA string  `json:"element_a"` // ID của phần tử A
	ElementB string  `json:"element_b"` // ID của phần tử B
	Value    float64 `json:"value"`     // A quan trọng hơn B bao nhiêu (1-9)

	// Value interpretation:
	// 1.0: A và B ngang nhau
	// 3.0: A quan trọng hơn B vừa phải
	// 5.0: A quan trọng hơn B rõ rệt
	// 0.333: A kém quan trọng hơn B vừa phải (= 1/3)
	// 0.2: A kém quan trọng hơn B rõ rệt (= 1/5)

}

// ComparisonMatrix là ma trận so sánh nxn
type ComparisonMatrix struct {
	Size     int         `json:"size"`     // n
	Matrix   [][]float64 `json:"matrix"`   // Ma trận nxn
	Elements []string    `json:"elements"` // IDs các phần tử
}

// RankItem represents ranking of an alternative
type RankItem struct {
	AlternativeID   string  `json:"alternative_id"`
	AlternativeName string  `json:"alternative_name"`
	Priority        float64 `json:"priority"`
	Rank            int     `json:"rank"`
}
