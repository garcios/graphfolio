import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import './UserPreferencesModal.css';

export interface UserPreferencesData {
  userId: string;
  email: string;
  displayName: string;
  displayCurrency: string;
  theme: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface CurrencyItem {
  code: string;
  name: string;
  symbol: string;
}

interface UserPreferencesModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentPreferences?: UserPreferencesData | null;
  supportedCurrencies?: CurrencyItem[];
  onSuccess: (updatedPrefs: UserPreferencesData, updatedPortfolio?: any) => void;
}

const CURRENCY_FLAGS: Record<string, string> = {
  USD: '🇺🇸',
  EUR: '🇪🇺',
  GBP: '🇬🇧',
  AUD: '🇦🇺',
  CAD: '🇨🇦',
  JPY: '🇯🇵',
  CHF: '🇨🇭',
};

export const UserPreferencesModal: React.FC<UserPreferencesModalProps> = ({
  isOpen,
  onClose,
  currentPreferences,
  supportedCurrencies = [],
  onSuccess,
}) => {
  const [displayName, setDisplayName] = useState('');
  const [displayCurrency, setDisplayCurrency] = useState('USD');
  const [theme, setTheme] = useState('DARK');
  const [currencies, setCurrencies] = useState<CurrencyItem[]>(supportedCurrencies);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Sync state when preferences or modal state changes
  useEffect(() => {
    if (currentPreferences) {
      setDisplayName(currentPreferences.displayName || '');
      setDisplayCurrency(currentPreferences.displayCurrency || 'USD');
      setTheme(currentPreferences.theme || 'DARK');
    }
    setError(null);
  }, [currentPreferences, isOpen]);

  // Load currencies if not provided
  useEffect(() => {
    if (currencies.length === 0 && isOpen) {
      client
        .query({
          supportedCurrencies: {
            code: true,
            name: true,
            symbol: true,
          },
        })
        .then((res) => {
          if (res.supportedCurrencies) {
            setCurrencies(res.supportedCurrencies);
          }
        })
        .catch((err) => {
          console.warn('Failed to fetch supported currencies:', err);
        });
    }
  }, [isOpen, currencies.length]);

  // Keyboard navigation & body scroll lock
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !saving) {
        onClose();
      }
    };

    document.body.style.overflow = 'hidden';
    window.addEventListener('keydown', handleKeyDown);

    return () => {
      document.body.style.overflow = '';
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, saving, onClose]);

  if (!isOpen) return null;

  const getInitials = (name: string): string => {
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase() || 'OG';
  };

  const selectedCurrencyInfo = currencies.find((c) => c.code === displayCurrency) || {
    code: displayCurrency,
    name: displayCurrency,
    symbol: '$',
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmedName = displayName.trim();
    if (!trimmedName) {
      setError('Display name cannot be empty');
      return;
    }
    if (trimmedName.length > 100) {
      setError('Display name must not exceed 100 characters');
      return;
    }

    setSaving(true);
    setError(null);

    try {
      const res = await client.mutation({
        updateUserPreferences: {
          __args: {
            input: {
              displayName: trimmedName,
              displayCurrency,
              theme,
            },
          },
          preferences: {
            userId: true,
            email: true,
            displayName: true,
            displayCurrency: true,
            theme: true,
            createdAt: true,
            updatedAt: true,
          },
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
            },
          },
        },
      });

      if (res.updateUserPreferences?.preferences) {
        onSuccess(
          res.updateUserPreferences.preferences,
          res.updateUserPreferences.portfolio
        );
        onClose();
      }
    } catch (err: any) {
      console.error('Failed to save user preferences:', err);
      setError(err?.message || 'Failed to save preferences. Please try again.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      className="prefs-modal-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget && !saving) {
          onClose();
        }
      }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="prefs-modal-title"
    >
      <div className="prefs-modal-container">
        <div className="prefs-modal-header">
          <h2 id="prefs-modal-title" className="prefs-modal-title">
            Investor Preferences
          </h2>
          <button
            type="button"
            className="prefs-close-btn"
            onClick={onClose}
            disabled={saving}
            aria-label="Close dialog"
          >
            &times;
          </button>
        </div>

        <form onSubmit={handleSubmit} className="prefs-modal-form">
          {/* User Info Banner */}
          <div className="prefs-user-banner">
            <div className="prefs-avatar">
              {getInitials(displayName || currentPreferences?.displayName || 'OG')}
            </div>
            <div className="prefs-user-meta">
              <span className="prefs-user-name">
                {displayName || currentPreferences?.displayName || 'Oscar Garcia'}
              </span>
              <span className="prefs-user-email">
                {currentPreferences?.email || 'demo@graphfolio.internal'}
              </span>
            </div>
          </div>

          {error && <div className="prefs-error-banner">{error}</div>}

          {/* Display Name Input */}
          <div className="prefs-form-group">
            <div className="prefs-label-row">
              <label htmlFor="prefs-display-name" className="prefs-label">
                Display Name
              </label>
              <span className="prefs-char-count">{displayName.length}/100</span>
            </div>
            <input
              id="prefs-display-name"
              type="text"
              className="prefs-input"
              value={displayName}
              maxLength={100}
              placeholder="e.g. Oscar Garcia"
              onChange={(e) => setDisplayName(e.target.value)}
              disabled={saving}
              required
            />
          </div>

          {/* Display Currency Select */}
          <div className="prefs-form-group">
            <label htmlFor="prefs-currency" className="prefs-label">
              Display Currency
            </label>
            <span className="prefs-subtitle">
              Anchors portfolio valuations, cash balances, and performance metrics.
            </span>
            <select
              id="prefs-currency"
              className="prefs-select"
              value={displayCurrency}
              onChange={(e) => setDisplayCurrency(e.target.value)}
              disabled={saving}
            >
              {currencies.length > 0 ? (
                currencies.map((c) => {
                  const flag = CURRENCY_FLAGS[c.code] || '🌐';
                  return (
                    <option key={c.code} value={c.code}>
                      {flag} {c.code} ({c.symbol}) — {c.name}
                    </option>
                  );
                })
              ) : (
                <option value="USD">🇺🇸 USD ($) — US Dollar</option>
              )}
            </select>

            <div className="prefs-currency-preview">
              <span>Sample formatted output:</span>
              <span className="prefs-currency-preview-tag">
                {selectedCurrencyInfo.symbol}12,345.67 {selectedCurrencyInfo.code}
              </span>
            </div>
          </div>

          {/* Interface Theme Segmented Control */}
          <div className="prefs-form-group">
            <label className="prefs-label">Interface Theme</label>
            <div className="prefs-theme-picker">
              <button
                type="button"
                className={`prefs-theme-btn ${theme === 'DARK' ? 'active' : ''}`}
                onClick={() => setTheme('DARK')}
                disabled={saving}
              >
                <span>🌙</span> Dark
              </button>
              <button
                type="button"
                className={`prefs-theme-btn ${theme === 'LIGHT' ? 'active' : ''}`}
                onClick={() => setTheme('LIGHT')}
                disabled={saving}
              >
                <span>☀️</span> Light
              </button>
              <button
                type="button"
                className={`prefs-theme-btn ${theme === 'SYSTEM' ? 'active' : ''}`}
                onClick={() => setTheme('SYSTEM')}
                disabled={saving}
              >
                <span>💻</span> System
              </button>
            </div>
          </div>

          <div className="prefs-modal-footer">
            <button
              type="button"
              className="prefs-btn-cancel"
              onClick={onClose}
              disabled={saving}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="prefs-btn-save"
              disabled={saving || !displayName.trim()}
            >
              {saving && <span className="prefs-spinner" />}
              {saving ? 'Saving...' : 'Save Preferences'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
