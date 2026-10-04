import { useEffect, useState } from 'react';
import { createClient } from '../generated';
import './Dashboard.css';
import chartImage from '../assets/portfolio_chart.jpg';

const client = createClient({
  url: 'http://localhost:8080/query',
});

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
        <div className="user-profile">
          <div className="avatar">OG</div>
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

      <section className="chart-container">
        <div className="chart-header">
          <h3>Performance</h3>
          <div className="time-filters">
            <button>1D</button>
            <button>1W</button>
            <button>1M</button>
            <button className="active">1Y</button>
            <button>ALL</button>
          </div>
        </div>
        <div className="chart-area">
          <img src={chartImage} alt="Portfolio Performance Chart" className="mock-chart-img" />
          <div className="chart-overlay-gradient"></div>
        </div>
      </section>

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
    </div>
  );
};
