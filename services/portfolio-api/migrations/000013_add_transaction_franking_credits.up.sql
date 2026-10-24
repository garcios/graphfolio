-- 000013_add_transaction_franking_credits.up.sql

ALTER TABLE portfolio.transactions
    ADD COLUMN franking_credit numeric(20,6) NOT NULL DEFAULT 0,
    ADD COLUMN franking_pct    numeric(5,2)  DEFAULT NULL;

ALTER TABLE portfolio.transactions
    ADD CONSTRAINT non_negative_franking CHECK (franking_credit >= 0);
