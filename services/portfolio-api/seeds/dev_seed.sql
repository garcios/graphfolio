-- GraphFolio Development Seed Data

BEGIN;

-- 1. Reference Data: Currencies
INSERT INTO portfolio.currencies (code, name, minor_units) VALUES
  ('USD', 'US Dollar', 2),
  ('AUD', 'Australian Dollar', 2),
  ('EUR', 'Euro', 2),
  ('GBP', 'British Pound', 2),
  ('CAD', 'Canadian Dollar', 2),
  ('JPY', 'Japanese Yen', 0),
  ('CHF', 'Swiss Franc', 2)
ON CONFLICT (code) DO NOTHING;

-- 2. Reference Data: Exchanges
INSERT INTO portfolio.exchanges (code, name, country, timezone) VALUES
  ('XNAS', 'NASDAQ Stock Market', 'US', 'America/New_York'),
  ('XNYS', 'New York Stock Exchange', 'US', 'America/New_York'),
  ('XASX', 'Australian Securities Exchange', 'AU', 'Australia/Sydney')
ON CONFLICT (code) DO NOTHING;


-- 4. Demo Portfolios
-- Primary demo portfolio (AVERAGE_COST)
INSERT INTO portfolio.portfolios (id, user_id, name, base_currency, cost_basis_method) VALUES
  ('018f0000-0002-7000-8000-000000000001', '018f0000-0000-7000-8000-000000000001', 'Main Portfolio', 'AUD', 'AVERAGE_COST'),
  ('018f0000-0002-7000-8000-000000000002', '018f0000-0000-7000-8000-000000000001', 'FIFO Growth Portfolio', 'AUD', 'FIFO')
ON CONFLICT (user_id, name) DO NOTHING;


COMMIT;
