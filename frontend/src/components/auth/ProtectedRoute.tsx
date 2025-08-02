import React, { ReactNode } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useRBAC, UserRole, Permission, Resource, Action } from '../../contexts/RBACContext';

interface ProtectedRouteProps {
  children: ReactNode;
  requireEmailVerification?: boolean;
  fallback?: ReactNode;
  redirectTo?: string;
  // RBAC props
  requiredRole?: UserRole;
  requiredRoles?: UserRole[];
  requiredPermission?: Permission;
  requiredPermissions?: Permission[];
  requireAllPermissions?: boolean;
  requiredResource?: Resource;
  requiredAction?: Action;
  adminOnly?: boolean;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requireEmailVerification = false,
  fallback,
  redirectTo = '/login',
  // RBAC props
  requiredRole,
  requiredRoles,
  requiredPermission,
  requiredPermissions,
  requireAllPermissions = false,
  requiredResource,
  requiredAction,
  adminOnly = false,
}) => {
  const { isAuthenticated, isLoading, isEmailVerified } = useAuth();
  const { 
    hasRole, 
    hasAnyRole, 
    hasPermission, 
    hasAnyPermission, 
    hasAllPermissions, 
    canAccessResource,
    isAdmin 
  } = useRBAC();

  // Show loading state while checking authentication
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  // Check authentication
  if (!isAuthenticated) {
    if (fallback) {
      return <>{fallback}</>;
    }
    
    // Redirect to login
    window.location.href = redirectTo;
    return null;
  }

  // Check email verification if required
  if (requireEmailVerification && !isEmailVerified) {
    if (fallback) {
      return <>{fallback}</>;
    }
    
    // Redirect to email verification
    window.location.href = '/verify-email';
    return null;
  }

  // Check RBAC authorization
  let hasAuthorization = true;

  // Check admin requirement
  if (adminOnly && !isAdmin()) {
    hasAuthorization = false;
  }

  // Check role requirements
  if (hasAuthorization && requiredRole && !hasRole(requiredRole)) {
    hasAuthorization = false;
  }

  if (hasAuthorization && requiredRoles && !hasAnyRole(requiredRoles)) {
    hasAuthorization = false;
  }

  // Check permission requirements
  if (hasAuthorization && requiredPermission && !hasPermission(requiredPermission)) {
    hasAuthorization = false;
  }

  if (hasAuthorization && requiredPermissions) {
    if (requireAllPermissions) {
      if (!hasAllPermissions(requiredPermissions)) {
        hasAuthorization = false;
      }
    } else {
      if (!hasAnyPermission(requiredPermissions)) {
        hasAuthorization = false;
      }
    }
  }

  // Check resource access requirements
  if (hasAuthorization && requiredResource && requiredAction) {
    if (!canAccessResource(requiredResource, requiredAction)) {
      hasAuthorization = false;
    }
  }

  // If not authorized, handle accordingly
  if (!hasAuthorization) {
    if (fallback) {
      return <>{fallback}</>;
    }
    
    // Redirect to unauthorized page
    window.location.href = '/unauthorized';
    return null;
  }

  // Render children if all checks pass
  return <>{children}</>;
};

// Convenience components for common use cases

// Admin-only protected route
export const AdminRoute: React.FC<Omit<ProtectedRouteProps, 'adminOnly'>> = (props) => (
  <ProtectedRoute {...props} adminOnly={true} />
);

// Role-based protected route
export const RoleRoute: React.FC<ProtectedRouteProps & { role: UserRole }> = ({ role, ...props }) => (
  <ProtectedRoute {...props} requiredRole={role} />
);

// Permission-based protected route
export const PermissionRoute: React.FC<ProtectedRouteProps & { permission: Permission }> = ({ permission, ...props }) => (
  <ProtectedRoute {...props} requiredPermission={permission} />
);

// Resource access protected route
export const ResourceRoute: React.FC<ProtectedRouteProps & { resource: Resource; action: Action }> = ({ resource, action, ...props }) => (
  <ProtectedRoute {...props} requiredResource={resource} requiredAction={action} />
);

export default ProtectedRoute;