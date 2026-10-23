import React, { useState, useEffect, useMemo } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney } from '@graphfolio/ui';
import type { TaxLotItem } from './OpenTaxLotsTable';
import './TaxLotInspectorModal.css';

export interface TaxLotInspectorModalProps {
  isOpen: boolean;
  onClose: () => void;
  instrumentId?: string;
  ticker?: string;
  name?: string;
  currentPrice?: { amount: string; currencyCode: string };
  baseCurrency?: string;
}

export const TaxLotInspectorModal: React.FC<TaxLotInspectorModalProps> = ({
  isOpen,
  onClose,
  instrumentId,
  ticker,
  name,
  currentPrice,
  baseCurrency = 'USD',
}) => {
  const [lots, setLots] = useState<TaxLotItem[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<'ALL' | 'OPEN' | 'CLOSED'>('ALL');

  useEffect(() => {
    if (!isOpen || !instrumentId) return;

    setLoading(true);
    setError(null);

    client
      .query({
        taxLots: {
          __args: {
            instrumentId,
            status: statusFilter === 'ALL' ? undefined : (statusFilter as any),
          },
          id: true,
          instrumentId: true,
          ticker: true,
          name: true,
          openTransactionId: true,
          acquiredDate: true,
          closedDate: true,
          originalQuantity: true,
          remainingQuantity: true,
          costBasis: { amount: true, currencyCode: true },
          costBasisBase: { amount: true, currencyCode: true },
          unitCostBase: { amount: true, currencyCode: true },
          currentPrice: { amount: true, currencyCode: true },
          marketValueBase: { amount: true, currencyCode: true },
          unrealizedPnlBase: { amount: true, currencyCode: true },
          unrealizedPnlPercent: true,
          holdingPeriodDays: true,
          termClassification: true,
          isCgtDiscountEligible: true,
          isLossHarvestingCandidate: true,
          hasWashSaleRisk: true,
        },
      })
      .then((res: any) => {
        if (res?.taxLots) {
          setLots(res.taxLots);
        } else {
          setLots([]);
        }
      })
      .catch((err: any) => {
        setError(err.message || 'Failed to load tax lots.');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [isOpen, instrumentId, statusFilter]);

  // Aggregate metrics across active open lots
  const { totalUnits, totalCostBasis, totalMarketValue, totalUnrealizedPnl, cgtEligibleUnits } = useMemo(() => {
    let units = 0;
    let cost = 0;
    let value = 0;
    let pnl = 0;
    let cgtUnits = 0;

    for (const lot of lots) {
      const remQty = parseFloat(lot.remainingQuantity) || 0;
      if (remQty > 0) {
        units += remQty;
        cost += parseFloat(lot.costBasisBase?.amount || '0') || 0;
        value += parseFloat(lot.marketValueBase?.amount || '0') || 0;
        pnl += parseFloat(lot.unrealizedPnlBase?.amount || '0') || 0;
        if (lot.isCgtDiscountEligible) {
          cgtUnits += remQty;
        }
      }
    }

    return {
      totalUnits: units,
      totalCostBasis: cost,
      totalMarketValue: value,
      totalUnrealizedPnl: pnl,
      cgtEligibleUnits: cgtUnits,
    };
  }, [lots]);

  if (!isOpen) return null;

  return (
    <div className="tax-inspector-overlay" onClick={onClose} data-testid="tax-lot-inspector-modal">
      <div className="tax-inspector-modal" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="tax-inspector-header">
          <div className="tax-inspector-title-group">
            <span className="inspector-ticker-badge">{ticker || 'ASSET'}</span>
            <div>
              <h2>{name || 'Tax Lot Inspector'}</h2>
              <p>
                Granular parcel inventory, acquisition dates, cost basis relief, and CGT discount tracking.
              </p>
            </div>
          </div>
          <button
            type="button"
            className="tax-inspector-close-btn"
            onClick={onClose}
            aria-label="Close"
          >
            &times;
          </button>
        </div>

        {/* Aggregate Summary Strip */}
        <div className="tax-inspector-summary-strip">
          <div className="summary-metric-box">
            <span className="summary-metric-label">Open Position</span>
            <span className="summary-metric-value">
              {totalUnits.toLocaleString('en-US', { maximumFractionDigits: 6 })} units
            </span>
          </div>

          <div className="summary-metric-box">
            <span className="summary-metric-label">Total Cost Basis</span>
            <span className="summary-metric-value">
              {formatMoney({ amount: totalCostBasis.toFixed(2), currencyCode: baseCurrency })}
            </span>
          </div>

          <div className="summary-metric-box">
            <span className="summary-metric-label">Market Value</span>
            <span className="summary-metric-value">
              {formatMoney({ amount: totalMarketValue.toFixed(2), currencyCode: baseCurrency })}
            </span>
          </div>

          <div className="summary-metric-box">
            <span className="summary-metric-label">Unrealized P&L</span>
            <span
              className={`summary-metric-value ${
                totalUnrealizedPnl >= 0 ? 'positive' : 'negative'
              }`}
            >
              {totalUnrealizedPnl >= 0 ? '+' : ''}
              {formatMoney({ amount: totalUnrealizedPnl.toFixed(2), currencyCode: baseCurrency })}
            </span>
          </div>

          <div className="summary-metric-box">
            <span className="summary-metric-label">50% CGT Eligible</span>
            <span className="summary-metric-value">
              {cgtEligibleUnits.toLocaleString('en-US', { maximumFractionDigits: 6 })} units
            </span>
          </div>
        </div>

        {/* Filter Controls */}
        <div className="tax-inspector-controls">
          <div className="inspector-filter-pills">
            <button
              type="button"
              className={`filter-pill ${statusFilter === 'ALL' ? 'active' : ''}`}
              onClick={() => setStatusFilter('ALL')}
            >
              All Parcels ({lots.length})
            </button>
            <button
              type="button"
              className={`filter-pill ${statusFilter === 'OPEN' ? 'active' : ''}`}
              onClick={() => setStatusFilter('OPEN')}
            >
              Open Only
            </button>
            <button
              type="button"
              className={`filter-pill ${statusFilter === 'CLOSED' ? 'active' : ''}`}
              onClick={() => setStatusFilter('CLOSED')}
            >
              Closed Only
            </button>
          </div>

          {currentPrice && (
            <div className="current-price-tag">
              <span className="price-tag-label">Current Price: </span>
              <span className="price-tag-value">{formatMoney(currentPrice)}</span>
            </div>
          )}
        </div>

        {/* Table Content */}
        <div className="tax-inspector-table-container">
          {error && <div className="tax-error-banner">⚠️ {error}</div>}

          {loading ? (
            <div className="tax-loading-state">
              <div className="tax-spinner" />
              <p>Loading parcel ledger for {ticker}...</p>
            </div>
          ) : lots.length === 0 ? (
            <div className="inspector-empty-state">
              <p>No tax lots found for this holding.</p>
            </div>
          ) : (
            <table className="inspector-lots-table">
              <thead>
                <tr>
                  <th>Acquired</th>
                  <th>Status</th>
                  <th>Term / Discount</th>
                  <th>Original Qty</th>
                  <th>Remaining Qty</th>
                  <th>Unit Cost</th>
                  <th>Cost Basis</th>
                  <th>Market Value</th>
                  <th>Unrealized Gain / Loss</th>
                  <th>Signals</th>
                </tr>
              </thead>
              <tbody>
                {lots.map((lot) => {
                  const isClosed = parseFloat(lot.remainingQuantity) <= 0;
                  const pnlAmount = parseFloat(lot.unrealizedPnlBase?.amount || '0');
                  const pnlPct = parseFloat(lot.unrealizedPnlPercent || '0');
                  const isPositive = pnlAmount >= 0;

                  return (
                    <tr
                      key={lot.id}
                      className={isClosed ? 'closed-lot-row' : ''}
                    >
                      <td className="acquired-date-cell">{lot.acquiredDate}</td>
                      <td>
                        <span
                          className={`parcel-status-badge ${
                            isClosed ? 'closed' : 'open'
                          }`}
                        >
                          {isClosed ? 'CLOSED' : 'OPEN'}
                        </span>
                      </td>
                      <td>
                        {lot.isCgtDiscountEligible ? (
                          <span className="term-badge long-term">
                            [LONG-TERM] ({lot.holdingPeriodDays}d)
                          </span>
                        ) : (
                          <span className="term-badge short-term">
                            [SHORT-TERM] ({lot.holdingPeriodDays}d /{' '}
                            {Math.max(0, 365 - lot.holdingPeriodDays)}d to 50%)
                          </span>
                        )}
                      </td>
                      <td>{parseFloat(lot.originalQuantity).toLocaleString('en-US', { maximumFractionDigits: 6 })}</td>
                      <td>
                        <strong>
                          {parseFloat(lot.remainingQuantity).toLocaleString('en-US', { maximumFractionDigits: 6 })}
                        </strong>
                      </td>
                      <td>{formatMoney(lot.unitCostBase)}</td>
                      <td>{formatMoney(lot.costBasisBase)}</td>
                      <td>
                        {isClosed ? '—' : formatMoney(lot.marketValueBase)}
                      </td>
                      <td>
                        {isClosed ? (
                          '—'
                        ) : (
                          <span className={isPositive ? 'gain-text' : 'loss-text'}>
                            {isPositive ? '+' : ''}
                            {formatMoney(lot.unrealizedPnlBase)} ({pnlPct.toFixed(2)}%)
                          </span>
                        )}
                      </td>
                      <td>
                        <div className="lot-signals-wrapper">
                          {lot.isLossHarvestingCandidate && (
                            <span
                              className="tlh-badge"
                              title="Unrealized loss > $500 & > -10%. Candidate for capital loss realization."
                            >
                              TLH CANDIDATE
                            </span>
                          )}
                          {lot.hasWashSaleRisk && (
                            <span
                              className="wash-sale-badge"
                              title="Recent purchase within 30 days. Wash sale rules may disallow loss."
                            >
                              WASH SALE RISK
                            </span>
                          )}
                          {!lot.isLossHarvestingCandidate && !lot.hasWashSaleRisk && '—'}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
};
