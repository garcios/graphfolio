import React, { useState } from 'react';
import { useAdminAuth } from '../context/AdminAuthContext';
import { Card, Button, Input } from '@graphfolio/ui';
import './AdminChangePasswordView.css';

interface PasswordCriteria {
  label: string;
  met: boolean;
}

function evaluatePassword(password: string): {
  score: number;
  criteria: PasswordCriteria[];
  isValid: boolean;
} {
  const criteria: PasswordCriteria[] = [
    { label: 'At least 10 characters', met: password.length >= 10 },
    { label: 'At least 1 uppercase letter (A-Z)', met: /[A-Z]/.test(password) },
    { label: 'At least 1 lowercase letter (a-z)', met: /[a-z]/.test(password) },
    { label: 'At least 1 number (0-9)', met: /[0-9]/.test(password) },
    { label: 'At least 1 special character (!@#$%^&*)', met: /[^A-Za-z0-9]/.test(password) },
  ];

  const metCount = criteria.filter(c => c.met).length;
  let score = 0;
  if (metCount >= 5) score = 4;
  else if (metCount >= 4) score = 3;
  else if (metCount >= 3) score = 2;
  else if (metCount >= 1) score = 1;

  const isValid = criteria.every(c => c.met);
  return { score, criteria, isValid };
}

interface AdminChangePasswordViewProps {
  onSuccess?: () => void;
  isModal?: boolean;
  onCancel?: () => void;
}

export const AdminChangePasswordView: React.FC<AdminChangePasswordViewProps> = ({
  onSuccess,
  isModal = false,
  onCancel,
}) => {
  const { changePassword, logout, isLoading, error, clearError } = useAdminAuth();

  const [currentPassword, setCurrentPassword] = useState('Password123!');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [revokeOtherSessions, setRevokeOtherSessions] = useState(true);
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);

  const { score, criteria, isValid } = evaluatePassword(newPassword);

  const getStrengthLabel = () => {
    switch (score) {
      case 0:
      case 1:
        return 'Weak';
      case 2:
        return 'Fair';
      case 3:
        return 'Good';
      case 4:
        return 'Strong';
      default:
        return '';
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);
    clearError();

    if (!currentPassword) {
      setValidationError('Please enter your current password.');
      return;
    }

    if (!isValid) {
      setValidationError('New password must satisfy all 5 complexity requirements.');
      return;
    }

    if (currentPassword === newPassword) {
      setValidationError('New password cannot be the same as your current password.');
      return;
    }

    if (newPassword !== confirmPassword) {
      setValidationError('New password and confirmation do not match.');
      return;
    }

    try {
      await changePassword(currentPassword, newPassword, revokeOtherSessions);
      if (onSuccess) {
        onSuccess();
      }
    } catch {
      // Error handled by AdminAuthContext
    }
  };

  const displayError = validationError || error;

  const content = (
    <Card className="admin-change-password-card">
      <div className="admin-change-header">
        <div className="admin-key-badge">🔑</div>
        <h1 className="admin-change-title">
          {isModal ? 'Change Administrator Password' : 'Initial Password Rotation Required'}
        </h1>
        <p className="admin-change-subtitle">
          {isModal
            ? 'Update your administrator password to maintain strong platform hygiene.'
            : 'You are currently authenticated with the default auto-provisioned password. You must establish a new, private password before accessing the administrative console.'}
        </p>
      </div>

      {!isModal && (
        <div className="admin-security-banner">
          <span className="admin-security-banner-icon">🛡️</span>
          <div className="admin-security-banner-text">
            <strong>NIST SP 800-63B Compliance Policy:</strong> Default bootstrap credentials
            must be rotated upon initial login to prevent unauthorized administrative takeover.
          </div>
        </div>
      )}

      {displayError && (
        <div className="admin-alert-box" role="alert">
          <span>⚠️</span>
          <span>{displayError}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="admin-change-form">
        <div className="admin-field-group">
          <div className="admin-label-row">
            <label htmlFor="current-pw" className="admin-label">
              Current Password
            </label>
            <button
              type="button"
              className="admin-toggle-btn"
              onClick={() => setShowCurrentPassword(!showCurrentPassword)}
              tabIndex={-1}
            >
              {showCurrentPassword ? 'Hide' : 'Show'}
            </button>
          </div>
          <Input
            id="current-pw"
            type={showCurrentPassword ? 'text' : 'password'}
            autoComplete="current-password"
            value={currentPassword}
            onChange={(e) => {
              setCurrentPassword(e.target.value);
              if (validationError) setValidationError(null);
            }}
            placeholder="Enter current password"
            required
            disabled={isLoading}
          />
        </div>

        <div className="admin-field-group">
          <div className="admin-label-row">
            <label htmlFor="new-pw" className="admin-label">
              New Password
            </label>
            <button
              type="button"
              className="admin-toggle-btn"
              onClick={() => setShowNewPassword(!showNewPassword)}
              tabIndex={-1}
            >
              {showNewPassword ? 'Hide' : 'Show'}
            </button>
          </div>
          <Input
            id="new-pw"
            type={showNewPassword ? 'text' : 'password'}
            autoComplete="new-password"
            value={newPassword}
            onChange={(e) => {
              setNewPassword(e.target.value);
              if (validationError) setValidationError(null);
            }}
            placeholder="Enter new strong password"
            required
            disabled={isLoading}
          />
        </div>

        {newPassword && (
          <div className="admin-strength-section">
            <div className="admin-strength-meter-header">
              <span className="admin-strength-title">Password Strength</span>
              <span className={`admin-strength-badge strength-${score}`}>
                {getStrengthLabel()}
              </span>
            </div>

            <div className="admin-strength-bar">
              <div
                className={`admin-strength-fill strength-${score}`}
                style={{ width: `${(score / 4) * 100}%` }}
              />
            </div>

            <ul className="admin-criteria-list">
              {criteria.map((item, index) => (
                <li key={index} className={`admin-criteria-item ${item.met ? 'met' : ''}`}>
                  <span className="criteria-icon">{item.met ? '✓' : '○'}</span>
                  <span>{item.label}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        <div className="admin-field-group">
          <label htmlFor="confirm-pw" className="admin-label">
            Confirm New Password
          </label>
          <Input
            id="confirm-pw"
            type={showNewPassword ? 'text' : 'password'}
            autoComplete="new-password"
            value={confirmPassword}
            onChange={(e) => {
              setConfirmPassword(e.target.value);
              if (validationError) setValidationError(null);
            }}
            placeholder="Re-enter new password"
            required
            disabled={isLoading}
          />
          {confirmPassword && newPassword !== confirmPassword && (
            <span className="admin-field-hint error">Passwords do not match</span>
          )}
          {confirmPassword && newPassword === confirmPassword && (
            <span className="admin-field-hint success">✓ Passwords match</span>
          )}
        </div>

        <div className="admin-checkbox-group">
          <label className="admin-checkbox-label">
            <input
              type="checkbox"
              checked={revokeOtherSessions}
              onChange={(e) => setRevokeOtherSessions(e.target.checked)}
              disabled={isLoading}
            />
            <span>Revoke all other active sessions across devices</span>
          </label>
        </div>

        <div className="admin-actions-row">
          {isModal && onCancel && (
            <Button
              type="button"
              variant="secondary"
              onClick={onCancel}
              disabled={isLoading}
            >
              Cancel
            </Button>
          )}

          {!isModal && (
            <Button
              type="button"
              variant="secondary"
              onClick={logout}
              disabled={isLoading}
            >
              Sign Out
            </Button>
          )}

          <Button
            type="submit"
            variant="primary"
            className="admin-change-submit-btn"
            disabled={isLoading || !isValid || newPassword !== confirmPassword || currentPassword === newPassword}
          >
            {isLoading ? 'Updating Credentials...' : 'Save Password & Access Portal'}
          </Button>
        </div>
      </form>
    </Card>
  );

  if (isModal) {
    return content;
  }

  return (
    <div className="admin-change-password-page">
      <div className="admin-change-glow" />
      {content}
    </div>
  );
};
