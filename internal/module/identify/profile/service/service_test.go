package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"personalfinancedss/internal/module/identify/profile/domain"
	"personalfinancedss/internal/module/identify/profile/dto"
	"personalfinancedss/internal/shared"
)

// ==================== Mocks ====================

type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) Create(ctx context.Context, profile *domain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserProfile), args.Error(1)
}

func (m *MockProfileRepository) Update(ctx context.Context, profile *domain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockProfileRepository) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// ==================== Tests ====================

// TestCreateProfile tests profile creation
func TestCreateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Create profile with full data", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()
		occupation := "Software Engineer"
		industry := "Technology"
		monthlyIncome := 50000000.0
		creditScore := 750

		req := dto.CreateProfileRequest{
			Occupation:       &occupation,
			Industry:         &industry,
			MonthlyIncomeAvg: &monthlyIncome,
			CreditScore:      &creditScore,
			RiskTolerance:    domain.RiskToleranceModerate,
		}

		mockRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.UserProfile) bool {
			return p.UserID == userID &&
				*p.Occupation == occupation &&
				*p.Industry == industry &&
				*p.MonthlyIncomeAvg == monthlyIncome &&
				*p.CreditScore == creditScore
		})).Return(nil)

		result, err := service.CreateProfile(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, occupation, *result.Occupation)
		assert.Equal(t, industry, *result.Industry)
		assert.Equal(t, monthlyIncome, *result.MonthlyIncomeAvg)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Create default profile", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()

		mockRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.UserProfile) bool {
			return p.UserID == userID &&
				p.RiskTolerance == domain.RiskToleranceModerate &&
				p.InvestmentHorizon == domain.InvestmentHorizonMedium &&
				p.InvestmentExperience == domain.InvestmentExperienceBeginner
		})).Return(nil)

		result, err := service.CreateDefaultProfile(ctx, userID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, domain.RiskToleranceModerate, result.RiskTolerance)
		assert.False(t, result.OnboardingCompleted)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid user ID", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		req := dto.CreateProfileRequest{}

		result, err := service.CreateProfile(ctx, "invalid-uuid", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()
		req := dto.CreateProfileRequest{}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.UserProfile")).
			Return(errors.New("database error"))

		result, err := service.CreateProfile(ctx, userID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

// TestGetProfile tests profile retrieval
func TestGetProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Get existing profile", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()
		occupation := "Software Engineer"
		monthlyIncome := 50000000.0

		profile := &domain.UserProfile{
			ID:               uuid.New(),
			UserID:           userID,
			Occupation:       &occupation,
			MonthlyIncomeAvg: &monthlyIncome,
			RiskTolerance:    domain.RiskToleranceModerate,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(profile, nil)

		result, err := service.GetProfile(ctx, userID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, occupation, *result.Occupation)
		assert.Equal(t, monthlyIncome, *result.MonthlyIncomeAvg)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Profile not found", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()

		mockRepo.On("GetByUserID", ctx, userID).Return(nil, shared.ErrProfileNotFound)

		result, err := service.GetProfile(ctx, userID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, shared.ErrProfileNotFound, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid user ID", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		result, err := service.GetProfile(ctx, "invalid-uuid")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestUpdateProfile tests profile updates
func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - Update profile with partial data", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()
		oldOccupation := "Software Engineer"
		newOccupation := "Senior Software Engineer"
		oldIncome := 50000000.0
		newIncome := 70000000.0

		existingProfile := &domain.UserProfile{
			ID:               uuid.New(),
			UserID:           userID,
			Occupation:       &oldOccupation,
			MonthlyIncomeAvg: &oldIncome,
			RiskTolerance:    domain.RiskToleranceModerate,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		req := dto.UpdateProfileRequest{
			Occupation:       &newOccupation,
			MonthlyIncomeAvg: &newIncome,
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(existingProfile, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *domain.UserProfile) bool {
			return p.UserID == userID &&
				*p.Occupation == newOccupation &&
				*p.MonthlyIncomeAvg == newIncome
		})).Return(nil)

		result, err := service.UpdateProfile(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newOccupation, *result.Occupation)
		assert.Equal(t, newIncome, *result.MonthlyIncomeAvg)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Update risk tolerance", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()

		existingProfile := &domain.UserProfile{
			ID:            uuid.New(),
			UserID:        userID,
			RiskTolerance: domain.RiskToleranceModerate,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		newRiskTolerance := domain.RiskToleranceAggressive
		req := dto.UpdateProfileRequest{
			RiskTolerance: &newRiskTolerance,
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(existingProfile, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *domain.UserProfile) bool {
			return p.UserID == userID && p.RiskTolerance == newRiskTolerance
		})).Return(nil)

		result, err := service.UpdateProfile(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newRiskTolerance, result.RiskTolerance)
		assert.True(t, result.IsAggressiveInvestor())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Mark onboarding completed", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()

		existingProfile := &domain.UserProfile{
			ID:                  uuid.New(),
			UserID:              userID,
			OnboardingCompleted: false,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		onboardingCompleted := true
		req := dto.UpdateProfileRequest{
			OnboardingCompleted: &onboardingCompleted,
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(existingProfile, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *domain.UserProfile) bool {
			return p.UserID == userID &&
				p.OnboardingCompleted == true &&
				p.OnboardingCompletedAt != nil
		})).Return(nil)

		result, err := service.UpdateProfile(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.OnboardingCompleted)
		assert.NotNil(t, result.OnboardingCompletedAt)
		assert.True(t, result.IsOnboardingComplete())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Profile not found", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		userID := uuid.New()
		req := dto.UpdateProfileRequest{}

		mockRepo.On("GetByUserID", ctx, userID).Return(nil, shared.ErrProfileNotFound)

		result, err := service.UpdateProfile(ctx, userID.String(), req)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid user ID", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)

		req := dto.UpdateProfileRequest{}

		result, err := service.UpdateProfile(ctx, "invalid-uuid", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestProfileBusinessLogic tests domain business logic
func TestProfileBusinessLogic(t *testing.T) {
	t.Run("Risk level calculation", func(t *testing.T) {
		profile := &domain.UserProfile{
			RiskTolerance: domain.RiskToleranceConservative,
		}
		assert.Equal(t, 1, profile.GetRiskLevel())
		assert.True(t, profile.IsConservativeInvestor())
		assert.False(t, profile.IsAggressiveInvestor())

		profile.RiskTolerance = domain.RiskToleranceModerate
		assert.Equal(t, 2, profile.GetRiskLevel())

		profile.RiskTolerance = domain.RiskToleranceAggressive
		assert.Equal(t, 3, profile.GetRiskLevel())
		assert.True(t, profile.IsAggressiveInvestor())
	})

	t.Run("Emergency fund check", func(t *testing.T) {
		profile := &domain.UserProfile{}
		assert.False(t, profile.HasEmergencyFund())

		months := 3.0
		profile.EmergencyFundMonths = &months
		assert.True(t, profile.HasEmergencyFund())

		zeroMonths := 0.0
		profile.EmergencyFundMonths = &zeroMonths
		assert.False(t, profile.HasEmergencyFund())
	})

	t.Run("Onboarding status", func(t *testing.T) {
		profile := &domain.UserProfile{
			OnboardingCompleted: false,
		}
		assert.False(t, profile.IsOnboardingComplete())

		profile.OnboardingCompleted = true
		assert.True(t, profile.IsOnboardingComplete())
	})
}

// TestProfileDefaults tests default values
func TestProfileDefaults(t *testing.T) {
	t.Run("Default profile has correct defaults", func(t *testing.T) {
		mockRepo := new(MockProfileRepository)
		service := NewService(mockRepo)
		ctx := context.Background()

		userID := uuid.New()

		mockRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.UserProfile) bool {
			return p.RiskTolerance == domain.RiskToleranceModerate &&
				p.InvestmentHorizon == domain.InvestmentHorizonMedium &&
				p.InvestmentExperience == domain.InvestmentExperienceBeginner &&
				p.BudgetMethod == domain.BudgetMethodCustom &&
				p.CurrencyPrimary == "VND" &&
				p.CurrencySecondary == "USD" &&
				p.OnboardingCompleted == false
		})).Return(nil)

		result, err := service.CreateDefaultProfile(ctx, userID.String())

		assert.NoError(t, err)
		assert.Equal(t, domain.RiskToleranceModerate, result.RiskTolerance)
		assert.Equal(t, domain.InvestmentHorizonMedium, result.InvestmentHorizon)
		assert.Equal(t, domain.InvestmentExperienceBeginner, result.InvestmentExperience)
		assert.Equal(t, "VND", result.CurrencyPrimary)
		assert.Equal(t, "USD", result.CurrencySecondary)
		assert.False(t, result.OnboardingCompleted)

		mockRepo.AssertExpectations(t)
	})
}

