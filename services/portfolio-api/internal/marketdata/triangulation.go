package marketdata

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrUnresolvableFXRate = errors.New("cannot resolve fx exchange rate")
)

const (
	// DefaultFXPrecision represents the standard decimal places for FX triangulation (10 digits).
	DefaultFXPrecision = 10
	// DefaultAnchorCurrency is EUR, conforming to the European Central Bank (ECB) reference fixing standard.
	DefaultAnchorCurrency = "EUR"
)

// TriangulationEngine resolves direct, inverse, and cross-currency exchange rates
// using exact fixed-point arithmetic without floating-point representation drift.
type TriangulationEngine struct {
	mu             sync.RWMutex
	anchorCurrency string
	rates          map[string]map[string]fxRateEntry
}

type fxRateEntry struct {
	rate   decimal.Decimal
	source string
}

// NewTriangulationEngine creates an engine configured with an anchor currency (e.g. "EUR" or "USD").
func NewTriangulationEngine(anchorCurrency string) *TriangulationEngine {
	anchor := strings.ToUpper(strings.TrimSpace(anchorCurrency))
	if anchor == "" {
		anchor = DefaultAnchorCurrency
	}
	return &TriangulationEngine{
		anchorCurrency: anchor,
		rates:          make(map[string]map[string]fxRateEntry),
	}
}

// AddRate records a known exchange rate for base -> quote.
func (e *TriangulationEngine) AddRate(base, quote string, rate decimal.Decimal, source string) {
	if !rate.IsPositive() {
		return
	}
	base = strings.ToUpper(strings.TrimSpace(base))
	quote = strings.ToUpper(strings.TrimSpace(quote))
	if base == quote {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.rates[base]; !ok {
		e.rates[base] = make(map[string]fxRateEntry)
	}
	e.rates[base][quote] = fxRateEntry{rate: rate, source: source}
}

// AddRates ingests a slice of FXRecord structs into the engine's known rate matrix.
func (e *TriangulationEngine) AddRates(records []FXRecord) {
	for _, r := range records {
		e.AddRate(r.BaseCurrency, r.QuoteCurrency, r.Rate, r.Source)
	}
}

// GetRate computes or derives the exchange rate from base to quote.
// Resolution strategy:
// 1. Identity: If base == quote, returns 1.0.
// 2. Direct: If base -> quote is known, returns rate.
// 3. Inverse: If quote -> base is known, returns 1 / (quote -> base).
// 4. Cross-rate via Anchor:
//   - If Anchor -> Base and Anchor -> Quote are known:
//     Base -> Quote = (Anchor -> Quote) / (Anchor -> Base)
//   - If Base -> Anchor and Anchor -> Quote are known:
//     Base -> Quote = (Base -> Anchor) * (Anchor -> Quote)
//   - If Base -> Anchor and Quote -> Anchor are known:
//     Base -> Quote = (Base -> Anchor) / (Quote -> Anchor)
func (e *TriangulationEngine) GetRate(base, quote string) (decimal.Decimal, string, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	quote = strings.ToUpper(strings.TrimSpace(quote))

	// 1. Identity
	if base == quote {
		return decimal.NewFromInt(1), "identity", nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	// 2. Direct rate
	if quotes, ok := e.rates[base]; ok {
		if entry, found := quotes[quote]; found {
			return entry.rate, entry.source, nil
		}
	}

	// 3. Inverse rate: Quote -> Base
	if quotes, ok := e.rates[quote]; ok {
		if entry, found := quotes[base]; found && entry.rate.IsPositive() {
			inv := decimal.NewFromInt(1).DivRound(entry.rate, DefaultFXPrecision)
			src := fmt.Sprintf("inverted(%s)", entry.source)
			return inv, src, nil
		}
	}

	// 4. Triangulation via Anchor Currency
	anchor := e.anchorCurrency

	// Sub-case 4a: Base is anchor -> handled in direct/inverse above.
	// Sub-case 4b: Quote is anchor -> handled in direct/inverse above.

	// Sub-case 4c: Both are distinct from anchor.
	// Retrieve Rate(Anchor -> Base)
	rateAnchorToBase, srcAnchorToBase, errA := e.lookupDirectOrInverse(anchor, base)
	// Retrieve Rate(Anchor -> Quote)
	rateAnchorToQuote, srcAnchorToQuote, errB := e.lookupDirectOrInverse(anchor, quote)

	if errA == nil && errB == nil && rateAnchorToBase.IsPositive() {
		// Base -> Quote = (Anchor -> Quote) / (Anchor -> Base)
		cross := rateAnchorToQuote.DivRound(rateAnchorToBase, DefaultFXPrecision)
		src := fmt.Sprintf("triangulated(%s via %s: %s / %s)", anchor, quote, srcAnchorToQuote, srcAnchorToBase)
		return cross, src, nil
	}

	return decimal.Zero, "", fmt.Errorf("%w: unable to resolve rate from %s to %s via anchor %s", ErrUnresolvableFXRate, base, quote, anchor)
}

// lookupDirectOrInverse is an internal helper that searches for A -> B or inverted B -> A.
// Caller must hold e.mu.RLock.
func (e *TriangulationEngine) lookupDirectOrInverse(from, to string) (decimal.Decimal, string, error) {
	if quotes, ok := e.rates[from]; ok {
		if entry, found := quotes[to]; found {
			return entry.rate, entry.source, nil
		}
	}
	if quotes, ok := e.rates[to]; ok {
		if entry, found := quotes[from]; found && entry.rate.IsPositive() {
			return decimal.NewFromInt(1).DivRound(entry.rate, DefaultFXPrecision), fmt.Sprintf("inverted(%s)", entry.source), nil
		}
	}
	return decimal.Zero, "", fmt.Errorf("no path between %s and %s", from, to)
}

// TriangulatePairs resolves exchange rates for a requested list of currency pairs for rateDate.
func (e *TriangulationEngine) TriangulatePairs(pairs []CurrencyPair, rateDate time.Time) ([]FXRecord, error) {
	results := make([]FXRecord, 0, len(pairs))
	for _, p := range pairs {
		rate, src, err := e.GetRate(p.Base, p.Quote)
		if err != nil {
			return nil, err
		}
		results = append(results, FXRecord{
			BaseCurrency:  strings.ToUpper(strings.TrimSpace(p.Base)),
			QuoteCurrency: strings.ToUpper(strings.TrimSpace(p.Quote)),
			RateDate:      rateDate,
			Rate:          rate,
			Source:        src,
		})
	}
	return results, nil
}
