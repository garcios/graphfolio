import React, { useState } from 'react';
import { LoginForm } from './LoginForm';
import { RegisterForm } from './RegisterForm';
import './AuthForms.css';

export const AuthView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'login' | 'register'>('login');

  return (
    <div className="auth-view-wrapper">
      <div className="auth-view-container">
        <div className="auth-brand-header">
          <div className="auth-logo">
            <span className="logo-icon">📈</span>
            <span className="logo-text">GraphFolio</span>
          </div>
          <p className="auth-tagline">
            Precision portfolio tracking for serious investors
          </p>
        </div>

        <div className="auth-card">
          <div className="auth-tabs" role="tablist">
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'login'}
              className={`auth-tab-btn ${activeTab === 'login' ? 'active' : ''}`}
              onClick={() => setActiveTab('login')}
            >
              Sign In
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'register'}
              className={`auth-tab-btn ${activeTab === 'register' ? 'active' : ''}`}
              onClick={() => setActiveTab('register')}
            >
              Create Account
            </button>
          </div>

          <div className="auth-card-body">
            {activeTab === 'login' ? (
              <LoginForm onSwitchToRegister={() => setActiveTab('register')} />
            ) : (
              <RegisterForm onSwitchToLogin={() => setActiveTab('login')} />
            )}
          </div>
        </div>

        <div className="auth-trust-badges">
          <div className="trust-badge">
            <span className="badge-icon">🔒</span>
            <span>Argon2id Encrypted</span>
          </div>
          <div className="trust-badge">
            <span className="badge-icon">📊</span>
            <span>GIPS-Compliant TWR & Modified Dietz</span>
          </div>
          <div className="trust-badge">
            <span className="badge-icon">⚡</span>
            <span>Multi-Currency Attribution</span>
          </div>
        </div>
      </div>
    </div>
  );
};
