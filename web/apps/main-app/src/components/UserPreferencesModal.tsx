import React, { useState, useEffect, useMemo } from 'react';
import { client } from '@graphfolio/api-client';
import './UserPreferencesModal.css';

export interface UserPreferencesData {
  userId: string;
  email: string;
  displayName: string;
  displayCurrency: string;
  theme: string;
  financialYearEndMonth?: number;
  financialYearEndDay?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface CurrencyItem {
  code: string;
  name: string;
  symbol: string;
}

export interface EOFYPreset {
  id: string;
  label: string;
  flag: string;
  month: number;
  day: number;
}

export const EOFY_PRESETS: EOFYPreset[] = [
  { id: 'AU', label: 'Australia (30 Jun)', flag: '🇦🇺', month: 6, day: 30 },
  { id: 'UK', label: 'UK (5 Apr)', flag: '🇬🇧', month: 4, day: 5 },
  { id: 'US', label: 'USA / Canada (31 Dec)', flag: '🇺🇸', month: 12, day: 31 },
  { id: 'NZ', label: 'NZ / India (31 Mar)', flag: '🇳🇿', month: 3, day: 31 },
  { id: 'ZA', label: 'South Africa (28 Feb)', flag: '🇿🇦', month: 2, day: 28 },
];

export const MONTH_OPTIONS = [
  { value: 1, label: 'January (01)' },
  { value: 2, label: 'February (02)' },
  { value: 3, label: 'March (03)' },
  { value: 4, label: 'April (04)' },
  { value: 5, label: 'May (05)' },
  { value: 6, label: 'June (06)' },
  { value: 7, label: 'July (07)' },
  { value: 8, label: 'August (08)' },
  { value: 9, label: 'September (09)' },
  { value: 10, label: 'October (10)' },
  { value: 11, label: 'November (11)' },
  { value: 12, label: 'December (12)' },
];

export function getMaxDaysInMonth(month: number): number {
  if (month === 2) return 29; // Allows leap-year enders (e.g. Feb 29)
  if ([4, 6, 9, 11].includes(month)) return 30;
  return 31;
}

export function detectPresetId(month: number, day: number): string {
  const match = EOFY_PRESETS.find((p) => p.month === month && p.day === day);
  return match ? match.id : 'CUSTOM';
}

export function isLastDayOfMonth(month: number, day: number): boolean {
  if (month === 2) return day >= 28;
  if ([4, 6, 9, 11].includes(month)) return day === 30;
  return day === 31;
}

export function calculateFiscalYearBounds(
  month: number,
  day: number,
  asOfDate: Date = new Date()
) {
  const currentYear = asOfDate.getFullYear();
  const isLeap = (year: number) => (year % 4 === 0 && year % 100 !== 0) || (year % 400 === 0);
  const resolveEndDay = (year: number, m: number, d: number) => {
    if (m === 2) {
      if (d >= 28) return isLeap(year) ? 29 : 28;
      return d;
    }
    return d;
  };

  const endThisYear = new Date(
    currentYear,
    month - 1,
    resolveEndDay(currentYear, month, day),
    23,
    59,
    59,
    999
  );
  let fyEndYear: number;

  if (asOfDate.getTime() <= endThisYear.getTime()) {
    fyEndYear = currentYear;
  } else {
    fyEndYear = currentYear + 1;
  }

  const buildBounds = (endYear: number) => {
    const endDay = resolveEndDay(endYear, month, day);
    const endDate = new Date(endYear, month - 1, endDay);
    let startDate: Date;

    if (isLastDayOfMonth(month, day)) {
      if (month === 12) {
        startDate = new Date(endYear, 0, 1);
      } else {
        startDate = new Date(endYear - 1, month, 1);
      }
    } else {
      startDate = new Date(endYear - 1, month - 1, day + 1);
    }

    const label =
      month === 12 && day === 31
        ? `FY${endYear}`
        : `FY${String(endYear).slice(-2)}`;

    return { startDate, endDate, label };
  };

  const current = buildBounds(fyEndYear);
  const prev = buildBounds(fyEndYear - 1);

  const formatDate = (d: Date) =>
    d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });

  return {
    currentPeriod: `${formatDate(current.startDate)} – ${formatDate(current.endDate)} (${current.label})`,
    prevPeriod: `${formatDate(prev.startDate)} – ${formatDate(prev.endDate)} (${prev.label})`,
    currentLabel: current.label,
    prevLabel: prev.label,
    startDate: current.startDate,
    endDate: current.endDate,
  };
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
  const [fyMonth, setFyMonth] = useState(6);
  const [fyDay, setFyDay] = useState(30);
  const [currencies, setCurrencies] = useState<CurrencyItem[]>(supportedCurrencies);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Sync state when preferences or modal state changes
  useEffect(() => {
    if (currentPreferences) {
      setDisplayName(currentPreferences.displayName || '');
      setDisplayCurrency(currentPreferences.displayCurrency || 'USD');
      setTheme(currentPreferences.theme || 'DARK');
      setFyMonth(currentPreferences.financialYearEndMonth ?? 6);
      setFyDay(currentPreferences.financialYearEndDay ?? 30);
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

  const activePresetId = useMemo(() => detectPresetId(fyMonth, fyDay), [fyMonth, fyDay]);

  const fiscalYearPreview = useMemo(
    () => calculateFiscalYearBounds(fyMonth, fyDay),
    [fyMonth, fyDay]
  );

  const availableDays = useMemo(() => {
    const max = getMaxDaysInMonth(fyMonth);
    return Array.from({ length: max }, (_, i) => i + 1);
  }, [fyMonth]);

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

  const handleMonthChange = (newMonth: number) => {
    setFyMonth(newMonth);
    const max = getMaxDaysInMonth(newMonth);
    if (fyDay > max) {
      setFyDay(max);
    }
  };

  const handlePresetSelect = (preset: EOFYPreset) => {
    setFyMonth(preset.month);
    setFyDay(preset.day);
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
              financialYearEndMonth: fyMonth,
              financialYearEndDay: fyDay,
            },
          },
          preferences: {
            userId: true,
            email: true,
            displayName: true,
            displayCurrency: true,
            theme: true,
            financialYearEndMonth: true,
            financialYearEndDay: true,
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
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Your preferred name or alias"
              disabled={saving}
              required
            />
          </div>

          {/* Display Currency Select */}
          <div className="prefs-form-group">
            <div className="prefs-label-row">
              <label htmlFor="prefs-currency-select" className="prefs-label">
                Reporting & Display Currency
              </label>
              <span className="prefs-hint">Converts portfolio values & returns</span>
            </div>
            <select
              id="prefs-currency-select"
              className="prefs-select"
              value={displayCurrency}
              onChange={(e) => setDisplayCurrency(e.target.value)}
              disabled={saving}
            >
              {currencies.length > 0 ? (
                currencies.map((curr) => {
                  const flag = CURRENCY_FLAGS[curr.code] || '🌐';
                  return (
                    <option key={curr.code} value={curr.code}>
                      {flag} {curr.code} ({curr.symbol}) — {curr.name}
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

          <div className="prefs-section-divider" />

          {/* End of Financial Year (EOFY) Section */}
          <div className="prefs-form-group">
            <div className="prefs-label-row">
              <label className="prefs-label">Financial Year End (EOFY)</label>
              <span className="prefs-hint">Tax year cutoff for statements & FYTD metrics</span>
            </div>

            {/* Regional Presets */}
            <div className="prefs-eofy-presets">
              {EOFY_PRESETS.map((preset) => {
                const isActive = activePresetId === preset.id;
                return (
                  <button
                    key={preset.id}
                    type="button"
                    className={`prefs-preset-pill ${isActive ? 'active' : ''}`}
                    onClick={() => handlePresetSelect(preset)}
                    disabled={saving}
                  >
                    <span>{preset.flag}</span> {preset.label}
                  </button>
                );
              })}
              <button
                type="button"
                className={`prefs-preset-pill ${activePresetId === 'CUSTOM' ? 'active' : ''}`}
                onClick={() => {
                  // Switch to custom (retaining current month & day)
                }}
                disabled={saving}
              >
                <span>🌐</span> Custom
              </button>
            </div>

            {/* Month & Day Dropdown Pickers */}
            <div className="prefs-eofy-pickers">
              <div className="prefs-picker-col">
                <label htmlFor="prefs-eofy-month" className="prefs-sublabel">
                  Ending Month
                </label>
                <select
                  id="prefs-eofy-month"
                  className="prefs-select"
                  value={fyMonth}
                  onChange={(e) => handleMonthChange(Number(e.target.value))}
                  disabled={saving}
                >
                  {MONTH_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="prefs-picker-col">
                <label htmlFor="prefs-eofy-day" className="prefs-sublabel">
                  Ending Day
                </label>
                <select
                  id="prefs-eofy-day"
                  className="prefs-select"
                  value={fyDay}
                  onChange={(e) => setFyDay(Number(e.target.value))}
                  disabled={saving}
                >
                  {availableDays.map((d) => (
                    <option key={d} value={d}>
                      {d}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Live Fiscal Year Preview Card */}
            <div className="prefs-eofy-card">
              <div className="prefs-eofy-card-header">
                <span className="prefs-eofy-badge">Fiscal Calendar Preview</span>
              </div>
              <div className="prefs-eofy-card-body">
                <div className="prefs-eofy-row">
                  <span className="prefs-eofy-label">Current Fiscal Period:</span>
                  <span className="prefs-eofy-value current">{fiscalYearPreview.currentPeriod}</span>
                </div>
                <div className="prefs-eofy-row">
                  <span className="prefs-eofy-label">Previous Fiscal Period:</span>
                  <span className="prefs-eofy-value">{fiscalYearPreview.prevPeriod}</span>
                </div>
              </div>
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
