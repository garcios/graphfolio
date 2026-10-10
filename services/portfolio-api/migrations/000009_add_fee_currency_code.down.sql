-- 000009_add_fee_currency_code.down.sql
ALTER TABLE portfolio.transactions
    DROP COLUMN IF EXISTS fee_currency_code;
