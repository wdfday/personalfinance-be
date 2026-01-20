-- ============================================================
-- Goal Module Migration: Type -> Category + Behavior
-- Run once, then delete this file
-- Date: 2026-01-16
-- ============================================================

-- Step 1: Add new columns
ALTER TABLE goals
ADD COLUMN IF NOT EXISTS behavior VARCHAR(20),
ADD COLUMN IF NOT EXISTS category VARCHAR(20),
ADD COLUMN IF NOT EXISTS converted_budget_id UUID;

-- Step 2: Copy data from old 'type' column to new 'category' column
UPDATE goals
SET
    category = type
WHERE
    category IS NULL
    AND type IS NOT NULL;

-- Step 3: Set default behavior based on existing conditions
-- - If has target_date and not recurring -> willing
-- - If has contribution_frequency -> recurring
-- - Default -> flexible
UPDATE goals
SET
    behavior = CASE
        WHEN contribution_frequency IS NOT NULL THEN 'recurring'
        WHEN target_date IS NOT NULL THEN 'willing'
        ELSE 'flexible'
    END
WHERE
    behavior IS NULL;

-- Step 4: Rename linked_account_id to account_id (if exists)
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'goals' AND column_name = 'linked_account_id'
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'goals' AND column_name = 'account_id'
  ) THEN
    ALTER TABLE goals RENAME COLUMN linked_account_id TO account_id;
  END IF;
END $$;

-- Step 5: Set NOT NULL constraints (after data is populated)
-- Only set if there's data to avoid errors on empty table
DO $$
BEGIN
  -- Check if table has rows and set defaults for NULL values
  UPDATE goals SET behavior = 'flexible' WHERE behavior IS NULL;
  UPDATE goals SET category = 'other' WHERE category IS NULL;
  
  -- Add NOT NULL constraints
  ALTER TABLE goals ALTER COLUMN behavior SET NOT NULL;
  ALTER TABLE goals ALTER COLUMN behavior SET DEFAULT 'flexible';
  ALTER TABLE goals ALTER COLUMN category SET NOT NULL;
END $$;

-- Step 6: Add foreign key for converted_budget_id (optional)
-- ALTER TABLE goals
--   ADD CONSTRAINT fk_goals_converted_budget
--   FOREIGN KEY (converted_budget_id) REFERENCES budgets(id);

-- Step 7: Drop old 'type' column (OPTIONAL - uncomment if ready)
-- ALTER TABLE goals DROP COLUMN IF EXISTS type;

-- ============================================================
-- Verification Query - Run to check migration success
-- ============================================================
-- SELECT
--   id, name,
--   behavior, category,
--   account_id,
--   type as old_type
-- FROM goals
-- LIMIT 10;

-- ============================================================
-- ROLLBACK (if needed)
-- ============================================================
-- ALTER TABLE goals
--   DROP COLUMN IF EXISTS behavior,
--   DROP COLUMN IF EXISTS category,
--   DROP COLUMN IF EXISTS converted_budget_id;
-- ALTER TABLE goals RENAME COLUMN account_id TO linked_account_id;