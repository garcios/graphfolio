-- Migration: 000005_add_financial_year_end.down.sql
-- Description: Revert financial year end columns

ALTER TABLE users.users
    DROP CONSTRAINT IF EXISTS users_fy_day_check,
    DROP CONSTRAINT IF EXISTS users_fy_month_check,
    DROP COLUMN IF EXISTS financial_year_end_day,
    DROP COLUMN IF EXISTS financial_year_end_month;
