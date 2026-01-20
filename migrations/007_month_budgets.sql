-- Month Module: YNAB-style Zero-Based Budgeting with Event Sourcing
-- Migration: 007_month_budgets.sql

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- TABLE: monthly_budgets (Read Model - JSONB State)
-- =====================================================
-- This table stores the materialized view of the budget state
-- The JSONB 'state' column contains the full budget breakdown

CREATE TABLE IF NOT EXISTS monthly_budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    budget_id UUID NOT NULL,
    month VARCHAR(10) NOT NULL,  -- Display name: "2024-02", "Jan 2024"

-- Budget period dates (supports custom periods like pay periods)
start_date DATE NOT NULL DEFAULT '2024-01-01',
end_date DATE NOT NULL DEFAULT '2024-01-31',
status VARCHAR(20) DEFAULT 'OPEN' CHECK (
    status IN ('OPEN', 'CLOSED', 'ARCHIVED')
),

-- The Big JSONB State
-- Structure: {"tbb": 0, "categories": {"<uuid>": {"rollover": 0, "assigned": 0, "activity": 0, "available": 0, ...}}}
state JSONB NOT NULL DEFAULT '{}'::jsonb,

-- Optimistic locking / Event stream position
version BIGINT DEFAULT 1 NOT NULL,
created_at TIMESTAMP DEFAULT NOW() NOT NULL,
updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
deleted_at TIMESTAMP,

-- Unique constraint: one record per budget per month
UNIQUE(budget_id, month) );

-- Indexes for monthly_budgets
CREATE INDEX idx_monthly_budgets_budget_month ON monthly_budgets (budget_id, month);

CREATE INDEX idx_monthly_budgets_status ON monthly_budgets (status);

CREATE INDEX idx_monthly_budgets_deleted_at ON monthly_budgets (deleted_at);

CREATE INDEX idx_monthly_budgets_date_range ON monthly_budgets (
    budget_id,
    start_date,
    end_date
);

-- GIN index for JSONB queries (e.g., searching within category states)
CREATE INDEX idx_monthly_budgets_state ON monthly_budgets USING GIN (state);

-- Updated_at trigger for monthly_budgets
CREATE OR REPLACE FUNCTION update_monthly_budgets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_monthly_budgets_updated_at
    BEFORE UPDATE ON monthly_budgets
    FOR EACH ROW
    EXECUTE FUNCTION update_monthly_budgets_updated_at();

-- =====================================================
-- TABLE: month_event_logs (Event Store - Append-Only)
-- =====================================================
-- This table stores all domain events for event sourcing

CREATE TABLE IF NOT EXISTS month_event_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id UUID NOT NULL,  -- Month ID (references monthly_budgets.id)
    event_type VARCHAR(50) NOT NULL CHECK (
        event_type IN (
            'month.created',
            'category.assigned',
            'money.moved',
            'income.received',
            'transaction.posted',
            'month.closed'
        )
    ),
    payload JSONB NOT NULL,  -- Event-specific data
    version BIGINT NOT NULL,  -- Sequence number for this aggregate (1, 2, 3, ...)
    occurred_at TIMESTAMP DEFAULT NOW() NOT NULL,
    user_id UUID,  -- Who triggered this event (NULL for system events)

-- Unique constraint: prevent duplicate versions for same aggregate
UNIQUE(aggregate_id, version) );

-- Indexes for month_event_logs
CREATE INDEX idx_month_event_logs_aggregate ON month_event_logs (aggregate_id, version);

CREATE INDEX idx_month_event_logs_type ON month_event_logs (event_type);

CREATE INDEX idx_month_event_logs_occurred_at ON month_event_logs (occurred_at);

CREATE INDEX idx_month_event_logs_user ON month_event_logs (user_id);

-- GIN index for JSONB payload queries
CREATE INDEX idx_month_event_logs_payload ON month_event_logs USING GIN (payload);

-- =====================================================
-- COMMENTS (Documentation)
-- =====================================================

COMMENT ON TABLE monthly_budgets IS 'Read model for budgeting months with JSONB state (YNAB-style)';

COMMENT ON COLUMN monthly_budgets.month IS 'Display name for the period (e.g., "2024-02", "Pay Period 1/15-2/14")';

COMMENT ON COLUMN monthly_budgets.start_date IS 'Actual start date of the budget period (supports custom periods)';

COMMENT ON COLUMN monthly_budgets.end_date IS 'Actual end date of the budget period (supports custom periods)';

COMMENT ON COLUMN monthly_budgets.state IS 'JSONB containing full budget state: TBB and category states';

COMMENT ON COLUMN monthly_budgets.version IS 'Optimistic locking version / event stream position';

COMMENT ON TABLE month_event_logs IS 'Event store for domain events (Event Sourcing pattern)';

COMMENT ON COLUMN month_event_logs.aggregate_id IS 'Month ID (references monthly_budgets.id)';

COMMENT ON COLUMN month_event_logs.version IS 'Event sequence number for this aggregate (auto-increment per aggregate)';

COMMENT ON COLUMN month_event_logs.payload IS 'JSON payload containing event-specific data';

-- =====================================================
-- SAMPLE DATA (Optional - for testing)
-- =====================================================
-- You can add sample data here or in a separate seed file