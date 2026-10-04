import { useEffect, useState } from 'react';
import { createClient } from '../generated';
import './Dashboard.css';
import chartImage from '../assets/portfolio_chart.jpg';

const client = createClient({
  url: 'http://localhost:8080/query',
});

export const Dashboard = () => {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    client.query({
      portfolio: {
        totalValue: true,
        todayReturnAmount: true,
        todayReturnPercent: true,
        annualizedReturnPercent: true,
        cashBalance: true,
        investments: {
          id: true,
          ticker: true,
          name: true,
          price: true,
          quantity: true,
          totalValue: true,
          todayReturnAmount: true,
          todayReturnPercent: true,
          totalReturnAmount: true,
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
          <h2>${data.totalValue.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</h2>
          <span className={`trend ${data.todayReturnAmount >= 0 ? 'positive' : 'negative'}`}>
            {data.todayReturnAmount >= 0 ? '+' : ''}${Math.abs(data.todayReturnAmount).toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})} ({data.todayReturnAmount >= 0 ? '+' : ''}{data.todayReturnPercent.toFixed(2)}%) Today
          </span>
        </div>
        <div className="metric-secondary">
          <div className="metric-card">
            <span className="label">Annualized Return (TWR)</span>
            <span className={`value ${data.annualizedReturnPercent >= 0 ? 'positive' : 'negative'}`}>
              {data.annualizedReturnPercent >= 0 ? '+' : ''}{data.annualizedReturnPercent.toFixed(1)}%
            </span>
          </div>
          <div className="metric-card">
            <span className="label">Cash Balance</span>
            <span className="value neutral">${data.cashBalance.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</span>
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
                {data.investments.map((inv: any) => (
                  <tr key={inv.id}>
                    <td>
                      <div className="asset-info">
                        <span className="ticker">{inv.ticker}</span>
                        <span className="name">{inv.name}</span>
                      </div>
                    </td>
                    <td>${inv.price.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</td>
                    <td>{inv.quantity}</td>
                    <td>${inv.totalValue.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</td>
                    <td>
                      <div className={`return-info ${inv.todayReturnAmount >= 0 ? 'positive' : 'negative'}`}>
                        <span className="amount">{inv.todayReturnAmount >= 0 ? '+' : ''}${Math.abs(inv.todayReturnAmount).toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</span>
                        <span className="percent">{inv.todayReturnPercent >= 0 ? '+' : ''}{inv.todayReturnPercent.toFixed(2)}%</span>
                      </div>
                    </td>
                    <td>
                      <div className={`return-info ${inv.totalReturnAmount >= 0 ? 'positive' : 'negative'}`}>
                        <span className="amount">{inv.totalReturnAmount >= 0 ? '+' : ''}${Math.abs(inv.totalReturnAmount).toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2})}</span>
                        <span className="percent">{inv.totalReturnPercent >= 0 ? '+' : ''}{inv.totalReturnPercent.toFixed(1)}%</span>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </div>
  );
};
