import React, { useState, useEffect } from 'react';
import { client } from '@graphfolio/api-client';
import { PasswordStrengthMeter, evaluatePassword } from './PasswordStrengthMeter';
import './ChangePasswordModal.css';

interface ChangePasswordModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const ChangePasswordModal: React.FC<ChangePasswordModalProps> = ({
  isOpen,
  onClose,
}) => {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [revokeOtherSessions, setRevokeOtherSessions] = useState(true);

  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Reset form state on open / close
  useEffect(() => {
    if (isOpen) {
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      setRevokeOtherSessions(true);
      setShowCurrentPassword(false);
      setShowNewPassword(false);
      setError(null);
      setSuccess(null);
    }
  }, [isOpen]);

  // Keyboard navigation & body scroll lock
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !submitting) {
        onClose();
      }
    };

    document.body.style.overflow = 'hidden';
    window.addEventListener('keydown', handleKeyDown);

    return () => {
      document.body.style.overflow = '';
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, submitting, onClose]);

  if (!isOpen) return null;

  const passwordEvaluation = evaluatePassword(newPassword);
  const passwordsMatch = newPassword && confirmPassword && newPassword === confirmPassword;
  const isSamePassword = currentPassword && newPassword && currentPassword === newPassword;

  const canSubmit =
    Boolean(currentPassword.trim()) &&
    passwordEvaluation.isValid &&
    passwordsMatch &&
    !isSamePassword &&
    !submitting;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!currentPassword) {
      setError('Please enter your current password.');
      return;
    }

    if (!passwordEvaluation.isValid) {
      setError('New password must satisfy all security complexity requirements.');
      return;
    }

    if (isSamePassword) {
      setError('New password cannot be the same as your current password.');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    setSubmitting(true);

    try {
      const res = await client.mutation({
        changePassword: {
          __args: {
            input: {
              currentPassword,
              newPassword,
              revokeOtherSessions,
            },
          },
          success: true,
          message: true,
        },
      });

      if (res.changePassword?.success) {
        setSuccess(
          res.changePassword.message ||
            'Password updated successfully. Other sessions have been signed out.'
        );
        setCurrentPassword('');
        setNewPassword('');
        setConfirmPassword('');

        // Auto-close after brief notification
        setTimeout(() => {
          onClose();
        }, 1500);
      } else {
        setError(res.changePassword?.message || 'Failed to change password.');
      }
    } catch (err: any) {
      console.error('Failed to change password:', err);
      setError(
        err?.message ||
          'Failed to update password. Please check your current password and try again.'
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="pwd-modal-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget && !submitting) {
          onClose();
        }
      }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="pwd-modal-title"
    >
      <div className="pwd-modal-container">
        <div className="pwd-modal-header">
          <h2 id="pwd-modal-title" className="pwd-modal-title">
            <span>🔑</span> Change Password
          </h2>
          <button
            type="button"
            className="pwd-close-btn"
            onClick={onClose}
            disabled={submitting}
            aria-label="Close dialog"
          >
            &times;
          </button>
        </div>

        <form onSubmit={handleSubmit} className="pwd-modal-form">
          {error && (
            <div className="pwd-error-banner" role="alert">
              <span>⚠️</span>
              <span>{error}</span>
            </div>
          )}

          {success && (
            <div className="pwd-success-banner" role="status">
              <span>✓</span>
              <span>{success}</span>
            </div>
          )}

          {/* Current Password */}
          <div className="pwd-form-group">
            <label htmlFor="pwd-current" className="pwd-label">
              Current Password
            </label>
            <div className="pwd-input-wrapper">
              <input
                id="pwd-current"
                type={showCurrentPassword ? 'text' : 'password'}
                className="pwd-input"
                placeholder="Enter your current password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                disabled={submitting}
                autoComplete="current-password"
                required
              />
              <button
                type="button"
                className="pwd-toggle-visibility"
                onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                tabIndex={-1}
                aria-label={showCurrentPassword ? 'Hide current password' : 'Show current password'}
              >
                {showCurrentPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
          </div>

          {/* New Password */}
          <div className="pwd-form-group">
            <label htmlFor="pwd-new" className="pwd-label">
              New Password
            </label>
            <div className="pwd-input-wrapper">
              <input
                id="pwd-new"
                type={showNewPassword ? 'text' : 'password'}
                className={`pwd-input ${isSamePassword ? 'pwd-input-error' : ''}`}
                placeholder="Enter a new secure password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                disabled={submitting}
                autoComplete="new-password"
                required
              />
              <button
                type="button"
                className="pwd-toggle-visibility"
                onClick={() => setShowNewPassword(!showNewPassword)}
                tabIndex={-1}
                aria-label={showNewPassword ? 'Hide new password' : 'Show new password'}
              >
                {showNewPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
            {isSamePassword && (
              <span className="pwd-field-warning">
                New password cannot be identical to current password
              </span>
            )}
            <PasswordStrengthMeter password={newPassword} />
          </div>

          {/* Confirm New Password */}
          <div className="pwd-form-group">
            <label htmlFor="pwd-confirm" className="pwd-label">
              Confirm New Password
            </label>
            <div className="pwd-input-wrapper">
              <input
                id="pwd-confirm"
                type={showNewPassword ? 'text' : 'password'}
                className={`pwd-input ${
                  confirmPassword && !passwordsMatch ? 'pwd-input-error' : ''
                }`}
                placeholder="Re-enter your new password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                disabled={submitting}
                autoComplete="new-password"
                required
              />
            </div>
            {confirmPassword && !passwordsMatch && (
              <span className="pwd-field-error">Passwords do not match</span>
            )}
          </div>

          {/* Revoke Other Sessions */}
          <div className="pwd-form-group">
            <label className="pwd-checkbox-label">
              <input
                type="checkbox"
                className="pwd-checkbox"
                checked={revokeOtherSessions}
                onChange={(e) => setRevokeOtherSessions(e.target.checked)}
                disabled={submitting}
              />
              <div>
                <span className="pwd-checkbox-text">
                  Sign out of all other devices
                </span>
                <span className="pwd-checkbox-hint">
                  Recommended. Revokes all other active sessions and background refresh tokens.
                </span>
              </div>
            </label>
          </div>

          <div className="pwd-modal-footer">
            <button
              type="button"
              className="pwd-btn-cancel"
              onClick={onClose}
              disabled={submitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="pwd-btn-save"
              disabled={!canSubmit}
            >
              {submitting && <span className="pwd-spinner" />}
              {submitting ? 'Updating...' : 'Update Password'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
