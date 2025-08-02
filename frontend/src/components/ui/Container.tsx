import { ReactNode } from 'react';
import { cn } from '@/utils/cn';

interface ContainerProps {
  children: ReactNode;
  className?: string;
}

function Container({ children, className }: ContainerProps) {
  return (
    <div className={cn('max-w-7xl mx-auto px-4 sm:px-6 lg:px-8', className)}>
      {children}
    </div>
  );
}

export default Container;
export { Container };