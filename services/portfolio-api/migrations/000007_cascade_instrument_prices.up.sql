-- 000007_cascade_instrument_prices.up.sql
-- Allow cascading deletion of historical prices when an instrument is deleted.

ALTER TABLE portfolio.instrument_prices
    DROP CONSTRAINT IF EXISTS instrument_prices_instrument_id_fkey,
    ADD CONSTRAINT instrument_prices_instrument_id_fkey
        FOREIGN KEY (instrument_id) REFERENCES portfolio.instruments(id) ON DELETE CASCADE;
