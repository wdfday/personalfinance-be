package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"personalfinancedss/internal/module/analytics/goal_prioritization/domain"
	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockService is a mock implementation of service.Service for testing
type mockService struct {
	executeFunc                      func(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error)
	executeAutoScoringFunc           func(ctx context.Context, input *dto.AutoScoringInput) (*dto.AHPOutput, error)
	executeDirectRatingFunc          func(ctx context.Context, input *dto.DirectRatingInput) (*dto.AHPOutput, error)
	getAutoScoresFunc                func(ctx context.Context, input *dto.AutoScoresRequest) (*dto.AutoScoresResponse, error)
	executeDirectRatingWithOverrides func(ctx context.Context, input *dto.DirectRatingWithOverridesInput) (*dto.AHPOutput, error)
}

func (m *mockService) ExecuteAHP(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ExecuteAutoScoring(ctx context.Context, input *dto.AutoScoringInput) (*dto.AHPOutput, error) {
	if m.executeAutoScoringFunc != nil {
		return m.executeAutoScoringFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ExecuteDirectRating(ctx context.Context, input *dto.DirectRatingInput) (*dto.AHPOutput, error) {
	if m.executeDirectRatingFunc != nil {
		return m.executeDirectRatingFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetAutoScores(ctx context.Context, input *dto.AutoScoresRequest) (*dto.AutoScoresResponse, error) {
	if m.getAutoScoresFunc != nil {
		return m.getAutoScoresFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ExecuteDirectRatingWithOverrides(ctx context.Context, input *dto.DirectRatingWithOverridesInput) (*dto.AHPOutput, error) {
	if m.executeDirectRatingWithOverrides != nil {
		return m.executeDirectRatingWithOverrides(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewHandler(t *testing.T) {
	logger := zap.NewNop()
	mockSvc := &mockService{}

	handler := NewHandler(mockSvc, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockSvc, handler.service)
	assert.Equal(t, logger, handler.logger)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc, logger)

	router := setupTestRouter()
	handler.RegisterRoutes(router)

	// Check that the route is registered
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/api/v1/analytics/goal-prioritization" && route.Method == "POST" {
			found = true
			break
		}
	}

	assert.True(t, found, "Route /api/v1/analytics/goal-prioritization should be registered")
}

func TestHandler_ExecuteAHP_Success(t *testing.T) {
	logger := zap.NewNop()

	expectedOutput := &dto.AHPOutput{
		AlternativePriorities: map[string]float64{
			"goal1": 0.6,
			"goal2": 0.4,
		},
		CriteriaWeights: map[string]float64{
			"c1": 0.7,
			"c2": 0.3,
		},
		LocalPriorities: map[string]map[string]float64{
			"c1": {"goal1": 0.6, "goal2": 0.4},
		},
		ConsistencyRatio: 0.05,
		IsConsistent:     true,
		Ranking: []domain.RankItem{
			{AlternativeID: "goal1", AlternativeName: "Goal 1", Priority: 0.6, Rank: 1},
			{AlternativeID: "goal2", AlternativeName: "Goal 2", Priority: 0.4, Rank: 2},
		},
	}

	mockSvc := &mockService{
		executeFunc: func(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
			assert.Equal(t, "user123", input.UserID)
			return expectedOutput, nil
		},
	}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	requestBody := dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Criteria 1"},
			{ID: "c2", Name: "Criteria 2"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "c1", ElementB: "c2", Value: 3.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "goal1", Name: "Goal 1"},
			{ID: "goal2", Name: "Goal 2"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"c1": {{ElementA: "goal1", ElementB: "goal2", Value: 2.0}},
			"c2": {{ElementA: "goal1", ElementB: "goal2", Value: 3.0}},
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response shared.SuccessResponse[dto.AHPOutput]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedOutput.AlternativePriorities, response.Data.AlternativePriorities)
	assert.Equal(t, expectedOutput.CriteriaWeights, response.Data.CriteriaWeights)
	assert.Equal(t, expectedOutput.ConsistencyRatio, response.Data.ConsistencyRatio)
	assert.Equal(t, expectedOutput.IsConsistent, response.Data.IsConsistent)
}

func TestHandler_ExecuteAHP_WithUserIDFromContext(t *testing.T) {
	logger := zap.NewNop()

	userUUID := uuid.New()
	expectedUserID := userUUID.String()

	mockSvc := &mockService{
		executeFunc: func(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
			// Verify that user ID from context overrides the one in request
			assert.Equal(t, expectedUserID, input.UserID)
			return &dto.AHPOutput{
				AlternativePriorities: map[string]float64{"goal1": 1.0},
				CriteriaWeights:       map[string]float64{"c1": 1.0},
				LocalPriorities:       map[string]map[string]float64{},
				ConsistencyRatio:      0.0,
				IsConsistent:          true,
				Ranking:               []domain.RankItem{},
			}, nil
		},
	}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()

	// Register route with middleware that sets user_id in context
	router.POST("/api/v1/analytics/goal-prioritization", func(c *gin.Context) {
		c.Set("user_id", userUUID)
		handler.ExecuteAHP(c)
	})

	requestBody := dto.AHPInput{
		UserID: "original-user-id", // This should be overridden
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Criteria 1"},
			{ID: "c2", Name: "Criteria 2"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "c1", ElementB: "c2", Value: 2.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "goal1", Name: "Goal 1"},
			{ID: "goal2", Name: "Goal 2"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"c1": {{ElementA: "goal1", ElementB: "goal2", Value: 2.0}},
			"c2": {{ElementA: "goal1", ElementB: "goal2", Value: 2.0}},
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_ExecuteAHP_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	mockSvc := &mockService{}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestHandler_ExecuteAHP_ValidationError(t *testing.T) {
	logger := zap.NewNop()

	mockSvc := &mockService{
		executeFunc: func(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
			return nil, errors.New("validation failed: at least 2 criteria required")
		},
	}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	// Send request with invalid data (only 1 criterion)
	requestBody := dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Criteria 1"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{},
		Alternatives: []domain.Alternative{
			{ID: "goal1", Name: "Goal 1"},
			{ID: "goal2", Name: "Goal 2"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Gin's binding validation triggers before service validation (400 instead of 500)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	// Gin validation message contains "Field validation for 'Criteria' failed"
	assert.Contains(t, response["message"].(string), "Criteria")
}

func TestHandler_ExecuteAHP_ServiceError(t *testing.T) {
	logger := zap.NewNop()

	mockSvc := &mockService{
		executeFunc: func(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
			return nil, errors.New("internal service error")
		},
	}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	requestBody := dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "c1", Name: "Criteria 1"},
			{ID: "c2", Name: "Criteria 2"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "c1", ElementB: "c2", Value: 2.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "goal1", Name: "Goal 1"},
			{ID: "goal2", Name: "Goal 2"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"c1": {{ElementA: "goal1", ElementB: "goal2", Value: 2.0}},
			"c2": {{ElementA: "goal1", ElementB: "goal2", Value: 2.0}},
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Equal(t, "Internal server error", response["message"])
}

func TestHandler_ExecuteAHP_MissingRequiredFields(t *testing.T) {
	logger := zap.NewNop()
	mockSvc := &mockService{}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name        string
		requestBody interface{}
	}{
		{
			name: "Missing criteria",
			requestBody: map[string]interface{}{
				"user_id":                 "user123",
				"criteria_comparisons":    []interface{}{},
				"alternatives":            []interface{}{},
				"alternative_comparisons": map[string]interface{}{},
			},
		},
		{
			name: "Missing alternatives",
			requestBody: map[string]interface{}{
				"user_id":                 "user123",
				"criteria":                []interface{}{},
				"criteria_comparisons":    []interface{}{},
				"alternative_comparisons": map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
		})
	}
}

func TestHandler_ExecuteAHP_CompleteScenario(t *testing.T) {
	logger := zap.NewNop()

	mockSvc := &mockService{
		executeFunc: func(ctx context.Context, input *dto.AHPInput) (*dto.AHPOutput, error) {
			// Simulate realistic output
			return &dto.AHPOutput{
				AlternativePriorities: map[string]float64{
					"emergency_fund": 0.45,
					"house_payment":  0.35,
					"retirement":     0.20,
				},
				CriteriaWeights: map[string]float64{
					"importance": 0.65,
					"urgency":    0.35,
				},
				LocalPriorities: map[string]map[string]float64{
					"importance": {
						"emergency_fund": 0.5,
						"house_payment":  0.3,
						"retirement":     0.2,
					},
					"urgency": {
						"emergency_fund": 0.6,
						"house_payment":  0.3,
						"retirement":     0.1,
					},
				},
				ConsistencyRatio: 0.08,
				IsConsistent:     true,
				Ranking: []domain.RankItem{
					{AlternativeID: "emergency_fund", AlternativeName: "Emergency Fund", Priority: 0.45, Rank: 1},
					{AlternativeID: "house_payment", AlternativeName: "House Down Payment", Priority: 0.35, Rank: 2},
					{AlternativeID: "retirement", AlternativeName: "Retirement Savings", Priority: 0.20, Rank: 3},
				},
			}, nil
		},
	}

	handler := NewHandler(mockSvc, logger)
	router := setupTestRouter()
	handler.RegisterRoutes(router)

	requestBody := dto.AHPInput{
		UserID: "user123",
		Criteria: []domain.Criteria{
			{ID: "importance", Name: "Importance", Description: "How important is this goal?"},
			{ID: "urgency", Name: "Urgency", Description: "How urgent is this goal?"},
		},
		CriteriaComparisons: []domain.PairwiseComparison{
			{ElementA: "importance", ElementB: "urgency", Value: 2.0},
		},
		Alternatives: []domain.Alternative{
			{ID: "emergency_fund", Name: "Emergency Fund"},
			{ID: "house_payment", Name: "House Down Payment"},
			{ID: "retirement", Name: "Retirement Savings"},
		},
		AlternativeComparisons: map[string][]domain.PairwiseComparison{
			"importance": {
				{ElementA: "emergency_fund", ElementB: "house_payment", Value: 2.0},
				{ElementA: "emergency_fund", ElementB: "retirement", Value: 3.0},
				{ElementA: "house_payment", ElementB: "retirement", Value: 2.0},
			},
			"urgency": {
				{ElementA: "emergency_fund", ElementB: "house_payment", Value: 3.0},
				{ElementA: "emergency_fund", ElementB: "retirement", Value: 5.0},
				{ElementA: "house_payment", ElementB: "retirement", Value: 2.0},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/goal-prioritization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response shared.SuccessResponse[dto.AHPOutput]
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify the response structure and content
	assert.NotNil(t, response.Data.AlternativePriorities)
	assert.Len(t, response.Data.AlternativePriorities, 3)

	assert.NotNil(t, response.Data.CriteriaWeights)
	assert.Len(t, response.Data.CriteriaWeights, 2)

	assert.NotNil(t, response.Data.Ranking)
	assert.Len(t, response.Data.Ranking, 3)

	assert.True(t, response.Data.IsConsistent)
	assert.Less(t, response.Data.ConsistencyRatio, 0.1)

	// Verify ranking order
	assert.Equal(t, 1, response.Data.Ranking[0].Rank)
	assert.Equal(t, 2, response.Data.Ranking[1].Rank)
	assert.Equal(t, 3, response.Data.Ranking[2].Rank)
}
