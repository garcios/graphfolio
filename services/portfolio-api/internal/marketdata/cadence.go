package marketdata

import (
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// DetectFrequency analyzes historical distribution intervals to determine dividend cadence and step size.
func DetectFrequency(history []DividendHistoryPoint) (string, int) {
	if len(history) < 2 {
		return "QUARTERLY", 91
	}

	// Sort chronologically
	sorted := make([]DividendHistoryPoint, len(history))
	copy(sorted, history)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	var intervals []int
	for i := 1; i < len(sorted); i++ {
		days := int(sorted[i].Date.Sub(sorted[i-1].Date).Hours() / 24)
		if days > 10 { // Exclude special dividends / corrections
			intervals = append(intervals, days)
		}
	}

	if len(intervals) == 0 {
		return "QUARTERLY", 91
	}

	sort.Ints(intervals)
	medianDays := intervals[len(intervals)/2]

	switch {
	case medianDays >= 25 && medianDays <= 35:
		return "MONTHLY", 30
	case medianDays >= 80 && medianDays <= 105:
		return "QUARTERLY", 91
	case medianDays >= 160 && medianDays <= 200:
		return "SEMI_ANNUAL", 182
	case medianDays >= 340 && medianDays <= 380:
		return "ANNUAL", 365
	default:
		return "QUARTERLY", 91
	}
}

// SynthesizeForwardEstimates projects unannounced future dividend cycles for the next 12 months,
// flagging them with IsEstimated = true. Verified declarations within ±15 days take precedence.
func SynthesizeForwardEstimates(
	inst InstrumentRef,
	announced []DividendScheduleEvent,
	history []DividendHistoryPoint,
	asOf time.Time,
) []DividendScheduleEvent {
	if len(announced) == 0 && len(history) == 0 {
		return announced
	}

	frequency, stepDays := DetectFrequency(history)

	var latestDate time.Time
	var latestDPS decimal.Decimal
	payGapDays := 14

	// Determine latest known dividend event
	if len(announced) > 0 {
		for _, a := range announced {
			if a.ExDate.After(latestDate) {
				latestDate = a.ExDate
				latestDPS = a.AmountPerShare
				if a.PaymentDate.After(a.ExDate) {
					payGapDays = int(a.PaymentDate.Sub(a.ExDate).Hours() / 24)
				}
			}
		}
	}

	if (latestDPS.IsZero() || latestDate.IsZero()) && len(history) > 0 {
		sortedHist := make([]DividendHistoryPoint, len(history))
		copy(sortedHist, history)
		sort.Slice(sortedHist, func(i, j int) bool {
			return sortedHist[i].Date.Before(sortedHist[j].Date)
		})
		lastHist := sortedHist[len(sortedHist)-1]
		if lastHist.Date.After(latestDate) {
			latestDate = lastHist.Date
		}
		if latestDPS.IsZero() {
			latestDPS = lastHist.Amount
		}
	}

	if latestDPS.IsZero() || latestDate.IsZero() {
		return announced
	}

	frankingPct := decimal.Zero
	if strings.HasSuffix(strings.ToUpper(inst.Symbol), ".AX") || strings.ToUpper(inst.Exchange) == "ASX" || strings.ToUpper(inst.Exchange) == "XASX" {
		frankingPct = decimal.NewFromInt(100)
	}

	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	asOf = asOf.UTC().Truncate(24 * time.Hour)
	horizonEnd := asOf.AddDate(1, 0, 0)

	var synthesized []DividendScheduleEvent
	currentDate := latestDate.AddDate(0, 0, stepDays)

	// Step forward until horizonEnd
	for !currentDate.After(horizonEnd) {
		if currentDate.After(asOf) {
			// Check if any verified event in announced falls within ±15 days
			hasVerified := false
			for _, a := range announced {
				diffDays := int(a.ExDate.Sub(currentDate).Hours() / 24)
				if diffDays < 0 {
					diffDays = -diffDays
				}
				if diffDays <= 15 {
					hasVerified = true
					break
				}
			}

			if !hasVerified {
				payDate := currentDate.AddDate(0, 0, payGapDays)
				recDate := currentDate.AddDate(0, 0, 1)

				synthesized = append(synthesized, DividendScheduleEvent{
					InstrumentID:      inst.ID,
					Symbol:            inst.Symbol,
					ExchangeCode:      inst.Exchange,
					ExDate:            currentDate,
					RecordDate:        &recDate,
					PaymentDate:       payDate,
					AmountPerShare:    latestDPS,
					CurrencyCode:      inst.Currency,
					Frequency:         frequency,
					Status:            "ANNOUNCED",
					IsEstimated:       true,
					FrankingPct:       frankingPct,
					WithholdingTaxPct: decimal.Zero,
					Source:            "cadence_engine",
				})
			}
		}

		currentDate = currentDate.AddDate(0, 0, stepDays)
	}

	combined := append(announced, synthesized...)
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].ExDate.Before(combined[j].ExDate)
	})

	return combined
}
