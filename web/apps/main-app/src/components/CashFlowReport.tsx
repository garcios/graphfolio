import React, { useState, useEffect, useMemo, useRef } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney } from '@graphfolio/ui';
import './CashFlowReport.css';

type CashFlowTimeframe = 'MTD' | 'YTD' | 'M1' | 'M3' | 'M6' | 'Y1' | 'ALL' | 'CUSTOM';
type SortField = 'date' | 'amount' | 'type';
type SortDirection = 'asc' | 'desc';

export const CATEGORY_OPTIONS = [
  { id: 'CAPITAL_DEPOSITS', label: 'Capital Deposits', shortLabel: 'Deposits' },
  { id: 'DIVIDENDS', label: 'Dividends Received', shortLabel: 'Dividends' },
  { id: 'INTEREST', label: 'Interest Earned', shortLabel: 'Interest' },
  { id: 'SALE_PROCEEDS', label: 'Stock Sale Proceeds', shortLabel: 'Sales' },
  { id: 'CAPITAL_WITHDRAWALS', label: 'Capital Withdrawals', shortLabel: 'Withdrawals' },
  { id: 'PURCHASES', label: 'Security Purchases', shortLabel: 'Purchases' },
  { id: 'FEES', label: 'Brokerage & Fees', shortLabel: 'Fees' },
  { id: 'TAXES', label: 'Taxes & Withholding', shortLabel: 'Taxes' },
] as const;

export const ALL_CATEGORY_IDS = CATEGORY_OPTIONS.map((c) => c.id);

interface CashFlowReportProps {
  preferredCurrency?: string;
}

export const CashFlowReport: React.FC<CashFlowReportProps> = ({ preferredCurrency }) => {
  const [timeframe, setTimeframe] = useState<CashFlowTimeframe>('MTD');
  const [customFromDate, setCustomFromDate] = useState<string>('');
  const [customToDate, setCustomToDate] = useState<string>('');
  const [reportData, setReportData] = useState<any>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const [searchQuery, setSearchQuery] = useState<string>('');
  const [selectedCategories, setSelectedCategories] = useState<string[]>([...ALL_CATEGORY_IDS]);
  const [isCategoryMenuOpen, setIsCategoryMenuOpen] = useState<boolean>(false);
  const categoryMenuRef = useRef<HTMLDivElement>(null);
  const [sortField, setSortField] = useState<SortField>('date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  const fetchReport = () => {
    setLoading(true);
    setError(null);

    const filterArgs: any = {
      timeframe,
    };
    if (timeframe === 'CUSTOM') {
      if (customFromDate) filterArgs.fromDate = customFromDate;
      if (customToDate) filterArgs.toDate = customToDate;
    }
    if (preferredCurrency) {
      filterArgs.currency = preferredCurrency;
    }

    client
      .query({
        cashFlowReport: {
          __args: { filter: filterArgs },
          baseCurrency: true,
          fromDate: true,
          toDate: true,
          summary: {
            startingCashBalance: { amount: true, currencyCode: true },
            totalInflows: { amount: true, currencyCode: true },
            totalOutflows: { amount: true, currencyCode: true },
            netCashFlow: { amount: true, currencyCode: true },
            endingCashBalance: { amount: true, currencyCode: true },
          },
          breakdown: {
            deposits: { amount: true, currencyCode: true },
            dividends: { amount: true, currencyCode: true },
            interest: { amount: true, currencyCode: true },
            salesProceeds: { amount: true, currencyCode: true },
            withdrawals: { amount: true, currencyCode: true },
            purchases: { amount: true, currencyCode: true },
            fees: { amount: true, currencyCode: true },
            taxes: { amount: true, currencyCode: true },
          },
          items: {
            id: true,
            eventDate: true,
            type: true,
            flowDirection: true,
            category: true,
            symbol: true,
            instrumentName: true,
            description: true,
            netAmount: { amount: true, currencyCode: true },
            runningBalance: { amount: true, currencyCode: true },
            localAmount: { amount: true, currencyCode: true },
            fee: { amount: true, currencyCode: true },
            withholdingTax: { amount: true, currencyCode: true },
          },
        },
      })
      .then((res: any) => {
        setReportData(res.cashFlowReport);
        setLoading(false);
      })
      .catch((err: any) => {
        setError(err.message || 'Failed to load cash flow report');
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchReport();
  }, [timeframe, customFromDate, customToDate, preferredCurrency]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (categoryMenuRef.current && !categoryMenuRef.current.contains(event.target as Node)) {
        setIsCategoryMenuOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsCategoryMenuOpen(false);
      }
    };

    if (isCategoryMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isCategoryMenuOpen]);

  const items = reportData?.items || [];

  const categoryCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const item of items) {
      if (item.category) {
        counts[item.category] = (counts[item.category] || 0) + 1;
      }
    }
    return counts;
  }, [items]);

  const handleToggleCategory = (catId: string) => {
    setSelectedCategories((prev) =>
      prev.includes(catId) ? prev.filter((id) => id !== catId) : [...prev, catId]
    );
  };

  const handleSelectAll = () => {
    setSelectedCategories([...ALL_CATEGORY_IDS]);
  };

  const handleClearAll = () => {
    setSelectedCategories([]);
  };

  const getSelectionLabel = () => {
    if (selectedCategories.length === 0) {
      return 'No Categories';
    }
    if (selectedCategories.length === ALL_CATEGORY_IDS.length) {
      return 'All Categories';
    }
    if (selectedCategories.length === 1) {
      const match = CATEGORY_OPTIONS.find((c) => c.id === selectedCategories[0]);
      return match ? match.shortLabel : '1 Category';
    }
    if (selectedCategories.length === 2) {
      const cat1 = CATEGORY_OPTIONS.find((c) => c.id === selectedCategories[0]);
      const cat2 = CATEGORY_OPTIONS.find((c) => c.id === selectedCategories[1]);
      if (cat1 && cat2) {
        return `${cat1.shortLabel}, ${cat2.shortLabel}`;
      }
    }
    return `${selectedCategories.length} Categories`;
  };

  const filteredItems = useMemo(() => {
    return items.filter((item: any) => {
      if (selectedCategories.length === 0) {
        return false;
      }
      if (!selectedCategories.includes(item.category)) {
        return false;
      }
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const sym = (item.symbol || '').toLowerCase();
        const desc = (item.description || '').toLowerCase();
        const type = (item.type || '').toLowerCase();
        return sym.includes(q) || desc.includes(q) || type.includes(q);
      }
      return true;
    });
  }, [items, selectedCategories, searchQuery]);

  const sortedItems = useMemo(() => {
    return [...filteredItems].sort((a: any, b: any) => {
      if (sortField === 'date') {
        const dateA = new Date(a.eventDate).getTime();
        const dateB = new Date(b.eventDate).getTime();
        return sortDirection === 'asc' ? dateA - dateB : dateB - dateA;
      }
      if (sortField === 'amount') {
        const amtA = parseFloat(a.netAmount?.amount) || 0;
        const amtB = parseFloat(b.netAmount?.amount) || 0;
        return sortDirection === 'asc' ? amtA - amtB : amtB - amtA;
      }
      if (sortField === 'type') {
        return sortDirection === 'asc'
          ? (a.type || '').localeCompare(b.type || '')
          : (b.type || '').localeCompare(a.type || '');
      }
      return 0;
    });
  }, [filteredItems, sortField, sortDirection]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDirection(field === 'date' ? 'desc' : 'desc');
    }
  };

  const handleExportCSV = () => {
    if (!reportData) return;

    const summary = reportData.summary;
    const breakdown = reportData.breakdown;
    const baseCcy = reportData.baseCurrency;

    const lines: string[] = [];
    lines.push('GraphFolio Portfolio Cash Flow Report');
    lines.push(`Reporting Period,${reportData.fromDate || 'Inception'} to ${reportData.toDate}`);
    lines.push(`Base Currency,${baseCcy}`);
    lines.push(`Exported At,${new Date().toISOString()}`);
    lines.push('');

    // Summary Section
    lines.push('--- EXECUTIVE CASH SUMMARY ---');
    lines.push(`Starting Cash Balance,${summary.startingCashBalance?.amount}`);
    lines.push(`Total Inflows (+),${summary.totalInflows?.amount}`);
    lines.push(`Total Outflows (-),${summary.totalOutflows?.amount}`);
    lines.push(`Net Cash Movement,${summary.netCashFlow?.amount}`);
    lines.push(`Ending Cash Balance,${summary.endingCashBalance?.amount}`);
    lines.push('');

    // Category Breakdown Section
    lines.push('--- CATEGORY BREAKDOWN ---');
    lines.push(`Capital Deposits,${breakdown.deposits?.amount}`);
    lines.push(`Dividends Received,${breakdown.dividends?.amount}`);
    lines.push(`Cash Interest Earned,${breakdown.interest?.amount}`);
    lines.push(`Stock Sale Proceeds,${breakdown.salesProceeds?.amount}`);
    lines.push(`Capital Withdrawals,${breakdown.withdrawals?.amount}`);
    lines.push(`Security Purchases,${breakdown.purchases?.amount}`);
    lines.push(`Brokerage & Custody Fees,${breakdown.fees?.amount}`);
    lines.push(`Withholding & Account Taxes,${breakdown.taxes?.amount}`);
    lines.push('');

    // Ledger Items Section
    lines.push('--- ITEMIZED CASH FLOW LEDGER ---');
    lines.push('Event Date,Type,Category,Symbol,Description,Net Amount,Running Balance,Local Amount,Fee,Withholding Tax');

    for (const item of items) {
      const row = [
        item.eventDate,
        item.type,
        item.category,
        `"${item.symbol || ''}"`,
        `"${(item.description || '').replace(/"/g, '""')}"`,
        item.netAmount?.amount,
        item.runningBalance?.amount,
        `${item.localAmount?.amount} ${item.localAmount?.currencyCode}`,
        item.fee?.amount,
        item.withholdingTax?.amount,
      ];
      lines.push(row.join(','));
    }

    const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `GraphFolio_CashFlow_${reportData.toDate || 'Report'}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const summary = reportData?.summary;
  const breakdown = reportData?.breakdown;
  const baseCurrency = reportData?.baseCurrency || 'USD';

  const totalInflowsNum = parseFloat(summary?.totalInflows?.amount) || 0;
  const totalOutflowsNum = parseFloat(summary?.totalOutflows?.amount) || 0;
  const netCashFlowNum = parseFloat(summary?.netCashFlow?.amount) || 0;

  const getInflowPct = (amtStr?: string) => {
    if (!totalInflowsNum) return 0;
    const n = parseFloat(amtStr || '0') || 0;
    return Math.min(100, Math.round((n / totalInflowsNum) * 100));
  };

  const getOutflowPct = (amtStr?: string) => {
    if (!totalOutflowsNum) return 0;
    const n = parseFloat(amtStr || '0') || 0;
    return Math.min(100, Math.round((n / totalOutflowsNum) * 100));
  };

  return (
    <div className="cash-flow-report">
      {/* Control Toolbar */}
      <div className="cfr-toolbar">
        <div className="cfr-timeframe-group">
          {(['MTD', 'YTD', 'M1', 'M3', 'M6', 'Y1', 'ALL', 'CUSTOM'] as CashFlowTimeframe[]).map((tf) => (
            <button
              key={tf}
              type="button"
              className={`cfr-pill-btn ${timeframe === tf ? 'active' : ''}`}
              onClick={() => setTimeframe(tf)}
            >
              {tf === 'M1' ? '1M' : tf === 'M3' ? '3M' : tf === 'M6' ? '6M' : tf === 'Y1' ? '1Y' : tf === 'ALL' ? 'All Time' : tf}
            </button>
          ))}

          {timeframe === 'CUSTOM' && (
            <div className="cfr-custom-dates">
              <input
                type="date"
                className="cfr-date-input"
                value={customFromDate}
                onChange={(e) => setCustomFromDate(e.target.value)}
                title="Start date"
              />
              <span style={{ color: 'var(--text-muted)' }}>to</span>
              <input
                type="date"
                className="cfr-date-input"
                value={customToDate}
                onChange={(e) => setCustomToDate(e.target.value)}
                title="End date"
              />
            </div>
          )}
        </div>

        <div className="cfr-toolbar-actions">
          <button
            type="button"
            className="btn-export-csv"
            onClick={handleExportCSV}
            disabled={!reportData || loading}
            title="Export complete report to CSV"
          >
            <span>📥 Export CSV</span>
          </button>
        </div>
      </div>

      {loading && !reportData ? (
        <div className="cfr-loading">Computing exact cash flows and reconciling running balances...</div>
      ) : error ? (
        <div className="cfr-empty-state" style={{ color: 'var(--accent-red)' }}>
          Error: {error}
        </div>
      ) : (
        <>
          {/* Top 5 KPI Summary Cards */}
          <div className="cfr-kpi-grid">
            <div className="cfr-kpi-card">
              <span className="cfr-kpi-label">Starting Cash</span>
              <span className="cfr-kpi-value">{formatMoney(summary?.startingCashBalance)}</span>
              <span className="cfr-kpi-subtext">As of period open</span>
            </div>

            <div className="cfr-kpi-card">
              <span className="cfr-kpi-label">Total Inflows</span>
              <span className="cfr-kpi-value positive">+{formatMoney(summary?.totalInflows)}</span>
              <span className="cfr-kpi-subtext">Deposits, dividends, interest, sales</span>
            </div>

            <div className="cfr-kpi-card">
              <span className="cfr-kpi-label">Total Outflows</span>
              <span className="cfr-kpi-value negative">-{formatMoney(summary?.totalOutflows)}</span>
              <span className="cfr-kpi-subtext">Withdrawals, purchases, fees, taxes</span>
            </div>

            <div className="cfr-kpi-card">
              <span className="cfr-kpi-label">Net Cash Movement</span>
              <span className={`cfr-kpi-value ${netCashFlowNum >= 0 ? 'positive' : 'negative'}`}>
                {netCashFlowNum >= 0 ? '+' : ''}{formatMoney(summary?.netCashFlow)}
              </span>
              <span className="cfr-kpi-subtext">Inflows minus Outflows</span>
            </div>

            <div className="cfr-kpi-card reconciled">
              <span className="cfr-kpi-label">Ending Cash</span>
              <span className="cfr-kpi-value">{formatMoney(summary?.endingCashBalance)}</span>
              <span className="reconciled-badge">✓ Reconciled Balance</span>
            </div>
          </div>

          {/* Category Breakdown Panels */}
          <div className="cfr-breakdown-section">
            <div className="cfr-breakdown-card inflows">
              <h4>
                <span>Inflows Breakdown</span>
                <span>+{formatMoney(summary?.totalInflows)}</span>
              </h4>
              <div className="cfr-breakdown-list">
                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Capital Deposits</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.deposits)} ({getInflowPct(breakdown?.deposits?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getInflowPct(breakdown?.deposits?.amount)}%` }} />
                  </div>
                </div>

                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Dividends Received</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.dividends)} ({getInflowPct(breakdown?.dividends?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getInflowPct(breakdown?.dividends?.amount)}%` }} />
                  </div>
                </div>

                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Interest Earned</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.interest)} ({getInflowPct(breakdown?.interest?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getInflowPct(breakdown?.interest?.amount)}%` }} />
                  </div>
                </div>

                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Stock Sale Proceeds</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.salesProceeds)} ({getInflowPct(breakdown?.salesProceeds?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getInflowPct(breakdown?.salesProceeds?.amount)}%` }} />
                  </div>
                </div>
              </div>
            </div>

            <div className="cfr-breakdown-card outflows">
              <h4>
                <span>Outflows Breakdown</span>
                <span>-{formatMoney(summary?.totalOutflows)}</span>
              </h4>
              <div className="cfr-breakdown-list">
                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Capital Withdrawals</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.withdrawals)} ({getOutflowPct(breakdown?.withdrawals?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getOutflowPct(breakdown?.withdrawals?.amount)}%` }} />
                  </div>
                </div>

                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Security Purchases</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.purchases)} ({getOutflowPct(breakdown?.purchases?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getOutflowPct(breakdown?.purchases?.amount)}%` }} />
                  </div>
                </div>

                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Brokerage & Custody Fees</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.fees)} ({getOutflowPct(breakdown?.fees?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getOutflowPct(breakdown?.fees?.amount)}%` }} />
                  </div>
                </div>

                <div className="cfr-breakdown-item">
                  <div className="cfr-breakdown-row">
                    <span className="cfr-category-name">Taxes & Withholding</span>
                    <span className="cfr-category-amt">{formatMoney(breakdown?.taxes)} ({getOutflowPct(breakdown?.taxes?.amount)}%)</span>
                  </div>
                  <div className="cfr-progress-track">
                    <div className="cfr-progress-fill" style={{ width: `${getOutflowPct(breakdown?.taxes?.amount)}%` }} />
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Itemized Cash Flow Table */}
          <div className="cfr-ledger-card">
            <div className="cfr-ledger-header">
              <h3>Itemized Cash Events ({filteredItems.length})</h3>

              <div className="cfr-ledger-filters">
                <input
                  type="text"
                  className="cfr-search-input"
                  placeholder="Filter by symbol or description..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                />

                <div className="cfr-multiselect" ref={categoryMenuRef}>
                  <button
                    type="button"
                    className={`cfr-multiselect-trigger ${isCategoryMenuOpen ? 'open' : ''} ${
                      selectedCategories.length < ALL_CATEGORY_IDS.length ? 'filtered' : ''
                    }`}
                    onClick={() => setIsCategoryMenuOpen((prev) => !prev)}
                    aria-haspopup="listbox"
                    aria-expanded={isCategoryMenuOpen}
                    title="Filter by categories"
                  >
                    <span className="cfr-multiselect-label">{getSelectionLabel()}</span>
                    {selectedCategories.length > 0 && selectedCategories.length < ALL_CATEGORY_IDS.length && (
                      <span className="cfr-multiselect-badge">{selectedCategories.length}</span>
                    )}
                    <span className="cfr-multiselect-caret">{isCategoryMenuOpen ? '▲' : '▼'}</span>
                  </button>

                  {isCategoryMenuOpen && (
                    <div className="cfr-multiselect-popover" role="listbox">
                      <div className="cfr-multiselect-actions">
                        <button
                          type="button"
                          className="cfr-multiselect-action-btn"
                          onClick={handleSelectAll}
                          disabled={selectedCategories.length === ALL_CATEGORY_IDS.length}
                        >
                          Select All
                        </button>
                        <span className="cfr-multiselect-action-divider">•</span>
                        <button
                          type="button"
                          className="cfr-multiselect-action-btn"
                          onClick={handleClearAll}
                          disabled={selectedCategories.length === 0}
                        >
                          Clear All
                        </button>
                      </div>

                      <div className="cfr-multiselect-list">
                        {CATEGORY_OPTIONS.map((cat) => {
                          const isChecked = selectedCategories.includes(cat.id);
                          const count = categoryCounts[cat.id] || 0;
                          return (
                            <label
                              key={cat.id}
                              className={`cfr-multiselect-item ${isChecked ? 'selected' : ''}`}
                            >
                              <input
                                type="checkbox"
                                className="cfr-multiselect-checkbox"
                                checked={isChecked}
                                onChange={() => handleToggleCategory(cat.id)}
                              />
                              <span className="cfr-multiselect-item-text">{cat.label}</span>
                              <button
                                type="button"
                                className="cfr-multiselect-only-btn"
                                onClick={(e) => {
                                  e.preventDefault();
                                  e.stopPropagation();
                                  setSelectedCategories([cat.id]);
                                }}
                                title={`Show only ${cat.label}`}
                              >
                                Only
                              </button>
                              <span className="cfr-multiselect-item-count">{count}</span>
                            </label>
                          );
                        })}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>

            <div className="cfr-table-responsive">
              <table className="cfr-table">
                <thead>
                  <tr>
                    <th className="sortable" onClick={() => handleSort('date')}>
                      Settlement Date {sortField === 'date' ? (sortDirection === 'asc' ? '▲' : '▼') : '↕'}
                    </th>
                    <th className="sortable" onClick={() => handleSort('type')}>
                      Type {sortField === 'type' ? (sortDirection === 'asc' ? '▲' : '▼') : '↕'}
                    </th>
                    <th>Asset</th>
                    <th>Description</th>
                    <th className="sortable text-right" onClick={() => handleSort('amount')}>
                      Cash Flow ({baseCurrency}) {sortField === 'amount' ? (sortDirection === 'asc' ? '▲' : '▼') : '↕'}
                    </th>
                    <th className="text-right">Running Balance</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedItems.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="cfr-empty-state">
                        {selectedCategories.length === 0 ? (
                          <div className="cfr-empty-state-content">
                            <p>No categories selected.</p>
                            <button
                              type="button"
                              className="cfr-pill-btn active"
                              onClick={handleSelectAll}
                              style={{ marginTop: '0.75rem' }}
                            >
                              Select All Categories
                            </button>
                          </div>
                        ) : searchQuery.trim() ? (
                          'No cash events match the search query.'
                        ) : (
                          'No cash events recorded for this timeframe.'
                        )}
                      </td>
                    </tr>
                  ) : (
                    sortedItems.map((item: any) => {
                      const isInflow = item.flowDirection === 'INFLOW';
                      const netAmtNum = parseFloat(item.netAmount?.amount) || 0;
                      return (
                        <tr key={item.id}>
                          <td>{item.eventDate}</td>
                          <td>
                            <span className={`cfr-type-badge ${isInflow ? 'inflow' : 'outflow'}`}>
                              {item.type}
                            </span>
                          </td>
                          <td>
                            {item.symbol ? (
                              <strong style={{ color: 'var(--text-primary)' }}>{item.symbol}</strong>
                            ) : (
                              <span style={{ color: 'var(--text-muted)' }}>—</span>
                            )}
                          </td>
                          <td style={{ color: 'var(--text-secondary)' }}>{item.description}</td>
                          <td className="text-right">
                            <span className={`cfr-flow-amt ${netAmtNum >= 0 ? 'positive' : 'negative'}`}>
                              {netAmtNum >= 0 ? '+' : ''}{formatMoney(item.netAmount)}
                            </span>
                          </td>
                          <td className="text-right">
                            <span className="cfr-running-bal">{formatMoney(item.runningBalance)}</span>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
