-- Migration: Broker Connections Table
-- Description: Create broker_connections table and migrate data from accounts.broker_integration JSONB
-- Date: 2025-01-29

-- =====================================================
-- STEP 1: Create broker_connections table
-- =====================================================

CREATE TABLE IF NOT EXISTS broker_connections (
    id UUID DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Broker information
    broker_type VARCHAR(20) NOT NULL,
    broker_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- Credentials (encrypted)
    api_key TEXT,
    api_secret TEXT,
    passphrase TEXT,
    consumer_id VARCHAR(100),
    consumer_secret TEXT,
    otp_method VARCHAR(20),

    -- Token management
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    last_refreshed_at TIMESTAMP,

    -- Sync settings
    auto_sync BOOLEAN DEFAULT true,
    sync_frequency INTEGER DEFAULT 60,
    last_sync_at TIMESTAMP,
    last_sync_status VARCHAR(20),
    last_sync_error TEXT,

    -- Sync statistics
    total_syncs INTEGER DEFAULT 0,
    successful_syncs INTEGER DEFAULT 0,
    failed_syncs INTEGER DEFAULT 0,

    -- Additional settings
    sync_assets BOOLEAN DEFAULT true,
    sync_transactions BOOLEAN DEFAULT true,
    sync_prices BOOLEAN DEFAULT true,
    sync_balance BOOLEAN DEFAULT true,

    -- External account info
    external_account_id VARCHAR(100),
    external_account_number VARCHAR(100),
    external_account_name VARCHAR(255),

    -- Metadata
    notes TEXT,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Indexes
    CONSTRAINT broker_connections_pkey PRIMARY KEY (id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_broker_connections_user_id ON broker_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_broker_connections_broker_type ON broker_connections(broker_type);
CREATE INDEX IF NOT EXISTS idx_broker_connections_status ON broker_connections(status);
CREATE INDEX IF NOT EXISTS idx_broker_connections_deleted_at ON broker_connections(deleted_at);

-- =====================================================
-- STEP 2: Add broker_connection_id to accounts table
-- =====================================================

ALTER TABLE accounts
ADD COLUMN IF NOT EXISTS broker_connection_id UUID REFERENCES broker_connections(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_accounts_broker_connection_id ON accounts(broker_connection_id);

-- =====================================================
-- STEP 3: Migrate data from accounts.broker_integration to broker_connections
-- =====================================================

-- This migration will:
-- 1. Read broker_integration JSONB from accounts
-- 2. Create a new broker_connection record
-- 3. Link the account to the new broker_connection
-- 4. Keep the old JSONB for backward compatibility

DO $$
DECLARE
    account_record RECORD;
    broker_data JSONB;
    new_broker_id UUID;
    broker_type_val VARCHAR(20);
BEGIN
    -- Loop through all accounts that have broker_integration
    FOR account_record IN
        SELECT id, user_id, broker_integration
        FROM accounts
        WHERE broker_integration IS NOT NULL
          AND broker_integration::TEXT != 'null'
          AND broker_integration::TEXT != '{}'
          AND broker_connection_id IS NULL  -- Only migrate if not already migrated
    LOOP
        broker_data := account_record.broker_integration;

        -- Extract broker_type, default to 'ssi' if not present
        broker_type_val := COALESCE(broker_data->>'broker_type', 'ssi');

        -- Create new broker_connection record
        INSERT INTO broker_connections (
            id,
            user_id,
            broker_type,
            broker_name,
            status,
            api_key,
            api_secret,
            passphrase,
            consumer_id,
            consumer_secret,
            otp_method,
            access_token,
            refresh_token,
            token_expires_at,
            last_refreshed_at,
            auto_sync,
            sync_frequency,
            last_sync_at,
            total_syncs,
            successful_syncs,
            failed_syncs,
            sync_assets,
            sync_transactions,
            sync_prices,
            sync_balance,
            created_at,
            updated_at
        ) VALUES (
            uuid_generate_v4(),
            account_record.user_id,
            broker_type_val,
            COALESCE(broker_data->>'broker_name', INITCAP(broker_type_val)),
            CASE
                WHEN (broker_data->>'is_active')::BOOLEAN THEN 'active'::VARCHAR
                ELSE 'disconnected'::VARCHAR
            END,
            broker_data->>'access_token',  -- Will be encrypted in application layer
            broker_data->>'access_token',  -- Use access_token as api_key for now
            broker_data->>'okx_passphrase',
            broker_data->>'ssi_consumer_id',
            broker_data->>'ssi_consumer_secret',
            broker_data->>'ssi_otp_method',
            broker_data->>'access_token',
            broker_data->>'refresh_token',
            CASE
                WHEN broker_data->>'token_expires_at' IS NOT NULL
                THEN (broker_data->>'token_expires_at')::TIMESTAMP
                ELSE NULL
            END,
            CASE
                WHEN broker_data->>'last_refreshed_at' IS NOT NULL
                THEN (broker_data->>'last_refreshed_at')::TIMESTAMP
                ELSE NULL
            END,
            COALESCE((broker_data->>'auto_sync')::BOOLEAN, true),
            COALESCE((broker_data->>'sync_frequency')::INTEGER, 60),
            CURRENT_TIMESTAMP,
            COALESCE((broker_data->>'total_syncs')::INTEGER, 0),
            COALESCE((broker_data->>'successful_syncs')::INTEGER, 0),
            COALESCE((broker_data->>'failed_syncs')::INTEGER, 0),
            COALESCE((broker_data->>'sync_assets')::BOOLEAN, true),
            COALESCE((broker_data->>'sync_transactions')::BOOLEAN, true),
            COALESCE((broker_data->>'sync_prices')::BOOLEAN, true),
            COALESCE((broker_data->>'sync_balance')::BOOLEAN, true),
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        ) RETURNING id INTO new_broker_id;

        -- Link account to new broker_connection
        UPDATE accounts
        SET broker_connection_id = new_broker_id
        WHERE id = account_record.id;

        RAISE NOTICE 'Migrated broker integration for account % to broker_connection %',
            account_record.id, new_broker_id;
    END LOOP;

    RAISE NOTICE 'Migration completed successfully';
END $$;

-- =====================================================
-- STEP 4: Create function to auto-update updated_at
-- =====================================================

CREATE OR REPLACE FUNCTION update_broker_connections_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_broker_connections_updated_at
    BEFORE UPDATE ON broker_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_broker_connections_updated_at();

-- =====================================================
-- ROLLBACK (if needed)
-- =====================================================

-- To rollback this migration:
-- DROP TRIGGER IF EXISTS trigger_update_broker_connections_updated_at ON broker_connections;
-- DROP FUNCTION IF EXISTS update_broker_connections_updated_at();
-- ALTER TABLE accounts DROP COLUMN IF EXISTS broker_connection_id;
-- DROP TABLE IF EXISTS broker_connections CASCADE;
