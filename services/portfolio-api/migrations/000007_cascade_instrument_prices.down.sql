-- 000007_cascade_instrument_prices.down.sql
-- Revert cascading deletion on instrument prices.

ALTER TABLE portfolio.instrument_prices
    DROP CONSTRAINT IF EXISTS instrument_prices_instrument_id_fkey,
    ADD CONSTRAINT instrument_prices_instrument_id_fkey
        FOREIGN KEY (instrument_id) REFERENCES portfolio.instruments(id);
