-- 000001_reference_data.up.sql

DO $$
BEGIN
  IF current_setting('server_version_num')::int < 180000 THEN
    CREATE OR REPLACE FUNCTION portfolio.uuidv7() RETURNS uuid AS $func$
    DECLARE
      unix_time_ms bytea;
      uuid_bytes bytea;
    BEGIN
      unix_time_ms := substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3 for 6);
      uuid_bytes := unix_time_ms || substring(uuid_send(gen_random_uuid()) from 7 for 10);
      uuid_bytes := set_byte(uuid_bytes, 6, (get_byte(uuid_bytes, 6) & 15) | 112);
      uuid_bytes := set_byte(uuid_bytes, 8, (get_byte(uuid_bytes, 8) & 63) | 128);
      RETURN encode(uuid_bytes, 'hex')::uuid;
    END;
    $func$ LANGUAGE plpgsql VOLATILE;
  END IF;
END $$;

CREATE TYPE portfolio.asset_class AS ENUM ('EQUITY', 'ETF', 'FUND', 'BOND', 'CRYPTO', 'CASH_EQUIVALENT');

CREATE TABLE portfolio.currencies (
    code         char(3)  PRIMARY KEY,
    name         text     NOT NULL,
    minor_units  smallint NOT NULL DEFAULT 2
);

CREATE TABLE portfolio.exchanges (
    code      text    PRIMARY KEY,
    name      text    NOT NULL,
    country   char(2) NOT NULL,
    timezone  text    NOT NULL
);

CREATE TABLE portfolio.instruments (
    id             uuid                  PRIMARY KEY DEFAULT uuidv7(),
    symbol         text                  NOT NULL,
    exchange_code  text                  NOT NULL REFERENCES portfolio.exchanges(code),
    name           text                  NOT NULL,
    asset_class    portfolio.asset_class NOT NULL,
    currency_code  char(3)               NOT NULL REFERENCES portfolio.currencies(code),
    isin           char(12)              UNIQUE,
    is_active      boolean               NOT NULL DEFAULT true,
    created_at     timestamptz           NOT NULL DEFAULT now(),
    updated_at     timestamptz           NOT NULL DEFAULT now(),
    UNIQUE (symbol, exchange_code)
);
