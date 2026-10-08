import React, { useEffect, useRef, useState } from 'react';
import { client } from '@graphfolio/api-client';
import { Badge } from '@graphfolio/ui';
import './FXTrendChart.css';

export type FXTimeframeKey = '1W' | '1M' | '1Y' | 'ALL';

const TIMEFRAME_MAP: Record<FXTimeframeKey, 'TIMEFRAME_1W' | 'TIMEFRAME_1M' | 'TIMEFRAME_1Y' | 'TIMEFRAME_ALL'> = {
  '1W': 'TIMEFRAME_1W',
  '1M': 'TIMEFRAME_1M',
  '1Y': 'TIMEFRAME_1Y',
  'ALL': 'TIMEFRAME_ALL',
};

interface FXHistoryPointItem {
  date: string;
  rate: string;
  invertedRate: string;
  source: string;
}

interface FXHistoryData {
  baseCurrency: string;
  quoteCurrency: string;
  pair: string;
  startRate: string;
  endRate: string;
  periodChange: string;
  periodChangePct: string;
  periodHigh: string;
  periodLow: string;
  points: FXHistoryPointItem[];
}

interface FXTrendChartProps {
  baseCurrency: string;
  quoteCurrency: string;
  isInverted?: boolean;
  timeframe: FXTimeframeKey;
  onTimeframeChange: (tf: FXTimeframeKey) => void;
}

const formatDateLabel = (dateStr: string) => {
  if (!dateStr) return '';
  const parts = dateStr.split('T')[0].split('-');
  if (parts.length < 3) return dateStr;
  const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
};

const formatShortDate = (dateStr: string) => {
  if (!dateStr) return '';
  const parts = dateStr.split('T')[0].split('-');
  if (parts.length < 3) return dateStr;
  const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
};

export const FXTrendChart: React.FC<FXTrendChartProps> = ({
  baseCurrency,
  quoteCurrency,
  isInverted = false,
  timeframe,
  onTimeframeChange,
}) => {
  const [historyData, setHistoryData] = useState<FXHistoryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [containerWidth, setContainerWidth] = useState(700);

  const containerRef = useRef<HTMLDivElement>(null);
  const svgRef = useRef<SVGSVGElement>(null);

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

  useEffect(() => {
    let isCancelled = false;
    setLoading(true);

    client
      .query({
        currencyPairHistory: {
          __args: {
            baseCurrency,
            quoteCurrency,
            timeframe: TIMEFRAME_MAP[timeframe],
          },
          baseCurrency: true,
          quoteCurrency: true,
          pair: true,
          startRate: true,
          endRate: true,
          periodChange: true,
          periodChangePct: true,
          periodHigh: true,
          periodLow: true,
          points: {
            date: true,
            rate: true,
            invertedRate: true,
            source: true,
          },
        },
      })
      .then((res) => {
        if (!isCancelled && res.currencyPairHistory) {
          setHistoryData(res.currencyPairHistory as FXHistoryData);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!isCancelled) {
          console.warn('Could not fetch currency pair history:', err);
          setHistoryData(null);
          setLoading(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [baseCurrency, quoteCurrency, timeframe]);

  const points = historyData?.points || [];

  // Determine active pair display
  const activePairLabel = isInverted
    ? `${quoteCurrency}/${baseCurrency}`
    : `${baseCurrency}/${quoteCurrency}`;

  // Rates based on inversion toggle
  const getPointRate = (p: FXHistoryPointItem) => {
    return isInverted ? Number(p.invertedRate) : Number(p.rate);
  };

  const currentRateNum = points.length > 0
    ? getPointRate(points[points.length - 1])
    : isInverted && historyData
    ? (Number(historyData.endRate) > 0 ? 1 / Number(historyData.endRate) : 0)
    : Number(historyData?.endRate || 0);

  const startRateNum = points.length > 0
    ? getPointRate(points[0])
    : isInverted && historyData
    ? (Number(historyData.startRate) > 0 ? 1 / Number(historyData.startRate) : 0)
    : Number(historyData?.startRate || 0);

  const deltaNum = currentRateNum - startRateNum;
  const deltaPct = startRateNum > 0 ? (deltaNum / startRateNum) * 100 : 0;
  const isPositive = deltaNum >= 0;

  // Chart layout geometry
  const height = 280;
  const padding = { top: 20, bottom: 35, left: 10, right: 80 };
  const chartW = Math.max(containerWidth, 300);

  const rateValues = points.map(getPointRate);
  const rawMin = rateValues.length > 0 ? Math.min(...rateValues) : 0;
  const rawMax = rateValues.length > 0 ? Math.max(...rateValues) : 1;
  const marginRange = (rawMax - rawMin) * 0.1 || (rawMin * 0.05 || 0.01);
  const minVal = Math.max(0, rawMin - marginRange);
  const maxVal = rawMax + marginRange;

  const avgVal = rateValues.length > 0
    ? rateValues.reduce((sum, v) => sum + v, 0) / rateValues.length
    : 0;

  const avgY = maxVal === minVal
    ? height / 2
    : padding.top + (1 - (avgVal - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);

  const coords = points.map((p, i) => {
    const x =
      points.length <= 1
        ? padding.left
        : padding.left + (i / (points.length - 1)) * (chartW - padding.left - padding.right);
    const rateVal = getPointRate(p);
    const y =
      maxVal === minVal
        ? height / 2
        : padding.top + (1 - (rateVal - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);
    return { x, y, point: p, rateVal };
  });

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

  // 4 horizontal grid lines
  const gridLevels = 4;
  const gridLines = Array.from({ length: gridLevels }).map((_, idx) => {
    const val = minVal + ((maxVal - minVal) * (idx + 1)) / (gridLevels + 1);
    const y = padding.top + (1 - (val - minVal) / (maxVal - minVal)) * (height - padding.top - padding.bottom);
    return { val, y };
  });

  // Date labels along X axis
  const dateStep = Math.max(1, Math.floor(coords.length / 5));
  const dateLabels = coords.filter((_, idx) => idx % dateStep === 0 || idx === coords.length - 1);

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

  const activeCoord = hoveredIndex !== null && coords[hoveredIndex] ? coords[hoveredIndex] : null;

  return (
    <div className="fx-chart" ref={containerRef}>
      <div className="fx-chart__header">
        <div className="fx-chart__title-area">
          <div className="fx-chart__pair-title">
            <span>{activePairLabel}</span>
            {isInverted && <Badge variant="purple">Inverted Quote (1/X)</Badge>}
          </div>

          <div className="fx-chart__rate-display">
            <span className="fx-chart__current-rate">
              {activeCoord
                ? activeCoord.rateVal.toFixed(6)
                : currentRateNum.toFixed(6)}
            </span>
            <span
              className={`fx-chart__change-badge ${
                isPositive ? 'fx-chart__change-badge--positive' : 'fx-chart__change-badge--negative'
              }`}
            >
              {isPositive ? '+' : ''}
              {deltaNum.toFixed(6)} ({isPositive ? '+' : ''}
              {deltaPct.toFixed(2)}%)
            </span>
          </div>
        </div>

        <div className="fx-chart__timeframes">
          {(['1W', '1M', '1Y', 'ALL'] as FXTimeframeKey[]).map((tf) => (
            <button
              key={tf}
              type="button"
              className={`fx-chart__tf-btn ${timeframe === tf ? 'fx-chart__tf-btn--active' : ''}`}
              onClick={() => onTimeframeChange(tf)}
            >
              {tf}
            </button>
          ))}
        </div>
      </div>

      <div className="fx-chart__canvas-container">
        {loading ? (
          <div className="fx-chart__loading">Loading historical rates...</div>
        ) : points.length === 0 ? (
          <div className="fx-chart__empty">
            <span>No historical rate observations found for this timeframe.</span>
          </div>
        ) : (
          <>
            <svg
              ref={svgRef}
              className="fx-chart__svg"
              width={chartW}
              height={height}
              onPointerMove={handlePointerMove}
              onPointerLeave={handlePointerLeave}
            >
              <defs>
                <linearGradient id="fxAreaGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#06b6d4" stopOpacity="0.35" />
                  <stop offset="100%" stopColor="#06b6d4" stopOpacity="0.0" />
                </linearGradient>
                <linearGradient id="fxLineGradient" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stopColor="#38bdf8" />
                  <stop offset="100%" stopColor="#06b6d4" />
                </linearGradient>
              </defs>

              {/* Grid Lines & Labels */}
              {gridLines.map((g, idx) => (
                <g key={idx}>
                  <line
                    x1={padding.left}
                    y1={g.y}
                    x2={chartW - padding.right}
                    y2={g.y}
                    className="fx-chart__grid-line"
                  />
                  <text
                    x={chartW - padding.right + 8}
                    y={g.y + 4}
                    className="fx-chart__grid-text"
                  >
                    {g.val.toFixed(4)}
                  </text>
                </g>
              ))}

              {/* Period Average Line */}
              {rateValues.length > 1 && (
                <line
                  x1={padding.left}
                  y1={avgY}
                  x2={chartW - padding.right}
                  y2={avgY}
                  className="fx-chart__avg-line"
                />
              )}

              {/* Area & Stroke Path */}
              {areaPath && <path d={areaPath} fill="url(#fxAreaGradient)" />}
              {linePath && (
                <path
                  d={linePath}
                  fill="none"
                  stroke="url(#fxLineGradient)"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              )}

              {/* Date Axis Labels */}
              {dateLabels.map((coord, idx) => (
                <text
                  key={idx}
                  x={coord.x}
                  y={height - 12}
                  textAnchor={idx === 0 ? 'start' : idx === dateLabels.length - 1 ? 'end' : 'middle'}
                  className="fx-chart__axis-text"
                >
                  {formatShortDate(coord.point.date)}
                </text>
              ))}

              {/* Active Hover Crosshairs & Dot */}
              {activeCoord && (
                <g>
                  <line
                    x1={activeCoord.x}
                    y1={padding.top}
                    x2={activeCoord.x}
                    y2={areaBottom}
                    className="fx-chart__crosshair-line"
                  />
                  <circle
                    cx={activeCoord.x}
                    cy={activeCoord.y}
                    r="5"
                    fill="#38bdf8"
                    stroke="#ffffff"
                    strokeWidth="2"
                  />
                </g>
              )}
            </svg>

            {/* Interactive Tooltip Card */}
            {activeCoord && (
              <div
                className="fx-chart__tooltip"
                style={{
                  left: `${activeCoord.x}px`,
                  top: `${activeCoord.y}px`,
                }}
              >
                <div className="fx-chart__tooltip-date">{formatDateLabel(activeCoord.point.date)}</div>
                <div className="fx-chart__tooltip-rate">
                  {activePairLabel}: {activeCoord.rateVal.toFixed(6)}
                </div>
                <div className="fx-chart__tooltip-reciprocal">
                  {isInverted
                    ? `${baseCurrency}/${quoteCurrency}: ${Number(activeCoord.point.rate).toFixed(6)}`
                    : `${quoteCurrency}/${baseCurrency}: ${Number(activeCoord.point.invertedRate).toFixed(6)}`}
                </div>
                <div className="fx-chart__tooltip-source">Source: {activeCoord.point.source}</div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};
