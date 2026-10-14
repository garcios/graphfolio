import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { Button } from '@graphfolio/ui';
import { client } from '@graphfolio/api-client';
import { PasswordStrengthMeter, evaluatePassword } from './PasswordStrengthMeter';
import './AuthForms.css';

interface RegisterFormProps {
  onSwitchToLogin: () => void;
}

const DEFAULT_CURRENCIES = [
  { code: 'USD', name: 'US Dollar', symbol: '$', flag: '🇺🇸' },
  { code: 'AUD', name: 'Australian Dollar', symbol: 'A$', flag: '🇦🇺' },
  { code: 'EUR', name: 'Euro', symbol: '€', flag: '🇪🇺' },
  { code: 'GBP', name: 'British Pound', symbol: '£', flag: '🇬🇧' },
  { code: 'CAD', name: 'Canadian Dollar', symbol: 'C$', flag: '🇨🇦' },
  { code: 'JPY', name: 'Japanese Yen', symbol: '¥', flag: '🇯🇵' },
  { code: 'CHF', name: 'Swiss Franc', symbol: 'CHF', flag: '🇨🇭' },
];

export const RegisterForm: React.FC<RegisterFormProps> = ({ onSwitchToLogin }) => {
  const { register, error, clearError } = useAuth();
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [displayCurrency, setDisplayCurrency] = useState('USD');
  const [currencies, setCurrencies] = useState(DEFAULT_CURRENCIES);
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  // Load supported currencies from backend if available
  useEffect(() => {
    client
      .query({
        supportedCurrencies: {
          code: true,
          name: true,
          symbol: true,
        },
      })
      .then(res => {
        if (res.supportedCurrencies && res.supportedCurrencies.length > 0) {
          const mapped = res.supportedCurrencies.map(c => {
            const match = DEFAULT_CURRENCIES.find(d => d.code === c.code);
            return {
              code: c.code,
              name: c.name,
              symbol: c.symbol,
              flag: match ? match.flag : '🌐',
            };
          });
          setCurrencies(mapped);
        }
      })
      .catch(() => {
        // Fallback to DEFAULT_CURRENCIES
      });
  }, []);

  const passwordEvaluation = evaluatePassword(password);
  const passwordsMatch = password && confirmPassword && password === confirmPassword;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    clearError();

    if (!displayName.trim()) {
      setFormError('Please enter your name.');
      return;
    }

    if (!email.trim()) {
      setFormError('Please enter a valid email address.');
      return;
    }

    if (!passwordEvaluation.isValid) {
      setFormError('Please ensure your password satisfies all security requirements.');
      return;
    }

    if (password !== confirmPassword) {
      setFormError('Passwords do not match.');
      return;
    }

    try {
      setIsSubmitting(true);
      await register(email.trim(), password, displayName.trim(), displayCurrency);
    } catch (err: any) {
      console.warn('Registration submission error:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const activeError = formError || error;

  return (
    <form className="auth-form" onSubmit={handleSubmit} noValidate>
      <div className="auth-form-header">
        <h2 className="auth-form-title">Create Account</h2>
        <p className="auth-form-subtitle">Start tracking your true annualized portfolio returns</p>
      </div>

      {activeError && (
        <div className="auth-error-banner" role="alert">
          <span className="error-icon">⚠️</span>
          <span className="error-message">{activeError}</span>
        </div>
      )}

      <div className="form-group">
        <label htmlFor="reg-name">Full Name</label>
        <input
          id="reg-name"
          type="text"
          className="auth-input"
          placeholder="e.g. Jane Doe"
          value={displayName}
          onChange={e => setDisplayName(e.target.value)}
          autoComplete="name"
          required
        />
      </div>

      <div className="form-group">
        <label htmlFor="reg-email">Email Address</label>
        <input
          id="reg-email"
          type="email"
          className="auth-input"
          placeholder="name@example.com"
          value={email}
          onChange={e => setEmail(e.target.value)}
          autoComplete="email"
          required
        />
      </div>

      <div className="form-group">
        <label htmlFor="reg-currency">Base Portfolio Currency</label>
        <div className="select-wrapper">
          <select
            id="reg-currency"
            className="auth-input auth-select"
            value={displayCurrency}
            onChange={e => setDisplayCurrency(e.target.value)}
          >
            {currencies.map(c => (
              <option key={c.code} value={c.code}>
                {c.flag} {c.code} - {c.name} ({c.symbol})
              </option>
            ))}
          </select>
        </div>
        <span className="form-hint">
          Your initial default portfolio and returns will be anchored in this currency.
        </span>
      </div>

      <div className="form-group">
        <div className="label-row">
          <label htmlFor="reg-password">Password</label>
        </div>
        <div className="password-input-wrapper">
          <input
            id="reg-password"
            type={showPassword ? 'text' : 'password'}
            className="auth-input password-input"
            placeholder="Create a strong password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            autoComplete="new-password"
            required
          />
          <button
            type="button"
            className="toggle-password-btn"
            onClick={() => setShowPassword(!showPassword)}
            aria-label={showPassword ? 'Hide password' : 'Show password'}
          >
            {showPassword ? '👁️' : '👁️‍🗨️'}
          </button>
        </div>
        <PasswordStrengthMeter password={password} />
      </div>

      <div className="form-group">
        <label htmlFor="reg-confirm-password">Confirm Password</label>
        <input
          id="reg-confirm-password"
          type={showPassword ? 'text' : 'password'}
          className={`auth-input ${
            confirmPassword && !passwordsMatch ? 'input-error' : ''
          }`}
          placeholder="Repeat your password"
          value={confirmPassword}
          onChange={e => setConfirmPassword(e.target.value)}
          autoComplete="new-password"
          required
        />
        {confirmPassword && !passwordsMatch && (
          <span className="field-error-text">Passwords do not match</span>
        )}
      </div>

      <Button
        type="submit"
        variant="primary"
        className="auth-submit-btn"
        disabled={isSubmitting || !passwordEvaluation.isValid || !passwordsMatch}
      >
        {isSubmitting ? 'Creating Account...' : 'Create Account'}
      </Button>

      <div className="auth-switch-prompt">
        <span>Already have an account?</span>{' '}
        <button
          type="button"
          className="switch-link-btn"
          onClick={onSwitchToLogin}
        >
          Sign In
        </button>
      </div>
    </form>
  );
};
