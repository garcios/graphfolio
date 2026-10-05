package marketdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	defaultTwelveDataBaseURL = "https://api.twelvedata.com"
	defaultYahooBaseURL      = "https://query1.finance.yahoo.com"
	twelveDataSourceID       = "twelvedata"
	yahooSourceID            = "yahoo"
)

var (
	ErrPriceNotFound      = errors.New("closing price not found for instrument")
	ErrProviderQuota      = errors.New("upstream provider rate limit or quota exceeded")
	ErrInvalidAPIResponse = errors.New("invalid or malformed provider response")
)

// ============================================================================
// 1. Twelve Data Provider
// ============================================================================

// TwelveDataConfig configures the Twelve Data API adapter.
type TwelveDataConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Limiter    *RateLimiter
	RetryCfg   RetryConfig
}

// TwelveDataProvider implements PriceProvider using the Twelve Data REST API.
type TwelveDataProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	limiter    *RateLimiter
	retryCfg   RetryConfig
}

// NewTwelveDataProvider constructs a Twelve Data price provider.
func NewTwelveDataProvider(cfg TwelveDataConfig) *TwelveDataProvider {
	base := cfg.BaseURL
	if base == "" {
		base = defaultTwelveDataBaseURL
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	retryCfg := cfg.RetryCfg
	if retryCfg.MaxRetries == 0 && retryCfg.InitialBackoff == 0 {
		retryCfg = DefaultRetryConfig()
	}

	return &TwelveDataProvider{
		baseURL:    strings.TrimRight(base, "/"),
		apiKey:     cfg.APIKey,
		httpClient: client,
		limiter:    cfg.Limiter,
		retryCfg:   retryCfg,
	}
}

func (p *TwelveDataProvider) Name() string {
	return "Twelve Data"
}

type twelveDataTimeSeriesResponse struct {
	Meta struct {
		Symbol   string `json:"symbol"`
		Currency string `json:"currency"`
		Exchange string `json:"exchange"`
	} `json:"meta"`
	Values []struct {
		Datetime string `json:"datetime"`
		Open     string `json:"open"`
		High     string `json:"high"`
		Low      string `json:"low"`
		Close    string `json:"close"`
		Volume   string `json:"volume"`
	} `json:"values"`
	Status  string `json:"status"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (p *TwelveDataProvider) fetchTimeSeries(ctx context.Context, queryParams url.Values) (*twelveDataTimeSeriesResponse, error) {
	if p.apiKey != "" && queryParams.Get("apikey") == "" {
		queryParams.Set("apikey", p.apiKey)
	}

	endpoint := fmt.Sprintf("%s/time_series?%s", p.baseURL, queryParams.Encode())

	var result twelveDataTimeSeriesResponse
	err := ExecuteWithRetry(ctx, p.retryCfg, func(attempt int) error {
		if p.limiter != nil {
			if err := p.limiter.Wait(ctx); err != nil {
				return err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidAPIResponse, err)
		}

		// Twelve Data returns 200 OK with status="error" and code 429 when quota is hit
		if result.Status == "error" {
			if result.Code == 429 || strings.Contains(strings.ToLower(result.Message), "credits") {
				return &HTTPStatusError{
					StatusCode: http.StatusTooManyRequests,
					Status:     "Too Many Requests",
					Body:       result.Message,
				}
			}
			return fmt.Errorf("twelve data error: %s (code %d)", result.Message, result.Code)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchClosingPrices retrieves EOD closing marks for a slice of instruments on a given date.
func (p *TwelveDataProvider) FetchClosingPrices(ctx context.Context, instruments []InstrumentRef, date time.Time) ([]PriceRecord, error) {
	if len(instruments) == 0 {
		return nil, nil
	}

	targetDateStr := ""
	if !date.IsZero() {
		targetDateStr = date.UTC().Format("2006-01-02")
	}

	var records []PriceRecord
	for _, inst := range instruments {
		sym := strings.TrimSpace(inst.Symbol)
		if sym == "" {
			continue
		}

		q := url.Values{}
		q.Set("symbol", sym)
		q.Set("interval", "1day")
		q.Set("outputsize", "5") // fetch small buffer in case of weekend/holiday LOCF

		tsResp, err := p.fetchTimeSeries(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("twelve data failed for %s: %w", sym, err)
		}

		if len(tsResp.Values) == 0 {
			continue
		}

		// Locate target date or use latest available
		var chosenVal *struct {
			Datetime string `json:"datetime"`
			Open     string `json:"open"`
			High     string `json:"high"`
			Low      string `json:"low"`
			Close    string `json:"close"`
			Volume   string `json:"volume"`
		}

		if targetDateStr != "" {
			for i := range tsResp.Values {
				if strings.HasPrefix(tsResp.Values[i].Datetime, targetDateStr) {
					chosenVal = &tsResp.Values[i]
					break
				}
			}
			// If not found (e.g. weekend), pick most recent observation on or before date
			if chosenVal == nil {
				for i := range tsResp.Values {
					if tsResp.Values[i].Datetime <= targetDateStr {
						chosenVal = &tsResp.Values[i]
						break
					}
				}
			}
		} else {
			// Zero date -> newest observation
			chosenVal = &tsResp.Values[0]
		}

		if chosenVal == nil {
			continue
		}

		closeDec, err := decimal.NewFromString(strings.TrimSpace(chosenVal.Close))
		if err != nil || !closeDec.IsPositive() {
			continue
		}

		obsDate, err := time.Parse("2006-01-02", strings.Split(chosenVal.Datetime, " ")[0])
		if err != nil {
			continue
		}

		rec := PriceRecord{
			InstrumentID: inst.ID,
			Symbol:       sym,
			Exchange:     inst.Exchange,
			PriceDate:    obsDate.UTC(),
			ClosePrice:   closeDec,
			Source:       twelveDataSourceID,
		}

		if err := ValidatePriceRecord(rec); err == nil {
			records = append(records, rec)
		}
	}

	return records, nil
}

// FetchHistoricalPrices retrieves daily closing marks across an inclusive date range [from, to].
func (p *TwelveDataProvider) FetchHistoricalPrices(ctx context.Context, inst InstrumentRef, from, to time.Time) ([]PriceRecord, error) {
	sym := strings.TrimSpace(inst.Symbol)
	if sym == "" {
		return nil, ErrEmptySymbol
	}

	fromUTC := from.UTC().Truncate(24 * time.Hour)
	toUTC := to.UTC().Truncate(24 * time.Hour)
	if toUTC.Before(fromUTC) {
		return nil, errors.New("to date cannot be before from date")
	}

	q := url.Values{}
	q.Set("symbol", sym)
	q.Set("interval", "1day")
	q.Set("start_date", fromUTC.Format("2006-01-02"))
	q.Set("end_date", toUTC.Format("2006-01-02"))
	q.Set("outputsize", "5000")

	tsResp, err := p.fetchTimeSeries(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("twelve data historical failed for %s: %w", sym, err)
	}

	var records []PriceRecord
	for _, val := range tsResp.Values {
		closeDec, err := decimal.NewFromString(strings.TrimSpace(val.Close))
		if err != nil || !closeDec.IsPositive() {
			continue
		}

		obsDate, err := time.Parse("2006-01-02", strings.Split(val.Datetime, " ")[0])
		if err != nil {
			continue
		}
		obsDate = obsDate.UTC()

		if obsDate.Before(fromUTC) || obsDate.After(toUTC) {
			continue
		}

		rec := PriceRecord{
			InstrumentID: inst.ID,
			Symbol:       sym,
			Exchange:     inst.Exchange,
			PriceDate:    obsDate,
			ClosePrice:   closeDec,
			Source:       twelveDataSourceID,
		}

		if err := ValidatePriceRecord(rec); err == nil {
			records = append(records, rec)
		}
	}

	// Sort ascending by date
	sort.Slice(records, func(i, j int) bool {
		return records[i].PriceDate.Before(records[j].PriceDate)
	})

	return records, nil
}

// ============================================================================
// 2. Yahoo Finance Provider
// ============================================================================

// YahooFinanceConfig configures the Yahoo Finance API adapter.
type YahooFinanceConfig struct {
	BaseURL    string
	HTTPClient *http.Client
	Limiter    *RateLimiter
	RetryCfg   RetryConfig
	UserAgent  string
}

// YahooFinanceProvider implements PriceProvider using Yahoo Finance Chart API.
type YahooFinanceProvider struct {
	baseURL    string
	httpClient *http.Client
	limiter    *RateLimiter
	retryCfg   RetryConfig
	userAgent  string
}

// NewYahooFinanceProvider constructs a Yahoo Finance price provider.
func NewYahooFinanceProvider(cfg YahooFinanceConfig) *YahooFinanceProvider {
	base := cfg.BaseURL
	if base == "" {
		base = defaultYahooBaseURL
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	ua := cfg.UserAgent
	if ua == "" {
		ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko)"
	}

	retryCfg := cfg.RetryCfg
	if retryCfg.MaxRetries == 0 && retryCfg.InitialBackoff == 0 {
		retryCfg = DefaultRetryConfig()
	}

	return &YahooFinanceProvider{
		baseURL:    strings.TrimRight(base, "/"),
		httpClient: client,
		limiter:    cfg.Limiter,
		retryCfg:   retryCfg,
		userAgent:  ua,
	}
}

func (p *YahooFinanceProvider) Name() string {
	return "Yahoo Finance"
}

// Yahoo chart JSON response with exact json.Number decimal preservation.
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency string `json:"currency"`
				Symbol   string `json:"symbol"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []json.Number `json:"close"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func (p *YahooFinanceProvider) fetchChart(ctx context.Context, symbol string, params url.Values) (*yahooChartResponse, error) {
	endpoint := fmt.Sprintf("%s/v8/finance/chart/%s?%s", p.baseURL, url.PathEscape(symbol), params.Encode())

	var result yahooChartResponse
	err := ExecuteWithRetry(ctx, p.retryCfg, func(attempt int) error {
		if p.limiter != nil {
			if err := p.limiter.Wait(ctx); err != nil {
				return err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", p.userAgent)

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

		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber() // Prevents IEEE-754 float drift

		if err := decoder.Decode(&result); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidAPIResponse, err)
		}

		if result.Chart.Error != nil {
			return fmt.Errorf("yahoo finance error: %s (%s)", result.Chart.Error.Description, result.Chart.Error.Code)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchClosingPrices retrieves daily closing marks using Yahoo Finance.
func (p *YahooFinanceProvider) FetchClosingPrices(ctx context.Context, instruments []InstrumentRef, date time.Time) ([]PriceRecord, error) {
	if len(instruments) == 0 {
		return nil, nil
	}

	targetDateStr := ""
	if !date.IsZero() {
		targetDateStr = date.UTC().Format("2006-01-02")
	}

	var records []PriceRecord
	for _, inst := range instruments {
		sym := strings.TrimSpace(inst.Symbol)
		if sym == "" {
			continue
		}

		params := url.Values{}
		params.Set("interval", "1d")
		params.Set("range", "5d")

		chartResp, err := p.fetchChart(ctx, sym, params)
		if err != nil {
			return nil, fmt.Errorf("yahoo finance failed for %s: %w", sym, err)
		}

		if len(chartResp.Chart.Result) == 0 {
			continue
		}

		res := chartResp.Chart.Result[0]
		if len(res.Timestamp) == 0 || len(res.Indicators.Quote) == 0 || len(res.Indicators.Quote[0].Close) == 0 {
			continue
		}

		closes := res.Indicators.Quote[0].Close
		timestamps := res.Timestamp

		// Scan backwards for target date or newest
		var chosenClose json.Number
		var chosenTime time.Time
		found := false

		for i := len(timestamps) - 1; i >= 0; i-- {
			if i >= len(closes) {
				continue
			}
			t := time.Unix(timestamps[i], 0).UTC()
			tStr := t.Format("2006-01-02")

			if targetDateStr == "" || tStr <= targetDateStr {
				raw := string(closes[i])
				if raw == "" || raw == "null" {
					continue
				}
				chosenClose = closes[i]
				chosenTime = t.Truncate(24 * time.Hour)
				found = true
				break
			}
		}

		if !found {
			continue
		}

		closeDec, err := decimal.NewFromString(string(chosenClose))
		if err != nil || !closeDec.IsPositive() {
			continue
		}

		rec := PriceRecord{
			InstrumentID: inst.ID,
			Symbol:       sym,
			Exchange:     inst.Exchange,
			PriceDate:    chosenTime,
			ClosePrice:   closeDec,
			Source:       yahooSourceID,
		}

		if err := ValidatePriceRecord(rec); err == nil {
			records = append(records, rec)
		}
	}

	return records, nil
}

// FetchHistoricalPrices retrieves daily closing marks across an inclusive date range [from, to].
func (p *YahooFinanceProvider) FetchHistoricalPrices(ctx context.Context, inst InstrumentRef, from, to time.Time) ([]PriceRecord, error) {
	sym := strings.TrimSpace(inst.Symbol)
	if sym == "" {
		return nil, ErrEmptySymbol
	}

	fromUTC := from.UTC().Truncate(24 * time.Hour)
	toUTC := to.UTC().Truncate(24 * time.Hour)
	if toUTC.Before(fromUTC) {
		return nil, errors.New("to date cannot be before from date")
	}

	params := url.Values{}
	params.Set("interval", "1d")
	params.Set("period1", strconv.FormatInt(fromUTC.Unix(), 10))
	// Period2 on Yahoo is exclusive; add 24 hours to include target day
	params.Set("period2", strconv.FormatInt(toUTC.Add(24*time.Hour).Unix(), 10))

	chartResp, err := p.fetchChart(ctx, sym, params)
	if err != nil {
		return nil, fmt.Errorf("yahoo finance historical failed for %s: %w", sym, err)
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, nil
	}

	res := chartResp.Chart.Result[0]
	if len(res.Timestamp) == 0 || len(res.Indicators.Quote) == 0 {
		return nil, nil
	}

	closes := res.Indicators.Quote[0].Close
	timestamps := res.Timestamp

	var records []PriceRecord
	for i, ts := range timestamps {
		if i >= len(closes) {
			break
		}
		raw := string(closes[i])
		if raw == "" || raw == "null" {
			continue
		}

		closeDec, err := decimal.NewFromString(raw)
		if err != nil || !closeDec.IsPositive() {
			continue
		}

		obsDate := time.Unix(ts, 0).UTC().Truncate(24 * time.Hour)
		if obsDate.Before(fromUTC) || obsDate.After(toUTC) {
			continue
		}

		rec := PriceRecord{
			InstrumentID: inst.ID,
			Symbol:       sym,
			Exchange:     inst.Exchange,
			PriceDate:    obsDate,
			ClosePrice:   closeDec,
			Source:       yahooSourceID,
		}

		if err := ValidatePriceRecord(rec); err == nil {
			records = append(records, rec)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].PriceDate.Before(records[j].PriceDate)
	})

	return records, nil
}

// ============================================================================
// 3. Resilient / Fallback Price Provider
// ============================================================================

// ResilientPriceProvider wraps a primary PriceProvider with a secondary fallback PriceProvider.
type ResilientPriceProvider struct {
	primary  PriceProvider
	fallback PriceProvider
}

// NewResilientPriceProvider creates a composite provider with seamless fallback.
func NewResilientPriceProvider(primary, fallback PriceProvider) *ResilientPriceProvider {
	return &ResilientPriceProvider{
		primary:  primary,
		fallback: fallback,
	}
}

func (p *ResilientPriceProvider) Name() string {
	if p.fallback != nil {
		return fmt.Sprintf("%s (fallback: %s)", p.primary.Name(), p.fallback.Name())
	}
	return p.primary.Name()
}

// FetchClosingPrices attempts primary provider first; on failure or missing symbols, falls back.
func (p *ResilientPriceProvider) FetchClosingPrices(ctx context.Context, instruments []InstrumentRef, date time.Time) ([]PriceRecord, error) {
	primaryRecords, primaryErr := p.primary.FetchClosingPrices(ctx, instruments, date)
	if primaryErr == nil && len(primaryRecords) == len(instruments) {
		return primaryRecords, nil
	}

	if p.fallback == nil {
		return primaryRecords, primaryErr
	}

	// Identify which instruments were missing or if primary entirely failed
	foundSymbols := make(map[string]bool)
	for _, r := range primaryRecords {
		foundSymbols[r.Symbol] = true
	}

	var missing []InstrumentRef
	for _, inst := range instruments {
		if !foundSymbols[inst.Symbol] {
			missing = append(missing, inst)
		}
	}

	if len(missing) == 0 {
		return primaryRecords, nil
	}

	fallbackRecords, fallbackErr := p.fallback.FetchClosingPrices(ctx, missing, date)
	if fallbackErr != nil && len(primaryRecords) == 0 {
		return nil, fmt.Errorf("primary failed (%v) and fallback failed (%w)", primaryErr, fallbackErr)
	}

	combined := append(primaryRecords, fallbackRecords...)
	return combined, nil
}

// FetchHistoricalPrices attempts primary provider first; falls back if primary returns an error.
func (p *ResilientPriceProvider) FetchHistoricalPrices(ctx context.Context, inst InstrumentRef, from, to time.Time) ([]PriceRecord, error) {
	records, err := p.primary.FetchHistoricalPrices(ctx, inst, from, to)
	if err == nil && len(records) > 0 {
		return records, nil
	}

	if p.fallback == nil {
		return records, err
	}

	fallbackRecords, fallbackErr := p.fallback.FetchHistoricalPrices(ctx, inst, from, to)
	if fallbackErr != nil {
		if err != nil {
			return nil, fmt.Errorf("primary failed (%v) and fallback failed (%w)", err, fallbackErr)
		}
		return nil, fallbackErr
	}

	return fallbackRecords, nil
}
