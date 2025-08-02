import React, { createContext, useContext, ReactNode } from 'react';
import { useAuth } from './AuthContext';

// Types
export type Permission = string;
export type Resource = string;
export type Action = string;
export type UserRole = 'user' | 'admin';

// System resources
export const RESOURCES = {
  USER: 'user',
  ADMIN: 'admin', 
  PROFILE: 'profile',
  AUTH: 'auth',
  AUDIT: 'audit',
  SYSTEM: 'system',
} as const;

// Actions
export const ACTIONS = {
  READ: 'read',
  WRITE: 'write',
  CREATE: 'create',
  UPDATE: 'update',
  DELETE: 'delete',
  MANAGE: 'manage',
} as const;

// Permissions
export const PERMISSIONS = {
  // User permissions
  USER_READ: 'user:read',
  USER_WRITE: 'user:write',
  USER_CREATE: 'user:create',
  USER_UPDATE: 'user:update',
  USER_DELETE: 'user:delete',
  USER_MANAGE: 'user:manage',

  // Profile permissions (own profile)
  PROFILE_READ: 'profile:read',
  PROFILE_UPDATE: 'profile:update',

  // Admin permissions
  ADMIN_READ: 'admin:read',
  ADMIN_WRITE: 'admin:write',
  ADMIN_MANAGE: 'admin:manage',

  // Auth permissions
  AUTH_READ: 'auth:read',
  AUTH_WRITE: 'auth:write',
  AUTH_MANAGE: 'auth:manage',

  // Audit permissions
  AUDIT_READ: 'audit:read',
  AUDIT_WRITE: 'audit:write',
  AUDIT_MANAGE: 'audit:manage',

  // System permissions
  SYSTEM_READ: 'system:read',
  SYSTEM_WRITE: 'system:write',
  SYSTEM_MANAGE: 'system:manage',
} as const;

// Role permissions mapping
const ROLE_PERMISSIONS: Record<UserRole, Permission[]> = {
  user: [
    PERMISSIONS.PROFILE_READ,
    PERMISSIONS.PROFILE_UPDATE,
    PERMISSIONS.AUTH_READ,
    PERMISSIONS.AUTH_WRITE,
  ],
  admin: [
    // All user permissions
    PERMISSIONS.PROFILE_READ,
    PERMISSIONS.PROFILE_UPDATE,
    PERMISSIONS.AUTH_READ,
    PERMISSIONS.AUTH_WRITE,
    // Plus admin-specific permissions
    PERMISSIONS.USER_READ,
    PERMISSIONS.USER_WRITE,
    PERMISSIONS.USER_CREATE,
    PERMISSIONS.USER_UPDATE,
    PERMISSIONS.USER_DELETE,
    PERMISSIONS.USER_MANAGE,
    PERMISSIONS.ADMIN_READ,
    PERMISSIONS.ADMIN_WRITE,
    PERMISSIONS.ADMIN_MANAGE,
    PERMISSIONS.AUDIT_READ,
    PERMISSIONS.AUDIT_WRITE,
    PERMISSIONS.SYSTEM_READ,
  ],
};

// RBAC Context Interface
interface RBACContextType {
  // Permission checking
  hasPermission: (permission: Permission) => boolean;
  hasAnyPermission: (permissions: Permission[]) => boolean;
  hasAllPermissions: (permissions: Permission[]) => boolean;
  canAccessResource: (resource: Resource, action: Action) => boolean;

  // Role checking
  hasRole: (role: UserRole) => boolean;
  hasAnyRole: (roles: UserRole[]) => boolean;
  isAdmin: () => boolean;
  isUser: () => boolean;

  // User management
  canManageUser: (targetUserId?: number) => boolean;
  canAccessOwnResource: (resourceOwnerId: number) => boolean;

  // Utility functions
  getUserRole: () => UserRole | null;
  getUserPermissions: () => Permission[];
  buildPermission: (resource: Resource, action: Action) => Permission;
}

// Create context
const RBACContext = createContext<RBACContextType | undefined>(undefined);

// Provider props
interface RBACProviderProps {
  children: ReactNode;
}

// RBAC Provider component
export const RBACProvider: React.FC<RBACProviderProps> = ({ children }) => {
  const { user, isAuthenticated } = useAuth();

  // Get current user role
  const getUserRole = (): UserRole | null => {
    if (!isAuthenticated || !user) {
      return null;
    }
    return user.role;
  };

  // Get user permissions based on role
  const getUserPermissions = (): Permission[] => {
    const role = getUserRole();
    if (!role) {
      return [];
    }
    return ROLE_PERMISSIONS[role] || [];
  };

  // Build permission string from resource and action
  const buildPermission = (resource: Resource, action: Action): Permission => {
    return `${resource}:${action}`;
  };

  // Check if user has a specific permission
  const hasPermission = (permission: Permission): boolean => {
    const permissions = getUserPermissions();
    return permissions.includes(permission);
  };

  // Check if user has any of the specified permissions
  const hasAnyPermission = (permissions: Permission[]): boolean => {
    const userPermissions = getUserPermissions();
    return permissions.some(permission => userPermissions.includes(permission));
  };

  // Check if user has all of the specified permissions
  const hasAllPermissions = (permissions: Permission[]): boolean => {
    const userPermissions = getUserPermissions();
    return permissions.every(permission => userPermissions.includes(permission));
  };

  // Check if user can access a resource with a specific action
  const canAccessResource = (resource: Resource, action: Action): boolean => {
    const permission = buildPermission(resource, action);
    return hasPermission(permission);
  };

  // Check if user has a specific role
  const hasRole = (role: UserRole): boolean => {
    const userRole = getUserRole();
    return userRole === role;
  };

  // Check if user has any of the specified roles
  const hasAnyRole = (roles: UserRole[]): boolean => {
    const userRole = getUserRole();
    return userRole ? roles.includes(userRole) : false;
  };

  // Check if user is admin
  const isAdmin = (): boolean => {
    return hasRole('admin');
  };

  // Check if user is regular user
  const isUser = (): boolean => {
    return hasRole('user');
  };

  // Check if user can manage another user
  const canManageUser = (targetUserId?: number): boolean => {
    // Must have user management permission
    if (!hasPermission(PERMISSIONS.USER_MANAGE)) {
      return false;
    }

    // Cannot manage yourself through admin interface
    if (targetUserId && user && user.id === targetUserId) {
      return false;
    }

    return true;
  };

  // Check if user can access their own resource or has permission
  const canAccessOwnResource = (resourceOwnerId: number): boolean => {
    // Allow if accessing own resource
    if (user && user.id === resourceOwnerId) {
      return true;
    }

    // Otherwise, need appropriate admin permissions
    return hasPermission(PERMISSIONS.USER_READ);
  };

  // Context value
  const contextValue: RBACContextType = {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    canAccessResource,
    hasRole,
    hasAnyRole,
    isAdmin,
    isUser,
    canManageUser,
    canAccessOwnResource,
    getUserRole,
    getUserPermissions,
    buildPermission,
  };

  return (
    <RBACContext.Provider value={contextValue}>
      {children}
    </RBACContext.Provider>
  );
};

// Custom hook to use RBAC context
export const useRBAC = (): RBACContextType => {
  const context = useContext(RBACContext);
  
  if (context === undefined) {
    throw new Error('useRBAC must be used within an RBACProvider');
  }
  
  return context;
};

// Custom hooks for common checks

// Hook to check if user has a specific permission
export const usePermission = (permission: Permission): boolean => {
  const { hasPermission } = useRBAC();
  return hasPermission(permission);
};

// Hook to check if user has any of the specified permissions
export const useAnyPermission = (permissions: Permission[]): boolean => {
  const { hasAnyPermission } = useRBAC();
  return hasAnyPermission(permissions);
};

// Hook to check if user has all of the specified permissions
export const useAllPermissions = (permissions: Permission[]): boolean => {
  const { hasAllPermissions } = useRBAC();
  return hasAllPermissions(permissions);
};

// Hook to check if user can access a resource
export const useResourceAccess = (resource: Resource, action: Action): boolean => {
  const { canAccessResource } = useRBAC();
  return canAccessResource(resource, action);
};

// Hook to check if user has a specific role
export const useRole = (role: UserRole): boolean => {
  const { hasRole } = useRBAC();
  return hasRole(role);
};

// Hook to check if user is admin
export const useIsAdmin = (): boolean => {
  const { isAdmin } = useRBAC();
  return isAdmin();
};

// Hook to check if user is regular user
export const useIsUser = (): boolean => {
  const { isUser } = useRBAC();
  return isUser();
};

// Hook to require specific permission (throws error if not authorized)
export const useRequirePermission = (permission: Permission): void => {
  const { hasPermission } = useRBAC();
  const { isAuthenticated } = useAuth();
  
  React.useEffect(() => {
    if (!isAuthenticated) {
      throw new Error('Authentication required');
    }
    
    if (!hasPermission(permission)) {
      throw new Error(`Permission required: ${permission}`);
    }
  }, [isAuthenticated, hasPermission, permission]);
};

// Hook to require admin role
export const useRequireAdmin = (): void => {
  const { isAdmin } = useRBAC();
  const { isAuthenticated } = useAuth();
  
  React.useEffect(() => {
    if (!isAuthenticated) {
      throw new Error('Authentication required');
    }
    
    if (!isAdmin()) {
      throw new Error('Admin role required');
    }
  }, [isAuthenticated, isAdmin]);
};

export default RBACContext;