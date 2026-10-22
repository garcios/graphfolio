-- 000011_create_dividend_schedules.up.sql

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE t.typname = 'dividend_frequency' AND n.nspname = 'portfolio') THEN
        CREATE TYPE portfolio.dividend_frequency AS ENUM (
            'MONTHLY', 'QUARTERLY', 'SEMI_ANNUAL', 'ANNUAL', 'IRREGULAR'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE t.typname = 'dividend_status' AND n.nspname = 'portfolio') THEN
        CREATE TYPE portfolio.dividend_status AS ENUM (
            'ANNOUNCED', 'EX_DATE_PASSED', 'PAID', 'CANCELLED'
        );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS portfolio.dividend_schedules (
    id                  uuid                            PRIMARY KEY DEFAULT uuidv7(),
    instrument_id       uuid                            NOT NULL REFERENCES portfolio.instruments(id) ON DELETE CASCADE,
    ex_date             date                            NOT NULL,
    record_date         date,
    payment_date        date                            NOT NULL,
    amount_per_share    numeric(20,8)                   NOT NULL CHECK (amount_per_share > 0),
    currency_code       char(3)                         NOT NULL REFERENCES portfolio.currencies(code),
    frequency           portfolio.dividend_frequency    NOT NULL DEFAULT 'QUARTERLY',
    status              portfolio.dividend_status       NOT NULL DEFAULT 'ANNOUNCED',
    is_estimated        boolean                         NOT NULL DEFAULT false,
    is_override         boolean                         NOT NULL DEFAULT false,
    franking_percentage numeric(5,2)                   DEFAULT 0.00,
    withholding_tax_pct numeric(5,2)                   DEFAULT 0.00,
    source              varchar(32)                     NOT NULL DEFAULT 'yahoo',
    created_at          timestamptz                     NOT NULL DEFAULT now(),
    updated_at          timestamptz                     NOT NULL DEFAULT now(),
    UNIQUE (instrument_id, ex_date, payment_date)
);

CREATE INDEX IF NOT EXISTS idx_dividend_schedules_dates ON portfolio.dividend_schedules (ex_date, payment_date);
CREATE INDEX IF NOT EXISTS idx_dividend_schedules_instrument ON portfolio.dividend_schedules (instrument_id);
CREATE INDEX IF NOT EXISTS idx_dividend_schedules_status ON portfolio.dividend_schedules (status, is_estimated);

-- Trigger: Protect manual administrator overrides from automated feed overwrites
CREATE OR REPLACE FUNCTION portfolio.prevent_dividend_override_overwrite()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.is_override = true AND NEW.is_override = false THEN
        RAISE EXCEPTION 'Cannot overwrite manual admin dividend override with automated feed data';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_protect_dividend_overrides ON portfolio.dividend_schedules;
CREATE TRIGGER trg_protect_dividend_overrides
BEFORE UPDATE ON portfolio.dividend_schedules
FOR EACH ROW EXECUTE FUNCTION portfolio.prevent_dividend_override_overwrite();
