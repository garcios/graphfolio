package service

import (
	"context"
	"fmt"
	"time"

	"portfolio-api/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	zero = decimal.Zero
	one  = decimal.NewFromInt(1)
)

func (s *portfolioService) RebuildProjections(ctx context.Context, userID string) error {
	portfolio, err := s.repo.FindPortfolioByUser(ctx, userID)
	if err != nil {
		return err
	}

	txs, err := s.repo.GetTransactions(ctx, portfolio.ID)
	if err != nil {
		return fmt.Errorf("service: get transactions: %w", err)
	}

	// Collect all instrument IDs in transactions
	seenInstruments := make(map[uuid.UUID]bool)
	var instrumentIDs []uuid.UUID
	for _, tx := range txs {
		if tx.InstrumentID != nil && !seenInstruments[*tx.InstrumentID] {
			seenInstruments[*tx.InstrumentID] = true
			instrumentIDs = append(instrumentIDs, *tx.InstrumentID)
		}
	}

	cas, err := s.repo.GetCorporateActions(ctx, instrumentIDs)
	if err != nil {
		return fmt.Errorf("service: get corporate actions: %w", err)
	}

	lots, disposals, holdings, cash, err := ProcessLedger(*portfolio, txs, cas)
	if err != nil {
		return fmt.Errorf("service: process ledger: %w", err)
	}

	return s.repo.SaveProjectionsTx(ctx, portfolio.ID, lots, disposals, holdings, cash)
}

// ProcessLedger replays the transaction ledger and corporate actions to produce
// exact tax lots, lot disposals, holdings, and cash balances for both AVERAGE_COST and FIFO.
func ProcessLedger(
	portfolio domain.Portfolio,
	txs []domain.Transaction,
	cas []domain.CorporateAction,
) ([]domain.TaxLot, []domain.LotDisposal, []domain.Holding, []domain.CashBalance, error) {
	// Index corporate actions by instrument and ex_date
	casByInstrument := make(map[uuid.UUID][]domain.CorporateAction)
	for _, ca := range cas {
		casByInstrument[ca.InstrumentID] = append(casByInstrument[ca.InstrumentID], ca)
	}

	cashBalances := make(map[string]decimal.Decimal)
	openLots := make(map[uuid.UUID][]*domain.TaxLot)
	allLots := make([]domain.TaxLot, 0)
	allDisposals := make([]domain.LotDisposal, 0)
	realizedPnLByInstrument := make(map[uuid.UUID]decimal.Decimal)
	dividendsByInstrument := make(map[uuid.UUID]decimal.Decimal)

	for _, tx := range txs {
		fxRate := one
		if tx.FXRateToBase != nil && tx.FXRateToBase.IsPositive() {
			fxRate = *tx.FXRateToBase
		}

		switch tx.Type {
		case domain.TxTypeDeposit:
			cashBalances[tx.CurrencyCode] = cashBalances[tx.CurrencyCode].Add(tx.Amount)

		case domain.TxTypeWithdrawal:
			cashBalances[tx.CurrencyCode] = cashBalances[tx.CurrencyCode].Sub(tx.Amount)

		case domain.TxTypeDividend, domain.TxTypeInterest:
			netIncome := tx.Amount.Sub(tx.WithholdingTax)
			cashBalances[tx.CurrencyCode] = cashBalances[tx.CurrencyCode].Add(netIncome)
			if tx.InstrumentID != nil {
				dividendsByInstrument[*tx.InstrumentID] = dividendsByInstrument[*tx.InstrumentID].Add(netIncome.Mul(fxRate))
			}

		case domain.TxTypeBuy, domain.TxTypeTransferIn:
			if tx.InstrumentID == nil || tx.Quantity == nil {
				continue
			}

			// Cash impact: deduct gross amount + fee
			totalCost := tx.Amount.Add(tx.Fee)
			cashBalances[tx.CurrencyCode] = cashBalances[tx.CurrencyCode].Sub(totalCost)

			lot := domain.TaxLot{
				ID:                uuid.New(),
				PortfolioID:       portfolio.ID,
				InstrumentID:      *tx.InstrumentID,
				OpenTransactionID: tx.ID,
				AcquiredDate:      tx.TradeDate,
				OriginalQuantity:  *tx.Quantity,
				RemainingQuantity: *tx.Quantity,
				CostBasis:         totalCost,
				CostBasisBase:     totalCost.Mul(fxRate),
			}

			openLots[*tx.InstrumentID] = append(openLots[*tx.InstrumentID], &lot)

		case domain.TxTypeSell, domain.TxTypeTransferOut:
			if tx.InstrumentID == nil || tx.Quantity == nil {
				continue
			}

			instID := *tx.InstrumentID
			sellQty := *tx.Quantity
			netProceeds := tx.Amount.Sub(tx.Fee)
			netProceedsBase := netProceeds.Mul(fxRate)
			cashBalances[tx.CurrencyCode] = cashBalances[tx.CurrencyCode].Add(netProceeds)

			instLots := openLots[instID]
			if len(instLots) == 0 {
				continue // oversold / short positions outside scope
			}

			if portfolio.CostBasisMethod == domain.CostBasisMethodFIFO {
				// FIFO: Consume oldest open lots first
				remainingToSell := sellQty
				for _, lot := range instLots {
					if remainingToSell.IsZero() || lot.RemainingQuantity.IsZero() {
						continue
					}

					qtyToRelieve := decimal.Min(lot.RemainingQuantity, remainingToSell)
					shareOfLot := qtyToRelieve.Div(lot.RemainingQuantity)

					costReleased := lot.CostBasis.Mul(shareOfLot).Round(6)
					costReleasedBase := lot.CostBasisBase.Mul(shareOfLot).Round(6)

					// Proceeds allocated pro-rata to relieved quantity
					lotProceedsBase := netProceedsBase.Mul(qtyToRelieve.Div(sellQty)).Round(6)
					realizedPnL := lotProceedsBase.Sub(costReleasedBase)

					lot.RemainingQuantity = lot.RemainingQuantity.Sub(qtyToRelieve)
					lot.CostBasis = lot.CostBasis.Sub(costReleased)
					lot.CostBasisBase = lot.CostBasisBase.Sub(costReleasedBase)
					if lot.RemainingQuantity.IsZero() {
						closeDate := tx.TradeDate
						lot.ClosedDate = &closeDate
					}

					disposal := domain.LotDisposal{
						ID:                    uuid.New(),
						TaxLotID:              lot.ID,
						SellTransactionID:     tx.ID,
						Quantity:              qtyToRelieve,
						CostBasisReleased:     costReleased,
						CostBasisReleasedBase: costReleasedBase,
						ProceedsBase:          lotProceedsBase,
						RealizedPnLBase:       realizedPnL,
					}
					allDisposals = append(allDisposals, disposal)
					realizedPnLByInstrument[instID] = realizedPnLByInstrument[instID].Add(realizedPnL)

					remainingToSell = remainingToSell.Sub(qtyToRelieve)
				}
			} else {
				// AVERAGE_COST: Relieve units proportionally across ALL open lots
				totalOpenQty := zero
				for _, lot := range instLots {
					totalOpenQty = totalOpenQty.Add(lot.RemainingQuantity)
				}

				if totalOpenQty.IsPositive() {
					remainingToRelieve := sellQty
					totalCostReleasedBase := zero

					for i, lot := range instLots {
						if lot.RemainingQuantity.IsZero() {
							continue
						}

						var qtyToRelieve decimal.Decimal
						if i == len(instLots)-1 {
							qtyToRelieve = remainingToRelieve
						} else {
							qtyToRelieve = sellQty.Mul(lot.RemainingQuantity.Div(totalOpenQty)).Round(10)
							if qtyToRelieve.GreaterThan(lot.RemainingQuantity) {
								qtyToRelieve = lot.RemainingQuantity
							}
						}

						if qtyToRelieve.IsPositive() {
							shareOfLot := qtyToRelieve.Div(lot.RemainingQuantity)
							costReleased := lot.CostBasis.Mul(shareOfLot).Round(6)
							costReleasedBase := lot.CostBasisBase.Mul(shareOfLot).Round(6)

							lotProceedsBase := netProceedsBase.Mul(qtyToRelieve.Div(sellQty)).Round(6)
							realizedPnL := lotProceedsBase.Sub(costReleasedBase)

							lot.RemainingQuantity = lot.RemainingQuantity.Sub(qtyToRelieve)
							lot.CostBasis = lot.CostBasis.Sub(costReleased)
							lot.CostBasisBase = lot.CostBasisBase.Sub(costReleasedBase)
							if lot.RemainingQuantity.IsZero() {
								closeDate := tx.TradeDate
								lot.ClosedDate = &closeDate
							}

							disposal := domain.LotDisposal{
								ID:                    uuid.New(),
								TaxLotID:              lot.ID,
								SellTransactionID:     tx.ID,
								Quantity:              qtyToRelieve,
								CostBasisReleased:     costReleased,
								CostBasisReleasedBase: costReleasedBase,
								ProceedsBase:          lotProceedsBase,
								RealizedPnLBase:       realizedPnL,
							}
							allDisposals = append(allDisposals, disposal)
							totalCostReleasedBase = totalCostReleasedBase.Add(costReleasedBase)
							realizedPnLByInstrument[instID] = realizedPnLByInstrument[instID].Add(realizedPnL)

							remainingToRelieve = remainingToRelieve.Sub(qtyToRelieve)
						}
					}
				}
			}

		case domain.TxTypeSplit:
			if tx.InstrumentID == nil || tx.Quantity == nil || !tx.Quantity.IsPositive() {
				continue
			}

			splitRatio := *tx.Quantity
			if tx.Price != nil && tx.Price.IsPositive() {
				splitRatio = splitRatio.Div(*tx.Price)
			}
			if !splitRatio.IsPositive() {
				continue
			}

			instID := *tx.InstrumentID
			instLots := openLots[instID]
			for _, lot := range instLots {
				lot.OriginalQuantity = lot.OriginalQuantity.Mul(splitRatio).Round(10)
				lot.RemainingQuantity = lot.RemainingQuantity.Mul(splitRatio).Round(10)
			}
		}
	}

	// Flatten all tax lots and build holdings projections
	holdingsMap := make(map[uuid.UUID]*domain.Holding)
	for instID, lots := range openLots {
		for _, lot := range lots {
			allLots = append(allLots, *lot)
			if lot.RemainingQuantity.IsPositive() {
				h, exists := holdingsMap[instID]
				if !exists {
					h = &domain.Holding{
						PortfolioID:     portfolio.ID,
						InstrumentID:    instID,
						Quantity:        zero,
						CostBasis:       zero,
						CostBasisBase:   zero,
						RealizedPnLBase: realizedPnLByInstrument[instID],
						DividendsBase:   dividendsByInstrument[instID],
					}
					holdingsMap[instID] = h
				}
				h.Quantity = h.Quantity.Add(lot.RemainingQuantity)
				h.CostBasis = h.CostBasis.Add(lot.CostBasis)
				h.CostBasisBase = h.CostBasisBase.Add(lot.CostBasisBase)
			}
		}
	}

	holdings := make([]domain.Holding, 0, len(holdingsMap))
	for _, h := range holdingsMap {
		holdings = append(holdings, *h)
	}

	cashList := make([]domain.CashBalance, 0, len(cashBalances))
	for ccy, bal := range cashBalances {
		cashList = append(cashList, domain.CashBalance{
			PortfolioID:  portfolio.ID,
			CurrencyCode: ccy,
			Balance:      bal,
		})
	}

	_ = time.Now()
	return allLots, allDisposals, holdings, cashList, nil
}
