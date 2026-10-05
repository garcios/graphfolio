import React from 'react';
import { cn } from '../../utils/classnames';
import './Card.css';

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  glow?: 'none' | 'green' | 'red' | 'blue' | 'amber';
  hoverable?: boolean;
}

export const Card: React.FC<CardProps> = ({
  children,
  className,
  glow = 'none',
  hoverable = false,
  ...props
}) => {
  return (
    <div
      className={cn(
        'gf-card',
        glow !== 'none' && `gf-card--glow-${glow}`,
        hoverable && 'gf-card--hoverable',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
};

export const CardHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => <div className={cn('gf-card__header', className)} {...props}>{children}</div>;

export const CardTitle: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  children,
  className,
  ...props
}) => <h3 className={cn('gf-card__title', className)} {...props}>{children}</h3>;

export const CardDescription: React.FC<React.HTMLAttributes<HTMLParagraphElement>> = ({
  children,
  className,
  ...props
}) => <p className={cn('gf-card__description', className)} {...props}>{children}</p>;

export const CardContent: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => <div className={cn('gf-card__content', className)} {...props}>{children}</div>;

export const CardFooter: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => <div className={cn('gf-card__footer', className)} {...props}>{children}</div>;
