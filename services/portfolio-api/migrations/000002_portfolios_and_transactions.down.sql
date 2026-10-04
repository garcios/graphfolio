-- 000002_portfolios_and_transactions.down.sql

DROP TABLE IF EXISTS portfolio.transactions;
DROP TABLE IF EXISTS portfolio.portfolios;
DROP TYPE IF EXISTS portfolio.transaction_type;
DROP TYPE IF EXISTS portfolio.cost_basis_method;
