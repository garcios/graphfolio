import React, { useState, useMemo } from 'react';
import { formatMoney } from '@graphfolio/ui';

export interface RealizedDisposalItem {
  disposalId: string;
  taxLotId: string;
  sellTransactionId: string;
  instrumentId: string;
  ticker: string;
  name: string;
  acquiredDate: string;
  disposedDate: string;
  quantityDisposed: string;
  costBasisRelievedBase: { amount: string; currencyCode: string };
  proceedsBase: { amount: string; currencyCode: string };
  realizedPnlBase: { amount: string; currencyCode: string };
  realizedPnlPercent: string;
  holdingPeriodDays: number;
  termClassification: string;
  isCgtDiscountEligible: boolean;
  discountedTaxablePnlBase: { amount: string; currencyCode: string };
}

interface RealizedDisposalsTableProps {
  disposals: RealizedDisposalItem[];
  onExportCSV: () => void;
  isExporting?: boolean;
}

export const RealizedDisposalsTable: React.FC<RealizedDisposalsTableProps> = ({
  disposals,
  onExportCSV,
  isExporting = false,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [termFilter, setTermFilter] = useState<'ALL' | 'LONG_TERM' | 'SHORT_TERM'>('ALL');

  const filteredDisposals = useMemo(() => {
    return disposals.filter((d) => {
      const matchesSearch =
        d.ticker.toLowerCase().includes(searchTerm.toLowerCase()) ||
        d.name.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesTerm =
        termFilter === 'ALL' || d.termClassification === termFilter;
      return matchesSearch && matchesTerm;
    });
  }, [disposals, searchTerm, termFilter]);

  return (
    <div className="tax-subview-panel" data-testid="realized-disposals-table">
      <div className="tax-panel-header">
        <div className="tax-panel-title-group">
          <h3 className="tax-panel-title">Realized Disposals Schedule</h3>
          <p className="tax-panel-desc">
            Granular lot-level depletion audit trail. Every sale is mapped back to its acquisition parcel with exact cost relief.
          </p>
        </div>

        <div className="tax-panel-actions">
          <div className="search-input-wrapper">
            <input
              type="text"
              placeholder="Search ticker or asset..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="tax-search-input"
            />
          </div>

          <div className="filter-pill-group">
            <button
              type="button"
              className={`filter-pill ${termFilter === 'ALL' ? 'active' : ''}`}
              onClick={() => setTermFilter('ALL')}
            >
              All ({disposals.length})
            </button>
            <button
              type="button"
              className={`filter-pill ${termFilter === 'LONG_TERM' ? 'active' : ''}`}
              onClick={() => setTermFilter('LONG_TERM')}
            >
              Long-Term
            </button>
            <button
              type="button"
              className={`filter-pill ${termFilter === 'SHORT_TERM' ? 'active' : ''}`}
              onClick={() => setTermFilter('SHORT_TERM')}
            >
              Short-Term
            </button>
          </div>

          <button
            type="button"
            className="export-csv-btn"
            onClick={onExportCSV}
            disabled={isExporting || disposals.length === 0}
            title="Download statutory audit schedule as RFC 4180 CSV"
          >
            {isExporting ? 'Exporting...' : '📥 Export CSV'}
          </button>
        </div>
      </div>

      <div className="table-responsive">
        <table className="tax-table">
          <thead>
            <tr>
              <th>Disposed</th>
              <th>Asset</th>
              <th>Acquired</th>
              <th className="num-col">Held Days</th>
              <th>Term</th>
              <th className="num-col">Qty Relieved</th>
              <th className="num-col">Cost Basis (Base)</th>
              <th className="num-col">Gross Proceeds</th>
              <th className="num-col">Realized PnL</th>
              <th className="num-col">CGT Discount</th>
              <th className="num-col">Taxable Net</th>
            </tr>
          </thead>
          <tbody>
            {filteredDisposals.length === 0 ? (
              <tr>
                <td colSpan={11} className="empty-table-cell">
                  {disposals.length === 0
                    ? 'No realized disposals found for this period.'
                    : 'No disposals match the search criteria.'}
                </td>
              </tr>
            ) : (
              filteredDisposals.map((d) => {
                const isGain = parseFloat(d.realizedPnlBase.amount) >= 0;
                const isLongTerm = d.termClassification === 'LONG_TERM';

                return (
                  <tr key={d.disposalId} className="tax-data-row">
                    <td className="date-cell">{d.disposedDate}</td>
                    <td>
                      <div className="tax-asset-cell">
                        <span className="tax-ticker">{d.ticker}</span>
                        <span className="tax-asset-name">{d.name}</span>
                      </div>
                    </td>
                    <td className="date-cell">{d.acquiredDate}</td>
                    <td className="num-col">{d.holdingPeriodDays}d</td>
                    <td>
                      <span className={`term-badge ${isLongTerm ? 'term-long' : 'term-short'}`}>
                        {isLongTerm ? 'LONG-TERM' : 'SHORT-TERM'}
                      </span>
                    </td>
                    <td className="num-col font-mono">{d.quantityDisposed}</td>
                    <td className="num-col font-mono">{formatMoney(d.costBasisRelievedBase)}</td>
                    <td className="num-col font-mono">{formatMoney(d.proceedsBase)}</td>
                    <td className={`num-col font-mono ${isGain ? 'positive-cell' : 'negative-cell'}`}>
                      <div>{isGain ? '+' : ''}{formatMoney(d.realizedPnlBase)}</div>
                      <div className="pct-subtext">({parseFloat(d.realizedPnlPercent).toFixed(2)}%)</div>
                    </td>
                    <td className="num-col font-mono">
                      {d.isCgtDiscountEligible && isGain ? (
                        <span className="discount-applied-text">
                          -{formatMoney({
                            amount: (parseFloat(d.realizedPnlBase.amount) * 0.50).toFixed(2),
                            currencyCode: d.realizedPnlBase.currencyCode,
                          })}
                        </span>
                      ) : (
                        <span className="neutral-dash">—</span>
                      )}
                    </td>
                    <td className={`num-col font-mono font-bold ${isGain ? 'positive-cell' : 'negative-cell'}`}>
                      {formatMoney(d.discountedTaxablePnlBase)}
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
