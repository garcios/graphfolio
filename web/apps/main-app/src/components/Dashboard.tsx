import { useEffect, useState, useMemo } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney, formatPercent, isPositive, Button } from '@graphfolio/ui';
import './Dashboard.css';
import { AddTransactionModal } from './AddTransactionModal';
import { ImportTransactionsModal } from './ImportTransactionsModal';
import { PerformanceChart } from './PerformanceChart';
import { TransactionLedger } from './TransactionLedger';
import { ReportsView } from './ReportsView';
import { UserPreferencesModal, type UserPreferencesData, type CurrencyItem } from './UserPreferencesModal';

type SortField = 'ticker' | 'price' | 'averageBuyPrice' | 'quantity' | 'totalValue' | 'capitalGain' | 'income' | 'currencyGain' | 'totalReturn' | 'todayReturn';
type SortDirection = 'asc' | 'desc';

const EMPTY_INVESTMENTS: any[] = [];

export const Dashboard = () => {
  const [data, setData] = useState<any>(null);
  const [userPrefs, setUserPrefs] = useState<UserPreferencesData | null>(null);
  const [supportedCurrencies, setSupportedCurrencies] = useState<CurrencyItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [isPreferencesOpen, setIsPreferencesOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'ledger' | 'reports'>('overview');
  const [ledgerRefreshKey, setLedgerRefreshKey] = useState(0);
  const [sortField, setSortField] = useState<SortField>('totalValue');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  const investments = data?.investments ?? EMPTY_INVESTMENTS;

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(prev => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDirection(field === 'ticker' ? 'asc' : 'desc');
    }
  };

  const sortedInvestments = useMemo(() => {
    if (!sortField) return investments;
    return [...investments].sort((a: any, b: any) => {
      let aVal = 0;
      let bVal = 0;

      if (sortField === 'ticker') {
        const aTicker = (a.ticker || a.name || '').toLowerCase();
        const bTicker = (b.ticker || b.name || '').toLowerCase();
        return sortDirection === 'asc'
          ? aTicker.localeCompare(bTicker)
          : bTicker.localeCompare(aTicker);
      }

      if (sortField === 'averageBuyPrice') {
        const aNum = parseFloat(a.averageBuyPrice?.amount);
        const bNum = parseFloat(b.averageBuyPrice?.amount);
        const aValid = !isNaN(aNum) && aNum > 0;
        const bValid = !isNaN(bNum) && bNum > 0;
        if (!aValid && !bValid) return 0;
        if (!aValid) return 1;
        if (!bValid) return -1;
        aVal = aNum;
        bVal = bNum;
      } else if (sortField === 'price') {
        aVal = parseFloat(a.price?.amount) || 0;
        bVal = parseFloat(b.price?.amount) || 0;
      } else if (sortField === 'quantity') {
        aVal = parseFloat(a.quantity) || 0;
        bVal = parseFloat(b.quantity) || 0;
      } else if (sortField === 'totalValue') {
        aVal = parseFloat(a.totalValue?.amount) || 0;
        bVal = parseFloat(b.totalValue?.amount) || 0;
      } else if (sortField === 'capitalGain') {
        aVal = parseFloat(a.capitalGainAmount?.amount) || 0;
        bVal = parseFloat(b.capitalGainAmount?.amount) || 0;
      } else if (sortField === 'income') {
        aVal = parseFloat(a.incomeAmount?.amount) || 0;
        bVal = parseFloat(b.incomeAmount?.amount) || 0;
      } else if (sortField === 'currencyGain') {
        aVal = parseFloat(a.currencyGainAmount?.amount) || 0;
        bVal = parseFloat(b.currencyGainAmount?.amount) || 0;
      } else if (sortField === 'totalReturn') {
        aVal = parseFloat(a.totalReturnAmount?.amount) || 0;
        bVal = parseFloat(b.totalReturnAmount?.amount) || 0;
      } else if (sortField === 'todayReturn') {
        aVal = parseFloat(a.todayReturnAmount?.amount) || 0;
        bVal = parseFloat(b.todayReturnAmount?.amount) || 0;
      }

      if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
      return (a.ticker || '').localeCompare(b.ticker || '');
    });
  }, [investments, sortField, sortDirection]);

  const renderSortableTh = (
    field: SortField,
    label: string,
    isNumCol: boolean = false,
    title?: string
  ) => {
    const isCurrent = sortField === field;
    const icon = isCurrent ? (sortDirection === 'asc' ? ' ▲' : ' ▼') : ' ↕';
    return (
      <th
        key={field}
        className={`sortable-th ${isNumCol ? 'num-col' : ''}`.trim()}
        onClick={() => handleSort(field)}
        aria-sort={isCurrent ? (sortDirection === 'asc' ? 'ascending' : 'descending') : 'none'}
        title={title || `Sort by ${label}`}
      >
        <div className="th-content">
          <span>{label}</span>
          <span className={`sort-icon ${isCurrent ? 'active' : ''}`}>{icon}</span>
        </div>
      </th>
    );
  };

  const getInitials = (name: string): string => {
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase() || 'OG';
  };

  useEffect(() => {
    client.query({
      portfolio: {
        totalValue: { amount: true, currencyCode: true },
        todayReturnAmount: { amount: true, currencyCode: true },
        todayReturnPercent: true,
        annualizedReturnPercent: true,
        cashBalance: { amount: true, currencyCode: true },
        investments: {
          id: true,
          ticker: true,
          name: true,
          price: { amount: true, currencyCode: true },
          averageBuyPrice: { amount: true, currencyCode: true },
          quantity: true,
          totalValue: { amount: true, currencyCode: true },
          todayReturnAmount: { amount: true, currencyCode: true },
          todayReturnPercent: true,
          totalReturnAmount: { amount: true, currencyCode: true },
          totalReturnPercent: true,
          capitalGainAmount: { amount: true, currencyCode: true },
          capitalGainPercent: true,
          incomeAmount: { amount: true, currencyCode: true },
          incomeYieldPercent: true,
          currencyGainAmount: { amount: true, currencyCode: true },
          currencyGainPercent: true,
          isInternational: true,
        }
      },
      userPreferences: {
        userId: true,
        email: true,
        displayName: true,
        displayCurrency: true,
        theme: true,
      },
      supportedCurrencies: {
        code: true,
        name: true,
        symbol: true,
      }
    })
    .then(res => {
      setData(res.portfolio);
      if (res.userPreferences) {
        setUserPrefs(res.userPreferences);
      }
      if (res.supportedCurrencies) {
        setSupportedCurrencies(res.supportedCurrencies);
      }
      setLoading(false);
    })
    .catch(err => {
      console.warn("Combined dashboard query failed, attempting portfolio-only fallback:", err);
      client.query({
        portfolio: {
          totalValue: { amount: true, currencyCode: true },
          todayReturnAmount: { amount: true, currencyCode: true },
          todayReturnPercent: true,
          annualizedReturnPercent: true,
          cashBalance: { amount: true, currencyCode: true },
          investments: {
            id: true,
            ticker: true,
            name: true,
            price: { amount: true, currencyCode: true },
            averageBuyPrice: { amount: true, currencyCode: true },
            quantity: true,
            totalValue: { amount: true, currencyCode: true },
            todayReturnAmount: { amount: true, currencyCode: true },
            todayReturnPercent: true,
            totalReturnAmount: { amount: true, currencyCode: true },
            totalReturnPercent: true,
            capitalGainAmount: { amount: true, currencyCode: true },
            capitalGainPercent: true,
            incomeAmount: { amount: true, currencyCode: true },
            incomeYieldPercent: true,
            currencyGainAmount: { amount: true, currencyCode: true },
            currencyGainPercent: true,
            isInternational: true,
          }
        }
      })
      .then(res => {
        setData(res.portfolio);
        setLoading(false);
      })
      .catch(fallbackErr => {
        console.error("Error fetching portfolio fallback:", fallbackErr);
        setLoading(false);
      });
    });
  }, []);

  if (loading || !data) {
    return <div className="dashboard loading">Loading portfolio...</div>;
  }

  const todayReturnPositive = isPositive(data.todayReturnAmount);
  const annualizedPositive = isPositive(data.annualizedReturnPercent);

  const currencyCode = data.totalValue?.currencyCode || userPrefs?.displayCurrency || 'USD';

  const investedAssetsTotal = investments.reduce(
    (sum: number, inv: any) => sum + (parseFloat(inv.totalValue?.amount) || 0),
    0
  );
  const subtotalTodayReturn = investments.reduce(
    (sum: number, inv: any) => sum + (parseFloat(inv.todayReturnAmount?.amount) || 0),
    0
  );
  const subtotalTotalReturn = investments.reduce(
    (sum: number, inv: any) => sum + (parseFloat(inv.totalReturnAmount?.amount) || 0),
    0
  );
  const subtotalCapitalGain = investments.reduce(
    (sum: number, inv: any) => sum + (parseFloat(inv.capitalGainAmount?.amount) || 0),
    0
  );
  const subtotalIncome = investments.reduce(
    (sum: number, inv: any) => sum + (parseFloat(inv.incomeAmount?.amount) || 0),
    0
  );
  const subtotalCurrencyGain = investments.reduce(
    (sum: number, inv: any) => sum + (parseFloat(inv.currencyGainAmount?.amount) || 0),
    0
  );

  const subtotalTodayPos = subtotalTodayReturn >= 0;
  const subtotalTotalPos = subtotalTotalReturn >= 0;
  const subtotalCapGainPos = subtotalCapitalGain >= 0;
  const subtotalIncomePos = subtotalIncome >= 0;
  const subtotalCurrencyPos = subtotalCurrencyGain >= 0;

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div>
          <h1 className="dashboard-title">GraphFolio</h1>
          <p className="dashboard-subtitle">Total Portfolio Value</p>
        </div>

        <div className="view-tabs">
          <button
            type="button"
            className={`view-tab-btn ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            Overview
          </button>
          <button
            type="button"
            className={`view-tab-btn ${activeTab === 'ledger' ? 'active' : ''}`}
            onClick={() => setActiveTab('ledger')}
          >
            Transaction Ledger
          </button>
          <button
            type="button"
            className={`view-tab-btn ${activeTab === 'reports' ? 'active' : ''}`}
            onClick={() => setActiveTab('reports')}
          >
            Reports
          </button>
        </div>

        <div className="header-actions">
          <Button
            variant="secondary"
            className="btn-import-header"
            onClick={() => setIsImportModalOpen(true)}
          >
            📥 Import CSV
          </Button>
          <Button
            variant="primary"
            className="btn-add-transaction"
            onClick={() => setIsModalOpen(true)}
          >
            + Add Transaction
          </Button>
          <button
            type="button"
            className="user-profile-btn"
            onClick={() => setIsPreferencesOpen(true)}
            title="Investor Preferences"
            aria-label="Open investor preferences"
          >
            <div className="avatar">
              {getInitials(userPrefs?.displayName || 'Oscar Garcia')}
            </div>
            <div className="user-profile-details">
              <span className="user-profile-name">
                {userPrefs?.displayName || 'Oscar Garcia'}
              </span>
              <span className="user-profile-currency">
                {userPrefs?.displayCurrency || 'USD'}
              </span>
            </div>
          </button>
        </div>
      </header>

      <section className="hero-metrics">
        <div className="metric-primary">
          <h2>{formatMoney(data.totalValue)}</h2>
          <span className={`trend ${todayReturnPositive ? 'positive' : 'negative'}`}>
            {todayReturnPositive ? '+' : ''}{formatMoney(data.todayReturnAmount)} ({formatPercent(data.todayReturnPercent, 2)}) Today
          </span>
        </div>
        <div className="metric-secondary">
          <div className="metric-card">
            <span className="label">Invested Assets</span>
            <span className="value neutral">
              {formatMoney({ amount: investedAssetsTotal.toFixed(2), currencyCode })}
            </span>
          </div>
          <div className="metric-card">
            <span className="label">Cash Balance</span>
            <span className="value neutral">{formatMoney(data.cashBalance)}</span>
          </div>
          <div className="metric-card">
            <span className="label">Annualized Return (TWR)</span>
            <span className={`value ${annualizedPositive ? 'positive' : 'negative'}`}>
              {formatPercent(data.annualizedReturnPercent, 1)}
            </span>
          </div>
        </div>
      </section>

      {activeTab === 'overview' ? (
        <>
          <PerformanceChart key={userPrefs?.displayCurrency || 'USD'} />

          <section className="investments-section">
            <div className="table-card">
              <h3>Your Investments</h3>
              <div className="table-responsive">
                <table className="investments-table">
                  <thead>
                    <tr>
                      {renderSortableTh('ticker', 'Asset')}
                      {renderSortableTh('price', 'Price', true)}
                      {renderSortableTh(
                        'averageBuyPrice',
                        'Avg Buy Price',
                        true,
                        'The weighted average price paid per share/unit across all open lots.'
                      )}
                      {renderSortableTh('quantity', 'Quantity', true)}
                      {renderSortableTh('totalValue', 'Total Value', true)}
                      {renderSortableTh('capitalGain', 'Capital Gain')}
                      {renderSortableTh('income', 'Income')}
                      {renderSortableTh('currencyGain', 'Currency Gain')}
                      {renderSortableTh('totalReturn', 'Total Return')}
                      {renderSortableTh('todayReturn', "Today's Return")}
                    </tr>
                  </thead>
                  <tbody>
                    {sortedInvestments.map((inv: any) => {
                      const todayPos = isPositive(inv.todayReturnAmount);
                      const totalPos = isPositive(inv.totalReturnAmount);
                      const capGainPos = isPositive(inv.capitalGainAmount);
                      const incomePos = isPositive(inv.incomeAmount);
                      const currencyPos = isPositive(inv.currencyGainAmount);

                      return (
                        <tr key={inv.id}>
                          <td>
                            <div className="asset-info">
                              <div className="ticker-wrapper">
                                <span className="ticker">{inv.ticker}</span>
                                {inv.isInternational && (
                                  <span className="intl-badge" title={`International holding denominated in ${inv.price?.currencyCode || 'foreign currency'}`}>INTL</span>
                                )}
                              </div>
                              <span className="name">{inv.name}</span>
                            </div>
                          </td>
                          <td className="num-col">{formatMoney(inv.price)}</td>
                          <td className="num-col">
                            {inv.averageBuyPrice && parseFloat(inv.averageBuyPrice.amount) > 0 ? (
                              <span className="avg-buy-price-val">{formatMoney(inv.averageBuyPrice)}</span>
                            ) : (
                              <span className="neutral-dash" title="Cost basis unavailable">—</span>
                            )}
                          </td>
                          <td className="num-col">{inv.quantity}</td>
                          <td className="num-col">{formatMoney(inv.totalValue)}</td>
                          <td>
                            <div className={`return-info ${capGainPos ? 'positive' : 'negative'}`}>
                              <span className="amount">{capGainPos ? '+' : ''}{formatMoney(inv.capitalGainAmount)}</span>
                              <span className="percent">{formatPercent(inv.capitalGainPercent, 2)}</span>
                            </div>
                          </td>
                          <td>
                            <div className={`return-info ${incomePos ? 'positive' : 'neutral'}`}>
                              <span className="amount">{incomePos ? '+' : ''}{formatMoney(inv.incomeAmount)}</span>
                              <span className="percent">{formatPercent(inv.incomeYieldPercent, 2)} YOC</span>
                            </div>
                          </td>
                          <td>
                            {inv.isInternational ? (
                              <div className={`return-info ${currencyPos ? 'positive' : 'negative'}`}>
                                <span className="amount">{currencyPos ? '+' : ''}{formatMoney(inv.currencyGainAmount)}</span>
                                <span className="percent">{formatPercent(inv.currencyGainPercent, 2)}</span>
                              </div>
                            ) : (
                              <span className="neutral-dash" title="Domestic holding — zero currency exposure">—</span>
                            )}
                          </td>
                          <td>
                            <div className={`return-info ${totalPos ? 'positive' : 'negative'}`}>
                              <span className="amount">{totalPos ? '+' : ''}{formatMoney(inv.totalReturnAmount)}</span>
                              <span className="percent">{formatPercent(inv.totalReturnPercent, 1)}</span>
                            </div>
                          </td>
                          <td>
                            <div className={`return-info ${todayPos ? 'positive' : 'negative'}`}>
                              <span className="amount">{todayPos ? '+' : ''}{formatMoney(inv.todayReturnAmount)}</span>
                              <span className="percent">{formatPercent(inv.todayReturnPercent, 2)}</span>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                  <tfoot>
                    <tr className="table-footer-subtotal">
                      <td colSpan={4}>
                        <div className="footer-title-cell">
                          <span className="footer-title">Invested Assets Subtotal</span>
                          <span className="footer-count">{investments.length} {investments.length === 1 ? 'asset' : 'assets'}</span>
                        </div>
                      </td>
                      <td className="footer-amount">
                        {formatMoney({ amount: investedAssetsTotal.toFixed(2), currencyCode })}
                      </td>
                      <td>
                        <div className={`return-info ${subtotalCapGainPos ? 'positive' : 'negative'}`}>
                          <span className="amount">
                            {subtotalCapGainPos ? '+' : ''}{formatMoney({ amount: subtotalCapitalGain.toFixed(2), currencyCode })}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div className={`return-info ${subtotalIncomePos ? 'positive' : 'neutral'}`}>
                          <span className="amount">
                            {subtotalIncomePos ? '+' : ''}{formatMoney({ amount: subtotalIncome.toFixed(2), currencyCode })}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div className={`return-info ${subtotalCurrencyPos ? 'positive' : 'negative'}`}>
                          <span className="amount">
                            {subtotalCurrencyPos ? '+' : ''}{formatMoney({ amount: subtotalCurrencyGain.toFixed(2), currencyCode })}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div className={`return-info ${subtotalTotalPos ? 'positive' : 'negative'}`}>
                          <span className="amount">
                            {subtotalTotalPos ? '+' : ''}{formatMoney({ amount: subtotalTotalReturn.toFixed(2), currencyCode })}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div className={`return-info ${subtotalTodayPos ? 'positive' : 'negative'}`}>
                          <span className="amount">
                            {subtotalTodayPos ? '+' : ''}{formatMoney({ amount: subtotalTodayReturn.toFixed(2), currencyCode })}
                          </span>
                        </div>
                      </td>
                    </tr>
                    <tr className="table-footer-cash">
                      <td colSpan={4}>
                        <div className="footer-title-cell">
                          <span className="footer-title">Cash Balance</span>
                          <span className="cash-pill">Liquid</span>
                        </div>
                      </td>
                      <td className="footer-amount">
                        {formatMoney(data.cashBalance)}
                      </td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                    </tr>
                    <tr className="table-footer-total">
                      <td colSpan={4}>
                        <div className="footer-title-cell">
                          <span className="footer-title-total">Total Portfolio Value</span>
                          <span className="footer-formula">Assets + Cash</span>
                        </div>
                      </td>
                      <td className="footer-amount-total">
                        {formatMoney(data.totalValue)}
                      </td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                      <td className="footer-muted">—</td>
                      <td>
                        <div className={`return-info ${todayReturnPositive ? 'positive' : 'negative'}`}>
                          <span className="amount">
                            {todayReturnPositive ? '+' : ''}{formatMoney(data.todayReturnAmount)}
                          </span>
                          <span className="percent">{formatPercent(data.todayReturnPercent, 2)}</span>
                        </div>
                      </td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          </section>
        </>
      ) : activeTab === 'ledger' ? (
        <TransactionLedger
          refreshTrigger={ledgerRefreshKey}
          onOpenImportModal={() => setIsImportModalOpen(true)}
          onTransactionDeleted={(updatedPortfolio) => {
            setData(updatedPortfolio);
            setToastMessage('Transaction deleted and projections recomputed!');
            setTimeout(() => setToastMessage(null), 4000);
          }}
        />
      ) : (
        <ReportsView preferredCurrency={userPrefs?.displayCurrency} />
      )}

      <AddTransactionModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        preferredCurrency={userPrefs?.displayCurrency}
        supportedCurrencies={supportedCurrencies}
        onSuccess={(updatedPortfolio) => {
          setData(updatedPortfolio);
          setLedgerRefreshKey((k) => k + 1);
          setToastMessage('Transaction recorded and projections updated!');
          setTimeout(() => setToastMessage(null), 4000);
        }}
      />

      <ImportTransactionsModal
        isOpen={isImportModalOpen}
        onClose={() => setIsImportModalOpen(false)}
        onSuccess={(updatedPortfolio, message) => {
          if (updatedPortfolio) {
            setData(updatedPortfolio);
          }
          setLedgerRefreshKey((k) => k + 1);
          setToastMessage(message || 'Transactions imported and projections updated!');
          setTimeout(() => setToastMessage(null), 4000);
        }}
      />

      <UserPreferencesModal
        isOpen={isPreferencesOpen}
        onClose={() => setIsPreferencesOpen(false)}
        currentPreferences={userPrefs}
        supportedCurrencies={supportedCurrencies}
        onSuccess={(updatedPrefs, updatedPortfolio) => {
          setUserPrefs(updatedPrefs);
          if (updatedPortfolio) {
            setData(updatedPortfolio);
          }
          setLedgerRefreshKey((k) => k + 1);
          const currSymbol = supportedCurrencies.find((c) => c.code === updatedPrefs.displayCurrency)?.symbol || '$';
          setToastMessage(`Preferences saved. Display currency updated to ${updatedPrefs.displayCurrency} (${currSymbol})`);
          setTimeout(() => setToastMessage(null), 4000);
        }}
      />

      {toastMessage && (
        <div className="toast-notification">
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
};
