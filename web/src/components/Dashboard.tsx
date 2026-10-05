import { useEffect, useState } from 'react';
import { client } from '../graphql/client';
import './Dashboard.css';
import { AddTransactionModal } from './AddTransactionModal';
import { PerformanceChart } from './PerformanceChart';
import { TransactionLedger } from './TransactionLedger';

interface Money {
  amount: string;
  currencyCode: string;
}

const formatMoney = (m?: Money | null) => {
  if (!m) return '$0.00';
  const n = Number(m.amount);
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: m.currencyCode || 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n);
};

const formatPercent = (percentStr?: string | null, decimals = 2) => {
  if (!percentStr) return '0.00%';
  const n = Number(percentStr);
  const sign = n >= 0 ? '+' : '';
  return `${sign}${n.toFixed(decimals)}%`;
};

const isPositive = (val?: Money | string | null) => {
  if (!val) return false;
  const raw = typeof val === 'string' ? val : val.amount;
  return Number(raw) >= 0;
};

export const Dashboard = () => {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'ledger'>('overview');
  const [ledgerRefreshKey, setLedgerRefreshKey] = useState(0);

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
      }
    })
    .then(res => {
      setData(res.portfolio);
      setLoading(false);
    })
    .catch(err => {
      console.error("Error fetching portfolio:", err);
      setLoading(false);
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
          <button
            type="button"
            className="btn-add-transaction"
            onClick={() => setIsModalOpen(true)}
          >
            + Add Transaction
          </button>
          <div className="user-profile">
            <div className="avatar">OG</div>
          </div>
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
          <PerformanceChart />

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

      {toastMessage && (
        <div className="toast-notification">
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
};

