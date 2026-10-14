import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { Button } from '@graphfolio/ui';
import './AuthForms.css';

interface LoginFormProps {
  onSwitchToRegister: () => void;
}

export const LoginForm: React.FC<LoginFormProps> = ({ onSwitchToRegister }) => {
  const { login, error, clearError } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    clearError();

    if (!email.trim() || !password) {
      setFormError('Please enter both email and password.');
      return;
    }

    try {
      setIsSubmitting(true);
      await login(email.trim(), password, rememberMe);
    } catch (err: any) {
      // Error handled by AuthContext state
      console.warn('Login submission error:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const activeError = formError || error;

  return (
    <form className="auth-form" onSubmit={handleSubmit} noValidate>
      <div className="auth-form-header">
        <h2 className="auth-form-title">Welcome Back</h2>
        <p className="auth-form-subtitle">Sign in to manage your investment portfolio</p>
      </div>

      {activeError && (
        <div className="auth-error-banner" role="alert">
          <span className="error-icon">⚠️</span>
          <span className="error-message">{activeError}</span>
        </div>
      )}

      <div className="form-group">
        <label htmlFor="login-email">Email Address</label>
        <input
          id="login-email"
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
        <div className="label-row">
          <label htmlFor="login-password">Password</label>
        </div>
        <div className="password-input-wrapper">
          <input
            id="login-password"
            type={showPassword ? 'text' : 'password'}
            className="auth-input password-input"
            placeholder="••••••••••••"
            value={password}
            onChange={e => setPassword(e.target.value)}
            autoComplete="current-password"
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
      </div>

      <div className="form-options-row">
        <label className="remember-me-label">
          <input
            type="checkbox"
            checked={rememberMe}
            onChange={e => setRememberMe(e.target.checked)}
            className="remember-me-checkbox"
          />
          <span>Remember me for 30 days</span>
        </label>
      </div>

      <Button
        type="submit"
        variant="primary"
        className="auth-submit-btn"
        disabled={isSubmitting}
      >
        {isSubmitting ? 'Signing In...' : 'Sign In'}
      </Button>

      <div className="auth-switch-prompt">
        <span>Don't have an account?</span>{' '}
        <button
          type="button"
          className="switch-link-btn"
          onClick={onSwitchToRegister}
        >
          Create an account
        </button>
      </div>
    </form>
  );
};
