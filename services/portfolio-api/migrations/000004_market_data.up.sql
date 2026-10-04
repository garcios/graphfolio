-- 000004_market_data.up.sql

CREATE TABLE portfolio.instrument_prices (
    instrument_id  uuid          NOT NULL REFERENCES portfolio.instruments(id),
    price_date     date          NOT NULL,
    close          numeric(20,8) NOT NULL CHECK (close >= 0),
    source         text          NOT NULL DEFAULT 'manual',
    created_at     timestamptz   NOT NULL DEFAULT now(),
    PRIMARY KEY (instrument_id, price_date)
);

CREATE TABLE portfolio.fx_rates (
    base_currency   char(3)        NOT NULL REFERENCES portfolio.currencies(code),
    quote_currency  char(3)        NOT NULL REFERENCES portfolio.currencies(code),
    rate_date       date           NOT NULL,
    rate            numeric(20,10) NOT NULL CHECK (rate > 0),
    source          text           NOT NULL DEFAULT 'manual',
    PRIMARY KEY (base_currency, quote_currency, rate_date),
    CHECK (base_currency <> quote_currency)
);
