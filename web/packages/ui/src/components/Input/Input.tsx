import React, { useId } from 'react';
import { cn } from '../../utils/classnames';
import './Input.css';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Input: React.FC<InputProps> = ({
  label,
  error,
  helperText,
  id,
  className,
  ...props
}) => {
  const generatedId = useId();
  const inputId = id || (label ? generatedId : undefined);

  return (
    <div className="gf-input-group">
      {label && (
        <label htmlFor={inputId} className="gf-input-label">
          {label}
        </label>
      )}
      <div className="gf-input-wrapper">
        <input
          id={inputId}
          className={cn('gf-input', error && 'gf-input--error', className)}
          {...props}
        />
      </div>
      {error && <span className="gf-input-error">{error}</span>}
      {!error && helperText && <span className="gf-input-helper">{helperText}</span>}
    </div>
  );
};
