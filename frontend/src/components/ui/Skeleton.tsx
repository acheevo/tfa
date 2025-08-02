import { HTMLAttributes, forwardRef } from 'react';
import { cn } from '@/utils/cn';

interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular' | 'rounded';
  animation?: 'pulse' | 'wave' | 'none';
  width?: string | number;
  height?: string | number;
}

const Skeleton = forwardRef<HTMLDivElement, SkeletonProps>(({
  variant = 'text',
  animation = 'pulse',
  width,
  height,
  className,
  style,
  ...props
}, ref) => {
  const baseClasses = 'bg-gradient-to-r from-secondary-200 via-secondary-100 to-secondary-200 bg-[length:200%_100%]';
  
  const variantClasses = {
    text: 'rounded h-4',
    circular: 'rounded-full',
    rectangular: 'rounded-none',
    rounded: 'rounded-lg',
  };

  const animationClasses = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer',
    none: '',
  };

  const inlineStyles = {
    ...style,
    ...(width && { width: typeof width === 'number' ? `${width}px` : width }),
    ...(height && { height: typeof height === 'number' ? `${height}px` : height }),
  };

  return (
    <div
      ref={ref}
      className={cn(
        baseClasses,
        variantClasses[variant],
        animationClasses[animation],
        className
      )}
      style={inlineStyles}
      {...props}
    />
  );
});

// Preset skeleton components for common patterns
const SkeletonText = forwardRef<HTMLDivElement, Omit<SkeletonProps, 'variant'>>(({ 
  className, 
  ...props 
}, ref) => (
  <Skeleton ref={ref} variant="text" className={cn('h-4', className)} {...props} />
));

const SkeletonCard = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(({ 
  className, 
  ...props 
}, ref) => (
  <div ref={ref} className={cn('space-y-4 p-6', className)} {...props}>
    <div className="flex items-center space-x-4">
      <Skeleton variant="circular" width={40} height={40} />
      <div className="space-y-2 flex-1">
        <SkeletonText className="w-3/4" />
        <SkeletonText className="w-1/2 h-3" />
      </div>
    </div>
    <div className="space-y-2">
      <SkeletonText />
      <SkeletonText />
      <SkeletonText className="w-5/6" />
    </div>
  </div>
));

const SkeletonTable = forwardRef<HTMLDivElement, { rows?: number } & HTMLAttributes<HTMLDivElement>>(({ 
  rows = 5,
  className, 
  ...props 
}, ref) => (
  <div ref={ref} className={cn('space-y-4', className)} {...props}>
    {/* Header */}
    <div className="flex space-x-4 pb-4 border-b border-secondary-200">
      <SkeletonText className="w-1/4 h-3" />
      <SkeletonText className="w-1/4 h-3" />
      <SkeletonText className="w-1/4 h-3" />
      <SkeletonText className="w-1/4 h-3" />
    </div>
    
    {/* Rows */}
    {Array.from({ length: rows }).map((_, index) => (
      <div key={index} className="flex space-x-4 py-2">
        <SkeletonText className="w-1/4" />
        <SkeletonText className="w-1/4" />
        <SkeletonText className="w-1/4" />
        <SkeletonText className="w-1/4" />
      </div>
    ))}
  </div>
));

const SkeletonStats = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(({ 
  className, 
  ...props 
}, ref) => (
  <div ref={ref} className={cn('grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6', className)} {...props}>
    {Array.from({ length: 4 }).map((_, index) => (
      <div key={index} className="p-6 border border-secondary-200 rounded-xl">
        <div className="flex items-center">
          <Skeleton variant="circular" width={40} height={40} className="mr-4" />
          <div className="space-y-2">
            <SkeletonText className="w-20 h-3" />
            <SkeletonText className="w-16 h-6" />
          </div>
        </div>
      </div>
    ))}
  </div>
));

Skeleton.displayName = 'Skeleton';
SkeletonText.displayName = 'SkeletonText';
SkeletonCard.displayName = 'SkeletonCard';
SkeletonTable.displayName = 'SkeletonTable';
SkeletonStats.displayName = 'SkeletonStats';

export { 
  Skeleton, 
  SkeletonText, 
  SkeletonCard, 
  SkeletonTable, 
  SkeletonStats 
};