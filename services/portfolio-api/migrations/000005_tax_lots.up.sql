-- 000005_tax_lots.up.sql

CREATE TABLE portfolio.tax_lots (
    id                    uuid           PRIMARY KEY DEFAULT uuidv7(),
    portfolio_id          uuid           NOT NULL REFERENCES portfolio.portfolios(id) ON DELETE CASCADE,
    instrument_id         uuid           NOT NULL REFERENCES portfolio.instruments(id),
    open_transaction_id   uuid           NOT NULL REFERENCES portfolio.transactions(id) ON DELETE CASCADE,
    acquired_date         date           NOT NULL,
    original_quantity     numeric(28,10) NOT NULL CHECK (original_quantity > 0),
    remaining_quantity    numeric(28,10) NOT NULL CHECK (remaining_quantity >= 0),
    cost_basis            numeric(20,6)  NOT NULL,
    cost_basis_base       numeric(20,6)  NOT NULL,
    closed_date           date,
    CHECK (remaining_quantity <= original_quantity)
);

CREATE INDEX tax_lots_open_idx
    ON portfolio.tax_lots (portfolio_id, instrument_id, acquired_date, id)
    WHERE remaining_quantity > 0;

CREATE TABLE portfolio.lot_disposals (
    id                        uuid           PRIMARY KEY DEFAULT uuidv7(),
    tax_lot_id                uuid           NOT NULL REFERENCES portfolio.tax_lots(id) ON DELETE CASCADE,
    sell_transaction_id       uuid           NOT NULL REFERENCES portfolio.transactions(id) ON DELETE CASCADE,
    quantity                  numeric(28,10) NOT NULL CHECK (quantity > 0),
    cost_basis_released       numeric(20,6)  NOT NULL,
    cost_basis_released_base  numeric(20,6)  NOT NULL,
    proceeds_base             numeric(20,6)  NOT NULL,
    realized_pnl_base         numeric(20,6)  NOT NULL,
    UNIQUE (tax_lot_id, sell_transaction_id)
);

CREATE INDEX lot_disposals_sell_idx ON portfolio.lot_disposals (sell_transaction_id);
