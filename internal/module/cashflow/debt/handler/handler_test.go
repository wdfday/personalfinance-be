package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"personalfinancedss/internal/module/cashflow/debt/domain"
	"personalfinancedss/internal/module/cashflow/debt/dto"
	"personalfinancedss/internal/module/cashflow/debt/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockService
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateDebt(ctx context.Context, debt *domain.Debt) error {
	args := m.Called(ctx, debt)
	return args.Error(0)
}

func (m *MockService) GetDebtByID(ctx context.Context, debtID uuid.UUID) (*domain.Debt, error) {
	return nil, nil
}
func (m *MockService) GetUserDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	return nil, nil
}
func (m *MockService) GetActiveDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	return nil, nil
}
func (m *MockService) GetDebtsByType(ctx context.Context, userID uuid.UUID, debtType domain.DebtType) ([]domain.Debt, error) {
	return nil, nil
}
func (m *MockService) GetPaidOffDebts(ctx context.Context, userID uuid.UUID) ([]domain.Debt, error) {
	return nil, nil
}
func (m *MockService) GetDebtSummary(ctx context.Context, userID uuid.UUID) (*service.DebtSummary, error) {
	return nil, nil
}
func (m *MockService) UpdateDebt(ctx context.Context, debt *domain.Debt) error {
	return nil
}
func (m *MockService) CalculateProgress(ctx context.Context, debtID uuid.UUID) error {
	return nil
}
func (m *MockService) CheckOverdueDebts(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *MockService) DeleteDebt(ctx context.Context, debtID uuid.UUID) error {
	return nil
}
func (m *MockService) AddPayment(ctx context.Context, debtID uuid.UUID, amount float64) (*domain.Debt, error) {
	return nil, nil
}
func (m *MockService) MarkAsPaidOff(ctx context.Context, debtID uuid.UUID) error {
	return nil
}

func TestHandler_CreateDebt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	logger := zap.NewNop()
	h := NewHandler(mockSvc, logger)

	router := gin.New()
	// Bypass auth middleware by setting user_id directly
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		c.Next()
	})

	router.POST("/api/v1/debts", h.CreateDebt)

	t.Run("Success", func(t *testing.T) {
		reqBody := dto.CreateDebtRequest{
			Name:            "Test Debt",
			Type:            domain.DebtTypeCreditCard,
			Behavior:        domain.DebtBehaviorRevolving, // Verify this field
			PrincipalAmount: 1000,
			CurrentBalance:  500,
			Currency:        "VND",
			StartDate:       time.Now(),
		}
		jsonBody, _ := json.Marshal(reqBody)

		mockSvc.On("CreateDebt", mock.Anything, mock.MatchedBy(func(d *domain.Debt) bool {
			return d.Behavior == domain.DebtBehaviorRevolving && d.Type == domain.DebtTypeCreditCard
		})).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
