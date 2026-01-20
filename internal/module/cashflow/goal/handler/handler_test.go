package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/dto"
	authDomain "personalfinancedss/internal/module/identify/auth/domain"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockService is a mock implementation of service.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateGoal(ctx context.Context, goal *domain.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockService) GetGoalByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Goal), args.Error(1)
}

func (m *MockService) GetUserGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Goal), args.Error(1)
}

func (m *MockService) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockService) DeleteGoal(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Add these to satisfy the interface for mocking only what we test
func (m *MockService) GetActiveGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Goal), args.Error(1)
}
func (m *MockService) GetArchivedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockService) GetGoalsByCategory(ctx context.Context, userID uuid.UUID, category domain.GoalCategory) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockService) GetCompletedGoals(ctx context.Context, userID uuid.UUID) ([]domain.Goal, error) {
	return nil, nil
}
func (m *MockService) GetGoalSummary(ctx context.Context, userID uuid.UUID) (*domain.GoalSummary, error) {
	return nil, nil
}
func (m *MockService) GetGoalAnalytics(ctx context.Context, goalID uuid.UUID) (*domain.GoalAnalytics, error) {
	return nil, nil
}
func (m *MockService) GetGoalProgress(ctx context.Context, goalID uuid.UUID) (*domain.GoalProgress, error) {
	return nil, nil
}
func (m *MockService) CalculateProgress(ctx context.Context, goalID uuid.UUID) error { return nil }
func (m *MockService) CheckOverdueGoals(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockService) MarkAsCompleted(ctx context.Context, goalID uuid.UUID) error   { return nil }
func (m *MockService) ArchiveGoal(ctx context.Context, goalID uuid.UUID) error       { return nil }
func (m *MockService) UnarchiveGoal(ctx context.Context, goalID uuid.UUID) error     { return nil }
func (m *MockService) AddContribution(ctx context.Context, goalID uuid.UUID, amount float64, source *string, method string) (*domain.Goal, error) {
	args := m.Called(ctx, goalID, amount, source, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Goal), args.Error(1)
}
func (m *MockService) WithdrawContribution(ctx context.Context, goalID uuid.UUID, amount float64, note *string, reversingID *uuid.UUID) (*domain.Goal, error) {
	return nil, nil
}
func (m *MockService) GetContributions(ctx context.Context, goalID uuid.UUID) ([]domain.GoalContribution, error) {
	return nil, nil
}
func (m *MockService) GetGoalNetContributions(ctx context.Context, goalID uuid.UUID) (float64, error) {
	return 0, nil
}

func (m *MockService) GetMonthContributions(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) ([]domain.GoalContribution, error) {
	return nil, nil
}

func (m *MockService) GetMonthSummary(ctx context.Context, goalID uuid.UUID, startDate, endDate time.Time) (*dto.GoalMonthlySummary, error) {
	return nil, nil
}

func (m *MockService) GetAllTimeSummary(ctx context.Context, goalID uuid.UUID) (*dto.GoalAllTimeSummary, error) {
	return nil, nil
}

func setupRouter(svc *MockService) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, zap.NewNop())
	return r, h
}

func TestHandler_CreateGoal(t *testing.T) {
	svc := &MockService{}
	r, h := setupRouter(svc)

	userID := uuid.New()
	validGoal := dto.CreateGoalRequest{
		Name:         "Travel Fund",
		TargetAmount: 5000000,
		Category:     domain.GoalCategoryTravel,
		Priority:     domain.GoalPriorityMedium,
		StartDate:    time.Now(),
		Behavior:     domain.GoalBehaviorFlexible,
		Currency:     "VND",
		AccountID:    uuid.New(),
	}

	svc.On("CreateGoal", mock.Anything, mock.MatchedBy(func(g *domain.Goal) bool {
		return g.Name == validGoal.Name && g.TargetAmount == validGoal.TargetAmount && g.UserID == userID
	})).Return(nil)

	r.POST("/goals", func(c *gin.Context) {
		c.Set(middleware.UserKey, authDomain.AuthUser{ID: userID})
		h.CreateGoal(c)
	})

	body, _ := json.Marshal(validGoal)
	req, _ := http.NewRequest("POST", "/goals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Response body: %s", w.Body.String())
	svc.AssertExpectations(t)
}

func TestHandler_GetUserGoals(t *testing.T) {
	svc := &MockService{}
	r, h := setupRouter(svc)
	userID := uuid.New()

	goals := []domain.Goal{
		{ID: uuid.New(), Name: "Goal 1", UserID: userID},
		{ID: uuid.New(), Name: "Goal 2", UserID: userID},
	}

	svc.On("GetUserGoals", mock.Anything, userID).Return(goals, nil)

	r.GET("/goals", func(c *gin.Context) {
		c.Set(middleware.UserKey, authDomain.AuthUser{ID: userID})
		h.GetUserGoals(c)
	})

	req, _ := http.NewRequest("GET", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Data []dto.GoalResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 2)
}

func TestHandler_AddContribution(t *testing.T) {
	svc := &MockService{}
	r, h := setupRouter(svc)
	userID := uuid.New()
	goalID := uuid.New()

	reqPayload := dto.AddContributionRequest{
		Amount: 1000,
		Note:   nil,
	}

	updatedGoal := &domain.Goal{
		ID:            goalID,
		UserID:        userID,
		CurrentAmount: 1000,
	}

	svc.On("AddContribution", mock.Anything, goalID, 1000.0, mock.Anything, "manual").Return(updatedGoal, nil)

	r.POST("/goals/:id/contribute", func(c *gin.Context) {
		c.Set(middleware.UserKey, authDomain.AuthUser{ID: userID})
		h.AddContribution(c)
	})

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/goals/"+goalID.String()+"/contribute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
