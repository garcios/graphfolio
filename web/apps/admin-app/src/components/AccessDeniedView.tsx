import React from 'react';
import { useAdminAuth } from '../context/AdminAuthContext';
import { Card, Button } from '@graphfolio/ui';
import './AccessDeniedView.css';

export const AccessDeniedView: React.FC = () => {
  const { user, logout, isLoading } = useAdminAuth();

  return (
    <div className="admin-access-denied-page">
      <div className="admin-denied-glow" />
      <Card className="admin-denied-card">
        <div className="admin-denied-icon">🚫</div>
        <h1 className="admin-denied-title">403 Forbidden: Access Denied</h1>
        <p className="admin-denied-message">
          The account <strong>{user?.email}</strong> does not possess administrative privileges to
          access the GraphFolio Operational Portal.
        </p>

        <div className="admin-denied-badge-row">
          <span className="denied-label">Assigned Role:</span>
          <span className="denied-role-pill">{user?.role || 'USER'}</span>
        </div>

        <p className="admin-denied-subtext">
          If you believe this is an error, please contact the platform operations team or sign out
          and authenticate with your administrative credentials.
        </p>

        <div className="admin-denied-actions">
          <Button
            variant="primary"
            onClick={logout}
            disabled={isLoading}
            className="admin-denied-signout-btn"
          >
            {isLoading ? 'Signing Out...' : 'Sign Out & Return to Login'}
          </Button>
        </div>
      </Card>
    </div>
  );
};
