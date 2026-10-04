-- GraphFolio Development Seed Data
-- Idempotent: deletes prior demo data and inserts fresh records.

BEGIN;

-- 1. Reference Data: Currencies
INSERT INTO portfolio.currencies (code, name, minor_units) VALUES
  ('USD', 'US Dollar', 2),
  ('AUD', 'Australian Dollar', 2),
  ('EUR', 'Euro', 2),
  ('GBP', 'British Pound', 2)
ON CONFLICT (code) DO NOTHING;

-- 2. Reference Data: Exchanges
INSERT INTO portfolio.exchanges (code, name, country, timezone) VALUES
  ('XNAS', 'NASDAQ Stock Market', 'US', 'America/New_York'),
  ('XNYS', 'New York Stock Exchange', 'US', 'America/New_York'),
  ('XASX', 'Australian Securities Exchange', 'AU', 'Australia/Sydney')
ON CONFLICT (code) DO NOTHING;

-- 3. Reference Data: Instruments (Fixed UUIDs for deterministic testing/demos)
INSERT INTO portfolio.instruments (id, symbol, exchange_code, name, asset_class, currency_code) VALUES
  ('018f0000-0001-7000-8000-000000000001', 'AAPL', 'XNAS', 'Apple Inc.', 'EQUITY', 'USD'),
  ('018f0000-0001-7000-8000-000000000002', 'MSFT', 'XNAS', 'Microsoft', 'EQUITY', 'USD'),
  ('018f0000-0001-7000-8000-000000000003', 'TSLA', 'XNAS', 'Tesla', 'EQUITY', 'USD'),
  ('018f0000-0001-7000-8000-000000000004', 'NVDA', 'XNAS', 'NVIDIA Corp.', 'EQUITY', 'USD'),
  ('018f0000-0001-7000-8000-000000000005', 'V',    'XNYS', 'Visa Inc.', 'EQUITY', 'USD')
ON CONFLICT (symbol, exchange_code) DO UPDATE
  SET name = EXCLUDED.name, currency_code = EXCLUDED.currency_code;

-- 4. Demo Portfolios
-- Primary demo portfolio (AVERAGE_COST)
INSERT INTO portfolio.portfolios (id, user_id, name, base_currency, cost_basis_method) VALUES
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0000-7000-8000-000000000001', 'Main Portfolio', 'USD', 'AVERAGE_COST'),
  ('018f0000-0002-7000-8000-000000000002', '018f0000-0000-7000-8000-000000000001', 'FIFO Growth Portfolio', 'USD', 'FIFO')
ON CONFLICT (user_id, name) DO NOTHING;

-- Clean prior seed transactions/projections for the demo portfolios
DELETE FROM portfolio.transactions WHERE portfolio_id IN (
  '018f0000-0002-7000-8000-000000000001',
  '018f0000-0002-7000-8000-000000000002'
);
DELETE FROM portfolio.portfolio_valuations WHERE portfolio_id IN (
  '018f0000-0002-7000-8000-000000000001',
  '018f0000-0002-7000-8000-000000000002'
);

-- 5. Seed Transactions for Main Portfolio
-- Initial Deposit
INSERT INTO portfolio.transactions (
  portfolio_id, type, trade_date, amount, currency_code, external_ref, notes
) VALUES (
  '018f0000-0002-7000-8000-000000000001', 'DEPOSIT', '2025-01-02', 120000.00, 'USD', 'DEP-001', 'Initial Funding'
);

-- Buys matching mock holdings
INSERT INTO portfolio.transactions (
  portfolio_id, instrument_id, type, trade_date, quantity, price, amount, currency_code, external_ref
) VALUES
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000001', 'BUY', '2025-01-03', 142.5, 156.235, 22263.50, 'USD', 'BUY-AAPL-1'),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000002', 'BUY', '2025-01-03', 85.0,  302.696, 25729.15, 'USD', 'BUY-MSFT-1'),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000003', 'BUY', '2025-01-03', 50.0,  235.450, 11772.50, 'USD', 'BUY-TSLA-1'),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000004', 'BUY', '2025-01-03', 35.0,  307.044, 10746.55, 'USD', 'BUY-NVDA-1'),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000005', 'BUY', '2025-01-03', 45.0,  253.211, 11394.50, 'USD', 'BUY-V-1');

-- 6. Market Prices (Latest Close and Previous Trading Day Close)
INSERT INTO portfolio.instrument_prices (instrument_id, price_date, close, source) VALUES
  ('018f0000-0001-7000-8000-000000000001', CURRENT_DATE, 185.92, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000002', CURRENT_DATE, 402.11, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000003', CURRENT_DATE, 210.45, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000004', CURRENT_DATE, 721.33, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000005', CURRENT_DATE, 278.10, 'dev_seed'),
  -- Previous trading day prices
  ('018f0000-0001-7000-8000-000000000001', CURRENT_DATE - INTERVAL '1 day', 182.02, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000002', CURRENT_DATE - INTERVAL '1 day', 396.11, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000003', CURRENT_DATE - INTERVAL '1 day', 217.17, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000004', CURRENT_DATE - INTERVAL '1 day', 691.33, 'dev_seed'),
  ('018f0000-0001-7000-8000-000000000005', CURRENT_DATE - INTERVAL '1 day', 277.82, 'dev_seed')
ON CONFLICT (instrument_id, price_date) DO UPDATE SET close = EXCLUDED.close;

-- 7. FX Rates
INSERT INTO portfolio.fx_rates (base_currency, quote_currency, rate_date, rate, source) VALUES
  ('USD', 'AUD', CURRENT_DATE, 1.5230, 'dev_seed'),
  ('USD', 'EUR', CURRENT_DATE, 0.9215, 'dev_seed'),
  ('USD', 'GBP', CURRENT_DATE, 0.7890, 'dev_seed')
ON CONFLICT (base_currency, quote_currency, rate_date) DO UPDATE SET rate = EXCLUDED.rate;

-- 8. Projections: Cash Balances ($8,450 USD)
INSERT INTO portfolio.cash_balances (portfolio_id, currency_code, balance) VALUES
  ('018f0000-0002-7000-8000-000000000001', 'USD', 8450.00)
ON CONFLICT (portfolio_id, currency_code) DO UPDATE SET balance = EXCLUDED.balance;

-- 9. Projections: Holdings (Exact metrics matching mock)
INSERT INTO portfolio.holdings (
  portfolio_id, instrument_id, quantity, cost_basis, cost_basis_base, realized_pnl_base, dividends_base
) VALUES
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000001', 142.5, 22263.50, 22263.50, 0, 0),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000002', 85.0,  25729.15, 25729.15, 0, 0),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000003', 50.0,  11772.50, 11772.50, 0, 0),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000004', 35.0,  10746.55, 10746.55, 0, 0),
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0001-7000-8000-000000000005', 45.0,  11394.50, 11394.50, 0, 0)
ON CONFLICT (portfolio_id, instrument_id) DO UPDATE SET
  quantity = EXCLUDED.quantity,
  cost_basis = EXCLUDED.cost_basis,
  cost_basis_base = EXCLUDED.cost_basis_base;

-- 10. Projections: 365-Day Historical Valuation Snapshots
WITH date_series AS (
  SELECT 
    s::date AS val_date,
    ROW_NUMBER() OVER (ORDER BY s ASC) - 1 AS day_idx,
    COUNT(*) OVER () - 1 AS total_days
  FROM generate_series(CURRENT_DATE - INTERVAL '365 days', CURRENT_DATE, '1 day') AS s
),
curves AS (
  SELECT
    val_date,
    day_idx,
    total_days,
    day_idx::numeric / NULLIF(total_days::numeric, 0) AS progress,
    sin((day_idx::numeric / NULLIF(total_days::numeric, 0)) * pi()) AS envelope,
    sin(day_idx::numeric * 0.12) * 1500.0 + cos(day_idx::numeric * 0.05) * 2200.0 + sin(day_idx::numeric * 0.25) * 600.0 AS wave
  FROM date_series
),
valuations AS (
  SELECT
    val_date,
    ROUND(
      (CASE 
        WHEN val_date = CURRENT_DATE THEN 8450.00
        ELSE 7500.00 + (950.00 * progress) + (envelope * sin(day_idx::numeric * 0.08) * 300.0)
      END)::numeric, 2
    ) AS cash_val,
    ROUND(
      (CASE 
        WHEN val_date = CURRENT_DATE THEN 116082.89
        ELSE 97500.00 + (18582.89 * progress) + (envelope * wave)
      END)::numeric, 2
    ) AS market_val,
    ROUND(
      (CASE
        WHEN val_date = CURRENT_DATE THEN 1.1420
        ELSE 1.0000 + (0.1420 * progress) + (envelope * wave / 100000.0)
      END)::numeric, 4
    ) AS twr
  FROM curves
),
computed AS (
  SELECT
    val_date,
    market_val,
    cash_val,
    twr,
    CASE 
      WHEN val_date = CURRENT_DATE THEN 0.010000000000
      WHEN LAG(market_val + cash_val) OVER (ORDER BY val_date) IS NULL THEN 0.000000000000
      ELSE ROUND(
        (((market_val + cash_val) - LAG(market_val + cash_val) OVER (ORDER BY val_date)) 
        / LAG(market_val + cash_val) OVER (ORDER BY val_date))::numeric,
        12
      )
    END AS ret
  FROM valuations
)
INSERT INTO portfolio.portfolio_valuations (
  portfolio_id, valuation_date, market_value_base, cash_value_base, net_flow_base, daily_return, twr_index
)
SELECT
  '018f0000-0002-7000-8000-000000000001',
  val_date,
  market_val,
  cash_val,
  0,
  ret,
  twr
FROM computed
ON CONFLICT (portfolio_id, valuation_date) DO UPDATE SET
  market_value_base = EXCLUDED.market_value_base,
  cash_value_base = EXCLUDED.cash_value_base,
  net_flow_base = EXCLUDED.net_flow_base,
  daily_return = EXCLUDED.daily_return,
  twr_index = EXCLUDED.twr_index;

-- 11. Cost Basis Fixtures for FIFO demo portfolio
-- Fixture scenario: Buy 10 @ 100, Buy 10 @ 200, Sell 10 @ 250
INSERT INTO portfolio.transactions (
  portfolio_id, instrument_id, type, trade_date, quantity, price, amount, currency_code, external_ref
) VALUES
  ('018f0000-0002-7000-8000-000000000002', '018f0000-0001-7000-8000-000000000001', 'BUY', '2025-01-10', 10.0, 100.00, 1000.00, 'USD', 'FIFO-B1'),
  ('018f0000-0002-7000-8000-000000000002', '018f0000-0001-7000-8000-000000000001', 'BUY', '2025-01-15', 10.0, 200.00, 2000.00, 'USD', 'FIFO-B2'),
  ('018f0000-0002-7000-8000-000000000002', '018f0000-0001-7000-8000-000000000001', 'SELL', '2025-01-20', 10.0, 250.00, 2500.00, 'USD', 'FIFO-S1');

COMMIT;
