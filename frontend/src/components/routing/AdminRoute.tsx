import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { useRBAC } from '../../contexts/RBACContext';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { UnauthorizedError } from '../ui/UnauthorizedError';

interface AdminRouteProps {
  children: React.ReactNode;
  fallbackPath?: string;
  showErrorPage?: boolean;
}

/**
 * AdminRoute component that ensures user has admin role before rendering children
 * Redirects to login if not authenticated
 * Shows unauthorized error or redirects if user doesn't have admin role
 */
export const AdminRoute: React.FC<AdminRouteProps> = ({ 
  children, 
  fallbackPath = '/dashboard',
  showErrorPage = true
}) => {
  const { isAuthenticated, isLoading } = useAuth();
  const { isAdmin } = useRBAC();
  const location = useLocation();

  // Show loading spinner while checking authentication
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-secondary-50">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // Redirect to login if not authenticated, preserving the intended location
  if (!isAuthenticated) {
    return (
      <Navigate 
        to="/login" 
        state={{ returnTo: location.pathname + location.search }} 
        replace 
      />
    );
  }

  // Check if user has admin role
  if (!isAdmin()) {
    if (showErrorPage) {
      return (
        <UnauthorizedError
          title="Admin Access Required"
          message="You need administrator privileges to access this page."
          actionText="Go to Dashboard"
          actionPath={fallbackPath}
        />
      );
    } else {
      return <Navigate to={fallbackPath} replace />;
    }
  }

  // User is authenticated and has admin role, render children
  return <>{children}</>;
};

export default AdminRoute;