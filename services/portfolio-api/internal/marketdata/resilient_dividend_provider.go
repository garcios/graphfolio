package marketdata

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ResilientDividendProvider cascades requests between a primary DividendProvider (Yahoo Finance)
// and a secondary fallback DividendProvider (Twelve Data) to prevent quota exhaustion and handle outages.
type ResilientDividendProvider struct {
	primary  DividendProvider
	fallback DividendProvider
}

// NewResilientDividendProvider constructs a resilient composite dividend provider.
func NewResilientDividendProvider(primary, fallback DividendProvider) *ResilientDividendProvider {
	return &ResilientDividendProvider{
		primary:  primary,
		fallback: fallback,
	}
}

func (r *ResilientDividendProvider) Name() string {
	if r.fallback != nil {
		return fmt.Sprintf("%s (fallback: %s)", r.primary.Name(), r.fallback.Name())
	}
	return r.primary.Name()
}

// GetUpcomingDividends attempts to retrieve forward dividend declarations from the primary provider,
// failing over to the fallback provider if primary errors or returns empty results.
func (r *ResilientDividendProvider) GetUpcomingDividends(ctx context.Context, inst InstrumentRef) ([]DividendScheduleEvent, error) {
	events, err := r.primary.GetUpcomingDividends(ctx, inst)
	if err == nil && len(events) > 0 {
		return events, nil
	}

	if r.fallback == nil {
		return events, err
	}

	log.Printf("[marketdata] primary dividend provider %q failed or empty for %s (err: %v), failing over to %q",
		r.primary.Name(), inst.Symbol, err, r.fallback.Name())

	fallbackEvents, fallbackErr := r.fallback.GetUpcomingDividends(ctx, inst)
	if fallbackErr != nil {
		if err != nil {
			return nil, fmt.Errorf("both primary (%v) and fallback (%w) feeds failed", err, fallbackErr)
		}
		return nil, fallbackErr
	}

	return fallbackEvents, nil
}

// GetHistoricalDividends attempts to retrieve historical dividend distributions from the primary provider,
// failing over to the fallback provider if primary errors or returns empty results.
func (r *ResilientDividendProvider) GetHistoricalDividends(ctx context.Context, inst InstrumentRef, fromDate, toDate time.Time) ([]DividendHistoryPoint, error) {
	points, err := r.primary.GetHistoricalDividends(ctx, inst, fromDate, toDate)
	if err == nil && len(points) > 0 {
		return points, nil
	}

	if r.fallback == nil {
		return points, err
	}

	log.Printf("[marketdata] primary historical dividend provider %q failed or empty for %s (err: %v), failing over to %q",
		r.primary.Name(), inst.Symbol, err, r.fallback.Name())

	fallbackPoints, fallbackErr := r.fallback.GetHistoricalDividends(ctx, inst, fromDate, toDate)
	if fallbackErr != nil {
		if err != nil {
			return nil, fmt.Errorf("both primary (%v) and fallback (%w) historical feeds failed", err, fallbackErr)
		}
		return nil, fallbackErr
	}

	return fallbackPoints, nil
}
