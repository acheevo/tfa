import React, { ReactNode } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useRBAC, UserRole, Permission, Resource, Action } from '../../contexts/RBACContext';

// Base guard props
interface BaseGuardProps {
  children: ReactNode;
  fallback?: ReactNode;
  redirectTo?: string;
  showError?: boolean;
  errorMessage?: string;
}

// Role-based guard props
interface RoleGuardProps extends BaseGuardProps {
  role?: UserRole;
  roles?: UserRole[];
  requireAll?: boolean;
}

// Permission-based guard props
interface PermissionGuardProps extends BaseGuardProps {
  permission?: Permission;
  permissions?: Permission[];
  requireAll?: boolean;
}

// Resource access guard props
interface ResourceGuardProps extends BaseGuardProps {
  resource: Resource;
  action: Action;
}

// Admin guard props
interface AdminGuardProps extends BaseGuardProps {}

// User management guard props
interface UserManagementGuardProps extends BaseGuardProps {
  targetUserId?: number;
}

// Own resource guard props
interface OwnResourceGuardProps extends BaseGuardProps {
  resourceOwnerId: number;
}

// Default error component
const DefaultError: React.FC<{ message?: string }> = ({ message }) => (
  <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-center">
    <div className="text-red-600 font-medium mb-2">Access Denied</div>
    <div className="text-red-500 text-sm">
      {message || 'You don\'t have permission to access this content.'}
    </div>
  </div>
);

// Default loading component
const DefaultLoading: React.FC = () => (
  <div className="flex items-center justify-center p-4">
    <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
    <span className="ml-2 text-gray-600">Checking permissions...</span>
  </div>
);

// Base guard component
const BaseGuard: React.FC<{
  condition: boolean;
  loading?: boolean;
  fallback?: ReactNode;
  redirectTo?: string;
  showError?: boolean;
  errorMessage?: string;
  children: ReactNode;
}> = ({ 
  condition, 
  loading = false, 
  fallback, 
  redirectTo, 
  showError = true, 
  errorMessage,
  children 
}) => {
  const { isLoading } = useAuth();

  // Show loading state
  if (isLoading || loading) {
    return <DefaultLoading />;
  }

  // Handle redirect
  if (!condition && redirectTo) {
    window.location.href = redirectTo;
    return <DefaultLoading />;
  }

  // Show content if condition is met
  if (condition) {
    return <>{children}</>;
  }

  // Show fallback or error
  if (fallback) {
    return <>{fallback}</>;
  }

  if (showError) {
    return <DefaultError message={errorMessage} />;
  }

  return null;
};

// Role Guard - Protects content based on user roles
export const RoleGuard: React.FC<RoleGuardProps> = ({
  role,
  roles,
  requireAll = false,
  children,
  ...baseProps
}) => {
  const { hasRole, hasAnyRole } = useRBAC();
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <BaseGuard
        condition={false}
        errorMessage="Please log in to access this content."
        {...baseProps}
      >
        {children}
      </BaseGuard>
    );
  }

  let condition = false;

  if (role) {
    condition = hasRole(role);
  } else if (roles) {
    if (requireAll) {
      // Check if user has all specified roles (not practical with current role system)
      condition = roles.every(r => hasRole(r));
    } else {
      // Check if user has any of the specified roles
      condition = hasAnyRole(roles);
    }
  }

  return (
    <BaseGuard condition={condition} {...baseProps}>
      {children}
    </BaseGuard>
  );
};

// Permission Guard - Protects content based on permissions
export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  permission,
  permissions,
  requireAll = false,
  children,
  ...baseProps
}) => {
  const { hasPermission, hasAnyPermission, hasAllPermissions } = useRBAC();
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <BaseGuard
        condition={false}
        errorMessage="Please log in to access this content."
        {...baseProps}
      >
        {children}
      </BaseGuard>
    );
  }

  let condition = false;

  if (permission) {
    condition = hasPermission(permission);
  } else if (permissions) {
    if (requireAll) {
      condition = hasAllPermissions(permissions);
    } else {
      condition = hasAnyPermission(permissions);
    }
  }

  return (
    <BaseGuard condition={condition} {...baseProps}>
      {children}
    </BaseGuard>
  );
};

// Resource Guard - Protects content based on resource access
export const ResourceGuard: React.FC<ResourceGuardProps> = ({
  resource,
  action,
  children,
  ...baseProps
}) => {
  const { canAccessResource } = useRBAC();
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <BaseGuard
        condition={false}
        errorMessage="Please log in to access this content."
        {...baseProps}
      >
        {children}
      </BaseGuard>
    );
  }

  const condition = canAccessResource(resource, action);

  return (
    <BaseGuard condition={condition} {...baseProps}>
      {children}
    </BaseGuard>
  );
};

// Admin Guard - Protects admin-only content
export const AdminGuard: React.FC<AdminGuardProps> = ({
  children,
  ...baseProps
}) => {
  return (
    <RoleGuard
      role="admin"
      errorMessage="Administrator access required."
      {...baseProps}
    >
      {children}
    </RoleGuard>
  );
};

// User Management Guard - Protects user management content
export const UserManagementGuard: React.FC<UserManagementGuardProps> = ({
  targetUserId,
  children,
  ...baseProps
}) => {
  const { canManageUser } = useRBAC();
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <BaseGuard
        condition={false}
        errorMessage="Please log in to access this content."
        {...baseProps}
      >
        {children}
      </BaseGuard>
    );
  }

  const condition = canManageUser(targetUserId);

  return (
    <BaseGuard 
      condition={condition}
      errorMessage="You don't have permission to manage users."
      {...baseProps}
    >
      {children}
    </BaseGuard>
  );
};

// Own Resource Guard - Protects content that user owns or has permission to access
export const OwnResourceGuard: React.FC<OwnResourceGuardProps> = ({
  resourceOwnerId,
  children,
  ...baseProps
}) => {
  const { canAccessOwnResource } = useRBAC();
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <BaseGuard
        condition={false}
        errorMessage="Please log in to access this content."
        {...baseProps}
      >
        {children}
      </BaseGuard>
    );
  }

  const condition = canAccessOwnResource(resourceOwnerId);

  return (
    <BaseGuard 
      condition={condition}
      errorMessage="You don't have permission to access this resource."
      {...baseProps}
    >
      {children}
    </BaseGuard>
  );
};

// Conditional render based on permissions
interface ConditionalRenderProps {
  condition: boolean;
  children: ReactNode;
  fallback?: ReactNode;
}

export const ConditionalRender: React.FC<ConditionalRenderProps> = ({
  condition,
  children,
  fallback = null,
}) => {
  return condition ? <>{children}</> : <>{fallback}</>;
};

// Higher-order component for route protection
export const withRoleGuard = <P extends object>(
  Component: React.ComponentType<P>,
  guardProps: Omit<RoleGuardProps, 'children'>
) => {
  return (props: P) => (
    <RoleGuard {...guardProps}>
      <Component {...props} />
    </RoleGuard>
  );
};

export const withPermissionGuard = <P extends object>(
  Component: React.ComponentType<P>,
  guardProps: Omit<PermissionGuardProps, 'children'>
) => {
  return (props: P) => (
    <PermissionGuard {...guardProps}>
      <Component {...props} />
    </PermissionGuard>
  );
};

export const withAdminGuard = <P extends object>(
  Component: React.ComponentType<P>,
  guardProps?: Omit<AdminGuardProps, 'children'>
) => {
  return (props: P) => (
    <AdminGuard {...guardProps}>
      <Component {...props} />
    </AdminGuard>
  );
};

// Navigation guard hook for programmatic route protection
export const useNavigationGuard = () => {
  const { hasPermission, isAdmin } = useRBAC();
  const { isAuthenticated } = useAuth();

  const canNavigateTo = (route: string): boolean => {
    if (!isAuthenticated) {
      return false;
    }

    // Define route protection rules
    const protectedRoutes: Record<string, () => boolean> = {
      '/admin': () => isAdmin(),
      '/admin/*': () => isAdmin(),
      '/profile/edit': () => isAuthenticated,
      '/users': () => hasPermission('user:read'),
      '/users/*': () => hasPermission('user:read'),
    };

    // Check if route is protected
    for (const [pattern, checkFn] of Object.entries(protectedRoutes)) {
      if (route.match(pattern.replace('*', '.*'))) {
        return checkFn();
      }
    }

    // Allow access to non-protected routes
    return true;
  };

  const redirectIfUnauthorized = (route: string, fallbackRoute: string = '/') => {
    if (!canNavigateTo(route)) {
      window.location.href = fallbackRoute;
      return false;
    }
    return true;
  };

  return {
    canNavigateTo,
    redirectIfUnauthorized,
  };
};

export default RoleGuard;