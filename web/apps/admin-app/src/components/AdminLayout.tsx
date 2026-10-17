import React, { useState, useRef, useEffect } from 'react';
import { Badge, Modal } from '@graphfolio/ui';
import { useAdminAuth } from '../context/AdminAuthContext';
import { AdminChangePasswordView } from './AdminChangePasswordView';
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
  const { user, logout } = useAdminAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <div className="admin-sidebar__brand">
          <div className="admin-sidebar__logo">
            GraphFolio <span>Admin</span>
          </div>
          <Badge variant="info">PORTAL</Badge>
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
            <div className="admin-profile-menu-container" ref={menuRef}>
              <button
                type="button"
                className="admin-profile-trigger"
                onClick={() => setMenuOpen(!menuOpen)}
                aria-expanded={menuOpen}
              >
                <span className="admin-avatar">👤</span>
                <span className="admin-email-text">{user?.email || 'admin@graphfolio.app'}</span>
                <span className="admin-role-badge">ADMIN</span>
                <span className="admin-dropdown-arrow">▾</span>
              </button>

              {menuOpen && (
                <div className="admin-profile-dropdown">
                  <div className="admin-dropdown-header">
                    <span className="admin-dropdown-name">{user?.displayName || 'System Administrator'}</span>
                    <span className="admin-dropdown-email">{user?.email}</span>
                  </div>
                  <div className="admin-dropdown-divider" />
                  <button
                    type="button"
                    className="admin-dropdown-item"
                    onClick={() => {
                      setMenuOpen(false);
                      setShowPasswordModal(true);
                    }}
                  >
                    <span>🔑</span>
                    <span>Change Password</span>
                  </button>
                  <div className="admin-dropdown-divider" />
                  <button
                    type="button"
                    className="admin-dropdown-item admin-dropdown-item--danger"
                    onClick={() => {
                      setMenuOpen(false);
                      logout();
                    }}
                  >
                    <span>🚪</span>
                    <span>Sign Out</span>
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>

        <div className="admin-content">{children}</div>
      </main>

      {showPasswordModal && (
        <Modal
          isOpen={showPasswordModal}
          onClose={() => setShowPasswordModal(false)}
          title="Change Administrator Password"
        >
          <AdminChangePasswordView
            isModal={true}
            onSuccess={() => setShowPasswordModal(false)}
            onCancel={() => setShowPasswordModal(false)}
          />
        </Modal>
      )}
    </div>
  );
};
