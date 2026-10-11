import React, { useState } from 'react';
import { CashFlowReport } from './CashFlowReport';
import './ReportsView.css';

export type ReportType = 'cash-flow' | 'tax-lots' | 'dividends' | 'owner-earnings';

interface ReportsViewProps {
  preferredCurrency?: string;
}

export const ReportsView: React.FC<ReportsViewProps> = ({ preferredCurrency }) => {
  const [activeReport, setActiveReport] = useState<ReportType>('cash-flow');

  return (
    <div className="reports-view">
      <div className="reports-header">
        <div className="reports-title-group">
          <h2>Portfolio Reports</h2>
          <p>Multi-dimensional financial analysis, tax reporting, and liquidity reconciliation.</p>
        </div>

        <div className="reports-nav-pills">
          <button
            type="button"
            className={`reports-nav-btn ${activeReport === 'cash-flow' ? 'active' : ''}`}
            onClick={() => setActiveReport('cash-flow')}
          >
            <span>💵 Cash Flow Report</span>
          </button>

          <button
            type="button"
            className="reports-nav-btn disabled"
            title="Tax Lots & Capital Gains Report (Coming Soon in Feature #22)"
            disabled
          >
            <span>📜 Tax Lots & CGT</span>
            <span className="report-soon-badge">SOON</span>
          </button>

          <button
            type="button"
            className="reports-nav-btn disabled"
            title="Dividend Calendar & Yield Analytics (Coming Soon in Feature #23)"
            disabled
          >
            <span>📅 Dividend Calendar</span>
            <span className="report-soon-badge">SOON</span>
          </button>

          <button
            type="button"
            className="reports-nav-btn disabled"
            title="Look-Through Fundamental & Owner Earnings (Coming Soon in Feature #15)"
            disabled
          >
            <span>🏢 Owner Earnings</span>
            <span className="report-soon-badge">SOON</span>
          </button>
        </div>
      </div>

      <div className="reports-content">
        {activeReport === 'cash-flow' && (
          <CashFlowReport preferredCurrency={preferredCurrency} />
        )}
      </div>
    </div>
  );
};
