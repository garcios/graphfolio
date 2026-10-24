-- 000013_add_transaction_franking_credits.down.sql

ALTER TABLE portfolio.transactions
    DROP CONSTRAINT IF EXISTS non_negative_franking,
    DROP COLUMN IF EXISTS franking_pct,
    DROP COLUMN IF EXISTS franking_credit;
