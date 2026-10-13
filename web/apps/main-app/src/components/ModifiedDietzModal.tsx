import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney, formatPercent, Button } from '@graphfolio/ui';
import './ModifiedDietzModal.css';

interface ModifiedDietzModalProps {
  isOpen: boolean;
  onClose: () => void;
  portfolioCurrency?: string;
  inceptionDate?: string;
}

type PresetRange = '1Y' | '3Y' | 'ALL' | 'CUSTOM';

export const ModifiedDietzModal: React.FC<ModifiedDietzModalProps> = ({
  isOpen,
  onClose,
  portfolioCurrency = 'USD',
  inceptionDate,
}) => {
  const todayStr = new Date().toISOString().split('T')[0];
  const oneYearAgoStr = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

  const [activePreset, setActivePreset] = useState<PresetRange>('1Y');
  const [startDate, setStartDate] = useState(oneYearAgoStr);
  const [endDate, setEndDate] = useState(todayStr);
  const [useMonthlyLinking, setUseMonthlyLinking] = useState(true);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dietzData, setDietzData] = useState<any | null>(null);

  const handlePresetChange = (preset: PresetRange) => {
    setActivePreset(preset);
    const end = new Date().toISOString().split('T')[0];
    setEndDate(end);

    if (preset === '1Y') {
      const start = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      setStartDate(start);
    } else if (preset === '3Y') {
      const start = new Date(Date.now() - 3 * 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      setStartDate(start);
    } else if (preset === 'ALL') {
      // Historical inception start: pre-select to true portfolio inception date
      setStartDate(inceptionDate || dietzData?.startDate || '2020-01-01');
    }
  };

  const fetchModifiedDietz = () => {
    if (!startDate || !endDate) return;
    setLoading(true);
    setError(null);

    client
      .query({
        modifiedDietzReturn: {
          __args: {
            startDate,
            endDate,
            useMonthlyLinking,
          },
          portfolioId: true,
          baseCurrency: true,
          startDate: true,
          endDate: true,
          elapsedDays: true,
          startValuation: { amount: true, currencyCode: true },
          endValuation: { amount: true, currencyCode: true },
          netExternalFlows: { amount: true, currencyCode: true },
          netInvestmentGain: { amount: true, currencyCode: true },
          weightedCapitalEmployed: { amount: true, currencyCode: true },
          cumulativeReturnPercent: true,
          annualizedReturnPercent: true,
          isAnnualized: true,
          isComputable: true,
          diagnosticCode: true,
          cashFlowWeights: {
            transactionId: true,
            date: true,
            type: true,
            amount: { amount: true, currencyCode: true },
            weight: true,
            weightedAmount: { amount: true, currencyCode: true },
            localAmount: { amount: true, currencyCode: true },
          },
        },
      })
      .then((res) => {
        const payload = res.modifiedDietzReturn;
        setDietzData(payload);
        if (payload?.startDate && payload.startDate !== startDate && activePreset === 'ALL') {
          setStartDate(payload.startDate);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to load Modified Dietz metrics:', err);
        setError(err.message || 'Failed to calculate Modified Dietz return');
        setLoading(false);
      });
  };

  useEffect(() => {
    if (isOpen) {
      fetchModifiedDietz();
    }
  }, [isOpen, startDate, endDate, useMonthlyLinking]);

  if (!isOpen) return null;

  const cumReturnVal = parseFloat(dietzData?.cumulativeReturnPercent) || 0;
  const annReturnVal = parseFloat(dietzData?.annualizedReturnPercent) || 0;
  const isGainPositive = parseFloat(dietzData?.netInvestmentGain?.amount) >= 0;

  return (
    <div className="dietz-modal-backdrop" onClick={onClose} role="dialog" aria-modal="true">
      <div className="dietz-modal-card" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="dietz-modal-header">
          <div className="dietz-header-title-group">
            <div className="dietz-badge-row">
              <span className="dietz-badge">METHODOLOGY INSPECTOR</span>
              {dietzData?.isAnnualized ? (
                <span className="gips-badge annualized" title="Conforms to GIPS Standard 2.A.20 (Geometric CAGR)">
                  GIPS ANNUALIZED (Actual/365.25)
                </span>
              ) : (
                <span className="gips-badge period" title="Per GIPS Standard 2.A.20, periods under 365 days must not be annualized">
                  GIPS PERIOD RETURN (&lt; 1 Year)
                </span>
              )}
            </div>
            <h2>Modified Dietz Return Calculator</h2>
            <p className="dietz-subtitle">
              Calculates cash-flow-weighted rate of return in {portfolioCurrency} by time-weighting every external deposit and withdrawal.
            </p>
          </div>
          <button className="dietz-close-btn" onClick={onClose} aria-label="Close modal">
            ✕
          </button>
        </div>

        {/* Date Range & Controls */}
        <div className="dietz-controls-bar">
          <div className="dietz-presets">
            {(['1Y', '3Y', 'ALL', 'CUSTOM'] as PresetRange[]).map((preset) => (
              <button
                key={preset}
                type="button"
                className={`preset-btn ${activePreset === preset ? 'active' : ''}`}
                onClick={() => handlePresetChange(preset)}
              >
                {preset}
              </button>
            ))}
          </div>

          <div className="dietz-date-inputs">
            <label className="date-field">
              <span>From:</span>
              <input
                type="date"
                value={startDate}
                max={endDate}
                onChange={(e) => {
                  setStartDate(e.target.value);
                  setActivePreset('CUSTOM');
                }}
              />
            </label>
            <label className="date-field">
              <span>To:</span>
              <input
                type="date"
                value={endDate}
                min={startDate}
                onChange={(e) => {
                  setEndDate(e.target.value);
                  setActivePreset('CUSTOM');
                }}
              />
            </label>
            <label className="toggle-monthly" title="Geometrically link monthly sub-period Modified Dietz calculations for multi-year horizons">
              <input
                type="checkbox"
                checked={useMonthlyLinking}
                onChange={(e) => setUseMonthlyLinking(e.target.checked)}
              />
              <span>Monthly sub-period linking</span>
            </label>
          </div>
        </div>

        {/* Error / Loading */}
        {error && <div className="dietz-error-banner">{error}</div>}

        {loading ? (
          <div className="dietz-loading-state">
            <div className="spinner"></div>
            <span>Evaluating cash-flow weights across valuation ledger...</span>
          </div>
        ) : dietzData ? (
          <div className="dietz-body">
            {/* KPI Summary Cards */}
            <div className="dietz-kpi-grid">
              <div className="kpi-card highlight">
                <span className="kpi-label">
                  {dietzData.isAnnualized ? 'Annualized Return (Dietz)' : 'Cumulative Period Return'}
                </span>
                <span className={`kpi-value ${dietzData.isAnnualized ? (annReturnVal >= 0 ? 'positive' : 'negative') : (cumReturnVal >= 0 ? 'positive' : 'negative')}`}>
                  {dietzData.isAnnualized
                    ? `${annReturnVal >= 0 ? '+' : ''}${formatPercent(dietzData.annualizedReturnPercent, 2)} p.a.`
                    : `${cumReturnVal >= 0 ? '+' : ''}${formatPercent(dietzData.cumulativeReturnPercent, 2)}`}
                </span>
                <span className="kpi-subtext">
                  {dietzData.elapsedDays} calendar days elapsed
                </span>
              </div>

              <div className="kpi-card">
                <span className="kpi-label">Cumulative Return</span>
                <span className={`kpi-value ${cumReturnVal >= 0 ? 'positive' : 'negative'}`}>
                  {cumReturnVal >= 0 ? '+' : ''}{formatPercent(dietzData.cumulativeReturnPercent, 2)}
                </span>
                <span className="kpi-subtext">
                  Total economic gain / Weighted capital
                </span>
              </div>

              <div className="kpi-card">
                <span className="kpi-label">Net Investment Gain</span>
                <span className={`kpi-value ${isGainPositive ? 'positive' : 'negative'}`}>
                  {isGainPositive ? '+' : ''}{formatMoney(dietzData.netInvestmentGain)}
                </span>
                <span className="kpi-subtext">
                  Ending value minus start value &amp; net flows
                </span>
              </div>

              <div className="kpi-card">
                <span className="kpi-label">Weighted Average Capital</span>
                <span className="kpi-value neutral">
                  {formatMoney(dietzData.weightedCapitalEmployed)}
                </span>
                <span className="kpi-subtext">
                  Denominator: Start value + ∑ (W_i × F_i)
                </span>
              </div>
            </div>

            {/* Formula & Balance Decomposition */}
            <div className="dietz-breakdown-row">
              <div className="decomposition-card">
                <h4>Capital Balance Components</h4>
                <div className="component-row">
                  <span className="label">Starting Valuation (V_start):</span>
                  <span className="val">{formatMoney(dietzData.startValuation)}</span>
                </div>
                <div className="component-row">
                  <span className="label">Ending Valuation (V_end):</span>
                  <span className="val">{formatMoney(dietzData.endValuation)}</span>
                </div>
                <div className="component-row">
                  <span className="label">Total Net External Flows (F):</span>
                  <span className="val">{formatMoney(dietzData.netExternalFlows)}</span>
                </div>
                <div className="component-row border-top">
                  <span className="label font-bold">Net Profit / Investment Gain:</span>
                  <span className={`val font-bold ${isGainPositive ? 'positive' : 'negative'}`}>
                    {isGainPositive ? '+' : ''}{formatMoney(dietzData.netInvestmentGain)}
                  </span>
                </div>
              </div>

              <div className="formula-explanation-card">
                <h4>Formula &amp; Time Weighting Logic</h4>
                <div className="formula-math">
                  <code>
                    R_Dietz = (V_end - V_start - F) / (V_start + ∑ (W_i × F_i))
                  </code>
                </div>
                <p className="formula-desc">
                  Each deposit or withdrawal <code>F_i</code> occurring at day <code>D_i</code> is weighted by
                  <code> W_i = (CD - D_i) / CD</code>, where <code>CD = {dietzData.elapsedDays}</code> days.
                  Flows introduced earlier work longer in the portfolio, receiving proportionally higher denominator weight.
                </p>
              </div>
            </div>

            {/* Itemized External Cash Flows Table */}
            <div className="dietz-table-section">
              <div className="dietz-table-header">
                <h4>Time-Weighted External Cash Flows ({dietzData.cashFlowWeights?.length || 0})</h4>
                <span className="table-caption">
                  Only capital injections (`DEPOSIT`) and extractions (`WITHDRAWAL`) enter the equation. Internal trades and dividends are isolated.
                </span>
              </div>

              {dietzData.cashFlowWeights && dietzData.cashFlowWeights.length > 0 ? (
                <div className="table-responsive">
                  <table className="dietz-flows-table">
                    <thead>
                      <tr>
                        <th>Date</th>
                        <th>Type</th>
                        <th className="num-col">Cash Flow Amount</th>
                        <th className="num-col">Time Weight (W_i)</th>
                        <th className="num-col">Weighted Contribution</th>
                      </tr>
                    </thead>
                    <tbody>
                      {dietzData.cashFlowWeights.map((w: any) => {
                        const isDeposit = w.type === 'DEPOSIT' || w.type === 'TRANSFER_IN';
                        return (
                          <tr key={w.transactionId}>
                            <td>{w.date}</td>
                            <td>
                              <span className={`flow-type-pill ${isDeposit ? 'inflow' : 'outflow'}`}>
                                {w.type}
                              </span>
                            </td>
                            <td className="num-col font-mono">
                              {formatMoney(w.amount)}
                            </td>
                            <td className="num-col font-mono text-cyan">
                              {(parseFloat(w.weight) || 0).toFixed(6)}
                            </td>
                            <td className="num-col font-mono font-bold">
                              {formatMoney(w.weightedAmount)}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="dietz-empty-flows">
                  <p>
                    No external deposits or withdrawals occurred in this window.
                    The Modified Dietz return is mathematically identical to the simple holding period return.
                  </p>
                </div>
              )}
            </div>
          </div>
        ) : null}

        {/* Footer */}
        <div className="dietz-modal-footer">
          <Button variant="secondary" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </div>
  );
};
