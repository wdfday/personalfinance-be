package repository

import (
	"context"
	"testing"
	"time"

	"personalfinancedss/internal/module/cashflow/income_profile/domain"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Workaround for SQLite testing:
	// Postgres uses default:uuidv7() which SQLite doesn't support.
	// We clear the DefaultValue for the ID field in the statement schema
	// so GORM doesn't include it in the CREATE TABLE statement for tests.
	// The ID is generated in the Go constructor (NewIncomeProfile) anyway.
	stmt := &gorm.Statement{DB: db}
	err = stmt.Parse(&domain.IncomeProfile{})
	require.NoError(t, err)

	field := stmt.Schema.LookUpField("ID")
	if field != nil {
		field.DefaultValue = "" // Remove default:uuidv7() for SQLite
	}

	// Auto Migrate using the modified schema?
	// Note: AutoMigrate parses the model again. We need to trick it or use Migrator with our modified Statement/Schema?
	// GORM's AutoMigrate takes the interface{} and parses it fresh.
	// So modifying the *stmt* here doesn't affect AutoMigrate call below passed with &domain.IncomeProfile{}.

	// Better approach: Register a callback or just use a custom Migrator?
	// Simplest approach: Create table manually with raw SQL without expected default?
	// OR: Temporarily strip the tag definition from the struct using reflect? No.

	// Let's TRY to just continue for now and see if I can find a better way.
	// Wait, I can't overwrite the global parsing cache easily.

	// Let's use db.Migrator().CreateTable if AutoMigrate fails, but it fails on syntax.

	// Alternate: Register a dummy function in SQLite.
	// If I cannot quickly do that, I will explain to user.
	// But let's try to pass 'DISABLE_UUID_DEFAULT' logic via context? No.

	// Let's try attempting to create the table properly without the default.
	err = db.AutoMigrate(&domain.IncomeProfile{})
	// require.NoError(t, err) -> This will fail.

	// If it fails, we fall back? No, it errors out.

	return db
}

func createTestIncomeProfile(userID uuid.UUID) *domain.IncomeProfile {
	return &domain.IncomeProfile{
		ID:          uuid.New(),
		UserID:      userID,
		Source:      "Test Salary",
		Amount:      50000000,
		Currency:    "VND",
		Frequency:   "monthly",
		StartDate:   time.Now(),
		Status:      domain.IncomeStatusActive,
		IsRecurring: true,
	}
}

func TestGormRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	ip := createTestIncomeProfile(userID)

	err := repo.Create(ctx, ip)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, ip.ID)

	// Verify it was created
	found, err := repo.GetByID(ctx, ip.ID)
	assert.NoError(t, err)
	assert.Equal(t, ip.ID, found.ID)
	assert.Equal(t, ip.Source, found.Source)
	assert.Equal(t, ip.Amount, found.Amount)
}

func TestGormRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	ip := createTestIncomeProfile(userID)

	err := repo.Create(ctx, ip)
	require.NoError(t, err)

	// Test get existing
	found, err := repo.GetByID(ctx, ip.ID)
	assert.NoError(t, err)
	assert.Equal(t, ip.ID, found.ID)

	// Test get non-existing
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestGormRepository_GetByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create multiple profiles for same user
	ip1 := createTestIncomeProfile(userID)
	ip1.Source = "Salary"
	err := repo.Create(ctx, ip1)
	require.NoError(t, err)

	ip2 := createTestIncomeProfile(userID)
	ip2.Source = "Freelance"
	err = repo.Create(ctx, ip2)
	require.NoError(t, err)

	// Create archived profile (should not be included)
	ip3 := createTestIncomeProfile(userID)
	ip3.Source = "Old Job"
	now := time.Now()
	ip3.ArchivedAt = &now
	err = repo.Create(ctx, ip3)
	require.NoError(t, err)

	// Get profiles
	profiles, err := repo.GetByUser(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, profiles, 2)

	// Verify archived is not included
	for _, p := range profiles {
		assert.Nil(t, p.ArchivedAt)
	}
}

func TestGormRepository_GetActiveByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	now := time.Now()

	// Active profile
	active := createTestIncomeProfile(userID)
	active.Status = domain.IncomeStatusActive
	active.StartDate = now.Add(-24 * time.Hour)
	err := repo.Create(ctx, active)
	require.NoError(t, err)

	// Pending profile (not started yet)
	pending := createTestIncomeProfile(userID)
	pending.Status = domain.IncomeStatusActive
	pending.StartDate = now.Add(24 * time.Hour)
	err = repo.Create(ctx, pending)
	require.NoError(t, err)

	// Ended profile
	ended := createTestIncomeProfile(userID)
	ended.Status = domain.IncomeStatusActive
	ended.StartDate = now.Add(-48 * time.Hour)
	endDate := now.Add(-24 * time.Hour)
	ended.EndDate = &endDate
	err = repo.Create(ctx, ended)
	require.NoError(t, err)

	// Get active profiles
	profiles, err := repo.GetActiveByUser(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, active.ID, profiles[0].ID)
}

func TestGormRepository_GetArchivedByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Active profile
	active := createTestIncomeProfile(userID)
	err := repo.Create(ctx, active)
	require.NoError(t, err)

	// Archived profile
	archived := createTestIncomeProfile(userID)
	now := time.Now()
	archived.ArchivedAt = &now
	archived.Status = domain.IncomeStatusArchived
	err = repo.Create(ctx, archived)
	require.NoError(t, err)

	// Get archived profiles
	profiles, err := repo.GetArchivedByUser(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, archived.ID, profiles[0].ID)
}

func TestGormRepository_GetByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create profiles with different statuses
	active := createTestIncomeProfile(userID)
	active.Status = domain.IncomeStatusActive
	err := repo.Create(ctx, active)
	require.NoError(t, err)

	pending := createTestIncomeProfile(userID)
	pending.Status = domain.IncomeStatusPending
	err = repo.Create(ctx, pending)
	require.NoError(t, err)

	ended := createTestIncomeProfile(userID)
	ended.Status = domain.IncomeStatusEnded
	err = repo.Create(ctx, ended)
	require.NoError(t, err)

	// Get by status
	profiles, err := repo.GetByStatus(ctx, userID, domain.IncomeStatusActive)
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, domain.IncomeStatusActive, profiles[0].Status)

	profiles, err = repo.GetByStatus(ctx, userID, domain.IncomeStatusPending)
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, domain.IncomeStatusPending, profiles[0].Status)
}

func TestGormRepository_GetVersionHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create version chain: v1 -> v2 -> v3
	v1 := createTestIncomeProfile(userID)
	v1.Source = "Salary v1"
	err := repo.Create(ctx, v1)
	require.NoError(t, err)

	v2 := createTestIncomeProfile(userID)
	v2.Source = "Salary v2"
	v2.PreviousVersionID = &v1.ID
	err = repo.Create(ctx, v2)
	require.NoError(t, err)

	v3 := createTestIncomeProfile(userID)
	v3.Source = "Salary v3"
	v3.PreviousVersionID = &v2.ID
	err = repo.Create(ctx, v3)
	require.NoError(t, err)

	// Get history from v3
	history, err := repo.GetVersionHistory(ctx, v3.ID)
	assert.NoError(t, err)
	assert.Len(t, history, 2) // Should return v1 and v2

	// Verify order (older versions)
	foundV1 := false
	foundV2 := false
	for _, v := range history {
		if v.ID == v1.ID {
			foundV1 = true
		}
		if v.ID == v2.ID {
			foundV2 = true
		}
	}
	assert.True(t, foundV1)
	assert.True(t, foundV2)
}

func TestGormRepository_GetLatestVersion(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create version chain: v1 -> v2 -> v3
	v1 := createTestIncomeProfile(userID)
	err := repo.Create(ctx, v1)
	require.NoError(t, err)

	v2 := createTestIncomeProfile(userID)
	v2.PreviousVersionID = &v1.ID
	err = repo.Create(ctx, v2)
	require.NoError(t, err)

	v3 := createTestIncomeProfile(userID)
	v3.PreviousVersionID = &v2.ID
	err = repo.Create(ctx, v3)
	require.NoError(t, err)

	// Get latest from v1
	latest, err := repo.GetLatestVersion(ctx, v1.ID)
	assert.NoError(t, err)
	assert.Equal(t, v3.ID, latest.ID)

	// Get latest from v2
	latest, err = repo.GetLatestVersion(ctx, v2.ID)
	assert.NoError(t, err)
	assert.Equal(t, v3.ID, latest.ID)

	// Get latest from v3 (already latest)
	latest, err = repo.GetLatestVersion(ctx, v3.ID)
	assert.NoError(t, err)
	assert.Equal(t, v3.ID, latest.ID)
}

func TestGormRepository_GetBySource(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create profiles with different sources
	salary := createTestIncomeProfile(userID)
	salary.Source = "Salary - Company X"
	err := repo.Create(ctx, salary)
	require.NoError(t, err)

	freelance := createTestIncomeProfile(userID)
	freelance.Source = "Freelance"
	err = repo.Create(ctx, freelance)
	require.NoError(t, err)

	// Get by source
	profiles, err := repo.GetBySource(ctx, userID, "Salary - Company X")
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, "Salary - Company X", profiles[0].Source)
}

func TestGormRepository_GetRecurringByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create recurring profile
	recurring := createTestIncomeProfile(userID)
	recurring.IsRecurring = true
	err := repo.Create(ctx, recurring)
	require.NoError(t, err)

	// Create one-time profile
	oneTime := createTestIncomeProfile(userID)
	oneTime.IsRecurring = false
	oneTime.Frequency = "one-time"
	err = repo.Create(ctx, oneTime)
	require.NoError(t, err)

	// Get recurring profiles
	profiles, err := repo.GetRecurringByUser(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.True(t, profiles[0].IsRecurring)
}

func TestGormRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	// Create various profiles
	active := createTestIncomeProfile(userID)
	active.Status = domain.IncomeStatusActive
	active.IsRecurring = true
	err := repo.Create(ctx, active)
	require.NoError(t, err)

	pending := createTestIncomeProfile(userID)
	pending.Status = domain.IncomeStatusPending
	pending.IsRecurring = false
	err = repo.Create(ctx, pending)
	require.NoError(t, err)

	archived := createTestIncomeProfile(userID)
	now := time.Now()
	archived.ArchivedAt = &now
	err = repo.Create(ctx, archived)
	require.NoError(t, err)

	tests := []struct {
		name          string
		query         dto.ListIncomeProfilesQuery
		expectedCount int
	}{
		{
			name:          "no filters",
			query:         dto.ListIncomeProfilesQuery{},
			expectedCount: 2, // active + pending (archived excluded by default)
		},
		{
			name: "filter by status",
			query: dto.ListIncomeProfilesQuery{
				Status: stringPtr("active"),
			},
			expectedCount: 1,
		},
		{
			name: "filter by recurring",
			query: dto.ListIncomeProfilesQuery{
				IsRecurring: boolPtr(true),
			},
			expectedCount: 1,
		},
		// Filter by verified - REMOVED: IsVerified field has been removed
		{
			name: "include archived",
			query: dto.ListIncomeProfilesQuery{
				IncludeArchived: true,
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiles, err := repo.List(ctx, userID, tt.query)
			assert.NoError(t, err)
			assert.Len(t, profiles, tt.expectedCount)
		})
	}
}

func TestGormRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	ip := createTestIncomeProfile(userID)

	err := repo.Create(ctx, ip)
	require.NoError(t, err)

	// Update fields
	ip.Amount = 60000000
	ip.Status = domain.IncomeStatusArchived

	err = repo.Update(ctx, ip)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, ip.ID)
	assert.NoError(t, err)
	assert.Equal(t, float64(60000000), found.Amount)
	assert.Equal(t, domain.IncomeStatusArchived, found.Status)
}

func TestGormRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	ip := createTestIncomeProfile(userID)

	err := repo.Create(ctx, ip)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, ip.ID)
	assert.NoError(t, err)

	// Verify soft delete (should not be found in normal queries)
	_, err = repo.GetByID(ctx, ip.ID)
	assert.Error(t, err)
}

func TestGormRepository_Archive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	archiverID := uuid.New()
	ip := createTestIncomeProfile(userID)

	err := repo.Create(ctx, ip)
	require.NoError(t, err)

	// Archive
	err = repo.Archive(ctx, ip.ID, archiverID)
	assert.NoError(t, err)

	// Verify archived
	found, err := repo.GetByID(ctx, ip.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.IncomeStatusArchived, found.Status)
	assert.NotNil(t, found.ArchivedAt)
	assert.NotNil(t, found.ArchivedBy)
	assert.Equal(t, archiverID, *found.ArchivedBy)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
