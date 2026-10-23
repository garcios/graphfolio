import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney, formatPercent } from '@graphfolio/ui';
import './DRIPSimulatorModal.css';

interface DRIPSimulationPoint {
  year: number;
  dripPortfolioValue: string;
  cashPortfolioValue: string;
  dripAccumulatedShares: string;
  cashAccumulatedCash: string;
  dripAnnualDividend: string;
  cashAnnualDividend: string;
}

interface DRIPSimulationResult {
  startingPortfolioValue: string;
  startingAnnualDividend: string;
  baseCurrency: string;
  horizonYears: number;
  points: DRIPSimulationPoint[];
  terminalDripValue: string;
  terminalCashValue: string;
  dripBenefitAmount: string;
}

export interface DRIPSimulatorModalProps {
  isOpen: boolean;
  onClose: () => void;
  baseCurrency?: string;
}

export const DRIPSimulatorModal: React.FC<DRIPSimulatorModalProps> = ({
  isOpen,
  onClose,
  baseCurrency = 'USD',
}) => {
  const [horizonYears, setHorizonYears] = useState<number>(10);
  const [dividendGrowthRatePct, setDividendGrowthRatePct] = useState<number>(5.0);
  const [capitalAppreciationPct, setCapitalAppreciationPct] = useState<number>(7.0);
  const [result, setResult] = useState<DRIPSimulationResult | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'chart' | 'table'>('chart');
  const [hoveredPoint, setHoveredPoint] = useState<DRIPSimulationPoint | null>(null);

  const fetchSimulation = async (years: number, divGrowth: number, capApp: number) => {
    setLoading(true);
    setError(null);
    try {
      const res = await client.query({
        simulateDRIP: {
          __args: {
            input: {
              horizonYears: years,
              dividendGrowthRatePct: divGrowth.toString(),
              capitalAppreciationPct: capApp.toString(),
            },
          },
          startingPortfolioValue: true,
          startingAnnualDividend: true,
          baseCurrency: true,
          horizonYears: true,
          points: {
            year: true,
            dripPortfolioValue: true,
            cashPortfolioValue: true,
            dripAccumulatedShares: true,
            cashAccumulatedCash: true,
            dripAnnualDividend: true,
            cashAnnualDividend: true,
          },
          terminalDripValue: true,
          terminalCashValue: true,
          dripBenefitAmount: true,
        },
      });

      if (res.simulateDRIP) {
        setResult(res.simulateDRIP as DRIPSimulationResult);
      }
    } catch (err: any) {
      console.error('Failed to simulate DRIP compounding:', err);
      setError(err.message || 'Failed to simulate DRIP compounding');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchSimulation(horizonYears, dividendGrowthRatePct, capitalAppreciationPct);
    }
  }, [isOpen]);

  const handleApply = (e: React.FormEvent) => {
    e.preventDefault();
    fetchSimulation(horizonYears, dividendGrowthRatePct, capitalAppreciationPct);
  };

  if (!isOpen) return null;

  // Chart coordinate calculations
  const chartWidth = 980;
  const chartHeight = 320;
  const padding = { top: 20, right: 30, bottom: 40, left: 75 };
  const innerWidth = chartWidth - padding.left - padding.right;
  const innerHeight = chartHeight - padding.top - padding.bottom;

  let maxVal = 1000;
  let minVal = 0;
  if (result && result.points.length > 0) {
    const dripVals = result.points.map((p) => Number(p.dripPortfolioValue));
    const cashVals = result.points.map((p) => Number(p.cashPortfolioValue));
    maxVal = Math.max(...dripVals, ...cashVals, Number(result.startingPortfolioValue)) * 1.05;
    minVal = Math.min(...dripVals, ...cashVals, Number(result.startingPortfolioValue)) * 0.95;
    if (minVal < 0) minVal = 0;
  }

  const getX = (year: number) => {
    if (!result || result.horizonYears === 0) return padding.left;
    return padding.left + (year / result.horizonYears) * innerWidth;
  };

  const getY = (val: number) => {
    if (maxVal === minVal) return padding.top + innerHeight / 2;
    const ratio = (val - minVal) / (maxVal - minVal);
    return padding.top + innerHeight - ratio * innerHeight;
  };

  const startingVal = result ? Number(result.startingPortfolioValue) : 0;
  const startPtX = getX(0);
  const startPtY = getY(startingVal);

  const dripPoints = result ? [[startPtX, startPtY], ...result.points.map((p) => [getX(p.year), getY(Number(p.dripPortfolioValue))])] : [];
  const cashPoints = result ? [[startPtX, startPtY], ...result.points.map((p) => [getX(p.year), getY(Number(p.cashPortfolioValue))])] : [];

  const dripPath = dripPoints.map((pt, i) => `${i === 0 ? 'M' : 'L'} ${pt[0]} ${pt[1]}`).join(' ');
  const cashPath = cashPoints.map((pt, i) => `${i === 0 ? 'M' : 'L'} ${pt[0]} ${pt[1]}`).join(' ');

  const dripArea = dripPoints.length > 0
    ? `${dripPath} L ${dripPoints[dripPoints.length - 1][0]} ${padding.top + innerHeight} L ${startPtX} ${padding.top + innerHeight} Z`
    : '';

  const terminalDrip = result ? Number(result.terminalDripValue) : 0;
  const terminalCash = result ? Number(result.terminalCashValue) : 0;
  const benefitPct = terminalCash > 0 ? ((terminalDrip - terminalCash) / terminalCash) * 100 : 0;
  const currencyCode = result?.baseCurrency || baseCurrency;

  return (
    <div className="modal-backdrop" onClick={onClose} role="dialog" aria-modal="true" aria-labelledby="drip-modal-title">
      <div className="modal-container drip-modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <div className="drip-modal-header-group">
            <span className="drip-badge">⚡ COMPOUNDING ENGINE</span>
            <h3 id="drip-modal-title" className="modal-title">DRIP Compounding Simulator</h3>
            <p className="drip-modal-subtitle">
              Forecast exponential wealth accumulation by comparing dividend reinvestment against cash withdrawals.
            </p>
          </div>
          <button
            type="button"
            className="modal-close-btn"
            onClick={onClose}
            aria-label="Close modal"
          >
            &times;
          </button>
        </div>

        {/* Input Parameters Form */}
        <form className="drip-params-card" onSubmit={handleApply}>
          <div className="drip-input-group">
            <label htmlFor="horizon-input">
              Horizon: <strong>{horizonYears} Years</strong>
            </label>
            <div className="drip-slider-wrap">
              <input
                id="horizon-input"
                type="range"
                min="1"
                max="30"
                value={horizonYears}
                onChange={(e) => setHorizonYears(parseInt(e.target.value, 10))}
                className="drip-range-slider"
              />
              <span className="drip-range-val">{horizonYears}y</span>
            </div>
          </div>

          <div className="drip-input-group">
            <label htmlFor="div-growth-input">Annual Dividend Growth Rate</label>
            <div className="drip-pct-input-wrap">
              <input
                id="div-growth-input"
                type="number"
                step="0.1"
                min="-10"
                max="50"
                value={dividendGrowthRatePct}
                onChange={(e) => setDividendGrowthRatePct(parseFloat(e.target.value) || 0)}
                className="drip-num-input"
              />
              <span className="drip-unit-suffix">%</span>
            </div>
          </div>

          <div className="drip-input-group">
            <label htmlFor="cap-app-input">Annual Capital Appreciation</label>
            <div className="drip-pct-input-wrap">
              <input
                id="cap-app-input"
                type="number"
                step="0.1"
                min="-20"
                max="50"
                value={capitalAppreciationPct}
                onChange={(e) => setCapitalAppreciationPct(parseFloat(e.target.value) || 0)}
                className="drip-num-input"
              />
              <span className="drip-unit-suffix">%</span>
            </div>
          </div>

          <button type="submit" className="drip-simulate-btn" disabled={loading}>
            {loading ? 'Simulating...' : 'Recalculate'}
          </button>
        </form>

        {error && <div className="drip-error-banner">{error}</div>}

        {result && (
          <div className="drip-results-body">
            {/* KPI Summary Cards */}
            <div className="drip-kpi-grid">
              <div className="drip-kpi-card">
                <span className="drip-kpi-label">Starting Portfolio</span>
                <span className="drip-kpi-val">
                  {formatMoney(result.startingPortfolioValue, currencyCode)}
                </span>
                <span className="drip-kpi-sub">
                  Initial Income: {formatMoney(result.startingAnnualDividend, currencyCode)}/yr
                </span>
              </div>

              <div className="drip-kpi-card highlight-drip">
                <span className="drip-kpi-label">Terminal DRIP Value ({result.horizonYears}y)</span>
                <span className="drip-kpi-val text-emerald">
                  {formatMoney(result.terminalDripValue, currencyCode)}
                </span>
                <span className="drip-kpi-sub">100% Reinvested into Compounding Shares</span>
              </div>

              <div className="drip-kpi-card">
                <span className="drip-kpi-label">Terminal Cash Strategy ({result.horizonYears}y)</span>
                <span className="drip-kpi-val text-sky">
                  {formatMoney(result.terminalCashValue, currencyCode)}
                </span>
                <span className="drip-kpi-sub">Stock Value + Accumulated Cash</span>
              </div>

              <div className="drip-kpi-card highlight-benefit">
                <span className="drip-kpi-label">DRIP Compounding Advantage</span>
                <span className="drip-kpi-val text-emerald">
                  +{formatMoney(result.dripBenefitAmount, currencyCode)}
                </span>
                <span className="drip-kpi-sub text-emerald">
                  +{formatPercent(benefitPct.toString(), 1, false)} Net Wealth Advantage vs Cash
                </span>
              </div>
            </div>

            {/* View Switcher Tabs */}
            <div className="drip-view-tabs">
              <button
                type="button"
                className={`drip-tab-btn ${activeTab === 'chart' ? 'active' : ''}`}
                onClick={() => setActiveTab('chart')}
              >
                📈 Compounding Curve
              </button>
              <button
                type="button"
                className={`drip-tab-btn ${activeTab === 'table' ? 'active' : ''}`}
                onClick={() => setActiveTab('table')}
              >
                📋 Progression Matrix
              </button>
            </div>

            {/* Chart View */}
            {activeTab === 'chart' && (
              <div className="drip-chart-container">
                <div className="drip-chart-legend">
                  <span className="legend-item drip-legend">
                    <span className="legend-dot dot-drip" />
                    <strong>DRIP Reinvestment</strong> ({formatMoney(result.terminalDripValue, currencyCode)})
                  </span>
                  <span className="legend-item cash-legend">
                    <span className="legend-dot dot-cash" />
                    <strong>Cash Strategy (Stock + Cash)</strong> ({formatMoney(result.terminalCashValue, currencyCode)})
                  </span>
                </div>

                <div className="drip-svg-wrapper">
                  <svg
                    viewBox={`0 0 ${chartWidth} ${chartHeight}`}
                    className="drip-chart-svg"
                    preserveAspectRatio="none"
                  >
                    <defs>
                      <linearGradient id="dripGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor="#10b981" stopOpacity="0.3" />
                        <stop offset="100%" stopColor="#10b981" stopOpacity="0.0" />
                      </linearGradient>
                    </defs>

                    {/* Horizontal gridlines */}
                    {[0, 0.25, 0.5, 0.75, 1].map((pct, idx) => {
                      const y = padding.top + innerHeight * (1 - pct);
                      const val = minVal + (maxVal - minVal) * pct;
                      return (
                        <g key={idx}>
                          <line
                            x1={padding.left}
                            y1={y}
                            x2={padding.left + innerWidth}
                            y2={y}
                            stroke="rgba(255, 255, 255, 0.08)"
                            strokeDasharray="4 4"
                          />
                          <text
                            x={padding.left - 8}
                            y={y + 4}
                            textAnchor="end"
                            fontSize="10"
                            fill="rgba(255, 255, 255, 0.45)"
                          >
                            {formatMoney(val.toFixed(0), currencyCode, 0)}
                          </text>
                        </g>
                      );
                    })}

                    {/* Area fill under DRIP curve */}
                    {dripArea && <path d={dripArea} fill="url(#dripGrad)" />}

                    {/* Cash Payout Line */}
                    <path
                      d={cashPath}
                      fill="none"
                      stroke="#38bdf8"
                      strokeWidth="2.5"
                      strokeDasharray="6 3"
                    />

                    {/* DRIP Line */}
                    <path
                      d={dripPath}
                      fill="none"
                      stroke="#10b981"
                      strokeWidth="3"
                    />

                    {/* Interactive Points on DRIP line */}
                    {result.points.map((pt) => {
                      const cx = getX(pt.year);
                      const cy = getY(Number(pt.dripPortfolioValue));
                      const isHovered = hoveredPoint?.year === pt.year;
                      return (
                        <circle
                          key={pt.year}
                          cx={cx}
                          cy={cy}
                          r={isHovered ? 6 : 3.5}
                          fill="#10b981"
                          stroke="#ffffff"
                          strokeWidth="1.5"
                          className="drip-chart-point"
                          onMouseEnter={() => setHoveredPoint(pt)}
                          onMouseLeave={() => setHoveredPoint(null)}
                        />
                      );
                    })}

                    {/* X-axis labels */}
                    {result.points.map((pt) => {
                      if (pt.year % Math.max(1, Math.floor(result.horizonYears / 5)) === 0 || pt.year === result.horizonYears) {
                        return (
                          <text
                            key={`lbl-${pt.year}`}
                            x={getX(pt.year)}
                            y={padding.top + innerHeight + 20}
                            textAnchor="middle"
                            fontSize="11"
                            fill="rgba(255, 255, 255, 0.6)"
                          >
                            Yr {pt.year}
                          </text>
                        );
                      }
                      return null;
                    })}
                  </svg>

                  {/* Tooltip on hover */}
                  {hoveredPoint && (
                    <div
                      className="drip-chart-tooltip"
                      style={{
                        left: `${(getX(hoveredPoint.year) / chartWidth) * 100}%`,
                        top: '15px',
                      }}
                    >
                      <div className="tooltip-year">Year {hoveredPoint.year}</div>
                      <div className="tooltip-row text-emerald">
                        <span>DRIP Value:</span>
                        <strong>{formatMoney(hoveredPoint.dripPortfolioValue, currencyCode)}</strong>
                      </div>
                      <div className="tooltip-row text-sky">
                        <span>Cash Strategy:</span>
                        <strong>{formatMoney(hoveredPoint.cashPortfolioValue, currencyCode)}</strong>
                      </div>
                      <div className="tooltip-row">
                        <span>Cash Accumulated:</span>
                        <strong>{formatMoney(hoveredPoint.cashAccumulatedCash, currencyCode)}</strong>
                      </div>
                      <div className="tooltip-row">
                        <span>Annual Div:</span>
                        <strong>{formatMoney(hoveredPoint.dripAnnualDividend, currencyCode)}</strong>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Table View */}
            {activeTab === 'table' && (
              <div className="drip-table-wrapper">
                <table className="drip-data-table">
                  <thead>
                    <tr>
                      <th>Year</th>
                      <th>DRIP Portfolio Value</th>
                      <th>Cash Portfolio Value</th>
                      <th>DRIP Shares</th>
                      <th>Cash Accumulated</th>
                      <th>DRIP Annual Div</th>
                      <th>Cash Annual Div</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.points.map((pt) => (
                      <tr key={pt.year}>
                        <td>
                          <strong>Yr {pt.year}</strong>
                        </td>
                        <td className="text-emerald font-semibold">
                          {formatMoney(pt.dripPortfolioValue, currencyCode)}
                        </td>
                        <td className="text-sky">
                          {formatMoney(pt.cashPortfolioValue, currencyCode)}
                        </td>
                        <td>{Number(pt.dripAccumulatedShares).toFixed(2)}</td>
                        <td>{formatMoney(pt.cashAccumulatedCash, currencyCode)}</td>
                        <td className="text-emerald">
                          {formatMoney(pt.dripAnnualDividend, currencyCode)}
                        </td>
                        <td>{formatMoney(pt.cashAnnualDividend, currencyCode)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
