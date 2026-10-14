import React from 'react';
import './BenchmarkMetricsStrip.css';

export interface BenchmarkMetricsData {
  portfolioCumulativeReturnPercent: string;
  benchmarkCumulativeReturnPercent: string;
  activeCumulativeReturnPercent: string;
  portfolioAnnualizedReturnPercent?: string | null;
  benchmarkAnnualizedReturnPercent?: string | null;
  activeAnnualizedReturnPercent?: string | null;
  isAnnualized: boolean;
  beta?: string | null;
  trackingErrorPercent?: string | null;
  informationRatio?: string | null;
}

export interface BenchmarkInfo {
  id: string;
  symbol: string;
  name: string;
  currencyCode: string;
  indexType: string;
}

interface BenchmarkMetricsStripProps {
  benchmark: BenchmarkInfo;
  metrics: BenchmarkMetricsData;
  currencyMode: 'PORTFOLIO_BASE' | 'BENCHMARK_NATIVE';
  baseCurrency: string;
  elapsedDays: number;
}

export const BenchmarkMetricsStrip: React.FC<BenchmarkMetricsStripProps> = ({
  benchmark,
  metrics,
  currencyMode,
  baseCurrency,
  elapsedDays,
}) => {
  const isAnn = metrics.isAnnualized;
  const activePercentVal = isAnn && metrics.activeAnnualizedReturnPercent != null
    ? Number(metrics.activeAnnualizedReturnPercent)
    : Number(metrics.activeCumulativeReturnPercent);

  const isPositive = activePercentVal >= 0;

  const portVal = isAnn && metrics.portfolioAnnualizedReturnPercent != null
    ? Number(metrics.portfolioAnnualizedReturnPercent)
    : Number(metrics.portfolioCumulativeReturnPercent);

  const benchVal = isAnn && metrics.benchmarkAnnualizedReturnPercent != null
    ? Number(metrics.benchmarkAnnualizedReturnPercent)
    : Number(metrics.benchmarkCumulativeReturnPercent);

  const displayCurrency = currencyMode === 'PORTFOLIO_BASE' ? baseCurrency : benchmark.currencyCode;

  return (
    <div className="benchmark-metrics-strip" role="region" aria-label="Benchmark Comparison Metrics">
      {/* Active Alpha Badge */}
      <div className={`benchmark-alpha-pill ${isPositive ? 'positive' : 'negative'}`}>
        <span className="alpha-badge-icon" aria-hidden="true">
          {isPositive ? '▲' : '▼'}
        </span>
        <div className="alpha-badge-text">
          <span className="alpha-badge-title">
            {isPositive ? '+' : ''}{activePercentVal.toFixed(2)}% {isAnn ? 'Annualized Alpha' : 'Active Spread'}
          </span>
          <span className="alpha-badge-subtitle">
            {isPositive ? 'Outperforming' : 'Underperforming'} vs {benchmark.symbol} ({isAnn ? 'p.a.' : `${elapsedDays}d`})
          </span>
        </div>
      </div>

      {/* Attribution Statistics */}
      <div className="benchmark-stats-row">
        {/* Portfolio Return */}
        <div className="benchmark-stat-card">
          <span className="stat-label">Portfolio {isAnn ? 'p.a.' : 'Return'}</span>
          <span className={`stat-value ${portVal >= 0 ? 'text-positive' : 'text-negative'}`}>
            {portVal >= 0 ? '+' : ''}{portVal.toFixed(2)}%
          </span>
        </div>

        {/* Benchmark Return */}
        <div className="benchmark-stat-card">
          <span className="stat-label">
            {benchmark.symbol} ({displayCurrency})
          </span>
          <span className={`stat-value ${benchVal >= 0 ? 'text-positive' : 'text-negative'}`}>
            {benchVal >= 0 ? '+' : ''}{benchVal.toFixed(2)}%
          </span>
        </div>

        {/* Beta */}
        <div className="benchmark-stat-card" title="Beta (β): Sensitivity of portfolio returns to benchmark movements. >1.0 implies higher volatility than the index.">
          <div className="stat-label-group">
            <span className="stat-label">Beta (β)</span>
            <span className="info-icon" aria-label="Beta info">ℹ</span>
          </div>
          <span className="stat-value font-mono">
            {metrics.beta ? Number(metrics.beta).toFixed(2) : '—'}
          </span>
        </div>

        {/* Tracking Error */}
        <div className="benchmark-stat-card" title="Tracking Error (TE): Annualized standard deviation of excess returns over benchmark.">
          <div className="stat-label-group">
            <span className="stat-label">Tracking Error</span>
            <span className="info-icon" aria-label="Tracking error info">ℹ</span>
          </div>
          <span className="stat-value font-mono">
            {metrics.trackingErrorPercent ? `${Number(metrics.trackingErrorPercent).toFixed(2)}%` : '—'}
          </span>
        </div>

        {/* Information Ratio */}
        <div className="benchmark-stat-card" title="Information Ratio (IR): Active return per unit of active risk (annualized alpha / tracking error).">
          <div className="stat-label-group">
            <span className="stat-label">Info Ratio</span>
            <span className="info-icon" aria-label="Information ratio info">ℹ</span>
          </div>
          <span className="stat-value font-mono">
            {metrics.informationRatio ? Number(metrics.informationRatio).toFixed(2) : '—'}
          </span>
        </div>
      </div>
    </div>
  );
};
