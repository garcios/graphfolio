import React, { useState, useMemo } from 'react';
import { formatMoney } from '@graphfolio/ui';

export interface TaxLotItem {
  id: string;
  instrumentId: string;
  ticker: string;
  name: string;
  openTransactionId: string;
  acquiredDate: string;
  closedDate?: string | null;
  originalQuantity: string;
  remainingQuantity: string;
  costBasis: { amount: string; currencyCode: string };
  costBasisBase: { amount: string; currencyCode: string };
  unitCostBase: { amount: string; currencyCode: string };
  currentPrice: { amount: string; currencyCode: string };
  marketValueBase: { amount: string; currencyCode: string };
  unrealizedPnlBase: { amount: string; currencyCode: string };
  unrealizedPnlPercent: string;
  holdingPeriodDays: number;
  termClassification: string;
  isCgtDiscountEligible: boolean;
  isLossHarvestingCandidate: boolean;
  hasWashSaleRisk: boolean;
}

interface OpenTaxLotsTableProps {
  lots: TaxLotItem[];
}

export const OpenTaxLotsTable: React.FC<OpenTaxLotsTableProps> = ({ lots }) => {
  const [filterMode, setFilterMode] = useState<'ALL' | 'TLH_ONLY' | 'WASH_SALE_ONLY'>('ALL');
  const [searchTerm, setSearchTerm] = useState('');

  const tlhCount = useMemo(() => lots.filter((l) => l.isLossHarvestingCandidate).length, [lots]);
  const washSaleCount = useMemo(() => lots.filter((l) => l.hasWashSaleRisk).length, [lots]);

  const filteredLots = useMemo(() => {
    return lots.filter((l) => {
      const matchesSearch =
        l.ticker.toLowerCase().includes(searchTerm.toLowerCase()) ||
        l.name.toLowerCase().includes(searchTerm.toLowerCase());
      if (!matchesSearch) return false;

      if (filterMode === 'TLH_ONLY') return l.isLossHarvestingCandidate;
      if (filterMode === 'WASH_SALE_ONLY') return l.hasWashSaleRisk;
      return true;
    });
  }, [lots, searchTerm, filterMode]);

  return (
    <div className="tax-subview-panel" data-testid="open-tax-lots-table">
      <div className="tax-panel-header">
        <div className="tax-panel-title-group">
          <h3 className="tax-panel-title">Open Tax Lots & TLH Scanner</h3>
          <p className="tax-panel-desc">
            Granular active parcel inventory. Track holding duration toward 50% CGT discount and uncover tax-loss harvesting candidates.
          </p>
        </div>

        <div className="tax-panel-actions">
          <div className="search-input-wrapper">
            <input
              type="text"
              placeholder="Filter by ticker or asset..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="tax-search-input"
            />
          </div>

          <div className="filter-pill-group">
            <button
              type="button"
              className={`filter-pill ${filterMode === 'ALL' ? 'active' : ''}`}
              onClick={() => setFilterMode('ALL')}
            >
              All Lots ({lots.length})
            </button>
            <button
              type="button"
              className={`filter-pill pill-tlh ${filterMode === 'TLH_ONLY' ? 'active' : ''}`}
              onClick={() => setFilterMode('TLH_ONLY')}
            >
              🔥 TLH Candidates ({tlhCount})
            </button>
            <button
              type="button"
              className={`filter-pill pill-wash ${filterMode === 'WASH_SALE_ONLY' ? 'active' : ''}`}
              onClick={() => setFilterMode('WASH_SALE_ONLY')}
            >
              ⚠️ Wash-Sale Guard ({washSaleCount})
            </button>
          </div>
        </div>
      </div>

      {washSaleCount > 0 && filterMode !== 'TLH_ONLY' && (
        <div className="wash-sale-banner" role="alert">
          <span className="wash-icon">⚠️</span>
          <div className="wash-text">
            <strong>Wash-Sale / Bed-and-Breakfasting Warning:</strong> One or more assets had a purchase transaction executed within the last 30 days. Realizing a loss on these positions now may disallow or defer the tax deduction under IRS §1091 or HMRC 30-day matching rules.
          </div>
        </div>
      )}

      <div className="table-responsive">
        <table className="tax-table">
          <thead>
            <tr>
              <th>Acquired</th>
              <th>Asset</th>
              <th className="num-col">Remaining / Orig</th>
              <th className="num-col">Unit Cost</th>
              <th className="num-col">Cost Basis (Base)</th>
              <th className="num-col">Price</th>
              <th className="num-col">Market Value</th>
              <th className="num-col">Unrealized PnL</th>
              <th>Holding Duration</th>
              <th>Tax Flags</th>
            </tr>
          </thead>
          <tbody>
            {filteredLots.length === 0 ? (
              <tr>
                <td colSpan={10} className="empty-table-cell">
                  {lots.length === 0
                    ? 'No open tax parcels currently held.'
                    : 'No tax lots match the selected filter.'}
                </td>
              </tr>
            ) : (
              filteredLots.map((l) => {
                const isGain = parseFloat(l.unrealizedPnlBase.amount) >= 0;
                const isLongTerm = l.termClassification === 'LONG_TERM' || l.isCgtDiscountEligible;
                const daysToDiscount = Math.max(0, 365 - l.holdingPeriodDays);

                return (
                  <tr key={l.id} className="tax-data-row">
                    <td className="date-cell">{l.acquiredDate}</td>
                    <td>
                      <div className="tax-asset-cell">
                        <span className="tax-ticker">{l.ticker}</span>
                        <span className="tax-asset-name">{l.name}</span>
                      </div>
                    </td>
                    <td className="num-col font-mono">
                      <span>{l.remainingQuantity}</span>
                      {l.originalQuantity !== l.remainingQuantity && (
                        <span className="orig-qty-subtext"> / {l.originalQuantity}</span>
                      )}
                    </td>
                    <td className="num-col font-mono">{formatMoney(l.unitCostBase)}</td>
                    <td className="num-col font-mono">{formatMoney(l.costBasisBase)}</td>
                    <td className="num-col font-mono">{formatMoney(l.currentPrice)}</td>
                    <td className="num-col font-mono">{formatMoney(l.marketValueBase)}</td>
                    <td className={`num-col font-mono ${isGain ? 'positive-cell' : 'negative-cell'}`}>
                      <div>{isGain ? '+' : ''}{formatMoney(l.unrealizedPnlBase)}</div>
                      <div className="pct-subtext">({parseFloat(l.unrealizedPnlPercent).toFixed(2)}%)</div>
                    </td>
                    <td>
                      {isLongTerm ? (
                        <span className="term-badge term-long" title="Held 365+ days. Eligible for 50% CGT discount.">
                          LONG-TERM ({l.holdingPeriodDays}d)
                        </span>
                      ) : (
                        <span className="term-badge term-countdown" title={`Held ${l.holdingPeriodDays} days. ${daysToDiscount} days remaining until CGT discount.`}>
                          {daysToDiscount}D TO DISCOUNT
                        </span>
                      )}
                    </td>
                    <td>
                      <div className="flags-cell">
                        {l.isLossHarvestingCandidate && (
                          <span
                            className="flag-badge flag-tlh"
                            title="Unrealized loss > $500 and > 10%. Prime candidate to harvest capital loss against realized capital gains."
                          >
                            TLH CANDIDATE
                          </span>
                        )}
                        {l.hasWashSaleRisk && (
                          <span
                            className="flag-badge flag-wash"
                            title="Recent buy transaction within 30 days. Loss realization may trigger wash-sale disallowance."
                          >
                            WASH SALE RISK
                          </span>
                        )}
                        {!l.isLossHarvestingCandidate && !l.hasWashSaleRisk && (
                          <span className="neutral-dash">—</span>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
