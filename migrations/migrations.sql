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
CREATE INDEX IF NOT EXISTS idx_expenses_user_id_created_at ON expenses(user_id, created_at DESC);

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

-- =========================
-- POINTS + REWARDS V1
-- =========================

-- Points ledger (append only)
CREATE TABLE IF NOT EXISTS points_ledger (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  source_txn_id TEXT NULL,
  points_delta INT NOT NULL,
  reason TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Cached points balance
CREATE TABLE IF NOT EXISTS points_balance (
  user_id UUID PRIMARY KEY,
  points_total BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tiers (simple for now; can be expanded later)
CREATE TABLE IF NOT EXISTS tiers (
  id BIGSERIAL PRIMARY KEY,
  tier_name TEXT NOT NULL UNIQUE,
  min_points BIGINT NOT NULL DEFAULT 0,
  multiplier NUMERIC(6,3) NOT NULL DEFAULT 1.000
);

INSERT INTO tiers (tier_name, min_points, multiplier)
VALUES
 ('STONE', 0, 1.000),
 ('SILVER', 2000, 1.050),
 ('OBSIDIAN', 10000, 1.100)
ON CONFLICT (tier_name) DO NOTHING;

-- Rewards catalog
CREATE TABLE IF NOT EXISTS rewards_catalog (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  type TEXT NOT NULL,               -- e.g. FLIGHT, HOTEL, PERK
  points_cost BIGINT NOT NULL,
  partner TEXT NULL,
  status TEXT NOT NULL DEFAULT 'COMING_SOON',  -- ACTIVE, COMING_SOON
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO rewards_catalog (title, type, points_cost, partner, status)
VALUES
 ('Flights (Coming Soon)', 'FLIGHT', 5000, 'Vantro Travel Partner', 'COMING_SOON')
ON CONFLICT DO NOTHING;

-- Additional active reward for testing
INSERT INTO rewards_catalog (title, type, points_cost, partner, status)
VALUES
 ('Airport Lounge Pass', 'PERK', 1500, 'Vantro Partner', 'ACTIVE')
ON CONFLICT DO NOTHING;

-- Redemptions table
CREATE TABLE IF NOT EXISTS redemptions (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  reward_id BIGINT NOT NULL REFERENCES rewards_catalog(id),
  points_spent BIGINT NOT NULL,
  status TEXT NOT NULL DEFAULT 'REQUESTED',     -- REQUESTED, APPROVED, FULFILLED, REJECTED
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_points_ledger_user_id_created_at
  ON points_ledger(user_id, created_at DESC);

-- avoid double-award for same transaction (only when source_txn_id present)
CREATE UNIQUE INDEX IF NOT EXISTS uq_points_ledger_user_txn
  ON points_ledger(user_id, source_txn_id)
  WHERE source_txn_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_redemptions_user_id_created_at
  ON redemptions(user_id, created_at DESC);

-- Unified transactions table (idempotent)
CREATE TABLE IF NOT EXISTS user_transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  amount BIGINT NOT NULL CHECK (amount >= 0),
  direction TEXT NOT NULL CHECK (direction IN ('IN','OUT')),
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_transactions_user_created_at
  ON user_transactions(user_id, created_at DESC);

-- ADMIN + TRACKING FIELDS (safe, minimal)
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;

-- Onboarding status
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS onboarding_step TEXT NOT NULL DEFAULT 'start';

-- Soft delete support for incomes/expenses
ALTER TABLE incomes
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE expenses
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_incomes_user_id_created_at_active
  ON incomes(user_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_user_id_created_at_active
  ON expenses(user_id, created_at DESC)
  WHERE deleted_at IS NULL;

-- Helpful index
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

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

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'fk_transactions_business'
      AND table_name = 'transactions'
      AND table_schema = 'public'
  ) THEN
    ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_business
    FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_transactions_business_created_at
ON transactions(business_id, created_at DESC);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    amount NUMERIC(12,2) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- TRANSACTIONS (UUID user_id)
-- Requires pgcrypto for gen_random_uuid (you already enabled it earlier)

CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('income','expense')),
  amount BIGINT NOT NULL CHECK (amount >= 0), -- store amount in smallest unit (paise) if you want; else treat as rupees
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id_created_at
ON transactions(user_id, created_at DESC);

-- =========================
-- V1 CORE TABLES (idempotent)
-- =========================

-- unified transactions for V1
CREATE TABLE IF NOT EXISTS transactions_v1 (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  amount BIGINT NOT NULL,
  direction TEXT NOT NULL CHECK (direction IN ('IN','OUT')),
  note TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_transactions_v1_user_created_at
  ON transactions_v1(user_id, created_at DESC);

-- Idempotency records scoped by logical "owner" (typically user_id or phone/client id)
CREATE TABLE IF NOT EXISTS idempotency_keys (
  id BIGSERIAL PRIMARY KEY,
  owner_id TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  request_hash TEXT NOT NULL,
  response_status INT NOT NULL,
  response_body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (owner_id, idempotency_key)
);

-- points ledger + balance
CREATE TABLE IF NOT EXISTS points_ledger (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  source_txn_id TEXT NULL,
  points_delta INT NOT NULL,
  reason TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_points_ledger_user_txn_reason
  ON points_ledger(user_id, source_txn_id, reason)
  WHERE source_txn_id IS NOT NULL AND reason = 'earn';

CREATE TABLE IF NOT EXISTS points_balance (
  user_id UUID PRIMARY KEY,
  points_total BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tiers (
  id BIGSERIAL PRIMARY KEY,
  tier_name TEXT NOT NULL UNIQUE,
  min_points BIGINT NOT NULL DEFAULT 0,
  multiplier NUMERIC(6,3) NOT NULL DEFAULT 1.000
);
INSERT INTO tiers (tier_name, min_points, multiplier)
VALUES
 ('STONE', 0, 1.000),
 ('SILVER', 2000, 1.050),
 ('OBSIDIAN', 10000, 1.100)
ON CONFLICT (tier_name) DO NOTHING;

CREATE TABLE IF NOT EXISTS rewards_catalog (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  type TEXT NOT NULL,
  points_cost BIGINT NOT NULL,
  partner TEXT NULL,
  status TEXT NOT NULL DEFAULT 'COMING_SOON',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO rewards_catalog (title, type, points_cost, partner, status)
VALUES
 ('Flights (Coming Soon)', 'FLIGHT', 5000, 'Vantro Travel Partner', 'COMING_SOON'),
 ('Airport Lounge Pass', 'PERK', 1500, 'Vantro Partner', 'ACTIVE')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS redemptions (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  reward_id BIGINT NOT NULL REFERENCES rewards_catalog(id),
  points_spent BIGINT NOT NULL,
  status TEXT NOT NULL DEFAULT 'REQUESTED',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_points_ledger_user_id_created_at
  ON points_ledger(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_redemptions_user_id_created_at
  ON redemptions(user_id, created_at DESC);

-- Vantro Expense Memory: expenses table

CREATE TABLE IF NOT EXISTS expenses (
  id BIGSERIAL PRIMARY KEY,
  user_phone TEXT NOT NULL,
  amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
  currency TEXT NOT NULL DEFAULT 'INR',
  category TEXT NOT NULL DEFAULT 'MISC',
  note TEXT,
  source TEXT NOT NULL DEFAULT 'manual', -- manual | whatsapp | app | upi (future)
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expenses_user_phone_created_at
  ON expenses (user_phone, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_expenses_category
  ON expenses (category);

-- Vantro Expense Memory: subscriptions (for paid PDF reports)

CREATE TABLE IF NOT EXISTS subscriptions (
  id BIGSERIAL PRIMARY KEY,
  user_phone TEXT NOT NULL UNIQUE,
  plan TEXT NOT NULL DEFAULT 'expense_memory_monthly',
  status TEXT NOT NULL DEFAULT 'inactive', -- inactive | active
  current_period_end TIMESTAMPTZ,          -- when access expires
  razorpay_payment_link_id TEXT,           -- last payment link id (optional)
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_status
  ON subscriptions (status);

-- Store generated PDF links (tokenized access)
CREATE TABLE IF NOT EXISTS reports (
  id BIGSERIAL PRIMARY KEY,
  user_phone TEXT NOT NULL,
  month TEXT NOT NULL, -- YYYY-MM
  token TEXT NOT NULL UNIQUE,
  file_path TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_user_month ON reports(user_phone, month);
CREATE INDEX IF NOT EXISTS idx_reports_expires ON reports(expires_at);

-- ============================
-- AUDIT LOGS (append-only)
-- ============================
CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NULL,
  action TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NULL,
  ip TEXT NULL,
  user_agent TEXT NULL,
  metadata JSONB NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created_at
  ON audit_logs(user_id, created_at DESC);

-- ============================
-- PHASE 3: Money normalization
-- Store all money as BIGINT paise (minor units)
-- ============================

-- 1) Convert legacy NUMERIC amounts → BIGINT paise
-- If your `transactions.amount` is currently NUMERIC(12,2), convert:
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'transactions'
      AND column_name = 'amount'
      AND data_type IN ('numeric', 'decimal')
  ) THEN
    -- Convert numeric rupees to paise BIGINT
    ALTER TABLE transactions
      ALTER COLUMN amount TYPE BIGINT
      USING ROUND(amount * 100)::BIGINT;
  END IF;
END $$;

-- If user_transactions.amount exists and is not BIGINT, convert too:
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'user_transactions'
      AND column_name = 'amount'
      AND data_type IN ('numeric', 'decimal')
  ) THEN
    ALTER TABLE user_transactions
      ALTER COLUMN amount TYPE BIGINT
      USING ROUND(amount * 100)::BIGINT;
  END IF;
END $$;

-- 2) Safety constraints (no negative money unless you explicitly want refunds)
-- Adjust if you support negative adjustments.
ALTER TABLE incomes
  ADD CONSTRAINT IF NOT EXISTS incomes_amount_nonneg CHECK (amount >= 0);

ALTER TABLE expenses
  ADD CONSTRAINT IF NOT EXISTS expenses_amount_nonneg CHECK (amount >= 0);

ALTER TABLE transactions
  ADD CONSTRAINT IF NOT EXISTS transactions_amount_nonneg CHECK (amount >= 0);

ALTER TABLE user_transactions
  ADD CONSTRAINT IF NOT EXISTS user_transactions_amount_nonneg CHECK (amount >= 0);

-- 3) Optional: upper bound protection (prevents insane values)
-- Example: max ₹10,00,00,000 (10 crore) => paise = 1000000000 * 100 = 100000000000
DO $$
DECLARE
  max_paise BIGINT := 100000000000;
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='transactions') THEN
    ALTER TABLE transactions
      ADD CONSTRAINT IF NOT EXISTS transactions_amount_max CHECK (amount <= max_paise);
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='user_transactions') THEN
    ALTER TABLE user_transactions
      ADD CONSTRAINT IF NOT EXISTS user_transactions_amount_max CHECK (amount <= max_paise);
  END IF;

  ALTER TABLE incomes
    ADD CONSTRAINT IF NOT EXISTS incomes_amount_max CHECK (amount <= max_paise);

  ALTER TABLE expenses
    ADD CONSTRAINT IF NOT EXISTS expenses_amount_max CHECK (amount <= max_paise);
END $$;


✅ This makes money storage consistent + safe.
