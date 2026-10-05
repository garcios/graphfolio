import React from 'react';
import { cn } from '../../utils/classnames';
import './Table.css';

export interface TableProps extends React.TableHTMLAttributes<HTMLTableElement> {
  striped?: boolean;
  hoverable?: boolean;
}

export const Table: React.FC<TableProps> = ({
  children,
  className,
  striped = false,
  hoverable = true,
  ...props
}) => {
  return (
    <div className="gf-table-container">
      <table
        className={cn(
          'gf-table',
          striped && 'gf-table--striped',
          hoverable && 'gf-table--hoverable',
          className
        )}
        {...props}
      >
        {children}
      </table>
    </div>
  );
};

export const TableHead: React.FC<React.HTMLAttributes<HTMLTableSectionElement>> = ({
  children,
  className,
  ...props
}) => <thead className={className} {...props}>{children}</thead>;

export const TableBody: React.FC<React.HTMLAttributes<HTMLTableSectionElement>> = ({
  children,
  className,
  ...props
}) => <tbody className={className} {...props}>{children}</tbody>;

export const TableRow: React.FC<React.HTMLAttributes<HTMLTableRowElement>> = ({
  children,
  className,
  ...props
}) => <tr className={cn('gf-table__row', className)} {...props}>{children}</tr>;

export interface TableCellProps extends React.TdHTMLAttributes<HTMLTableCellElement> {
  align?: 'left' | 'center' | 'right';
}

export const TableCell: React.FC<TableCellProps> = ({
  children,
  className,
  align = 'left',
  ...props
}) => {
  return (
    <td
      className={cn('gf-table__td', align !== 'left' && `gf-table__td--${align}`, className)}
      {...props}
    >
      {children}
    </td>
  );
};

export interface TableHeaderCellProps extends React.ThHTMLAttributes<HTMLTableCellElement> {
  align?: 'left' | 'center' | 'right';
}

export const TableHeaderCell: React.FC<TableHeaderCellProps> = ({
  children,
  className,
  align = 'left',
  ...props
}) => {
  return (
    <th
      className={cn('gf-table__th', align !== 'left' && `gf-table__th--${align}`, className)}
      {...props}
    >
      {children}
    </th>
  );
};
