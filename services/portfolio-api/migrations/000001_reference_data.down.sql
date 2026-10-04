-- 000001_reference_data.down.sql

DROP TABLE IF EXISTS portfolio.instruments;
DROP TABLE IF EXISTS portfolio.exchanges;
DROP TABLE IF EXISTS portfolio.currencies;
DROP TYPE IF EXISTS portfolio.asset_class;
DROP FUNCTION IF EXISTS portfolio.uuidv7();
