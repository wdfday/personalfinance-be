package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/module/identify/user/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of Service
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateIncomeProfile(ctx any, userID string, req dto.CreateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetIncomeProfile(ctx any, userID string, profileID string) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetIncomeProfileWithHistory(ctx any, userID string, profileID string) (*domain.IncomeProfile, []*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var history []*domain.IncomeProfile
	if args.Get(1) != nil {
		history = args.Get(1).([]*domain.IncomeProfile)
	}
	return args.Get(0).(*domain.IncomeProfile), history, args.Error(2)
}

func (m *MockService) ListIncomeProfiles(ctx any, userID string, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetActiveIncomes(ctx any, userID string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetArchivedIncomes(ctx any, userID string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetRecurringIncomes(ctx any, userID string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) UpdateIncomeProfile(ctx any, userID string, profileID string, req dto.UpdateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) VerifyIncomeProfile(ctx any, userID string, profileID string, verified bool) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID, verified)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) UpdateDSSMetadata(ctx any, userID string, profileID string, req dto.UpdateDSSMetadataRequest) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) ArchiveIncomeProfile(ctx any, userID string, profileID string) error {
	args := m.Called(ctx, userID, profileID)
	return args.Error(0)
}

func (m *MockService) CheckAndArchiveEnded(ctx any, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockService) DeleteIncomeProfile(ctx any, userID string, profileID string) error {
	args := m.Called(ctx, userID, profileID)
	return args.Error(0)
}

func setupRouter() (*gin.Engine, *MockService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockService := new(MockService)
	return r, mockService
}

func setUserContext(c *gin.Context, userID uuid.UUID) {
	user := &entity.User{
		ID:    userID,
		Email: "test@example.com",
	}
	c.Set(middleware.CurrentUserKey, user)
}

func TestHandler_CreateIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	req := dto.CreateIncomeProfileRequest{
		Source:    "Salary",
		Amount:    50000000,
		Frequency: "monthly",
		StartDate: time.Now(),
	}

	expectedProfile := &domain.IncomeProfile{
		ID:     uuid.New(),
		UserID: userID,
		Source: req.Source,
		Amount: req.Amount,
	}

	mockService.On("CreateIncomeProfile", mock.Anything, userID.String(), req).Return(expectedProfile, nil)

	r.POST("/income-profiles", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.createIncomeProfile(c)
	})

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/income-profiles", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_GetIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	expectedProfile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
		Source: "Salary",
		Amount: 50000000,
	}

	mockService.On("GetIncomeProfile", mock.Anything, userID.String(), profileID.String()).Return(expectedProfile, nil)

	r.GET("/income-profiles/:id", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/income-profiles/"+profileID.String(), nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_ListIncomeProfiles(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	expectedProfiles := []*domain.IncomeProfile{
		{
			ID:     uuid.New(),
			UserID: userID,
			Status: domain.IncomeStatusActive,
		},
	}

	mockService.On("ListIncomeProfiles", mock.Anything, userID.String(), mock.AnythingOfType("dto.ListIncomeProfilesQuery")).Return(expectedProfiles, nil)

	r.GET("/income-profiles", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.listIncomeProfiles(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/income-profiles", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_GetIncomeProfileHistory(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	currentProfile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
		Source: "Salary v2",
	}

	historyProfiles := []*domain.IncomeProfile{
		{
			ID:     uuid.New(),
			UserID: userID,
			Source: "Salary v1",
		},
	}

	mockService.On("GetIncomeProfileWithHistory", mock.Anything, userID.String(), profileID.String()).Return(currentProfile, historyProfiles, nil)

	r.GET("/income-profiles/:id/history", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getIncomeProfileHistory(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/income-profiles/"+profileID.String()+"/history", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_GetActiveIncomes(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	expectedProfiles := []*domain.IncomeProfile{
		{
			ID:     uuid.New(),
			UserID: userID,
			Status: domain.IncomeStatusActive,
		},
	}

	mockService.On("GetActiveIncomes", mock.Anything, userID.String()).Return(expectedProfiles, nil)

	r.GET("/income-profiles/active", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getActiveIncomes(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/income-profiles/active", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()
	newVersionID := uuid.New()

	req := dto.UpdateIncomeProfileRequest{
		Amount: floatPtr(60000000),
	}

	newVersion := &domain.IncomeProfile{
		ID:                newVersionID,
		UserID:            userID,
		Amount:            60000000,
		PreviousVersionID: &profileID,
	}

	mockService.On("UpdateIncomeProfile", mock.Anything, userID.String(), profileID.String(), req).Return(newVersion, nil)

	r.PUT("/income-profiles/:id", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.updateIncomeProfile(c)
	})

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", "/income-profiles/"+profileID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_VerifyIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	verifyReq := dto.VerifyIncomeRequest{
		Verified: true,
	}

	verifiedProfile := &domain.IncomeProfile{
		ID:         profileID,
		UserID:     userID,
		IsVerified: true,
	}

	mockService.On("VerifyIncomeProfile", mock.Anything, userID.String(), profileID.String(), true).Return(verifiedProfile, nil)

	r.POST("/income-profiles/:id/verify", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.verifyIncomeProfile(c)
	})

	body, _ := json.Marshal(verifyReq)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/income-profiles/"+profileID.String()+"/verify", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_ArchiveIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	mockService.On("ArchiveIncomeProfile", mock.Anything, userID.String(), profileID.String()).Return(nil)

	r.POST("/income-profiles/:id/archive", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.archiveIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/income-profiles/"+profileID.String()+"/archive", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateDSSMetadata(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	dssReq := dto.UpdateDSSMetadataRequest{
		StabilityScore: floatPtr(0.95),
		RiskLevel:      stringPtr("low"),
	}

	updatedProfile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
	}

	mockService.On("UpdateDSSMetadata", mock.Anything, userID.String(), profileID.String(), dssReq).Return(updatedProfile, nil)

	r.POST("/income-profiles/:id/dss-metadata", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.updateDSSMetadata(c)
	})

	body, _ := json.Marshal(dssReq)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/income-profiles/"+profileID.String()+"/dss-metadata", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_CheckAndArchiveEnded(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	mockService.On("CheckAndArchiveEnded", mock.Anything, userID.String()).Return(3, nil)

	r.POST("/income-profiles/check-ended", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.checkAndArchiveEnded(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/income-profiles/check-ended", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["archived_count"])

	mockService.AssertExpectations(t)
}

func TestHandler_DeleteIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	mockService.On("DeleteIncomeProfile", mock.Anything, userID.String(), profileID.String()).Return(nil)

	r.DELETE("/income-profiles/:id", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.deleteIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/income-profiles/"+profileID.String(), nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_CreateIncomeProfile_InvalidRequest(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	r.POST("/income-profiles", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.createIncomeProfile(c)
	})

	// Invalid JSON
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/income-profiles", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_GetIncomeProfile_InvalidID(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	r.GET("/income-profiles/:id", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/income-profiles/invalid-uuid", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
