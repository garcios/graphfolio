import React, { useEffect, useRef, useState } from 'react';
import { client } from '../graphql/client';
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

  // Fetch portfolio history whenever timeframe changes
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

  const points = historyData?.points || [];
  const currency = historyData?.endValue.currencyCode || 'USD';
  const overallReturnAmount = Number(historyData?.returnAmount.amount || '0');
  const overallReturnPercent = Number(historyData?.returnPercent || '0');
  const isOverallPositive = overallReturnAmount >= 0;

  // Chart dimensions & coordinates
  const height = 320;
  const padding = { top: 20, bottom: 35, left: 10, right: 70 };
  const chartW = Math.max(containerWidth, 300);

  const values = points.map((p) => Number(p.totalValue.amount));
  const rawMin = values.length > 0 ? Math.min(...values) : 0;
  const rawMax = values.length > 0 ? Math.max(...values) : 100;
  const marginRange = (rawMax - rawMin) * 0.08 || 10;
  const minVal = rawMin - marginRange;
  const maxVal = rawMax + marginRange;

  const coords = points.map((p, i) => {
    const x =
      points.length === 1
        ? padding.left
        : padding.left + (i / (points.length - 1)) * (chartW - padding.left - padding.right);
    const y =
      maxVal === minVal
        ? height / 2
        : padding.top + (1 - (Number(p.totalValue.amount) - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);
    return { x, y, point: p };
  });

  // Smooth Catmull-Rom / Bezier curve
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

  const linePath = getSmoothPath(coords);
  const areaBottom = height - padding.bottom;
  const areaPath =
    coords.length > 1
      ? `${linePath} L ${coords[coords.length - 1].x.toFixed(1)} ${areaBottom} L ${coords[0].x.toFixed(1)} ${areaBottom} Z`
      : coords.length === 1
      ? `M ${padding.left} ${coords[0].y.toFixed(1)} L ${(chartW - padding.right).toFixed(1)} ${coords[0].y.toFixed(1)} L ${(chartW - padding.right).toFixed(1)} ${areaBottom} L ${padding.left} ${areaBottom} Z`
      : '';

  // Horizontal price grid lines
  const gridLevels = 4;
  const gridLines = Array.from({ length: gridLevels }).map((_, idx) => {
    const val = minVal + ((maxVal - minVal) * (idx + 1)) / (gridLevels + 1);
    const y = padding.top + (1 - (val - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);
    return { val, y };
  });

  // Date labels along X axis
  const dateStep = Math.max(1, Math.floor(coords.length / 5));
  const dateLabels = coords.filter((_, idx) => idx % dateStep === 0 || idx === coords.length - 1);

  // Mouse / Pointer tracker
  const handlePointerMove = (e: React.PointerEvent<SVGSVGElement>) => {
    if (!svgRef.current || coords.length === 0) return;
    const rect = svgRef.current.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;

    let closestIdx = 0;
    let closestDist = Infinity;
    for (let i = 0; i < coords.length; i++) {
      const dist = Math.abs(coords[i].x - mouseX);
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

  // Hovered or latest metrics for header display
  const activeCoord = hoveredIndex !== null && coords[hoveredIndex] ? coords[hoveredIndex] : null;
  const startVal = Number(historyData?.startValue.amount || '0');
  const activeVal = activeCoord ? Number(activeCoord.point.totalValue.amount) : Number(historyData?.endValue.amount || '0');
  const activeDelta = activeCoord ? activeVal - startVal : overallReturnAmount;
  const activePercent = startVal > 0 ? (activeDelta / startVal) * 100 : overallReturnPercent;
  const isActivePositive = activeDelta >= 0;

  const strokeColor = isActivePositive ? '#10b981' : '#ef4444';
  const gradientColor = isOverallPositive ? '#10b981' : '#ef4444';

  return (
    <section className="performance-chart-card" ref={containerRef}>
      <header className="performance-chart-header">
        <div className="performance-chart-title-area">
          <h3>Performance</h3>
          <div className="performance-chart-highlight">
            <span className="performance-highlight-value">{formatCurrency(activeVal, currency)}</span>
            <span className={`performance-highlight-trend ${isActivePositive ? 'positive' : 'negative'}`}>
              {isActivePositive ? '+' : ''}{formatCurrency(activeDelta, currency)} ({formatPercent(activePercent, 2)})
            </span>
          </div>
          <span className="performance-highlight-date">
            {activeCoord ? formatDateLabel(activeCoord.point.date) : TIMEFRAME_LABELS[timeframe]}
          </span>
        </div>

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

      <div className="chart-svg-container">
        {loading && !historyData ? (
          <div className="chart-loading-state">
            <div className="chart-loading-spinner" />
            <span>Loading historical valuations...</span>
          </div>
        ) : coords.length === 0 ? (
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
                  <stop offset="0%" stopColor={gradientColor} stopOpacity="0.32" />
                  <stop offset="60%" stopColor={gradientColor} stopOpacity="0.08" />
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
                    {formatCurrency(gl.val, currency)}
                  </text>
                </g>
              ))}

              {/* Area fill */}
              {areaPath && (
                <path
                  d={areaPath}
                  fill="url(#perfChartGradient)"
                  style={{ transition: 'd 0.3s ease-out' }}
                />
              )}

              {/* Main curve stroke */}
              {linePath && (
                <path
                  d={linePath}
                  fill="none"
                  stroke={strokeColor}
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  filter="url(#glowEffect)"
                  style={{ transition: 'stroke 0.25s ease' }}
                />
              )}

              {/* Single point fallback circle */}
              {coords.length === 1 && (
                <circle
                  cx={(padding.left + chartW - padding.right) / 2}
                  cy={coords[0].y}
                  r="5"
                  fill="#fff"
                  stroke={strokeColor}
                  strokeWidth="2.5"
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
                  {formatShortDate(lbl.point.date)}
                </text>
              ))}

              {/* Crosshair indicator on hover */}
              {activeCoord && (
                <g>
                  {/* Vertical Tracker Line */}
                  <line
                    x1={activeCoord.x}
                    y1={padding.top}
                    x2={activeCoord.x}
                    y2={areaBottom}
                    className="chart-crosshair-line"
                  />
                  {/* Outer pulse ring */}
                  <circle
                    cx={activeCoord.x}
                    cy={activeCoord.y}
                    r="8"
                    className="chart-focus-outer-ring"
                  />
                  {/* Inner glowing dot */}
                  <circle
                    cx={activeCoord.x}
                    cy={activeCoord.y}
                    r="4.5"
                    fill="#fff"
                    stroke={strokeColor}
                    strokeWidth="2.5"
                  />
                </g>
              )}
            </svg>

            {/* Hover Tooltip Card */}
            {activeCoord && (
              <div
                className="chart-tooltip-floating"
                style={{
                  left: Math.min(
                    Math.max(activeCoord.x - 90, 10),
                    chartW - 220
                  ),
                }}
              >
                <span className="chart-tooltip-date">{formatDateLabel(activeCoord.point.date)}</span>
                <span className="chart-tooltip-total">{formatCurrency(activeVal, currency)}</span>
                <span className={`chart-tooltip-delta ${isActivePositive ? 'positive' : 'negative'}`}>
                  {isActivePositive ? '+' : ''}{formatCurrency(activeDelta, currency)} ({formatPercent(activePercent, 2)})
                </span>
                <div className="chart-tooltip-divider" />
                <div className="chart-tooltip-row">
                  <span>Market:</span>
                  <span className="val">{formatCurrency(Number(activeCoord.point.marketValue.amount), currency)}</span>
                </div>
                <div className="chart-tooltip-row">
                  <span>Cash:</span>
                  <span className="val">{formatCurrency(Number(activeCoord.point.cashValue.amount), currency)}</span>
                </div>
                {activeCoord.point.dailyReturn !== undefined && activeCoord.point.dailyReturn !== null && (
                  <div className="chart-tooltip-row">
                    <span>Daily Return:</span>
                    <span className="val">{formatPercent(Number(activeCoord.point.dailyReturn) * 100, 2)}</span>
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </section>
  );
};
