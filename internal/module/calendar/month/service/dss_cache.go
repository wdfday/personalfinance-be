package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"personalfinancedss/internal/module/calendar/month/dto"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// DSS cache TTL: 3 hours (user session typically shorter)
	dssCacheTTL = 3 * time.Hour
)

// DSSCache handles caching of DSS preview results in Redis
type DSSCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewDSSCache creates a new DSS cache service
func NewDSSCache(client *redis.Client, logger *zap.Logger) *DSSCache {
	return &DSSCache{
		client: client,
		logger: logger,
	}
}

// DSSCachedState holds the complete DSS workflow state in Redis (single key approach)
type DSSCachedState struct {
	MonthID   uuid.UUID `json:"month_id"`
	UserID    uuid.UUID `json:"user_id"`
	UpdatedAt time.Time `json:"updated_at"`

	// ===== Input Snapshot - captured when DSS workflow is initialized =====
	MonthlyIncome    float64                   `json:"monthly_income,omitempty"`
	InputGoals       []dto.InitGoalInput       `json:"input_goals,omitempty"`
	InputDebts       []dto.InitDebtInput       `json:"input_debts,omitempty"`
	InputConstraints []dto.InitConstraintInput `json:"input_constraints,omitempty"`

	// ===== Step 0: Auto-scoring results =====
	AutoScoring interface{} `json:"auto_scoring,omitempty"`

	// ===== Step 1: Goal prioritization =====
	GoalPrioritizationPreview interface{} `json:"goal_prioritization_preview,omitempty"`
	AcceptedGoalRanking       []uuid.UUID `json:"accepted_goal_ranking,omitempty"`

	// ===== Step 2: Debt strategy =====
	DebtStrategyPreview  interface{} `json:"debt_strategy_preview,omitempty"`
	AcceptedDebtStrategy string      `json:"accepted_debt_strategy,omitempty"`
	// DebtAllocationWeights lưu phân bổ extra payment theo chiến lược:
	//   strategy -> (debtID -> weight trong [0,1], tổng mỗi strategy = 1)
	DebtAllocationWeights map[string]map[uuid.UUID]float64 `json:"debt_allocation_weights,omitempty"`

	// ===== Step 4: Budget allocation =====
	BudgetAllocationPreview interface{}           `json:"budget_allocation_preview,omitempty"`
	AcceptedScenario        string                `json:"accepted_scenario,omitempty"`
	AcceptedAllocations     map[uuid.UUID]float64 `json:"accepted_allocations,omitempty"`
}

// buildStateKey constructs Redis key for complete DSS state
// Format: dss:state:{monthID}:{userID}
func (c *DSSCache) buildStateKey(monthID, userID uuid.UUID) string {
	return fmt.Sprintf("dss:state:%s:%s", monthID.String(), userID.String())
}

// GetState retrieves the complete DSS workflow state from Redis
func (c *DSSCache) GetState(ctx context.Context, monthID, userID uuid.UUID) (*DSSCachedState, error) {
	if c.client == nil {
		return nil, fmt.Errorf("redis unavailable")
	}

	key := c.buildStateKey(monthID, userID)

	bytes, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// State doesn't exist yet - return nil without error
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get DSS state: %w", err)
	}

	var state DSSCachedState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DSS state: %w", err)
	}

	c.logger.Debug("Retrieved DSS state from cache",
		zap.String("key", key),
		zap.Time("updated_at", state.UpdatedAt))

	return &state, nil
}

// ErrDSSNotInitialized is returned when DSS workflow is accessed before initialization
var ErrDSSNotInitialized = fmt.Errorf("DSS not initialized - call POST /dss/initialize first")

// GetOrInitState gets existing state or returns error if not initialized
// Use this for Preview steps - they REQUIRE Initialize to be called first
func (c *DSSCache) GetOrInitState(ctx context.Context, monthID, userID uuid.UUID) (*DSSCachedState, error) {
	state, err := c.GetState(ctx, monthID, userID)
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, ErrDSSNotInitialized
	}

	return state, nil
}

// SaveState saves the complete DSS workflow state to Redis
func (c *DSSCache) SaveState(ctx context.Context, state *DSSCachedState) error {
	if c.client == nil {
		c.logger.Debug("Redis unavailable, skipping save")
		return nil
	}

	state.UpdatedAt = time.Now()
	key := c.buildStateKey(state.MonthID, state.UserID)

	bytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal DSS state: %w", err)
	}

	if err := c.client.Set(ctx, key, bytes, dssCacheTTL).Err(); err != nil {
		c.logger.Error("Failed to save DSS state",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Saved DSS state to cache",
		zap.String("key", key),
		zap.Duration("ttl", dssCacheTTL))

	return nil
}

// ClearState deletes the complete DSS workflow state from Redis
func (c *DSSCache) ClearState(ctx context.Context, monthID, userID uuid.UUID) error {
	if c.client == nil {
		return nil
	}

	key := c.buildStateKey(monthID, userID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Error("Failed to clear DSS state",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	c.logger.Info("Cleared DSS state from cache",
		zap.String("month_id", monthID.String()),
		zap.String("user_id", userID.String()))

	return nil
}
