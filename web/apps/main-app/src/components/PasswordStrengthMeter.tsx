import React from 'react';
import './PasswordStrengthMeter.css';

interface PasswordCriteria {
  label: string;
  met: boolean;
}

export function evaluatePassword(password: string): {
  score: number; // 0 to 4
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
  // Score: 0 (weak) to 4 (strong)
  let score = 0;
  if (metCount >= 5) score = 4;
  else if (metCount >= 4) score = 3;
  else if (metCount >= 3) score = 2;
  else if (metCount >= 1) score = 1;

  const isValid = criteria.every(c => c.met);

  return { score, criteria, isValid };
}

interface PasswordStrengthMeterProps {
  password: string;
}

export const PasswordStrengthMeter: React.FC<PasswordStrengthMeterProps> = ({ password }) => {
  if (!password) {
    return null;
  }

  const { score, criteria } = evaluatePassword(password);

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

  const getStrengthClass = () => {
    switch (score) {
      case 0:
      case 1:
        return 'strength-weak';
      case 2:
        return 'strength-fair';
      case 3:
        return 'strength-good';
      case 4:
        return 'strength-strong';
      default:
        return '';
    }
  };

  return (
    <div className="password-strength-container" role="region" aria-label="Password strength indicator">
      <div className="strength-header">
        <span className="strength-title">Password Strength:</span>
        <span className={`strength-badge ${getStrengthClass()}`}>{getStrengthLabel()}</span>
      </div>

      <div className="strength-bars">
        <div className={`bar ${score >= 1 ? getStrengthClass() : ''}`} />
        <div className={`bar ${score >= 2 ? getStrengthClass() : ''}`} />
        <div className={`bar ${score >= 3 ? getStrengthClass() : ''}`} />
        <div className={`bar ${score >= 4 ? getStrengthClass() : ''}`} />
      </div>

      <ul className="criteria-list">
        {criteria.map((c, i) => (
          <li key={i} className={`criteria-item ${c.met ? 'met' : 'unmet'}`}>
            <span className="criteria-icon">{c.met ? '✓' : '○'}</span>
            <span className="criteria-text">{c.label}</span>
          </li>
        ))}
      </ul>
    </div>
  );
};
