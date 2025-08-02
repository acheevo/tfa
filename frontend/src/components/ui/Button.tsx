import { ButtonHTMLAttributes, forwardRef } from 'react';
import { cn } from '@/utils/cn';
import { LucideIcon } from 'lucide-react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'success' | 'warning' | 'gradient';
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  icon?: LucideIcon;
  iconPosition?: 'left' | 'right';
  loading?: boolean;
  loadingText?: string;
  fullWidth?: boolean;
}

const variants = {
  primary: 'bg-primary-600 text-white hover:bg-primary-700 active:bg-primary-800 focus:ring-primary-500 shadow-sm hover:shadow-md active:shadow-sm transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  secondary: 'bg-secondary-100 text-secondary-900 hover:bg-secondary-200 active:bg-secondary-300 focus:ring-secondary-500 shadow-xs hover:shadow-sm transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  outline: 'border-2 border-primary-600 text-primary-600 hover:bg-primary-50 active:bg-primary-100 focus:ring-primary-500 hover:border-primary-700 transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  ghost: 'text-secondary-700 hover:bg-secondary-100 active:bg-secondary-200 focus:ring-secondary-500 hover:text-secondary-900 transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  destructive: 'bg-error-600 text-white hover:bg-error-700 active:bg-error-800 focus:ring-error-500 shadow-sm hover:shadow-md active:shadow-sm transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  success: 'bg-success-600 text-white hover:bg-success-700 active:bg-success-800 focus:ring-success-500 shadow-sm hover:shadow-md active:shadow-sm transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  warning: 'bg-warning-600 text-white hover:bg-warning-700 active:bg-warning-800 focus:ring-warning-500 shadow-sm hover:shadow-md active:shadow-sm transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
  gradient: 'bg-gradient-to-r from-primary-600 to-accent-600 text-white hover:from-primary-700 hover:to-accent-700 active:from-primary-800 active:to-accent-800 focus:ring-primary-500 shadow-md hover:shadow-lg active:shadow-md transform hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200',
};

const sizes = {
  xs: 'px-2.5 py-1.5 text-xs gap-1 min-h-[2rem]',
  sm: 'px-3 py-2 text-sm gap-1.5 min-h-[2.25rem]',
  md: 'px-4 py-2.5 text-base gap-2 min-h-[2.75rem]',
  lg: 'px-6 py-3 text-lg gap-2 min-h-[3rem]',
  xl: 'px-8 py-4 text-xl gap-3 min-h-[3.5rem]',
};

const Button = forwardRef<HTMLButtonElement, ButtonProps>(({
  variant = 'primary',
  size = 'md',
  icon: Icon,
  iconPosition = 'left',
  loading = false,
  loadingText,
  fullWidth = false,
  className,
  children,
  disabled,
  ...props
}, ref) => {
  const isDisabled = disabled || loading;

  return (
    <button
      ref={ref}
      className={cn(
        'inline-flex items-center justify-center font-medium rounded-lg transition-all duration-200',
        'focus:outline-none focus:ring-2 focus:ring-offset-2',
        'disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none disabled:hover:translate-y-0',
        'active:scale-[0.98]',
        variants[variant],
        sizes[size],
        fullWidth && 'w-full',
        className
      )}
      disabled={isDisabled}
      aria-label={loading ? `Loading... ${loadingText || ''}` : undefined}
      {...props}
    >
      {loading && (
        <svg
          className={cn(
            "animate-spin",
            size === 'xs' && "h-3 w-3",
            size === 'sm' && "h-3.5 w-3.5", 
            size === 'md' && "h-4 w-4",
            size === 'lg' && "h-4 w-4",
            size === 'xl' && "h-5 w-5"
          )}
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      
      {Icon && iconPosition === 'left' && !loading && (
        <Icon 
          className={cn(
            size === 'xs' && "h-3 w-3",
            size === 'sm' && "h-3.5 w-3.5", 
            size === 'md' && "h-4 w-4",
            size === 'lg' && "h-4 w-4",
            size === 'xl' && "h-5 w-5"
          )}
          aria-hidden="true"
        />
      )}
      
      <span className={loading ? 'animate-pulse' : ''}>
        {loading && loadingText ? loadingText : children}
      </span>
      
      {Icon && iconPosition === 'right' && !loading && (
        <Icon 
          className={cn(
            size === 'xs' && "h-3 w-3",
            size === 'sm' && "h-3.5 w-3.5", 
            size === 'md' && "h-4 w-4",
            size === 'lg' && "h-4 w-4",
            size === 'xl' && "h-5 w-5"
          )}
          aria-hidden="true"
        />
      )}
    </button>
  );
});

Button.displayName = 'Button';

export default Button;
export { Button };