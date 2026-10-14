import React, { useState, useRef, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import './UserProfileMenu.css';

interface UserProfileMenuProps {
  onOpenPreferences: () => void;
}

function getInitials(name: string): string {
  if (!name) return 'U';
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) {
    return parts[0].substring(0, 2).toUpperCase();
  }
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export const UserProfileMenu: React.FC<UserProfileMenuProps> = ({ onOpenPreferences }) => {
  const { user, logout } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  const displayName = user?.displayName || 'Investor';
  const email = user?.email || '';
  const currency = user?.displayCurrency || 'USD';
  const initials = getInitials(displayName);

  return (
    <div className="user-profile-menu-container" ref={menuRef}>
      <button
        type="button"
        className={`user-profile-btn ${isOpen ? 'active' : ''}`}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        title="Investor Profile Menu"
      >
        <div className="avatar">{initials}</div>
        <div className="user-profile-details">
          <span className="user-profile-name">{displayName}</span>
          <span className="user-profile-currency">{currency}</span>
        </div>
        <span className="dropdown-caret">▾</span>
      </button>

      {isOpen && (
        <div className="user-profile-dropdown" role="menu">
          <div className="user-dropdown-header">
            <div className="dropdown-avatar">{initials}</div>
            <div className="dropdown-user-info">
              <span className="dropdown-name">{displayName}</span>
              <span className="dropdown-email">{email}</span>
              <span className="dropdown-currency-badge">Base: {currency}</span>
            </div>
          </div>

          <div className="dropdown-divider" />

          <button
            type="button"
            className="dropdown-item"
            role="menuitem"
            onClick={() => {
              setIsOpen(false);
              onOpenPreferences();
            }}
          >
            <span className="item-icon">⚙️</span>
            <span>Investor Preferences</span>
          </button>

          <div className="dropdown-divider" />

          <button
            type="button"
            className="dropdown-item dropdown-item-danger"
            role="menuitem"
            onClick={async () => {
              setIsOpen(false);
              await logout();
            }}
          >
            <span className="item-icon">🚪</span>
            <span>Sign Out</span>
          </button>
        </div>
      )}
    </div>
  );
};
