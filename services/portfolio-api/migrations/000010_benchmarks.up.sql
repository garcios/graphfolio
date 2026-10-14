-- 000010_benchmarks.up.sql

CREATE TABLE portfolio.benchmarks (
    id            text PRIMARY KEY, -- e.g. 'SP500TR', 'NASDAQ100', 'ASX200TR'
    instrument_id uuid NOT NULL REFERENCES portfolio.instruments(id) ON DELETE CASCADE,
    name          text NOT NULL,
    index_type    text NOT NULL DEFAULT 'TOTAL_RETURN', -- 'TOTAL_RETURN' or 'PRICE_RETURN'
    is_default    boolean NOT NULL DEFAULT false,
    display_order int NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_benchmarks_display_order ON portfolio.benchmarks(display_order ASC);
