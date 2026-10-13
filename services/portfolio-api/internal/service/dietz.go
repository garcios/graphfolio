package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/shopspring/decimal"
)

type monthInterval struct {
	start time.Time
	end   time.Time
}

func generateMonthIntervals(start, end time.Time) []monthInterval {
	var intervals []monthInterval
	currStart := start
	for currStart.Before(end) {
		// First day of next month
		nextMonth := time.Date(currStart.Year(), currStart.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		currEnd := nextMonth
		if currEnd.After(end) {
			currEnd = end
		}
		intervals = append(intervals, monthInterval{
			start: currStart,
			end:   currEnd,
		})
		currStart = currEnd
	}
	return intervals
}

func (s *portfolioService) GetModifiedDietzReturn(
	ctx context.Context,
	userID string,
	startDate, endDate time.Time,
	useMonthlyLinking bool,
) (*domain.ModifiedDietzResult, string, error) {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	startUTC := startDate.UTC().Truncate(24 * time.Hour)
	endUTC := endDate.UTC().Truncate(24 * time.Hour)
	if endUTC.Before(startUTC) {
		return nil, "", fmt.Errorf("service: end date %s cannot be before start date %s", endUTC.Format("2006-01-02"), startUTC.Format("2006-01-02"))
	}

	// Clamp startUTC to portfolio inception if earlier (portfolio did not exist before inception)
	earliestVal, err := s.repo.GetEarliestValuation(ctx, portfolio.ID)
	var earliestDate time.Time
	if err == nil && earliestVal != nil && !earliestVal.ValuationDate.IsZero() {
		earliestDate = earliestVal.ValuationDate.UTC().Truncate(24 * time.Hour)
		if startUTC.Before(earliestDate) {
			startUTC = earliestDate
		}
	}

	// Fetch all external cash flows in [startUTC, endUTC]
	flows, err := s.repo.GetExternalCashFlows(ctx, portfolio.ID, startUTC, endUTC, portfolio.BaseCurrency)
	if err != nil {
		return nil, "", fmt.Errorf("service: failed to fetch external cash flows: %w", err)
	}

	cdDays := int32(endUTC.Sub(startUTC).Hours() / 24.0)

	// If monthly linking is requested and elapsed days >= 365
	if useMonthlyLinking && cdDays >= 365 {
		intervals := generateMonthIntervals(startUTC, endUTC)
		if len(intervals) > 1 {
			subPeriods := make([]*domain.ModifiedDietzResult, 0, len(intervals))

			for i, interval := range intervals {
				var vMStart decimal.Decimal
				if i == 0 {
					if !earliestDate.IsZero() && interval.start.Equal(earliestDate) {
						vMStart = decimal.Zero
					} else {
						mStartVal, _ := s.repo.GetValuationAtOrBefore(ctx, portfolio.ID, interval.start)
						if mStartVal != nil {
							vMStart = mStartVal.TotalValue()
						}
					}
				} else {
					mStartVal, _ := s.repo.GetValuationAtOrBefore(ctx, portfolio.ID, interval.start)
					if mStartVal != nil {
						vMStart = mStartVal.TotalValue()
					}
				}

				mEndVal, _ := s.repo.GetValuationAtOrBefore(ctx, portfolio.ID, interval.end)
				vMEnd := decimal.Zero
				if mEndVal != nil {
					vMEnd = mEndVal.TotalValue()
				}

				// Filter flows in this interval:
				// First interval includes start boundary; subsequent intervals exclude start boundary to avoid double-counting
				var mFlows []domain.ExternalCashFlow
				for _, f := range flows {
					flowDate := f.Date.UTC().Truncate(24 * time.Hour)
					if i == 0 {
						if !flowDate.Before(interval.start) && !flowDate.After(interval.end) {
							mFlows = append(mFlows, f)
						}
					} else {
						if flowDate.After(interval.start) && !flowDate.After(interval.end) {
							mFlows = append(mFlows, f)
						}
					}
				}

				if vMStart.IsZero() && vMEnd.IsZero() && len(mFlows) == 0 {
					// Inactive interval prior to funding or during dormancy
					continue
				}

				mRes, err := domain.CalculateModifiedDietz(vMStart, vMEnd, interval.start, interval.end, mFlows)
				if err != nil {
					return nil, "", fmt.Errorf("service: failed to compute monthly Modified Dietz for interval %s to %s: %w",
						interval.start.Format("2006-01-02"), interval.end.Format("2006-01-02"), err)
				}
				subPeriods = append(subPeriods, mRes)
			}

			if len(subPeriods) > 0 {
				chainedRes, err := domain.ChainModifiedDietzSubPeriods(subPeriods, startUTC, endUTC)
				if err != nil {
					return nil, "", fmt.Errorf("service: failed to chain monthly sub-periods: %w", err)
				}
				return chainedRes, portfolio.BaseCurrency, nil
			}
		}
	}

	// Single-period calculation
	var vStart decimal.Decimal
	if !earliestDate.IsZero() && startUTC.Equal(earliestDate) {
		vStart = decimal.Zero
	} else {
		startVal, err := s.repo.GetValuationAtOrBefore(ctx, portfolio.ID, startUTC)
		if err != nil {
			return nil, "", fmt.Errorf("service: failed to fetch start valuation: %w", err)
		}
		if startVal != nil {
			vStart = startVal.TotalValue()
		}
	}

	endVal, err := s.repo.GetValuationAtOrBefore(ctx, portfolio.ID, endUTC)
	if err != nil {
		return nil, "", fmt.Errorf("service: failed to fetch end valuation: %w", err)
	}

	vEnd := decimal.Zero
	if endVal != nil {
		vEnd = endVal.TotalValue()
	}

	result, err := domain.CalculateModifiedDietz(vStart, vEnd, startUTC, endUTC, flows)
	if err != nil {
		return nil, "", err
	}

	return result, portfolio.BaseCurrency, nil
}
