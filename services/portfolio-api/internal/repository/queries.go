package repository

const (
	findPortfolioByUserSQL = `
SELECT 
    id, user_id, name, base_currency, cost_basis_method, created_at
FROM portfolio.portfolios
WHERE ($1 = '1' OR user_id::text = $1 OR id::text = $1)
  AND archived_at IS NULL
ORDER BY created_at ASC
LIMIT 1;`

	getHoldingsWithMarketDataSQL = `
SELECT 
    h.portfolio_id,
    h.instrument_id,
    i.symbol,
    i.name,
    i.currency_code AS instrument_currency,
    h.quantity,
    h.cost_basis,
    h.cost_basis_base,
    h.realized_pnl_base,
    h.dividends_base,
    COALESCE(curr_p.close, 0) AS latest_price,
    COALESCE(prev_p.close, curr_p.close, 0) AS prev_price,
    COALESCE(
        direct_fx.rate,
        CASE WHEN inverse_fx.rate > 0 THEN 1.0 / inverse_fx.rate ELSE NULL END,
        1.0
    ) AS fx_rate_to_base
FROM portfolio.holdings h
JOIN portfolio.instruments i ON i.id = h.instrument_id
JOIN portfolio.portfolios p ON p.id = h.portfolio_id
-- Latest price (most recent trading day)
LEFT JOIN LATERAL (
    SELECT close, price_date
    FROM portfolio.instrument_prices 
    WHERE instrument_id = h.instrument_id 
    ORDER BY price_date DESC LIMIT 1
) curr_p ON true
-- Previous price (second most recent trading day)
LEFT JOIN LATERAL (
    SELECT close 
    FROM portfolio.instrument_prices 
    WHERE instrument_id = h.instrument_id 
      AND price_date < curr_p.price_date
    ORDER BY price_date DESC LIMIT 1
) prev_p ON true
-- FX direct rate (instrument -> base)
LEFT JOIN LATERAL (
    SELECT rate 
    FROM portfolio.fx_rates 
    WHERE base_currency = i.currency_code 
      AND quote_currency = p.base_currency 
    ORDER BY rate_date DESC LIMIT 1
) direct_fx ON (i.currency_code <> p.base_currency)
-- FX inverse rate (base -> instrument)
LEFT JOIN LATERAL (
    SELECT rate 
    FROM portfolio.fx_rates 
    WHERE base_currency = p.base_currency 
      AND quote_currency = i.currency_code 
    ORDER BY rate_date DESC LIMIT 1
) inverse_fx ON (i.currency_code <> p.base_currency)
WHERE h.portfolio_id = $1 AND h.quantity > 0
ORDER BY (h.quantity * COALESCE(curr_p.close, 0)) DESC;`

	getCashBalancesSQL = `
SELECT 
    portfolio_id, currency_code, balance
FROM portfolio.cash_balances
WHERE portfolio_id = $1
ORDER BY currency_code;`

	getLatestValuationSQL = `
SELECT 
    portfolio_id, valuation_date, market_value_base, cash_value_base, net_flow_base, daily_return, twr_index
FROM portfolio.portfolio_valuations
WHERE portfolio_id = $1
ORDER BY valuation_date DESC
LIMIT 1;`

	getPortfolioValuationsSQL = `
SELECT 
    portfolio_id, valuation_date, market_value_base, cash_value_base,
    net_flow_base, COALESCE(daily_return, 0) AS daily_return, twr_index
FROM portfolio.portfolio_valuations
WHERE portfolio_id = $1 AND valuation_date >= $2
ORDER BY valuation_date ASC;`

	getCashFXRatesSQL = `
SELECT DISTINCT ON (base_currency) 
    base_currency, rate
FROM portfolio.fx_rates
WHERE quote_currency = $1
ORDER BY base_currency, rate_date DESC;`

	getTransactionsSQL = `
SELECT 
    id, portfolio_id, instrument_id, type, trade_date, quantity, price, amount,
    currency_code, fee, withholding_tax, fx_rate_to_base, external_ref, notes
FROM portfolio.transactions
WHERE portfolio_id = $1
ORDER BY trade_date ASC, id ASC;`

	getCorporateActionsSQL = `
SELECT 
    id, instrument_id, type, ex_date, ratio_from, ratio_to, new_instrument_id,
    cash_amount, currency_code, cost_basis_pct
FROM portfolio.corporate_actions
WHERE instrument_id = ANY($1)
ORDER BY ex_date ASC, id ASC;`

	insertTransactionSQL = `
INSERT INTO portfolio.transactions (
    portfolio_id, instrument_id, type, trade_date, quantity, price, amount,
    currency_code, fee, withholding_tax, fx_rate_to_base, external_ref, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING id;`

	findInstrumentBySymbolSQL = `
SELECT id, symbol, exchange_code, name, asset_class, currency_code, isin, is_active
FROM portfolio.instruments
WHERE UPPER(symbol) = UPPER($1) AND is_active = true
LIMIT 1;`

	listActiveInstrumentsSQL = `
SELECT id, symbol, exchange_code, name, asset_class, currency_code, isin, is_active
FROM portfolio.instruments
WHERE is_active = true
ORDER BY symbol ASC;`

	getDirectFXRateSQL = `
SELECT rate
FROM portfolio.fx_rates
WHERE base_currency = $1 AND quote_currency = $2
ORDER BY rate_date DESC
LIMIT 1;`

	getInverseFXRateSQL = `
SELECT rate
FROM portfolio.fx_rates
WHERE base_currency = $2 AND quote_currency = $1
ORDER BY rate_date DESC
LIMIT 1;`

	listTransactionsSQL = `
SELECT 
    t.id, t.portfolio_id, t.instrument_id, t.type, t.trade_date,
    t.quantity, t.price, t.amount, t.currency_code, t.fee,
    t.withholding_tax, t.fx_rate_to_base, t.external_ref, t.notes,
    t.created_at, i.symbol, i.name AS instrument_name,
    COUNT(*) OVER() AS total_count
FROM portfolio.transactions t
LEFT JOIN portfolio.instruments i ON i.id = t.instrument_id
WHERE t.portfolio_id = $1
  AND ($2::text IS NULL OR t.type::text = $2)
  AND ($3::text IS NULL OR UPPER(i.symbol) = UPPER($3))
ORDER BY t.trade_date DESC, t.created_at DESC, t.id DESC
LIMIT $4 OFFSET $5;`

	deleteTransactionSQL = `
DELETE FROM portfolio.transactions
WHERE id = $1 AND portfolio_id = $2;`
)
