package marketdata

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type mockDividendProvider struct {
	name             string
	upcomingEvents   []DividendScheduleEvent
	upcomingErr      error
	historicalPoints []DividendHistoryPoint
	historicalErr    error
	upcomingCalls    int
	historicalCalls  int
}

func (m *mockDividendProvider) Name() string {
	return m.name
}

func (m *mockDividendProvider) GetUpcomingDividends(ctx context.Context, inst InstrumentRef) ([]DividendScheduleEvent, error) {
	m.upcomingCalls++
	if m.upcomingErr != nil {
		return nil, m.upcomingErr
	}
	return m.upcomingEvents, nil
}

func (m *mockDividendProvider) GetHistoricalDividends(ctx context.Context, inst InstrumentRef, fromDate, toDate time.Time) ([]DividendHistoryPoint, error) {
	m.historicalCalls++
	if m.historicalErr != nil {
		return nil, m.historicalErr
	}
	return m.historicalPoints, nil
}

func TestResilientDividendProvider_PrimarySuccess(t *testing.T) {
	ctx := context.Background()
	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	primary := &mockDividendProvider{
		name: "Yahoo Finance",
		upcomingEvents: []DividendScheduleEvent{
			{Symbol: "AAPL", AmountPerShare: decimal.NewFromFloat(0.24)},
		},
	}
	fallback := &mockDividendProvider{
		name: "Twelve Data",
	}

	provider := NewResilientDividendProvider(primary, fallback)

	events, err := provider.GetUpcomingDividends(ctx, inst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if primary.upcomingCalls != 1 {
		t.Errorf("expected 1 primary call, got %d", primary.upcomingCalls)
	}
	if fallback.upcomingCalls != 0 {
		t.Errorf("expected 0 fallback calls when primary succeeds, got %d", fallback.upcomingCalls)
	}
}

func TestResilientDividendProvider_FailoverOnPrimaryError(t *testing.T) {
	ctx := context.Background()
	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "AAPL",
		Exchange: "NASDAQ",
		Currency: "USD",
	}

	primary := &mockDividendProvider{
		name:        "Yahoo Finance",
		upcomingErr: errors.New("upstream 429 rate limit"),
	}
	fallback := &mockDividendProvider{
		name: "Twelve Data",
		upcomingEvents: []DividendScheduleEvent{
			{Symbol: "AAPL", AmountPerShare: decimal.NewFromFloat(0.24), Source: "twelvedata"},
		},
	}

	provider := NewResilientDividendProvider(primary, fallback)

	events, err := provider.GetUpcomingDividends(ctx, inst)
	if err != nil {
		t.Fatalf("unexpected error on fallback: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event from fallback, got %d", len(events))
	}
	if events[0].Source != "twelvedata" {
		t.Errorf("expected source twelvedata, got %s", events[0].Source)
	}
	if primary.upcomingCalls != 1 {
		t.Errorf("expected 1 primary call, got %d", primary.upcomingCalls)
	}
	if fallback.upcomingCalls != 1 {
		t.Errorf("expected 1 fallback call, got %d", fallback.upcomingCalls)
	}
}

func TestResilientDividendProvider_BothFail(t *testing.T) {
	ctx := context.Background()
	inst := InstrumentRef{
		ID:       uuid.New(),
		Symbol:   "UNKNOWN",
		Exchange: "NASDAQ",
	}

	primary := &mockDividendProvider{
		name:        "Yahoo Finance",
		upcomingErr: errors.New("yahoo 404 not found"),
	}
	fallback := &mockDividendProvider{
		name:        "Twelve Data",
		upcomingErr: errors.New("twelve data quota exceeded"),
	}

	provider := NewResilientDividendProvider(primary, fallback)

	_, err := provider.GetUpcomingDividends(ctx, inst)
	if err == nil {
		t.Fatalf("expected error when both fail, got nil")
	}
}
