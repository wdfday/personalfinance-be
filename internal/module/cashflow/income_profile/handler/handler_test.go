package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	authDomain "personalfinancedss/internal/module/identify/auth/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	basePath          = "/income-profiles"
	pathWithID        = "/income-profiles/:id"
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

// MockService is a mock implementation of Service
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateIncomeProfile(ctx context.Context, userID string, req dto.CreateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetIncomeProfile(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetIncomeProfileWithHistory(ctx context.Context, userID string, profileID string) (*domain.IncomeProfile, []*domain.IncomeProfile, error) {
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

func (m *MockService) ListIncomeProfiles(ctx context.Context, userID string, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetActiveIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetArchivedIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) GetRecurringIncomes(ctx context.Context, userID string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) UpdateIncomeProfile(ctx context.Context, userID string, profileID string, req dto.UpdateIncomeProfileRequest) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) VerifyIncomeProfile(ctx context.Context, userID string, profileID string, verified bool) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID, verified)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) UpdateDSSMetadata(ctx context.Context, userID string, profileID string, req dto.UpdateDSSMetadataRequest) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, profileID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockService) ArchiveIncomeProfile(ctx context.Context, userID string, profileID string) error {
	args := m.Called(ctx, userID, profileID)
	return args.Error(0)
}

func (m *MockService) CheckAndArchiveEnded(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockService) DeleteIncomeProfile(ctx context.Context, userID string, profileID string) error {
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
	user := authDomain.AuthUser{
		ID:       userID,
		Username: "test@example.com",
	}
	c.Set(middleware.UserKey, user)
}

func TestHandlerCreateIncomeProfile(t *testing.T) {
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

	// Strip monotonic clock from time.Now() by round-tripping through JSON
	// This ensures the mock expectation matches what the handler receives (JSON-decoded struct)
	body, _ := json.Marshal(req)
	_ = json.Unmarshal(body, &req)

	mockService.On("CreateIncomeProfile", mock.Anything, userID.String(), req).Return(expectedProfile, nil)

	r.POST(basePath, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.createIncomeProfile(c)
	})

	body, _ = json.Marshal(req)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", basePath, bytes.NewBuffer(body))
	request.Header.Set(contentTypeHeader, applicationJSON)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerGetIncomeProfile(t *testing.T) {
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

	r.GET(pathWithID, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", basePath+"/"+profileID.String(), nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerListIncomeProfiles(t *testing.T) {
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

	r.GET(basePath, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.listIncomeProfiles(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", basePath, nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerGetIncomeProfileHistory(t *testing.T) {
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

	r.GET(pathWithID+"/history", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getIncomeProfileHistory(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", basePath+"/"+profileID.String()+"/history", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerGetActiveIncomes(t *testing.T) {
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

	r.GET(basePath+"/active", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getActiveIncomes(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", basePath+"/active", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerLoadIncomeProfile(t *testing.T) {
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

	r.PUT(pathWithID, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.updateIncomeProfile(c)
	})

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", basePath+"/"+profileID.String(), bytes.NewBuffer(body))
	request.Header.Set(contentTypeHeader, applicationJSON)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerVerifyIncomeProfile(t *testing.T) {
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

	r.POST(pathWithID+"/verify", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.verifyIncomeProfile(c)
	})

	body, _ := json.Marshal(verifyReq)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", basePath+"/"+profileID.String()+"/verify", bytes.NewBuffer(body))
	request.Header.Set(contentTypeHeader, applicationJSON)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerArchiveIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	mockService.On("ArchiveIncomeProfile", mock.Anything, userID.String(), profileID.String()).Return(nil)

	r.POST(pathWithID+"/archive", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.archiveIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", basePath+"/"+profileID.String()+"/archive", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerUpdateDSSMetadata(t *testing.T) {
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

	r.POST(pathWithID+"/dss-metadata", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.updateDSSMetadata(c)
	})

	body, _ := json.Marshal(dssReq)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", basePath+"/"+profileID.String()+"/dss-metadata", bytes.NewBuffer(body))
	request.Header.Set(contentTypeHeader, applicationJSON)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerCheckAndArchiveEnded(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	mockService.On("CheckAndArchiveEnded", mock.Anything, userID.String()).Return(3, nil)

	r.POST(basePath+"/check-ended", func(c *gin.Context) {
		setUserContext(c, userID)
		handler.checkAndArchiveEnded(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", basePath+"/check-ended", nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["archived_count"])

	mockService.AssertExpectations(t)
}

func TestHandlerDeleteIncomeProfile(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()
	profileID := uuid.New()

	mockService.On("DeleteIncomeProfile", mock.Anything, userID.String(), profileID.String()).Return(nil)

	r.DELETE(pathWithID, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.deleteIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", basePath+"/"+profileID.String(), nil)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerCreateIncomeProfileInvalidRequest(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	r.POST(basePath, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.createIncomeProfile(c)
	})

	// Invalid JSON
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", basePath, bytes.NewBufferString("invalid json"))
	request.Header.Set(contentTypeHeader, applicationJSON)

	r.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandlerGetIncomeProfileInvalidID(t *testing.T) {
	r, mockService := setupRouter()
	handler := NewHandler(mockService)

	userID := uuid.New()

	r.GET(pathWithID, func(c *gin.Context) {
		setUserContext(c, userID)
		handler.getIncomeProfile(c)
	})

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", basePath+"/invalid-uuid", nil)

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
