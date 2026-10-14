import React from 'react';
import { Badge } from '@graphfolio/ui';
import './AdminLayout.css';

export type AdminTab = 'assets' | 'prices' | 'missing-prices' | 'missing-fx' | 'fx' | 'ingestion';

interface AdminLayoutProps {
  currentTab: AdminTab;
  onTabChange: (tab: AdminTab) => void;
  missingPriceCount?: number;
  missingFXCount?: number;
  children: React.ReactNode;
}

export const AdminLayout: React.FC<AdminLayoutProps> = ({
  currentTab,
  onTabChange,
  missingPriceCount,
  missingFXCount,
  children,
}) => {

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <div className="admin-sidebar__brand">
          <div className="admin-sidebar__logo">
            GraphFolio <span>Admin</span>
          </div>
          <Badge variant="info">LOCAL</Badge>
        </div>

        <nav className="admin-sidebar__nav">
          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'assets' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('assets')}
          >
            <span className="admin-nav-item__icon">📊</span>
            <span className="admin-nav-item__label">Asset Management</span>
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'prices' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('prices')}
          >
            <span className="admin-nav-item__icon">📈</span>
            <span className="admin-nav-item__label">Market Prices</span>
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'missing-prices' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('missing-prices')}
          >
            <span className="admin-nav-item__icon">⚠️</span>
            <span className="admin-nav-item__label">Missing Prices</span>
            {typeof missingPriceCount === 'number' && missingPriceCount > 0 && (
              <Badge variant="warning">{missingPriceCount}</Badge>
            )}
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'missing-fx' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('missing-fx')}
          >
            <span className="admin-nav-item__icon admin-nav-item__icon--composite">
              <span className="admin-nav-item__icon-base">💱</span>
              <span className="admin-nav-item__icon-badge">⚠️</span>
            </span>
            <span className="admin-nav-item__label">Missing FX Rates</span>
            {typeof missingFXCount === 'number' && missingFXCount > 0 && (
              <Badge variant="warning">{missingFXCount}</Badge>
            )}
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'fx' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('fx')}
          >
            <span className="admin-nav-item__icon">💱</span>
            <span className="admin-nav-item__label">Currency Pairs (FX)</span>
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'ingestion' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('ingestion')}
          >
            <span className="admin-nav-item__icon">⚡</span>
            <span className="admin-nav-item__label">Ingestion Pipeline</span>
          </button>
        </nav>

        <div className="admin-sidebar__footer">
          <div className="system-status">
            <div className="status-indicator" />
            <span>BFF GraphQL (:8080) Online</span>
          </div>
          <div className="system-status">
            <div className="status-indicator" />
            <span>PostgreSQL (portfolio) Ready</span>
          </div>
        </div>
      </aside>

      <main className="admin-main">
        <header className="admin-topbar">
          <h2 className="admin-topbar__title">
            {currentTab === 'assets' && 'Tradable Asset & Instrument Directory'}
            {currentTab === 'prices' && 'Market Data, Closing Prices & Overrides'}
            {currentTab === 'missing-prices' && 'Diagnostic Report: Unpriced Transactions on Trade Date'}
            {currentTab === 'missing-fx' && 'Diagnostic Report: Unconverted Transactions Lacking Trade-Date FX'}
            {currentTab === 'fx' && 'Foreign Exchange (FX) & Currency Pair History'}
            {currentTab === 'ingestion' && 'Market Ingestion Pipeline & FX Monitor'}
          </h2>

          <div className="admin-topbar__actions">
            <Badge variant="purple">Internal Portal</Badge>
          </div>
        </header>

        <div className="admin-content">{children}</div>
      </main>
    </div>
  );
};
