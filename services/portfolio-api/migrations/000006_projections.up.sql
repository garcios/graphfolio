-- 000006_projections.up.sql

CREATE TABLE portfolio.holdings (
    portfolio_id         uuid           NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    instrument_id        uuid           NOT NULL REFERENCES portfolio.instruments(id),
    quantity             numeric(28,10) NOT NULL,
    cost_basis           numeric(20,6)  NOT NULL,
    cost_basis_base      numeric(20,6)  NOT NULL,
    realized_pnl_base    numeric(20,6)  NOT NULL DEFAULT 0,
    dividends_base       numeric(20,6)  NOT NULL DEFAULT 0,
    last_transaction_id  uuid           REFERENCES portfolio.transactions(id) ON DELETE SET NULL,
    updated_at           timestamptz    NOT NULL DEFAULT now(),
    PRIMARY KEY (portfolio_id, instrument_id)
);

CREATE TABLE portfolio.cash_balances (
    portfolio_id   uuid          NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    currency_code  char(3)       NOT NULL REFERENCES portfolio.currencies(code),
    balance        numeric(20,6) NOT NULL,
    updated_at     timestamptz   NOT NULL DEFAULT now(),
    PRIMARY KEY (portfolio_id, currency_code)
);

CREATE TABLE portfolio.portfolio_valuations (
    portfolio_id        uuid           NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    valuation_date      date           NOT NULL,
    market_value_base   numeric(20,6)  NOT NULL,
    cash_value_base     numeric(20,6)  NOT NULL,
    net_flow_base       numeric(20,6)  NOT NULL DEFAULT 0,
    daily_return        numeric(20,12),
    twr_index           numeric(20,12) NOT NULL,
    PRIMARY KEY (portfolio_id, valuation_date)
);
