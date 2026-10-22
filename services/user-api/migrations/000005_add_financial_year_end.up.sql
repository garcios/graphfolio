-- Migration: 000005_add_financial_year_end.up.sql
-- Description: Add financial_year_end_month and financial_year_end_day to users.users

ALTER TABLE users.users
    ADD COLUMN IF NOT EXISTS financial_year_end_month smallint NOT NULL DEFAULT 6,
    ADD COLUMN IF NOT EXISTS financial_year_end_day   smallint NOT NULL DEFAULT 30;

-- Range sanity constraints
ALTER TABLE users.users
    DROP CONSTRAINT IF EXISTS users_fy_month_check;
ALTER TABLE users.users
    ADD CONSTRAINT users_fy_month_check CHECK (financial_year_end_month BETWEEN 1 AND 12);

ALTER TABLE users.users
    DROP CONSTRAINT IF EXISTS users_fy_day_check;
ALTER TABLE users.users
    ADD CONSTRAINT users_fy_day_check CHECK (financial_year_end_day BETWEEN 1 AND 31);

-- Default backfill based on current base_currency:
-- US, Canadian, European, and Swiss investors typically observe calendar year-end (31 December).
UPDATE users.users
SET financial_year_end_month = 12,
    financial_year_end_day = 31
WHERE base_currency IN ('USD', 'CAD', 'EUR', 'CHF')
  AND financial_year_end_month = 6
  AND financial_year_end_day = 30;
