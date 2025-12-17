-- VANTARO MVP (Freelancer Money OS) schema
-- Postgres / Neon friendly

BEGIN;

-- 1) Extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- USERS
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text UNIQUE NOT NULL,
  password_hash text NOT NULL,
  full_name text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name text NOT NULL DEFAULT '';

-- INCOMES
CREATE TABLE IF NOT EXISTS incomes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  client_name text NOT NULL,
  amount bigint NOT NULL CHECK (amount > 0),
  currency text NOT NULL DEFAULT 'INR',
  received_on date NOT NULL,
  note text NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_incomes_user_id_created_at ON incomes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_incomes_user_date ON incomes(user_id, received_on DESC);
CREATE INDEX IF NOT EXISTS idx_incomes_user_client ON incomes(user_id, client_name);

-- EXPENSES
CREATE TABLE IF NOT EXISTS expenses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  vendor_name text NOT NULL,
  category text NOT NULL DEFAULT 'General',
  amount bigint NOT NULL CHECK (amount > 0),
  currency text NOT NULL DEFAULT 'INR',
  spent_on date NOT NULL,
  note text NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Backward compatibility: older schema used `merchant` instead of `vendor_name`.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'expenses'
      AND column_name = 'merchant'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'expenses'
      AND column_name = 'vendor_name'
  ) THEN
    ALTER TABLE expenses RENAME COLUMN merchant TO vendor_name;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_expenses_user_date ON expenses(user_id, spent_on DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_user_category ON expenses(user_id, category);

COMMIT;

-- Quick notes (so you don't get wrecked later)
-- Amount type: I used BIGINT so you can store money safely (recommended). Decide one rule:
-- store rupees (e.g., 1500) or
-- store paise (e.g., 150000).
-- Pick one and keep it everywhere.
--
-- This schema is enough for:
-- signup/login users
-- add income/expense
-- monthly totals (sum by date range)
-- per-client + per-category breakdown later
