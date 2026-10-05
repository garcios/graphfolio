package marketdata

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const (
	defaultECBBaseURL   = "https://www.ecb.europa.eu/stats/eurofxref"
	defaultECBDailyURL  = defaultECBBaseURL + "/eurofxref-daily.xml"
	defaultECBHist90URL = defaultECBBaseURL + "/eurofxref-hist-90d.xml"
	ecbSourceIdentifier = "ecb"
)

var (
	ErrNoECBRatesFound     = errors.New("no ecb fx rates found for the requested date")
	ErrInvalidECBXMLFormat = errors.New("unexpected ecb xml schema structure")
)

// ECBProviderConfig holds configuration parameters for the ECB FX rate provider.
type ECBProviderConfig struct {
	BaseURL    string
	DailyURL   string
	Hist90dURL string
	HTTPClient *http.Client
	Limiter    *RateLimiter
	RetryCfg   RetryConfig
	CacheTTL   time.Duration
}

// ECBProvider implements FXRateProvider against the European Central Bank XML/SDMX feed.
type ECBProvider struct {
	dailyURL   string
	hist90dURL string
	httpClient *http.Client
	limiter    *RateLimiter
	retryCfg   RetryConfig
	cacheTTL   time.Duration

	mu            sync.RWMutex
	cachedDailyAt time.Time
	cachedDaily   map[string][]FXRecord
	cachedHistAt  time.Time
	cachedHist    map[string][]FXRecord
}

// NewECBProvider constructs a configured ECB FXRateProvider.
func NewECBProvider(cfg ECBProviderConfig) *ECBProvider {
	daily := cfg.DailyURL
	if daily == "" {
		if cfg.BaseURL != "" {
			daily = strings.TrimRight(cfg.BaseURL, "/") + "/eurofxref-daily.xml"
		} else {
			daily = defaultECBDailyURL
		}
	}

	hist := cfg.Hist90dURL
	if hist == "" {
		if cfg.BaseURL != "" {
			hist = strings.TrimRight(cfg.BaseURL, "/") + "/eurofxref-hist-90d.xml"
		} else {
			hist = defaultECBHist90URL
		}
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	retryCfg := cfg.RetryCfg
	if retryCfg.MaxRetries == 0 && retryCfg.InitialBackoff == 0 {
		retryCfg = DefaultRetryConfig()
	}

	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	return &ECBProvider{
		dailyURL:   daily,
		hist90dURL: hist,
		httpClient: client,
		limiter:    cfg.Limiter,
		retryCfg:   retryCfg,
		cacheTTL:   ttl,
	}
}

// Name returns the provider's human-readable identifier.
func (p *ECBProvider) Name() string {
	return "European Central Bank"
}

// XML mapping structs for ECB Euro Foreign Exchange Reference Rates.
type ecbEnvelope struct {
	XMLName xml.Name    `xml:"Envelope"`
	Root    ecbRootCube `xml:"Cube"`
}

type ecbRootCube struct {
	TimeCubes []ecbTimeCube `xml:"Cube"`
}

type ecbTimeCube struct {
	Time  string        `xml:"time,attr"`
	Rates []ecbRateCube `xml:"Cube"`
}

type ecbRateCube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}

// ParseECBXML parses raw ECB XML bytes into a map of date string (YYYY-MM-DD) to base EUR FXRecords.
func ParseECBXML(data []byte) (map[string][]FXRecord, error) {
	var env ecbEnvelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidECBXMLFormat, err)
	}

	result := make(map[string][]FXRecord)

	for _, timeCube := range env.Root.TimeCubes {
		dateStr := strings.TrimSpace(timeCube.Time)
		if dateStr == "" {
			continue
		}
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		var dateRecords []FXRecord
		for _, rateCube := range timeCube.Rates {
			curr := strings.ToUpper(strings.TrimSpace(rateCube.Currency))
			rateStr := strings.TrimSpace(rateCube.Rate)
			if curr == "" || rateStr == "" {
				continue
			}

			rateDec, err := decimal.NewFromString(rateStr)
			if err != nil || !rateDec.IsPositive() {
				continue
			}

			rec := FXRecord{
				BaseCurrency:  "EUR",
				QuoteCurrency: curr,
				RateDate:      parsedDate.UTC(),
				Rate:          rateDec,
				Source:        ecbSourceIdentifier,
			}
			if err := ValidateFXRecord(rec); err == nil {
				dateRecords = append(dateRecords, rec)
			}
		}

		if len(dateRecords) > 0 {
			result[dateStr] = dateRecords
		}
	}

	if len(result) == 0 {
		return nil, ErrNoECBRatesFound
	}

	return result, nil
}

// fetchAndParseFeed fetches an XML endpoint with rate limiting, retries, and unmarshaling.
func (p *ECBProvider) fetchAndParseFeed(ctx context.Context, url string) (map[string][]FXRecord, error) {
	var bodyBytes []byte

	err := ExecuteWithRetry(ctx, p.retryCfg, func(attempt int) error {
		if p.limiter != nil {
			if err := p.limiter.Wait(ctx); err != nil {
				return err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "GraphFolio-MarketData/1.0")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			return &HTTPStatusError{StatusCode: http.StatusServiceUnavailable, Status: err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return &HTTPStatusError{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Body:       string(bodySnippet),
			}
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyBytes = b
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch ecb feed from %s: %w", url, err)
	}

	return ParseECBXML(bodyBytes)
}

// getDailyRates returns the latest daily feed, checking memory cache first.
func (p *ECBProvider) getDailyRates(ctx context.Context) (map[string][]FXRecord, error) {
	p.mu.RLock()
	if len(p.cachedDaily) > 0 && time.Since(p.cachedDailyAt) < p.cacheTTL {
		cached := p.cachedDaily
		p.mu.RUnlock()
		return cached, nil
	}
	p.mu.RUnlock()

	parsed, err := p.fetchAndParseFeed(ctx, p.dailyURL)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.cachedDaily = parsed
	p.cachedDailyAt = time.Now()
	p.mu.Unlock()

	return parsed, nil
}

// getHistoricalRates returns the 90-day historical feed, checking memory cache first.
func (p *ECBProvider) getHistoricalRates(ctx context.Context) (map[string][]FXRecord, error) {
	p.mu.RLock()
	if len(p.cachedHist) > 0 && time.Since(p.cachedHistAt) < p.cacheTTL {
		cached := p.cachedHist
		p.mu.RUnlock()
		return cached, nil
	}
	p.mu.RUnlock()

	parsed, err := p.fetchAndParseFeed(ctx, p.hist90dURL)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.cachedHist = parsed
	p.cachedHistAt = time.Now()
	p.mu.Unlock()

	return parsed, nil
}

// findLatestObservationDate finds the closest available date in the map on or before targetDate.
func findLatestObservationDate(recordsByDate map[string][]FXRecord, targetDate time.Time) (string, []FXRecord, bool) {
	if len(recordsByDate) == 0 {
		return "", nil, false
	}

	targetStr := targetDate.UTC().Format("2006-01-02")
	if recs, ok := recordsByDate[targetStr]; ok {
		return targetStr, recs, true
	}

	// Collect and sort all available dates descending
	dates := make([]string, 0, len(recordsByDate))
	for d := range recordsByDate {
		dates = append(dates, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// If targetDate is zero, return the absolute newest date
	if targetDate.IsZero() {
		newest := dates[0]
		return newest, recordsByDate[newest], true
	}

	// Carry forward the latest observation on or before targetDate (LOCF)
	for _, d := range dates {
		if d <= targetStr {
			return d, recordsByDate[d], true
		}
	}

	return "", nil, false
}

// FetchFXRates retrieves daily fixing rates for a slice of currency pairs on a given date.
// If requested date is during a weekend or holiday, the last available trading day is carried forward.
func (p *ECBProvider) FetchFXRates(ctx context.Context, pairs []CurrencyPair, date time.Time) ([]FXRecord, error) {
	if len(pairs) == 0 {
		return nil, nil
	}

	// If date is within the last 5 days or zero, try daily feed first
	var recordsByDate map[string][]FXRecord
	var err error

	todayUTC := time.Now().UTC().Truncate(24 * time.Hour)
	isRecent := date.IsZero() || date.UTC().After(todayUTC.AddDate(0, 0, -5))

	if isRecent {
		recordsByDate, err = p.getDailyRates(ctx)
		if err != nil {
			// Fallback to hist feed if daily fails
			recordsByDate, err = p.getHistoricalRates(ctx)
		}
	} else {
		recordsByDate, err = p.getHistoricalRates(ctx)
	}

	if err != nil {
		return nil, err
	}

	effectiveDateStr, baseRecords, found := findLatestObservationDate(recordsByDate, date)
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrNoECBRatesFound, date.Format("2006-01-02"))
	}

	effectiveDate, _ := time.Parse("2006-01-02", effectiveDateStr)
	effectiveDate = effectiveDate.UTC()

	// Build triangulation engine with the EUR base rates for this date
	engine := NewTriangulationEngine(DefaultAnchorCurrency)
	engine.AddRates(baseRecords)

	results := make([]FXRecord, 0, len(pairs))
	for _, pair := range pairs {
		base := strings.ToUpper(strings.TrimSpace(pair.Base))
		quote := strings.ToUpper(strings.TrimSpace(pair.Quote))

		rate, _, err := engine.GetRate(base, quote)
		if err != nil {
			continue
		}

		results = append(results, FXRecord{
			BaseCurrency:  base,
			QuoteCurrency: quote,
			RateDate:      effectiveDate,
			Rate:          rate,
			Source:        ecbSourceIdentifier,
		})
	}

	return results, nil
}

// FetchHistoricalFXRates retrieves daily fixing rates for a currency pair across an inclusive date range [from, to].
func (p *ECBProvider) FetchHistoricalFXRates(ctx context.Context, pair CurrencyPair, from, to time.Time) ([]FXRecord, error) {
	fromUTC := from.UTC().Truncate(24 * time.Hour)
	toUTC := to.UTC().Truncate(24 * time.Hour)
	if toUTC.Before(fromUTC) {
		return nil, errors.New("to date cannot be before from date")
	}

	recordsByDate, err := p.getHistoricalRates(ctx)
	if err != nil {
		return nil, err
	}

	base := strings.ToUpper(strings.TrimSpace(pair.Base))
	quote := strings.ToUpper(strings.TrimSpace(pair.Quote))

	// Sort dates ascending
	dates := make([]string, 0, len(recordsByDate))
	for d := range recordsByDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	var results []FXRecord
	for _, d := range dates {
		parsedDate, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		parsedDate = parsedDate.UTC()

		if parsedDate.Before(fromUTC) || parsedDate.After(toUTC) {
			continue
		}

		engine := NewTriangulationEngine(DefaultAnchorCurrency)
		engine.AddRates(recordsByDate[d])
		rate, _, err := engine.GetRate(base, quote)
		if err != nil {
			continue
		}

		results = append(results, FXRecord{
			BaseCurrency:  base,
			QuoteCurrency: quote,
			RateDate:      parsedDate,
			Rate:          rate,
			Source:        ecbSourceIdentifier,
		})
	}

	return results, nil
}
