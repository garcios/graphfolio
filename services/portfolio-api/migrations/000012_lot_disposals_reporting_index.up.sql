-- 000012_lot_disposals_reporting_index.up.sql

CREATE INDEX IF NOT EXISTS lot_disposals_cgt_reporting_idx
    ON portfolio.lot_disposals (sell_transaction_id, tax_lot_id);
