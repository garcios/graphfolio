import React from 'react';
import { Badge } from '@graphfolio/ui';
import './AdminLayout.css';

export type AdminTab = 'assets' | 'prices' | 'ingestion';

interface AdminLayoutProps {
  currentTab: AdminTab;
  onTabChange: (tab: AdminTab) => void;
  children: React.ReactNode;
}

export const AdminLayout: React.FC<AdminLayoutProps> = ({
  currentTab,
  onTabChange,
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
            <span>📊</span>
            <span>Asset Management</span>
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'prices' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('prices')}
          >
            <span>📈</span>
            <span>Market Prices</span>
          </button>

          <button
            type="button"
            className={`admin-nav-item ${currentTab === 'ingestion' ? 'admin-nav-item--active' : ''}`}
            onClick={() => onTabChange('ingestion')}
          >
            <span>⚡</span>
            <span>Ingestion Pipeline</span>
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
