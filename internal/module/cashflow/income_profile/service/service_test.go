package service

import (
	"context"
	"testing"
	"time"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, ip *domain.IncomeProfile) error {
	args := m.Called(ctx, ip)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetActiveByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetArchivedByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetByStatus(ctx context.Context, userID uuid.UUID, status domain.IncomeStatus) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetVersionHistory(ctx context.Context, profileID uuid.UUID) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetLatestVersion(ctx context.Context, profileID uuid.UUID) (*domain.IncomeProfile, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetBySource(ctx context.Context, userID uuid.UUID, source string) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, source)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) GetRecurringByUser(ctx context.Context, userID uuid.UUID) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListIncomeProfilesQuery) ([]*domain.IncomeProfile, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.IncomeProfile), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, ip *domain.IncomeProfile) error {
	args := m.Called(ctx, ip)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Archive(ctx context.Context, id uuid.UUID, archivedBy uuid.UUID) error {
	args := m.Called(ctx, id, archivedBy)
	return args.Error(0)
}

func setupService() (*incomeProfileService, *MockRepository) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	service := &incomeProfileService{
		repo:   mockRepo,
		logger: logger,
	}
	return service, mockRepo
}

func TestService_CreateIncomeProfile(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	req := dto.CreateIncomeProfileRequest{
		Source:     "Salary - Company X",
		Amount:     50000000,
		Currency:   "VND",
		Frequency:  "monthly",
		StartDate:  time.Now(),
		BaseSalary: floatPtr(40000000),
		Bonus:      floatPtr(10000000),
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.IncomeProfile")).Return(nil)

	ip, err := service.CreateIncomeProfile(ctx, userID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, req.Source, ip.Source)
	assert.Equal(t, req.Amount, ip.Amount)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateIncomeProfile_InvalidUserID(t *testing.T) {
	service, _ := setupService()
	ctx := context.Background()

	req := dto.CreateIncomeProfileRequest{
		Source:    "Salary",
		Amount:    50000000,
		Frequency: "monthly",
		StartDate: time.Now(),
	}

	_, err := service.CreateIncomeProfile(ctx, "invalid-uuid", req)

	assert.Error(t, err)
}

func TestService_GetIncomeProfile(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	expectedProfile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
		Source: "Salary",
		Amount: 50000000,
	}

	mockRepo.On("GetByID", ctx, profileID).Return(expectedProfile, nil)

	ip, err := service.GetIncomeProfile(ctx, userID.String(), profileID.String())

	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, profileID, ip.ID)
	mockRepo.AssertExpectations(t)
}

func TestService_GetIncomeProfile_NotFound(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	mockRepo.On("GetByID", ctx, profileID).Return(nil, shared.ErrNotFound)

	_, err := service.GetIncomeProfile(ctx, userID.String(), profileID.String())

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetIncomeProfile_WrongUser(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	profileID := uuid.New()

	profile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: otherUserID, // Different user
		Source: "Salary",
	}

	mockRepo.On("GetByID", ctx, profileID).Return(profile, nil)

	_, err := service.GetIncomeProfile(ctx, userID.String(), profileID.String())

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetIncomeProfileWithHistory(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
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

	mockRepo.On("GetByID", ctx, profileID).Return(currentProfile, nil)
	mockRepo.On("GetVersionHistory", ctx, profileID).Return(historyProfiles, nil)

	current, history, err := service.GetIncomeProfileWithHistory(ctx, userID.String(), profileID.String())

	assert.NoError(t, err)
	assert.NotNil(t, current)
	assert.Len(t, history, 1)
	mockRepo.AssertExpectations(t)
}

func TestService_ListIncomeProfiles(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	query := dto.ListIncomeProfilesQuery{
		Status: stringPtr("active"),
	}

	expectedProfiles := []*domain.IncomeProfile{
		{
			ID:     uuid.New(),
			UserID: userID,
			Status: domain.IncomeStatusActive,
		},
	}

	mockRepo.On("List", ctx, userID, query).Return(expectedProfiles, nil)

	profiles, err := service.ListIncomeProfiles(ctx, userID.String(), query)

	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateIncomeProfile(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	existingProfile := &domain.IncomeProfile{
		ID:         profileID,
		UserID:     userID,
		Source:     "Salary v1",
		Amount:     50000000,
		Frequency:  "monthly",
		StartDate:  time.Now().Add(-30 * 24 * time.Hour),
		Status:     domain.IncomeStatusActive,
		BaseSalary: 40000000,
		Bonus:      10000000,
	}

	req := dto.UpdateIncomeProfileRequest{
		Amount:     floatPtr(60000000),
		BaseSalary: floatPtr(50000000),
	}

	mockRepo.On("GetByID", ctx, profileID).Return(existingProfile, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(ip *domain.IncomeProfile) bool {
		return ip.ID == profileID && ip.Status == domain.IncomeStatusArchived
	})).Return(nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(ip *domain.IncomeProfile) bool {
		return ip.ID != profileID && ip.PreviousVersionID != nil && *ip.PreviousVersionID == profileID
	})).Return(nil)

	newVersion, err := service.UpdateIncomeProfile(ctx, userID.String(), profileID.String(), req)

	assert.NoError(t, err)
	assert.NotNil(t, newVersion)
	assert.NotEqual(t, profileID, newVersion.ID)
	assert.NotNil(t, newVersion.PreviousVersionID)
	assert.Equal(t, profileID, *newVersion.PreviousVersionID)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateIncomeProfile_AlreadyArchived(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	archivedProfile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
		Status: domain.IncomeStatusArchived,
	}
	now := time.Now()
	archivedProfile.ArchivedAt = &now

	req := dto.UpdateIncomeProfileRequest{
		Amount: floatPtr(60000000),
	}

	mockRepo.On("GetByID", ctx, profileID).Return(archivedProfile, nil)

	_, err := service.UpdateIncomeProfile(ctx, userID.String(), profileID.String(), req)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_VerifyIncomeProfile(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	profile := &domain.IncomeProfile{
		ID:         profileID,
		UserID:     userID,
		IsVerified: false,
	}

	mockRepo.On("GetByID", ctx, profileID).Return(profile, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(ip *domain.IncomeProfile) bool {
		return ip.ID == profileID && ip.IsVerified == true
	})).Return(nil)

	ip, err := service.VerifyIncomeProfile(ctx, userID.String(), profileID.String(), true)

	assert.NoError(t, err)
	assert.True(t, ip.IsVerified)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateDSSMetadata(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	profile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
	}

	req := dto.UpdateDSSMetadataRequest{
		StabilityScore: floatPtr(0.95),
		RiskLevel:      stringPtr("low"),
		Confidence:     floatPtr(0.85),
	}

	mockRepo.On("GetByID", ctx, profileID).Return(profile, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(ip *domain.IncomeProfile) bool {
		return ip.ID == profileID && len(ip.DSSMetadata) > 0
	})).Return(nil)

	ip, err := service.UpdateDSSMetadata(ctx, userID.String(), profileID.String(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, ip.DSSMetadata)
	mockRepo.AssertExpectations(t)
}

func TestService_ArchiveIncomeProfile(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	profile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
		Status: domain.IncomeStatusActive,
	}

	mockRepo.On("GetByID", ctx, profileID).Return(profile, nil)
	mockRepo.On("Archive", ctx, profileID, userID).Return(nil)

	err := service.ArchiveIncomeProfile(ctx, userID.String(), profileID.String())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_ArchiveIncomeProfile_AlreadyArchived(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	archivedProfile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
		Status: domain.IncomeStatusArchived,
	}
	now := time.Now()
	archivedProfile.ArchivedAt = &now

	mockRepo.On("GetByID", ctx, profileID).Return(archivedProfile, nil)

	err := service.ArchiveIncomeProfile(ctx, userID.String(), profileID.String())

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_CheckAndArchiveEnded(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	now := time.Now()
	endedDate := now.Add(-24 * time.Hour)

	activeProfiles := []*domain.IncomeProfile{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Status:    domain.IncomeStatusActive,
			StartDate: now.Add(-48 * time.Hour),
			EndDate:   &endedDate,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Status:    domain.IncomeStatusActive,
			StartDate: now.Add(-24 * time.Hour),
			EndDate:   nil, // No end date, should not archive
		},
	}

	mockRepo.On("GetActiveByUser", ctx, userID).Return(activeProfiles, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(ip *domain.IncomeProfile) bool {
		return ip.Status == domain.IncomeStatusEnded
	})).Return(nil).Once()

	count, err := service.CheckAndArchiveEnded(ctx, userID.String())

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteIncomeProfile(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	profileID := uuid.New()

	profile := &domain.IncomeProfile{
		ID:     profileID,
		UserID: userID,
	}

	mockRepo.On("GetByID", ctx, profileID).Return(profile, nil)
	mockRepo.On("Delete", ctx, profileID).Return(nil)

	err := service.DeleteIncomeProfile(ctx, userID.String(), profileID.String())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetActiveIncomes(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	expectedProfiles := []*domain.IncomeProfile{
		{
			ID:     uuid.New(),
			UserID: userID,
			Status: domain.IncomeStatusActive,
		},
	}

	mockRepo.On("GetActiveByUser", ctx, userID).Return(expectedProfiles, nil)

	profiles, err := service.GetActiveIncomes(ctx, userID.String())

	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	mockRepo.AssertExpectations(t)
}

func TestService_GetArchivedIncomes(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	expectedProfiles := []*domain.IncomeProfile{
		{
			ID:     uuid.New(),
			UserID: userID,
			Status: domain.IncomeStatusArchived,
		},
	}

	mockRepo.On("GetArchivedByUser", ctx, userID).Return(expectedProfiles, nil)

	profiles, err := service.GetArchivedIncomes(ctx, userID.String())

	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	mockRepo.AssertExpectations(t)
}

func TestService_GetRecurringIncomes(t *testing.T) {
	service, mockRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()

	expectedProfiles := []*domain.IncomeProfile{
		{
			ID:          uuid.New(),
			UserID:      userID,
			IsRecurring: true,
		},
	}

	mockRepo.On("GetRecurringByUser", ctx, userID).Return(expectedProfiles, nil)

	profiles, err := service.GetRecurringIncomes(ctx, userID.String())

	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	mockRepo.AssertExpectations(t)
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
