-- Test Data for Broker Migration
-- This script creates sample accounts with broker_integration JSONB to test migration

-- Clean up any existing test data
DELETE FROM accounts WHERE account_name LIKE 'TEST:%';

-- Test user (assume exists or create one)
INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, is_verified, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'test_broker@example.com', '$2a$10$test_hash', 'Test', 'User', true, true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Account 1: SSI Securities (Active)
INSERT INTO accounts (
    id,
    user_id,
    account_name,
    account_type,
    current_balance,
    currency,
    is_active,
    is_auto_sync,
    broker_integration,
    created_at,
    updated_at
) VALUES (
    '10000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'TEST: SSI Securities Account',
    'investment',
    50000000.00,
    'VND',
    true,
    true,
    jsonb_build_object(
        'broker_type', 'ssi',
        'broker_name', 'SSI Securities',
        'is_active', true,
        'access_token', 'encrypted_ssi_access_token_123',
        'refresh_token', 'encrypted_ssi_refresh_token_123',
        'token_expires_at', (NOW() + INTERVAL '1 hour')::text,
        'last_refreshed_at', NOW()::text,
        'auto_sync', true,
        'sync_frequency', 60,
        'total_syncs', 15,
        'successful_syncs', 14,
        'failed_syncs', 1,
        'sync_assets', true,
        'sync_transactions', true,
        'sync_prices', true,
        'sync_balance', true,
        'ssi_consumer_id', 'encrypted_consumer_id_123',
        'ssi_consumer_secret', 'encrypted_consumer_secret_123',
        'ssi_otp_method', 'PIN'
    ),
    NOW(),
    NOW()
);

-- Test Account 2: OKX Exchange (Active)
INSERT INTO accounts (
    id,
    user_id,
    account_name,
    account_type,
    current_balance,
    currency,
    is_active,
    is_auto_sync,
    broker_integration,
    created_at,
    updated_at
) VALUES (
    '10000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'TEST: OKX Crypto Account',
    'crypto_wallet',
    15000.00,
    'USD',
    true,
    true,
    jsonb_build_object(
        'broker_type', 'okx',
        'broker_name', 'OKX Exchange',
        'is_active', true,
        'access_token', 'encrypted_okx_access_token_456',
        'refresh_token', 'encrypted_okx_refresh_token_456',
        'token_expires_at', (NOW() + INTERVAL '2 hours')::text,
        'last_refreshed_at', NOW()::text,
        'auto_sync', true,
        'sync_frequency', 30,
        'total_syncs', 50,
        'successful_syncs', 48,
        'failed_syncs', 2,
        'sync_assets', true,
        'sync_transactions', true,
        'sync_prices', true,
        'sync_balance', true,
        'okx_passphrase', 'encrypted_passphrase_456'
    ),
    NOW(),
    NOW()
);

-- Test Account 3: SePay Banking (Inactive)
INSERT INTO accounts (
    id,
    user_id,
    account_name,
    account_type,
    current_balance,
    currency,
    is_active,
    is_auto_sync,
    broker_integration,
    created_at,
    updated_at
) VALUES (
    '10000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000001',
    'TEST: SePay Banking Account',
    'bank',
    25000000.00,
    'VND',
    true,
    false,
    jsonb_build_object(
        'broker_type', 'sepay',
        'broker_name', 'SePay',
        'is_active', false,
        'access_token', 'encrypted_sepay_token_789',
        'refresh_token', 'encrypted_sepay_token_789',
        'token_expires_at', (NOW() + INTERVAL '24 hours')::text,
        'last_refreshed_at', NOW()::text,
        'auto_sync', false,
        'sync_frequency', 120,
        'total_syncs', 5,
        'successful_syncs', 5,
        'failed_syncs', 0,
        'sync_assets', false,
        'sync_transactions', true,
        'sync_prices', false,
        'sync_balance', true
    ),
    NOW(),
    NOW()
);

-- Test Account 4: No broker integration (should be skipped)
INSERT INTO accounts (
    id,
    user_id,
    account_name,
    account_type,
    current_balance,
    currency,
    is_active,
    created_at,
    updated_at
) VALUES (
    '10000000-0000-0000-0000-000000000004',
    '00000000-0000-0000-0000-000000000001',
    'TEST: Regular Account (No Broker)',
    'bank',
    10000000.00,
    'VND',
    true,
    NOW(),
    NOW()
);

-- Verify test data
SELECT
    id,
    account_name,
    account_type,
    broker_integration->>'broker_type' as broker_type,
    broker_integration->>'is_active' as is_active,
    broker_integration IS NOT NULL as has_broker
FROM accounts
WHERE account_name LIKE 'TEST:%'
ORDER BY account_name;

-- Show count
SELECT COUNT(*) as test_accounts_with_broker
FROM accounts
WHERE account_name LIKE 'TEST:%'
  AND broker_integration IS NOT NULL;
