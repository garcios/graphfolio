package marketdata_test

import (
	"errors"
	"testing"
	"time"

	"portfolio-api/internal/marketdata"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValidatePriceRecord(t *testing.T) {
	validDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	futureDate := time.Now().UTC().AddDate(0, 0, 5)

	tests := []struct {
		name    string
		rec     marketdata.PriceRecord
		wantErr error
	}{
		{
			name: "valid price record",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "AAPL",
				Exchange:     "NASDAQ",
				PriceDate:    validDate,
				ClosePrice:   decimal.NewFromFloat(228.50),
				Source:       "twelve_data",
			},
			wantErr: nil,
		},
		{
			name: "empty symbol",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "",
				Exchange:     "NASDAQ",
				PriceDate:    validDate,
				ClosePrice:   decimal.NewFromInt(100),
				Source:       "twelve_data",
			},
			wantErr: marketdata.ErrEmptySymbol,
		},
		{
			name: "zero close price",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "MSFT",
				Exchange:     "NASDAQ",
				PriceDate:    validDate,
				ClosePrice:   decimal.Zero,
				Source:       "twelve_data",
			},
			wantErr: marketdata.ErrInvalidPrice,
		},
		{
			name: "negative close price",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "MSFT",
				Exchange:     "NASDAQ",
				PriceDate:    validDate,
				ClosePrice:   decimal.NewFromFloat(-10.5),
				Source:       "twelve_data",
			},
			wantErr: marketdata.ErrInvalidPrice,
		},
		{
			name: "zero date",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "GOOGL",
				Exchange:     "NASDAQ",
				PriceDate:    time.Time{},
				ClosePrice:   decimal.NewFromInt(150),
				Source:       "twelve_data",
			},
			wantErr: marketdata.ErrZeroDate,
		},
		{
			name: "future date",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "GOOGL",
				Exchange:     "NASDAQ",
				PriceDate:    futureDate,
				ClosePrice:   decimal.NewFromInt(150),
				Source:       "twelve_data",
			},
			wantErr: marketdata.ErrFutureDate,
		},
		{
			name: "empty source",
			rec: marketdata.PriceRecord{
				InstrumentID: uuid.New(),
				Symbol:       "GOOGL",
				Exchange:     "NASDAQ",
				PriceDate:    validDate,
				ClosePrice:   decimal.NewFromInt(150),
				Source:       "",
			},
			wantErr: marketdata.ErrEmptySource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := marketdata.ValidatePriceRecord(tt.rec)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestValidateFXRecord(t *testing.T) {
	validDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	futureDate := time.Now().UTC().AddDate(0, 0, 5)

	tests := []struct {
		name    string
		rec     marketdata.FXRecord
		wantErr error
	}{
		{
			name: "valid fx record",
			rec: marketdata.FXRecord{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      validDate,
				Rate:          decimal.RequireFromString("1.0850000000"),
				Source:        "ecb",
			},
			wantErr: nil,
		},
		{
			name: "short base currency code",
			rec: marketdata.FXRecord{
				BaseCurrency:  "EU",
				QuoteCurrency: "USD",
				RateDate:      validDate,
				Rate:          decimal.NewFromInt(1),
				Source:        "ecb",
			},
			wantErr: marketdata.ErrInvalidCurrencyCode,
		},
		{
			name: "identical base and quote currencies",
			rec: marketdata.FXRecord{
				BaseCurrency:  "USD",
				QuoteCurrency: "USD",
				RateDate:      validDate,
				Rate:          decimal.NewFromInt(1),
				Source:        "ecb",
			},
			wantErr: marketdata.ErrIdenticalCurrencies,
		},
		{
			name: "zero fx rate",
			rec: marketdata.FXRecord{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      validDate,
				Rate:          decimal.Zero,
				Source:        "ecb",
			},
			wantErr: marketdata.ErrInvalidFXRate,
		},
		{
			name: "future fx date",
			rec: marketdata.FXRecord{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      futureDate,
				Rate:          decimal.RequireFromString("1.0850"),
				Source:        "ecb",
			},
			wantErr: marketdata.ErrFutureDate,
		},
		{
			name: "empty source",
			rec: marketdata.FXRecord{
				BaseCurrency:  "EUR",
				QuoteCurrency: "USD",
				RateDate:      validDate,
				Rate:          decimal.RequireFromString("1.0850"),
				Source:        "   ",
			},
			wantErr: marketdata.ErrEmptySource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := marketdata.ValidateFXRecord(tt.rec)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

func TestValidatePriceSpike(t *testing.T) {
	threshold := decimal.NewFromInt(50) // 50% max allowable change

	tests := []struct {
		name          string
		prevClose     decimal.Decimal
		newClose      decimal.Decimal
		wantSpike     bool
		wantPctChange decimal.Decimal
	}{
		{
			name:          "normal 10% price increase",
			prevClose:     decimal.NewFromInt(100),
			newClose:      decimal.NewFromInt(110),
			wantSpike:     false,
			wantPctChange: decimal.NewFromInt(10),
		},
		{
			name:          "normal 20% price drop",
			prevClose:     decimal.NewFromInt(100),
			newClose:      decimal.NewFromInt(80),
			wantSpike:     false,
			wantPctChange: decimal.NewFromInt(20),
		},
		{
			name:          "spike: 60% price jump",
			prevClose:     decimal.NewFromInt(100),
			newClose:      decimal.NewFromInt(160),
			wantSpike:     true,
			wantPctChange: decimal.NewFromInt(60),
		},
		{
			name:          "spike: 70% price crash (e.g. unadjusted stock split)",
			prevClose:     decimal.NewFromInt(100),
			newClose:      decimal.NewFromInt(30),
			wantSpike:     true,
			wantPctChange: decimal.NewFromInt(70),
		},
		{
			name:          "zero previous close returns no spike",
			prevClose:     decimal.Zero,
			newClose:      decimal.NewFromInt(100),
			wantSpike:     false,
			wantPctChange: decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSpike, pctChange := marketdata.ValidatePriceSpike(tt.prevClose, tt.newClose, threshold)
			if isSpike != tt.wantSpike {
				t.Errorf("ValidatePriceSpike() isSpike = %v, want %v", isSpike, tt.wantSpike)
			}
			if !pctChange.Equal(tt.wantPctChange) {
				t.Errorf("ValidatePriceSpike() pctChange = %s, want %s", pctChange.String(), tt.wantPctChange.String())
			}
		})
	}
}
