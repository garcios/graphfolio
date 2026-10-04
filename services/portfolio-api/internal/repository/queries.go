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
)
