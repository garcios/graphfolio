-- 000002_portfolios_and_transactions.up.sql

CREATE TYPE portfolio.cost_basis_method AS ENUM ('AVERAGE_COST', 'FIFO');

CREATE TYPE portfolio.transaction_type AS ENUM (
    'BUY', 'SELL',
    'DIVIDEND', 'INTEREST',
    'DEPOSIT', 'WITHDRAWAL',
    'FEE', 'TAX',
    'TRANSFER_IN', 'TRANSFER_OUT',
    'FX_CONVERSION'
);

CREATE TABLE portfolio.portfolios (
    id                 uuid                        PRIMARY KEY DEFAULT uuidv7(),
    user_id            uuid                        NOT NULL,
    name               text                        NOT NULL,
    base_currency      char(3)                     NOT NULL REFERENCES portfolio.currencies(code),
    cost_basis_method  portfolio.cost_basis_method NOT NULL DEFAULT 'AVERAGE_COST',
    created_at         timestamptz                 NOT NULL DEFAULT now(),
    updated_at         timestamptz                 NOT NULL DEFAULT now(),
    archived_at        timestamptz,
    UNIQUE (user_id, name)
);

CREATE INDEX portfolios_user_id_idx ON portfolio.portfolios (user_id) WHERE archived_at IS NULL;

CREATE TABLE portfolio.transactions (
    id               uuid                       PRIMARY KEY DEFAULT uuidv7(),
    portfolio_id     uuid                       NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    instrument_id    uuid                       REFERENCES portfolio.instruments(id),
    type             portfolio.transaction_type NOT NULL,
    trade_date       date                       NOT NULL,
    settle_date      date,
    quantity         numeric(28,10),
    price            numeric(20,8),
    amount           numeric(20,6)              NOT NULL,
    currency_code    char(3)                    NOT NULL REFERENCES portfolio.currencies(code),
    fee              numeric(20,6)              NOT NULL DEFAULT 0,
    withholding_tax  numeric(20,6)              NOT NULL DEFAULT 0,
    fx_rate_to_base  numeric(20,10),
    external_ref     text,
    notes            text,
    created_at       timestamptz                NOT NULL DEFAULT now(),

    CONSTRAINT instrument_required CHECK (
        type NOT IN ('BUY','SELL','DIVIDEND','TRANSFER_IN','TRANSFER_OUT')
        OR instrument_id IS NOT NULL
    ),
    CONSTRAINT trade_fields_required CHECK (
        type NOT IN ('BUY','SELL','TRANSFER_IN','TRANSFER_OUT') OR (quantity > 0 AND price >= 0)
    ),
    CONSTRAINT non_negative_charges CHECK (fee >= 0 AND withholding_tax >= 0),
    UNIQUE (portfolio_id, external_ref)
);

CREATE INDEX transactions_portfolio_date_idx       ON portfolio.transactions (portfolio_id, trade_date, id);
CREATE INDEX transactions_portfolio_instrument_idx ON portfolio.transactions (portfolio_id, instrument_id, trade_date, id);
