package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"pkg/database"
	"portfolio-api/internal/domain"
	"portfolio-api/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaxableIncomeSummary_RatioPercentages tests Test Case D in isolation:
// verifying that ratio bar percentages in TaxableIncomeSummary correctly sum to 100%.
func TestTaxableIncomeSummary_RatioPercentages(t *testing.T) {
	// Case D: Div/Int $700.00, CGT $500.00, Franking $300.00 -> Total $1,500.00
	divInt := decimal.NewFromFloat(700.00)
	cgt := decimal.NewFromFloat(500.00)
	franking := decimal.NewFromFloat(300.00)

	summary := domain.CalculateTaxableIncomeSummary(divInt, cgt, franking)

	assert.True(t, decimal.NewFromFloat(1500.00).Equal(summary.TotalAssessableIncome), "Total assessable income should be $1500.00")
	assert.True(t, decimal.NewFromFloat(46.67).Equal(summary.DividendInterestRatioPct), "Div/Int ratio should be 46.67%")
	assert.True(t, decimal.NewFromFloat(33.33).Equal(summary.CapitalGainsRatioPct), "CGT ratio should be 33.33%")
	assert.True(t, decimal.NewFromFloat(20.00).Equal(summary.FrankingCreditsRatioPct), "Franking ratio should be 20.00%")

	totalPct := summary.DividendInterestRatioPct.Add(summary.CapitalGainsRatioPct).Add(summary.FrankingCreditsRatioPct)
	assert.True(t, decimal.NewFromInt(100).Equal(totalPct), "Ratios must sum to exactly 100.00%")
}

// TestGetTaxReportIncomeAndOffsets_Integration runs live integration tests against PostgreSQL
// covering Test Cases A, B, and C for the Two-Tier Waterfall resolution.
func TestGetTaxReportIncomeAndOffsets_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_PORTFOLIO_DB_URL")
	if dsn == "" {
		dsn = os.Getenv("PORTFOLIO_DB_URL")
	}
	if dsn == "" {
		t.Skip("No PORTFOLIO_DB_URL set; skipping live repository integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(dsn))
	if err != nil {
		t.Skipf("cannot connect to postgres (%v); skipping", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)

	// Setup: create test portfolio
	portfolioID := uuid.New()
	userID := "test-user-" + uuid.NewString()[:8]
	_, err = pool.Exec(ctx, `
		INSERT INTO portfolio.portfolios (id, user_id, name, base_currency)
		VALUES ($1, $2, 'Tax Integration Test', 'AUD')
	`, portfolioID, userID)
	require.NoError(t, err)

	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM portfolio.transactions WHERE portfolio_id = $1`, portfolioID)
		_, _ = pool.Exec(ctx, `DELETE FROM portfolio.portfolios WHERE id = $1`, portfolioID)
	}()

	// Setup: find or create an ASX instrument
	instrumentID := uuid.New()
	symbol := "TEST-" + uuid.NewString()[:6]
	_, err = pool.Exec(ctx, `
		INSERT INTO portfolio.instruments (id, symbol, name, currency_code, exchange_mic)
		VALUES ($1, $2, 'Test ASX Asset', 'AUD', 'XASX')
	`, instrumentID, symbol)
	require.NoError(t, err)

	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM portfolio.dividend_schedules WHERE instrument_id = $1`, instrumentID)
		_, _ = pool.Exec(ctx, `DELETE FROM portfolio.instruments WHERE id = $1`, instrumentID)
	}()

	fromDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)

	// Test Case A: Fully franked dividend with schedule fallback (Tier 2)
	// Schedule: Ex-date 2024-09-10, Payment date 2024-09-25, franking_percentage = 100.00%
	// Transaction: Trade date 2024-09-25, Amount $700.00, tx.franking_credit = 0
	// Expected: Franking Credit = 700.00 * (100 / 100) * (0.30 / 0.70) = 300.00
	t.Run("TestCaseA_FullyFranked_ScheduleFallback", func(t *testing.T) {
		scheduleID := uuid.New()
		exDate := time.Date(2024, 9, 10, 0, 0, 0, 0, time.UTC)
		payDate := time.Date(2024, 9, 25, 0, 0, 0, 0, time.UTC)
		_, err := pool.Exec(ctx, `
			INSERT INTO portfolio.dividend_schedules 
			(id, instrument_id, ex_date, payment_date, record_date, dividend_per_share, currency_code, franking_percentage, status, is_estimated, source)
			VALUES ($1, $2, $3, $4, $3, 1.50, 'AUD', 100.00, 'PAID', false, 'test')
		`, scheduleID, instrumentID, exDate, payDate)
		require.NoError(t, err)
		defer func() {
			_, _ = pool.Exec(ctx, `DELETE FROM portfolio.dividend_schedules WHERE id = $1`, scheduleID)
		}()

		txID := uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO portfolio.transactions
			(id, portfolio_id, instrument_id, type, trade_date, amount, fee, franking_credit, franking_pct)
			VALUES ($1, $2, $3, 'DIVIDEND', $4, 700.00, 0.00, 0.00, NULL)
		`, txID, portfolioID, instrumentID, payDate)
		require.NoError(t, err)
		defer func() {
			_, _ = pool.Exec(ctx, `DELETE FROM portfolio.transactions WHERE id = $1`, txID)
		}()

		divInc, intInc, frankingCredits, wht, err := repo.GetTaxReportIncomeAndOffsets(ctx, portfolioID, fromDate, toDate)
		require.NoError(t, err)

		assert.True(t, decimal.NewFromFloat(700.00).Equal(divInc), "Dividend income should be $700.00")
		assert.True(t, decimal.Zero.Equal(intInc))
		assert.True(t, decimal.Zero.Equal(wht))
		assert.True(t, decimal.NewFromFloat(300.00).Equal(frankingCredits), "Franking credit should be $300.00 (from 100%% schedule fallback)")
	})

	// Test Case B: Partially franked dividend with explicit tx.franking_credit (Tier 1 override)
	// Schedule: Ex-date 2024-10-15, payment 2024-10-30, franking_percentage = 100.00% (would yield $300.00 if used)
	// Transaction: Trade date 2024-10-30, Amount $700.00, explicit tx.franking_credit = $150.00
	// Expected: Franking Credit = 150.00 (Tier 1 takes precedence)
	t.Run("TestCaseB_ExplicitFrankingCredit_StatementOverride", func(t *testing.T) {
		scheduleID := uuid.New()
		exDate := time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC)
		payDate := time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC)
		_, err := pool.Exec(ctx, `
			INSERT INTO portfolio.dividend_schedules 
			(id, instrument_id, ex_date, payment_date, record_date, dividend_per_share, currency_code, franking_percentage, status, is_estimated, source)
			VALUES ($1, $2, $3, $4, $3, 2.00, 'AUD', 100.00, 'PAID', false, 'test')
		`, scheduleID, instrumentID, exDate, payDate)
		require.NoError(t, err)
		defer func() {
			_, _ = pool.Exec(ctx, `DELETE FROM portfolio.dividend_schedules WHERE id = $1`, scheduleID)
		}()

		txID := uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO portfolio.transactions
			(id, portfolio_id, instrument_id, type, trade_date, amount, fee, franking_credit, franking_pct)
			VALUES ($1, $2, $3, 'DIVIDEND', $4, 700.00, 0.00, 150.00, 50.00)
		`, txID, portfolioID, instrumentID, payDate)
		require.NoError(t, err)
		defer func() {
			_, _ = pool.Exec(ctx, `DELETE FROM portfolio.transactions WHERE id = $1`, txID)
		}()

		divInc, _, frankingCredits, _, err := repo.GetTaxReportIncomeAndOffsets(ctx, portfolioID, fromDate, toDate)
		require.NoError(t, err)

		assert.True(t, decimal.NewFromFloat(700.00).Equal(divInc))
		assert.True(t, decimal.NewFromFloat(150.00).Equal(frankingCredits), "Franking credit should be $150.00 from explicit statement entry")
	})

	// Test Case C: Dividend with settlement lag (4 days after ex-date, verifying 45-day window match)
	// Schedule: Ex-date 2024-11-01, payment_date is unset/null or far out
	// Transaction: Trade date 2024-11-05 (4 days after ex-date), Amount $1400.00, tx.franking_credit = 0
	// Expected: Schedule matches via (ex_date <= tx.trade_date AND tx.trade_date <= ex_date + INTERVAL '45 days')
	// Franking Credit = 1400.00 * (100 / 100) * (0.30 / 0.70) = 600.00
	t.Run("TestCaseC_SettlementLag_WindowMatch", func(t *testing.T) {
		scheduleID := uuid.New()
		exDate := time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)
		payDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		_, err := pool.Exec(ctx, `
			INSERT INTO portfolio.dividend_schedules 
			(id, instrument_id, ex_date, payment_date, record_date, dividend_per_share, currency_code, franking_percentage, status, is_estimated, source)
			VALUES ($1, $2, $3, $4, $3, 2.50, 'AUD', 100.00, 'PAID', false, 'test')
		`, scheduleID, instrumentID, exDate, payDate)
		require.NoError(t, err)
		defer func() {
			_, _ = pool.Exec(ctx, `DELETE FROM portfolio.dividend_schedules WHERE id = $1`, scheduleID)
		}()

		txID := uuid.New()
		tradeDate := time.Date(2024, 11, 5, 0, 0, 0, 0, time.UTC) // 4 days lag from exDate
		_, err = pool.Exec(ctx, `
			INSERT INTO portfolio.transactions
			(id, portfolio_id, instrument_id, type, trade_date, amount, fee, franking_credit, franking_pct)
			VALUES ($1, $2, $3, 'DIVIDEND', $4, 1400.00, 0.00, 0.00, NULL)
		`, txID, portfolioID, instrumentID, tradeDate)
		require.NoError(t, err)
		defer func() {
			_, _ = pool.Exec(ctx, `DELETE FROM portfolio.transactions WHERE id = $1`, txID)
		}()

		divInc, _, frankingCredits, _, err := repo.GetTaxReportIncomeAndOffsets(ctx, portfolioID, fromDate, toDate)
		require.NoError(t, err)

		assert.True(t, decimal.NewFromFloat(1400.00).Equal(divInc))
		assert.True(t, decimal.NewFromFloat(600.00).Equal(frankingCredits), "Franking credit should be $600.00 within 45-day window")
	})
}
