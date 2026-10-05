import React from 'react';
import { cn } from '../../utils/classnames';
import './Button.css';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  className,
  variant = 'primary',
  size = 'md',
  isLoading = false,
  disabled,
  leftIcon,
  rightIcon,
  ...props
}) => {
  return (
    <button
      className={cn(
        'gf-button',
        `gf-button--${variant}`,
        `gf-button--${size}`,
        isLoading && 'gf-button--loading',
        className
      )}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? (
        <span className="gf-button__spinner" aria-hidden="true" />
      ) : (
        leftIcon && <span className="gf-button__icon">{leftIcon}</span>
      )}
      <span>{children}</span>
      {!isLoading && rightIcon && <span className="gf-button__icon">{rightIcon}</span>}
    </button>
  );
};
