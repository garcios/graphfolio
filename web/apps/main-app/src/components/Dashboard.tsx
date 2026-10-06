import { useEffect, useState } from 'react';
import { client } from '@graphfolio/api-client';
import { formatMoney, formatPercent, isPositive, Button } from '@graphfolio/ui';
import './Dashboard.css';
import { AddTransactionModal } from './AddTransactionModal';
import { PerformanceChart } from './PerformanceChart';
import { TransactionLedger } from './TransactionLedger';
import { UserPreferencesModal, type UserPreferencesData, type CurrencyItem } from './UserPreferencesModal';

export const Dashboard = () => {
  const [data, setData] = useState<any>(null);
  const [userPrefs, setUserPrefs] = useState<UserPreferencesData | null>(null);
  const [supportedCurrencies, setSupportedCurrencies] = useState<CurrencyItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPreferencesOpen, setIsPreferencesOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'ledger'>('overview');
  const [ledgerRefreshKey, setLedgerRefreshKey] = useState(0);

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
          quantity: true,
          totalValue: { amount: true, currencyCode: true },
          todayReturnAmount: { amount: true, currencyCode: true },
          todayReturnPercent: true,
          totalReturnAmount: { amount: true, currencyCode: true },
          totalReturnPercent: true,
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
            quantity: true,
            totalValue: { amount: true, currencyCode: true },
            todayReturnAmount: { amount: true, currencyCode: true },
            todayReturnPercent: true,
            totalReturnAmount: { amount: true, currencyCode: true },
            totalReturnPercent: true,
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
        </div>

        <div className="header-actions">
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
            <span className="label">Annualized Return (TWR)</span>
            <span className={`value ${annualizedPositive ? 'positive' : 'negative'}`}>
              {formatPercent(data.annualizedReturnPercent, 1)}
            </span>
          </div>
          <div className="metric-card">
            <span className="label">Cash Balance</span>
            <span className="value neutral">{formatMoney(data.cashBalance)}</span>
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
                      <th>Asset</th>
                      <th>Price</th>
                      <th>Quantity</th>
                      <th>Total Value</th>
                      <th>Today's Return</th>
                      <th>Total Return</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.investments.map((inv: any) => {
                      const todayPos = isPositive(inv.todayReturnAmount);
                      const totalPos = isPositive(inv.totalReturnAmount);

                      return (
                        <tr key={inv.id}>
                          <td>
                            <div className="asset-info">
                              <span className="ticker">{inv.ticker}</span>
                              <span className="name">{inv.name}</span>
                            </div>
                          </td>
                          <td>{formatMoney(inv.price)}</td>
                          <td>{inv.quantity}</td>
                          <td>{formatMoney(inv.totalValue)}</td>
                          <td>
                            <div className={`return-info ${todayPos ? 'positive' : 'negative'}`}>
                              <span className="amount">{todayPos ? '+' : ''}{formatMoney(inv.todayReturnAmount)}</span>
                              <span className="percent">{formatPercent(inv.todayReturnPercent, 2)}</span>
                            </div>
                          </td>
                          <td>
                            <div className={`return-info ${totalPos ? 'positive' : 'negative'}`}>
                              <span className="amount">{totalPos ? '+' : ''}{formatMoney(inv.totalReturnAmount)}</span>
                              <span className="percent">{formatPercent(inv.totalReturnPercent, 1)}</span>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          </section>
        </>
      ) : (
        <TransactionLedger
          refreshTrigger={ledgerRefreshKey}
          onTransactionDeleted={(updatedPortfolio) => {
            setData(updatedPortfolio);
            setToastMessage('Transaction deleted and projections recomputed!');
            setTimeout(() => setToastMessage(null), 4000);
          }}
        />
      )}

      <AddTransactionModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={(updatedPortfolio) => {
          setData(updatedPortfolio);
          setLedgerRefreshKey((k) => k + 1);
          setToastMessage('Transaction recorded and projections updated!');
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
