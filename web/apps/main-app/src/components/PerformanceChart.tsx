import React, { useEffect, useRef, useState } from 'react';
import { client } from '@graphfolio/api-client';
import { BenchmarkMetricsStrip } from './BenchmarkMetricsStrip';
import type { BenchmarkMetricsData, BenchmarkInfo } from './BenchmarkMetricsStrip';
import './PerformanceChart.css';

export type TimeframeKey = '1D' | '1W' | '1M' | '1Y' | 'ALL';

const TIMEFRAME_MAP: Record<TimeframeKey, 'TIMEFRAME_1D' | 'TIMEFRAME_1W' | 'TIMEFRAME_1M' | 'TIMEFRAME_1Y' | 'TIMEFRAME_ALL'> = {
  '1D': 'TIMEFRAME_1D',
  '1W': 'TIMEFRAME_1W',
  '1M': 'TIMEFRAME_1M',
  '1Y': 'TIMEFRAME_1Y',
  'ALL': 'TIMEFRAME_ALL',
};

const TIMEFRAME_LABELS: Record<TimeframeKey, string> = {
  '1D': 'Past Day',
  '1W': 'Past Week',
  '1M': 'Past Month',
  '1Y': 'Past Year',
  'ALL': 'All Time',
};

interface ValuationPoint {
  date: string;
  totalValue: { amount: string; currencyCode: string };
  marketValue: { amount: string; currencyCode: string };
  cashValue: { amount: string; currencyCode: string };
  twrIndex: string;
  dailyReturn?: string | null;
}

interface PortfolioHistoryData {
  points: ValuationPoint[];
  startValue: { amount: string; currencyCode: string };
  endValue: { amount: string; currencyCode: string };
  returnAmount: { amount: string; currencyCode: string };
  returnPercent: string;
}

interface BenchmarkComparisonPoint {
  date: string;
  portfolioReturnPercent: string;
  benchmarkReturnPercent: string;
  activeReturnPercent: string;
  portfolioValue: { amount: string; currencyCode: string };
  benchmarkPriceAdjusted: { amount: string; currencyCode: string };
}

interface BenchmarkComparisonPayload {
  portfolioId: string;
  baseCurrency: string;
  benchmark: BenchmarkInfo;
  currencyMode: 'PORTFOLIO_BASE' | 'BENCHMARK_NATIVE';
  startDate: string;
  endDate: string;
  elapsedDays: number;
  points: BenchmarkComparisonPoint[];
  metrics: BenchmarkMetricsData;
}

const formatCurrency = (val: number, currency = 'USD') => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(val);
};

const formatPercent = (val: number, decimals = 2) => {
  const sign = val >= 0 ? '+' : '';
  return `${sign}${val.toFixed(decimals)}%`;
};

const formatDateLabel = (dateStr: string) => {
  if (!dateStr) return '';
  const parts = dateStr.split('-');
  if (parts.length < 3) return dateStr;
  const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
};

const formatShortDate = (dateStr: string) => {
  if (!dateStr) return '';
  const parts = dateStr.split('-');
  if (parts.length < 3) return dateStr;
  const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
};

export const PerformanceChart: React.FC = () => {
  const [timeframe, setTimeframe] = useState<TimeframeKey>('1Y');
  const [historyData, setHistoryData] = useState<PortfolioHistoryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [containerWidth, setContainerWidth] = useState(800);

  // Benchmark comparison state
  const [benchmarks, setBenchmarks] = useState<BenchmarkInfo[]>([]);
  const [selectedBenchmarkId, setSelectedBenchmarkId] = useState<string>(() => {
    return localStorage.getItem('graphfolio_benchmark_id') || 'SP500TR';
  });
  const [currencyMode, setCurrencyMode] = useState<'PORTFOLIO_BASE' | 'BENCHMARK_NATIVE'>(() => {
    return (localStorage.getItem('graphfolio_benchmark_currency_mode') as 'PORTFOLIO_BASE' | 'BENCHMARK_NATIVE') || 'PORTFOLIO_BASE';
  });
  const [benchmarkData, setBenchmarkData] = useState<BenchmarkComparisonPayload | null>(null);
  const [benchmarkLoading, setBenchmarkLoading] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);

  // ResizeObserver to dynamically adapt to container dimensions
  useEffect(() => {
    if (!containerRef.current) return;
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (entry.contentRect.width > 0) {
          setContainerWidth(entry.contentRect.width);
        }
      }
    });
    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, []);

  // Fetch benchmark catalog on mount
  useEffect(() => {
    client.query({
      benchmarks: {
        id: true,
        symbol: true,
        name: true,
        currencyCode: true,
        exchangeCode: true,
        indexType: true,
        isDefault: true,
      },
    })
      .then((res) => {
        const list = (res.benchmarks as BenchmarkInfo[]) || [];
        setBenchmarks(list);
        // If stored benchmark is invalid and list is present, pick default
        if (list.length > 0 && selectedBenchmarkId !== 'NONE') {
          const match = list.find((b) => b.id === selectedBenchmarkId);
          if (!match) {
            const def = list.find((b) => b.id === 'SP500TR') || list[0];
            setSelectedBenchmarkId(def.id);
          }
        }
      })
      .catch((err) => {
        console.error('Error fetching benchmarks catalog:', err);
      });
  }, []);

  // Persist benchmark selection
  const handleBenchmarkChange = (id: string) => {
    setSelectedBenchmarkId(id);
    localStorage.setItem('graphfolio_benchmark_id', id);
  };

  const handleCurrencyModeChange = (mode: 'PORTFOLIO_BASE' | 'BENCHMARK_NATIVE') => {
    setCurrencyMode(mode);
    localStorage.setItem('graphfolio_benchmark_currency_mode', mode);
  };

  // Fetch standard portfolio history
  useEffect(() => {
    let isCancelled = false;
    setLoading(true);

    client.query({
      portfolioHistory: {
        __args: { timeframe: TIMEFRAME_MAP[timeframe] },
        points: {
          date: true,
          totalValue: { amount: true, currencyCode: true },
          marketValue: { amount: true, currencyCode: true },
          cashValue: { amount: true, currencyCode: true },
          twrIndex: true,
          dailyReturn: true,
        },
        startValue: { amount: true, currencyCode: true },
        endValue: { amount: true, currencyCode: true },
        returnAmount: { amount: true, currencyCode: true },
        returnPercent: true,
      },
    })
      .then((res) => {
        if (!isCancelled) {
          setHistoryData(res.portfolioHistory as PortfolioHistoryData);
          setLoading(false);
          setHoveredIndex(null);
        }
      })
      .catch((err) => {
        if (!isCancelled) {
          console.error('Error fetching portfolio history:', err);
          setLoading(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [timeframe]);

  // Fetch benchmark comparison when a benchmark is selected
  useEffect(() => {
    if (selectedBenchmarkId === 'NONE') {
      setBenchmarkData(null);
      return;
    }

    let isCancelled = false;
    setBenchmarkLoading(true);

    client.query({
      benchmarkComparison: {
        __args: {
          benchmarkId: selectedBenchmarkId,
          timeframe: TIMEFRAME_MAP[timeframe],
          currencyMode,
        },
        portfolioId: true,
        baseCurrency: true,
        startDate: true,
        endDate: true,
        elapsedDays: true,
        currencyMode: true,
        benchmark: {
          id: true,
          symbol: true,
          name: true,
          exchangeCode: true,
          currencyCode: true,
          indexType: true,
          isDefault: true,
        },
        points: {
          date: true,
          portfolioReturnPercent: true,
          benchmarkReturnPercent: true,
          activeReturnPercent: true,
          portfolioValue: { amount: true, currencyCode: true },
          benchmarkPriceAdjusted: { amount: true, currencyCode: true },
        },
        metrics: {
          portfolioCumulativeReturnPercent: true,
          benchmarkCumulativeReturnPercent: true,
          activeCumulativeReturnPercent: true,
          portfolioAnnualizedReturnPercent: true,
          benchmarkAnnualizedReturnPercent: true,
          activeAnnualizedReturnPercent: true,
          isAnnualized: true,
          beta: true,
          trackingErrorPercent: true,
          informationRatio: true,
        },
      },
    })
      .then((res) => {
        if (!isCancelled) {
          setBenchmarkData(res.benchmarkComparison as BenchmarkComparisonPayload);
          setBenchmarkLoading(false);
        }
      })
      .catch((err) => {
        if (!isCancelled) {
          console.error('Error fetching benchmark comparison:', err);
          setBenchmarkLoading(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [selectedBenchmarkId, timeframe, currencyMode]);

  const isBenchmarkActive = selectedBenchmarkId !== 'NONE' && benchmarkData && benchmarkData.points.length > 0;
  const currentBenchmark = benchmarks.find((b) => b.id === selectedBenchmarkId) || benchmarkData?.benchmark;

  // Chart dimensions & coordinates
  const height = 320;
  const padding = { top: 20, bottom: 35, left: 10, right: 75 };
  const chartW = Math.max(containerWidth, 300);

  // Handle data mapping based on benchmark mode vs standard mode
  const points = isBenchmarkActive ? benchmarkData.points : (historyData?.points || []);
  const baseCurrency = isBenchmarkActive
    ? benchmarkData.baseCurrency
    : (historyData?.endValue.currencyCode || 'USD');

  // Compute Scales
  let minVal = 0;
  let maxVal = 100;
  let zeroY: number | null = null;

  if (isBenchmarkActive && benchmarkData) {
    const allReturns = [
      ...benchmarkData.points.map((p) => Number(p.portfolioReturnPercent)),
      ...benchmarkData.points.map((p) => Number(p.benchmarkReturnPercent)),
      0, // ensure 0 baseline is included
    ];
    const rawMin = Math.min(...allReturns);
    const rawMax = Math.max(...allReturns);
    const marginRange = (rawMax - rawMin) * 0.12 || 2;
    minVal = rawMin - marginRange;
    maxVal = rawMax + marginRange;

    zeroY =
      maxVal === minVal
        ? height / 2
        : padding.top + (1 - (0 - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);
  } else {
    const values = (historyData?.points || []).map((p) => Number(p.totalValue.amount));
    const rawMin = values.length > 0 ? Math.min(...values) : 0;
    const rawMax = values.length > 0 ? Math.max(...values) : 100;
    const marginRange = (rawMax - rawMin) * 0.08 || 10;
    minVal = rawMin - marginRange;
    maxVal = rawMax + marginRange;
  }

  // Generate coordinates for portfolio line
  const portCoords = points.map((p, i) => {
    const x =
      points.length === 1
        ? padding.left
        : padding.left + (i / (points.length - 1)) * (chartW - padding.left - padding.right);

    const val = isBenchmarkActive
      ? Number((p as BenchmarkComparisonPoint).portfolioReturnPercent)
      : Number((p as ValuationPoint).totalValue.amount);

    const y =
      maxVal === minVal
        ? height / 2
        : padding.top + (1 - (val - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);

    return { x, y, point: p };
  });

  // Generate coordinates for benchmark line (if active)
  const benchCoords = isBenchmarkActive && benchmarkData
    ? benchmarkData.points.map((p, i) => {
        const x =
          points.length === 1
            ? padding.left
            : padding.left + (i / (points.length - 1)) * (chartW - padding.left - padding.right);

        const val = Number(p.benchmarkReturnPercent);
        const y =
          maxVal === minVal
            ? height / 2
            : padding.top + (1 - (val - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);

        return { x, y, point: p };
      })
    : [];

  // Smooth Catmull-Rom / Bezier curve generator
  const getSmoothPath = (pts: Array<{ x: number; y: number }>) => {
    if (pts.length === 0) return '';
    if (pts.length === 1) {
      return `M ${padding.left} ${pts[0].y.toFixed(1)} L ${(chartW - padding.right).toFixed(1)} ${pts[0].y.toFixed(1)}`;
    }
    let d = `M ${pts[0].x.toFixed(1)} ${pts[0].y.toFixed(1)}`;
    for (let i = 0; i < pts.length - 1; i++) {
      const p0 = pts[Math.max(i - 1, 0)];
      const p1 = pts[i];
      const p2 = pts[i + 1];
      const p3 = pts[Math.min(i + 2, pts.length - 1)];

      const cp1x = p1.x + (p2.x - p0.x) / 6;
      const cp1y = p1.y + (p2.y - p0.y) / 6;
      const cp2x = p2.x - (p3.x - p1.x) / 6;
      const cp2y = p2.y - (p3.y - p1.y) / 6;

      d += ` C ${cp1x.toFixed(1)} ${cp1y.toFixed(1)}, ${cp2x.toFixed(1)} ${cp2y.toFixed(1)}, ${p2.x.toFixed(1)} ${p2.y.toFixed(1)}`;
    }
    return d;
  };

  const portLinePath = getSmoothPath(portCoords);
  const benchLinePath = isBenchmarkActive ? getSmoothPath(benchCoords) : '';

  const areaBottom = height - padding.bottom;
  const portAreaPath =
    portCoords.length > 1
      ? `${portLinePath} L ${portCoords[portCoords.length - 1].x.toFixed(1)} ${areaBottom} L ${portCoords[0].x.toFixed(1)} ${areaBottom} Z`
      : '';

  // Grid levels
  const gridLevels = 4;
  const gridLines = Array.from({ length: gridLevels }).map((_, idx) => {
    const val = minVal + ((maxVal - minVal) * (idx + 1)) / (gridLevels + 1);
    const y = padding.top + (1 - (val - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);
    return { val, y };
  });

  // Date labels along X axis
  const dateStep = Math.max(1, Math.floor(portCoords.length / 5));
  const dateLabels = portCoords.filter((_, idx) => idx % dateStep === 0 || idx === portCoords.length - 1);

  // Mouse / Pointer tracker
  const handlePointerMove = (e: React.PointerEvent<SVGSVGElement>) => {
    if (!svgRef.current || portCoords.length === 0) return;
    const rect = svgRef.current.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;

    let closestIdx = 0;
    let closestDist = Infinity;
    for (let i = 0; i < portCoords.length; i++) {
      const dist = Math.abs(portCoords[i].x - mouseX);
      if (dist < closestDist) {
        closestDist = dist;
        closestIdx = i;
      }
    }
    setHoveredIndex(closestIdx);
  };

  const handlePointerLeave = () => {
    setHoveredIndex(null);
  };

  // Header metric highlights
  const activePortCoord = hoveredIndex !== null && portCoords[hoveredIndex] ? portCoords[hoveredIndex] : null;
  const activeBenchCoord = isBenchmarkActive && hoveredIndex !== null && benchCoords[hoveredIndex] ? benchCoords[hoveredIndex] : null;

  const overallReturnAmount = Number(historyData?.returnAmount.amount || '0');
  const overallReturnPercent = Number(historyData?.returnPercent || '0');
  const isOverallPositive = overallReturnAmount >= 0;

  const startVal = Number(historyData?.startValue.amount || '0');
  const activeVal = activePortCoord && !isBenchmarkActive
    ? Number((activePortCoord.point as ValuationPoint).totalValue.amount)
    : (isBenchmarkActive && activePortCoord
        ? Number((activePortCoord.point as BenchmarkComparisonPoint).portfolioValue.amount)
        : Number(historyData?.endValue.amount || '0'));

  const activeDelta = activePortCoord && !isBenchmarkActive
    ? activeVal - startVal
    : overallReturnAmount;

  const activePercent = isBenchmarkActive && activePortCoord
    ? Number((activePortCoord.point as BenchmarkComparisonPoint).portfolioReturnPercent)
    : (startVal > 0 ? (activeDelta / startVal) * 100 : overallReturnPercent);

  const isActivePositive = activeDelta >= 0;

  const strokeColor = isActivePositive ? '#10b981' : '#ef4444';
  const gradientColor = isOverallPositive ? '#10b981' : '#ef4444';

  return (
    <section className="performance-chart-card" ref={containerRef}>
      <header className="performance-chart-header">
        <div className="performance-chart-title-area">
          <div className="performance-chart-title-row">
            <h3>Performance</h3>
            {/* Benchmark Selector Dropdown */}
            <div className="benchmark-select-wrapper">
              <label htmlFor="benchmark-dropdown" className="benchmark-label">
                Compare:
              </label>
              <select
                id="benchmark-dropdown"
                className="benchmark-dropdown"
                value={selectedBenchmarkId}
                onChange={(e) => handleBenchmarkChange(e.target.value)}
                aria-label="Select market benchmark to compare"
              >
                <option value="NONE">None (Portfolio Only)</option>
                {benchmarks.map((b) => (
                  <option key={b.id} value={b.id}>
                    {b.name} ({b.symbol})
                  </option>
                ))}
              </select>
            </div>

            {/* Currency Mode Toggle (visible when benchmark is active) */}
            {isBenchmarkActive && currentBenchmark && currentBenchmark.currencyCode !== baseCurrency && (
              <div className="benchmark-currency-toggle" role="group" aria-label="Currency mode">
                <button
                  type="button"
                  className={`currency-toggle-btn ${currencyMode === 'PORTFOLIO_BASE' ? 'active' : ''}`}
                  onClick={() => handleCurrencyModeChange('PORTFOLIO_BASE')}
                  title={`Translate ${currentBenchmark.symbol} into portfolio base currency (${baseCurrency})`}
                >
                  In {baseCurrency} (Base)
                </button>
                <button
                  type="button"
                  className={`currency-toggle-btn ${currencyMode === 'BENCHMARK_NATIVE' ? 'active' : ''}`}
                  onClick={() => handleCurrencyModeChange('BENCHMARK_NATIVE')}
                  title={`Express ${currentBenchmark.symbol} in native currency (${currentBenchmark.currencyCode})`}
                >
                  In {currentBenchmark.currencyCode} (Native)
                </button>
              </div>
            )}
          </div>

          <div className="performance-chart-highlight">
            <span className="performance-highlight-value">{formatCurrency(activeVal, baseCurrency)}</span>
            <span className={`performance-highlight-trend ${isActivePositive ? 'positive' : 'negative'}`}>
              {isActivePositive ? '+' : ''}
              {formatPercent(activePercent, 2)}
            </span>
          </div>

          <span className="performance-highlight-date">
            {activePortCoord
              ? formatDateLabel(
                  isBenchmarkActive
                    ? (activePortCoord.point as BenchmarkComparisonPoint).date
                    : (activePortCoord.point as ValuationPoint).date
                )
              : TIMEFRAME_LABELS[timeframe]}
          </span>
        </div>

        {/* Timeframe Selector Bar */}
        <nav className="timeframe-bar" aria-label="Timeframe selection">
          {(['1D', '1W', '1M', '1Y', 'ALL'] as TimeframeKey[]).map((tf) => (
            <button
              key={tf}
              type="button"
              className={`timeframe-btn ${timeframe === tf ? 'active' : ''}`}
              onClick={() => setTimeframe(tf)}
              aria-pressed={timeframe === tf}
            >
              {tf}
            </button>
          ))}
        </nav>
      </header>

      {/* Benchmark Alpha & Risk Attribution Metrics Strip */}
      {isBenchmarkActive && benchmarkData && (
        <BenchmarkMetricsStrip
          benchmark={benchmarkData.benchmark}
          metrics={benchmarkData.metrics}
          currencyMode={currencyMode}
          baseCurrency={baseCurrency}
          elapsedDays={benchmarkData.elapsedDays}
        />
      )}

      {/* Chart Canvas */}
      <div className="chart-svg-container">
        {(loading && !historyData) || (benchmarkLoading && !benchmarkData) ? (
          <div className="chart-loading-state">
            <div className="chart-loading-spinner" />
            <span>Loading historical performance data...</span>
          </div>
        ) : portCoords.length === 0 ? (
          <div className="chart-empty-state">
            <span>No valuation history recorded for this timeframe.</span>
          </div>
        ) : (
          <>
            <svg
              ref={svgRef}
              className="chart-svg"
              viewBox={`0 0 ${chartW} ${height}`}
              onPointerMove={handlePointerMove}
              onPointerLeave={handlePointerLeave}
            >
              <defs>
                <linearGradient id="perfChartGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={gradientColor} stopOpacity="0.25" />
                  <stop offset="60%" stopColor={gradientColor} stopOpacity="0.05" />
                  <stop offset="100%" stopColor={gradientColor} stopOpacity="0.0" />
                </linearGradient>

                <filter id="glowEffect" x="-20%" y="-20%" width="140%" height="140%">
                  <feGaussianBlur stdDeviation="3" result="blur" />
                  <feComposite in="SourceGraphic" in2="blur" operator="over" />
                </filter>
              </defs>

              {/* Horizontal Price Grid Lines */}
              {gridLines.map((gl, i) => (
                <g key={`grid-${i}`}>
                  <line
                    x1={padding.left}
                    y1={gl.y}
                    x2={chartW - padding.right}
                    y2={gl.y}
                    className="chart-grid-line"
                  />
                  <text
                    x={chartW - padding.right + 8}
                    y={gl.y + 3}
                    className="chart-grid-label"
                  >
                    {isBenchmarkActive ? formatPercent(gl.val, 1) : formatCurrency(gl.val, baseCurrency)}
                  </text>
                </g>
              ))}

              {/* Zero baseline when benchmark comparison is active */}
              {zeroY !== null && zeroY >= padding.top && zeroY <= areaBottom && (
                <g className="chart-zero-baseline-group">
                  <line
                    x1={padding.left}
                    y1={zeroY}
                    x2={chartW - padding.right}
                    y2={zeroY}
                    className="chart-zero-baseline"
                  />
                  <text
                    x={chartW - padding.right + 8}
                    y={zeroY + 3}
                    className="chart-grid-label chart-zero-label"
                  >
                    0.00%
                  </text>
                </g>
              )}

              {/* Area fill for Portfolio */}
              {portAreaPath && (
                <path
                  d={portAreaPath}
                  fill="url(#perfChartGradient)"
                  style={{ transition: 'd 0.3s ease-out' }}
                />
              )}

              {/* Benchmark dashed curve (if active) */}
              {isBenchmarkActive && benchLinePath && (
                <path
                  d={benchLinePath}
                  fill="none"
                  stroke="#f59e0b"
                  strokeWidth="2.25"
                  strokeDasharray="5 5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  style={{ transition: 'all 0.25s ease' }}
                />
              )}

              {/* Portfolio main curve stroke */}
              {portLinePath && (
                <path
                  d={portLinePath}
                  fill="none"
                  stroke={strokeColor}
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  filter="url(#glowEffect)"
                  style={{ transition: 'stroke 0.25s ease' }}
                />
              )}

              {/* Date ticks on bottom axis */}
              {dateLabels.map((lbl, idx) => (
                <text
                  key={`date-${idx}`}
                  x={lbl.x}
                  y={height - 10}
                  textAnchor={idx === 0 ? 'start' : idx === dateLabels.length - 1 ? 'end' : 'middle'}
                  className="chart-grid-label"
                >
                  {formatShortDate(
                    isBenchmarkActive
                      ? (lbl.point as BenchmarkComparisonPoint).date
                      : (lbl.point as ValuationPoint).date
                  )}
                </text>
              ))}

              {/* Crosshair indicator on hover */}
              {activePortCoord && (
                <g>
                  {/* Vertical Tracker Line */}
                  <line
                    x1={activePortCoord.x}
                    y1={padding.top}
                    x2={activePortCoord.x}
                    y2={areaBottom}
                    className="chart-crosshair-line"
                  />

                  {/* Benchmark Point on Tracker Line */}
                  {isBenchmarkActive && activeBenchCoord && (
                    <circle
                      cx={activeBenchCoord.x}
                      cy={activeBenchCoord.y}
                      r="4.5"
                      fill="#fff"
                      stroke="#f59e0b"
                      strokeWidth="2.5"
                    />
                  )}

                  {/* Outer pulse ring for portfolio */}
                  <circle
                    cx={activePortCoord.x}
                    cy={activePortCoord.y}
                    r="8"
                    className="chart-focus-outer-ring"
                  />
                  {/* Inner glowing dot for portfolio */}
                  <circle
                    cx={activePortCoord.x}
                    cy={activePortCoord.y}
                    r="4.5"
                    fill="#fff"
                    stroke={strokeColor}
                    strokeWidth="2.5"
                  />
                </g>
              )}
            </svg>

            {/* Hover Tooltip Card */}
            {activePortCoord && (
              <div
                className="chart-tooltip-floating"
                style={{
                  left: Math.min(
                    Math.max(activePortCoord.x - 100, 10),
                    chartW - 250
                  ),
                }}
              >
                <span className="chart-tooltip-date">
                  {formatDateLabel(
                    isBenchmarkActive
                      ? (activePortCoord.point as BenchmarkComparisonPoint).date
                      : (activePortCoord.point as ValuationPoint).date
                  )}
                </span>

                {isBenchmarkActive ? (
                  <>
                    {/* Synchronized Dual Comparison Tooltip */}
                    <div className="chart-tooltip-row">
                      <span className="font-semibold text-white">Portfolio:</span>
                      <span className={`val font-bold ${Number((activePortCoord.point as BenchmarkComparisonPoint).portfolioReturnPercent) >= 0 ? 'text-positive' : 'text-negative'}`}>
                        {formatPercent(Number((activePortCoord.point as BenchmarkComparisonPoint).portfolioReturnPercent), 2)}
                      </span>
                    </div>

                    <div className="chart-tooltip-row">
                      <span className="font-semibold" style={{ color: '#f59e0b' }}>
                        {currentBenchmark?.symbol || 'Benchmark'}:
                      </span>
                      <span className={`val font-bold ${Number((activePortCoord.point as BenchmarkComparisonPoint).benchmarkReturnPercent) >= 0 ? 'text-positive' : 'text-negative'}`}>
                        {formatPercent(Number((activePortCoord.point as BenchmarkComparisonPoint).benchmarkReturnPercent), 2)}
                      </span>
                    </div>

                    <div className="chart-tooltip-divider" />

                    <div className="chart-tooltip-row">
                      <span>Active Spread:</span>
                      <span className={`val font-bold ${Number((activePortCoord.point as BenchmarkComparisonPoint).activeReturnPercent) >= 0 ? 'text-positive' : 'text-negative'}`}>
                        {formatPercent(Number((activePortCoord.point as BenchmarkComparisonPoint).activeReturnPercent), 2)}
                      </span>
                    </div>

                    <div className="chart-tooltip-row">
                      <span>Portfolio Value:</span>
                      <span className="val">
                        {formatCurrency(Number((activePortCoord.point as BenchmarkComparisonPoint).portfolioValue.amount), baseCurrency)}
                      </span>
                    </div>
                  </>
                ) : (
                  <>
                    <span className="chart-tooltip-total">{formatCurrency(activeVal, baseCurrency)}</span>
                    <span className={`chart-tooltip-delta ${isActivePositive ? 'positive' : 'negative'}`}>
                      {isActivePositive ? '+' : ''}{formatCurrency(activeDelta, baseCurrency)} ({formatPercent(activePercent, 2)})
                    </span>
                    <div className="chart-tooltip-divider" />
                    <div className="chart-tooltip-row">
                      <span>Market:</span>
                      <span className="val">{formatCurrency(Number((activePortCoord.point as ValuationPoint).marketValue.amount), baseCurrency)}</span>
                    </div>
                    <div className="chart-tooltip-row">
                      <span>Cash:</span>
                      <span className="val">{formatCurrency(Number((activePortCoord.point as ValuationPoint).cashValue.amount), baseCurrency)}</span>
                    </div>
                  </>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {/* Chart Legend (visible when benchmark is active) */}
      {isBenchmarkActive && (
        <footer className="chart-legend-strip">
          <div className="chart-legend-item">
            <span className="chart-legend-line solid-green" />
            <span className="chart-legend-text">Your Portfolio</span>
          </div>
          <div className="chart-legend-item">
            <span className="chart-legend-line dashed-amber" />
            <span className="chart-legend-text">
              {currentBenchmark?.name || currentBenchmark?.symbol} ({currencyMode === 'PORTFOLIO_BASE' ? `In ${baseCurrency}` : `In ${currentBenchmark?.currencyCode || 'USD'}`})
            </span>
          </div>
        </footer>
      )}
    </section>
  );
};
