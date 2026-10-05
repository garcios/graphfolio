package marketdata

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrEmptySymbol         = errors.New("symbol cannot be empty")
	ErrInvalidPrice        = errors.New("closing price must be greater than zero")
	ErrZeroDate            = errors.New("date cannot be zero")
	ErrFutureDate          = errors.New("price date cannot be in the future")
	ErrEmptySource         = errors.New("source cannot be empty")
	ErrInvalidCurrencyCode = errors.New("currency code must be exactly 3 characters")
	ErrIdenticalCurrencies = errors.New("base and quote currencies cannot be identical")
	ErrInvalidFXRate       = errors.New("fx rate must be greater than zero")
)

// ValidatePriceRecord verifies that a price record complies with zero-drift financial integrity rules.
func ValidatePriceRecord(rec PriceRecord) error {
	if strings.TrimSpace(rec.Symbol) == "" {
		return ErrEmptySymbol
	}
	if !rec.ClosePrice.IsPositive() {
		return ErrInvalidPrice
	}
	if rec.PriceDate.IsZero() {
		return ErrZeroDate
	}
	// Check against future calendar day in UTC
	todayUTC := time.Now().UTC().Truncate(24 * time.Hour)
	if rec.PriceDate.UTC().Truncate(24 * time.Hour).After(todayUTC) {
		return ErrFutureDate
	}
	if strings.TrimSpace(rec.Source) == "" {
		return ErrEmptySource
	}
	return nil
}

// ValidateFXRecord verifies that an FX fixing rate complies with currency standards.
func ValidateFXRecord(rec FXRecord) error {
	base := strings.TrimSpace(rec.BaseCurrency)
	quote := strings.TrimSpace(rec.QuoteCurrency)

	if len(base) != 3 || len(quote) != 3 {
		return ErrInvalidCurrencyCode
	}
	if strings.EqualFold(base, quote) {
		return ErrIdenticalCurrencies
	}
	if !rec.Rate.IsPositive() {
		return ErrInvalidFXRate
	}
	if rec.RateDate.IsZero() {
		return ErrZeroDate
	}
	todayUTC := time.Now().UTC().Truncate(24 * time.Hour)
	if rec.RateDate.UTC().Truncate(24 * time.Hour).After(todayUTC) {
		return ErrFutureDate
	}
	if strings.TrimSpace(rec.Source) == "" {
		return ErrEmptySource
	}
	return nil
}

// ValidatePriceSpike evaluates whether a new closing mark deviates from the previous close
// by more than maxThresholdPercent (e.g. 50%). Returns whether a spike was detected and the percentage change.
// Percentage is calculated as: |newClose - prevClose| / prevClose * 100.
func ValidatePriceSpike(prevClose, newClose, maxThresholdPercent decimal.Decimal) (bool, decimal.Decimal) {
	if !prevClose.IsPositive() || !newClose.IsPositive() {
		return false, decimal.Zero
	}

	diff := newClose.Sub(prevClose).Abs()
	pctChange := diff.Div(prevClose).Mul(decimal.NewFromInt(100))

	if pctChange.GreaterThan(maxThresholdPercent) {
		return true, pctChange
	}
	return false, pctChange
}

// FormatCurrencyPair returns canonical "BASE/QUOTE" format in uppercase.
func FormatCurrencyPair(base, quote string) string {
	return fmt.Sprintf("%s/%s", strings.ToUpper(strings.TrimSpace(base)), strings.ToUpper(strings.TrimSpace(quote)))
}
