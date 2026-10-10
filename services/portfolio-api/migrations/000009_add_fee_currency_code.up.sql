-- 000009_add_fee_currency_code.up.sql
ALTER TABLE portfolio.transactions
    ADD COLUMN fee_currency_code char(3) REFERENCES portfolio.currencies(code);
