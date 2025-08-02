import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { LoadingSpinner } from '../ui/LoadingSpinner';

interface PublicRouteProps {
  children: React.ReactNode;
  redirectPath?: string;
}

/**
 * PublicRoute component for routes that should only be accessible to unauthenticated users
 * Redirects authenticated users to dashboard or specified path
 * Used for login, register, forgot-password pages
 */
export const PublicRoute: React.FC<PublicRouteProps> = ({ 
  children, 
  redirectPath = '/dashboard' 
}) => {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  // Show loading spinner while checking authentication
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-secondary-50">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // If user is authenticated, redirect to intended destination or dashboard
  if (isAuthenticated) {
    // Check if there's a return path from login flow
    const returnTo = location.state?.returnTo || redirectPath;
    return <Navigate to={returnTo} replace />;
  }

  // User is not authenticated, render the public route
  return <>{children}</>;
};

export default PublicRoute;