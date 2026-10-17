import React, { useState } from 'react';
import { useAdminAuth } from '../context/AdminAuthContext';
import { Card, Button, Input } from '@graphfolio/ui';
import './AdminLoginView.css';

export const AdminLoginView: React.FC = () => {
  const { login, isLoading, error, clearError } = useAdminAuth();
  const [email, setEmail] = useState('admin@graphfolio.app');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);
    clearError();

    const trimmedEmail = email.trim();
    if (!trimmedEmail) {
      setValidationError('Please enter your administrator email.');
      return;
    }
    if (!password) {
      setValidationError('Please enter your password.');
      return;
    }

    try {
      await login(trimmedEmail, password);
    } catch {
      // Error is stored in AdminAuthContext
    }
  };

  const displayError = validationError || error;

  return (
    <div className="admin-login-container">
      <div className="admin-login-glow" />
      <Card className="admin-login-card">
        <div className="admin-login-header">
          <div className="admin-logo-row">
            <span className="admin-logo-icon">📊</span>
            <span className="admin-brand-title">GraphFolio</span>
            <span className="admin-badge-pill">ADMIN</span>
          </div>
          <h1 className="admin-login-title">Administrator Portal</h1>
          <p className="admin-login-subtitle">
            Restricted Operations, Market Ingestion & Valuations Console
          </p>
        </div>

        {displayError && (
          <div className="admin-login-alert" role="alert">
            <span className="admin-alert-icon">⚠️</span>
            <span className="admin-alert-text">{displayError}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="admin-login-form">
          <div className="admin-field-group">
            <label htmlFor="admin-email" className="admin-label">
              Administrator Email
            </label>
            <Input
              id="admin-email"
              type="email"
              autoComplete="username"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
                if (validationError) setValidationError(null);
              }}
              placeholder="admin@graphfolio.app"
              required
              disabled={isLoading}
            />
          </div>

          <div className="admin-field-group">
            <div className="admin-label-row">
              <label htmlFor="admin-password" className="admin-label">
                Password
              </label>
              <button
                type="button"
                className="admin-toggle-password"
                onClick={() => setShowPassword(!showPassword)}
                tabIndex={-1}
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
            <Input
              id="admin-password"
              type={showPassword ? 'text' : 'password'}
              autoComplete="current-password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                if (validationError) setValidationError(null);
              }}
              placeholder="Enter password"
              required
              disabled={isLoading}
            />
          </div>

          <Button
            type="submit"
            variant="primary"
            className="admin-submit-btn"
            disabled={isLoading}
          >
            {isLoading ? (
              <span className="admin-btn-content">
                <span className="admin-spinner" />
                <span>Authenticating...</span>
              </span>
            ) : (
              'Sign In to Admin Portal'
            )}
          </Button>
        </form>

        <div className="admin-login-footer">
          <p className="admin-security-note">
            🔒 Authorized personnel only. All access attempts are monitored and logged.
          </p>
        </div>
      </Card>
    </div>
  );
};
