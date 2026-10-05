import React, { useId } from 'react';
import { cn } from '../../utils/classnames';
import './Select.css';

export interface SelectOption {
  label: string;
  value: string | number;
}

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  options?: SelectOption[];
  error?: string;
  helperText?: string;
}

export const Select: React.FC<SelectProps> = ({
  label,
  options,
  error,
  helperText,
  id,
  className,
  children,
  ...props
}) => {
  const generatedId = useId();
  const selectId = id || (label ? generatedId : undefined);

  return (
    <div className="gf-select-group">
      {label && (
        <label htmlFor={selectId} className="gf-select-label">
          {label}
        </label>
      )}
      <div className="gf-select-wrapper">
        <select
          id={selectId}
          className={cn('gf-select', error && 'gf-select--error', className)}
          {...props}
        >
          {options
            ? options.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))
            : children}
        </select>
      </div>
      {error && <span className="gf-select-error">{error}</span>}
      {!error && helperText && <span className="gf-select-helper">{helperText}</span>}
    </div>
  );
};
