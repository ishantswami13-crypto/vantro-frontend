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

-- 004_transactions.sql

CREATE TABLE IF NOT EXISTS transactions (
  id BIGSERIAL PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
  amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id_created_at
ON transactions(user_id, created_at DESC);

-- 005_businesses.sql

CREATE TABLE IF NOT EXISTS businesses (
  id BIGSERIAL PRIMARY KEY,
  owner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  currency TEXT NOT NULL DEFAULT 'INR',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_businesses_owner_user_id_created_at
ON businesses(owner_user_id, created_at DESC);

ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS business_id BIGINT;

-- backfill for existing rows: create a default business per user
DO $$
DECLARE
  u RECORD;
  bid BIGINT;
BEGIN
  FOR u IN SELECT DISTINCT user_id FROM transactions LOOP
    INSERT INTO businesses (owner_user_id, name, currency)
    VALUES (u.user_id, 'Default Business', 'INR')
    ON CONFLICT DO NOTHING
    RETURNING id INTO bid;

    IF bid IS NULL THEN
      SELECT id INTO bid
      FROM businesses
      WHERE owner_user_id = u.user_id
      ORDER BY created_at ASC
      LIMIT 1;
    END IF;

    UPDATE transactions
    SET business_id = bid
    WHERE user_id = u.user_id AND business_id IS NULL;
  END LOOP;
END $$;

ALTER TABLE transactions
ALTER COLUMN business_id SET NOT NULL;

ALTER TABLE transactions
ADD CONSTRAINT fk_transactions_business
FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_transactions_business_created_at
ON transactions(business_id, created_at DESC);
