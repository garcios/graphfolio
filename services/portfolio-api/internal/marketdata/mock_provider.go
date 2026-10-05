package marketdata

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// MockPriceProvider is a thread-safe implementation of PriceProvider for deterministic unit tests.
type MockPriceProvider struct {
	mu           sync.RWMutex
	name         string
	prices       map[string]map[string]decimal.Decimal // symbol -> YYYY-MM-DD -> price
	defaultPrice decimal.Decimal
	err          error
}

// NewMockPriceProvider creates a mock price provider with an optional provider name.
func NewMockPriceProvider(name string) *MockPriceProvider {
	if name == "" {
		name = "MockPriceProvider"
	}
	return &MockPriceProvider{
		name:         name,
		prices:       make(map[string]map[string]decimal.Decimal),
		defaultPrice: decimal.NewFromInt(100),
	}
}

func (m *MockPriceProvider) Name() string {
	return m.name
}

// SetPrice assigns an explicit price for a symbol on a specific date.
func (m *MockPriceProvider) SetPrice(symbol string, date time.Time, price decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	dateKey := date.Format("2006-01-02")
	if _, ok := m.prices[sym]; !ok {
		m.prices[sym] = make(map[string]decimal.Decimal)
	}
	m.prices[sym][dateKey] = price
}

// SetDefaultPrice sets the fallback price returned when an explicit date price is not set.
func (m *MockPriceProvider) SetDefaultPrice(price decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultPrice = price
}

// SetError instructs the mock to return an error on subsequent calls.
func (m *MockPriceProvider) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *MockPriceProvider) FetchClosingPrices(ctx context.Context, instruments []InstrumentRef, date time.Time) ([]PriceRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.err != nil {
		return nil, m.err
	}

	dateKey := date.Format("2006-01-02")
	records := make([]PriceRecord, 0, len(instruments))

	for _, inst := range instruments {
		sym := strings.ToUpper(strings.TrimSpace(inst.Symbol))
		price := m.defaultPrice
		if dateMap, ok := m.prices[sym]; ok {
			if explicitPrice, found := dateMap[dateKey]; found {
				price = explicitPrice
			}
		}

		if price.IsPositive() {
			records = append(records, PriceRecord{
				InstrumentID: inst.ID,
				Symbol:       sym,
				Exchange:     inst.Exchange,
				PriceDate:    date,
				ClosePrice:   price,
				Source:       m.name,
			})
		}
	}

	return records, nil
}

func (m *MockPriceProvider) FetchHistoricalPrices(ctx context.Context, instrument InstrumentRef, from, to time.Time) ([]PriceRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.err != nil {
		return nil, m.err
	}

	if from.After(to) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}

	sym := strings.ToUpper(strings.TrimSpace(instrument.Symbol))
	var records []PriceRecord

	current := from
	for !current.After(to) {
		// Skip weekends (Saturday, Sunday)
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			dateKey := current.Format("2006-01-02")
			price := m.defaultPrice
			if dateMap, ok := m.prices[sym]; ok {
				if explicitPrice, found := dateMap[dateKey]; found {
					price = explicitPrice
				}
			}

			if price.IsPositive() {
				records = append(records, PriceRecord{
					InstrumentID: instrument.ID,
					Symbol:       sym,
					Exchange:     instrument.Exchange,
					PriceDate:    current,
					ClosePrice:   price,
					Source:       m.name,
				})
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return records, nil
}

// MockFXRateProvider is a thread-safe implementation of FXRateProvider for deterministic unit tests.
type MockFXRateProvider struct {
	mu     sync.RWMutex
	name   string
	rates  map[string]map[string]decimal.Decimal // "BASE/QUOTE" -> YYYY-MM-DD -> rate
	err    error
	engine *TriangulationEngine
}

// NewMockFXRateProvider creates a mock FX provider.
func NewMockFXRateProvider(name string) *MockFXRateProvider {
	if name == "" {
		name = "MockFXRateProvider"
	}
	return &MockFXRateProvider{
		name:   name,
		rates:  make(map[string]map[string]decimal.Decimal),
		engine: NewTriangulationEngine(DefaultAnchorCurrency),
	}
}

func (m *MockFXRateProvider) Name() string {
	return m.name
}

// SetRate assigns an explicit rate for a currency pair on a specific date.
func (m *MockFXRateProvider) SetRate(base, quote string, date time.Time, rate decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	pair := FormatCurrencyPair(base, quote)
	dateKey := date.Format("2006-01-02")
	if _, ok := m.rates[pair]; !ok {
		m.rates[pair] = make(map[string]decimal.Decimal)
	}
	m.rates[pair][dateKey] = rate
	m.engine.AddRate(base, quote, rate, m.name)
}

// SetError instructs the mock to return an error on subsequent calls.
func (m *MockFXRateProvider) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *MockFXRateProvider) FetchFXRates(ctx context.Context, pairs []CurrencyPair, date time.Time) ([]FXRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.err != nil {
		return nil, m.err
	}

	dateKey := date.Format("2006-01-02")
	records := make([]FXRecord, 0, len(pairs))

	for _, p := range pairs {
		pairKey := FormatCurrencyPair(p.Base, p.Quote)
		if dateMap, ok := m.rates[pairKey]; ok {
			if explicitRate, found := dateMap[dateKey]; found {
				records = append(records, FXRecord{
					BaseCurrency:  strings.ToUpper(strings.TrimSpace(p.Base)),
					QuoteCurrency: strings.ToUpper(strings.TrimSpace(p.Quote)),
					RateDate:      date,
					Rate:          explicitRate,
					Source:        m.name,
				})
				continue
			}
		}

		// Fallback to triangulation engine
		rate, src, err := m.engine.GetRate(p.Base, p.Quote)
		if err != nil {
			return nil, fmt.Errorf("mock fx rate not configured for %s on %s: %w", pairKey, dateKey, err)
		}

		records = append(records, FXRecord{
			BaseCurrency:  strings.ToUpper(strings.TrimSpace(p.Base)),
			QuoteCurrency: strings.ToUpper(strings.TrimSpace(p.Quote)),
			RateDate:      date,
			Rate:          rate,
			Source:        src,
		})
	}

	return records, nil
}

func (m *MockFXRateProvider) FetchHistoricalFXRates(ctx context.Context, pair CurrencyPair, from, to time.Time) ([]FXRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.err != nil {
		return nil, m.err
	}

	if from.After(to) {
		return nil, fmt.Errorf("from date cannot be after to date")
	}

	var records []FXRecord
	current := from
	for !current.After(to) {
		recs, err := m.FetchFXRates(ctx, []CurrencyPair{pair}, current)
		if err != nil {
			return nil, err
		}
		records = append(records, recs...)
		current = current.AddDate(0, 0, 1)
	}

	return records, nil
}
