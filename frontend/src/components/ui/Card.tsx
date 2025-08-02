import { HTMLAttributes, forwardRef, ReactNode } from 'react';
import { cn } from '@/utils/cn';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'elevated' | 'outlined' | 'glass';
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl';
  hover?: boolean;
  interactive?: boolean;
}

interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

interface CardContentProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

interface CardFooterProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

const cardVariants = {
  default: 'bg-white border border-secondary-200 shadow-soft',
  elevated: 'bg-white shadow-medium border border-secondary-100',
  outlined: 'bg-white border-2 border-secondary-300 shadow-xs',
  glass: 'bg-white/70 backdrop-blur-md border border-white/20 shadow-soft',
};

const cardPadding = {
  none: '',
  sm: 'p-4',
  md: 'p-6',
  lg: 'p-8',
  xl: 'p-10',
};

const Card = forwardRef<HTMLDivElement, CardProps>(({
  variant = 'default',
  padding = 'md',
  hover = false,
  interactive = false,
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        'rounded-xl transition-all duration-300',
        cardVariants[variant],
        cardPadding[padding],
        hover && 'hover:shadow-large hover:-translate-y-1',
        interactive && 'cursor-pointer hover:shadow-large hover:-translate-y-1 active:translate-y-0 active:shadow-medium',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
});

const CardHeader = forwardRef<HTMLDivElement, CardHeaderProps>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn('flex flex-col space-y-1.5 pb-6', className)}
      {...props}
    >
      {children}
    </div>
  );
});

const CardContent = forwardRef<HTMLDivElement, CardContentProps>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn('pt-0', className)}
      {...props}
    >
      {children}
    </div>
  );
});

const CardFooter = forwardRef<HTMLDivElement, CardFooterProps>(({
  className,
  children,
  ...props
}, ref) => {
  return (
    <div
      ref={ref}
      className={cn('flex items-center pt-6', className)}
      {...props}
    >
      {children}
    </div>
  );
});

Card.displayName = 'Card';
CardHeader.displayName = 'CardHeader';
CardContent.displayName = 'CardContent';
CardFooter.displayName = 'CardFooter';

export { Card, CardHeader, CardContent, CardFooter };