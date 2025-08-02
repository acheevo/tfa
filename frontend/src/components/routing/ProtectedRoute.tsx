import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { LoadingSpinner } from '../ui/LoadingSpinner';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireEmailVerification?: boolean;
}

/**
 * ProtectedRoute component that ensures user is authenticated before rendering children
 * Redirects to login with return path if not authenticated
 * Shows loading spinner while authentication status is being checked
 */
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ 
  children, 
  requireEmailVerification = false 
}) => {
  const { isAuthenticated, isLoading, isEmailVerified } = useAuth();
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

  // Redirect to email verification if required and not verified
  if (requireEmailVerification && !isEmailVerified) {
    return (
      <Navigate 
        to="/verify-email" 
        state={{ returnTo: location.pathname + location.search }} 
        replace 
      />
    );
  }

  // User is authenticated and verified (if required), render children
  return <>{children}</>;
};

export default ProtectedRoute;