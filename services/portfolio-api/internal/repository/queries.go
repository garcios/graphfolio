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

	listActivePortfoliosSQL = `
SELECT id
FROM portfolio.portfolios
WHERE archived_at IS NULL
ORDER BY created_at ASC;`

	updatePortfolioBaseCurrencySQL = `
UPDATE portfolio.portfolios
SET base_currency = $2, updated_at = now()
WHERE id = $1;`

	updateValuationsBaseCurrencySQL = `
UPDATE portfolio.portfolio_valuations
SET market_value_base = ROUND((market_value_base * $2)::numeric, 10),
    cash_value_base   = ROUND((cash_value_base * $2)::numeric, 10),
    net_flow_base     = ROUND((net_flow_base * $2)::numeric, 10)
WHERE portfolio_id = $1;`

	updateHoldingsBaseCurrencySQL = `
UPDATE portfolio.holdings
SET cost_basis_base   = ROUND((cost_basis_base * $2)::numeric, 10),
    realized_pnl_base = ROUND((realized_pnl_base * $2)::numeric, 10),
    dividends_base    = ROUND((dividends_base * $2)::numeric, 10)
WHERE portfolio_id = $1;`

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

	listAllInstrumentsSQL = `
SELECT id, symbol, exchange_code, name, asset_class, currency_code, isin, is_active
FROM portfolio.instruments
WHERE ($1::boolean IS NULL OR is_active = $1)
  AND ($2::text IS NULL OR symbol ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%')
ORDER BY symbol ASC;`

	listExchangesSQL = `
SELECT code, name, country, timezone
FROM portfolio.exchanges
ORDER BY code ASC;`

	createInstrumentSQL = `
INSERT INTO portfolio.instruments (
    symbol, exchange_code, name, asset_class, currency_code, isin, is_active
) VALUES (
    $1, $2, $3, $4::portfolio.asset_class, $5, $6, true
)
RETURNING id, symbol, exchange_code, name, asset_class, currency_code, isin, is_active;`

	updateInstrumentSQL = `
UPDATE portfolio.instruments
SET name = COALESCE($1, name),
    is_active = COALESCE($2, is_active),
    isin = COALESCE($3, isin),
    updated_at = now()
WHERE id = $4
RETURNING id, symbol, exchange_code, name, asset_class, currency_code, isin, is_active;`

	listInstrumentPricesSQL = `
SELECT 
    ip.instrument_id, i.symbol, ip.price_date, ip.close, i.currency_code, ip.source, ip.created_at,
    COUNT(*) OVER() AS total_count
FROM portfolio.instrument_prices ip
JOIN portfolio.instruments i ON i.id = ip.instrument_id
WHERE ($1::text IS NULL OR UPPER(i.symbol) = UPPER($1))
  AND ($2::date IS NULL OR ip.price_date >= $2)
  AND ($3::date IS NULL OR ip.price_date <= $3)
ORDER BY ip.price_date DESC, i.symbol ASC
LIMIT $4 OFFSET $5;`

	upsertInstrumentPriceSQL = `
WITH upserted AS (
    INSERT INTO portfolio.instrument_prices (instrument_id, price_date, close, source, created_at)
    VALUES ($1, $2, $3, $4, now())
    ON CONFLICT (instrument_id, price_date)
    DO UPDATE SET close = EXCLUDED.close, source = EXCLUDED.source, created_at = now()
    RETURNING instrument_id, price_date, close, source, created_at
)
SELECT u.instrument_id, i.symbol, u.price_date, u.close, i.currency_code, u.source, u.created_at
FROM upserted u
JOIN portfolio.instruments i ON i.id = u.instrument_id;`

	findPortfoliosHoldingInstrumentSQL = `
SELECT DISTINCT portfolio_id
FROM portfolio.holdings
WHERE instrument_id = $1 AND quantity > 0;`

	countActiveInstrumentsSQL = `
SELECT COUNT(*) FROM portfolio.instruments WHERE is_active = true;`

	countCurrenciesSQL = `
SELECT COUNT(*) FROM portfolio.currencies;`

	latestPriceDateSQL = `
SELECT MAX(price_date) FROM portfolio.instrument_prices;`

	latestFXDateSQL = `
SELECT MAX(rate_date) FROM portfolio.fx_rates;`

	batchUpsertInstrumentPriceWithIDSQL = `
INSERT INTO portfolio.instrument_prices (instrument_id, price_date, close, source, created_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (instrument_id, price_date)
DO UPDATE SET close = EXCLUDED.close, source = EXCLUDED.source, created_at = now();`

	batchUpsertInstrumentPriceWithSymbolSQL = `
INSERT INTO portfolio.instrument_prices (instrument_id, price_date, close, source, created_at)
SELECT id, $2, $3, $4, now()
FROM portfolio.instruments
WHERE UPPER(symbol) = UPPER($1)
ON CONFLICT (instrument_id, price_date)
DO UPDATE SET close = EXCLUDED.close, source = EXCLUDED.source, created_at = now();`

	batchUpsertFXRateSQL = `
INSERT INTO portfolio.fx_rates (base_currency, quote_currency, rate_date, rate, source)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (base_currency, quote_currency, rate_date)
DO UPDATE SET rate = EXCLUDED.rate, source = EXCLUDED.source;`

	listActiveCurrenciesSQL = `
WITH active AS (
    SELECT DISTINCT c.code
    FROM portfolio.currencies c
    WHERE EXISTS (SELECT 1 FROM portfolio.portfolios p WHERE p.base_currency = c.code)
       OR EXISTS (SELECT 1 FROM portfolio.instruments i WHERE i.currency_code = c.code AND i.is_active = true)
       OR EXISTS (SELECT 1 FROM portfolio.cash_balances cb WHERE cb.currency_code = c.code)
)
SELECT code FROM active
UNION
SELECT code FROM portfolio.currencies WHERE code IN ('USD', 'EUR', 'GBP', 'AUD', 'CAD', 'JPY', 'CHF')
ORDER BY code ASC;`

	hasPricesForRangeSQL = `
SELECT EXISTS (
    SELECT 1 FROM portfolio.instrument_prices
    WHERE instrument_id = $1
      AND price_date >= $2
      AND price_date <= $3
);`

	upsertValuationsBatchSQL = `
INSERT INTO portfolio.portfolio_valuations (
    portfolio_id, valuation_date, market_value_base, cash_value_base,
    net_flow_base, daily_return, twr_index
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (portfolio_id, valuation_date)
DO UPDATE SET
    market_value_base = EXCLUDED.market_value_base,
    cash_value_base = EXCLUDED.cash_value_base,
    net_flow_base = EXCLUDED.net_flow_base,
    daily_return = EXCLUDED.daily_return,
    twr_index = EXCLUDED.twr_index;`

	getLatestValuationBeforeSQL = `
SELECT 
    portfolio_id, valuation_date, market_value_base, cash_value_base,
    net_flow_base, COALESCE(daily_return, 0) AS daily_return, twr_index
FROM portfolio.portfolio_valuations
WHERE portfolio_id = $1 AND valuation_date < $2
ORDER BY valuation_date DESC
LIMIT 1;`

	getHistoricalPricesForMatrixSQL = `
WITH latest_before AS (
    SELECT DISTINCT ON (instrument_id) instrument_id, price_date, close
    FROM portfolio.instrument_prices
    WHERE instrument_id = ANY($1) AND price_date <= $2
    ORDER BY instrument_id, price_date DESC
),
in_range AS (
    SELECT instrument_id, price_date, close
    FROM portfolio.instrument_prices
    WHERE instrument_id = ANY($1) AND price_date >= $2 AND price_date <= $3
)
SELECT instrument_id, price_date, close FROM latest_before
UNION
SELECT instrument_id, price_date, close FROM in_range
ORDER BY instrument_id, price_date ASC;`

	getHistoricalFXForMatrixSQL = `
WITH latest_before AS (
    SELECT DISTINCT ON (base_currency, quote_currency) base_currency, quote_currency, rate_date, rate
    FROM portfolio.fx_rates
    WHERE ((base_currency = ANY($1) AND quote_currency = $2)
        OR (base_currency = $2 AND quote_currency = ANY($1)))
      AND rate_date <= $3
    ORDER BY base_currency, quote_currency, rate_date DESC
),
in_range AS (
    SELECT base_currency, quote_currency, rate_date, rate
    FROM portfolio.fx_rates
    WHERE ((base_currency = ANY($1) AND quote_currency = $2)
        OR (base_currency = $2 AND quote_currency = ANY($1)))
      AND rate_date >= $3 AND rate_date <= $4
)
SELECT base_currency, quote_currency, rate_date, rate FROM latest_before
UNION
SELECT base_currency, quote_currency, rate_date, rate FROM in_range
ORDER BY rate_date ASC;`

	deleteValuationsFromDateSQL = `
DELETE FROM portfolio.portfolio_valuations
WHERE portfolio_id = $1 AND valuation_date >= $2;`

	listCurrencyPairsSQL = `
WITH ranked_rates AS (
    SELECT 
        base_currency,
        quote_currency,
        rate_date,
        rate,
        source,
        ROW_NUMBER() OVER (
            PARTITION BY base_currency, quote_currency 
            ORDER BY rate_date DESC
        ) as rn
    FROM portfolio.fx_rates
),
pair_stats AS (
    SELECT 
        base_currency,
        quote_currency,
        COUNT(*) as total_records,
        MIN(rate_date) as first_date,
        MAX(rate_date) as last_date
    FROM portfolio.fx_rates
    GROUP BY base_currency, quote_currency
)
SELECT 
    r1.base_currency,
    r1.quote_currency,
    r1.rate as latest_rate,
    r1.rate_date as latest_date,
    r1.source as latest_source,
    r2.rate as previous_rate,
    ps.total_records,
    ps.first_date,
    ps.last_date
FROM ranked_rates r1
LEFT JOIN ranked_rates r2 
    ON r1.base_currency = r2.base_currency 
   AND r1.quote_currency = r2.quote_currency 
   AND r2.rn = 2
JOIN pair_stats ps 
    ON r1.base_currency = ps.base_currency 
   AND r1.quote_currency = ps.quote_currency
WHERE r1.rn = 1
ORDER BY r1.base_currency ASC, r1.quote_currency ASC;`

	listFXRatesSQL = `
SELECT 
    base_currency,
    quote_currency,
    rate_date,
    rate,
    source,
    COUNT(*) OVER() AS total_count
FROM portfolio.fx_rates
WHERE ($1::text IS NULL OR UPPER(base_currency) = UPPER($1))
  AND ($2::text IS NULL OR UPPER(quote_currency) = UPPER($2))
  AND ($3::date IS NULL OR rate_date >= $3)
  AND ($4::date IS NULL OR rate_date <= $4)
ORDER BY rate_date DESC, base_currency ASC, quote_currency ASC
LIMIT $5 OFFSET $6;`

	getHistoricalFXRatesSQL = `
SELECT 
    base_currency,
    quote_currency,
    rate_date,
    rate,
    source
FROM portfolio.fx_rates
WHERE UPPER(base_currency) = UPPER($1) 
  AND UPPER(quote_currency) = UPPER($2)
  AND ($3::date IS NULL OR rate_date >= $3)
  AND ($4::date IS NULL OR rate_date <= $4)
ORDER BY rate_date ASC;`

	upsertFXRateSQL = `
INSERT INTO portfolio.fx_rates (base_currency, quote_currency, rate_date, rate, source)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (base_currency, quote_currency, rate_date)
DO UPDATE SET rate = EXCLUDED.rate, source = EXCLUDED.source
RETURNING base_currency, quote_currency, rate_date, rate, source;`
)
