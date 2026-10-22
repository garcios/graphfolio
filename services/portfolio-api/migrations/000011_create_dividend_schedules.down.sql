-- 000011_create_dividend_schedules.down.sql

DROP TRIGGER IF EXISTS trg_protect_dividend_overrides ON portfolio.dividend_schedules;
DROP FUNCTION IF EXISTS portfolio.prevent_dividend_override_overwrite();
DROP TABLE IF EXISTS portfolio.dividend_schedules;
DROP TYPE IF EXISTS portfolio.dividend_status;
DROP TYPE IF EXISTS portfolio.dividend_frequency;
