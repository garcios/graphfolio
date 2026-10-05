import React from 'react';
import { cn } from '../../utils/classnames';
import './Badge.css';

export type BadgeVariant =
  | 'buy'
  | 'sell'
  | 'dividend'
  | 'deposit'
  | 'withdrawal'
  | 'active'
  | 'inactive'
  | 'success'
  | 'danger'
  | 'warning'
  | 'info'
  | 'equity'
  | 'etf'
  | 'crypto'
  | 'neutral'
  | 'default';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant | string;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  className,
  variant = 'neutral',
  ...props
}) => {
  const normVariant = variant.toLowerCase().replace(/_/g, '-');

  return (
    <span
      className={cn('gf-badge', `gf-badge--${normVariant}`, className)}
      {...props}
    >
      {children}
    </span>
  );
};
