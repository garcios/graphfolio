-- 000003_corporate_actions.up.sql

CREATE TYPE portfolio.corporate_action_type AS ENUM (
    'SPLIT', 'REVERSE_SPLIT', 'SPIN_OFF', 'MERGER', 'SYMBOL_CHANGE', 'RETURN_OF_CAPITAL'
);

CREATE TABLE portfolio.corporate_actions (
    id                 uuid                            PRIMARY KEY DEFAULT uuidv7(),
    instrument_id      uuid                            NOT NULL REFERENCES portfolio.instruments(id),
    type               portfolio.corporate_action_type NOT NULL,
    ex_date            date                            NOT NULL,
    ratio_from         numeric(20,10),
    ratio_to           numeric(20,10),
    new_instrument_id  uuid                            REFERENCES portfolio.instruments(id),
    cash_amount        numeric(20,6),
    currency_code      char(3)                         REFERENCES portfolio.currencies(code),
    cost_basis_pct     numeric(9,6),
    notes              text,
    created_at         timestamptz                     NOT NULL DEFAULT now(),
    UNIQUE (instrument_id, type, ex_date),
    CONSTRAINT ratio_positive CHECK (
        type NOT IN ('SPLIT','REVERSE_SPLIT') OR (ratio_from > 0 AND ratio_to > 0)
    )
);
