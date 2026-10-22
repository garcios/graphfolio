import React, { useState, useEffect, useMemo } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney, formatPercent, formatQuantity, formatDate } from '@graphfolio/ui';
import { DRIPSimulatorModal } from './DRIPSimulatorModal';
import './DividendCalendar.css';

interface MoneyValue {
  amount: string;
  currencyCode?: string | null;
}

interface DividendCalendarEvent {
  id: string;
  instrumentId: string;
  symbol: string;
  name: string;
  exDate: string;
  recordDate?: string | null;
  paymentDate: string;
  sharesHeld: string;
  dpsDeclared: MoneyValue;
  grossAmountBase: MoneyValue;
  netAmountBase: MoneyValue;
  frequency: string;
  status: 'ANNOUNCED' | 'EX_DATE_PASSED' | 'PAID' | 'CANCELLED';
  isUserEntitled: boolean;
  isEstimated: boolean;
  frankingPct: string;
}

interface MonthlyProjection {
  year: number;
  month: number;
  projectedIncome: MoneyValue;
  dividendCount: number;
}

interface HoldingYieldItem {
  instrumentId: string;
  symbol: string;
  name: string;
  quantity: string;
  currentPrice: MoneyValue;
  averageBuyPrice: MoneyValue;
  currentMarketValue: MoneyValue;
  costBasis: MoneyValue;
  annualDps: MoneyValue;
  projectedAnnualCash: MoneyValue;
  currentYieldPct: string;
  yieldOnCostPct?: string | null;
  dividendCagr3y?: string | null;
  frequency: string;
  nextExDate?: string | null;
  nextPaymentDate?: string | null;
}

interface DividendYieldAnalytics {
  annualForwardIncome: MoneyValue;
  monthlyAverageIncome: MoneyValue;
  portfolioCurrentYield: string;
  portfolioYieldOnCost?: string | null;
  pendingCashInflow: MoneyValue;
  monthlyProjections: MonthlyProjection[];
  holdingsYield: HoldingYieldItem[];
}

export type DividendViewMode = 'calendar' | 'agenda' | 'matrix';

export interface DividendCalendarProps {
  preferredCurrency?: string;
}

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
];

const WEEKDAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

export const DividendCalendar: React.FC<DividendCalendarProps> = ({ preferredCurrency = 'USD' }) => {
  const today = new Date();
  const [currentYear, setCurrentYear] = useState<number>(today.getFullYear());
  const [currentMonth, setCurrentMonth] = useState<number>(today.getMonth() + 1); // 1..12

  const [viewMode, setViewMode] = useState<DividendViewMode>('calendar');
  const [analytics, setAnalytics] = useState<DividendYieldAnalytics | null>(null);
  const [calendarEvents, setCalendarEvents] = useState<DividendCalendarEvent[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [selectedEvent, setSelectedEvent] = useState<DividendCalendarEvent | null>(null);
  const [isDRIPOpen, setIsDRIPOpen] = useState<boolean>(false);
  const [hoveredBar, setHoveredBar] = useState<MonthlyProjection | null>(null);

  // Fetch summary analytics & 12-month projections
  const fetchAnalytics = async () => {
    try {
      const res = await client.query({
        dividendYieldAnalytics: {
          annualForwardIncome: { amount: true, currencyCode: true },
          monthlyAverageIncome: { amount: true, currencyCode: true },
          portfolioCurrentYield: true,
          portfolioYieldOnCost: true,
          pendingCashInflow: { amount: true, currencyCode: true },
          monthlyProjections: {
            year: true,
            month: true,
            projectedIncome: { amount: true, currencyCode: true },
            dividendCount: true,
          },
          holdingsYield: {
            instrumentId: true,
            symbol: true,
            name: true,
            quantity: true,
            currentPrice: { amount: true, currencyCode: true },
            averageBuyPrice: { amount: true, currencyCode: true },
            currentMarketValue: { amount: true, currencyCode: true },
            costBasis: { amount: true, currencyCode: true },
            annualDps: { amount: true, currencyCode: true },
            projectedAnnualCash: { amount: true, currencyCode: true },
            currentYieldPct: true,
            yieldOnCostPct: true,
            dividendCagr3y: true,
            frequency: true,
            nextExDate: true,
            nextPaymentDate: true,
          },
        },
      });

      if (res.dividendYieldAnalytics) {
        setAnalytics(res.dividendYieldAnalytics as DividendYieldAnalytics);
      }
    } catch (err: any) {
      console.error('Failed to load dividend yield analytics:', err);
      setError(err.message || 'Failed to load dividend yield analytics');
    }
  };

  // Fetch monthly calendar events
  const fetchCalendar = async (year: number, month: number) => {
    setLoading(true);
    try {
      const res = await client.query({
        dividendCalendar: {
          __args: { year, month },
          id: true,
          instrumentId: true,
          symbol: true,
          name: true,
          exDate: true,
          recordDate: true,
          paymentDate: true,
          sharesHeld: true,
          dpsDeclared: { amount: true, currencyCode: true },
          grossAmountBase: { amount: true, currencyCode: true },
          netAmountBase: { amount: true, currencyCode: true },
          frequency: true,
          status: true,
          isUserEntitled: true,
          isEstimated: true,
          frankingPct: true,
        },
      });

      if (res.dividendCalendar) {
        setCalendarEvents(res.dividendCalendar as DividendCalendarEvent[]);
      }
    } catch (err: any) {
      console.error('Failed to load dividend calendar events:', err);
      setError(err.message || 'Failed to load dividend calendar events');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAnalytics();
  }, [preferredCurrency]);

  useEffect(() => {
    fetchCalendar(currentYear, currentMonth);
  }, [currentYear, currentMonth]);

  const handlePrevMonth = () => {
    if (currentMonth === 1) {
      setCurrentMonth(12);
      setCurrentYear(currentYear - 1);
    } else {
      setCurrentMonth(currentMonth - 1);
    }
  };

  const handleNextMonth = () => {
    if (currentMonth === 12) {
      setCurrentMonth(1);
      setCurrentYear(currentYear + 1);
    } else {
      setCurrentMonth(currentMonth + 1);
    }
  };

  const handleToday = () => {
    setCurrentYear(today.getFullYear());
    setCurrentMonth(today.getMonth() + 1);
  };

  // Build calendar matrix days
  const calendarGrid = useMemo(() => {
    const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();
    const firstDayIndex = (new Date(currentYear, currentMonth - 1, 1).getDay() + 6) % 7; // Monday = 0

    const prevMonthDays = new Date(currentYear, currentMonth - 1, 0).getDate();

    interface CalendarCell {
      dayNumber: number;
      dateString: string;
      isCurrentMonth: boolean;
      isToday: boolean;
      exEvents: DividendCalendarEvent[];
      payEvents: DividendCalendarEvent[];
    }

    const cells: CalendarCell[] = [];

    // Leading padding from prev month
    for (let i = firstDayIndex - 1; i >= 0; i--) {
      const day = prevMonthDays - i;
      const prevM = currentMonth === 1 ? 12 : currentMonth - 1;
      const prevY = currentMonth === 1 ? currentYear - 1 : currentYear;
      const dateStr = `${prevY}-${String(prevM).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
      cells.push({
        dayNumber: day,
        dateString: dateStr,
        isCurrentMonth: false,
        isToday: false,
        exEvents: [],
        payEvents: [],
      });
    }

    // Days in current month
    const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
    for (let day = 1; day <= daysInMonth; day++) {
      const dateStr = `${currentYear}-${String(currentMonth).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
      const isToday = dateStr === todayStr;

      const ex = calendarEvents.filter((e) => e.exDate === dateStr);
      const pay = calendarEvents.filter((e) => e.paymentDate === dateStr);

      cells.push({
        dayNumber: day,
        dateString: dateStr,
        isCurrentMonth: true,
        isToday,
        exEvents: ex,
        payEvents: pay,
      });
    }

    // Trailing padding from next month
    const remaining = (7 - (cells.length % 7)) % 7;
    for (let day = 1; day <= remaining; day++) {
      const nextM = currentMonth === 12 ? 1 : currentMonth + 1;
      const nextY = currentMonth === 12 ? currentYear + 1 : currentYear;
      const dateStr = `${nextY}-${String(nextM).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
      cells.push({
        dayNumber: day,
        dateString: dateStr,
        isCurrentMonth: false,
        isToday: false,
        exEvents: [],
        payEvents: [],
      });
    }

    return cells;
  }, [currentYear, currentMonth, calendarEvents]);

  // Compounding yield on cost expansion
  const currentYield = Number(analytics?.portfolioCurrentYield || '0') * 100;
  const yieldOnCost = analytics?.portfolioYieldOnCost ? Number(analytics.portfolioYieldOnCost) * 100 : null;
  const yieldExpansion = yieldOnCost !== null ? yieldOnCost - currentYield : null;

  // Monthly Binned Distribution SVG Chart Calculations
  const chartHeight = 110;
  const chartWidth = 720;
  const barPadding = 12;
  const projections = analytics?.monthlyProjections || [];
  const maxProjected = Math.max(1, ...projections.map((p) => Number(p.projectedIncome.amount)));

  return (
    <div className="dividend-calendar-hub">
      {/* 1. Summary Metric Strip */}
      <div className="div-metric-strip">
        <div className="div-metric-card">
          <span className="div-card-label">Annual Forward Income</span>
          <span className="div-card-val text-emerald">
            {formatMoney(analytics?.annualForwardIncome, preferredCurrency)}
          </span>
          <span className="div-card-sub">
            Monthly Avg: {formatMoney(analytics?.monthlyAverageIncome, preferredCurrency)}
          </span>
        </div>

        <div className="div-metric-card">
          <span className="div-card-label">Portfolio Current Yield</span>
          <span className="div-card-val">
            {formatPercent(analytics?.portfolioCurrentYield, 2, false)}
          </span>
          <span className="div-card-sub">Annualized against market value</span>
        </div>

        <div className="div-metric-card highlight-yoc">
          <span className="div-card-label">Yield on Cost (YOC)</span>
          <div className="div-val-badge-row">
            <span className="div-card-val text-emerald">
              {yieldOnCost !== null ? `${yieldOnCost.toFixed(2)}%` : 'N/A'}
            </span>
            {yieldExpansion !== null && yieldExpansion > 0 && (
              <span className="yoc-expansion-badge">
                +{yieldExpansion.toFixed(2)}% EXPANSION
              </span>
            )}
          </div>
          <span className="div-card-sub">Yield against purchase cost basis</span>
        </div>

        <div className="div-metric-card">
          <span className="div-card-label">Pending Cash Inflows</span>
          <span className="div-card-val text-sky">
            {formatMoney(analytics?.pendingCashInflow, preferredCurrency)}
          </span>
          <span className="div-card-sub">Ex-Date passed, awaiting pay date</span>
        </div>

        <div className="div-metric-card action-card">
          <span className="div-card-label">Wealth Projection</span>
          <button
            type="button"
            className="drip-trigger-btn"
            onClick={() => setIsDRIPOpen(true)}
          >
            ⚡ Simulate DRIP Compounding
          </button>
          <span className="div-card-sub">Model 1–30y reinvestment curves</span>
        </div>
      </div>

      {/* 2. 12-Month Forward Distribution Binned Chart */}
      {projections.length > 0 && (
        <div className="div-chart-card">
          <div className="div-chart-header">
            <h4>12-Month Forward Dividend Distribution</h4>
            <span className="div-chart-sub">Projected monthly dividend cash flow cadence</span>
          </div>

          <div className="div-distribution-svg-wrap">
            <svg
              viewBox={`0 0 ${chartWidth} ${chartHeight}`}
              className="div-dist-svg"
              preserveAspectRatio="none"
            >
              <defs>
                <linearGradient id="barGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#10b981" stopOpacity="0.9" />
                  <stop offset="100%" stopColor="#047857" stopOpacity="0.4" />
                </linearGradient>
                <linearGradient id="barGradHover" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#34d399" stopOpacity="1" />
                  <stop offset="100%" stopColor="#10b981" stopOpacity="0.7" />
                </linearGradient>
              </defs>

              {projections.map((p, idx) => {
                const totalBars = projections.length;
                const slotWidth = chartWidth / totalBars;
                const barW = slotWidth - barPadding;
                const income = Number(p.projectedIncome.amount);
                const barH = maxProjected > 0 ? (income / maxProjected) * (chartHeight - 34) : 0;
                const x = idx * slotWidth + barPadding / 2;
                const y = chartHeight - 24 - barH;
                const isHovered = hoveredBar?.year === p.year && hoveredBar?.month === p.month;

                return (
                  <g
                    key={`${p.year}-${p.month}`}
                    onMouseEnter={() => setHoveredBar(p)}
                    onMouseLeave={() => setHoveredBar(null)}
                    className="dist-bar-group"
                  >
                    {/* Background track */}
                    <rect
                      x={x}
                      y={8}
                      width={barW}
                      height={chartHeight - 32}
                      rx={4}
                      fill="rgba(255, 255, 255, 0.02)"
                    />
                    {/* Active Bar */}
                    <rect
                      x={x}
                      y={y}
                      width={barW}
                      height={Math.max(barH, 3)}
                      rx={4}
                      fill={isHovered ? 'url(#barGradHover)' : 'url(#barGrad)'}
                      className="dist-bar"
                    />
                    {/* Month Label */}
                    <text
                      x={x + barW / 2}
                      y={chartHeight - 8}
                      textAnchor="middle"
                      fontSize="10"
                      fill={isHovered ? '#fff' : 'rgba(255, 255, 255, 0.55)'}
                      fontWeight={isHovered ? '700' : '500'}
                    >
                      {MONTH_NAMES[p.month - 1].slice(0, 3)}
                    </text>
                  </g>
                );
              })}
            </svg>

            {hoveredBar && (
              <div className="dist-tooltip">
                <strong>{MONTH_NAMES[hoveredBar.month - 1]} {hoveredBar.year}</strong>
                <span className="text-emerald">
                  {formatMoney(hoveredBar.projectedIncome, preferredCurrency)}
                </span>
                <span className="tooltip-count">
                  {hoveredBar.dividendCount} {hoveredBar.dividendCount === 1 ? 'payment' : 'payments'}
                </span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* 3. Toolbar & Tri-View Switcher */}
      <div className="div-view-toolbar">
        <div className="div-view-switcher">
          <button
            type="button"
            className={`div-switch-btn ${viewMode === 'calendar' ? 'active' : ''}`}
            onClick={() => setViewMode('calendar')}
          >
            📅 Monthly Calendar Grid
          </button>
          <button
            type="button"
            className={`div-switch-btn ${viewMode === 'agenda' ? 'active' : ''}`}
            onClick={() => setViewMode('agenda')}
          >
            📋 Chronological Agenda List
          </button>
          <button
            type="button"
            className={`div-switch-btn ${viewMode === 'matrix' ? 'active' : ''}`}
            onClick={() => setViewMode('matrix')}
          >
            📊 Holdings Yield Matrix
          </button>
        </div>

        {viewMode === 'calendar' && (
          <div className="div-month-nav">
            <button type="button" className="nav-arrow-btn" onClick={handlePrevMonth} aria-label="Previous month">
              ‹
            </button>
            <span className="month-display">
              {MONTH_NAMES[currentMonth - 1]} {currentYear}
            </span>
            <button type="button" className="nav-arrow-btn" onClick={handleNextMonth} aria-label="Next month">
              ›
            </button>
            <button type="button" className="nav-today-btn" onClick={handleToday}>
              Today
            </button>
            {loading && <span className="div-loading-indicator">Updating...</span>}
          </div>
        )}
      </div>

      {error && <div className="div-error-banner">{error}</div>}

      {/* 4. View 1: Monthly Calendar Grid */}
      {viewMode === 'calendar' && (
        <div className="calendar-grid-card">
          <div className="calendar-week-header">
            {WEEKDAY_NAMES.map((w) => (
              <div key={w} className="calendar-week-cell">
                {w}
              </div>
            ))}
          </div>

          <div className="calendar-days-grid">
            {calendarGrid.map((cell, idx) => (
              <div
                key={idx}
                className={`calendar-day-cell ${cell.isCurrentMonth ? '' : 'outside-month'} ${
                  cell.isToday ? 'is-today' : ''
                }`}
              >
                <div className="day-cell-top">
                  <span className="day-number">{cell.dayNumber}</span>
                  {cell.isToday && <span className="today-badge">TODAY</span>}
                </div>

                <div className="day-events-list">
                  {/* Ex-Date Events */}
                  {cell.exEvents.map((ev) => (
                    <button
                      key={`ex-${ev.id}`}
                      type="button"
                      className={`div-event-pill pill-ex ${ev.isUserEntitled ? 'entitled' : 'unentitled'}`}
                      onClick={() => setSelectedEvent(ev)}
                      title={`Ex-Date: ${ev.symbol} declared ${formatMoney(ev.dpsDeclared)}`}
                    >
                      <span className="pill-type">EX</span>
                      <span className="pill-sym">{ev.symbol}</span>
                      <span className="pill-amt">{formatMoney(ev.dpsDeclared)}</span>
                    </button>
                  ))}

                  {/* Payment Date Events */}
                  {cell.payEvents.map((ev) => (
                    <button
                      key={`pay-${ev.id}`}
                      type="button"
                      className={`div-event-pill pill-pay ${ev.isUserEntitled ? 'entitled' : 'unentitled'}`}
                      onClick={() => setSelectedEvent(ev)}
                      title={`Pay Date: ${ev.symbol} estimated ${formatMoney(ev.grossAmountBase, preferredCurrency)}`}
                    >
                      <span className="pill-type">PAY</span>
                      <span className="pill-sym">{ev.symbol}</span>
                      <span className="pill-amt">{formatMoney(ev.grossAmountBase, preferredCurrency)}</span>
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 5. View 2: Chronological Agenda List */}
      {viewMode === 'agenda' && (
        <div className="agenda-table-card">
          <table className="div-table">
            <thead>
              <tr>
                <th>Ex-Date</th>
                <th>Payment Date</th>
                <th>Asset</th>
                <th>Shares Held</th>
                <th>DPS Declared</th>
                <th>Projected Payout</th>
                <th>Frequency</th>
                <th>Status</th>
                <th>Entitlement</th>
              </tr>
            </thead>
            <tbody>
              {calendarEvents.length === 0 ? (
                <tr>
                  <td colSpan={9} className="table-empty">
                    No scheduled dividend events in {MONTH_NAMES[currentMonth - 1]} {currentYear}.
                  </td>
                </tr>
              ) : (
                calendarEvents.map((ev) => (
                  <tr
                    key={ev.id}
                    onClick={() => setSelectedEvent(ev)}
                    className="agenda-row"
                  >
                    <td>
                      <strong>{formatDate(ev.exDate)}</strong>
                    </td>
                    <td>{formatDate(ev.paymentDate)}</td>
                    <td>
                      <div className="asset-cell">
                        <span className="asset-symbol">{ev.symbol}</span>
                        <span className="asset-name">{ev.name}</span>
                      </div>
                    </td>
                    <td>{formatQuantity(ev.sharesHeld)}</td>
                    <td>{formatMoney(ev.dpsDeclared)}</td>
                    <td className="text-emerald font-semibold">
                      {formatMoney(ev.grossAmountBase, preferredCurrency)}
                    </td>
                    <td>
                      <span className="freq-badge">{ev.frequency}</span>
                    </td>
                    <td>
                      <span className={`status-badge status-${ev.status.toLowerCase()}`}>
                        {ev.status}
                      </span>
                    </td>
                    <td>
                      <span
                        className={`entitled-pill ${ev.isUserEntitled ? 'is-entitled' : 'not-entitled'}`}
                        title={
                          ev.isUserEntitled
                            ? 'Trade settled before Ex-Date'
                            : 'Trade settled after Ex-Date; not eligible for this dividend'
                        }
                      >
                        {ev.isUserEntitled ? '✓ Entitled' : '✗ Unentitled'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* 6. View 3: Holdings Yield Matrix */}
      {viewMode === 'matrix' && (
        <div className="matrix-table-card">
          <table className="div-table">
            <thead>
              <tr>
                <th>Asset</th>
                <th>Quantity</th>
                <th>Current Price</th>
                <th>Avg Buy Price</th>
                <th>Market Value</th>
                <th>Annual DPS</th>
                <th>Projected Cash</th>
                <th>Current Yield</th>
                <th>Yield on Cost</th>
                <th>3Y Div CAGR</th>
                <th>Next Ex-Date</th>
                <th>Next Pay Date</th>
              </tr>
            </thead>
            <tbody>
              {!analytics || analytics.holdingsYield.length === 0 ? (
                <tr>
                  <td colSpan={12} className="table-empty">
                    No active dividend-paying holdings found.
                  </td>
                </tr>
              ) : (
                analytics.holdingsYield.map((h) => {
                  const hYield = Number(h.currentYieldPct) * 100;
                  const hYoc = h.yieldOnCostPct ? Number(h.yieldOnCostPct) * 100 : null;
                  const isExpanding = hYoc !== null && hYoc > hYield;

                  return (
                    <tr key={h.instrumentId}>
                      <td>
                        <div className="asset-cell">
                          <span className="asset-symbol">{h.symbol}</span>
                          <span className="asset-name">{h.name}</span>
                        </div>
                      </td>
                      <td>{formatQuantity(h.quantity)}</td>
                      <td>{formatMoney(h.currentPrice)}</td>
                      <td>{formatMoney(h.averageBuyPrice)}</td>
                      <td>{formatMoney(h.currentMarketValue, preferredCurrency)}</td>
                      <td>{formatMoney(h.annualDps)}</td>
                      <td className="text-emerald font-semibold">
                        {formatMoney(h.projectedAnnualCash, preferredCurrency)}
                      </td>
                      <td>{formatPercent(h.currentYieldPct, 2, false)}</td>
                      <td>
                        <div className="yoc-cell">
                          <span className={isExpanding ? 'text-emerald font-semibold' : ''}>
                            {hYoc !== null ? `${hYoc.toFixed(2)}%` : '-'}
                          </span>
                          {isExpanding && <span className="yoc-arrow" title="Yield on cost expanding">▲</span>}
                        </div>
                      </td>
                      <td>
                        {h.dividendCagr3y !== null && h.dividendCagr3y !== undefined ? (
                          <span className="cagr-badge">
                            {formatPercent(h.dividendCagr3y, 1, true)}
                          </span>
                        ) : (
                          <span className="text-muted">-</span>
                        )}
                      </td>
                      <td>{h.nextExDate ? formatDate(h.nextExDate) : '-'}</td>
                      <td>{h.nextPaymentDate ? formatDate(h.nextPaymentDate) : '-'}</td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* 7. Event Detail Drawer Modal */}
      {selectedEvent && (
        <div
          className="modal-backdrop"
          onClick={() => setSelectedEvent(null)}
          role="dialog"
          aria-modal="true"
        >
          <div className="modal-container div-drawer-container" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div className="drawer-title-group">
                <span className={`status-badge status-${selectedEvent.status.toLowerCase()}`}>
                  {selectedEvent.status}
                </span>
                <h3>{selectedEvent.symbol} – {selectedEvent.name}</h3>
              </div>
              <button
                type="button"
                className="modal-close-btn"
                onClick={() => setSelectedEvent(null)}
              >
                &times;
              </button>
            </div>

            <div className="drawer-body">
              <div className="drawer-grid">
                <div className="drawer-field">
                  <span className="drawer-label">Ex-Dividend Date</span>
                  <strong className="text-amber">{formatDate(selectedEvent.exDate)}</strong>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">Payment Date</span>
                  <strong className="text-emerald">{formatDate(selectedEvent.paymentDate)}</strong>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">Record Date</span>
                  <span>{selectedEvent.recordDate ? formatDate(selectedEvent.recordDate) : 'Not specified'}</span>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">Entitlement Status</span>
                  <span className={`entitled-pill ${selectedEvent.isUserEntitled ? 'is-entitled' : 'not-entitled'}`}>
                    {selectedEvent.isUserEntitled ? '✓ Entitled to Dividend' : '✗ Unentitled (Trade settled after Ex-Date)'}
                  </span>
                </div>

                <div className="drawer-divider" />

                <div className="drawer-field">
                  <span className="drawer-label">Shares Entitled</span>
                  <strong>{formatQuantity(selectedEvent.sharesHeld)}</strong>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">DPS Declared</span>
                  <strong>{formatMoney(selectedEvent.dpsDeclared)}</strong>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">Gross Payout (Base)</span>
                  <strong className="text-emerald">{formatMoney(selectedEvent.grossAmountBase, preferredCurrency)}</strong>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">Net Payout (Base)</span>
                  <strong className="text-emerald">{formatMoney(selectedEvent.netAmountBase, preferredCurrency)}</strong>
                </div>

                <div className="drawer-field">
                  <span className="drawer-label">Cadence Frequency</span>
                  <span>{selectedEvent.frequency}</span>
                </div>
                <div className="drawer-field">
                  <span className="drawer-label">Franking Credits</span>
                  <span>{Number(selectedEvent.frankingPct) > 0 ? `${selectedEvent.frankingPct}% Franked` : 'Unfranked'}</span>
                </div>
              </div>

              {selectedEvent.isEstimated && (
                <div className="drawer-notice">
                  ℹ️ This event date or declared DPS is an estimated projection based on historical corporate action cadence.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 8. DRIP Simulator Modal */}
      <DRIPSimulatorModal
        isOpen={isDRIPOpen}
        onClose={() => setIsDRIPOpen(false)}
        baseCurrency={preferredCurrency}
      />
    </div>
  );
};
